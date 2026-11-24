package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sifter Wurm — Creature — Wurm {5}{G}{G}, 7/7 (Commander Legends):
//
//	"Trample
//	 When this creature enters, scry 3, then reveal the top card of your
//	 library. You gain life equal to that card's mana value."
//
// The scry's continuation (a registered tail, so the prompt is a restore
// point) reveals the card that is now on top, after the player has put the
// scried cards back, and gains life equal to its mana value as read from the
// library (a land is 0). An empty library reveals nothing and gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ed36b1f6-2922-46a9-b48d-b9ebedb9a142",
		Name:            "Sifter Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Scry: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sifter Wurm — scry 3, then reveal the top card of your library and gain life equal to its mana value",
				func(g *game.Game, item *game.StackItem) error {
					return effects.Scry{Player: item.Controller, N: 3,
						ThenTail: game.KeyedTail{Ref: siftersWurmReveal, Player: item.Controller}}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// siftersWurmReveal is the clause after the scry: reveal the top card, gain
// life equal to its mana value. An on-disk key.
var siftersWurmReveal = effects.CardTail("sifter-wurm/reveal-gain-life", func(ctx *effects.Context, _ game.TailOutcome) error {
	card, mv, err := effects.RevealTopReadingManaValue(ctx, "Sifter Wurm — reveal the top card of your library")
	if err != nil || card == uuid.Nil || mv <= 0 {
		return err
	}
	return effects.GainLife{Amount: mv}.Apply(ctx)
})
