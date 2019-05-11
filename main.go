package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type GiteaUser struct {
	Id int `json:"id"`
}

type GithubRepo struct {
	Name        string `json:"name"`
	CloneUrl    string `json:"clone_url"`
	Description string `json:"description"`
}

type GiteaRepo struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Mirror      bool   `json:"mirror"`
}

type GiteaMirrorRepo struct {
	AuthPassword string `json:"auth_password"`
	AuthUsername string `json:"auth_username"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	CloneAddr    string `json:"clone_addr"`
	Mirror       bool   `json:"mirror"`
	Private      bool   `json:"private"`
	RepoName     string `json:"repo_name"`
	Uid          int    `json:"uid"`
}

func setupRequest(request *http.Request) *http.Request {
	token := fmt.Sprintf("token %s", os.Getenv("GITEA_TOKEN"))
	request.Header.Set("Authorization", token)
	request.Header.Set("Content-Type", "application/json")
	return request
}

func CreateGiteaMirrorRepo(repo GiteaMirrorRepo) {
	client := &http.Client{}
	payload, _ := json.Marshal(repo)

	url := fmt.Sprintf("%s/api/v1/repos/migrate", os.Getenv("GITEA_HOST"))
	request, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	request = setupRequest(request)
	response, _ := client.Do(request)
	if response.StatusCode == 201 {
		fmt.Printf("%s mirror created\n", repo.Name)
	}
}

func GiteaUserInfo(username string) GiteaUser {
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/v1/users/%s", os.Getenv("GITEA_HOST"), username)
	request, _ := http.NewRequest("GET", url, nil)
	request = setupRequest(request)
	response, _ := client.Do(request)
	var user GiteaUser
	body, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(body, &user)
	return user
}

func GithubRepos(username string) []GithubRepo {
	url := fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=%d", username, 1000)
	response, _ := http.Get(url)
	content, _ := ioutil.ReadAll(response.Body)

	var repos []GithubRepo
	json.Unmarshal(content, &repos)
	return repos
}

func GetGiteaRepos(username string) []GiteaRepo {
	client := &http.Client{}
	url := fmt.Sprintf("%s/api/v1/users/%s/repos", os.Getenv("GITEA_HOST"), username)
	request, _ := http.NewRequest("GET", url, nil)
	request = setupRequest(request)
	response, _ := client.Do(request)
	var repos []GiteaRepo
	body, _ := ioutil.ReadAll(response.Body)
	json.Unmarshal(body, &repos)
	return repos
}

func SyncGiteaMirrors(username string) {
	client := &http.Client{}
	repos := GetGiteaRepos(username)
	for _, repo := range repos {
		if !repo.Mirror {
			continue
		}
		fmt.Printf("Syncing %s...\n", repo.Name)
		url := fmt.Sprintf("%s/api/v1/repos/%s/%s/mirror-sync", os.Getenv("GITEA_HOST"), username, repo.Name)
		request, _ := http.NewRequest("POST", url, nil)
		request = setupRequest(request)
		response, _ := client.Do(request)
		if response.StatusCode == 200 {
			fmt.Printf("%s Synced\n", repo.Name)
		}
	}
}

func existGiteaRepo(repos []GiteaRepo, repoName string) bool {
	for _, repo := range repos {
		if repo.Name == repoName {
			return true
		}
	}
	return false
}

func FilterGihubReposWithNoGiteaMirror(githubRepos []GithubRepo, giteaRepos []GiteaRepo) []GithubRepo {
	var result []GithubRepo

	for _, repo := range githubRepos {
		if !existGiteaRepo(giteaRepos, repo.Name) {
			result = append(result, repo)
		}
	}

	return result
}

func CreateMirrorFromGithub(repo GithubRepo, giteaUser GiteaUser) {
	fmt.Printf("Name: %s\n", repo.Name)
	fmt.Printf("Clone url: %s\n", repo.CloneUrl)
	fmt.Printf("Description: %s\n", repo.Description)
	giteaRepo := GiteaMirrorRepo{
		Name:        repo.Name,
		Description: repo.Description,
		CloneAddr:   repo.CloneUrl,
		Mirror:      true,
		Private:     true,
		RepoName:    repo.Name,
		Uid:         giteaUser.Id,
	}
	CreateGiteaMirrorRepo(giteaRepo)
}

func CreateGiteaMirrors(githubUsername string, giteaUsername string) {
	githubRepos := GithubRepos(githubUsername)
	giteaRepos := GetGiteaRepos(giteaUsername)
	repos := FilterGihubReposWithNoGiteaMirror(githubRepos, giteaRepos)
	giteaUser := GiteaUserInfo(giteaUsername)

	for _, repo := range repos {
		CreateMirrorFromGithub(repo, giteaUser)
	}
}

func main() {
	var action string
	flag.StringVar(&action, "action", "mirror", "Action to be run")
	flag.Parse()

	fmt.Printf("Action %s\n", action)
	GITEA_HOST := os.Getenv("GITEA_HOST")
	GITEA_TOKEN := os.Getenv("GITEA_TOKEN")
	GITEA_USERNAME := os.Getenv("GITEA_USERNAME")
	GITHUB_USERNAME := os.Getenv("GITHUB_USERNAME")

	fmt.Println("GITEA_HOST:", GITEA_HOST)
	fmt.Println("GITEA_TOKEN:", GITEA_TOKEN)
	fmt.Println("GITEA_USERNAME:", GITEA_USERNAME)
	fmt.Println("GITHUB_USERNAME:", GITHUB_USERNAME)

	if action == "mirror" {
		CreateGiteaMirrors(GITHUB_USERNAME, GITEA_USERNAME)
	} else if action == "sync" {
		SyncGiteaMirrors(GITEA_USERNAME)
	}
}
