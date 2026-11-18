package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "fee73267-b0ef-4f43-a0bd-7728691d6f17",
		Name:            "Aven Heartstabber",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{sncSelfPT(2, 2), sncSelfKeywords("deathtouch")},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Aven Heartstabber — mill two cards, then draw a card", Do(MillCards{N: 2}, DrawCards{N: 1})),
		},
	})
}
