// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package azappconfig

import (
	"fmt"
	"log"
	"sync"

	"github.com/Azure/AppConfiguration-GoProvider/azureappconfiguration"
	fm "github.com/microsoft/Featuremanagement-Go/featuremanagement"
)

type featureConfig struct {
	FeatureManagement fm.FeatureManagement `json:"feature_management"`
}

type FeatureFlagProvider struct {
	azappcfg     *azureappconfiguration.AzureAppConfiguration
	featureFlags []fm.FeatureFlag
	mu           sync.RWMutex
}

func NewFeatureFlagProvider(azappcfg *azureappconfiguration.AzureAppConfiguration) (*FeatureFlagProvider, error) {
	var fc featureConfig
	if err := azappcfg.Unmarshal(&fc, nil); err != nil {
		return nil, fmt.Errorf("failed to unmarshal feature management: %w", err)
	}
	provider := &FeatureFlagProvider{
		azappcfg:     azappcfg,
		featureFlags: fc.FeatureManagement.FeatureFlags,
	}

	// Register refresh callback to update feature management on configuration changes
	azappcfg.OnRefreshSuccess(func() {
		var updatedFC featureConfig
		err := azappcfg.Unmarshal(&updatedFC, nil)
		if err != nil {
			log.Printf("Error unmarshalling updated configuration: %s", err)
			return
		}
		provider.mu.Lock()
		defer provider.mu.Unlock()
		provider.featureFlags = updatedFC.FeatureManagement.FeatureFlags
	})

	return provider, nil
}

func (p *FeatureFlagProvider) GetFeatureFlags() ([]fm.FeatureFlag, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.featureFlags, nil
}

func (p *FeatureFlagProvider) GetFeatureFlag(id string) (fm.FeatureFlag, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, flag := range p.featureFlags {
		if flag.ID == id {
			return flag, nil
		}
	}

	return fm.FeatureFlag{}, fmt.Errorf("feature flag with ID %s not found", id)
}
