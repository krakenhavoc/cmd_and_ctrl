package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vivien's Stampede — Sorcery {4}{G}{G} (Streets of New Capenna
// Commander, #5491):
//
//	"Each creature you control gains vigilance, trample, and melee until
//	 end of turn. (Whenever a creature with melee attacks, it gets +1/+1
//	 until end of turn for each opponent you attacked this combat.)
//	 At the beginning of the next main phase this turn, draw a card for
//	 each player who was dealt combat damage this turn."
//
// The creatures are locked as the spell resolves (CR 611.2c). The second
// paragraph is a delayed trigger (CR 603.7) for the next main phase's
// beginning: cast in a first main phase it is the postcombat main phase;
// cast in the postcombat main phase it fires only if another main phase
// begins this turn. The body fires once and does nothing on a later turn
// (the "this turn" limit is the turn it was scheduled in, so a postcombat
// cast with no further main phase never draws). The number of cards is read
// as the trigger resolves: each player, any controller, who has been dealt
// combat damage by any source this turn counts once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0700c79-f948-4235-ad87-f8c6079a368a",
		Name:         "Vivien's Stampede",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerCombatGrant},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.YouControl()),
				Keywords: []string{"vigilance", "trample", game.KeywordMelee},
				Label:    "Vivien's Stampede — vigilance, trample and melee",
			}).Apply(ctx); err != nil {
				return err
			}
			return effects.ScheduleDelayedTrigger{
				At:     game.StepPostcombatMain,
				Label:  "Vivien's Stampede — draw a card for each player who was dealt combat damage this turn",
				Body:   viviensStampedeDrawBody,
				Params: game.EffectParams{Amount: ctx.Game.Turn.Seq},
			}.Apply(ctx)
		},
	})
}

// viviensStampedeDrawBody draws a card for each player dealt combat damage
// this turn, when the turn is the one that scheduled it. An on-disk key:
// never renamed, never reused.
var viviensStampedeDrawBody = game.DelayedBody("draw/viviens-stampede", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if g.Turn.Seq != p.Amount {
		return nil
	}
	n := 0
	for _, seat := range g.Seats {
		if g.CombatDamageTakenThisTurn(seat.ID) > 0 {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return effects.DrawCards{Player: item.Controller, N: n}.Apply(effects.NewContext(g, item))
})
