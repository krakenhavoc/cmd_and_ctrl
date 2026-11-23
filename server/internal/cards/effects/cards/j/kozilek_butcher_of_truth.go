package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kozilek, Butcher of Truth — Legendary Creature — Eldrazi {10}, 12/12:
//
//	"When you cast this spell, draw four cards.
//	 Annihilator 4 (Whenever this creature attacks, defending player
//	 sacrifices four permanents of their choice.)
//	 When Kozilek is put into a graveyard from anywhere, its owner
//	 shuffles their graveyard into their library."
//
// The cast trigger fires from the stack and resolves above the spell,
// so a countered Kozilek still draws four (the 2018-12-07 ruling).
// Annihilator 4 is the canonical keyword token: the engine makes the
// attack trigger, and the defending player sacrifices four permanents
// of their choice in one choice before blockers are declared (CR
// 702.86, ADR 0113 §2). The shuffle watches from the graveyard, so it
// fires however Kozilek got there: killed, sacrificed, discarded,
// milled or countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7b8528b0-71eb-4c9e-bed9-aa2d2e84038f",
		Name:            "Kozilek, Butcher of Truth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"annihilator 4"},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell("Kozilek, Butcher of Truth — draw four cards", effects.Do(effects.DrawCards{N: 4})),
			effects.WhenThisIsPutIntoAGraveyardFromAnywhere(
				"Kozilek, Butcher of Truth — its owner shuffles their graveyard into their library",
				effects.ShuffleYourGraveyardIntoYourLibrary),
		},
	})
}
