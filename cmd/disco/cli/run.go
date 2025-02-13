package cli

import (
	"github.com/mchmarny/disco/pkg/disco"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	c "github.com/urfave/cli/v2"
)

var (
	projectIDFlag = &c.StringFlag{
		Name:     "project",
		Aliases:  []string{"p"},
		Usage:    "project ID",
		Required: false,
	}

	outputPathFlag = &c.StringFlag{
		Name:     "output",
		Aliases:  []string{"o"},
		Usage:    "path where to save the output",
		Required: false,
	}

	outputFormatFlag = &c.StringFlag{
		Name:     "format",
		Aliases:  []string{"f"},
		Usage:    "output format (json, yaml, raw)",
		Required: false,
	}

	outputDigestOnlyFlag = &c.BoolFlag{
		Name:  "digest",
		Usage: "output only image digests",
		Value: false,
	}

	caAPIExecFlag = &c.BoolFlag{
		Name:  "ca",
		Usage: "invokes Container Analysis API in stead of local scanner",
		Value: false,
	}

	cveFlag = &c.StringFlag{
		Name:     "cve",
		Aliases:  []string{"e"},
		Usage:    "exposure ID (CVE number, e.g. CVE-2019-19378)",
		Required: false,
	}

	runCmd = &c.Command{
		Name:  "run",
		Usage: "Cloud Run commands",
		Subcommands: []*c.Command{
			{
				Name:    "images",
				Aliases: []string{"img", "i"},
				Usage:   "List deployed container images",
				Action:  runImagesCmd,
				Flags: []c.Flag{
					projectIDFlag,
					outputPathFlag,
					outputFormatFlag,
					outputDigestOnlyFlag,
				},
			},
			{
				Name:    "vulnerabilities",
				Aliases: []string{"vul", "v"},
				Usage:   "Check for OS-level exposures in deployed images (supports specific CVE filter)",
				Action:  runVulnsCmd,
				Flags: []c.Flag{
					projectIDFlag,
					outputPathFlag,
					outputFormatFlag,
					cveFlag,
					caAPIExecFlag,
				},
			},
			{
				Name:    "licenses",
				Aliases: []string{"lic", "l"},
				Usage:   "Scans images for license types (requires OSS scanner, e.g. Trivy)",
				Action:  runLicenseCmd,
				Flags: []c.Flag{
					projectIDFlag,
					outputPathFlag,
					outputFormatFlag,
				},
			},
		},
	}
)

func printVersion(c *c.Context) {
	log.Info().Msgf(c.App.Version)
}

func runImagesCmd(c *c.Context) error {
	in := &disco.ImagesQuery{}
	in.ProjectID = c.String(projectIDFlag.Name)
	in.OutputPath = c.String(outputPathFlag.Name)
	in.OutputFmt = disco.ParseOutputFormatOrDefault(c.String(outputFormatFlag.Name))
	in.OnlyDigest = c.Bool(outputDigestOnlyFlag.Name)

	printVersion(c)
	if err := disco.DiscoverImages(c.Context, in); err != nil {
		return errors.Wrap(err, "error discovering images")
	}

	return nil
}

func runVulnsCmd(c *c.Context) error {
	in := &disco.VulnsQuery{}
	in.ProjectID = c.String(projectIDFlag.Name)
	in.OutputPath = c.String(outputPathFlag.Name)
	in.CVE = c.String(cveFlag.Name)
	in.OutputFmt = disco.ParseOutputFormatOrDefault(c.String(outputFormatFlag.Name))
	in.CAAPI = c.Bool(caAPIExecFlag.Name)

	printVersion(c)

	if in.CAAPI {
		log.Info().Msg("Note: Container Analysis scans currently are limited to base OS only")
	}

	if err := disco.DiscoverVulns(c.Context, in); err != nil {
		return errors.Wrap(err, "error excuting command")
	}

	return nil
}

func runLicenseCmd(c *c.Context) error {
	in := &disco.SimpleQuery{}
	in.ProjectID = c.String(projectIDFlag.Name)
	in.OutputPath = c.String(outputPathFlag.Name)
	in.OutputFmt = disco.ParseOutputFormatOrDefault(c.String(outputFormatFlag.Name))

	printVersion(c)
	if err := disco.DiscoverLicense(c.Context, in); err != nil {
		return errors.Wrap(err, "error discovering licenses")
	}

	return nil
}
