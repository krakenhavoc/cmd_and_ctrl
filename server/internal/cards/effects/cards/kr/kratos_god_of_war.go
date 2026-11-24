package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kratos, God of War — Legendary Creature — God Warrior {R}{R}{R}, 2/3
// (Secret Lair Drop, #5462):
//
//	"Double strike
//	 All creatures have haste.
//	 At the beginning of each player's end step, Kratos deals damage to
//	 that player equal to the number of creatures that player controls that
//	 didn't attack this turn."
//
// The haste is Concordant Crossroads' layer 6 grant over every creature. The
// end-step trigger is one ability on every player's end step: the step's
// player is read off the event, and the count is the creatures they control
// as the trigger resolves that were not declared as attackers this turn (the
// per-object attack tally, so a creature that entered after combat counts and
// one that attacked and has since gone does not).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "42e55b84-8250-441d-a09e-7c61bf7661cc",
		Name:            "Kratos, God of War",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Static: []game.StaticAbility{{
			Label: "All creatures have haste.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "haste")
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEnd, "Kratos, God of War — deals damage to that player equal to the number of creatures they control that didn't attack this turn",
				func(g *game.Game, item *game.StackItem) error {
					player := effects.TriggeringActor(item)
					if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
						return nil
					}
					n := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == player && c.IsCreature() && !g.AttackedThisTurn(c.InstanceID) {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
