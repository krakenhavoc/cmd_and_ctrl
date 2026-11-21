package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skyclave Cleric // Skyclave Basilica — Modal DFC. Front: Creature —
// Kor Cleric {1}{W}, 1/3: "When this creature enters, you gain 2 life."
// Back: Land (mdfc_lands.go).
//
// An ordinary enters trigger. A Cleric played as its land back is the
// other face and has no trigger. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da9e3910-9a1c-43a9-9138-ca971b2bccae",
		Name:         "Skyclave Cleric",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Skyclave Cleric — gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
