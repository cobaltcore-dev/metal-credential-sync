// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and CobaltCore contributors
// SPDX-License-Identifier: Apache-2.0

package secretbackend

import (
	"fmt"
	"strings"

	configv1alpha1 "github.com/ironcore-dev/metal-credential-sync/api/v1alpha1"
	"github.com/ironcore-dev/metal-credential-sync/internal/secretbackend/vault"
)

// EngineBackend represents a backend configured for a specific secret engine
type EngineBackend struct {
	Backend      Backend
	EngineName   string
	PathBuilder  *PathBuilder
	SyncLabel    string
	SyncLabelKey string
	SyncLabelVal string
}

// MatchesLabels checks if the given labels match this engine's sync label requirements
func (e *EngineBackend) MatchesLabels(labels map[string]string) bool {
	if labels == nil {
		return false
	}

	// If syncLabelVal is empty, just check if the key exists with any value
	if e.SyncLabelVal == "" {
		_, exists := labels[e.SyncLabelKey]
		return exists
	}

	// Otherwise, check for exact match
	val, exists := labels[e.SyncLabelKey]
	return exists && val == e.SyncLabelVal
}

// parseSecretEngineConfig parses SecretEngineConfig and creates EngineBackend instances
func parseSecretEngineConfig(
	engines []configv1alpha1.SecretEngineConfig,
	baseVaultConfig *VaultConfigInternal,
	metricsCollector MetricsCollector,
) ([]*EngineBackend, error) {
	var engineBackends []*EngineBackend

	for _, engine := range engines {
		// Create vault config for this engine
		vaultConfig := &vault.Config{
			Address:            baseVaultConfig.Address,
			AuthMethod:         baseVaultConfig.AuthMethod,
			KubernetesAuthRole: baseVaultConfig.KubernetesAuthRole,
			KubernetesAuthPath: baseVaultConfig.KubernetesAuthPath,
			Token:              baseVaultConfig.Token,
			MountPath:          engine.MountPath,
			SkipVerify:         baseVaultConfig.SkipVerify,
			CACert:             baseVaultConfig.CACert,
		}

		// Create backend for this engine
		backend, err := vault.NewVaultBackend(vaultConfig, metricsCollector)
		if err != nil {
			return nil, fmt.Errorf("failed to create backend for engine %s: %w", engine.Name, err)
		}

		// Parse sync label (format: "key" or "key=value")
		syncLabelKey, syncLabelVal := parseSyncLabel(engine.SyncLabel)

		// Create path builder
		pathTemplate := engine.PathTemplate
		if pathTemplate == "" {
			pathTemplate = "bmc/{{.Region}}/{{.Hostname}}/{{.Username}}"
		}
		pathBuilder, err := NewPathBuilder(pathTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to create path builder for engine %s: %w", engine.Name, err)
		}

		engineBackends = append(engineBackends, &EngineBackend{
			Backend:      backend,
			EngineName:   engine.Name,
			PathBuilder:  pathBuilder,
			SyncLabel:    engine.SyncLabel,
			SyncLabelKey: syncLabelKey,
			SyncLabelVal: syncLabelVal,
		})
	}

	return engineBackends, nil
}

// parseSyncLabel parses a sync label into key and value components
// Format: "key" or "key=value"
// Returns: (key, value) where value is empty string if format is just "key"
func parseSyncLabel(syncLabel string) (string, string) {
	parts := strings.SplitN(syncLabel, "=", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}
