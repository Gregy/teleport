/*
Copyright 2022 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sqlserver

import (
	"context"
	"fmt"
	"net"

	"github.com/gravitational/teleport/lib/srv/db/common"
	"github.com/gravitational/trace"

	"github.com/jonboulle/clockwork"
	"github.com/sirupsen/logrus"
)

//
type Engine struct {
	// Auth handles database access authentication.
	Auth common.Auth
	// Audit emits database access audit events.
	Audit common.Audit
	// Context is the database server close context.
	Context context.Context
	// Clock is the clock interface.
	Clock clockwork.Clock
	// Log is used for logging.
	Log logrus.FieldLogger
}

//
func (e *Engine) HandleConnection(ctx context.Context, sessionCtx *common.Session, clientConn net.Conn) (err error) {
	fmt.Println("=== [AGENT] Received SQL Server connection ===")
	return trace.NotImplemented("asdasdasd")
}
