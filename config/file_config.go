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
	"io/ioutil"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/mc0239/logm"
)

type fileConfigSource struct {
	config map[string]interface{}
	logger *logm.Logm
}

func newFileConfigSource(configPath string, lgr *logm.Logm) configSource {
	var c fileConfigSource
	lgr.Verbose("Initializing %s config source", c.Name())
	c.logger = lgr

	var joinedPath string
	if configPath == "" {
		// set default
		joinedPath = "config.yaml"
	} else {
		joinedPath = configPath
	}

	lgr.Verbose(fmt.Sprintf("Config file path: %s\n", joinedPath))

	bytes, err := ioutil.ReadFile(joinedPath)
	if err != nil {
		lgr.Error(fmt.Sprintf("Failed to read file on path: %s, error: %s", joinedPath, err.Error()))
		return nil
	}
	//fmt.Printf("Read: %s", bytes)

	err = yaml.Unmarshal(bytes, &c.config)
	if err != nil {
		lgr.Error(fmt.Sprintf("Failed tu unmarshal yaml: %s", err.Error()))
		return nil
	}

	lgr.Verbose("Initialized %s config source", c.Name())
	return c
}

func (c fileConfigSource) Get(key string) interface{} {
	//fmt.Println("[fileConfigSource] Get: " + key)
	tree := strings.Split(key, ".")

	// move deeper into maps for every dot delimiter
	val := c.config
	var assertOk bool
	for i := 0; i < len(tree)-1; i++ {
		if val == nil {
			return nil
		}
		val, assertOk = val[tree[i]].(map[string]interface{})

		if !assertOk {
			return nil
		}
		//fmt.Printf("%d ::: %v\n", i, val)
	}

	return val[tree[len(tree)-1]]
}

func (c fileConfigSource) Subscribe(key string, callback func(key string, value string)) {
	return
}

func (c fileConfigSource) Name() string {
	return "file"
}

func (c fileConfigSource) ordinal() int {
	return 100
}

//
