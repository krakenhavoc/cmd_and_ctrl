package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spinewoods Paladin — {4}{G} Creature — Human Knight 5/4:
//
//	"Trample
//	 When this creature enters, you gain 3 life.
//	 Plot {3}{G} (You may pay {3}{G} and exile this card from your
//	 hand. Cast it as a sorcery on a later turn without paying its mana
//	 cost. Plot only as a sorcery.)"
//
// A plot proof card (#1342) whose enters trigger fires however the
// creature was cast — a free cast of a plotted card is still a cast,
// and the creature still enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d2aa9649-8394-48ca-bee6-a044e5a86b38",
		Name:            "Spinewoods Paladin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SpecialActions:  []game.SpecialAction{effects.Plot("{3}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spinewoods Paladin — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
		},
	})
}
