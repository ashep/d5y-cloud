package update

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v63/github"
	"github.com/rs/zerolog"
)

type Asset struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Release struct {
	Version *semver.Version `json:"version"`
	Assets  []Asset         `json:"assets"`
}

type ReleaseSet []Release

// Next returns the release which version number is after v.
func (r ReleaseSet) Next(v *semver.Version) *Release {
	for i, rl := range r {
		if rl.Version.GreaterThan(v) {
			return &r[i]
		}
	}

	return nil
}

type Service struct {
	gh *github.Client
	l  zerolog.Logger
}

func New(gh *github.Client, l zerolog.Logger) *Service {
	return &Service{
		gh: gh,
		l:  l,
	}
}

func (s *Service) List(ctx context.Context, app, arch, hw string) (ReleaseSet, error) {
	res := make(ReleaseSet, 0)

	app = strings.ToLower(app)
	arch = strings.ToLower(arch)
	hw = strings.ToLower(hw)

	for page := 1; ; page++ {
		rsp, _, err := s.gh.Repositories.ListReleases(ctx, "ashep", app, &github.ListOptions{Page: page})

		ghErr := &github.ErrorResponse{}
		if errors.As(err, &ghErr) && ghErr.Response.StatusCode == http.StatusNotFound {
			return ReleaseSet{}, ErrAppNotFound
		} else if err != nil {
			return nil, fmt.Errorf("gitbhub: list releases: %w", err)
		}

		if len(rsp) == 0 {
			break
		}

		assetName := fmt.Sprintf("%s-%s-%s", app, arch, hw)

		for _, ghRel := range rsp {
			ver, err := semver.NewVersion(*ghRel.TagName)
			if err != nil {
				s.l.Error().Err(err).Msg("failed to parse a version from GitHub tag name")
				continue
			}

			rel := Release{
				Version: ver,
				Assets:  make([]Asset, 0),
			}

			for _, ast := range ghRel.Assets {
				if !strings.Contains(*ast.Name, assetName) {
					continue
				}

				rel.Assets = append(rel.Assets, Asset{
					Name: *ast.Name,
					URL:  *ast.BrowserDownloadURL,
				})
			}

			res = append(res, rel)
		}
	}

	slices.SortFunc(res, func(a, b Release) int {
		return a.Version.Compare(b.Version)
	})

	return res, nil
}
