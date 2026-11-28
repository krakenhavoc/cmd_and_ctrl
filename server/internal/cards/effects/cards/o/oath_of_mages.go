package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oath of Mages — Enchantment {1}{R} (Exodus, S82 OLD b738, #5626):
//
//	"At the beginning of each player's upkeep, that player chooses target
//	 player who has more life than they do and is their opponent. The first
//	 player may have this enchantment deal 1 damage to the second player."
//
// Oath of Druids' shape: the Oath's controller controls the trigger, but the
// upkeep player chooses its target (OfThatPlayersChoice), judged against that
// player: an opponent of theirs who has more life than they do
// (WithMoreLifeThanTheChooser). With no such player the trigger is removed as
// it would go on the stack (CR 603.3d), and if the target no longer qualifies
// when it resolves the ability does nothing (CR 608.2b). On resolution the
// upkeep player, not the controller, is asked whether the enchantment deals the
// damage.
//
// No simplification.
const oathOfMagesName = "Oath of Mages"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b749918-f425-4243-aabb-7350bb74403c",
		Name:         oathOfMagesName,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.AtEachUpkeep(oathOfMagesName+" — that player may have this enchantment deal 1 damage to the chosen player", oathOfMagesAsk),
			effects.WithMoreLifeThanTheChooser(effects.OfThatPlayersChoice(
				effects.TargetPlayer("target player who has more life than they do and is their opponent"))),
		)},
	})
}

// oathOfMagesAsk asks the upkeep player whether to deal the damage.
func oathOfMagesAsk(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	first := item.Trigger.Event.Actor
	if first == uuid.Nil || g.PlayerByIDForEffect(first) == nil {
		return nil
	}
	return effects.MayChoice{
		Player:   first,
		Question: oathOfMagesName + " — have it deal 1 damage to the chosen player?",
		Then:     oathOfMagesDamage,
	}.Apply(effects.NewContext(g, item))
}

// oathOfMagesDamage is the upkeep player's "yes": 1 damage from the Oath to the
// chosen player, if that player is still a legal target. An on-disk key: never
// renamed, never reused.
var oathOfMagesDamage = effects.MayYes("confirm/oath-of-mages-damage", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return effects.DealDamage{Source: ctx.Item.SourceCardID, Target: t.ID, Amount: 1}.Apply(ctx)
		}
	}
	return nil
})
