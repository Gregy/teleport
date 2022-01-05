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
	"crypto/tls"
	"fmt"
	"net"
	"strconv"

	"github.com/gravitational/teleport/lib/srv/db/common"
	"github.com/gravitational/trace"

	mssql "github.com/denisenkom/go-mssqldb"
	"github.com/denisenkom/go-mssqldb/msdsn"

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

	// TODO: Add authz

	host, port, err := net.SplitHostPort(sessionCtx.Database.GetURI())
	if err != nil {
		return trace.Wrap(err)
	}

	portI, err := strconv.ParseUint(port, 10, 64)
	if err != nil {
		return trace.Wrap(err)
	}

	connector := mssql.NewConnectorConfig(msdsn.Config{
		Host:       host,
		Port:       portI,
		User:       "sa",
		Encryption: msdsn.EncryptionOff,
		TLSConfig:  &tls.Config{InsecureSkipVerify: true},
	}, nil)

	conn, err := connector.Connect(ctx)
	if err != nil {
		return trace.Wrap(err)
	}
	defer conn.Close()

	mssqlConn, ok := conn.(*mssql.Conn)
	if !ok {
		return trace.BadParameter("expected *mssql.Conn, got: %T", conn)
	}

	rawConn := mssqlConn.GetUnderlyingConn()

	fmt.Println("Connected to SQL server", host, rawConn)

	return nil
}
