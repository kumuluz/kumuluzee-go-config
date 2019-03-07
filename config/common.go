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

func loadServiceConfiguration(conf Util) (envName, name, version string, startRD, maxRD int64) {
	if e, ok := conf.GetString("kumuluzee.env.name"); ok {
		envName = e
	} else {
		envName = "dev"
	}

	name, _ = conf.GetString("kumuluzee.name")

	if v, ok := conf.GetString("kumuluzee.version"); ok {
		version = v
	} else {
		version = "1.0.0"
	}

	if sdl, ok := conf.GetInt("kumuluzee.config.start-retry-delay-ms"); ok {
		startRD = int64(sdl)
	} else {
		startRD = 500
	}

	if mdl, ok := conf.GetInt("kumuluzee.config.max-retry-delay-ms"); ok {
		maxRD = int64(mdl)
	} else {
		maxRD = 900000
	}

	return
}

func assertAsNumber(val interface{}) (num float64, ok bool) {
	switch t := val.(type) {
	case int:
		return float64(t), true
	case int8:
		return float64(t), true
	case int16:
		return float64(t), true
	case int32:
		return float64(t), true
	case int64:
		return float64(t), true
	case uint:
		return float64(t), true
	case uint8:
		return float64(t), true
	case uint16:
		return float64(t), true
	case uint32:
		return float64(t), true
	case uint64:
		return float64(t), true

	case float32:
		return float64(t), true
	case float64:
		return float64(t), true
	default:
		return 0, false
	}
}
