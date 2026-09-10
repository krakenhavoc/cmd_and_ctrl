package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reanimation_helpers.go — the shared body of the "creature card
// from a graveyard onto the battlefield" family.
//
// Its own file rather than helpers.go, per the convention
// fetchland_helpers.go and aang_helpers.go set: concurrent card
// batches collide on shared helper files.
//
// The family splits on one word. "From YOUR graveyard" (Zombify,
// Late to Dinner) reads only the caster's pile, and the creature
// returns to its owner, who is the caster either way. "From A
// graveyard ... under YOUR control" (Reanimate) reads every pile and
// the creature changes hands. Those are different target specs AND
// different controllers, and getting the second half right is what
// ReturnFromGraveyardUnderControlForEffect was added for — before
// it, reanimating across the table put the creature back under the
// opponent's control.

// targetCreatureInYourGraveyard is "target creature card in your
// graveyard".
func targetCreatureInYourGraveyard() *game.TargetSpec {
	return TargetCardInGraveyard("target creature card in your graveyard", Creature(), YouOwn())
}

// targetCreatureInAnyGraveyard is "target creature card in a
// graveyard" — every pile at the table, which is the whole point of
// the cards that print it.
func targetCreatureInAnyGraveyard() *game.TargetSpec {
	return TargetCardInGraveyard("target creature card in a graveyard", Creature())
}

// manaValueOf is a card's mana value: generic pips plus one per
// coloured pip. {X} contributes 0, which is what CR 202.3b says for
// a card anywhere other than the stack — and a reanimated card is
// always somewhere other than the stack when its mana value is read.
//
// Mirrors the arithmetic ManaValueLE already uses rather than
// introducing a second notion of the same number. An unparseable
// cost yields 0 rather than an error: a card with no mana cost
// (a reanimated token-turned-card, a fixture) has mana value 0.
func manaValueOf(c game.Card) int {
	cost, err := game.ParseCost(c.ManaCost)
	if err != nil {
		return 0
	}
	return cost.Generic + len(cost.Required)
}

// reanimateSingleTarget is the body every card in the family shares:
// read the spell's single graveyard target, then move it to the
// battlefield under `controller`. It returns the card AS IT SAT IN
// THE GRAVEYARD, so a rider ("you lose life equal to that card's
// mana value") reads the card's value rather than the permanent's,
// and ok=false when there was nothing to reanimate.
//
// Reading before the move is deliberate on both counts: it is the
// last moment the card is guaranteed findable in a graveyard, and
// CR 608.2 has the rider use the card's characteristics.
//
// A missing or illegal target is not an error. The engine already
// re-checks targets at resolution (CR 608.2b) and drops illegal
// ones, so an empty slot here means the spell fizzled its clause,
// not that something went wrong.
func reanimateSingleTarget(ctx *Context, controller uuid.UUID) (game.Card, bool) {
	targets := ctx.Targets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return game.Card{}, false
	}
	id := targets[0].ID
	card, ok := ctx.Game.LookupCardForEffect(id)
	if !ok {
		return game.Card{}, false
	}
	if err := (ReturnFromGraveyard{
		Target:     id,
		Dest:       game.ZoneBattlefield,
		Controller: controller,
	}).Apply(ctx); err != nil {
		return game.Card{}, false
	}
	return card, true
}
