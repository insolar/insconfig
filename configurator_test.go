//
// Copyright 2019 Insolar Technologies GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package insconfig_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/insolar/insconfig"
)

type Level3 struct {
	Level3text string
	NullString *string
}
type Level2 struct {
	Level2text string
	Level3     Level3
}
type CfgStruct struct {
	Level1text string
	Level2     Level2
	MapField   map[string]Level2
	Map2       map[string]Level3
}

type anonymousEmbeddedStruct struct {
	CfgStruct `mapstructure:",squash"`
	Level4    string
}

type testPathGetter struct {
	Path string
}

func (g testPathGetter) GetConfigPath() string {
	return g.Path
}

func Test_Load(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		cfg := CfgStruct{}
		params := insconfig.Params{
			EnvPrefix:        "testprefix",
			ConfigPathGetter: testPathGetter{"test_config.yaml"},
		}

		insConfigurator := insconfig.New(params)
		err := insConfigurator.Load(&cfg)
		require.NoError(t, err)
		require.Equal(t, cfg.Level1text, "text1")
		require.Equal(t, cfg.Level2.Level2text, "text2")
		require.Equal(t, cfg.Level2.Level3.Level3text, "text3")
		require.Len(t, cfg.MapField, 2)
		key1 := cfg.MapField["key1"]
		require.Equal(t, key1.Level2text, "key1text2")
		require.Equal(t, key1.Level3.Level3text, "key1text3")
		require.Nil(t, key1.Level3.NullString)
		key2 := cfg.MapField["key2"]
		require.Equal(t, key2.Level2text, "key2text2")
		require.Equal(t, key2.Level3.Level3text, "key2text3")
		require.NotNil(t, key2.Level3.NullString)
	})

	t.Run("ENV overriding", func(t *testing.T) {
		_ = os.Setenv("TESTPREFIX_LEVEL2_LEVEL2TEXT", "newTextValue")
		defer os.Unsetenv("TESTPREFIX_LEVEL2_LEVEL2TEXT")
		cfg := CfgStruct{}
		params := insconfig.Params{
			EnvPrefix:        "testprefix",
			ConfigPathGetter: testPathGetter{"test_config.yaml"},
		}

		insConfigurator := insconfig.New(params)
		err := insConfigurator.Load(&cfg)
		require.NoError(t, err)
		require.Equal(t, cfg.Level1text, "text1")
		require.Equal(t, cfg.Level2.Level2text, "newTextValue")
		require.Equal(t, cfg.Level2.Level3.Level3text, "text3")
	})

	t.Run("ENV has values, that is not in config, but it should", func(t *testing.T) {
		_ = os.Setenv("TESTPREFIX_LEVEL1TEXT", "newTextValue1")
		_ = os.Setenv("TESTPREFIX_MAP2_ONE_LEVEL3TEXT", "newTextValue1")
		_ = os.Setenv("TESTPREFIX_MAP2_ONE_NULLSTRING", "newTextValue1")
		defer os.Unsetenv("TESTPREFIX_LEVEL1TEXT")
		defer os.Unsetenv("TESTPREFIX_MAP2_ONE_LEVEL3TEXT")
		defer os.Unsetenv("TESTPREFIX_MAP2_ONE_NULLSTRING")
		cfg := CfgStruct{}
		params := insconfig.Params{
			EnvPrefix:        "testprefix",
			ConfigPathGetter: testPathGetter{"test_config_wrong2.yaml"},
		}

		insConfigurator := insconfig.New(params)
		err := insConfigurator.Load(&cfg)
		require.NoError(t, err)
		require.Equal(t, cfg.Level1text, "newTextValue1")
		require.Equal(t, cfg.Level2.Level2text, "text2")
		require.Equal(t, cfg.Level2.Level3.Level3text, "text3")
	})

	t.Run("ENV only, no config files", func(t *testing.T) {
		_ = os.Setenv("TESTPREFIX_LEVEL1TEXT", "newTextValue1")
		_ = os.Setenv("TESTPREFIX_LEVEL2_LEVEL2TEXT", "newTextValue2")
		_ = os.Setenv("TESTPREFIX_LEVEL2_LEVEL3_LEVEL3TEXT", "newTextValue3")
		_ = os.Setenv("TESTPREFIX_LEVEL2_LEVEL3_NULLSTRING", "text")
		_ = os.Setenv("TESTPREFIX_MAPFIELD_KEY1_LEVEL2TEXT", "1")
		_ = os.Setenv("TESTPREFIX_MAPFIELD_KEY1_LEVEL3_LEVEL3TEXT", "2")
		_ = os.Setenv("TESTPREFIX_MAPFIELD_KEY1_LEVEL3_NULLSTRING", "3")
		_ = os.Setenv("TESTPREFIX_MAPFIELD_KEY2_LEVEL2TEXT", "21")
		_ = os.Setenv("TESTPREFIX_MAPFIELD_KEY2_LEVEL3_LEVEL3TEXT", "22")
		_ = os.Setenv("TESTPREFIX_MAPFIELD_KEY2_LEVEL3_NULLSTRING", "23")
		_ = os.Setenv("TESTPREFIX_MAP2_KEY3_LEVEL3TEXT", "32")
		_ = os.Setenv("TESTPREFIX_MAP2_KEY3_NULLSTRING", "33")
		defer os.Unsetenv("TESTPREFIX_LEVEL1TEXT")
		defer os.Unsetenv("TESTPREFIX_LEVEL2_LEVEL2TEXT")
		defer os.Unsetenv("TESTPREFIX_LEVEL2_LEVEL3_LEVEL3TEXT")
		defer os.Unsetenv("TESTPREFIX_LEVEL2_LEVEL3_NULLSTRING")
		defer os.Unsetenv("TESTPREFIX_MAPFIELD_KEY1_LEVEL2TEXT")
		defer os.Unsetenv("TESTPREFIX_MAPFIELD_KEY1_LEVEL3_LEVEL3TEXT")
		defer os.Unsetenv("TESTPREFIX_MAPFIELD_KEY1_LEVEL3_NULLSTRING")
		defer os.Unsetenv("TESTPREFIX_MAPFIELD_KEY2_LEVEL2TEXT")
		defer os.Unsetenv("TESTPREFIX_MAPFIELD_KEY2_LEVEL3_LEVEL3TEXT")
		defer os.Unsetenv("TESTPREFIX_MAPFIELD_KEY2_LEVEL3_NULLSTRING")
		defer os.Unsetenv("TESTPREFIX_MAP2_KEY3_LEVEL3TEXT")
		defer os.Unsetenv("TESTPREFIX_MAP2_KEY3_NULLSTRING")

		cfg := CfgStruct{}
		params := insconfig.Params{
			EnvPrefix:        "testprefix",
			ConfigPathGetter: testPathGetter{""},
			FileNotRequired:  true,
		}

		insConfigurator := insconfig.New(params)
		err := insConfigurator.Load(&cfg)
		require.NoError(t, err)
		require.Equal(t, "newTextValue1", cfg.Level1text)
		require.Equal(t, "newTextValue2", cfg.Level2.Level2text)
		require.Equal(t, "newTextValue3", cfg.Level2.Level3.Level3text)
		mapField := cfg.MapField
		require.Len(t, mapField, 2)
		require.Equal(t, "1", mapField["key1"].Level2text)
		require.Equal(t, "2", mapField["key1"].Level3.Level3text)
		require.Equal(t, "3", *mapField["key1"].Level3.NullString)
		require.Equal(t, "21", mapField["key2"].Level2text)
		require.Equal(t, "22", mapField["key2"].Level3.Level3text)
		require.Equal(t, "23", *mapField["key2"].Level3.NullString)
		map2 := cfg.Map2
		require.Len(t, map2, 1)
		require.Equal(t, "32", map2["key3"].Level3text)
		require.Equal(t, "33", *map2["key3"].NullString)
	})

	t.Run("ENV only, not enough keys fail", func(t *testing.T) {
		_ = os.Setenv("TESTPREFIX_LEVEL1TEXT", "newTextValue1")
		_ = os.Setenv("TESTPREFIX_LEVEL2_LEVEL3_LEVEL3TEXT", "newTextValue3")
		_ = os.Setenv("TESTPREFIX_LEVEL2_LEVEL3_NULLSTRING", "text")
		_ = os.Setenv("TESTPREFIX_MAPFIELD_KEY1_LEVEL2TEXT", `1`)
		_ = os.Setenv("TESTPREFIX_MAPFIELD_KEY1_LEVEL3_NULLSTRING", `3`)
		_ = os.Setenv("TESTPREFIX_MAPFIELD_KEY3_LEVEL2TEXT", `1`)
		_ = os.Setenv("TESTPREFIX_MAPFIELD_KEY2_LEVEL2TEXT", `1`)
		_ = os.Setenv("TESTPREFIX_MAPFIELD_KEY2_LEVEL3_LEVEL3TEXT", `2`)
		_ = os.Setenv("TESTPREFIX_MAPFIELD_KEY2_LEVEL3_NULLSTRING", `3`)
		defer os.Unsetenv("TESTPREFIX_LEVEL1TEXT")
		defer os.Unsetenv("TESTPREFIX_LEVEL2_LEVEL3_LEVEL3TEXT")
		defer os.Unsetenv("TESTPREFIX_LEVEL2_LEVEL3_NULLSTRING")
		defer os.Unsetenv("TESTPREFIX_MAPFIELD_KEY1_LEVEL2TEXT")
		defer os.Unsetenv("TESTPREFIX_MAPFIELD_KEY1_LEVEL3_NULLSTRING")
		defer os.Unsetenv("TESTPREFIX_MAPFIELD_KEY3_LEVEL2TEXT")
		defer os.Unsetenv("TESTPREFIX_MAPFIELD_KEY2_LEVEL2TEXT")
		defer os.Unsetenv("TESTPREFIX_MAPFIELD_KEY2_LEVEL3_LEVEL3TEXT")
		defer os.Unsetenv("TESTPREFIX_MAPFIELD_KEY2_LEVEL3_NULLSTRING")

		cfg := CfgStruct{}
		params := insconfig.Params{
			EnvPrefix:        "testprefix",
			ConfigPathGetter: testPathGetter{""},
			FileNotRequired:  true,
		}

		insConfigurator := insconfig.New(params)
		err := insConfigurator.Load(&cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "level2.level2text")
		require.Contains(t, err.Error(), "mapfield.key1.level3.level3text")
		require.Contains(t, err.Error(), "mapfield.key3.level3.level3text")
		require.Contains(t, err.Error(), "mapfield.key3.level3.nullstring")
		require.Contains(t, err.Error(), "map2.<-key->.nullstring")
		require.Contains(t, err.Error(), "map2.<-key->.level3text")
	})

	t.Run("extra env fail", func(t *testing.T) {
		_ = os.Setenv("TESTPREFIX_NONEXISTENT_VALUE", "123")
		defer os.Unsetenv("TESTPREFIX_NONEXISTENT_VALUE")

		cfg := CfgStruct{}
		params := insconfig.Params{
			EnvPrefix:        "testprefix",
			ConfigPathGetter: testPathGetter{"test_config.yaml"},
		}

		insConfigurator := insconfig.New(params)
		err := insConfigurator.Load(&cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "nonexistent")
	})

	t.Run("extra env with empty value, fails", func(t *testing.T) {
		_ = os.Setenv("TESTPREFIX_NONEXISTENT_VALUE1", "")
		_ = os.Setenv("TESTPREFIX_NONEXISTENT_VALUE2", "")
		defer os.Unsetenv("TESTPREFIX_NONEXISTENT_VALUE1")
		defer os.Unsetenv("TESTPREFIX_NONEXISTENT_VALUE2")

		cfg := CfgStruct{}
		params := insconfig.Params{
			EnvPrefix:        "testprefix",
			ConfigPathGetter: testPathGetter{"test_config.yaml"},
		}

		insConfigurator := insconfig.New(params)
		err := insConfigurator.Load(&cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "nonexistent.value1")
		require.Contains(t, err.Error(), "nonexistent.value2")
	})

	t.Run("extra in file fail", func(t *testing.T) {
		cfg := CfgStruct{}
		params := insconfig.Params{
			EnvPrefix:        "testprefix",
			ConfigPathGetter: testPathGetter{"test_config_wrong.yaml"},
		}

		insConfigurator := insconfig.New(params)
		err := insConfigurator.Load(&cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "nonexistent")
	})

	t.Run("not set in file fail", func(t *testing.T) {
		cfg := CfgStruct{}
		params := insconfig.Params{
			EnvPrefix:        "testprefix",
			ConfigPathGetter: testPathGetter{"test_config_wrong2.yaml"},
		}

		insConfigurator := insconfig.New(params)
		err := insConfigurator.Load(&cfg)
		require.Error(t, err)

		require.Contains(t, err.Error(), "level1text")
		require.Contains(t, err.Error(), "map2.<-key->.level3text")
		require.Contains(t, err.Error(), "map2.<-key->.nullstring")
	})

	t.Run("required file not found", func(t *testing.T) {
		cfg := CfgStruct{}
		params := insconfig.Params{
			EnvPrefix:        "testprefix",
			ConfigPathGetter: testPathGetter{"nonexistent.yaml"},
		}

		insConfigurator := insconfig.New(params)
		err := insConfigurator.Load(&cfg)
		require.Error(t, err)
		require.Contains(t, err.Error(), "nonexistent.yaml")
	})

	t.Run("null string test", func(t *testing.T) {
		cfg := CfgStruct{}
		params := insconfig.Params{
			EnvPrefix:        "testprefix",
			ConfigPathGetter: testPathGetter{"test_config2.yaml"},
		}

		insConfigurator := insconfig.New(params)
		err := insConfigurator.Load(&cfg)
		require.NoError(t, err)
		require.Nil(t, cfg.Level2.Level3.NullString)
	})

	t.Run("embedded struct flatten test", func(t *testing.T) {
		cfg := anonymousEmbeddedStruct{}
		params := insconfig.Params{
			EnvPrefix:        "testprefix",
			ConfigPathGetter: testPathGetter{"test_config3.yaml"},
		}

		insConfigurator := insconfig.New(params)
		err := insConfigurator.Load(&cfg)
		require.NoError(t, err)
		require.Equal(t, cfg.Level4, "text4")
	})

	t.Run("embedded struct override by env", func(t *testing.T) {
		_ = os.Setenv("TESTPREFIX_LEVEL2_LEVEL2TEXT", "newTextValue")
		defer os.Unsetenv("TESTPREFIX_LEVEL2_LEVEL2TEXT")

		cfg := anonymousEmbeddedStruct{}
		params := insconfig.Params{
			EnvPrefix:        "testprefix",
			ConfigPathGetter: testPathGetter{"test_config3.yaml"},
		}

		insConfigurator := insconfig.New(params)
		err := insConfigurator.Load(&cfg)
		require.NoError(t, err)
		require.Equal(t, cfg.Level2.Level2text, "newTextValue")
	})
}
