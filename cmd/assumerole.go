/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// assumeroleCmd represents the assumerole command
var (
	assumeroleCmd = &cobra.Command{
		Use:   "assumerole",
		Short: "Assume an AWS role",
		Long:  `Assume an AWS role.`,
		Run: func(cmd *cobra.Command, args []string) {
			// getTokenSTS(args[0])

			unsetEnvironment()
			arn, err := arnBuilder(args[0], args[1], "AWS::IAM::Role")
			if err != nil {
				fmt.Println("Something went wrong while buidling arn:", err)
				return
			}
			stsResponse, err := getTokenSTS(arn)
			if err != nil {
				fmt.Println("Something went wrong while getting sts token", err)
			}
			createProfile(stsResponse, args[0], args[1])

		},
	}
)

func init() {

	rootCmd.AddCommand(assumeroleCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// assumeroleCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// assumeroleCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// Sets required environment variables in a new profile in given
// config file. Sets the following variables under that profile:
// AWS_SECRET_ACCESS_KEY, AWS_ACCESS_KEY_ID and AWS_SESSION_TOKEN
func createProfile(stsResponse *sts.AssumeRoleOutput, role string, accountID string) {

	viper.SetConfigName("credentials")
	viper.SetConfigType("ini")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	profileName := role + "-" + accountID
	viper.Set(profileName+".aws_access_key_id", stsResponse.Credentials.AccessKeyId)
	viper.Set(profileName+".aws_secret_access_key", stsResponse.Credentials.SecretAccessKey)
	viper.Set(profileName+".aws_session_token", stsResponse.Credentials.SessionToken)
	viper.WriteConfigAs("./credentials")
}

// Unsets the current assumed role environment variables
// AWS_SECRET_ACCESS_KEY, AWS_ACCESS_KEY_ID and AWS_SESSION_TOKEN
func unsetEnvironment() {
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SESSION_TOKEN")
}

// Gets a AccessKeyId, SecretAccessKey, SessionToken
// using the AWS STS client. Receives a role and an
// account
func getTokenSTS(role string) (*sts.AssumeRoleOutput, error) {
	// TODO: Pick up aws default config profile from
	// the configuration file

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})

	stsSvc := sts.New(sess)
	sessionName := "get_token_sess"

	result, err := stsSvc.AssumeRole(&sts.AssumeRoleInput{
		RoleArn:         &role,
		RoleSessionName: &sessionName,
	})

	if err != nil {
		fmt.Println("AssumeRole Error", err)
		return nil, err
	}

	return result, nil
}

// Returns an built ARN from an account id and a role
//
// IAM Roles: arn:aws:iam::${accountID}:role/${resourceID}
func arnBuilder(resourceID string, accountID string, resourceType string) (string, error) {
	if resourceType == "AWS::IAM::Role" {
		arn := fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, resourceID)
		return arn, nil
	}
	return "", errors.New("some error")
}
