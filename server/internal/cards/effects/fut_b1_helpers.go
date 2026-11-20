package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fut_b1_helpers.go — shapes shared by the Future Sight batch 1 cards
// (#5594, S82). Each is a package-level func or a constructor whose
// closure captures only constants, so nothing here holds a game.

// futB1SelfWhileHellbent is "this creature …, as long as you have no
// cards in hand" (hellbent): the source, while it is a creature, and
// while its controller's hand is empty. A hand's size is public.
func futB1SelfWhileHellbent(target *game.Card, g *game.Game, source *game.Card) bool {
	return selfCreature(target, g, source) && stickerHandSize(g, source.Controller) == 0
}

// futB1HellbentKeyword is "Hellbent — This creature has <keyword> as long
// as you have no cards in hand" (Cutthroat il-Dal): a layer 6 grant
// (CR 613.1f) that is on while the controller's hand is empty. It
// declares DependsOnHandSize, so the layers are recomputed as the hand
// empties and refills.
func futB1HellbentKeyword(keyword string) game.StaticAbility {
	ab := KeywordGrant(futB1SelfWhileHellbent, keyword)
	ab.Label = "Hellbent — This creature has " + keyword + " as long as you have no cards in hand."
	ab.DependsOnHandSize = true
	return ab
}

// futB1OwePact is the Pact cycle's "At the beginning of your next upkeep,
// pay <cost>. If you don't, you lose the game." (Slaughter Pact, Pact of
// the Titan): Pact of Negation's delayed trigger, owed whatever became of
// the rest of the spell.
func futB1OwePact(ctx *Context, name, cost string) error {
	return ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              name + " — pay " + cost + " or lose the game",
		Body:               pactPaymentBody,
		Params:             game.EffectParams{Cost: cost, Name: name},
	}.Apply(ctx)
}

// futB1SourcePowerAtLeast is "Activate only if this creature's power is n
// or greater" (Bloodshot Trainee): the source's power through the layers
// as the ability is announced. Read-only and public, like every
// activation condition.
func futB1SourcePowerAtLeast(n int) ActivationCondition {
	return func(g *game.Game, _, source uuid.UUID) bool {
		c, ok := g.LookupCardForEffect(source)
		return ok && c.IsCreature() && c.CurrentPower() >= n
	}
}

// futB1YouControlFewerCreaturesThanEachOpponent is Even the Odds' "Cast
// this spell only if you control fewer creatures than each opponent": the
// caster's creatures against every opponent still in the game, one at a
// time. With no opponent left there is no one to compare with, and the
// condition is not met.
func futB1YouControlFewerCreaturesThanEachOpponent(g *game.Game, caster uuid.UUID, _ game.Card) bool {
	creature := func(c game.Card) bool { return c.IsCreature() }
	mine := countControlled(g, caster, creature)
	anyOpponent := false
	notFewer := eachOpponent(g, caster, func(opp uuid.UUID) bool {
		anyOpponent = true
		return countControlled(g, opp, creature) <= mine
	})
	return anyOpponent && !notFewer
}
