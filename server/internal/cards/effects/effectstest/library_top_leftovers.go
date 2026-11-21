package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_top_leftovers_test.go — the proof cards for #1298, ADR 0088's
// 2026-09-23 amendment: Jace, the Mind Sculptor and Portent (a look at
// ANOTHER player's library), Temporal Cleansing (a top lane at a depth,
// chosen by the owner), Cream of the Crop (an exact top count), Hinder
// and Spell Crumple (a counter to a position, chosen by the counterer).

const (
	JaceTheMindSculptorOracle = "7f77a84e-5a4b-4834-aefa-3cecc175ae8e"
	PortentOracle             = "e744dbb6-2d56-462a-8d49-d79c73944048"
	TemporalCleansingOracle   = "5f67698a-28db-43cd-aaf1-52371b0b47eb"
	CreamOfTheCropOracle      = "b61a87e3-dc98-4db9-abed-b47b677d81ab"
	HinderOracle              = "c9db6b94-a7b1-4b93-b454-4dead8f85e34"
	SpellCrumpleOracle        = "d4ab7848-5c37-4c6b-be29-0bb703333e5b"
)

// libraryTopIDs is the top n cards of a library, top-first.
func LibraryTopIDs(p *game.Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, p.Library.Cards[p.Library.Size()-1-i].InstanceID)
	}
	return out
}

func KnownTo(g *game.Game, id, viewer uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && c.IsKnownTo(viewer)
}

// --- Hinder and Spell Crumple: a counter to a position -------------

// castShockAt puts an opponent's Shock on the stack, aimed at `at`.
func CastShockAt(t *testing.T, g *game.Game, caster *game.Player, at uuid.UUID) uuid.UUID {
	t.Helper()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Shock", TypeLine: "Instant",
		OracleID: B5ShockOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: at}},
	}); err != nil {
		t.Fatalf("CastSpell Shock: %v", err)
	}
	return id
}

// castCounterAt casts a catalog counterspell from `caster`'s hand at a
// spell on the stack.
func CastCounterAt(t *testing.T, g *game.Game, caster *game.Player, name, oracle string, spell uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Instant",
		OracleID: oracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}},
	}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}
