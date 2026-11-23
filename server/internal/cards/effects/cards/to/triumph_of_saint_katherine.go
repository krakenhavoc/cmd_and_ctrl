package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triumph of Saint Katherine — Creature — Human Warrior {4}{W}, 5/5 (Warhammer
// 40,000 Commander, S82 GAP batch 196, #5948):
//
//	"Lifelink
//	 Praesidium Protectiva — When this creature is put into your graveyard from the
//	 battlefield, exile it and the top six cards of your library in a face-down
//	 pile. If you do, shuffle that pile and put it back on top of your library.
//	 Miracle {1}{W} (You may cast this card for its miracle cost when you draw it
//	 if it's the first card you drew this turn.)"
//
// The trigger is a dies trigger for a creature its controller owns (a stolen one
// goes to its owner's graveyard, which is not "your" graveyard). It exiles the
// card from the graveyard and the top six cards of the library face down, as a
// pile; then, if anything was exiled, the whole pile is shuffled and put back on
// top of the library in one random order (PutOnTopInRandomOrderForEffect), the
// seam added with this card. A library with fewer than six cards exiles what it
// has. Miracle is the keyword's own alternative cost, bound to the hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "6f8a4998-35c7-4e30-a9e7-7d4ec933238a",
		Name:             "Triumph of Saint Katherine",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"lifelink"},
		AlternativeCosts: []game.AlternativeCost{effects.Miracle("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.GapB196PutIntoYourGraveyardFromTheBattlefield,
				"Triumph of Saint Katherine — exile it and the top six cards of your library in a face-down pile; if you do, shuffle that pile and put it back on top of your library",
				effects.GapB196PraesidiumProtectiva),
		},
	})
}
