/*
 *  Copyright (c) 2019 Kumuluz and/or its affiliates
 *  and other contributors as indicated by the @author tags and
 *  the contributor list.
 *
 *  Licensed under the MIT License (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  https://opensource.org/licenses/MIT
 *
 *  The software is provided "AS IS", WITHOUT WARRANTY OF ANY KIND, express or
 *  implied, including but not limited to the warranties of merchantability,
 *  fitness for a particular purpose and noninfringement. in no event shall the
 *  authors or copyright holders be liable for any claim, damages or other
 *  liability, whether in an action of contract, tort or otherwise, arising from,
 *  out of or in connection with the software or the use or other dealings in the
 *  software. See the License for the specific language governing permissions and
 *  limitations under the License.
 */

// Package config provides configuration management for the KumuluzEE microservice framework.
package config

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/mc0239/logm"
)

// Util is used for retrieving config values from available sources.
// Util should be initialized with config.NewUtil() function
type Util struct {
	configSources []configSource
	logger        *logm.Logm
}

// Bundle is used for filling a user-defined struct with config values.
// Bundle should be initialized with config.NewBundle() function
type Bundle struct {
	prefixKey string
	fields    interface{}
	conf      Util
	Logger    logm.Logm
}

// Options struct is used when instantiating a new Util or Bundle.
type Options struct {
	// ConfigPath is a path to configuration file, including the configuration file name.
	// Passing an empty string will default to config/config.yaml
	ConfigPath string
	// Additional configuration source to connect to. Possible values are: "consul", "etcd"
	Extension string
	// Additional configuration source's namespace to use (i.e. path prefix). Setting this to a
	// non-empty value overwrites default namespace or namespace defined in configuration file
	ExtensionNamespace string
	// LogLevel can be used to limit the amount of logging output. Default log level is 0. Level 4
	// will only output Warnings and Errors, and level 5 will only output errors.
	// See package github.com/mc0239/logm for more details on logging and log levels.
	LogLevel int
}

type configSource interface {
	Name() string
	ordinal() int
	Get(key string) interface{}
	Subscribe(key string, callback func(key string, value string))
}

// NewUtil instantiates a new Util with given options
func NewUtil(options Options) Util {
	lgr := logm.New("KumuluzEE-config")
	lgr.LogLevel = options.LogLevel

	configs := make([]configSource, 0)

	if envConfigSource := newEnvConfigSource(&lgr); envConfigSource != nil {
		configs = append(configs, envConfigSource)
	}

	fileConfigSource := newFileConfigSource(options.ConfigPath, &lgr)
	if fileConfigSource != nil {
		configs = append(configs, fileConfigSource)
	} else {
		lgr.Error("File configuration source failed to load!")
	}

	k := Util{
		configs,
		&lgr,
	}

	k.sortConfigSources()

	// use already initialized env/file config util to get values for initialization of extension
	// config source (consul/etcd)
	var extConfigSource configSource
	switch options.Extension {
	case "consul":
		extConfigSource = newConsulConfigSource(k, options.ExtensionNamespace, &lgr)
		break
	case "etcd":
		extConfigSource = newEtcdConfigSource(k, options.ExtensionNamespace, &lgr)
		break
	case "":
		// no extension
		break
	default:
		lgr.Error("Invalid extension specified, extension configuration source will not be available")
		break
	}

	// if extension config source was successfuly initialized, add it to sources and sort again
	if extConfigSource != nil {
		k.configSources = append(k.configSources, extConfigSource)
	}

	k.sortConfigSources()

	return k
}

// NewBundle fills the given fields struct with values from loaded configuration
func NewBundle(prefixKey string, fields interface{}, options Options) Bundle {
	lgr := logm.New("KumuluzEE-config")
	lgr.LogLevel = options.LogLevel

	util := NewUtil(options)

	bun := Bundle{
		prefixKey: prefixKey,
		fields:    &fields,
		conf:      util,
		Logger:    lgr,
	}

	traverseStruct(fields, prefixKey,
		func(key string, value reflect.Value, field reflect.StructField, tags reflect.StructTag) {

			// fill struct value using util
			setValueWithReflect(key, value, field, bun)

			// register watch on fields with tag config:",watch"

			if tag, ok := tags.Lookup("config"); ok {
				tagVals := strings.Split(tag, ",")
				if len(tagVals) > 1 && tagVals[1] == "watch" {
					util.Subscribe(key, func(watchKey string, newValue string) {
						setValueWithReflect(key, value, field, bun)
						//value.Set(reflect.ValueOf(newValue))
						lgr.Verbose("Watched value %s updated, new value: %s", key, newValue)
					})
				}
			}

		},
	)

	return bun
}

// Subscribe creates a watch on a given configuration key.
// Note that watch will be enabled on an extension configuration source, if one has been defined
// when Util was created.
// When value in configuration updates, callback is fired with the key and the new value.
func (c Util) Subscribe(key string, callback func(key string, value string)) {

	// find extension configSource and deploy a watch
	for _, cs := range c.configSources {
		cs.Subscribe(key, callback)
	}

}

// Get returns the value for a given key, stored in configuration.
// Configuration sources are checked by their ordinal numbers, and value is returned from first
// configuration source it was found in.
func (c Util) Get(key string) interface{} {
	// iterate through configSources and try to get some value ...
	var val interface{}

	for _, cs := range c.configSources {
		val = cs.Get(key)
		if val != nil {
			break
		}
	}
	return val
}

// GetBool is a helper method that calls Util.Get() internally and type asserts the value to
// bool before returning it.
// If value is not found in any configuration source or the value could not be type asserted to
// bool, a false is returned with ok equal to false.
func (c Util) GetBool(key string) (value bool, ok bool) {
	rvalue := c.Get(key)

	bvalue, ok := c.Get(key).(bool)
	if ok {
		return bvalue, true
	}

	svalue, ok := rvalue.(string)
	if ok {
		bvalue, err := strconv.ParseBool(string(svalue))
		if err == nil {
			return bvalue, true
		}
	}

	return false, false
}

// GetInt is a helper method that calls Util.Get() internally and type asserts the value to
// int before returning it.
// If value is not found in any configuration source or the value could not be type asserted to
// int, a zero is returned with ok equal to false.
func (c Util) GetInt(key string) (value int, ok bool) {
	rvalue := c.Get(key)

	// try to assert as any number type
	nvalue, ok := assertAsNumber(rvalue)
	if ok {
		return int(nvalue), true
	}

	// try to assert as string and convert to int
	svalue, ok := rvalue.(string)
	if ok {
		ivalue64, err := strconv.ParseInt(svalue, 0, 64)
		if err == nil {
			return int(ivalue64), true
		}
		fvalue64, err := strconv.ParseFloat(svalue, 64)
		if err == nil {
			return int(fvalue64), true
		}
	}

	return 0, false
}

// GetFloat is a helper method that calls Util.Get() internally and type asserts the value to
// float64 before returning it.
// If value is not found in any configuration source or the value could not be type asserted to
// float64, a zero is returned with ok equal to false.
func (c Util) GetFloat(key string) (value float64, ok bool) {
	rvalue := c.Get(key)

	// try to assert as any number type
	nvalue, ok := assertAsNumber(rvalue)
	if ok {
		return nvalue, true
	}

	// try to assert as string and convert to float64
	svalue, ok := rvalue.(string)
	if ok {
		fvalue64, err := strconv.ParseFloat(svalue, 64)
		if err == nil {
			return fvalue64, true
		}
	}

	return 0, false
}

// GetString is a helper method that calls Util.Get() internally and type asserts the value to
// string before returning it.
// If value is not found in any configuration source or the value could not be type asserted to
// string, an empty string is returned with ok equal to false.
func (c Util) GetString(key string) (value string, ok bool) {
	// try to type assert as string
	svalue, ok := c.Get(key).(string)
	if ok {
		return svalue, true
	}

	// can't assert to string, return nil
	return "", false
}

// sort config sources by ordinal numbers
func (c Util) sortConfigSources() {
	// insertion sort
	for i := 1; i < len(c.configSources); i++ {
		for k := i; k > 0 && c.configSources[k].ordinal() > c.configSources[k-1].ordinal(); k-- {
			// swap
			temp := c.configSources[k]
			c.configSources[k] = c.configSources[k-1]
			c.configSources[k-1] = temp
		}
	}
}
