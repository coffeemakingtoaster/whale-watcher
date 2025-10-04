package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func AddConfigFlags(cmd *cobra.Command, prefix string, val any, envPrefix string) error {
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

		envKey := field.Tag.Get("env")
		structEnvPrefix := field.Tag.Get("envPrefix")

		// Combine prefixes (for nested structs)
		fullKey := mapKey
		if prefix != "" {
			fullKey = prefix + "." + mapKey
		}

		// Compute environment variable prefix for nested structs
		fullEnvPrefix := envPrefix
		if structEnvPrefix != "" {
			fullEnvPrefix = fullEnvPrefix + structEnvPrefix
		}

		// Build environment variable name
		envVar := ""
		if envKey != "" {
			envVar = fullEnvPrefix + envKey
		} else {
			// fallback: derive from key
			envVar = fullEnvPrefix + strings.ToUpper(strings.ReplaceAll(mapKey, ".", "_"))
		}

		// Normalize to safe format (e.g. MYAPP_TARGET_IMAGE)
		envVar = strings.ToUpper(strings.ReplaceAll(envVar, ".", "_"))

		description := field.Tag.Get("desc")

		switch fieldVal.Kind() {
		case reflect.Struct:
			// Recursively register nested struct fields
			if err := AddConfigFlags(cmd, fullKey, fieldVal.Addr().Interface(), fullEnvPrefix); err != nil {
				return err
			}
			continue

		case reflect.String:
			cmd.Flags().String(fullKey, "", fmt.Sprintf("%s - (string) Env: %s", description, envVar))
		case reflect.Bool:
			cmd.Flags().Bool(fullKey, false, fmt.Sprintf("%s - (bool) Env: %s", description, envVar))
		case reflect.Int, reflect.Int32, reflect.Int64:
			cmd.Flags().Int(fullKey, 0, fmt.Sprintf("%s - (int) Env: %s", description, envVar))
		default:
			continue
		}

		// Bind flag to viper
		if err := viper.BindPFlag(fullKey, cmd.Flags().Lookup(fullKey)); err != nil {
			return fmt.Errorf("failed to bind flag %s: %w", fullKey, err)
		}

		// Bind environment variable
		if err := viper.BindEnv(fullKey, envVar); err != nil {
			return fmt.Errorf("failed to bind env var %s for %s: %w", envVar, fullKey, err)
		}
	}
	return nil
}

func AllowsTarget(target string) bool {
	val := viper.GetString("targetlist")
	if len(val) == 0 {
		return true
	}
	return strings.Contains(val, target)

}

func ShouldInteractWithVSC() bool {
	return ValidateGitea() == nil || ValidateGithub() == nil
}
