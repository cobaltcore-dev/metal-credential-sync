// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and CobaltCore contributors
// SPDX-License-Identifier: Apache-2.0

package vault

import (
	"fmt"
	"os"
	"time"
)

const (
	defaultServiceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

// authenticate authenticates with Vault using the configured method
func (v *VaultBackend) authenticate(config *Config) error {
	switch config.AuthMethod {
	case "kubernetes":
		return v.authenticateKubernetes(config)
	case "token":
		return v.authenticateToken(config)
	case "approle":
		return fmt.Errorf("approle authentication not yet implemented")
	default:
		return fmt.Errorf("unsupported auth method: %s", config.AuthMethod)
	}
}

// authenticateKubernetes authenticates using Kubernetes service account
func (v *VaultBackend) authenticateKubernetes(config *Config) error {
	start := time.Now()

	// Read service account token
	tokenBytes, err := os.ReadFile(defaultServiceAccountTokenPath)
	if err != nil {
		if v.metricsCollector != nil {
			v.metricsCollector.RecordAuth("kubernetes", "vault", time.Since(start), err)
		}
		return fmt.Errorf("failed to read service account token: %w", err)
	}
	jwt := string(tokenBytes)

	// Prepare login data
	loginData := map[string]any{
		"jwt":  jwt,
		"role": config.KubernetesAuthRole,
	}

	// Login to Vault
	authPath := fmt.Sprintf("auth/%s/login", config.KubernetesAuthPath)
	secret, err := v.client.Logical().Write(authPath, loginData)
	if err != nil {
		if v.metricsCollector != nil {
			v.metricsCollector.RecordAuth("kubernetes", "vault", time.Since(start), err)
		}
		return fmt.Errorf("kubernetes auth login failed: %w", err)
	}

	if secret == nil || secret.Auth == nil || secret.Auth.ClientToken == "" {
		err = fmt.Errorf("kubernetes auth returned no token")
		if v.metricsCollector != nil {
			v.metricsCollector.RecordAuth("kubernetes", "vault", time.Since(start), err)
		}
		return err
	}

	// Set the token
	v.client.SetToken(secret.Auth.ClientToken)

	if v.metricsCollector != nil {
		v.metricsCollector.RecordAuth("kubernetes", "vault", time.Since(start), nil)
	}

	return nil
}

// authenticateToken authenticates using a pre-configured token
func (v *VaultBackend) authenticateToken(config *Config) error {
	start := time.Now()

	if config.Token == "" {
		err := fmt.Errorf("token is required for token authentication")
		if v.metricsCollector != nil {
			v.metricsCollector.RecordAuth("token", "vault", time.Since(start), err)
		}
		return err
	}

	v.client.SetToken(config.Token)

	// Verify token is valid
	_, err := v.client.Auth().Token().LookupSelf()
	if v.metricsCollector != nil {
		v.metricsCollector.RecordAuth("token", "vault", time.Since(start), err)
	}

	if err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	return nil
}
