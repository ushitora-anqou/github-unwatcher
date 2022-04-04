package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

func listWatchedRepos(client *github.Client, ctx context.Context) ([]*github.Repository, error) {
	opt := &github.ListOptions{PerPage: 100 /* max */}
	var watchedRepos []*github.Repository
	for {
		repos, resp, err := client.Activity.ListWatched(ctx, "", opt)
		if err != nil {
			return nil, err
		}
		watchedRepos = append(watchedRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return watchedRepos, nil
}

func printWatchedRepos(client *github.Client, ctx context.Context) error {
	watchedRepos, err := listWatchedRepos(client, ctx)
	for i, repo := range watchedRepos {
		fmt.Printf("%d\t%s/%s\n", i, *repo.Owner.Login, *repo.Name)
	}
	return err
}

func askYN(def bool) bool {
	var line string
	_, err := fmt.Scanln(&line)
	if err != nil {
		return false
	}
	if line == "" {
		return def
	}
	if strings.ToLower(line[0:1]) == "y" {
		return true
	}
	return false
}

func unwatchRepos(client *github.Client, ctx context.Context, org string) error {
	watchedRepos, err := listWatchedRepos(client, ctx)
	if err != nil {
		return err
	}

	var reposToBeUnwatched []*github.Repository
	for _, repo := range watchedRepos {
		if *repo.Owner.Login != org {
			continue
		}
		reposToBeUnwatched = append(reposToBeUnwatched, repo)
	}

	if len(reposToBeUnwatched) == 0 {
		fmt.Printf("No repos to be unwatched found\n")
		return nil
	} else {
		fmt.Printf("Unwatching following repos:\n")
		for _, repo := range reposToBeUnwatched {
			fmt.Printf("\t%s/%s\n", org, *repo.Name)
		}
	}

	fmt.Printf("Are you sure? [y/N]: ")
	if !askYN(false) {
		fmt.Printf("Canceled\n")
		return nil
	}

	for _, repo := range reposToBeUnwatched {
		fmt.Printf("Unwatching %s/%s\n", org, *repo.Name)
		_, err := client.Activity.DeleteRepositorySubscription(ctx, org, *repo.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	// Usage: GITHUB_ACCESS_TOKEN="..." $0
	//     List watched repos
	// Usage: GITHUB_ACCESS_TOKEN="..." $0 ORGANIZATION-NAME
	//     Unwatch watching repos of the organization

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	var err error
	if len(os.Args) == 1 {
		fmt.Printf("Listing watched repos...\n")
		err = printWatchedRepos(client, ctx)
	} else {
		fmt.Printf("Unwatch repos of organization %s\n", os.Args[1])
		err = unwatchRepos(client, ctx, os.Args[1])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}
}
