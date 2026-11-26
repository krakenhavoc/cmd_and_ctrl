package game

import "testing"

// Altar of the Lost: "Spend this mana only to cast spells with flashback from
// a graveyard." The clause is two tags that must both hold: the spell has
// flashback (ManaRestrictSpellFlashback) and it is cast from a graveyard
// (ManaRestrictCastFromGraveyard).

func altarTags() []string {
	return []string{ManaRestrictSpellFlashback, ManaRestrictCastFromGraveyard}
}

func TestSpellFlashbackRestriction(t *testing.T) {
	creature := creatureSpendContext()
	fromGraveyard := withCastFrom(creature, ZoneGraveyard)
	hasFlashback := func(c ManaSpendContext) ManaSpendContext { c.FlashbackSpell = true; return c }

	for _, tc := range []struct {
		name string
		ctx  ManaSpendContext
		want bool
	}{
		{"a flashback spell cast from a graveyard", hasFlashback(fromGraveyard), true},
		{"a spell that claims the flashback cost, from a graveyard", withAltCost(fromGraveyard, "flashback"), true},
		{"a spell with flashback cast from a hand", hasFlashback(withCastFrom(creature, ZoneHand)), false},
		{"a spell without flashback cast from a graveyard", fromGraveyard, false},
		{"a spell cast from a graveyard for another alternative cost", withAltCost(fromGraveyard, "escape"), false},
		{"an activation of a card with flashback", ManaSpendContext{Purpose: SpendPurposeActivate, FlashbackSpell: true, CastFrom: ZoneGraveyard}, false},
		{"an unknown purpose", ManaSpendContext{FlashbackSpell: true, CastFrom: ZoneGraveyard}, false},
	} {
		if got := tc.ctx.allows(altarTags()); got != tc.want {
			t.Errorf("%s: allows = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// The engine path: Altar's mana pays a flashback cast and nothing else.
func TestSpellFlashbackManaPaysAFlashbackCastOnly(t *testing.T) {
	const oracle = "test-looting"
	g := newActiveGame(t)
	me := g.Seats[0]
	id := seedFlashbackCard(t, g, me, oracle)

	altarPool := func() ManaPool {
		return ManaPool{
			{Color: "R", Restrictions: altarTags()},
			{Color: "R", Restrictions: altarTags()},
			{Color: "R", Restrictions: altarTags()},
		}
	}
	me.ManaPool = altarPool()
	if err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true}); err != nil {
		t.Fatalf("flashback cast with Altar mana: %v", err)
	}
}
