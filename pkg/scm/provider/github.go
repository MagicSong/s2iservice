package provider

import (
	"context"
	"errors"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	giturl "github.com/MagicSong/s2iservice/pkg/scm/url"
)

const GithubHost = "github.com"

type githubProvider struct {
	URL    *giturl.URL
	Client *github.Client
}

func NewGithub(token string, url string) (*githubProvider, error) {
	g := new(githubProvider)
	err := g.SetURL(url)
	if err != nil {
		return nil, err
	}
	g.Client = openGithubClient(token)
	return g, nil
}
func (g *githubProvider) ListLanguages() ([]string, error) {
	langs := make([]string, 0)
	if g.URL == nil {
		return nil, errors.New("Github URL has not been initialized")
	}
	m, _, err := g.Client.Repositories.ListLanguages(context.Background(), g.URL.Organization, g.URL.Repository)
	if err != nil {
		return nil, err
	}
	for name := range m {
		langs = append(langs, name)
	}
	return langs, nil
}

func (g *githubProvider) SetURL(u string) error {
	url, err := giturl.Parse(u)
	if err != nil {
		return err
	}
	if url.URL.Host != GithubHost {
		return errors.New("URL doesn't belong to github")
	}
	g.URL = url
	return nil
}

func (g *githubProvider) GetURL() string {
	if g.URL == nil {
		return ""
	}
	return g.URL.String()
}

func openGithubClient(token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	return client
}
