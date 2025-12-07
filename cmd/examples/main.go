// Example demonstrates how to use the goPAS SDK to interact with CyberArk.
//
// This example shows common operations including:
// - Creating a session
// - Listing safes
// - Managing accounts
// - Retrieving passwords
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/chrisranney/gopas"
	"github.com/chrisranney/gopas/pkg/accounts"
	"github.com/chrisranney/gopas/pkg/safemembers"
	"github.com/chrisranney/gopas/pkg/safes"
	"github.com/chrisranney/gopas/pkg/users"
)

func main() {
	// Get credentials from environment variables
	baseURL := os.Getenv("CYBERARK_URL")
	username := os.Getenv("CYBERARK_USER")
	password := os.Getenv("CYBERARK_PASSWORD")

	if baseURL == "" || username == "" || password == "" {
		log.Fatal("Please set CYBERARK_URL, CYBERARK_USER, and CYBERARK_PASSWORD environment variables")
	}

	ctx := context.Background()

	// Example 1: Create a session
	fmt.Println("=== Creating Session ===")
	sess, err := gopas.NewSession(ctx, gopas.SessionOptions{
		BaseURL: baseURL,
		Credentials: gopas.Credentials{
			Username: username,
			Password: password,
		},
		AuthMethod: gopas.AuthMethodCyberArk,
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer gopas.CloseSession(ctx, sess)

	fmt.Printf("Session created successfully. User: %s\n", sess.User)
	fmt.Printf("CyberArk Version: %s\n", sess.ExternalVersion)

	// Example 2: Get server info
	fmt.Println("\n=== Server Info ===")
	serverInfo, err := gopas.GetServerInfo(ctx, sess)
	if err != nil {
		log.Printf("Failed to get server info: %v", err)
	} else {
		fmt.Printf("Server: %s (v%s)\n", serverInfo.ServerName, serverInfo.ExternalVersion)
	}

	// Example 3: List safes
	fmt.Println("\n=== Listing Safes ===")
	safesResp, err := gopas.ListSafes(ctx, sess, gopas.ListSafesOptions{
		Limit: 10,
	})
	if err != nil {
		log.Printf("Failed to list safes: %v", err)
	} else {
		fmt.Printf("Found %d safes:\n", safesResp.Count)
		for _, safe := range safesResp.Value {
			fmt.Printf("  - %s (%s)\n", safe.SafeName, safe.Description)
		}
	}

	// Example 4: Create a safe (commented out by default)
	/*
		fmt.Println("\n=== Creating Safe ===")
		newSafe, err := gopas.CreateSafe(ctx, sess, gopas.CreateSafeOptions{
			SafeName:    "TestSafe",
			Description: "Created by goPAS example",
		})
		if err != nil {
			log.Printf("Failed to create safe: %v", err)
		} else {
			fmt.Printf("Created safe: %s\n", newSafe.SafeName)
		}
	*/

	// Example 5: List accounts
	fmt.Println("\n=== Listing Accounts ===")
	accountsResp, err := gopas.ListAccounts(ctx, sess, gopas.ListAccountsOptions{
		Limit: 10,
	})
	if err != nil {
		log.Printf("Failed to list accounts: %v", err)
	} else {
		fmt.Printf("Found %d accounts:\n", accountsResp.Count)
		for _, acct := range accountsResp.Value {
			fmt.Printf("  - %s@%s (Safe: %s)\n", acct.UserName, acct.Address, acct.SafeName)
		}
	}

	// Example 6: Search for specific accounts
	fmt.Println("\n=== Searching Accounts ===")
	searchResp, err := accounts.List(ctx, sess, accounts.ListOptions{
		Search: "admin",
		Limit:  5,
	})
	if err != nil {
		log.Printf("Failed to search accounts: %v", err)
	} else {
		fmt.Printf("Found %d matching accounts\n", searchResp.Count)
	}

	// Example 7: List users
	fmt.Println("\n=== Listing Users ===")
	usersResp, err := users.List(ctx, sess, users.ListOptions{
		Limit: 10,
	})
	if err != nil {
		log.Printf("Failed to list users: %v", err)
	} else {
		fmt.Printf("Found %d users:\n", usersResp.Total)
		for _, user := range usersResp.Users {
			fmt.Printf("  - %s (%s)\n", user.Username, user.UserType)
		}
	}

	// Example 8: Get safe members
	fmt.Println("\n=== Safe Members ===")
	if len(safesResp.Value) > 0 {
		safeName := safesResp.Value[0].SafeName
		membersResp, err := safemembers.List(ctx, sess, safeName, safemembers.ListOptions{})
		if err != nil {
			log.Printf("Failed to get safe members for %s: %v", safeName, err)
		} else {
			fmt.Printf("Members of '%s':\n", safeName)
			for _, member := range membersResp.Value {
				fmt.Printf("  - %s (%s)\n", member.MemberName, member.MemberType)
			}
		}
	}

	// Example 9: Get specific safe details
	fmt.Println("\n=== Safe Details ===")
	if len(safesResp.Value) > 0 {
		safeName := safesResp.Value[0].SafeName
		safe, err := safes.Get(ctx, sess, safeName)
		if err != nil {
			log.Printf("Failed to get safe %s: %v", safeName, err)
		} else {
			fmt.Printf("Safe: %s\n", safe.SafeName)
			fmt.Printf("  Description: %s\n", safe.Description)
			fmt.Printf("  Managing CPM: %s\n", safe.ManagingCPM)
			fmt.Printf("  OLAC Enabled: %v\n", safe.OLACEnabled)
		}
	}

	// Example 10: Retrieve password (commented out by default - requires permissions)
	/*
		fmt.Println("\n=== Retrieving Password ===")
		if len(accountsResp.Value) > 0 {
			accountID := accountsResp.Value[0].ID
			password, err := gopas.GetAccountPassword(ctx, sess, accountID, "Testing goPAS SDK")
			if err != nil {
				log.Printf("Failed to retrieve password: %v", err)
			} else {
				fmt.Printf("Password retrieved successfully (length: %d)\n", len(password))
			}
		}
	*/

	fmt.Println("\n=== Example Complete ===")
}
