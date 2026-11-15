package fwupdate

import (
	"testing"

	"github.com/Masterminds/semver/v3"
)

func testRelease(ver string, assetNames ...string) Release {
	rel := Release{Version: semver.MustParse(ver)}
	for _, n := range assetNames {
		rel.Assets = append(rel.Assets, Asset{
			Name:   n,
			Size:   len(n),
			SHA256: "sha-" + n,
			URL:    "https://github.com/ashep/cronus/releases/download/" + ver + "/" + n,
		})
	}

	return rel
}

func TestNextHTTPBridge(t *testing.T) {
	const max7219 = "cronus-0.0.1-esp32s3_max7219v32x16.bin"

	set := ReleaseSet{
		Owner: "ashep",
		Name:  "cronus",
		List: []Release{
			testRelease("0.0.1-alpha9", "cronus-0.0.1-alpha9-esp32s3_max7219v32x16.bin"),
			testRelease("0.0.1", max7219),
			testRelease("0.0.2", "cronus-0.0.2-esp32s3_max7219v32x16.bin"),
		},
	}

	for _, cur := range []string{"0.0.1-alpha1", "0.0.1-alpha5", "0.0.1-alpha9"} {
		t.Run(cur, func(t *testing.T) {
			got := set.Next(semver.MustParse(cur))
			if got == nil {
				t.Fatal("got nil release")
			}

			if !got.Version.Equal(semver.MustParse("0.0.1")) {
				t.Errorf("version = %s, want 0.0.1", got.Version)
			}

			if len(got.Assets) != 1 {
				t.Fatalf("got %d assets, want 1", len(got.Assets))
			}

			ast := got.Assets[0]
			if want := "http://33v.xyz/" + max7219; ast.URL != want {
				t.Errorf("url = %q, want %q", ast.URL, want)
			}

			if ast.Name != max7219 || ast.Size != len(max7219) || ast.SHA256 != "sha-"+max7219 {
				t.Errorf("asset metadata changed: %+v", ast)
			}
		})
	}

	// The source release set must stay untouched.
	if u := set.List[1].Assets[0].URL; u != "https://github.com/ashep/cronus/releases/download/0.0.1/"+max7219 {
		t.Errorf("original asset url modified: %q", u)
	}
}

func TestNextHTTPBridgeNotApplied(t *testing.T) {
	list := []Release{
		testRelease("0.0.1", "cronus-0.0.1-esp32s3_max7219v32x16.bin"),
		testRelease("0.0.2", "cronus-0.0.2-esp32s3_max7219v32x16.bin"),
	}

	t.Run("stable source version", func(t *testing.T) {
		got := ReleaseSet{Owner: "ashep", Name: "cronus", List: list}.Next(semver.MustParse("0.0.1"))
		if got == nil || got.Version.String() != "0.0.2" || got.Assets[0].URL[:8] != "https://" {
			t.Fatalf("got %+v, want 0.0.2 over https", got)
		}
	})

	t.Run("other app", func(t *testing.T) {
		got := ReleaseSet{Owner: "ashep", Name: "other", List: list}.Next(semver.MustParse("0.0.1-alpha5"))
		if got == nil || got.Version.String() != "0.0.1" || got.Assets[0].URL[:8] != "https://" {
			t.Fatalf("got %+v, want 0.0.1 over https", got)
		}
	})

	t.Run("bridge target missing", func(t *testing.T) {
		set := ReleaseSet{Owner: "ashep", Name: "cronus", List: list[1:]}
		if got := set.Next(semver.MustParse("0.0.1-alpha5")); got != nil {
			t.Fatalf("got %+v, want nil", got)
		}
	})

	t.Run("ambiguous assets", func(t *testing.T) {
		set := ReleaseSet{Owner: "ashep", Name: "cronus", List: []Release{
			testRelease("0.0.1", "cronus-0.0.1-esp32s3_max7219v32x16.bin", "cronus-0.0.1-esp32s3_ws2812v32x16.bin"),
		}}
		if got := set.Next(semver.MustParse("0.0.1-alpha2")); got != nil {
			t.Fatalf("got %+v, want nil", got)
		}
	})
}

func TestNextHTTPBridgeAlpha1PicksMAX7219(t *testing.T) {
	const max7219 = "cronus-0.0.1-esp32s3_max7219v32x16.bin"

	// alpha1 reports a bare "esp32s3" arch, so every display type's image matches it.
	set := ReleaseSet{Owner: "ashep", Name: "cronus", List: []Release{
		testRelease("0.0.1", "cronus-0.0.1-esp32s3_ws2812v32x16.bin", max7219),
	}}

	got := set.Next(semver.MustParse("0.0.1-alpha1"))
	if got == nil || len(got.Assets) != 1 {
		t.Fatalf("got %+v, want a single asset", got)
	}

	if want := "http://33v.xyz/" + max7219; got.Assets[0].URL != want {
		t.Errorf("url = %q, want %q", got.Assets[0].URL, want)
	}
}
