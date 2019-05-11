# Github Gitea mirror script

Mirror and sync your public github repositories into your Gitea account.

These environment variables are required to run properly:

- `GITEA_TOKEN`: "your gitea token"
- `GITEA_HOST`: "https://gitea.io or whatever gitea instance"
- `GITEA_USERNAME` 
- `GITHUB_USERNAME`

## Usage

Mirror Github repositories

Run `go run main.go`

Sync Gitea mirror repositories

Run `go run main.go --action=sync`
