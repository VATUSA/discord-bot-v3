package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VATUSA/discord-bot-v3/internal/commands"
)

func TestAssignRolesSendsSyncCommand(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		cmds := make(chan commands.Command, 1)
		e := App(cmds, func() bool { return true })

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, httptest.NewRequest(method, "/assignRoles/123456789012345678", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", method, rec.Code)
		}
		select {
		case cmd := <-cmds:
			if cmd != (commands.SyncMember{UserID: "123456789012345678"}) {
				t.Fatalf("%s: unexpected command %#v", method, cmd)
			}
		default:
			t.Fatalf("%s: no command sent", method)
		}
	}
}

func TestAssignRolesRejectsInvalidId(t *testing.T) {
	cmds := make(chan commands.Command, 1)
	e := App(cmds, func() bool { return true })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/assignRoles/not-an-id", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if len(cmds) != 0 {
		t.Fatal("expected no command to be sent")
	}
}

func TestAssignRolesReturnsUnavailableWhenQueueFull(t *testing.T) {
	cmds := make(chan commands.Command) // unbuffered with no reader: always full
	e := App(cmds, func() bool { return true })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/assignRoles/123", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestHealthz(t *testing.T) {
	ready := false
	e := App(make(chan commands.Command), func() bool { return ready })

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 before ready, got %d", rec.Code)
	}

	ready = true
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when ready, got %d", rec.Code)
	}
}
