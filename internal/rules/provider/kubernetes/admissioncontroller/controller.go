// Copyright 2023 Dimitrij Drus <dadrus@gmx.de>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package admissioncontroller

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/dadrus/heimdall/internal/config"
	"github.com/dadrus/heimdall/internal/handler/fxlcm"
	"github.com/dadrus/heimdall/internal/rules/rule"
)

// available here for test purposes
//
//nolint:gochecknoglobals
var listeningAddress = ":4458"

type AdmissionController interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type noopController struct{}

func (noopController) Start(context.Context) error { return nil }
func (noopController) Stop(context.Context) error  { return nil }

func New(
	tlsConf *config.TLS,
	logger zerolog.Logger,
	authClass string,
	ruleFactory rule.Factory,
) AdmissionController {
	if tlsConf == nil {
		return noopController{}
	}

	return &fxlcm.LifecycleManager{
		ServiceName:    "Validating Admission Controller",
		ServiceAddress: listeningAddress,
		Server:         newService(listeningAddress, ruleFactory, authClass, logger),
		Logger:         logger,
		TLSConf:        tlsConf,
	}
}
