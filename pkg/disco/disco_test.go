package disco

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/mchmarny/disco/pkg/gcp"
	"github.com/mchmarny/disco/pkg/scanner"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func setTestImplementations() {
	getProjectsFunc = getTestProjects
	getLocationsFunc = getTestLocations
	getServicesFunc = getTestServices
	getImageInfoFunc = getTestImageInfo
	getCVEVulnsFunc = getTestCVEVulns
	getImageVulnsFunc = getTestImageVulns
	isAPIEnabledFunc = isTestAPIEnabled

	scanner.ScanVulnerability = func(digest, path string) *exec.Cmd {
		return exec.Command("cp", "../../etc/test-license.json", path) //nolint
	}

	scanner.ScanLicense = func(digest, path string) *exec.Cmd {
		return exec.Command("cp", "../../etc/test-vuln.json", path) //nolint
	}
}

func TestImage(t *testing.T) {
	setTestImplementations()
	ctx := context.Background()

	err := DiscoverImages(ctx, nil)
	assert.Error(t, err, "error discovering images with nil query")

	err = DiscoverImages(ctx, &ImagesQuery{})
	assert.NoError(t, err, "error discovering images")

	err = DiscoverImages(ctx, &ImagesQuery{
		SimpleQuery: SimpleQuery{
			ProjectID: "test-project",
			OutputFmt: ParseOutputFormatOrDefault("yaml"),
		},
	})
	assert.NoError(t, err, "error discovering images with project")

	err = DiscoverImages(ctx, &ImagesQuery{
		OnlyDigest: true,
	})
	assert.NoError(t, err, "error discovering images with digests only")
}

func TestLicense(t *testing.T) {
	setTestImplementations()
	ctx := context.Background()

	err := DiscoverLicense(ctx, nil)
	assert.Error(t, err, "error licenses images with nil query")

	err = DiscoverLicense(ctx, &SimpleQuery{})
	assert.NoError(t, err, "error discovering license")

	err = DiscoverLicense(ctx, &SimpleQuery{
		ProjectID:  "test-project",
		OutputPath: "../../license.tmp",
	})
	assert.NoError(t, err, "error discovering license")
}

func TestVuln(t *testing.T) {
	setTestImplementations()
	ctx := context.Background()

	err := DiscoverVulns(ctx, nil)
	assert.Error(t, err, "error vulns images with nil query")

	err = DiscoverVulns(ctx, &VulnsQuery{})
	assert.NoError(t, err, "error discovering vulns")

	err = DiscoverVulns(ctx, &VulnsQuery{
		SimpleQuery: SimpleQuery{
			ProjectID: "test-project",
		},
	})
	assert.NoError(t, err, "error discovering vulns with project")

	err = DiscoverVulns(ctx, &VulnsQuery{
		CAAPI: true,
	})
	assert.NoError(t, err, "error discovering vulns with CAAPI")

	err = DiscoverVulns(ctx, &VulnsQuery{
		CAAPI: true,
		SimpleQuery: SimpleQuery{
			ProjectID: "test-project",
			OutputFmt: ParseOutputFormatOrDefault("raw"),
		},
	})
	assert.NoError(t, err, "error discovering vulns with CAAPI and project ID")
}

func getTestProjects(ctx context.Context) ([]*gcp.Project, error) {
	var list []*gcp.Project
	if err := loadTestData("../../etc/test-project.json", &list); err != nil {
		return nil, err
	}
	return list, nil
}

func getTestLocations(ctx context.Context, projectNumber string) ([]*gcp.Location, error) {
	list := []*gcp.Location{
		{
			ID:   "us-west1",
			Name: "us-west1",
		},
	}
	return list, nil
}

func getTestServices(ctx context.Context, projectNumber string, region string) ([]*gcp.Service, error) {
	var list []*gcp.Service
	if err := loadTestData("../../etc/test-service.json", &list); err != nil {
		return nil, err
	}
	return list, nil
}

func getTestImageInfo(ctx context.Context, image string) (*gcp.ImageInfo, error) {
	img, err := gcp.ParseImageInfo("us-docker.pkg.dev/cloudy-demos/art/artomator@sha256:1234567890")
	if err != nil {
		return nil, errors.Wrap(err, "error parsing image info")
	}
	return img, nil
}

func getTestCVEVulns(ctx context.Context, projectID string, cveID string) ([]*gcp.Occurrence, error) {
	var list []*gcp.Occurrence
	if err := loadTestData("../../etc/test-occurrence.json", &list); err != nil {
		return nil, err
	}
	return list, nil
}

func getTestImageVulns(ctx context.Context, projectID string, imageURL string) ([]*gcp.Occurrence, error) {
	var list []*gcp.Occurrence
	if err := loadTestData("../../etc/test-occurrence.json", &list); err != nil {
		return nil, err
	}
	return list, nil
}

func isTestAPIEnabled(ctx context.Context, projectNumber string, uri string) (bool, error) {
	return true, nil
}

func loadTestData(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "error reading test data")
	}

	if err := json.NewDecoder(bytes.NewReader(b)).Decode(v); err != nil {
		return errors.Wrap(err, "error decoding test data")
	}
	return nil
}

func TestFormatParse(t *testing.T) {
	f := ParseOutputFormatOrDefault("")
	assert.Equal(t, f, DefaultOutputFormat)
	f = ParseOutputFormatOrDefault("json")
	assert.Equal(t, f, JSONFormat)
	f = ParseOutputFormatOrDefault("yaml")
	assert.Equal(t, f, YAMLFormat)
	f = ParseOutputFormatOrDefault("raw")
	assert.Equal(t, f, RawFormat)
}

func TestWriteOutput(t *testing.T) {
	err := writeOutput("", JSONFormat, nil)
	assert.Error(t, err, "error writing output with nil data")
	f := struct {
		Name string
	}{
		Name: "test",
	}
	err = writeOutput("", JSONFormat, f)
	assert.Nil(t, err, "error writing output with JSON format")

	err = writeOutput("", RawFormat, f)
	assert.Nil(t, err, "error writing output with RAW format")
}

func TestScan(t *testing.T) {
	ctx := context.Background()
	err := scan(ctx, scanner.LicenseScanner, nil, nil)
	assert.Error(t, err, "error scanning with nil query")
}
