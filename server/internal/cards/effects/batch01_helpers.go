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
// Post-layer so a type-adding effect composes; exact match so
// "Human" does not catch a hypothetical "Humanoid".
func hasSubtype(c game.Card, subtype string) bool {
	for _, s := range c.Effective().Subtypes {
		if s == subtype {
			return true
		}
	}
	return false
}

// isTreasure is the sacrifice-cost predicate for "Sacrifice a
// Treasure". Reads the Treasure subtype rather than the token's name,
// so a Treasure that is a copy of something else, or a nontoken
// artifact with the subtype, is still one.
func isTreasure(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return hasSubtype(c, "Treasure")
}

// triggerAlreadyPendingFrom reports whether a triggered ability
// sourced from `source` is already waiting on PendingTriggers — the
// harvester has built it this event batch and it has not yet drained
// onto the stack.
//
// This is how a "whenever ONE OR MORE creatures you control deal
// combat damage to a player" ability (Professional Face-Breaker)
// fires once per combat damage step instead of once per creature.
// The engine emits one EventDealDamage per creature (the CR 603.1
// batching gap recorded on EventAttack), and the whole damage step's
// events fire inside one mutation before any priority boundary, so
// by the time the second creature's event reaches AppliesTo the
// first creature's trigger is already queued. Checking the queue,
// rather than the stack, is what keeps first-strike and regular
// damage as two separate triggers: the first-strike trigger has
// drained and resolved before regular damage is dealt.
//
// Without this the card would ship STRONGER than printed — three
// attackers connecting would make three Treasures — which is the
// #259 direction, and the reason Malcolm, Keen-Eyed Navigator
// declares its over-fire as a known gap rather than pretending
// otherwise.
func triggerAlreadyPendingFrom(g *game.Game, source *game.Card) bool {
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == source.InstanceID {
			return true
		}
	}
	return false
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
