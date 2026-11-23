package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arclight Phoenix — Creature — Phoenix {3}{R}, 3/2 (Ravnica Remastered,
// #5948):
//
//	"Flying, haste
//	 At the beginning of combat on your turn, if you've cast three or more
//	 instant and sorcery spells this turn, return this card from your
//	 graveyard to the battlefield."
//
// Darklight Phoenix's graveyard trigger (CR 113.6k) with an intervening "if"
// (CR 603.4), checked as the step begins; "you" is the card's owner while it
// is in the graveyard (CR 108.4). The turn's tally of instant and sorcery
// spells cast only ever rises, and a spell countered still was cast, so the
// condition cannot stop being true while the ability waits; the return does
// nothing if the card has left the graveyard meanwhile (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a80ea391-54d1-4b61-95a3-f8deb4cbe842",
		Name:            "Arclight Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventStepBegan,
				effects.AllOf(effects.StepBegan(game.StepBeginCombat, true),
					func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return g.CastTallyFor(source.Controller).InstantSorcery >= 3
					}),
				"Arclight Phoenix — return it from your graveyard to the battlefield",
				effects.ReturnThisCardFromYourGraveyard)),
		},
	})
}
