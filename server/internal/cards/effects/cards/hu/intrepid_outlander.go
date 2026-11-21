package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Intrepid Outlander — Creature — Orc Ranger {1}{G}, 2/3:
//
//	"Reach
//	 Pack tactics — Whenever this creature attacks, if you attacked with
//	 creatures with total power 6 or greater this combat, venture into
//	 the dungeon."
//
// Pack tactics is Werewolf Pack Leader's intervening "if" (CR 603.4):
// the attacking power is read as the ability triggers and again, over
// the same declaration, as it resolves.
//
// No simplification.
func init() {
	const label = "Intrepid Outlander — pack tactics: venture into the dungeon"
	effects.Register(effects.Spec{
		OracleID:        "ff8d3532-bf4f-4209-ac01-351ddc9b9423",
		Name:            "Intrepid Outlander",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && ev.Actor == source.Controller &&
					effects.PackTacticsAttackingPower(g, source.Controller) >= 6
			}, label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if effects.PackTacticsTotalPower(g, item.Controller, ctx.Trigger().Event.Batch) < 6 {
					return nil
				}
				return effects.VentureIntoTheDungeon{}.Apply(ctx)
			}),
		},
	})
}
