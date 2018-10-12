package provider

import (
	"errors"

	"github.com/MagicSong/s2iservice/pkg/scm/url"
	gitlabapi "github.com/xanzy/go-gitlab"
)

type gitlabProvider struct {
	Client *gitlabapi.Client
	GitURL *url.URL
}

func NewGitlab(token string, source string) (*gitlabProvider, error) {
	git := gitlabapi.NewClient(nil, token)
	u, err := url.Parse(source)
	if err != nil {
		return nil, err
	}
	err = git.SetBaseURL(u.URL.Scheme + "://" + u.URL.Host + "/api/v4")
	if err != nil {
		return nil, err
	}
	return &gitlabProvider{
		Client: git,
		GitURL: u,
	}, nil
}
func (g *gitlabProvider) ListLanguages() ([]string, error) {
	langs := make([]string, 0)
	if g.GitURL == nil {
		return nil, errors.New("Github URL has not been initialized")
	}
	l, _, err := g.Client.Projects.GetProjectLanguages(g.GitURL.GetProjectName())
	if err != nil {
		return nil, err
	}
	for name := range *l {
		langs = append(langs, name)
	}
	return langs, nil
}

func (g *gitlabProvider) SetURL(u string) error {
	giturl, err := url.Parse(u)
	if err != nil {
		return err
	}
	err = g.Client.SetBaseURL(giturl.URL.Scheme + "://" + giturl.URL.Host + "/api/v4")
	if err != nil {
		return err
	}
	g.GitURL = giturl
	return nil
}

func (g *gitlabProvider) GetURL() string {
	if g.GitURL == nil {
		return ""
	}
	return g.GitURL.String()
}
