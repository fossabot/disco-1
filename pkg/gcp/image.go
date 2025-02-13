package gcp

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

const (
	partNameTag      = 2
	partRegionalName = 3
	partGlobalName   = 2
	partCountGCR     = 3
	partCountAR      = 4

	arNameSuffix = "-docker"
)

// ImageInfo represents parsed GCP registry image (GCR or AR).
type ImageInfo struct {
	IsGCR    bool   `json:"-"`
	IsAR     bool   `json:"-"`
	IsLatest bool   `json:"isLatest"`
	Deployed string `json:"deployed"`
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	Digest   string `json:"digest"`
	Region   string `json:"region"`
	Registry string `json:"registry"`
	Project  string `json:"project"`
	Folder   string `json:"registryFolder"`
}

func (i *ImageInfo) withPrefix(v string) string {
	// gcr.io/cloudy-demos/hello-broken@sha256:0900c08e7
	if i.IsGCR {
		if i.Digest != "" {
			return fmt.Sprintf("%s%s/%s/%s@%s", v, i.Registry, i.Project, i.Name, i.Digest)
		}
		return fmt.Sprintf("%s%s/%s/%s", v, i.Registry, i.Project, i.Name)
	}

	// us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator@sha256:b4a094e55244bc
	if i.IsAR {
		if i.Digest != "" {
			return fmt.Sprintf("%s%s/%s/%s/%s@%s", v, i.Registry, i.Project, i.Folder, i.Name, i.Digest)
		}
		return fmt.Sprintf("%s%s/%s/%s/%s", v, i.Registry, i.Project, i.Folder, i.Name)
	}
	return ""
}

func (i *ImageInfo) URI() string {
	return i.withPrefix("")
}

func (i *ImageInfo) URL() string {
	return i.withPrefix("https://")
}

// ManifestURL returns manifest URL for the image.
func (i *ImageInfo) ManifestURL() string {
	tag := i.Tag
	if tag == "" {
		tag = "latest"
	}

	// https://gcr.io/v2/cloudy-demos/hello-broken/manifests/latest
	if i.IsGCR {
		return fmt.Sprintf("https://%s/v2/%s/%s/manifests/%s",
			i.Registry, i.Project, i.Name, tag)
	}

	// https://us-west1-docker.pkg.dev/v2/cloudy-demos/artomator/artomator/manifests/v0.8.3
	if i.IsAR {
		return fmt.Sprintf("https://%s/v2/%s/%s/%s/manifests/%s",
			i.Registry, i.Project, i.Folder, i.Name, tag)
	}
	return ""
}

// ParseImageInfo parses image string into ImageInfo struct.
// Supported formats:
//   - gcr.io/cloudy-demos/hello-broken
//   - gcr.io/cloudy-demos/hello-broken:latest
//   - gcr.io/cloudy-demos/hello-broken@sha256:1234567890
//   - gcr.io/cloudy-demos/hello-broken:v0.8.3
//   - us.gcr.io/cloudy-demos/hello-broken:v0.8.3
//   - us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator
//   - us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator:latest
//   - us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator:v0.8.3
//   - us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator@sha256:1234567890
//   - us-docker.pkg.dev/cloudy-demos/test-multiregion
func ParseImageInfo(uri string) (*ImageInfo, error) {
	if uri == "" {
		return nil, errors.New("image is empty")
	}

	parts := strings.Split(uri, "/")
	if len(parts) < partCountGCR || len(parts) > partCountAR {
		return nil, errors.Errorf("invalid image URI: %s", uri)
	}

	info := &ImageInfo{
		Deployed: uri,
	}

	// gcr.io/cloudy-demos/hello-broken
	// gcr.io/cloudy-demos/hello-broken:latest
	// gcr.io/cloudy-demos/hello-broken@sha256:1234567890
	// gcr.io/cloudy-demos/hello-broken:v0.8.3
	if len(parts) == partCountGCR {
		info.IsGCR = true
		if !parseRegistryAndRegion(parts[0], info) {
			return nil, errors.Errorf("error parsing registry and region: %s", uri)
		}
		info.Project = parts[1]
		if !parseName(parts[2], info) {
			return nil, errors.Errorf("invalid GCR image URI: %s", uri)
		}
		return info, nil
	}

	// us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator
	// us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator:latest
	// us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator:v0.8.3
	// us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator@sha256:1234567890
	// us-docker.pkg.dev/cloudy-demos/test-multiregion
	if len(parts) == partCountAR {
		info.IsAR = true
		if !parseRegistryAndRegion(parts[0], info) {
			return nil, errors.Errorf("error parsing registry and region: %s", uri)
		}
		info.Project = parts[1]
		info.Folder = parts[2]
		if !parseName(parts[3], info) {
			return nil, errors.Errorf("error parsing image name: %s", uri)
		}
		return info, nil
	}

	return nil, errors.Errorf("invalid image URI: %s", uri)
}

// parseName parses image name into ImageInfo struct.
// Supported formats:
// - artomator:latest
// - artomator:v0.8.3
// - artomator@sha256:1234567890.
func parseName(name string, info *ImageInfo) bool {
	if !strings.Contains(name, "@") && !strings.Contains(name, ":") {
		info.Name = name
		return true
	}

	if strings.Contains(name, "@") {
		parts := strings.Split(name, "@")
		info.Name = parts[0]
		info.Digest = parts[1]
		return true
	}

	parts := strings.Split(name, ":")
	if len(parts) >= partNameTag {
		info.Name = parts[0]
		if parts[1] == "latest" {
			info.IsLatest = true
		} else {
			info.Tag = parts[1]
		}
		return true
	}

	return false
}

// parseRegistryAndRegion parses registry and region into ImageInfo struct.
// - us-west1-docker.pkg.dev
// - us-docker.pkg.dev
// - us.gcr.io
// - gcr.io.
func parseRegistryAndRegion(uri string, info *ImageInfo) bool {
	if uri == "" || info == nil {
		return false
	}

	info.Registry = uri

	parts := strings.Split(uri, ".")

	// gcr.io
	if len(parts) == partGlobalName {
		return true
	}

	// us.gcr.io
	if info.IsGCR && len(parts) == partRegionalName {
		info.Region = parts[0]
		return true
	}

	// us-west1-docker.pkg.dev
	// us-docker.pkg.dev
	// us.gcr.io
	if info.IsAR && len(parts) == partRegionalName {
		info.Region = strings.ReplaceAll(parts[0], arNameSuffix, "")
		return true
	}

	return false
}
