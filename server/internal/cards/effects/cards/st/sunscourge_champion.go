package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sunscourge Champion — {2}{W} Creature — Human Wizard 2/3 (Hour of Devastation):
//
//	"When this creature enters, you gain life equal to its power.
//	 Eternalize—{2}{W}{W}, Discard a card."
//
// The power is read as the trigger resolves, or as the creature last had it
// (CR 608.2h): the 4/4 eternalized token gains 4.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5e9c69f-b949-4e9a-91d8-e0d52f8d8d3e",
		Name:         "Sunscourge Champion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sunscourge Champion — you gain life equal to its power",
				effects.Do(effects.GainLifeEqualToSubject{Stat: effects.SubjectPower})),
		},
		Activated: []effects.ActivatedAbility{effects.OldB238EternalizeDiscard("{2}{W}{W}")},
	})
}
