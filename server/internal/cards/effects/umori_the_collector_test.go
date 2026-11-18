package effects

import (
	"testing"
)

const umoriTheCollectorOracle = "913383a1-9c0d-40a1-8c5a-b0a8a95c09a6"

func TestUmoriDiscountsOnlyTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castCatalogSpell(t, g, "Umori, the Collector", "Legendary Creature — Ooze", umoriTheCollectorOracle, nil)
	passPriorityAroundTable(t, g)
	answerAnchorWord(t, g, me.ID, "sorcery")
	if got := g.ChosenOptionOf(id); got != "sorcery" {
		t.Fatalf("Umori chose %q, want \"sorcery\"", got)
	}
	if got := priceInHand(t, g, me, "Divination", "Sorcery", "{2}{U}"); got != 2 {
		t.Errorf("sorcery spell (chosen type): %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Lightning Bolt", "Instant", "{1}{R}"); got != 2 {
		t.Errorf("instant spell (not the chosen type) should be unaffected: %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Grizzly Bears", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("creature spell (not the chosen type) should be unaffected: %d, want 2", got)
	}
}

func TestUmoriDiscountsAnyChosenTypeIncludingBattleAndKindred(t *testing.T) {
	for _, tc := range []struct{ chosen, name, typeLine string }{
		{"battle", "A Battle", "Battle — Siege"},
		{"kindred", "A Kindred", "Kindred Instant — Goblin"},
		{"planeswalker", "A Walker", "Legendary Planeswalker — Test"},
	} {
		t.Run(tc.chosen, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			castCatalogSpell(t, g, "Umori, the Collector", "Legendary Creature — Ooze", umoriTheCollectorOracle, nil)
			passPriorityAroundTable(t, g)
			answerAnchorWord(t, g, me.ID, tc.chosen)
			if got := priceInHand(t, g, me, tc.name, tc.typeLine, "{3}"); got != 2 {
				t.Errorf("%s spell costs %d, want 2", tc.chosen, got)
			}
		})
	}
}

func TestUmoriDiscountsOnlyItsControllersSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Umori, the Collector", "Legendary Creature — Ooze", umoriTheCollectorOracle, nil)
	passPriorityAroundTable(t, g)
	answerAnchorWord(t, g, me.ID, "instant")
	if got := priceInHand(t, g, opp, "Lightning Bolt", "Instant", "{1}{R}"); got != 2 {
		t.Errorf("an opponent's instant costs %d, want 2", got)
	}
}
