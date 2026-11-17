package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scorn-Blade Berserker — Creature — Human Berserker {B}, 0/1 (March of the
// Machine):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target
//	 creature. If that's another creature, it gains the following ability
//	 until end of turn.)
//	 {1}, Sacrifice this creature: Draw a card."
//
// Brass Secretary's sacrifice-to-draw row. The sacrifice is part of the
// cost, so the card is drawn though the creature is gone by resolution.
// Backup grants the activated ability as a bundle: a creature that gained
// it sacrifices itself to pay.
//
// No simplification.
const scornBladeBerserkerGrant = "scorn-blade-berserker/sacrifice-draw"

func init() {
	Register(Spec{
		OracleID:     "93107139-b743-4d62-b0ac-9997202e0562",
		Name:         "Scorn-Blade Berserker",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       scornBladeBerserkerGrant,
			Activated: []ActivatedAbility{scornBladeBerserkerSacrifice()},
			Text:      "{1}, Sacrifice this creature: Draw a card.",
		}},
		Triggered: []game.TriggeredAbility{
			Backup("Scorn-Blade Berserker", 1, scornBladeBerserkerGrant),
		},
		Activated: []ActivatedAbility{scornBladeBerserkerSacrifice()},
	})
}

func scornBladeBerserkerSacrifice() ActivatedAbility {
	return ActivatedAbility{
		Label:   "{1}, Sacrifice this creature: Draw a card.",
		Purpose: game.Purpose{Answers: game.AnswerValue},
		Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
		Effect:  ebDrawACard,
	}
}
