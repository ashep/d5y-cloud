package fwupdate

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Cronus 0.0.1-alphaN builds cannot download firmware over HTTPS: alpha1-3 verify the server
// against an outdated pinned CA set, alpha4-5 against an empty global CA store. A plain-HTTP
// download skips TLS entirely, so such devices are served 0.0.1 from an HTTP mirror and continue
// upgrading over HTTPS from there.
const (
	httpBridgeOwner  = "ashep"
	httpBridgeName   = "cronus"
	httpBridgeMirror = "http://33v.xyz/"
)

var httpBridgeTarget = semver.MustParse("0.0.1")

func needsHTTPBridge(owner, name string, v *semver.Version) bool {
	return owner == httpBridgeOwner && name == httpBridgeName &&
		v.Major() == 0 && v.Minor() == 0 && v.Patch() == 1 &&
		strings.HasPrefix(v.Prerelease(), "alpha")
}

// httpBridge returns the bridge target release with asset URLs pointing to the HTTP mirror.
//
// It returns nil if the target release is absent or the asset choice is ambiguous: 0.0.1-alpha1
// reports a bare arch without a hardware id, which matches images for every display type.
func (r ReleaseSet) httpBridge() *Release {
	for _, rel := range r.List {
		if !rel.Version.Equal(httpBridgeTarget) {
			continue
		}

		if len(rel.Assets) != 1 {
			return nil
		}

		ast := rel.Assets[0]
		ast.URL = httpBridgeMirror + ast.Name

		return &Release{Version: rel.Version, Assets: []Asset{ast}}
	}

	return nil
}
