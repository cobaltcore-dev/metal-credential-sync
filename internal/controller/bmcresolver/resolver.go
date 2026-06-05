// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and CobaltCore contributors
// SPDX-License-Identifier: Apache-2.0

package bmcresolver

import (
	"context"
	"fmt"

	metalv1alpha1 "github.com/ironcore-dev/metal-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FindBMCsForSecret returns all BMC resources referencing the given BMCSecret name
func FindBMCsForSecret(ctx context.Context, c client.Client, secretName string) ([]metalv1alpha1.BMC, error) {
	var bmcList metalv1alpha1.BMCList
	if err := c.List(ctx, &bmcList); err != nil {
		return nil, fmt.Errorf("failed to list BMC resources: %w", err)
	}

	var matchingBMCs []metalv1alpha1.BMC
	for _, bmc := range bmcList.Items {
		if bmc.Spec.BMCSecretRef.Name == secretName {
			matchingBMCs = append(matchingBMCs, bmc)
		}
	}

	return matchingBMCs, nil
}

// ExtractRegionFromBMC gets the region from BMC labels using configurable key
func ExtractRegionFromBMC(bmc *metalv1alpha1.BMC, regionLabelKey string) string {
	if bmc.Labels == nil {
		return "unknown"
	}

	region, ok := bmc.Labels[regionLabelKey]
	if !ok || region == "" {
		return "unknown"
	}

	return region
}

// GetHostnameFromBMC extracts the hostname field, with fallback to name
func GetHostnameFromBMC(bmc *metalv1alpha1.BMC) string {
	if bmc.Spec.Hostname != nil && *bmc.Spec.Hostname != "" {
		return *bmc.Spec.Hostname
	}

	// Check EndpointRef
	if bmc.Spec.EndpointRef != nil && bmc.Spec.EndpointRef.Name != "" {
		return bmc.Spec.EndpointRef.Name
	}

	// Fallback to BMC name
	return bmc.Name
}
