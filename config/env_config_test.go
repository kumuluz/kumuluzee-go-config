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

func envAssert(t *testing.T, expected interface{}, got interface{}) {
	if expected != got {
		t.Errorf("expected=%v, got=%v", expected, got)
	}
}

func TestEnvKey(t *testing.T) {
	keys := []string{"kumuluzee", "KumuluzEE[0]", "lev1.lev2[5].LEV3", "v€ry-c00l"}
	expNorm := []string{"kumuluzee", "KumuluzEE_0_", "lev1_lev2_5__LEV3", "v_ry_c00l"}
	expNorm2 := []string{"KUMULUZEE", "KUMULUZEE_0_", "LEV1_LEV2_5__LEV3", "V_RY_C00L"}
	expLeg1 := []string{"KUMULUZEE", "KUMULUZEE0", "LEV1_LEV25_LEV3", "V€RYC00L"}
	expLeg2 := []string{"KUMULUZEE", "KUMULUZEE[0]", "LEV1_LEV2[5]_LEV3", "V€RY-C00L"}

	for i, keyName := range keys {
		envAssert(t, expNorm[i], normalizeKey(keyName))
		envAssert(t, expNorm2[i], normalizeKeyUpper(keyName))
		envAssert(t, expLeg1[i], parseKeyLegacy1(keyName))
		envAssert(t, expLeg2[i], parseKeyLegacy2(keyName))
	}
}
