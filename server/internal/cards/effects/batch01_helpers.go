package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch01_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 01 (#294, `edhrec_rank` 9–237), the first slice of
// docs/decklists/card-coverage-roadmap.md.
//
// Its own file rather than helpers.go, per the convention #231 set:
// concurrent card batches collide on shared helper files.

// hasSubtype reports whether a card's post-layer subtypes carry the
// given word — "Human" for Return of the Wildspeaker's non-Human
// clause, "Treasure" for Professional Face-Breaker's sacrifice cost.
// Post-layer so a type-adding effect composes; whole-token match so
// "Human" does not catch a hypothetical "Humanoid".
//
// Thin wrapper over game.Card.HasSubtype, which is where the engine
// keeps the same question — an Urborg-granted Swamp and a
// Conspiracy-granted Goblin have to answer identically whether the
// asker is the mana pipeline or a card file.
func hasSubtype(c game.Card, subtype string) bool {
	return c.HasSubtype(subtype)
}

// isTreasure is the sacrifice-cost predicate for "Sacrifice a
// Treasure". Reads the Treasure subtype rather than the token's name,
// so a Treasure that is a copy of something else, or a nontoken
// artifact with the subtype, is still one.
func isTreasure(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return hasSubtype(c, "Treasure")
}

// manaValueOnStack is CR 202.3e: the mana value of a spell ON THE
// STACK, where {X} is the value chosen for it rather than zero. Mana
// Drain reads this off the countered spell; every off-stack read in
// the catalog (Reanimate, Feed the Swarm) keeps using manaValueOf.
func manaValueOnStack(c game.Card, item *game.StackItem) int {
	cost, err := game.ParseCost(c.ManaCost)
	if err != nil {
		return 0
	}
	mv := cost.Generic + len(cost.Required)
	if item != nil && item.XValue > 0 {
		mv += cost.XSlots * item.XValue
	}
	return mv
}

// eachOpponentLosesLife is "each opponent loses N life" — life loss,
// not damage, so no prevention or damage replacement sees it.
func eachOpponentLosesLife(g *game.Game, item *game.StackItem, n int) error {
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		if err := g.ChangePlayerLifeForEffect(ctx.Source(), opp, -n); err != nil {
			return err
		}
	}
	return nil
}
