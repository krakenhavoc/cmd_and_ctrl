package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fathom Feeder — Creature — Eldrazi Drone {U}{B}, 1/1 (Battle for Zendikar,
// #5547):
//
//	"Devoid (This card has no color.)
//	 Deathtouch
//	 Ingest (Whenever this creature deals combat damage to a player, that
//	 player exiles the top card of their library.)
//	 {3}{U}{B}: Draw a card. Each opponent exiles the top card of their
//	 library."
//
// Ingest is OldB305Ingest. The activated ability draws, then each opponent
// exiles the top card of their own library (a library with none exiles none).
//
// No simplification.
func init() {
	const name = "Fathom Feeder"
	effects.Register(effects.Spec{
		OracleID:        "0767fb65-a621-494b-97a4-f23a4c8b4bec",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "deathtouch"},
		Triggered:       []game.TriggeredAbility{effects.OldB305Ingest(name)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}{B}: Draw a card. Each opponent exiles the top card of their library.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}{U}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				for _, opp := range ctx.Opponents() {
					if err := (effects.MillToZone{Player: opp, N: 1, To: game.ZoneExile}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
