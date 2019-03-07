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

package config

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/mc0239/logm"

	"github.com/hashicorp/consul/api"
)

type consulConfigSource struct {
	client          *api.Client
	startRetryDelay int64
	maxRetryDelay   int64
	namespace       string
	logger          *logm.Logm
}

func newConsulConfigSource(conf Util, namespace string, lgr *logm.Logm) configSource {
	var consulConfig consulConfigSource
	lgr.Verbose("Initializing %s config source", consulConfig.Name())
	consulConfig.logger = lgr

	var consulAddress string
	if addr, ok := conf.GetString("kumuluzee.config.consul.hosts"); ok {
		consulAddress = addr
	} else {
		consulAddress = "http://localhost:8500"
	}

	if client, err := createConsulClient(consulAddress); err == nil {
		lgr.Info("Consul client address set to %v", consulAddress)
		consulConfig.client = client
	} else {
		lgr.Error("Failed to create Consul client: %s", err.Error())
	}

	envName, name, version, startRD, maxRD := loadServiceConfiguration(conf)
	consulConfig.startRetryDelay = startRD
	consulConfig.maxRetryDelay = maxRD
	lgr.Verbose("start-retry-delay-ms=%d, max-retry-delay-ms=%d", consulConfig.startRetryDelay, consulConfig.maxRetryDelay)

	consulConfig.namespace = fmt.Sprintf("environments/%s/services/%s/%s/config", envName, name, version)
	// namespace can be overwritten from configuration file ...
	if ns, ok := conf.GetString("kumuluzee.config.namespace"); ok {
		if ns != "" {
			consulConfig.namespace = ns
		}
	}
	// ... or programmatically by passing it into config.Options
	if namespace != "" {
		consulConfig.namespace = namespace
	}

	lgr.Info("%s key-value namespace: %s", consulConfig.Name(), consulConfig.namespace)
	lgr.Verbose("Initialized %s config source", consulConfig.Name())
	return consulConfig
}

func (c consulConfigSource) Get(key string) interface{} {
	//fmt.Println("[consulConfigSource] Get: " + key)
	kv := c.client.KV()

	key = strings.Replace(key, ".", "/", -1)
	//fmt.Printf("KV path: %s\n", path.Join(c.namespace, key))

	pair, _, err := kv.Get(path.Join(c.namespace, key), nil)
	if err != nil {
		c.logger.Warning("Error getting value: %v", err)
		return nil
	}

	//fmt.Printf("Pair received: %v\n", pair)
	if pair == nil {
		return nil
	}
	// pair.Value is type []byte
	return string(pair.Value)
}

func (c consulConfigSource) Subscribe(key string, callback func(key string, value string)) {
	c.logger.Info("Creating a watch: key=%s. namespace=%s source=%s", key, c.namespace, c.Name())
	go c.watch(key, "", c.startRetryDelay, callback, 0)
}

func (c consulConfigSource) Name() string {
	return "consul"
}

func (c consulConfigSource) ordinal() int {
	return 150
}

// functions that aren't configSource methods

func (c consulConfigSource) watch(key string, previousValue string, retryDelay int64, callback func(key string, value string), waitIndex uint64) {

	q := api.QueryOptions{
		WaitIndex: waitIndex,
		WaitTime:  10 * time.Minute,
	}

	key = strings.Replace(key, ".", "/", -1)
	c.logger.Verbose("Setting a watch on key %s with %s wait time", key, q.WaitTime)

	pair, meta, err := c.client.KV().Get(path.Join(c.namespace, key), &q)

	if err != nil {
		c.logger.Warning("Watch on %s failed with error: %s, retry delay: %d ms", key, err.Error(), retryDelay)

		// sleep for current delay
		time.Sleep(time.Duration(retryDelay) * time.Millisecond)

		// exponentially extend retry delay, but keep it at most maxRetryDelay
		newRetryDelay := retryDelay * 2
		if newRetryDelay > c.maxRetryDelay {
			newRetryDelay = c.maxRetryDelay
		}
		c.watch(key, "", newRetryDelay, callback, 0)
		return
	}

	c.logger.Verbose("Wait time (%s) on watch for key %s reached.", q.WaitTime, key)

	if pair != nil {
		if string(pair.Value) != previousValue {
			callback(key, string(pair.Value))
		}
		c.watch(key, string(pair.Value), c.startRetryDelay, callback, meta.LastIndex)
	} else {
		if previousValue != "" {
			callback(key, "")
		}
		var lastIndex uint64
		if meta != nil {
			lastIndex = meta.LastIndex
		}
		c.watch(key, "", c.startRetryDelay, callback, lastIndex)
	}
}

// functions that aren't configSource methods or etcdCondigSource methods

func createConsulClient(address string) (*api.Client, error) {
	clientConfig := api.DefaultConfig()
	clientConfig.Address = address

	client, err := api.NewClient(clientConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}
