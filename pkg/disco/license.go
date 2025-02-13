package disco

import (
	"context"

	"github.com/mchmarny/disco/pkg/scanner"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func DiscoverLicense(ctx context.Context, in *SimpleQuery) error {
	if in == nil {
		return errors.New("nil input")
	}

	licenseFilter := func(v string) bool {
		return false
	}

	log.Debug().Msgf("Discovering licenses with: %s", in)
	printProjectScope(in.ProjectID)

	if err := scan(ctx, scanner.LicenseScanner, in, licenseFilter); err != nil {
		return errors.Wrap(err, "error scanning for vulnerabilities")
	}

	return nil
}
