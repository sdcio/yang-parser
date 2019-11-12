// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile

import (
	"os"
)

type FeatureStatus int

const (
	DISABLED FeatureStatus = iota
	ENABLED
	NOTPRESENT
)

func (f FeatureStatus) String() string {
	switch f {
	case DISABLED:
		return "Disabled"
	case ENABLED:
		return "Enabled"
	case NOTPRESENT:
		return "Not Present"
	default:
		return "unknown"
	}
}
func StatusFromBool(b bool) FeatureStatus {
	if b == true {
		return ENABLED
	}
	return DISABLED
}

// FeaturesChecker checks the status of a feature
// feature is formatted <module>:<feature name>
type FeaturesChecker interface {
	Status(feature string) FeatureStatus
}

type featuresMap struct {
	features map[string]bool
}

func newFeaturesMap() featuresMap {
	return featuresMap{features: make(map[string]bool)}
}

func (f *featuresMap) Status(feature string) FeatureStatus {
	if enabled, ok := f.features[feature]; ok {
		return StatusFromBool(enabled)
	}
	return NOTPRESENT
}

// set is used to enable or disable a feature
func (f *featuresMap) set(feature string, enabled bool) {
	f.features[feature] = enabled
}

// getFeatures add features to a featuresMap
//
// If the file location/yang_module_name/feature_name exists, then the
// feature is enabled, otherwise, its disabled.
func (f *featuresMap) getFeatures(location string, enable bool) {
	if location == "" {
		// None defined
		return
	}
	fi, err := os.Stat(location)
	if err != nil {
		//  features does not exist
		return
	}

	if fi.Mode().IsDir() {
		d, err := os.Open(location)
		if err != nil {
			return
		}
		defer d.Close()

		names, err := d.Readdir(0)
		if err != nil {
			return
		}

		for _, name := range names {
			if name.IsDir() {
				featDir, err := os.Open(location + "/" + name.Name())
				features, err := featDir.Readdir(0)
				if err != nil {
					// Skip any problematic directories
					continue
				}
				for _, feat := range features {
					if !feat.IsDir() {
						f.features[name.Name()+":"+feat.Name()] = enable
					}
				}
				featDir.Close()
			}
		}
	}
}

func FeaturesFromLocations(enabled bool, locations ...string) FeaturesChecker {
	f := make(map[string]bool)
	m := featuresMap{features: f}
	for _, l := range locations {
		m.getFeatures(l, enabled)
	}

	return &m
}

func FeaturesFromNames(enabled bool, features ...string) FeaturesChecker {
	f := make(map[string]bool)
	m := featuresMap{features: f}
	for _, n := range features {
		m.set(n, enabled)
	}
	return &m
}

type checkers struct {
	checkers []FeaturesChecker
}

func MultiFeatureCheckers(f ...FeaturesChecker) FeaturesChecker {
	return &checkers{checkers: f}
}

func (f *checkers) Status(feature string) FeatureStatus {
	// Check each FeaturesChecker in order, the last to report
	// Enabled or Disabled wins. Disabled is reported if
	// not found.
	var status FeatureStatus = DISABLED
	for _, chkr := range f.checkers {
		if chkr == nil {
			continue
		}
		s := chkr.Status(feature)
		if s != NOTPRESENT {
			status = s
		}
	}

	return status
}
