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
	"os/exec"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/spf13/cobra"
)

// assumeroleCmd represents the assumerole command
var (
	assumeroleCmd = &cobra.Command{
		Use:   "assumerole",
		Short: "Assume an AWS role",
		Long:  `Assume an AWS role.`,
		Run: func(cmd *cobra.Command, args []string) {
			// getTokenSTS(args[0])

			arn, err := arnBuilder(args[0], args[1], "AWS::IAM::Role")
			if err != nil {
				fmt.Println("Something went wrong while buidling arn:", err)
				return
			}
			stsResponse, err := getTokenSTS(arn)
			if err != nil {
				fmt.Println("Something went wrong while getting sts token", err)
			}
			spawnTerminal(stsResponse, args[0], args[1])
		},
	}
)

func init() {
	rootCmd.AddCommand(assumeroleCmd)
}

// Spanws a new terminal with the assumed role
// with the correct environment variables
func spawnTerminal(stsOutput *sts.AssumeRoleOutput, role string, awsAccountID string) {
	binary, lookErr := exec.LookPath("bash")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"bash"}

	os.Setenv("AWS_ACCESS_KEY_ID", *stsOutput.Credentials.AccessKeyId)
	os.Setenv("AWS_SECRET_ACCESS_KEY", *stsOutput.Credentials.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", *stsOutput.Credentials.SessionToken)
	env := os.Environ()

	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		panic(execErr)
	}
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
