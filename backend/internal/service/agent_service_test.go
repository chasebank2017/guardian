
package service

import (
	"context"
	"regexp"
	"testing"

	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"guardian-backend/pkg/grpc/api"
)

func TestAgentServer_RegisterAgent(t *testing.T) {
	// 1. Initialize mock
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mock.Close()

	// 2. Define expected SQL and mock behavior
	// We use regexp because the exact SQL string can be hard to match.
	expectedSQL := `INSERT INTO agents (hostname, os_version, status, last_seen_at) VALUES ($1, $2, 'online', NOW()) RETURNING id`
	mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).
		WithArgs("test-host", "windows-11").
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int32(1)))

	// 3. Create service instance with mock DB
	agentServer := &AgentServer{
		DB: mock,
	}

	// 4. Call the method to be tested
	req := &api.RegisterAgentRequest{
		Hostname:  "test-host",
		OsVersion: "windows-11",
	}
	resp, err := agentServer.RegisterAgent(context.Background(), req)

	// 5. Assert results
	assert.NoError(t, err, "expected no error")
	assert.NotNil(t, resp, "response should not be nil")
	assert.Equal(t, int32(1), resp.AgentId, "expected agent ID to be 1")
	assert.Equal(t, "ok", resp.Status, "expected status to be 'ok'")

	// 6. Verify that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
