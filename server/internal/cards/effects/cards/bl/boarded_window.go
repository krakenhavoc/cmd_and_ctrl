package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boarded Window — Artifact {3} (Innistrad: Crimson Vow, #5495):
//
//	"Creatures attacking you get -1/-0.
//	 At the beginning of each end step, if you were dealt 4 or more damage
//	 this turn, exile this artifact."
//
// The static is a layer 7c modification of every creature attacking this
// artifact's controller, whoever controls the attacker; a creature attacking a planeswalker or a battle you control
// is not "attacking you". It is declared DependsOnAttackingStatus so a
// declaration, a removal from combat and the end of combat rebuild the
// cached characteristics (Blast-Furnace Hellkite's shape).
//
// The trigger is every player's end step, with an intervening "if"
// (CR 603.4) on the damage dealt to you this turn, combat and noncombat
// together, damage actually dealt (PlayerTurnTally.DamageTaken, S82), checked as
// the end step begins and again as it resolves. It exiles only the permanent
// the ability came from (ExileTarget, CR 400.7).
//
// No simplification.
func init() {
	const label = "Boarded Window — exile this artifact"
	effects.Register(effects.Spec{
		OracleID:     "3e159175-78e5-40df-8fe9-e694938e7222",
		Name:         "Boarded Window",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:                    game.Layer7PT,
			SubLayer:                 game.SubLayer7C_Modify,
			Label:                    "Creatures attacking you get -1/-0.",
			DependsOnAttackingStatus: true,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.AttackingTarget == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power--
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(boardedWindowDealtFour,
				effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Kind == game.EventBeginEndStep && g.DamageTakenThisTurn(source.Controller) >= 4
				}, label, func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTarget{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}

var boardedWindowDealtFour = effects.CondBoard{
	Name: "you were dealt 4 or more damage this turn",
	Test: func(ctx *effects.Context) bool { return ctx.Game.DamageTakenThisTurn(ctx.Controller()) >= 4 },
}
