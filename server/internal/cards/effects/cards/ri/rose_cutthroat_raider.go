package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rose, Cutthroat Raider — Legendary Artifact Creature — Robot {2}{R}{R},
// 3/2 (Fallout, #5948):
//
//	"First strike
//	 Raid — At end of combat on your turn, if you attacked this turn, create
//	 a Junk token for each opponent you attacked. (It's an artifact with "{T},
//	 Sacrifice this token: Exile the top card of your library. You may play
//	 that card this turn. Activate only as a sorcery.")
//	 Whenever you sacrifice a Junk, add {R}."
//
// The end-of-combat trigger has a CR 603.4 intervening "if", true as the step
// begins and again as it resolves, read from the turn's attack records. The
// number of Junk is the opponents attacked with a creature this turn, each
// once (gapB63OpponentsAttackedThisTurn): an attack on a planeswalker or a
// battle counts no player. The second ability is a plain triggered ability
// (not a mana ability: it triggers off a sacrifice, CR 605.1b), so the {R} is
// added as it resolves and empties at the end of the step (CR 106.4). It reads
// the sacrifice event (CR 701.21): a Junk you sacrificed to its own ability
// counts, as does one sacrificed any other way, and only yours.
//
// No simplification.
func init() {
	cond := effects.CondBoard{
		Name: "you attacked this turn",
		Test: func(ctx *effects.Context) bool { return effects.B18AttackedThisTurn(ctx.Game, ctx.Controller()) },
	}
	effects.Register(effects.Spec{
		OracleID:        "5a721c27-3e02-4f60-995e-b29d678d93b3",
		Name:            "Rose, Cutthroat Raider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(cond, effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepEndCombat, true), effects.TmcB6YouAttackedThisTurn),
				"Rose, Cutthroat Raider — create a Junk token for each opponent you attacked",
				func(g *game.Game, item *game.StackItem) error {
					n := effects.GapB63OpponentsAttackedThisTurn(g, item.Controller)
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.JunkToken(), N: n}.Apply(effects.NewContext(g, item))
				})),
			effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventSacrifice || ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.HasSubtype("Junk")
			}, "Rose, Cutthroat Raider — add {R}", effects.Do(effects.AddMana{Produced: "{R}"})),
		},
	})
}
