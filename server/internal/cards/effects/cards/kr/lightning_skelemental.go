package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Skelemental — Creature — Elemental Skeleton {B}{R}{R}, 6/1 (Modern Horizons, #5519):
//
//	"Trample, haste
//	 Whenever this creature deals combat damage to a player, that player
//	 discards two cards.
//	 At the beginning of the end step, sacrifice this creature."
//
// Abyssal Specter's discard at two cards, combat damage only, and
// Blistering Firecat's end-step sacrifice (every player's end step).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7c3f2b3d-1517-4818-9675-189949175623",
		Name:            "Lightning Skelemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayer,
				"Lightning Skelemental — that player discards two cards",
				func(g *game.Game, item *game.StackItem) error {
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player:   effects.NewContext(g, item).Trigger().Event.Target,
						Source:   item.SourceCardID,
						N:        2,
						Question: "Lightning Skelemental — discard two cards",
					})
					return nil
				}),
			effects.AtEachStep(game.StepEnd, "Lightning Skelemental — sacrifice it", effects.SacrificeThisIfStillOnBattlefield),
		},
	})
}
