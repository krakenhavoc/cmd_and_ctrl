package game

import (
	"testing"

	"github.com/google/uuid"
)

// cost_cause_spell_test.go — #4281: delve's and escape's exiles are costs
// of the spell being cast, so each names a cost cause (the caster, and the
// spell as the item) instead of inferring whatever the resolving slot
// holds. The readers it was checked against: ExiledBySpellOrAbilityOf and
// DiscardCausedByOpponentOf read only MoveCauseEffect; the log line reads
// CauseItem only for MoveCauseEffect; the exiled-with link is set by the
// route's ExiledWith, which a cost never names.

const delveCauseOracle = "test-delve-cause"

// requireSpellCostCause checks that every card in ids was exiled with the
// spell's cost cause, and that an opponent whose ability was resolving at
// the time is not credited with the exile.
func requireSpellCostCause(t *testing.T, g *Game, ids []uuid.UUID, caster, opp, spell uuid.UUID) {
	t.Helper()
	for _, id := range ids {
		ev := zoneMoveFor(t, g, id)
		if ev.Cause != MoveCauseCost || ev.CauseController != caster || ev.CauseItem != spell {
			t.Errorf("exile of %s: cause %q / %s / %s, want cost / caster / the spell", id, ev.Cause, ev.CauseController, ev.CauseItem)
		}
		if ev.Actor != caster || ev.Source != spell {
			t.Errorf("exile of %s: actor %s source %s, want the caster and the spell", id, ev.Actor, ev.Source)
		}
		if ExiledBySpellOrAbilityOf(ev, opp) || ExiledBySpellOrAbilityOf(ev, caster) {
			t.Errorf("exile of %s was credited to a spell or ability's controller", id)
		}
	}
}

// A delve spell cast from inside an opponent's open resolution (a cast
// made off an effect) exiles its delve cards as a cost of THAT spell, not
// as the opponent's effect.
func TestDelveExileNamesTheSpellCostCause(t *testing.T) {
	prev := CatalogDelve
	CatalogDelve = func(key string) bool { return key == delveCauseOracle }
	t.Cleanup(func() { CatalogDelve = prev })

	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fuel := graveyardFuel(me, 2)
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Test Delver", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{2}{U}"
	c.OracleID = delveCauseOracle
	me.Hand.PushTop(c)

	g.mu.Lock()
	resolvingAbilityOf(g, opp.ID) // the opponent's item is resolving, with its slot open
	g.mu.Unlock()

	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{DelveIDs: fuel}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for _, id := range fuel {
		if !g.Exile.Contains(id) {
			t.Fatalf("delved card %s was not exiled", id)
		}
	}
	requireSpellCostCause(t, g, fuel, me.ID, opp.ID, c.InstanceID)
}

// Escape's graveyard exile is a cost of the escaping spell the same way.
func TestEscapeExileNamesTheSpellCostCause(t *testing.T) {
	const oracle = "test-escape-cause"
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id, yard := seedEscapeSpell(t, g, me, oracle, 3)
	pay := yard[:3]

	g.mu.Lock()
	resolvingAbilityOf(g, opp.ID)
	g.mu.Unlock()

	if err := castEscape(g, me, id, pay); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	requireSpellCostCause(t, g, pay, me.ID, opp.ID, id)
}
