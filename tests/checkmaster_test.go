package testing

import (
	"axfrcheck/pkg"
	"fmt"
	"testing"
)

// TestCheckMasters tests the CheckMasters function for a basic scenario.
func TestCheckMasters(t *testing.T) {
	// Setup: Define a path to a mock configuration file or use a mock library
	configFile := "../example/named.conf"
	zones, err := pkg.ParseNamedConf(configFile)
	if err != nil {
		fmt.Printf("Could not parse %s", configFile)
	}

	// Action: Call CheckMasters with the mock config file path
	err = pkg.CheckMasters(zones)
	if err != nil {
		t.Errorf("CheckMasters() with mock config returned an error: %v", err)
	}
}
