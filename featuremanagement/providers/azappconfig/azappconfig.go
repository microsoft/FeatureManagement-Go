// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azappconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/Azure/AppConfiguration-GoProvider/azureappconfiguration"
)

type FeatureFlagProvider struct {
	azappcfg *azureappconfiguration.AzureAppConfiguration
	fm       FeatureManagement
	mu       sync.RWMutex
}

func NewFeatureFlagProvider(azappcfg *azureappconfiguration.AzureAppConfiguration) (*FeatureFlagProvider, error) {
	jsonBytes, err := azappcfg.GetBytes(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get bytes from Azure App Configuration: %w", err)
	}

	var fm FeatureManagement
	if err := json.Unmarshal(jsonBytes, &fm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal feature management: %w", err)
	}

	provider := &FeatureFlagProvider{
		azappcfg: azappcfg,
		fm:       fm,
	}

	// Register refresh callback to update feature management on configuration changes
	azappcfg.OnRefreshSuccess(func() {
		var updatedFM FeatureManagement
		err := azappcfg.Unmarshal(&updatedFM, nil)
		if err != nil {
			log.Printf("Error unmarshalling updated configuration: %s", err)
			return
		}

		provider.mu.Lock()
		defer provider.mu.Unlock()
		provider.fm = updatedFM
	})

	return provider, nil
}

func (p *FeatureFlagProvider) GetFeatureFlags() ([]FeatureFlag, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.fm.FeatureFlags, nil
}

func (p *FeatureFlagProvider) GetFeatureFlag(id string) (FeatureFlag, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, flag := range p.fm.FeatureFlags {
		if flag.ID == id {
			return flag, nil
		}
	}

	return FeatureFlag{}, fmt.Errorf("feature flag with ID %s not found", id)
}
