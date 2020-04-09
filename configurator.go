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

package insconfig

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

const placeholder = "<-key->"

// Params for config parsing
type Params struct {
	// EnvPrefix is a prefix for environment variables
	EnvPrefix string
	// ViperHooks is custom viper decoding hooks
	ViperHooks []mapstructure.DecodeHookFunc
	// ConfigPathGetter should return config path
	ConfigPathGetter ConfigPathGetter
	// FileNotRequired - do not return error on file not found
	FileNotRequired bool
}

// ConfigPathGetter - implement this if you don't want to use config path from --config flag
type ConfigPathGetter interface {
	GetConfigPath() string
}

type insConfigurator struct {
	params Params
	viper  *viper.Viper
}

// New creates new insConfigurator with params
func New(params Params) insConfigurator {
	return insConfigurator{
		params: params,
		viper:  viper.New(),
	}
}

// Load loads configuration from path, env and makes checks
// configStruct is a pointer to your config
func (i *insConfigurator) Load(configStruct interface{}) error {
	if i.params.EnvPrefix == "" {
		return errors.New("EnvPrefix should be defined")
	}
	if i.params.ConfigPathGetter == nil {
		return errors.New("ConfigPathGetter should be defined")
	}

	configPath := i.params.ConfigPathGetter.GetConfigPath()
	return i.load(configPath, configStruct)
}

func (i *insConfigurator) load(path string, configStruct interface{}) error {

	i.viper.AutomaticEnv()
	i.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	i.viper.SetEnvPrefix(i.params.EnvPrefix)

	i.viper.SetConfigFile(path)
	if err := i.viper.ReadInConfig(); err != nil {
		if !i.params.FileNotRequired {
			return err
		}
		fmt.Printf("failed to load config from '%s'\n", path)
	}
	i.params.ViperHooks = append(i.params.ViperHooks, mapstructure.StringToTimeDurationHookFunc(), mapstructure.StringToSliceHookFunc(","))
	err := i.viper.UnmarshalExact(configStruct, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		i.params.ViperHooks...,
	)))
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal config file into configuration structure")
	}
	configStructKeys := deepFieldNames(configStruct, "")
	configStructKeys, mapKeys := separateKeys(configStructKeys)
	configStructKeys, err = i.checkNoExtraENVValues(configStructKeys, mapKeys)
	if err != nil {
		return err
	}

	err = i.checkAllValuesIsSet(configStructKeys)
	if err != nil {
		return err
	}

	// Second Unmarshal needed because of bug https://github.com/spf13/viper/issues/761
	// This should be evaluated after manual values overriding is done
	err = i.viper.UnmarshalExact(configStruct, viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		i.params.ViperHooks...,
	)))
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal config file into configuration structure 2")
	}
	return nil
}

func (i *insConfigurator) checkNoExtraENVValues(structKeys []string, mapKeys []string) ([]string, error) {
	var errorKeys []string
	prefixLen := len(i.params.EnvPrefix)
	for _, e := range os.Environ() {
		if len(e) > prefixLen && e[0:prefixLen]+"_" == strings.ToUpper(i.params.EnvPrefix)+"_" {
			kv := strings.SplitN(e, "=", 2)
			key := strings.ReplaceAll(strings.Replace(strings.ToLower(kv[0]), i.params.EnvPrefix+"_", "", 1), "_", ".")

			if k, pref, match := matchMapKey(mapKeys, key); match && !stringInSlice(key, structKeys) {
				structKeys = append(structKeys, newKeys(mapKeys, k, pref)...)
			}

			if stringInSlice(key, structKeys) {
				// This manually sets value from ENV and overrides everything, this temporarily fix issue https://github.com/spf13/viper/issues/761
				i.viper.Set(key, kv[1])
			} else {
				errorKeys = append(errorKeys, key)
			}
		}
	}
	if len(errorKeys) > 0 {
		return structKeys, errors.New(fmt.Sprintf("Wrong config keys found in ENV: %s", strings.Join(errorKeys, ", ")))
	}
	return structKeys, nil
}

func separateKeys(list []string) (names []string, keys []string) {
	for _, s := range list {
		if strings.Contains(s, placeholder) {
			keys = append(keys, s)
		} else {
			names = append(names, s)
		}
	}
	return names, keys
}

func newKeys(keys []string, key, pref string) []string {
	var names []string
	oldStr := strings.Join([]string{pref, placeholder}, "")
	newStr := strings.Join([]string{pref, key}, "")
	for _, k := range keys {
		if strings.HasPrefix(k, oldStr) {
			names = append(names, strings.Replace(k, oldStr, newStr, 1))
		}
	}
	return names
}

func matchMapKey(keys []string, key string) (string, string, bool) {
	for _, k := range keys {
		l := strings.ToLower(k)
		pattern := strings.ReplaceAll(l, ".", "\\.")
		pattern = strings.Replace(pattern, placeholder, ".+", 1)
		match, err := regexp.MatchString(pattern, key)
		if err != nil {
			fmt.Println(err)
		}
		if match {
			parts := strings.Split(l, placeholder)
			return strings.TrimSuffix(strings.TrimPrefix(key, parts[0]), parts[1]), parts[0], true
		}
	}
	return "", "", false
}

func (i *insConfigurator) checkAllValuesIsSet(cstructKeys []string) error {
	var errorKeys []string
	allKeys := i.viper.AllKeys()
	for _, keyName := range cstructKeys {
		if !i.viper.IsSet(keyName) {
			// Due to a bug https://github.com/spf13/viper/issues/447 we can't use InConfig, so
			if !stringInSlice(keyName, allKeys) && !strings.Contains(keyName, placeholder) {
				errorKeys = append(errorKeys, keyName)
			}
			// Value of this key is "null" but it's set in config file
		}
	}
	if len(errorKeys) > 0 {
		return errors.New(fmt.Sprintf("Keys is not defined in config: %s", strings.Join(errorKeys, ", ")))
	}
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.ToLower(b) == strings.ToLower(a) {
			return true
		}
	}
	return false
}

func deepFieldNames(iface interface{}, prefix string) []string {
	names := make([]string, 0)
	v := reflect.ValueOf(iface)
	ifv := reflect.Indirect(v)
	s := ifv.Type()

	for i := 0; i < s.NumField(); i++ {
		v := ifv.Field(i)
		tagValue := ifv.Type().Field(i).Tag.Get("mapstructure")
		tagParts := strings.Split(tagValue, ",")

		// If "squash" is specified in the tag, we squash the field down.
		squash := false
		for _, tag := range tagParts[1:] {
			if tag == "squash" {
				squash = true
				break
			}
		}

		switch v.Kind() {
		case reflect.Struct:
			newPrefix := ""
			currPrefix := ""
			if !squash {
				currPrefix = ifv.Type().Field(i).Name
			}
			if prefix != "" {
				newPrefix = strings.Join([]string{prefix, currPrefix}, ".")
			} else {
				newPrefix = currPrefix
			}

			names = append(names, deepFieldNames(v.Interface(), strings.ToLower(newPrefix))...)
		case reflect.Map:
			if len(v.MapKeys()) != 0 {
				for _, k := range v.MapKeys() {
					key := k.String()
					newPrefix := ""
					currPrefix := ifv.Type().Field(i).Name
					if prefix != "" {
						newPrefix = strings.Join([]string{prefix, currPrefix, key}, ".")
					} else {
						newPrefix = strings.Join([]string{currPrefix, key}, ".")
					}
					names = append(names, deepFieldNames(v.MapIndex(k).Interface(), strings.ToLower(newPrefix))...)
				}
			} else {
				newPrefix := ""
				currPrefix := ifv.Type().Field(i).Name
				if prefix != "" {
					newPrefix = strings.Join([]string{prefix, currPrefix, placeholder}, ".")
				} else {
					newPrefix = strings.Join([]string{currPrefix, placeholder}, ".")
				}
				e := v.Type().Elem()
				names = append(names, deep(e, strings.ToLower(newPrefix))...)
			}
		default:
			prefWithPoint := ""
			if prefix != "" {
				prefWithPoint = prefix + "."
			}
			names = append(names, strings.ToLower(prefWithPoint+ifv.Type().Field(i).Name))
		}
	}

	return names
}

func deep(t reflect.Type, prefix string) []string {
	names := make([]string, 0)

	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			tf := t.Field(i)

			var newPref string
			if prefix != "" {
				newPref = strings.Join([]string{prefix, tf.Name}, ".")
			} else {
				newPref = tf.Name
			}

			z := reflect.Zero(tf.Type)
			names = append(names, deep(z.Type(), strings.ToLower(newPref))...)
		}
	default:
		if prefix != "" {
			names = append(names, strings.ToLower(prefix))
		}
	}
	return names
}

// ToYaml returns yaml marshalled struct
func (i *insConfigurator) ToYaml(c interface{}) string {
	// todo clean password
	out, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("failed to marshal config structure: %v", err)
	}
	return string(out)
}
