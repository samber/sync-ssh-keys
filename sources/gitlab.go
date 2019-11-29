package sources

import (
	"context"
	"fmt"
	"log"
	"net/http"
	logger "sync-ssh-keys/logger"
	"time"

	"github.com/thoas/go-funk"
	gitlab "github.com/xanzy/go-gitlab"
)

type GitlabSource struct {
	httpClient *http.Client
	ctx        context.Context
	client     *gitlab.Client
	token      string

	Groups    []string
	Usernames []string
	Exclude   []string
}

const sleepTime = 100 * time.Millisecond // rate limit: 10 req/s/IP and 600 req/min

func NewGitlabSource(endpoint *string, token string, groups []string, usernames []string, exclude []string) *GitlabSource {
	g := &GitlabSource{
		token: token,

		Groups:    groups,
		Usernames: usernames,
		Exclude:   exclude,
	}

	g.ctx = context.Background()
	g.httpClient = http.DefaultClient
	g.client = gitlab.NewClient(g.httpClient, token)

	if endpoint != nil {
		err := g.client.SetBaseURL(*endpoint)
		if err != nil {
			log.Fatalf("Endpoint is invalid: %s\n", *endpoint)
		}
	}

	return g
}

func (g GitlabSource) GetName() string {
	return "Gitlab"
}

func (g GitlabSource) CheckInputErrors() string {
	if g.token == "" {
		return "--gitlab-token is missing.\n"
	}

	if len(g.Groups) == 0 && len(g.Usernames) == 0 {
		return "--gitlab-group or --gitlab-usernames must be provided.\n"
	}

	return ""
}

func (g GitlabSource) GetKeys() []string {
	userIDs := map[int]string{}

	// get users from groups
	for _, group := range g.Groups {
		userIDs = mapUnion(userIDs, g.getGitlabGroupMembers(group))
	}

	// get users provided in cli
	if g.Usernames != nil {
		funk.ForEach(g.Usernames, func(username string) {
			userID, err := g.getGitlabUserID(username)
			if err != nil {
				logger.Warning(err, fmt.Sprintf("[warning] Gitlab user \"%s\" not found.\n", username))
				return
			}
			userIDs[userID] = username
		})
	}

	// ban some people ;)
	if g.Exclude != nil {
		funk.ForEach(g.Exclude, func(exclude string) {
			excludedUserID, err := g.getGitlabUserID(exclude)
			if err != nil {
				logger.Warning(err, fmt.Sprintf("[warning] Gitlab user \"%s\" not found.\n", exclude))
				return
			}
			delete(userIDs, excludedUserID)
		})
	}

	// get ssh keys from users
	var sshKeys []string
	funk.ForEach(userIDs, func(userID int, username string) {
		time.Sleep(sleepTime) // slow down in order to protect against rate limits
		sshKeys = append(sshKeys, g.getUserSSHKeys(userID, username)...)
	})
	return funk.Uniq(sshKeys).([]string)
}

func (g *GitlabSource) getGitlabUserID(username string) (int, error) {
	opt := &gitlab.ListUsersOptions{
		Username: &username,
	}
	users, _, err := g.client.Users.ListUsers(opt)
	if err != nil {
		return -1, err
	}
	if len(users) != 1 {
		return -1, fmt.Errorf("Gitlab user not found: %s", username)
	}

	return users[0].ID, nil
}

func (g *GitlabSource) getGitlabGroupMembers(groupName string) map[int]string {
	// urlEncodedGroupName := url.PathEscape(groupName)

	// fetch 100 results per api call
	opt := &gitlab.ListGroupMembersOptions{
		ListOptions: gitlab.ListOptions{PerPage: 100},
	}

	var allMembers []*gitlab.GroupMember
	for {
		members, resp, err := g.client.Groups.ListAllGroupMembers(groupName, opt)
		if err != nil {
			logger.Warning(err, fmt.Sprintf("[warning] Gitlab Group or Subgroup \"%s\" not found.\n", groupName))
			break
		}

		allMembers = append(allMembers, members...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return funk.Map(allMembers, func(member *gitlab.GroupMember) (int, string) {
		return member.ID, member.Username
	}).(map[int]string)
}

func (g *GitlabSource) getUserSSHKeys(userID int, username string) []string {
	opt := &gitlab.ListSSHKeysForUserOptions{PerPage: 100}
	keys, _, err := g.client.Users.ListSSHKeysForUser(userID, opt)
	if err != nil {
		logger.Warning(err, fmt.Sprintf("[warning] Gitlab User \"%s\" not found.\n", username))
		return []string{}
	}

	return funk.Map(keys, func(key *gitlab.SSHKey) string {
		return fmt.Sprintf("%s %s@gitlab", key.Key, username)
	}).([]string)
}

func mapUnion(a map[int]string, b map[int]string) map[int]string {
	c := map[int]string{}

	for k, v := range a {
		c[k] = v
	}
	for k, v := range b {
		c[k] = v
	}

	return c
}
