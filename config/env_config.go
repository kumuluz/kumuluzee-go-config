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
	"os"
	"regexp"
	"strings"

	"github.com/mc0239/logm"
)

type envConfigSource struct {
}

func newEnvConfigSource(lgr *logm.Logm) configSource {
	var c envConfigSource
	lgr.Verbose("Initializing %s config source", c.Name())
	lgr.Verbose("Initialized %s config source", c.Name())
	return c
}

func (c envConfigSource) Get(key string) interface{} {

	for _, keyName := range getPossibleNames(key) {
		value, exists := os.LookupEnv(keyName)
		if exists {
			return value
		}
	}

	return nil
}

func (c envConfigSource) Subscribe(key string, callback func(key string, value string)) {
	return
}

func (c envConfigSource) Name() string {
	return "env"
}

func (c envConfigSource) ordinal() int {
	return 300
}

//

// https://github.com/kumuluz/kumuluzee/blob/master/common/src/main/java/com/kumuluz/ee/configuration/sources/EnvironmentConfigurationSource.java#L224
func getPossibleNames(key string) []string {
	possibleNames := []string{
		// MP Config 1.3: raw key
		key,
		normalizeKey(key),
		normalizeKeyUpper(key),
		parseKeyLegacy1(key),
		parseKeyLegacy2(key),
	}

	return possibleNames
}

// MP Config 1.3: replaces non alpha-numeric characters with '_'
func normalizeKey(key string) string {
	re1 := regexp.MustCompile("[^a-zA-Z0-9]")
	normKey := re1.ReplaceAllString(key, "_")
	return normKey
}

func normalizeKeyUpper(key string) string {
	return strings.ToUpper(normalizeKey(key))
}

// legacy 1: removes characters '[]-' and replaces dots with '_', to uppercase
func parseKeyLegacy1(key string) string {
	return strings.ToUpper(
		strings.Replace(strings.Replace(strings.Replace(strings.Replace(
			key,
			"[", "", -1),
			"]", "", -1),
			"-", "", -1),
			".", "_", -1))
}

// legacy 2: replaces dots with '_', to uppercase
func parseKeyLegacy2(key string) string {
	return strings.ToUpper(strings.Replace(key, ".", "_", -1))
}
