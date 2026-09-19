package legal_test

// table_settings_test.go — ADR 0075 §3: "Bots never host and cannot
// change settings. The legal-move enumerator does not offer
// set_table_settings or spawn."
//
// The enumerator's move types are a closed list (legal.Type*), so this
// holds by construction today. The test is here because the property
// is a POLICY, not an accident of which constants exist: the closed
// list is easy to widen, and a bot that could turn its own undo budget
// up — or turn spawning on — is a bot that can rewrite the table's
// rules mid-game. A move type that reads like a table setting fails
// here rather than shipping.

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

func TestEnumeratorNeverOffersTableSettings(t *testing.T) {
	forbidden := map[string]bool{
		string(actions.TypeSetTableSettings): true,
		string(actions.TypeSetUndoLimit):     true,
	}

	check := func(t *testing.T, label string, g *game.Game) {
		t.Helper()
		for _, p := range g.Seats {
			for _, m := range legal.EnumerateFor(g, p.ID) {
				if forbidden[m.Type] {
					t.Errorf("%s: seat %s was offered %q", label, p.Name, m.Type)
				}
			}
		}
	}

	// Mulligan window, the main phase, and combat: the three shapes of
	// the enumerator's answer.
	check(t, "mulligan", newTableMulligans(t))
	g := newTable(t)
	check(t, "precombat main", g)
	advanceTo(t, g, game.StepDeclareAttackers)
	check(t, "declare attackers", g)

	// And with the host designated and spawning on, in case a future
	// enumerator ever consults the settings before deciding what to
	// offer.
	on := true
	if err := g.UpdateSettings(g.Seats[0].ID, game.SettingsPatch{AllowSpawn: &on}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	check(t, "spawning allowed", g)
}
