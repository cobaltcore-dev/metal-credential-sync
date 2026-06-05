// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and CobaltCore contributors
// SPDX-License-Identifier: Apache-2.0

package bmcresolver

import (
	"fmt"

	metalv1alpha1 "github.com/ironcore-dev/metal-operator/api/v1alpha1"
)

// ExtractCredentials gets username and password from BMCSecret data/stringData
func ExtractCredentials(bmcSecret *metalv1alpha1.BMCSecret) (username, password string, err error) {
	// Try to get username from Data first
	if usernameBytes, ok := bmcSecret.Data["username"]; ok {
		username = string(usernameBytes)
	}

	// Try to get password from Data
	if passwordBytes, ok := bmcSecret.Data["password"]; ok {
		password = string(passwordBytes)
	}

	// Validate
	if username == "" {
		return "", "", fmt.Errorf("username not found in BMCSecret data")
	}

	if password == "" {
		return "", "", fmt.Errorf("password not found in BMCSecret data")
	}

	return username, password, nil
}
