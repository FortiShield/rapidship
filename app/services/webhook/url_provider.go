// Copyright 2023 Nxenv, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"context"

	"github.com/nxenv/rapidship/types"
)

var _ URLProvider = (*RapidshipURLProvider)(nil)

type RapidshipURLProvider struct{}

func NewURLProvider(_ context.Context) *RapidshipURLProvider {
	return &RapidshipURLProvider{}
}

func (u *RapidshipURLProvider) GetWebhookURL(_ context.Context, webhook *types.Webhook) (string, error) {
	// set URL as is (already has been validated, any other error will be caught in request creation)
	return webhook.URL, nil
}
