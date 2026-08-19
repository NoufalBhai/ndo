package config

import (
	"reflect"
	"testing"

	"github.com/green-threads/ndo/internal/recipe"
)

func TestMerge(t *testing.T) {
	central := RecipeFile{Recipes: map[string]recipe.Recipe{
		"test":   {Command: "go test ./..."},
		"deploy": {Command: "./deploy.sh central"},
	}}
	local := RecipeFile{Recipes: map[string]recipe.Recipe{
		"deploy": {Command: "./deploy.sh local"},
		"build":  {Command: "go build ./..."},
	}}

	tests := []struct {
		name           string
		central, local RecipeFile
		want           map[string]recipe.Recipe
	}{
		{
			name:    "empty central",
			central: RecipeFile{Recipes: map[string]recipe.Recipe{}},
			local:   local,
			want:    local.Recipes,
		},
		{
			name:    "empty local",
			central: central,
			local:   RecipeFile{Recipes: map[string]recipe.Recipe{}},
			want:    central.Recipes,
		},
		{
			name:    "disjoint names",
			central: RecipeFile{Recipes: map[string]recipe.Recipe{"test": {Command: "go test"}}},
			local:   RecipeFile{Recipes: map[string]recipe.Recipe{"build": {Command: "go build"}}},
			want: map[string]recipe.Recipe{
				"test":  {Command: "go test"},
				"build": {Command: "go build"},
			},
		},
		{
			name:    "same-name collision: local wins wholesale",
			central: central,
			local:   local,
			want: map[string]recipe.Recipe{
				"test":   {Command: "go test ./..."},
				"deploy": {Command: "./deploy.sh local"},
				"build":  {Command: "go build ./..."},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Merge(tt.central, tt.local)
			if !reflect.DeepEqual(got.Recipes, tt.want) {
				t.Fatalf("Merge() = %+v, want %+v", got.Recipes, tt.want)
			}
		})
	}
}

func TestMergeVarsKeyLevel(t *testing.T) {
	central := RecipeFile{Vars: map[string]map[string]string{
		"folder": {"work": "C:\\Users\\alex\\projects", "home": "C:\\Users\\me"},
	}}
	local := RecipeFile{Vars: map[string]map[string]string{
		"folder": {"work": "C:\\Users\\alex\\projects\\ndo", "proj": "D:\\proj"},
	}}

	got := Merge(central, local)
	want := map[string]map[string]string{
		"folder": {
			"work": "C:\\Users\\alex\\projects\\ndo", // local overrides this key
			"home": "C:\\Users\\me",                  // central-only key survives
			"proj": "D:\\proj",                       // local-only key added
		},
	}
	if !reflect.DeepEqual(got.Vars, want) {
		t.Fatalf("Merge().Vars = %+v, want %+v", got.Vars, want)
	}
}

func TestMergeVarsGroupOnlyInOneSide(t *testing.T) {
	central := RecipeFile{Vars: map[string]map[string]string{
		"folder": {"work": "C:\\Users\\alex\\projects"},
	}}
	local := RecipeFile{Vars: map[string]map[string]string{}}

	got := Merge(central, local)
	want := map[string]map[string]string{"folder": {"work": "C:\\Users\\alex\\projects"}}
	if !reflect.DeepEqual(got.Vars, want) {
		t.Fatalf("Merge().Vars = %+v, want %+v", got.Vars, want)
	}
}
