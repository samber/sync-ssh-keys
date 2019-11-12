package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v28/github"
	"github.com/thoas/go-funk"
	"golang.org/x/oauth2"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	githubToken        = kingpin.Flag("github-token", "Github oauth2 token.").Envar("GITHUB_TOKEN").String()
	githubOrg          = kingpin.Flag("github-org", "Github organization name.").String()
	githubTeams        = kingpin.Flag("github-team", "List of teams allowed to access server.").Strings()
	githubUsers        = kingpin.Flag("github-user", "List of usernames allowed to access server.").Strings()
	excludeGithubUsers = kingpin.Flag("exclude-github-user", "List of users to explicitly exclude.").Strings()

	werror = kingpin.Flag("Werror", "Treat warning as errors. Fatal error if organization, team or user does not exist.").Bool()

	githubClient    *github.Client
	githubClientCtx context.Context
)

func githubBootContext() (*github.Client, context.Context) {
	ctx := context.Background()

	if *githubToken == "" {
		return github.NewClient(nil), ctx
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc), ctx
}

func warningMsg(err error, msg string) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}

	if werror != nil && *werror {
		log.Fatal(msg)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
}

func getGithubOrgMembers() []string {
	// fetch 100 results per api call
	opt := &github.ListMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allUsers []*github.User
	for {
		users, resp, err := githubClient.Organizations.ListMembers(githubClientCtx, *githubOrg, opt)
		if err != nil {
			warningMsg(err, fmt.Sprintf("[warning] Github Organisation \"%s\" not found", *githubOrg))
			break
		}

		allUsers = append(allUsers, users...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return funk.Map(allUsers, func(user *github.User) string {
		return user.GetLogin()
	}).([]string)

}

func getGithubTeamMembers(teamName string) []string {
	// get team id from slug
	team, _, err := githubClient.Teams.GetTeamBySlug(githubClientCtx, *githubOrg, teamName)
	if err != nil {
		warningMsg(err, fmt.Sprintf("[warning] Github Organisation \"%s\" not found.\n", *githubOrg))
		return []string{}
	}

	// fetch 100 results per api call
	opt := &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allUsers []*github.User
	for {
		users, resp, err := githubClient.Teams.ListTeamMembers(githubClientCtx, team.GetID(), opt)
		if err != nil {
			warningMsg(err, fmt.Sprintf("[warning] Github teams \"%s\" not found.\n", teamName))
			break
		}

		allUsers = append(allUsers, users...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return funk.Map(allUsers, func(user *github.User) string {
		return user.GetLogin()
	}).([]string)
}

func getUserSSHKeys(username string) []string {
	resp, err := http.Get("https://github.com/" + username + ".keys")
	if err != nil {
		warningMsg(err, fmt.Sprintf("[warning] Failed to fetch public ssh key of user \"%s\"", username))
		return []string{}
	}

	body, _ := ioutil.ReadAll(resp.Body)
	sshKeys := strings.Split(string(body), "\n")

	// removing empty lines
	sshKeys = funk.Filter(sshKeys, func(sshKey string) bool {
		return len(sshKey) > 0
	}).([]string)

	// adding ssh key owner
	return funk.Map(sshKeys, func(sshKey string) string {
		return fmt.Sprintf("%s %s@github-team-ssh-key", sshKey, username)
	}).([]string)
}

func output(sshKeys []string) {
	if len(sshKeys) > 0 {
		fmt.Println("#\n# Generated with https://github.com/samber/github-team-ssh-keys\n#\n")
		fmt.Println(strings.Join(sshKeys, "\n\n"))
	}
}

func checkFlags() {
	// error
	if len(*githubOrg)+len(*githubTeams)+len(*githubUsers) == 0 {
		kingpin.FatalUsage("--github-org, --github-team or --github-user must be provided.\n")
	}
	if len(*githubTeams) > 0 && len(*githubOrg) == 0 {
		kingpin.FatalUsage("--github-team cannot be provided without --github-org.\n")
	}
	if len(*githubTeams) > 0 && len(*githubToken) == 0 {
		kingpin.FatalUsage("--github-team cannot be provided without --github-token.\n")
	}

	// warnings
	if len(*githubOrg) > 0 && len(*githubToken) == 0 {
		fmt.Fprintln(os.Stderr, "[warning] You provided --github-org without --github-token: organization private members won't be fetched.")
	}
}

func main() {
	kingpin.Version("0.2.0")
	kingpin.Parse()
	checkFlags()

	githubClient, githubClientCtx = githubBootContext()
	if githubClient == nil {
		log.Fatal("Invalid Github client")
	}

	var usernames []string

	// get users from teams
	if githubOrg != nil && len(*githubOrg) > 0 {
		if githubTeams != nil && len(*githubTeams) > 0 {
			for _, team := range *githubTeams {
				usernames = append(usernames, getGithubTeamMembers(team)...)
			}
		} else {
			usernames = append(usernames, getGithubOrgMembers()...)
		}
	}

	// get users provided in cli
	if githubUsers != nil {
		usernames = append(usernames, *githubUsers...)
	}

	// removes duplicates
	usernames = funk.Uniq(usernames).([]string)

	// ban some people ;)
	if excludeGithubUsers != nil {
		usernames = funk.Filter(usernames, func(username string) bool {
			return !funk.Contains(*excludeGithubUsers, username)
		}).([]string)
	}

	// get ssh keys from users
	var sshKeys []string
	funk.ForEach(usernames, func(username string) {
		sshKeys = append(sshKeys, getUserSSHKeys(username)...)
	})
	sshKeys = funk.Uniq(sshKeys).([]string)

	// prints ssh keys
	output(sshKeys)
}
