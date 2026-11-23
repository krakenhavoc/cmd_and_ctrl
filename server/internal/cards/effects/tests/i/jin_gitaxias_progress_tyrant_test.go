package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castLightningBoltAt is castCatalogSpell for the registered
// Lightning Bolt (3 damage to any target), used here as an inert,
// observable instant.
func castLightningBoltAt(t *testing.T, g *game.Game, caster uuid.UUID, targetID uuid.UUID) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	if active.ID != caster {
		t.Fatalf("caster %s is not the active seat", caster)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: targetID}})
}

// TestJinGitaxiasCopiesYourFirstSpellEachTurnOnly — Lightning Bolt #1
// deals its 3 plus a copy's 3 (6 total); Lightning Bolt #2 the same
// turn deals only its own 3 — the copy trigger already fired once.
func TestJinGitaxiasCopiesYourFirstSpellEachTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Jin-Gitaxias, Progress Tyrant", "Legendary Creature — Phyrexian Praetor", effectstest.JinGitaxiasOracle, false)

	before := opp.Life
	castLightningBoltAt(t, g, me.ID, opp.ID)
	effectstest.SettleThroughCopyRetarget(t, g, me.ID, opp.ID)
	if got := before - opp.Life; got != 6 {
		t.Fatalf("first Bolt with Jin-Gitaxias out: %d damage, want 6 (bolt + one copy)", got)
	}

	before = opp.Life
	castLightningBoltAt(t, g, me.ID, opp.ID)
	effectstest.SettleThroughCopyRetarget(t, g, me.ID, opp.ID)
	if got := before - opp.Life; got != 3 {
		t.Errorf("second Bolt this turn: %d damage, want 3 (no second copy)", got)
	}
}

// TestJinGitaxiasCountersAnOpponentsFirstSpellEachTurn — the opposing
// half: an opponent's spell is countered before it can deal damage.
func TestJinGitaxiasCountersAnOpponentsFirstSpellEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Jin-Gitaxias, Progress Tyrant", "Legendary Creature — Phyrexian Praetor", effectstest.JinGitaxiasOracle, false)

	for g.Turn.ActiveSeat != 1 || (g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain) {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	before := me.Life
	castLightningBoltAt(t, g, opp.ID, me.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if got := before - me.Life; got != 0 {
		t.Errorf("Jin-Gitaxias should have countered the opponent's Bolt: took %d damage, want 0", got)
	}
}
