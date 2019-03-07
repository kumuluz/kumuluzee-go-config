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
	"testing"
)

func consulAssert(t *testing.T, expected interface{}, got interface{}) {
	t.Errorf("expected=%v, got=%v", expected, got)
}

func TestConsulConfigGetInt(t *testing.T) {
	c := NewUtil(Options{
		Extension:  "consul",
		ConfigPath: "../test/config-consul.yaml",
		LogLevel:   100, // turn off logging
	})
	if i, ok := c.GetInt("integer-value"); !(ok && i == 36) {
		consulAssert(t, 36, i)
	}
	if i, ok := c.GetInt("float-value"); !(ok && i == 11) {
		consulAssert(t, 11, i)
	}
	if i, ok := c.GetInt("negative-integer-value"); !(ok && i == -4) {
		consulAssert(t, -4, i)
	}
}

func TestConsulConfigGetString(t *testing.T) {
	c := NewUtil(Options{
		Extension:  "consul",
		ConfigPath: "../test/config-consul.yaml",
		LogLevel:   100, // turn off logging
	})
	if s, ok := c.GetString("string-value"); !(ok && s == "hey ho") {
		consulAssert(t, "hey ho", s)
	}
	if s, ok := c.GetString("empty-string-value"); !(ok && s == "") {
		consulAssert(t, "", s)
	}
}

func TestConsulConfigGetFloat(t *testing.T) {
	c := NewUtil(Options{
		Extension:  "consul",
		ConfigPath: "../test/config-consul.yaml",
		LogLevel:   100, // turn off logging
	})
	if f, ok := c.GetFloat("float-value"); !(ok && f == 11.65425) {
		consulAssert(t, 11.65425, f)
	}
	if f, ok := c.GetFloat("integer-value"); !(ok && f == 36.0) {
		consulAssert(t, 36.0, f)
	}
	if f, ok := c.GetFloat("negative-float-value"); !(ok && f == -0.411) {
		consulAssert(t, -0.411, f)
	}
}

func TestConsulConfigGetBool(t *testing.T) {
	c := NewUtil(Options{
		Extension:  "consul",
		ConfigPath: "../test/config-consul.yaml",
		LogLevel:   100, // turn off logging
	})
	if b, ok := c.GetBool("boolean-value-1"); !(ok && b) {
		consulAssert(t, true, b)
	}
	if b, ok := c.GetBool("boolean-value-2"); !(ok && b) {
		consulAssert(t, true, b)
	}
	if b, ok := c.GetBool("boolean-value-3"); !(ok && b) {
		consulAssert(t, true, b)
	}
}

func TestConsulConfigUseCase(t *testing.T) {
	c := NewUtil(Options{
		Extension:  "consul",
		ConfigPath: "../test/config-consul.yaml",
		LogLevel:   100, // turn off logging
	})
	if s, ok := c.GetString("some-config.protocol"); !(ok && s == "tcp") {
		consulAssert(t, "tcp", s)
	}
	if s, ok := c.GetString("some-config.address.ip"); !(ok && s == "127.0.0.2") {
		consulAssert(t, "127.0.0.2", s)
	}
	if i, ok := c.GetInt("some-config.address.port"); !(ok && i == 3000) {
		consulAssert(t, 3000, i)
	}
}

func TestConsulConfigBundle(t *testing.T) {
	type someConfig struct {
		Protocol string
		Address  struct {
			IP   string `mapstructure:"ip"`
			Port int
		}
		Version  string
		SomeBool bool `mapstructure:"some-boolean"`
	}

	sc := someConfig{}

	NewBundle("some-config", &sc, Options{
		Extension:  "consul",
		ConfigPath: "../test/config-consul.yaml",
		LogLevel:   100, // turn off logging
	})

	if sc.Protocol != "tcp" {
		t.Errorf("Test failed: bundle, expected: %s, got: %s", "tcp", sc.Protocol)
	}
	if sc.Address.IP != "127.0.0.2" {
		t.Errorf("Test failed: bundle, expected: %s, got: %s", "127.0.0.2", sc.Address.IP)
	}
	if sc.Address.Port != 3000 {
		t.Errorf("Test failed: bundle, expected: %d, got: %d", 3000, sc.Address.Port)
	}
	if sc.SomeBool != true {
		t.Errorf("Test failed: bundle, expected: %v, got: %v", true, sc.SomeBool)
	}
}

func TestConsulConfigDeep(t *testing.T) {
	c := NewUtil(Options{
		Extension:  "consul",
		ConfigPath: "../test/config-consul.yaml",
		LogLevel:   100, // turn off logging
	})

	if i, ok := c.GetInt("deep-config.l1.l2.l_3.l-4.l 5.6l"); !(ok && i == 6) {
		consulAssert(t, 6, i)
	}
}
