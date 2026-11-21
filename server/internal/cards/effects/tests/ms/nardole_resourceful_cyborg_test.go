package testsms

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const nardoleOracle = "81c515b7-9174-4b34-ad07-4144e3dbaaac"

// #2136: Nardole's mana pays a noncreature spell, and only that.
func TestNardoleManaCastsNoncreatureSpellsOnly(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	nardole := effectstest.PushCatalogPermanent(g, me.ID, "Nardole, Resourceful Cyborg", "Legendary Artifact Creature — Scientist", nardoleOracle, false)
	if err := g.AddCounter(nardole, "+1/+1", 2); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, nardole, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Nardole: %v", err)
	}
	if n := me.ManaPool.Total(); n != 2 {
		t.Fatalf("pool has %d mana, want one {U} per counter (2)", n)
	}

	// A creature spell, including an artifact creature, is refused.
	for _, tl := range []string{"Creature — Merfolk", "Artifact Creature — Construct"} {
		id := effectstest.HandSpell(me, "A Creature", tl, "{U}")
		effectstest.RefusedForMana(t, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}), tl)
	}
	// So is an activation, even of a noncreature source.
	cost, _ := game.ParseCost("{U}")
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Some Rock", "Artifact", "", false)
	rockCard, _ := effectstest.BattlefieldCard(g, rock)
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(rockCard)) {
		t.Error("Nardole's mana paid for an activated ability — stronger than printed")
	}
	// A noncreature spell is paid.
	for _, tl := range []string{"Instant", "Artifact"} {
		id := effectstest.HandSpell(me, "A Spell", tl, "{U}")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("Nardole's {U} for a %s: %v", tl, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
}

func TestNardoleWithNoCountersAddsNothing(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	nardole := effectstest.PushCatalogPermanent(g, me.ID, "Nardole, Resourceful Cyborg", "Legendary Artifact Creature — Scientist", nardoleOracle, false)
	_ = g.ActivateManaAbility(me.ID, nardole, 0, game.ManaAbilityParams{})
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %v, want empty", me.ManaPool)
	}
}
