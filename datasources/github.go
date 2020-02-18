package datasources

import (
	"context"
	"fmt"
	"log"
	"net/http"

	logger "github.com/samber/sync-ssh-keys/logger"

	"github.com/google/go-github/v28/github"
	"github.com/thoas/go-funk"
	"golang.org/x/oauth2"
)

type GithubSource struct {
	httpClient *http.Client
	ctx        context.Context
	client     *github.Client
	token      *string

	Org       *string
	Teams     []string
	Usernames []string
	Exclude   []string
}

func NewGithubSource(endpoint *string, token *string, org *string, teams []string, usernames []string, exclude []string) *GithubSource {
	g := &GithubSource{
		token: token,

		Org:       org,
		Teams:     teams,
		Usernames: usernames,
		Exclude:   exclude,
	}

	g.ctx = context.Background()

	if token != nil {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *token},
		)
		g.httpClient = oauth2.NewClient(g.ctx, ts)
	}
	if endpoint != nil {
		var err error
		g.client, err = github.NewEnterpriseClient(*endpoint, *endpoint, g.httpClient)
		if err != nil {
			log.Fatal("Failed to connect to Github Enterprise")
		}
	} else {
		g.client = github.NewClient(g.httpClient)
	}

	return g
}

func (g GithubSource) GetName() string {
	return "Github"
}

func (g GithubSource) CheckInputErrors() string {
	if g.Org == nil && len(g.Teams) == 0 && len(g.Usernames) == 0 {
		return "--github-org, --github-team or --github-username must be provided.\n"
	}
	if len(g.Teams) > 0 && g.Org == nil {
		return "--github-team cannot be provided without --github-org.\n"
	}
	if len(g.Teams) > 0 && g.token == nil {
		return "--github-team cannot be provided without --github-token.\n"
	}

	// warnings
	if g.Org != nil && g.token == nil {
		return "[warning] You provided --github-org without --github-token: organization private members won't be fetched.\n"
	}
	return ""
}

func (g GithubSource) GetKeys() []string {
	var usernames []string

	// get users from teams
	if g.Org != nil {
		if len(g.Teams) > 0 {
			for _, team := range g.Teams {
				usernames = append(usernames, g.getGithubTeamMembers(team)...)
			}
		} else {
			usernames = append(usernames, g.getGithubOrgMembers()...)
		}
	}

	// get users provided in cli
	if g.Usernames != nil {
		usernames = append(usernames, g.Usernames...)
	}

	// removes duplicates
	usernames = funk.Uniq(usernames).([]string)

	// ban some people ;)
	if g.Exclude != nil {
		usernames = funk.Filter(usernames, func(username string) bool {
			return !funk.Contains(g.Exclude, username)
		}).([]string)
	}

	// get ssh keys from users
	var sshKeys []string
	funk.ForEach(usernames, func(username string) {
		sshKeys = append(sshKeys, g.getUserSSHKeys(username)...)
	})
	return funk.Uniq(sshKeys).([]string)
}

func (g *GithubSource) getGithubOrgMembers() []string {
	// fetch 100 results per api call
	opt := &github.ListMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allUsers []*github.User
	for {
		users, resp, err := g.client.Organizations.ListMembers(g.ctx, *g.Org, opt)
		if err != nil {
			logger.Warning(err, fmt.Sprintf("[warning] Github Organisation \"%s\" not found", *g.Org))
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

func (g *GithubSource) getGithubTeamMembers(teamName string) []string {
	// get team id from slug
	team, _, err := g.client.Teams.GetTeamBySlug(g.ctx, *g.Org, teamName)
	if err != nil {
		logger.Warning(err, fmt.Sprintf("[warning] Github Team \"%s\" not found.\n", teamName))
		return []string{}
	}

	// fetch 100 results per api call
	opt := &github.TeamListTeamMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allUsers []*github.User
	for {
		users, resp, err := g.client.Teams.ListTeamMembers(g.ctx, team.GetID(), opt)
		if err != nil {
			logger.Warning(err, fmt.Sprintf("[warning] Github Team \"%s\" not found.\n", teamName))
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

func (g *GithubSource) getUserSSHKeys(username string) []string {
	opt := &github.ListOptions{PerPage: 100}
	keys, _, err := g.client.Users.ListKeys(g.ctx, username, opt)
	if err != nil {
		logger.Warning(err, fmt.Sprintf("[warning] Github User \"%s\" not found.\n", username))
		return []string{}
	}

	return funk.Map(keys, func(sshKey *github.Key) string {
		return fmt.Sprintf("%s %s@github", sshKey.GetKey(), username)
	}).([]string)
}
