package scm

import (
	"errors"

	"github.com/MagicSong/s2iservice/pkg/scm/provider"
)

type Gitter interface {
	ListLanguages() ([]string, error)
	SetURL(u string) error
	GetURL() string
}
type GitType int

const (
	Github GitType = iota
	Gitlab
	SVN
	BitBucket
)

func NewTokenClient(token string, url string, gitType GitType) (Gitter, error) {
	switch gitType {
	case Github:
		return provider.NewGithub(token, url)
	case Gitlab:
		return provider.NewGitlab(token, url)
	default:
		return nil, errors.New("unknow git type")
	}
}
