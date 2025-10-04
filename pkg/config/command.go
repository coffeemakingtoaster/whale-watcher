package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCommand() *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "config",
		Short: "Get the config",
		Long: ` Get the current config. Mainly meant to verify precendence behaviour

Expected arguments:  <policy set location> <Dockerfile location> [<oci tar location>] [<docker tar location>]
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg Config
			// 5️⃣ Unmarshal merged config into struct
			if err := viper.Unmarshal(&cfg); err != nil {
				return err
			}
			fmt.Printf("Final merged config: %+v\n", cfg)
			return nil
		},
	}
	return cmd
}
