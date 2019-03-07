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
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/mc0239/logm"

	"go.etcd.io/etcd/client"
)

type etcdConfigSource struct {
	client          *client.Client
	startRetryDelay int64
	maxRetryDelay   int64
	namespace       string
	logger          *logm.Logm
}

func newEtcdConfigSource(conf Util, namespace string, lgr *logm.Logm) configSource {
	var etcdConfig etcdConfigSource
	lgr.Verbose("Initializing %s config source", etcdConfig.Name())
	etcdConfig.logger = lgr

	var etcdAddress string
	if addr, ok := conf.GetString("kumuluzee.config.etcd.hosts"); ok {
		etcdAddress = addr
	} else {
		etcdAddress = "http://localhost:2379"
	}

	if client, err := createEtcdClient(etcdAddress); err == nil {
		lgr.Info("etcd client address set to %v", etcdAddress)
		etcdConfig.client = client
	} else {
		lgr.Error("Failed to create etcd client: %s", err.Error())
	}

	envName, name, version, startRD, maxRD := loadServiceConfiguration(conf)
	etcdConfig.startRetryDelay = startRD
	etcdConfig.maxRetryDelay = maxRD
	lgr.Verbose("start-retry-delay-ms=%d, max-retry-delay-ms=%d", etcdConfig.startRetryDelay, etcdConfig.maxRetryDelay)

	etcdConfig.namespace = fmt.Sprintf("environments/%s/services/%s/%s/config", envName, name, version)
	// namespace can be overwritten from configuration file ...
	if ns, ok := conf.GetString("kumuluzee.config.namespace"); ok {
		if ns != "" {
			etcdConfig.namespace = ns
		}
	}
	// ... or programmatically by passing it into config.Options
	if namespace != "" {
		etcdConfig.namespace = namespace
	}

	lgr.Info("etcd key-value namespace: %s", etcdConfig.namespace)
	lgr.Verbose("Initialized %s config source", etcdConfig.Name())
	return etcdConfig
}

func (c etcdConfigSource) Get(key string) interface{} {
	kv := client.NewKeysAPI(*c.client)

	key = strings.Replace(key, ".", "/", -1)
	//fmt.Printf("KV path: %s\n", path.Join(c.namespace, key))

	resp, err := kv.Get(context.Background(), path.Join(c.namespace, key), nil)
	if err != nil {
		c.logger.Warning("Error getting value: %v", err)
		return nil
	}

	return resp.Node.Value
}

func (c etcdConfigSource) Subscribe(key string, callback func(key string, value string)) {
	c.logger.Info("Creating a watch for key %s, source: %s", key, c.Name())
	go c.watch(key, "", c.startRetryDelay, callback)
}

func (c etcdConfigSource) Name() string {
	return "etcd"
}

func (c etcdConfigSource) ordinal() int {
	return 150
}

// functions that aren't configSource methods

func (c etcdConfigSource) watch(key string, previousValue string, retryDelay int64, callback func(key string, value string)) {

	c.logger.Verbose("Set a watch on key %s", key)

	key = strings.Replace(key, ".", "/", -1)
	kv := client.NewKeysAPI(*c.client)

	watcher := kv.Watcher(path.Join(c.namespace, key), nil)

	resp, err := watcher.Next(context.Background())
	if err != nil {
		c.logger.Warning("Watch on %s failed with error: %s, retry delay: %d ms", key, err.Error(), retryDelay)

		// sleep for current delay
		time.Sleep(time.Duration(retryDelay) * time.Millisecond)

		// exponentially extend retry delay, but keep it at most maxRetryDelay
		newRetryDelay := retryDelay * 2
		if newRetryDelay > c.maxRetryDelay {
			newRetryDelay = c.maxRetryDelay
		}
		c.watch(key, "", newRetryDelay, callback)
		return
	}

	c.logger.Verbose("Wait time on watch for key %s reached.", key)

	if string(resp.Node.Value) != previousValue {
		callback(key, string(resp.Node.Value))
	}
	c.watch(key, string(resp.Node.Value), c.startRetryDelay, callback)
}

// functions that aren't configSource methods or etcdCondigSource methods

func createEtcdClient(address string) (*client.Client, error) {
	clientConfig := client.Config{}
	clientConfig.Endpoints = []string{address}

	cl, err := client.New(clientConfig)
	if err != nil {
		return nil, err
	}
	return &cl, nil
}
