package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// AddConfigFlagsWithGroups registers flags for each nested struct
// as a separate flag group, and binds both Viper + environment variables.
func AddConfigFlagsWithGroups(cmd *cobra.Command, prefix string, val any, envPrefix string) error {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		if !fieldVal.CanInterface() {
			continue
		}

		mapKey := field.Tag.Get("mapstructure")
		if mapKey == "" {
			mapKey = strings.ToLower(field.Name)
		}
		structEnvPrefix := field.Tag.Get("envPrefix")
		fullEnvPrefix := envPrefix
		if structEnvPrefix != "" {
			fullEnvPrefix = envPrefix + structEnvPrefix
		}

		switch fieldVal.Kind() {
		case reflect.Struct:
			groupName := field.Tag.Get("group")

			var err error

			if len(groupName) == 0 {
				err = AddConfigFlagsWithGroups(cmd, mapKey, fieldVal.Addr().Interface(), fullEnvPrefix)
			} else {
				groupFlags := pflag.NewFlagSet(groupName, pflag.ContinueOnError)
				err = addStructFlags(groupFlags, mapKey, fieldVal.Addr().Interface(), fullEnvPrefix)
				cmd.PersistentFlags().AddFlagSet(groupFlags)
			}
			if err != nil {
				return err
			}

		default:
			// top-level (non-struct) fields
			if err := addFieldFlag(cmd.PersistentFlags(), prefix, &field, &fieldVal, envPrefix); err != nil {
				return err
			}
		}
	}

	return nil
}

func addStructFlags(flagSet *pflag.FlagSet, prefix string, val any, envPrefix string) error {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		if !fieldVal.CanInterface() {
			continue
		}

		mapKey := field.Tag.Get("mapstructure")
		if mapKey == "" {
			mapKey = strings.ToLower(field.Name)
		}

		fullKey := prefix + "." + mapKey
		envKey := field.Tag.Get("env")
		if envKey == "" {
			envKey = strings.ToUpper(strings.ReplaceAll(mapKey, ".", "_"))
		}
		envVar := envPrefix + envKey
		desc := field.Tag.Get("desc")

		// Register flag
		switch fieldVal.Kind() {
		case reflect.String:
			flagSet.String(fullKey, "", fmt.Sprintf("%s - (string) Env: %s", desc, envVar))
		case reflect.Bool:
			flagSet.Bool(fullKey, false, fmt.Sprintf("%s - (bool) Env: %s", desc, envVar))
		case reflect.Int, reflect.Int32, reflect.Int64:
			flagSet.Int(fullKey, 0, fmt.Sprintf("%s - (int) Env: %s", desc, envVar))
		}

		flagSet.SetAnnotation(fullKey, "group", []string{flagSet.Name()})

		// Bind to Viper and environment
		_ = viper.BindPFlag(fullKey, flagSet.Lookup(fullKey))
		_ = viper.BindEnv(fullKey, envVar)
	}

	return nil
}

func addFieldFlag(flagSet *pflag.FlagSet, prefix string, field *reflect.StructField, value *reflect.Value, envPrefix string) error {
	mapKey := field.Tag.Get("mapstructure")
	if mapKey == "" {
		mapKey = strings.ToLower(field.Name)
	}
	fullKey := mapKey
	if prefix != "" {
		fullKey = prefix + "." + mapKey
	}

	envKey := field.Tag.Get("env")
	if envKey == "" {
		envKey = strings.ToUpper(strings.ReplaceAll(mapKey, ".", "_"))
	}
	envVar := envPrefix + envKey

	desc := field.Tag.Get("desc")

	switch value.Kind() {
	case reflect.String:
		flagSet.String(fullKey, "", fmt.Sprintf("%s - (string) Env: %s", desc, envVar))
	case reflect.Bool:
		flagSet.Bool(fullKey, false, fmt.Sprintf("%s - (bool) Env: %s", desc, envVar))
	case reflect.Int, reflect.Int32, reflect.Int64:
		flagSet.Int(fullKey, 0, fmt.Sprintf("%s - (int) Env: %s", desc, envVar))
	}

	_ = viper.BindPFlag(fullKey, flagSet.Lookup(fullKey))
	_ = viper.BindEnv(fullKey, envVar)

	return nil
}

func AllowsTarget(target string) bool {
	allowList := viper.GetString("target_list")
	log.Info().Str("allow", allowList).Send()
	if len(allowList) == 0 {
		return true
	}
	return strings.Contains(allowList, target)

}

func ShouldInteractWithVSC() bool {
	return ValidateGitea() == nil || ValidateGithub() == nil
}
