package config

import (
	"reflect"
	"testing"

	"github.com/NoufalBhai/ndo/internal/recipe"
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
