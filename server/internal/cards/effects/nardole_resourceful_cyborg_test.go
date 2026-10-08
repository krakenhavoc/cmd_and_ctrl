package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const nardoleOracle = "81c515b7-9174-4b34-ad07-4144e3dbaaac"

// #2136: Nardole's mana pays a noncreature spell, and only that.
func TestNardoleManaCastsNoncreatureSpellsOnly(t *testing.T) {
	g, me, _ := spendTable(t)
	nardole := pushCatalogPermanent(g, me.ID, "Nardole, Resourceful Cyborg", "Legendary Artifact Creature — Scientist", nardoleOracle, false)
	if err := g.AddCounter(nardole, "+1/+1", 2); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, nardole, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Nardole: %v", err)
	}
	if n := len(me.ManaPool); n != 2 {
		t.Fatalf("pool has %d mana, want one {U} per counter (2)", n)
	}

	// A creature spell, including an artifact creature, is refused.
	for _, tl := range []string{"Creature — Merfolk", "Artifact Creature — Construct"} {
		id := handSpell(me, "A Creature", tl, "{U}")
		refusedForMana(t, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}), tl)
	}
	// So is an activation, even of a noncreature source.
	cost, _ := game.ParseCost("{U}")
	rock := pushCatalogPermanent(g, me.ID, "Some Rock", "Artifact", "", false)
	rockCard, _ := battlefieldCard(g, rock)
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(rockCard)) {
		t.Error("Nardole's mana paid for an activated ability — stronger than printed")
	}
	// A noncreature spell is paid.
	for _, tl := range []string{"Instant", "Artifact"} {
		id := handSpell(me, "A Spell", tl, "{U}")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("Nardole's {U} for a %s: %v", tl, err)
		}
		passPriorityAroundTable(t, g)
	}
}

func TestNardoleWithNoCountersAddsNothing(t *testing.T) {
	g, me, _ := spendTable(t)
	nardole := pushCatalogPermanent(g, me.ID, "Nardole, Resourceful Cyborg", "Legendary Artifact Creature — Scientist", nardoleOracle, false)
	_ = g.ActivateManaAbility(me.ID, nardole, 0, game.ManaAbilityParams{})
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want empty", me.ManaPool)
	}
}
