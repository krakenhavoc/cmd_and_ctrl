package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "ff8d3532-bf4f-4209-ac01-351ddc9b9423",
		Name:            "Intrepid Outlander",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && ev.Actor == source.Controller &&
					packTacticsAttackingPower(g, source.Controller) >= 6
			}, label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if packTacticsTotalPower(g, item.Controller, ctx.Trigger().Event.Batch) < 6 {
					return nil
				}
				return VentureIntoTheDungeon{}.Apply(ctx)
			}),
		},
	})
}
