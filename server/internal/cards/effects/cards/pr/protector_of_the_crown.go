package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Protector of the Crown — Creature — Giant Soldier {5}{W}, 2/5:
//
//	"When this creature enters, you become the monarch.
//	 All damage that would be dealt to you is dealt to this creature
//	 instead."
//
// ADR 0108 §9 decision 4 (#1905): a static redirection (CR 614.9) to the
// Protector itself, whoever is the monarch (the ruling). Two Protectors
// are two effects: you choose which applies, and the damage is never
// split (CR 616.1).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1052b21-ba96-499f-b9e6-9dba4ed82e1e",
		Name:         "Protector of the Crown",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Protector of the Crown — you become the monarch", effects.Do(effects.BecomeTheMonarch{})),
		},
		Replacements: []game.ReplacementEffect{
			effects.RedirectYourDamageToThis("Protector of the Crown — damage to you is dealt to it instead"),
		},
	})
}
