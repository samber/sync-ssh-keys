package main

import (
	"sync-ssh-keys/sources"
	src "sync-ssh-keys/sources"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	Version string
	Build   string
)

var (
	outputPath = kingpin.Flag("output", "Write output to <file>. Default to stdout").Short('o').String()
	wError     = kingpin.Flag("Werror", "Treat warning as errors. Fatal error if organization, team or user does not exist.").String()

	localPath = kingpin.Flag("local-path", "Path to a local authorized_keys file. It can be useful in case of network failure ;)").String()

	githubEndpoint         = kingpin.Flag("github-endpoint", "Github Enterprise endpoint.").Envar("GITHUB_ENDPOINT").String()
	githubToken            = kingpin.Flag("github-token", "Github personal token.").Envar("GITHUB_TOKEN").String()
	githubOrg              = kingpin.Flag("github-org", "Github organization name.").String()
	githubTeams            = kingpin.Flag("github-team", "Team(s) allowed to access server.").Strings()
	githubUsernames        = kingpin.Flag("github-username", "Username(s) allowed to access server.").Strings()
	excludeGithubUsernames = kingpin.Flag("exclude-github-username", "Username(s) to explicitly exclude.").Strings()

	gitlabEndpoint         = kingpin.Flag("gitlab-endpoint", "Gitlab endpoint.").Envar("GITLAB_ENDPOINT").String()
	gitlabToken            = kingpin.Flag("gitlab-token", "Gitlab personal token.").Envar("GITLAB_TOKEN").String()
	gitlabGroups           = kingpin.Flag("gitlab-group", "Group allowed to access server.").Strings()
	gitlabUsernames        = kingpin.Flag("gitlab-username", "Username(s) allowed to access server.").Strings()
	excludeGitlabUsernames = kingpin.Flag("exclude-gitlab-username", "Username(s) to explicitly exclude.").Strings()
)

func argvToNullable() {
	// main
	if outputPath != nil && len(*outputPath) == 0 {
		outputPath = nil
	}

	// local
	if localPath != nil && len(*localPath) == 0 {
		localPath = nil
	}

	// github
	if githubEndpoint != nil && len(*githubEndpoint) == 0 {
		githubEndpoint = nil
	}
	if githubToken != nil && len(*githubToken) == 0 {
		githubToken = nil
	}
	if githubOrg != nil && len(*githubOrg) == 0 {
		githubOrg = nil
	}

	// gitlab
	if gitlabEndpoint != nil && len(*gitlabEndpoint) == 0 {
		gitlabEndpoint = nil
	}
}

func checkInputErrors(srcs []sources.Source) {
	if len(srcs) == 0 {
		kingpin.FatalUsage("Please provide a key source.")
	}

	// Check data source errors
	for _, src := range srcs {
		if err := src.CheckInputErrors(); err != "" {
			kingpin.FatalUsage(err)
		}
	}
}

func main() {
	kingpin.Version(Version + "-" + Build)
	kingpin.Parse()

	argvToNullable()

	srcs := []sources.Source{}

	// Init Local key source
	if localPath != nil {
		srcs = append(srcs, src.NewLocalSource(
			*localPath,
		))
	}
	// Init Github key source
	if githubOrg != nil || len(*githubTeams) > 0 || len(*githubUsernames) > 0 || len(*excludeGithubUsernames) > 0 {
		srcs = append(srcs, src.NewGithubSource(
			githubEndpoint,
			githubToken,
			githubOrg,
			*githubTeams,
			*githubUsernames,
			*excludeGithubUsernames,
		))
	}
	// Init Gitlab key source
	if len(*gitlabGroups) > 0 || len(*gitlabUsernames) > 0 || len(*excludeGitlabUsernames) > 0 {
		srcs = append(srcs, src.NewGitlabSource(
			gitlabEndpoint,
			*gitlabToken,
			*gitlabGroups,
			*gitlabUsernames,
			*excludeGitlabUsernames,
		))
	}

	checkInputErrors(srcs)

	// fetch keys 🚀
	keys := map[string][]string{}
	for _, src := range srcs {
		keys[src.GetName()] = src.GetKeys()
	}

	// prints ssh keys
	PrintKeys(outputPath, keys)
}
