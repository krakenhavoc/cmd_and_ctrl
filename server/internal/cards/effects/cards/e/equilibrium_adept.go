package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Equilibrium Adept — Creature — Dog Monk {3}{R}, 2/4 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, exile the top card of your library. Until
//	 the end of your next turn, you may play that card.
//	 Flurry — Whenever you cast your second spell each turn, this creature
//	 gains double strike until end of turn."
//
// The enters trigger is Annie Flash's impulse exile (ExileTopWithPermission)
// with the window "until the end of your next turn" (ADR 0063's seat-turn
// duration: it ends at the cleanup of your next turn, so this turn's
// remainder counts too), a play permission, so a land may be played as the
// turn's land drop and a spell is paid for at its own cost and timing. The
// card stays in exile when the window passes unused. Flurry is an ability
// word (CR 207.2c): the trigger is YouCastYourSecondSpellEachTurn, counted
// from the turn's start, so spells cast before the Adept arrived count and a
// copy is not a cast. The double strike goes only to the creature that
// triggered it; one that has left or become a new object gets none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42bf9435-02e6-428f-a93c-fe4652deb4fc",
		Name:         "Equilibrium Adept",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Equilibrium Adept — exile the top card of your library; until the end of your next turn, you may play that card",
				func(g *game.Game, item *game.StackItem) error {
					window := g.UntilEndOfYourNextTurnDuration(item.Controller)
					return effects.ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1, Window: &window}.Apply(effects.NewContext(g, item))
				}),
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Equilibrium Adept — flurry: this creature gains double strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					return effects.GrantKeywordUntilEOT{
						Target:   item.SourceCardID,
						Keywords: []string{"double strike"},
						Label:    "Equilibrium Adept — double strike until end of turn",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
