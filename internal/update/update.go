package update

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v63/github"
	"github.com/rs/zerolog"
)

type Asset struct {
	Name   string `json:"name"`
	Size   int    `json:"size"`
	SHA256 string `json:"sha256"`
	URL    string `json:"url"`
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
	gh            *github.Client
	checkSumCache map[string]string
	l             zerolog.Logger
}

func New(gh *github.Client, l zerolog.Logger) *Service {
	return &Service{
		gh:            gh,
		checkSumCache: make(map[string]string),
		l:             l,
	}
}

// List returns all available assets for all releases sorted by version in ascending order.
//
// Only assets named `{name}-{arch}*` are returned.
//
// `incAlpha` arg controls whether assets named `*-alpha*` are returned.
func (s *Service) List(
	ctx context.Context,
	repoOwner string,
	repoName string,
	arch string,
	incAlpha bool,
) (ReleaseSet, error) {
	res := make(ReleaseSet, 0)

	arch = strings.ReplaceAll(strings.ToLower(arch), "-", "_")

	for page := 1; ; page++ {
		rsp, _, err := s.gh.Repositories.ListReleases(ctx, repoOwner, repoName, &github.ListOptions{Page: page})

		ghErr := &github.ErrorResponse{}
		if errors.As(err, &ghErr) && ghErr.Response.StatusCode == http.StatusNotFound {
			return ReleaseSet{}, ErrAppNotFound
		} else if err != nil {
			return nil, fmt.Errorf("gitbhub: list releases: %w", err)
		}

		if len(rsp) == 0 {
			break
		}

		for _, ghRel := range rsp {
			s.l.Debug().
				Str("repo", repoOwner+"/"+repoName).
				Str("tag_name", ghRel.GetTagName()).
				Msg("found release tag")

			ver, err := semver.NewVersion(ghRel.GetTagName())
			if err != nil {
				s.l.Error().
					Str("repo", repoOwner+"/"+repoName).
					Str("tag_name", ghRel.GetTagName()).
					Err(err).
					Msg("failed to parse a version from GitHub tag name")
				continue
			}

			rel := Release{
				Version: ver,
				Assets:  make([]Asset, 0),
			}

			for _, ast := range ghRel.Assets {
				if strings.HasSuffix(ast.GetName(), ".sha256") {
					continue
				}

				if !strings.HasPrefix(ast.GetName(), repoName) {
					s.l.Debug().
						Str("repo", repoOwner+"/"+repoName).
						Str("tag_name", ghRel.GetTagName()).
						Str("asset_name", ast.GetName()).
						Msg("skip asset: name does not match app")
					continue
				}

				if !strings.Contains(ast.GetName(), arch) {
					s.l.Debug().
						Str("repo", repoOwner+"/"+repoName).
						Str("tag_name", ghRel.GetTagName()).
						Str("asset_name", ast.GetName()).
						Str("arch", arch).
						Msg("skip asset: name does not match arch")
					continue
				}

				if strings.Contains(ast.GetName(), "-alpha") && !incAlpha {
					s.l.Debug().
						Str("repo", repoOwner+"/"+repoName).
						Str("tag_name", ghRel.GetTagName()).
						Str("asset_name", ast.GetName()).
						Msg("skip asset: alpha releases is not allowed")
					continue
				}

				s.l.Debug().
					Str("repo", repoOwner+"/"+repoName).
					Str("tag_name", ghRel.GetTagName()).
					Str("asset_name", ast.GetName()).
					Msg("found asset")

				rel.Assets = append(rel.Assets, Asset{
					Name:   ast.GetName(),
					Size:   ast.GetSize(),
					SHA256: s.assetChecksum(ctx, ast.GetBrowserDownloadURL()),
					URL:    ast.GetBrowserDownloadURL(),
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

func (s *Service) assetChecksum(ctx context.Context, url string) string {
	url += ".sha256"

	if v, ok := s.checkSumCache[url]; ok {
		return v
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		s.l.Error().Err(err).Msg("failed to create a request for asset checksum")
		return ""
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		s.l.Error().Err(err).Str("url", url).Msg("failed to fetch asset checksum")
		return ""
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		s.l.Error().Str("url", url).Int("code", res.StatusCode).Msg("asset checksum not found")
		return ""
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		s.l.Error().Err(err).Str("url", url).Msg("failed to read asset checksum")
		return ""
	}

	if len(b) < 64 {
		s.l.Error().Err(err).Str("url", url).Int("len", len(b)).Msg("asset checksum is too short")
		return ""
	}

	v := string(b)[:64]
	s.checkSumCache[url] = v

	return v
}
