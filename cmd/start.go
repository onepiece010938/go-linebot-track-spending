/*
Copyright © 2023 Raymond onepiece010938@gmail.com
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long:  `Start the linebot server and init with your channel secret & channel access token`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start called")
		fmt.Println(args)
		// cmd.Flags().GetString()
		s, _ := cmd.PersistentFlags().GetString("secret")
		fmt.Println("---->", s)
		fmt.Println(viper.GetString("baseDomain"))
		fmt.Println(viper.Get("CHANNEL_SECRET"))
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.PersistentFlags().StringP("secret", "s", "", "Line Channel Secret")
	startCmd.PersistentFlags().StringP("token", "t", "", "Line Channel Access Token")

	startCmd.MarkPersistentFlagRequired("secret")
	startCmd.MarkPersistentFlagRequired("token")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
