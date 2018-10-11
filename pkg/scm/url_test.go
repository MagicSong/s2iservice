package scm

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

type parseTest struct {
	rawurl         string
	expectedGitURL *URL
	expectedError  bool
}

func TestParse(t *testing.T) {
	var tests []parseTest

	tests = append(tests,
		// http://
		parseTest{
			rawurl: "http://user:pass@github.com:443/user/repo.git?query#fragment",
			expectedGitURL: &URL{
				URL: url.URL{
					Scheme:   "http",
					User:     url.UserPassword("user", "pass"),
					Host:     "github.com:443",
					Path:     "/user/repo.git",
					RawQuery: "query",
					Fragment: "fragment",
				},
				Type:         URLTypeURL,
				Organization: "user",
				Repository:   "repo",
			},
		},
		parseTest{
			rawurl: "http://user@1.2.3.4:443/repo?query#fragment",
			expectedGitURL: &URL{
				URL: url.URL{
					Scheme:   "http",
					User:     url.User("user"),
					Host:     "1.2.3.4:443",
					Path:     "/repo",
					RawQuery: "query",
					Fragment: "fragment",
				},
				Type:       URLTypeURL,
				Repository: "repo",
			},
		},
		parseTest{
			rawurl: "http://[::ffff:1.2.3.4]:443",
			expectedGitURL: &URL{
				URL: url.URL{
					Scheme: "http",
					Host:   "[::ffff:1.2.3.4]:443",
				},
				Type: URLTypeURL,
			},
		},
		parseTest{
			rawurl: "http://github.com/openshift/origin",
			expectedGitURL: &URL{
				URL: url.URL{
					Scheme: "http",
					Host:   "github.com",
					Path:   "/openshift/origin",
				},
				Type:         URLTypeURL,
				Organization: "openshift",
				Repository:   "origin",
			},
		},

		// git@host ...
		parseTest{
			rawurl: "user@github.com:/user/repo.git#fragment",
			expectedGitURL: &URL{
				URL: url.URL{
					User:     url.User("user"),
					Host:     "github.com",
					Path:     "/user/repo.git",
					Fragment: "fragment",
				},
				Type:         URLTypeSCP,
				Organization: "user",
				Repository:   "repo",
			},
		},
		parseTest{
			rawurl: "user@github.com:user/repo.git#fragment",
			expectedGitURL: &URL{
				URL: url.URL{
					User:     url.User("user"),
					Host:     "github.com",
					Path:     "user/repo.git",
					Fragment: "fragment",
				},
				Type:         URLTypeSCP,
				Organization: "user",
				Repository:   "repo",
			},
		},
		parseTest{
			rawurl: "user@1.2.3.4:repo#fragment",
			expectedGitURL: &URL{
				URL: url.URL{
					User:     url.User("user"),
					Host:     "1.2.3.4",
					Path:     "repo",
					Fragment: "fragment",
				},
				Type:       URLTypeSCP,
				Repository: "repo",
			},
		},
		parseTest{
			rawurl: "[::ffff:1.2.3.4]:",
			expectedGitURL: &URL{
				URL: url.URL{
					Host: "[::ffff:1.2.3.4]",
				},
				Type: URLTypeSCP,
			},
		},
		parseTest{
			rawurl: "git@github.com:openshift/origin",
			expectedGitURL: &URL{
				URL: url.URL{
					User: url.User("git"),
					Host: "github.com",
					Path: "openshift/origin",
				},
				Type:         URLTypeSCP,
				Organization: "openshift",
				Repository:   "origin",
			},
		},

		// path ...
		parseTest{
			rawurl: "/absolute#fragment",
			expectedGitURL: &URL{
				URL: url.URL{
					Path:     "/absolute",
					Fragment: "fragment",
				},
				Type: URLTypeLocal,
			},
		},
		parseTest{
			rawurl: "relative#fragment",
			expectedGitURL: &URL{
				URL: url.URL{
					Path:     "relative",
					Fragment: "fragment",
				},
				Type: URLTypeLocal,
			},
		},
	)

	for _, test := range tests {
		parsedURL, err := Parse(test.rawurl)
		if test.expectedError != (err != nil) {
			t.Errorf("%s: Parse() returned err: %v", test.rawurl, err)
		}
		if err != nil {
			continue
		}

		if !reflect.DeepEqual(parsedURL, test.expectedGitURL) {
			t.Errorf("%s: Parse() returned\n\t%#v\nWanted\n\t%#v", test.rawurl, parsedURL, test.expectedGitURL)
		}

		if parsedURL.String() != test.rawurl {
			t.Errorf("%s: String() returned %s", test.rawurl, parsedURL.String())
		}

		if parsedURL.StringNoFragment() != strings.SplitN(test.rawurl, "#", 2)[0] {
			t.Errorf("%s: StringNoFragment() returned %s", test.rawurl, parsedURL.StringNoFragment())
		}
	}
}

func TestStringNoFragment(t *testing.T) {
	u := MustParse("part#fragment")
	if u.StringNoFragment() != "part" {
		t.Errorf("StringNoFragment() returned %s", u.StringNoFragment())
	}
	if !reflect.DeepEqual(u, MustParse("part#fragment")) {
		t.Errorf("StringNoFragment() modified its argument")
	}
}
