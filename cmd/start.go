/*
Copyright © 2023 Raymond onepiece010938@gmail.com
*/
package cmd

import (
	"errors"
	"go-line-bot/cmd/bot"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long:  `Start the linebot server and init with your channel secret & channel access token`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return configCheck(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		bot.ServerStart()
		return nil

	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	// cobra.OnInitialize(initToken)
	startCmd.PersistentFlags().StringP("secret", "s", "", "Line Channel Secret")
	startCmd.PersistentFlags().StringP("token", "t", "", "Line Channel Access Token")
	// startCmd.MarkPersistentFlagRequired("secret")
	// startCmd.MarkPersistentFlagRequired("token")

}

func configCheck(cmd *cobra.Command) error {
	// Get flag
	secert, err := cmd.PersistentFlags().GetString("secret")
	if err != nil {
		return err
	}
	token, err := cmd.PersistentFlags().GetString("token")
	if err != nil {
		return err
	}
	if secert != "" {
		// get falg save to viper
		viper.Set("CHANNEL_SECRET", secert)
	}
	if token != "" {
		// get falg save to viper
		viper.Set("CHANNEL_ACCESS_TOKEN", token)
	}
	// Both flag and ENV didn't set value
	if viper.Get("CHANNEL_SECRET") == nil {
		return errors.New("line channel secret is not set")
	}
	if viper.Get("CHANNEL_ACCESS_TOKEN") == nil {
		return errors.New("line channel access token is not set")
	}
	return nil
}
