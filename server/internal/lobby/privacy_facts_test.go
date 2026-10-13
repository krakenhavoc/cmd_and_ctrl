package lobby

// The privacy notice (client/src/routes/Privacy.svelte, #2983, ADR 0149)
// states numbers. Each one is a constant or a setting somewhere in this
// repo. This test reads client/src/lib/privacy.ts, where the page takes
// them from, and fails when one no longer matches its source, so the
// page cannot go on saying "7 days" after the code says 14.
//
// It also fails when a new migration lands: the notice records the schema
// version it was last checked against, and a migration is how a new place
// to keep something about a person arrives. Reading the page again is
// the cheap part; shipping a table the page does not know about is the
// outcome this exists to prevent. accountdelete.Inventory is the other
// half: it fails on a column with no deletion decision.

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/bugstore"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/playmat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

const repoRoot = "../../.."

func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(raw)
}

// privacyFact reads `key: <number>,` from privacy.ts.
func privacyFact(t *testing.T, src, key string) int {
	t.Helper()
	m := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(key) + `:\s*(\d+),`).FindStringSubmatch(src)
	if m == nil {
		t.Fatalf("client/src/lib/privacy.ts has no `%s: <number>,` line", key)
	}
	n, _ := strconv.Atoi(m[1])
	return n
}

func TestPrivacyNoticeNumbersMatchTheServer(t *testing.T) {
	src := readRepoFile(t, "client/src/lib/privacy.ts")
	day := 24 * time.Hour

	check := func(key string, want int, source string) {
		t.Helper()
		if got := privacyFact(t, src, key); got != want {
			t.Errorf("privacy.ts %s = %d, but %s says %d: update the notice's source constant", key, got, source, want)
		}
	}
	check("endedTableDays", int(ws.EndedTableRetention/day), "ws.EndedTableRetention")
	check("bugArtifactDays", int(bugstore.DefaultRetention/day), "bugstore.DefaultRetention")
	check("bugArtifactMiB", int(bugstore.DefaultMaxStoreBytes>>20), "bugstore.DefaultMaxStoreBytes")
	check("bugImages", bugstore.MaxImages, "bugstore.MaxImages")
	check("signInDays", int(identityTTL/day), "identityTTL")
	check("guestSessionHours", int(sessionTTL/time.Hour), "sessionTTL")
	check("deckRequestsPerWindow", deckRequestLimit, "deckRequestLimit")
	check("deckRequestWindowHours", int(deckRequestWindow/time.Hour), "deckRequestWindow")
	check("playmatSlots", playmat.MaxSlots, "playmat.MaxSlots")

	// Backups: scripts/backup-offsite.sh is the source.
	script := readRepoFile(t, "scripts/backup-offsite.sh")
	m := regexp.MustCompile(`--keep-daily (\d+) --keep-weekly (\d+) --keep-monthly (\d+)`).FindStringSubmatch(script)
	if m == nil {
		t.Fatal("scripts/backup-offsite.sh has no `--keep-daily N --keep-weekly N --keep-monthly N`")
	}
	for i, key := range []string{"backupDaily", "backupWeekly", "backupMonthly"} {
		n, _ := strconv.Atoi(m[i+1])
		check(key, n, "scripts/backup-offsite.sh")
	}

	// Logs: ADR 0123 decision 7 is the source (Loki's 720 h).
	adr := readRepoFile(t, "docs/decisions/0123-monitoring-metrics-logs-dashboards-and-alerts.md")
	lm := regexp.MustCompile(`\*\*Retention:\*\* metrics 1 year, logs (\d+) days`).FindStringSubmatch(adr)
	if lm == nil {
		t.Fatal("ADR 0123 decision 7 no longer reads `Retention: metrics 1 year, logs N days`")
	}
	n, _ := strconv.Atoi(lm[1])
	check("logDays", n, "ADR 0123 decision 7")
}

func TestPrivacyNoticeUsesTheServersDeletePhrase(t *testing.T) {
	src := readRepoFile(t, "client/src/lib/privacy.ts")
	m := regexp.MustCompile(`export const DELETE_PHRASE = "([^"]+)"`).FindStringSubmatch(src)
	if m == nil || m[1] != DeleteAccountPhrase {
		t.Errorf("client DELETE_PHRASE = %v, server DeleteAccountPhrase = %q", m, DeleteAccountPhrase)
	}
}

// TestPrivacyNoticeIsCurrentWithTheSchema is the maintenance rule: the
// notice names the migration it was last checked against, and fails here
// when a newer one exists. Fix it by reading Privacy.svelte against the
// new migration, updating the page if the migration stores anything
// about a person (and accountdelete.Inventory with it), then bumping
// PRIVACY_SCHEMA_VERSION in client/src/lib/privacy.ts.
func TestPrivacyNoticeIsCurrentWithTheSchema(t *testing.T) {
	files, err := filepath.Glob(filepath.Join(repoRoot, "server/internal/db/migrations/*.sql"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no migrations found: %v", err)
	}
	newest := 0
	for _, f := range files {
		m := regexp.MustCompile(`^(\d+)_`).FindStringSubmatch(filepath.Base(f))
		if m == nil {
			continue
		}
		if v, _ := strconv.Atoi(m[1]); v > newest {
			newest = v
		}
	}
	src := readRepoFile(t, "client/src/lib/privacy.ts")
	m := regexp.MustCompile(`export const PRIVACY_SCHEMA_VERSION = (\d+);`).FindStringSubmatch(src)
	if m == nil {
		t.Fatal("client/src/lib/privacy.ts has no PRIVACY_SCHEMA_VERSION")
	}
	got, _ := strconv.Atoi(m[1])
	if got != newest {
		t.Errorf("the privacy notice was last checked against schema version %d, but migration %04d exists.\n"+
			"Read client/src/routes/Privacy.svelte against that migration. If it stores anything about a person, "+
			"update the page and accountdelete.Inventory; then set PRIVACY_SCHEMA_VERSION to %d (ADR 0149 §6).",
			got, newest, newest)
	}
}
