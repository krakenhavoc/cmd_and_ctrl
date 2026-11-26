package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Predatory Advantage — Enchantment {3}{R}{G} (Alara Reborn, #5586):
//
//	"At the beginning of each opponent's end step, if that player didn't cast
//	 a creature spell this turn, create a 2/2 green Lizard creature token."
//
// Archfiend of Depravity's "each opponent's end step" watcher, the active player
// being the event's actor. The intervening "if" (CR 603.4) reads the turn's cast
// tally of THAT player (total minus noncreature spells) when the step begins
// (AppliesTo) and again as the trigger resolves (If; the active player is still
// the one whose end step it is). The token is created under the enchantment's
// controller.
//
// No simplification.
func init() {
	const label = "Predatory Advantage — create a 2/2 green Lizard token"
	effects.Register(effects.Spec{
		OracleID:     "24547885-2e2c-4118-9d5e-4058471f289b",
		Name:         "Predatory Advantage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginEndStep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B14OpponentsEndStepBegan(ev, source) && !castACreatureSpell(g, ev.Actor)
			},
			Key: label,
			Effect: effects.Do(effects.If{
				Cond: effects.CondBoard{Name: "that player didn't cast a creature spell this turn", Test: func(ctx *effects.Context) bool {
					return !castACreatureSpell(ctx.Game, ctx.Game.Seats[ctx.Game.Turn.ActiveSeat].ID)
				}},
				Then: []effects.Applier{effects.CreateToken{Template: effects.TokenCard("2/2 green Lizard"), N: 1}},
			}),
		}},
	})
}

// castACreatureSpell is whether the player cast a creature spell this turn:
// every spell cast minus the noncreature ones.
func castACreatureSpell(g *game.Game, player uuid.UUID) bool {
	t := g.CastTallyFor(player)
	return t.Total-t.Noncreature > 0
}
