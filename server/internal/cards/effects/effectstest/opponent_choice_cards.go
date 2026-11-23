package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opponent_choice_cards_test.go — #568. The five cards the
// opponent-facing resolution choice unblocked.

const (
	FactOrFictionOracle       = "437b2dab-15e0-4b9a-a204-58622d37a3b3"
	TormentOfHailfireOracle   = "30a0a6c2-1fbb-4784-ab96-22611d57e62c"
	CombustibleGearhulkOracle = "3494a575-f68e-4e78-9f6b-142cd8a0edea"
	CharismaticConquerorOracl = "3dcc35f3-74dc-46a8-8aa9-3411189f0547"
	PainfulQuandaryOracle     = "c37051cc-6683-4dbb-b5ff-5c3a5bdab1df"
)

// --- Fact or Fiction ------------------------------------------------

// seedFactOrFictionLibrary puts five named cards on top of a player's
// library, top card LAST in the returned slice's reading order — the
// reveal reports them top-first, so the helper returns them in reveal
// order.
func SeedFactOrFictionLibrary(p *game.Player, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(names))
	for _, name := range names {
		id := uuid.New()
		p.Library.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Artifact",
			Owner: p.ID, Controller: p.ID,
		})
		ids = append(ids, id)
	}
	// PushTop leaves the LAST pushed on top, and the reveal reads
	// top-first, so reverse.
	out := make([]uuid.UUID, 0, len(ids))
	for i := len(ids) - 1; i >= 0; i-- {
		out = append(out, ids[i])
	}
	return out
}

// castTormentOfHailfire casts the sorcery for the given X.
func CastTormentOfHailfire(t *testing.T, g *game.Game, caster uuid.UUID, x int) {
	t.Helper()
	p := g.PlayerByIDForEffect(caster)
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Torment of Hailfire", TypeLine: "Sorcery",
		OracleID: TormentOfHailfireOracle, Owner: caster, Controller: caster,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(caster, id, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("CastSpell Torment of Hailfire: %v", err)
	}
}

// drainRemainingTormentPrompts answers every option pick still open
// with its always-legal first branch, so the chain runs to the end and
// the table is left clean.
func DrainRemainingTormentPrompts(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceOptionPick {
				open = c
				break
			}
		}
		if open == nil {
			return
		}
		if err := g.ResolveOptionPick(open.ID, open.Chooser, 0); err != nil {
			t.Fatalf("ResolveOptionPick: %v", err)
		}
	}
	t.Fatal("the Torment chain did not terminate")
}
