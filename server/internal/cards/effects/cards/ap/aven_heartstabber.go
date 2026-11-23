package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Heartstabber — Creature — Bird Assassin {U}{B}, 1/1 (Streets of New
// Capenna):
//
//	"Flying
//	 As long as there are five or more mana values among cards in your
//	 graveyard, this creature gets +2/+2 and has deathtouch.
//	 When this creature dies, mill two cards, then draw a card."
//
// Flying is printed; the pump and the deathtouch are the two self statics on
// the graveyard's mana values (sncSelfPT, sncSelfKeywords). The dies trigger
// mills first and draws second, in the printed order, so a milled card can
// never be the one drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fee73267-b0ef-4f43-a0bd-7728691d6f17",
		Name:            "Aven Heartstabber",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.SncSelfPT(2, 2), effects.SncSelfKeywords("deathtouch")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Aven Heartstabber — mill two cards, then draw a card", effects.Do(effects.MillCards{N: 2}, effects.DrawCards{N: 1})),
		},
	})
}
