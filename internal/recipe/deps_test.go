package recipe

import (
	"reflect"
	"strings"
	"testing"
)

func TestResolveDependenciesNoDeps(t *testing.T) {
	recipes := map[string]Recipe{
		"build": {Command: "go build"},
	}
	got, err := ResolveDependencies(recipes, "build")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"build"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestResolveDependenciesLinearChain(t *testing.T) {
	recipes := map[string]Recipe{
		"deploy": {Command: "deploy.sh", Depends: []string{"build"}},
		"build":  {Command: "go build", Depends: []string{"lint"}},
		"lint":   {Command: "go vet"},
	}
	got, err := ResolveDependencies(recipes, "deploy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"lint", "build", "deploy"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestResolveDependenciesDeclarationOrder(t *testing.T) {
	recipes := map[string]Recipe{
		"deploy": {Command: "deploy.sh", Depends: []string{"lint", "build"}},
		"build":  {Command: "go build"},
		"lint":   {Command: "go vet"},
	}
	got, err := ResolveDependencies(recipes, "deploy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"lint", "build", "deploy"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestResolveDependenciesDiamondRunsOnceInFirstSeenPosition(t *testing.T) {
	// deploy -> [build, test], build -> [lint], test -> [lint]
	// lint must appear exactly once, before its first dependent (build).
	recipes := map[string]Recipe{
		"deploy": {Command: "deploy.sh", Depends: []string{"build", "test"}},
		"build":  {Command: "go build", Depends: []string{"lint"}},
		"test":   {Command: "go test", Depends: []string{"lint"}},
		"lint":   {Command: "go vet"},
	}
	got, err := ResolveDependencies(recipes, "deploy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"lint", "build", "test", "deploy"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

func TestResolveDependenciesUnknownRootErrors(t *testing.T) {
	_, err := ResolveDependencies(map[string]Recipe{}, "nope")
	if err == nil || !strings.Contains(err.Error(), "no such recipe: nope") {
		t.Fatalf("error = %v, want mention of 'no such recipe: nope'", err)
	}
}

func TestResolveDependenciesUnknownDependencyErrors(t *testing.T) {
	recipes := map[string]Recipe{
		"deploy": {Command: "deploy.sh", Depends: []string{"missing"}},
	}
	_, err := ResolveDependencies(recipes, "deploy")
	if err == nil || !strings.Contains(err.Error(), "deploy depends on missing, which does not exist") {
		t.Fatalf("error = %v, want mention of the missing dependency", err)
	}
}

func TestResolveDependenciesDirectCycleErrors(t *testing.T) {
	recipes := map[string]Recipe{
		"a": {Command: "echo a", Depends: []string{"b"}},
		"b": {Command: "echo b", Depends: []string{"a"}},
	}
	_, err := ResolveDependencies(recipes, "a")
	if err == nil || !strings.Contains(err.Error(), "dependency cycle") {
		t.Fatalf("error = %v, want mention of a dependency cycle", err)
	}
}

func TestResolveDependenciesSelfCycleErrors(t *testing.T) {
	recipes := map[string]Recipe{
		"a": {Command: "echo a", Depends: []string{"a"}},
	}
	_, err := ResolveDependencies(recipes, "a")
	if err == nil || !strings.Contains(err.Error(), "dependency cycle") {
		t.Fatalf("error = %v, want mention of a dependency cycle", err)
	}
}

func TestResolveDependenciesDependencyWithRequiredParamsErrors(t *testing.T) {
	recipes := map[string]Recipe{
		"deploy": {Command: "deploy.sh {{env}}", Params: []string{"env"}, Depends: nil},
		"build":  {Command: "deploy.sh", Depends: []string{"deploy"}},
	}
	_, err := ResolveDependencies(recipes, "build")
	if err == nil || !strings.Contains(err.Error(), "deploy has required params") {
		t.Fatalf("error = %v, want mention that deploy has required params", err)
	}
}

func TestResolveDependenciesRootWithRequiredParamsIsFine(t *testing.T) {
	// The top-level recipe being invoked directly is allowed to have
	// required params — only recipes reached purely as a dependency are
	// restricted.
	recipes := map[string]Recipe{
		"open":  {Command: "code {{file}}", Params: []string{"file"}, Depends: []string{"build"}},
		"build": {Command: "go build"},
	}
	got, err := ResolveDependencies(recipes, "open")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"build", "open"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}
