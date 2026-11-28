package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urborg Stalker — {3}{B} Creature — Horror, 2/4 (Weatherlight, S82 OLD batch 771,
// #5629):
//
//	"At the beginning of each player's upkeep, if that player controls a nonblack,
//	 nonland permanent, this creature deals 1 damage to that player."
//
// Rivalry's each-upkeep trigger, with an intervening "if" (CR 603.4) about the
// player whose upkeep it is: they control a nonblack, nonland permanent (a
// colorless one is nonblack), checked as the upkeep begins and again as the
// ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5ef45e4-8c19-40bf-8756-c1b408dcb2fb",
		Name:         "Urborg Stalker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor != uuid.Nil && effects.S82OldB771ControlsANonblackNonlandPermanent(g, ev.Actor)
			}, "Urborg Stalker — it deals 1 damage to that player", func(g *game.Game, item *game.StackItem) error {
				p := effects.TriggeringActor(item)
				if p == uuid.Nil || !effects.S82OldB771ControlsANonblackNonlandPermanent(g, p) {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: p, Amount: 1}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
