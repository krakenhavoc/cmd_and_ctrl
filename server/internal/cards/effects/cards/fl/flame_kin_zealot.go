package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flame-Kin Zealot — Creature — Elemental Berserker {1}{R}{R}{W}, 2/2
// (Ravnica: City of Guilds, #5600):
//
//	"When this creature enters, creatures you control get +1/+1 and gain
//	 haste until end of turn."
//
// Catharsis's red half as a plain enters trigger: the creatures you
// control as it resolves get +1/+1 and haste until cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4af23012-16eb-47be-b666-3e536d8f6c1b",
		Name:         "Flame-Kin Zealot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Flame-Kin Zealot — creatures you control get +1/+1 and gain haste until end of turn", effects.Do(
				effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1, Label: "Flame-Kin Zealot — +1/+1"},
				effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"haste"}, Label: "Flame-Kin Zealot — haste"},
			)),
		},
	})
}
