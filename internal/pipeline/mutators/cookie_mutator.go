package mutators

import (
	"github.com/rs/zerolog"

	"github.com/dadrus/heimdall/internal/config"
	"github.com/dadrus/heimdall/internal/heimdall"
	"github.com/dadrus/heimdall/internal/keystore"
	"github.com/dadrus/heimdall/internal/pipeline/subject"
	"github.com/dadrus/heimdall/internal/x/errorchain"
)

// by intention. Used only during application bootstrap
// nolint
func init() {
	registerMutatorTypeFactory(
		func(typ config.PipelineObjectType, conf map[string]any) (bool, Mutator, error) {
			if typ != config.POTCookie {
				return false, nil, nil
			}

			mut, err := newCookieMutator(conf)

			return true, mut, err
		})
}

type cookieMutator struct {
	cookies map[string]Template
}

func newCookieMutator(rawConfig map[string]any) (*cookieMutator, error) {
	type _config struct {
		Cookies map[string]Template `mapstructure:"cookies"`
	}

	var conf _config
	if err := decodeConfig(rawConfig, &conf); err != nil {
		return nil, errorchain.
			NewWithMessage(heimdall.ErrConfiguration, "failed to unmarshal cookie mutator config").
			CausedBy(err)
	}

	return &cookieMutator{
		cookies: conf.Cookies,
	}, nil
}

func (m *cookieMutator) Mutate(ctx heimdall.Context, sub *subject.Subject, _ *keystore.Entry) error {
	logger := zerolog.Ctx(ctx.AppContext())
	logger.Debug().Msg("Mutating using cookie mutator")

	for name, tmpl := range m.cookies {
		value, err := tmpl.Render(sub)
		if err != nil {
			return err
		}

		ctx.AddResponseCookie(name, value)
	}

	return nil
}

func (m *cookieMutator) WithConfig(config map[string]any) (Mutator, error) {
	if len(config) == 0 {
		return m, nil
	}

	return newCookieMutator(config)
}
