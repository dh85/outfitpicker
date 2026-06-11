package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchitectureGuardrails_CLIProductionCodeAvoidsInfrastructureImports(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}

	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		contents, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", name, err)
		}

		if strings.Contains(string(contents), "internal/infrastructure/") {
			t.Fatalf("%s imports infrastructure directly; move wiring to cmd/outfitpicker", name)
		}
	}
}

func TestArchitectureGuardrails_RuntimeInterfacesStayNarrow(t *testing.T) {
	contents, err := os.ReadFile("runtime_interfaces.go")
	if err != nil {
		t.Fatalf("ReadFile(runtime_interfaces.go) error = %v", err)
	}

	if strings.Contains(string(contents), "type Picker interface") {
		t.Fatal("runtime_interfaces.go reintroduced Picker; wire MenuSystem from explicit dependencies instead")
	}
}
