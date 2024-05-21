package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
)

var usage = `usage: vault-policy-cp [<source-vault-policy-name> <destination-vault-policy-name>]

examples:

# show help
vault-policy-cp -h

# show help
vault-policy-cp --help

# copies all vault policies from source vault to destination vault.
# if a destination vault policy with the same name already exists,
# it will be overwritten.
vault-policy-cp

# copies allow_read policy from source vault to destination vault.
# if a destination vault policy with the same name already exists,
# it will be overwritten.
vault-policy-cp allow_read allow_read
`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "%s", usage)
		os.Exit(0)
	}
	showHelp := flag.Bool("h", false, "help")
	flag.Parse()

	if *showHelp {
		flag.Usage()
	}

	if !(flag.NArg() == 2 || flag.NArg() == 0) {
		fmt.Fprintf(os.Stderr, "invalid number of arguments: %d. expected 0 or 2 arguments.\n\n", flag.NArg())
		flag.Usage()
	}

	// Get Config for Source Vault
	sourceConfig := getSourceVaultConfig()

	// Create a new client to the source vault
	sourceDefaultConfig := api.DefaultConfig()
	sourceDefaultConfig.Address = sourceConfig.Address

	if sourceConfig.CACertPath != "" {
		sourceDefaultConfig.ConfigureTLS(&api.TLSConfig{CACert: sourceConfig.CACertPath})
	}
	sourceClient, err := api.NewClient(sourceDefaultConfig)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating source vault client: %s\n", err)
		os.Exit(1)
	}

	// Set the token for the source vault client
	sourceClient.SetToken(sourceConfig.Token)

	// Get Config for Destination Vault
	destinationConfig := getDestinationVaultConfig()

	// Create a new client to the destination vault
	destinationDefaultConfig := api.DefaultConfig()
	destinationDefaultConfig.Address = destinationConfig.Address
	if destinationConfig.CACertPath != "" {
		destinationDefaultConfig.ConfigureTLS(&api.TLSConfig{CACert: destinationConfig.CACertPath})
	}
	destinationClient, err := api.NewClient(destinationDefaultConfig)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating destination vault client: %s\n", err)
		os.Exit(1)
	}

	// Set the token for the destination vault client
	destinationClient.SetToken(destinationConfig.Token)

	if flag.NArg() == 0 {
		vaultPolicyCopyAll(sourceClient, destinationClient)
		return
	}

	sourceVaultPolicyName := flag.Args()[0]
	destinationVaultPolicyName := flag.Args()[1]

	vaultPolicyCopy(sourceClient, destinationClient, sourceVaultPolicyName, destinationVaultPolicyName)
}

func vaultPolicyCopy(sourceClient *api.Client, destinationClient *api.Client, sourceVaultPolicyName string, destinationVaultPolicyName string) {
	// TODO: Take a backup of the destination policy in case one is already present,
	// regardless of if they have the same name as source policy.
	// So that we have a backup just in case, especially before overwriting.

	sourceVaultPolicy, err := sourceClient.Sys().GetPolicy(sourceVaultPolicyName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading '%s' vault policy from source vault: %s\n", sourceVaultPolicyName, err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "\ncopying `%s` policy in source vault to `%s` policy in destination vault\n", sourceVaultPolicyName, destinationVaultPolicyName)
	fmt.Fprintf(os.Stdout, "\nsource vault policy `%s` rules: %+v\n", sourceVaultPolicyName, sourceVaultPolicy)

	err = destinationClient.Sys().PutPolicy(destinationVaultPolicyName, sourceVaultPolicy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing `%s` vault policy to destination vault: %s\n", destinationVaultPolicyName, err)
		os.Exit(1)
	}
}

func vaultPolicyCopyAll(sourceClient *api.Client, destinationClient *api.Client) {
	// TODO: Take a backup of the destination policies in case they are already present,
	// regardless of if they have the same name as source policies.
	// So that we have a backup just in case, especially before overwriting.

	// Question: Should we do a complete backup before doing policy copy one by one?
	// Or should we do a backup of each policy one by one? When we copy them one by one
	// that is. Basically - take a backup one by one, at vaultPolicyCopy() function level or
	// take a complete backup of all destination vault policies at vaultPolicyCopyAll()
	// function level.

	policies, err := sourceClient.Sys().ListPolicies()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing source vault policies: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "\ncopying the following vault policies in source vault to destination vault: %+v\n", policies)

	// Note: Ignore root policy as it cannot be updated

	// Copy all policies to destination vault
	for _, policyName := range policies {
		if policyName == "root" {
			// Ignore root policy as it cannot be updated
			continue
		}
		vaultPolicyCopy(sourceClient, destinationClient, policyName, policyName)
	}
}
