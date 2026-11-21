package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emrakul, the Aeons Torn — Legendary Creature — Eldrazi {15}, 15/15
// (Double Masters 2022):
//
//	"This spell can't be countered.
//	 When you cast this spell, take an extra turn after this one.
//	 Flying, protection from spells that are one or more colors, annihilator 6
//	 When Emrakul is put into a graveyard from anywhere, its owner shuffles
//	 their graveyard into their library."
//
// The cast trigger fires from the stack and resolves above the spell, so the
// extra turn is taken whatever becomes of Emrakul. The spell can't be
// countered (Spec.CantBeCountered, CR 113.6g). Protection from spells that are
// one or more colors (CR 702.16a) is the engine's colored_spells quality: a
// colored spell can't target Emrakul or damage it, while a colorless spell and
// every ability still can. Annihilator is the engine keyword. The shuffle
// watches from the graveyard, so it fires however Emrakul got there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "900ca697-ad38-4b2b-bc74-2ff7eb6ea951",
		Name:            "Emrakul, the Aeons Torn",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"flying", "protection from spells that are one or more colors", "annihilator 6"},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell("Emrakul, the Aeons Torn — take an extra turn after this one", effects.Do(effects.TakeExtraTurn{})),
			effects.WhenThisIsPutIntoAGraveyardFromAnywhere(
				"Emrakul, the Aeons Torn — its owner shuffles their graveyard into their library",
				effects.ShuffleYourGraveyardIntoYourLibrary),
		},
	})
}
