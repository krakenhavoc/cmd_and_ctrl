package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_cap_view_test.go — #3503: the view stamps a capped granted
// offer into `alternative_costs` while the creature count covers the
// spell and drops it when the count falls, read from the same derivation
// the cast gate uses. No new wire field.

func TestViewFollowsAComputedManaValueCap(t *testing.T) {
	const omnipresence = "test-3503-view-omnipresence"
	prev := game.CatalogGrantedAlternativeCosts
	game.CatalogGrantedAlternativeCosts = func(key string) []game.GrantedAlternativeCost {
		if key != omnipresence {
			return nil
		}
		return []game.GrantedAlternativeCost{{
			Offer: game.AlternativeCost{Key: "granted-free-up-to-creature-count", Label: "Cast it without paying its mana cost"},
			Zones: []game.ZoneKind{game.ZoneHand},
			MaxManaValueFunc: func(g *game.Game, caster uuid.UUID) int {
				n := 0
				for i := range g.Battlefield.Cards {
					if c := &g.Battlefield.Cards[i]; c.Controller == caster && c.IsCreature() {
						n++
					}
				}
				return n
			},
			MaxManaValueText: "the number of creatures you control",
		}}
	}
	t.Cleanup(func() { game.CatalogGrantedAlternativeCosts = prev })

	g := buildActiveGame(t)
	castWindowOpen(t, g)
	me, them := g.Seats[0], g.Seats[1]
	seen := map[uuid.UUID]bool{me.ID: true, them.ID: true}

	src := game.NewCard("Test Omnipresence", me.ID)
	src.TypeLine = "Enchantment"
	src.OracleID = omnipresence
	src.Controller = me.ID
	src.KnownBy = seen
	g.Battlefield.PushTop(src)

	spell := game.NewCard("Test Wurm", me.ID)
	spell.TypeLine = "Creature — Wurm"
	spell.ManaCost = "{1}{G}"
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(spell)

	offers := func() int {
		v := ViewOfGameFor(g, me.ID.String())
		for i := range v.Seats[0].Hand.Cards {
			if v.Seats[0].Hand.Cards[i].InstanceID == spell.InstanceID.String() {
				return len(v.Seats[0].Hand.Cards[i].AlternativeCosts)
			}
		}
		t.Fatalf("hand view missing the seeded card")
		return -1
	}
	addCreature := func() {
		c := game.NewCard("Test Bear", me.ID)
		c.TypeLine = "Creature — Bear"
		c.Controller = me.ID
		c.KnownBy = seen
		g.Battlefield.PushTop(c)
	}

	if got := offers(); got != 0 {
		t.Fatalf("offers with no creatures against a mana value 2 spell = %d, want 0", got)
	}
	addCreature()
	if got := offers(); got != 0 {
		t.Fatalf("offers with one creature = %d, want 0", got)
	}
	addCreature()
	if got := offers(); got != 1 {
		t.Fatalf("offers with two creatures = %d, want the capped free cast", got)
	}
}
