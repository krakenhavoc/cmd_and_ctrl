package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pandemonium — Enchantment {3}{R} (Exodus, S82 OLD b738, #5626):
//
//	"Whenever a creature enters, that creature's controller may have it deal
//	 damage equal to its power to any target of their choice."
//
// Confusion in the Ranks' watch of every permanent that enters, whoever
// controls it, Pandemonium itself included if it were a creature. The ability is
// Pandemonium's controller's, but the entering creature's controller chooses
// the "any target" (OfThatPlayersChoice, read off the trigger's snapshot of the
// creature, CR 603.3d) and is asked whether to have it deal the damage. The
// creature is the source and the damage equals its power as the ability resolves
// (its last-known power if it has left, CR 608.2h); a target that is no longer
// legal makes the ability do nothing (CR 608.2b).
//
// No simplification.
const pandemoniumName = "Pandemonium"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "073551ca-2e84-4d49-8484-9890718cc483",
		Name:         pandemoniumName,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.On(game.EventETB, pandemoniumACreatureEntered, pandemoniumName+" — that creature's controller may have it deal damage equal to its power to any target of their choice", pandemoniumAsk),
			effects.OfThatPlayersChoice(effects.TargetAny()),
		)},
	})
}

// pandemoniumACreatureEntered is "a creature enters", under anyone's control.
func pandemoniumACreatureEntered(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}

// pandemoniumAsk asks the entering creature's controller whether to have it deal
// the damage.
func pandemoniumAsk(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Object == nil {
		return nil
	}
	chooser := item.Trigger.Object.Controller
	if chooser == uuid.Nil || g.PlayerByIDForEffect(chooser) == nil {
		return nil
	}
	return effects.MayChoice{
		Player:   chooser,
		Question: pandemoniumName + " — have that creature deal damage equal to its power to the target?",
		Then:     pandemoniumDamage,
	}.Apply(effects.NewContext(g, item))
}

// pandemoniumDamage is the "yes": the creature deals damage equal to its power
// to the chosen target, if still legal. An on-disk key: never renamed, never
// reused.
var pandemoniumDamage = effects.MayYes("confirm/pandemonium-damage", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	creature, ok := ctx.TriggeringPermanent()
	if !ok || ctx.Trigger().Object == nil {
		return nil
	}
	ref := ctx.Trigger().Object.Ref()
	for _, t := range ctx.LegalTargets() {
		return effects.DealDamage{SourceObject: &ref, Target: t.ID, Amount: creature.Power}.Apply(ctx)
	}
	return nil
})
