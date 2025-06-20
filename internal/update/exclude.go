package update

import (
	"github.com/Masterminds/semver/v3"
)

type excludeItem struct {
	Owner   string
	Name    string
	Version *semver.Version
}

type excludeList []excludeItem

func (e excludeList) Contains(owner, name string, version *semver.Version) bool {
	for _, item := range e {
		if item.Owner == owner && item.Name == name && item.Version.Equal(version) {
			return true
		}
	}
	return false
}

// excVerSrc contains versions that cannot be upgraded from.
var excVerSrc = excludeList{
	{
		Owner:   "ashep",
		Name:    "cronus",
		Version: semver.MustParse("0.0.1-alpha1"),
	},
	{
		Owner:   "ashep",
		Name:    "cronus",
		Version: semver.MustParse("0.0.1-alpha2"),
	},
	{
		Owner:   "ashep",
		Name:    "cronus",
		Version: semver.MustParse("0.0.1-alpha3"),
	},
	{
		Owner:   "ashep",
		Name:    "cronus",
		Version: semver.MustParse("0.0.1-alpha4"),
	},
	{
		Owner:   "ashep",
		Name:    "cronus",
		Version: semver.MustParse("0.0.1-alpha5"),
	},
}

// excVerDst contains versions that cannot be upgraded to.
var excVerDst = excludeList{
	{
		Owner:   "ashep",
		Name:    "cronus",
		Version: semver.MustParse("0.0.1-alpha2"),
	},
	{
		Owner:   "ashep",
		Name:    "cronus",
		Version: semver.MustParse("0.0.1-alpha3"),
	},
	{
		Owner:   "ashep",
		Name:    "cronus",
		Version: semver.MustParse("0.0.1-alpha4"),
	},
	{
		Owner:   "ashep",
		Name:    "cronus",
		Version: semver.MustParse("0.0.1-alpha5"),
	},
}
