package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const electroduplicateOracle = "112f2b3f-32e7-40b1-b80e-0b99184a840c"

func TestElectroduplicateMakesAHastyCopyThatSacrificesAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Big Bear", 4, 4)

	castCatalogSpell(t, g, "Electroduplicate", "Sorcery", electroduplicateOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear},
	})
	passPriorityAroundTable(t, g)

	var token *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Big Bear" && c.IsToken() {
			token = c
		}
	}
	if token == nil {
		t.Fatalf("no token copy of the bear on the battlefield")
	}
	if token.Power != 4 || token.Toughness != 4 {
		t.Errorf("token P/T = %d/%d, want 4/4", token.Power, token.Toughness)
	}
	if !hasKeywordForTest(token.Keywords, "haste") {
		t.Errorf("the token should have haste: %v", token.Keywords)
	}

	// Advance to the end step; the delayed trigger sacrifices it.
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token.InstanceID) {
		t.Errorf("the token should be sacrificed at the end step")
	}
}

func TestElectroduplicateFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Big Bear", 4, 4)
	id := seedGraveyardCard(t, g, "Electroduplicate", "Sorcery", electroduplicateOracle)

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := b43TokensNamed(g, me.ID, "Big Bear"); got != 1 {
		t.Fatalf("token copies = %d, want 1", got)
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back sorcery should be exiled")
	}
}
