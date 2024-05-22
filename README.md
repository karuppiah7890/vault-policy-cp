# vault-policy-cp

Using this CLI tool, you can copy Vault Policies from one Vault instance to another Vault instance! :D

Note: The tool is written in Golang and uses Vault Official Golang API. The official Vault API documentation is here - https://pkg.go.dev/github.com/hashicorp/vault/api

Note: The tool needs Vault credentials of a user/account that has access to both the source and destination Vault, to read the Vault Policies from source Vault and to configure the Vault Policies in the destination Vault

Note: We have tested this only with some versions of Vault (like v1.15.x). So beware to test this in a testing environment with whatever version of Vault you are using, before using this in critical environments like production! Also, ensure that the testing environment is as close to your production environment as possible so that your testing makes sense

Note ⚠️‼️🚨: If the destination Vault has some policies already defined with the same name as the source Vault, when copying from source Vault to destination Vault, it will be overwritten! All the Vault Policies in source Vault will be present in destination Vault. If the destination Vault has some extra Vault Policies configured, it might have those untouched and intact. This scenario has NOT been tested currently though

Note: This does NOT copy the `default` and `root` Vault Policy as Vault does not support updating it / changing it

Future version ideas:
- Support for copying Vault Policies referred to in the Kubernetes Auth Method Roles (Roles Configuration) from source Vault to destination Vault
- Support for providing the Token Reviewer JWT Token as user input, say through environment variable, so that you can copy that to destination Vault and configure it as part of Kubernetes Auth Method Configuration, since it CANNOT be read from the source Vault through the Vault API
- Remove log verbosity for default settings. It's too verbose now, by default.
- Remove log which shows no information (`nil`) about the data copied to destination Vault, as Write API does NOT seem to be returning any data in the reponse, so, we can just ignore it

## Building

```bash
go build -v
```

## Usage

```bash
$ vault-policy-cp --help
usage: vault-policy-cp [<source-vault-policy-name> <destination-vault-policy-name>]

examples:

# show help
vault-policy-cp -h

# show help
$ vault-policy-cp --help

# copies all vault policies from source vault to destination vault.
# if a destination vault policy with the same name already exists,
# it will be overwritten.
$ vault-policy-cp

# copies allow_read policy from source vault to destination vault.
# if a destination vault policy with the same name already exists,
# it will be overwritten.
$ vault-policy-cp allow_read allow_read
```

# Demo

Source Vault, it's a secured Vault with HTTPS API enabled and a big token for root. It has some Vault Policies configured.

I'm using the Vault Root Token here for full access to the source Vault. But you don't need Vault Root Token. You just need any Vault Token / Credentials that has enough access to read Vault Policies in the source Vault

```bash
$ export VAULT_ADDR='https://127.0.0.1:8200'
$ export VAULT_TOKEN="some-big-token-here"
$ export VAULT_CACERT=$HOME/vault-ca.crt

$ vault status
Key                     Value
---                     -----
Seal Type               shamir
Initialized             true
Sealed                  false
Total Shares            5
Threshold               3
Version                 1.15.6
Build Date              2024-02-28T17:07:34Z
Storage Type            raft
Cluster Name            vault-cluster-9f170feb
Cluster ID              151e903e-e1e7-541e-d089-ce8db2da0a34
HA Enabled              true
HA Cluster              https://karuppiah-vault-0:8201
HA Mode                 active
Active Since            2024-04-27T23:15:36.130464099Z
Raft Committed Index    91107
Raft Applied Index      91107

$ vault policy list
allow_secrets
default
root

$ vault policy read allow_secrets
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

We can see that it has one policy already configured, named `allow_secrets` and we can also see the built-in `default` and `root` Vault Policies.

Let's also create two more Vault Policy to have some more Vault Policies in the source Vault, haha

```bash
$ cat /Users/karuppiah.n/every-day-log/allow_test_kv_secrets.hcl
# KV v2 secrets engine mount path is "test-kv"
path "test-kv/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

$ vault policy write allow_test_kv_secrets /Users/karuppiah.n/every-day-log/allow_test_kv_secrets.hcl
Success! Uploaded policy: allow_test_kv_secrets

$ vault policy list
allow_secrets
allow_test_kv_secrets
default
root

$ vault policy read allow_test_kv_secrets
# KV v2 secrets engine mount path is "test-kv"
path "test-kv/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

Now we have one more Vault Policy in the source Vault named `allow_test_kv_secrets` :)

Let's just create one more, and then that's it :)

```bash
$ cat /Users/karuppiah.n/every-day-log/allow_stage_kv_secrets.hcl
# KV v2 secrets engine mount path is "stage-kv"
path "stage-kv/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

$ vault policy write allow_stage_kv_secrets /Users/karuppiah.n/every-day-log/allow_stage_kv_secrets.hcl
Success! Uploaded policy: allow_stage_kv_secrets

$ vault policy list
allow_secrets
allow_stage_kv_secrets
allow_test_kv_secrets
default
root

$ vault policy read allow_stage_kv_secrets
# KV v2 secrets engine mount path is "stage-kv"
path "stage-kv/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

Now, let's move on to the destination Vault.

Destination Vault, it's a local dev Vault server, with no HTTPS, with token as `root`. It was run using

```bash
$ vault server -dev -dev-root-token-id root -dev-listen-address 127.0.0.1:8300
```

We see that it has no auth enabled.

I'm using the Vault Root Token here for full access

```bash
$ export VAULT_ADDR='http://127.0.0.1:8300'
$ export VAULT_TOKEN="root"

$ vault status
Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    1
Threshold       1
Version         1.15.4
Build Date      2023-12-04T17:45:28Z
Storage Type    inmem
Cluster Name    vault-cluster-4bf5d460
Cluster ID      a735f318-84c4-e1a5-b2d2-b11517a62463
HA Enabled      false

$ vault policy list
default
root
```

As we can see it has no Vault Policies defined, and it has `default` and `root` Vault Policies.

Now let's copy the Vault Policies from source Vault to destination Vault :) First I'll copy just one Vault Policy, then I'll copy all Vault Policies, just to show that both can be done :)

I'm using the Vault Root Token here for both source Vault and destination Vault, for full access. But you don't need Vault Root Token. You just need any Vault Token / Credentials that has enough access to read Vault Policies from source Vault and configure (write) Vault Policies in destination Vault

```bash
$ export SOURCE_VAULT_ADDR='https://127.0.0.1:8200'
$ export SOURCE_VAULT_TOKEN="some-big-token-here"
$ export SOURCE_VAULT_CACERT=$HOME/vault-ca.crt

$ export DESTINATION_VAULT_ADDR='http://127.0.0.1:8300'
$ export DESTINATION_VAULT_TOKEN="root"

$ ./vault-policy-cp allow_secrets allow_secrets

copying `allow_secrets` policy in source vault to `allow_secrets` policy in destination vault

source vault policy `allow_secrets` rules: path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

Now, we have copied just one Vault Policy from source Vault to destination Vault. It's the `allow_secrets` Vault Policy.

Now, let's look at the destination Vault and see if `allow_secrets` Vault Policy is copied to destination Vault

```bash
$ export VAULT_ADDR='http://127.0.0.1:8300'
$ export VAULT_TOKEN="root"

$ vault status
Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    1
Threshold       1
Version         1.15.4
Build Date      2023-12-04T17:45:28Z
Storage Type    inmem
Cluster Name    vault-cluster-4bf5d460
Cluster ID      a735f318-84c4-e1a5-b2d2-b11517a62463
HA Enabled      false

$ vault policy list
allow_secrets
default
root
```

Everything looks good! :D The `allow_secrets` Vault Policy is present in the destination Vault :D

Now, let's try to copy all the Vault Policies from source Vault to destination Vault. I'll use the `allow_test_kv_secrets` Vault Policy as an example. It's present in source Vault but not in destination Vault. I'll also use the `allow_stage_kv_secrets` Vault Policy as an example. It's present in source Vault and but not in destination Vault. Note that `allow_secrets` Vault Policy is already present in destination Vault. But this is not a problem. We can still just copy all the Vault Policies from source Vault to destination Vault, and we will get all the Vault Policies present in source Vault and destination Vault. Any policies that are already present in destination Vault with the same name will **be overwritten**!

Let's look at how to copy all the Vault Policies from source Vault to destination Vault

```bash
$ export SOURCE_VAULT_ADDR='https://127.0.0.1:8200'
$ export SOURCE_VAULT_TOKEN="some-big-token-here"
$ export SOURCE_VAULT_CACERT=$HOME/vault-ca.crt

$ export DESTINATION_VAULT_ADDR='http://127.0.0.1:8300'
$ export DESTINATION_VAULT_TOKEN="root"

$ ./vault-policy-cp

copying the following vault policies in source vault to destination vault: [allow_secrets allow_stage_kv_secrets allow_test_kv_secrets default root]

copying `allow_secrets` policy in source vault to `allow_secrets` policy in destination vault

source vault policy `allow_secrets` rules: path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}


copying `allow_stage_kv_secrets` policy in source vault to `allow_stage_kv_secrets` policy in destination vault

source vault policy `allow_stage_kv_secrets` rules: # KV v2 secrets engine mount path is "stage-kv"
path "stage-kv/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}


copying `allow_test_kv_secrets` policy in source vault to `allow_test_kv_secrets` policy in destination vault

source vault policy `allow_test_kv_secrets` rules: # KV v2 secrets engine mount path is "test-kv"
path "test-kv/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

That's all! :D It's done :)

Now, let's check the destination Vault and see if all the Vault Policies are copied to destination Vault

```bash
$ export VAULT_ADDR='http://127.0.0.1:8300'
$ export VAULT_TOKEN="root"

$ vault status
Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    1
Threshold       1
Version         1.15.4
Build Date      2023-12-04T17:45:28Z
Storage Type    inmem
Cluster Name    vault-cluster-4bf5d460
Cluster ID      a735f318-84c4-e1a5-b2d2-b11517a62463
HA Enabled      false

$ vault policy list
allow_secrets
allow_stage_kv_secrets
allow_test_kv_secrets
default
root
```

Note: If you try to copy the `root` Vault Policy from source Vault to destination Vault, it will throw an error similar to below as there's nothing to read from `root` policy, it's empty -

```bash
$ export SOURCE_VAULT_ADDR='https://127.0.0.1:8200'
$ export SOURCE_VAULT_TOKEN="some-big-token-here"
$ export SOURCE_VAULT_CACERT=$HOME/vault-ca.crt

$ export DESTINATION_VAULT_ADDR='http://127.0.0.1:8300'
$ export DESTINATION_VAULT_TOKEN="root"

$ ./vault-policy-cp root root

copying `root` policy in source vault to `root` policy in destination vault

source vault policy `root` rules:
error writing `root` vault policy to destination vault: Error making API request.

URL: PUT http://127.0.0.1:8300/v1/sys/policies/acl/root
Code: 400. Errors:

* 'policy' parameter not supplied or empty

$ ./vault-policy-cp root root_copy

copying `root` policy in source vault to `root_copy` policy in destination vault

source vault policy `root` rules:
error writing `root_copy` vault policy to destination vault: Error making API request.

URL: PUT http://127.0.0.1:8300/v1/sys/policies/acl/root_copy
Code: 400. Errors:

* 'policy' parameter not supplied or empty
```

You can also notice that the `root` Vault Policy is empty if you try to read the policy using Vault CLI, but using `vault read` command, as `vault policy read` command does **NOT** workout here, probably because they (Vault CLI developers) put some check there and removed the `root` policy, hence the `No policy named: root` error as seen below -

```bash
# straightforward way to list
$ vault policy list
allow_secrets
allow_stage_kv_secrets
allow_test_kv_secrets
default
root

# straightforward way to read a policy generally
$ vault policy read allow_secrets
path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# straightforward way to read but it doesn't work for root policy
$ vault policy read root
No policy named: root

# a bit complicated way to list - needs knowledge about list API and the API
# path / endpoint
$ vault list sys/policies/acl
Keys
----
allow_secrets
allow_stage_kv_secrets
allow_test_kv_secrets
default
root

# You can also put an extra forward slash ("/") at the end of the API path like
# the below, near `acl`
$ vault list sys/policies/acl/
Keys
----
allow_secrets
allow_stage_kv_secrets
allow_test_kv_secrets
default
root

# a bit complicated way to read - needs knowledge about read API and the API
# path / endpoint
$ vault read sys/policies/acl/allow_secrets
Key       Value
---       -----
name      allow_secrets
policy    path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# a bit complicated way to read, and it works for root policy, showing that
# the `root` policy is empty
$ vault read sys/policies/acl/root
Key       Value
---       -----
name      root
policy    n/a
```

Note: If you try to copy the any Vault Policy, other than `root`, from source Vault to destination Vault, but with the name of `root` in the destination Vault, it will throw an error similar to below as one cannot write to the `root` policy in general, it's not allowed -

```bash
$ export SOURCE_VAULT_ADDR='https://127.0.0.1:8200'
$ export SOURCE_VAULT_TOKEN="some-big-token-here"
$ export SOURCE_VAULT_CACERT=$HOME/vault-ca.crt

$ export DESTINATION_VAULT_ADDR='http://127.0.0.1:8300'
$ export DESTINATION_VAULT_TOKEN="root"

$ ./vault-policy-cp allow_secrets root

copying `allow_secrets` policy in source vault to `root` policy in destination vault

source vault policy `allow_secrets` rules: path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

error writing `root` vault policy to destination vault: Error making API request.

URL: PUT http://127.0.0.1:8300/v1/sys/policies/acl/root
Code: 400. Errors:

* cannot update "root" policy
```

You can also notice that it's not allowed to write to the `root` policy in Vault in general by doing something like this in the source or destination Vault -

```bash
$ vault policy write root /Users/karuppiah.n/every-day-log/allow_test_kv_secrets.hcl
Error uploading policy: Error making API request.

URL: PUT https://127.0.0.1:8200/v1/sys/policies/acl/root
Code: 400. Errors:

* cannot update "root" policy
```

Note: If the Vault Token / Credentials used for the destination Vault is not valid / wrong / does not have enough access, then the tool throws errors similar to this -

```bash
$ export SOURCE_VAULT_ADDR='https://127.0.0.1:8200'
$ export SOURCE_VAULT_TOKEN="some-big-token-here"
$ export SOURCE_VAULT_CACERT=$HOME/vault-ca.crt

$ export DESTINATION_VAULT_ADDR='http://127.0.0.1:8300'
$ export DESTINATION_VAULT_TOKEN="blah" # wrong Vault Token

$ ./vault-policy-cp allow_secrets allow_secrets

copying `allow_secrets` policy in source vault to `allow_secrets` policy in destination vault

source vault policy `allow_secrets` rules: path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

error writing `allow_secrets` vault policy to destination vault: Error making API request.

URL: PUT http://127.0.0.1:8300/v1/sys/policies/acl/allow_secrets
Code: 403. Errors:

* permission denied

$ ./vault-policy-cp 

copying the following vault policies in source vault to destination vault: [allow_secrets allow_stage_kv_secrets allow_test_kv_secrets default root]

copying `allow_secrets` policy in source vault to `allow_secrets` policy in destination vault

source vault policy `allow_secrets` rules: path "secret/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

error writing `allow_secrets` vault policy to destination vault: Error making API request.

URL: PUT http://127.0.0.1:8300/v1/sys/policies/acl/allow_secrets
Code: 403. Errors:

* permission denied
```

Note: If the Vault Token / Credentials used for the source Vault is not valid / wrong / does not have enough access, then the tool throws errors similar to this -

```bash
$ export SOURCE_VAULT_ADDR='https://127.0.0.1:8200'
$ export SOURCE_VAULT_TOKEN="blah"
$ export SOURCE_VAULT_CACERT=$HOME/vault-ca.crt

$ export DESTINATION_VAULT_ADDR='http://127.0.0.1:8300'
$ export DESTINATION_VAULT_TOKEN="root"

$ ./vault-policy-cp allow_secrets allow_secrets
error reading 'allow_secrets' vault policy from source vault: Error making API request.

URL: GET https://127.0.0.1:8200/v1/sys/policies/acl/allow_secrets
Code: 403. Errors:

* permission denied

$ ./vault-policy-cp 
error listing source vault policies: Error making API request.

URL: GET https://127.0.0.1:8200/v1/sys/policies/acl?list=true
Code: 403. Errors:

* permission denied
```

## Future Ideas

Talking about future ideas, here are some of the ideas for the future -
- To backup before doing any writes - regardless of just plain new writes / overwrites
- Show clear logs and mention if the copy is doing an overwrite or not (if it's a plain new write)
- Ability to give a specific set of policies alone to be copied from source Vault to destination Vault. As of now it's either one or all policies, in one command. I want to allow users to just run one command and copy N number of policies. Maybe take a file as input, say YAML file, or JSON file, with all the policies to be copied - and just run one command to do that. The file would say what's the Vault Policy name at source Vault and what should be the Vault Policy name at destination Vault. Something like -

```json
[
  {
    "source-vault-policy-name": "some-policy-1",
    "destination-vault-policy-name": "some-policy-1"
  },
  {
    "source-vault-policy-name": "some-policy-2",
    "destination-vault-policy-name": "some-policy-2"
  }
]
```
