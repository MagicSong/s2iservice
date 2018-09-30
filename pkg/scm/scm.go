package scm

import (
	"github.com/s2iservice/pkg/build"
	"github.com/s2iservice/pkg/errors"
	"github.com/s2iservice/pkg/scm/downloaders/empty"
	"github.com/s2iservice/pkg/scm/downloaders/file"
	gitdownloader "github.com/s2iservice/pkg/scm/downloaders/git"
	"github.com/s2iservice/pkg/scm/git"
	"github.com/s2iservice/pkg/utils/cmd"
	"github.com/s2iservice/pkg/utils/fs"
	utilglog "github.com/s2iservice/pkg/utils/glog"
)

var glog = utilglog.StderrLog

// DownloaderForSource determines what SCM plugin should be used for downloading
// the sources from the repository.
func DownloaderForSource(fs fs.FileSystem, s *git.URL, forceCopy bool) (build.Downloader, error) {
	glog.V(4).Infof("DownloadForSource %s", s)

	if s == nil {
		return &empty.Noop{}, nil
	}

	if s.IsLocal() {
		if forceCopy {
			return &file.File{FileSystem: fs}, nil
		}

		isLocalNonBareGitRepo, err := git.IsLocalNonBareGitRepository(fs, s.LocalPath())
		if err != nil {
			return nil, err
		}
		if !isLocalNonBareGitRepo {
			return &file.File{FileSystem: fs}, nil
		}

		isEmpty, err := git.LocalNonBareGitRepositoryIsEmpty(fs, s.LocalPath())
		if err != nil {
			return nil, err
		}
		if isEmpty {
			return nil, errors.NewEmptyGitRepositoryError(s.LocalPath())
		}

		if !git.HasGitBinary() {
			return &file.File{FileSystem: fs}, nil
		}
	}

	return &gitdownloader.Clone{Git: git.New(fs, cmd.NewCommandRunner()), FileSystem: fs}, nil
}
