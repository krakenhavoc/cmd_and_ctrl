package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// L'Rell, Fierce Visionary — Legendary Creature — Klingon Warrior
// {1}{B}{R}, 1/4 (Star Trek, #4973):
//
//	"At the beginning of combat on your turn, create a 2/1 red Klingon
//	 Warrior creature token with haste. Sacrifice it at the beginning of
//	 the next end step.
//	 {2}, Sacrifice another creature or artifact: Surveil 1."
//
// The combat trigger is D-7 Battle Cruiser's 3+ line without the charge
// gate (trekKlingonWarriorSacrificedAtNextEndStep): the token comes through
// the keyed creation door, and a CR 603.7 delayed trigger sacrifices
// exactly the tokens that landed at the beginning of the next end step.
// The activated ability's sacrifice is "another" by object identity
// (SacrificeAnotherN), paid at announce with the {2}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ea684e78-483c-4b89-b184-b2aeede83cc7",
		Name:         "L'Rell, Fierce Visionary",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("L'Rell, Fierce Visionary — create a 2/1 Klingon Warrior with haste; sacrifice it at the next end step",
				trekKlingonWarriorSacrificedAtNextEndStep(lrellFierceVisionarySacrificeAtEnd)),
		},
		Activated: []ActivatedAbility{{
			Label:  "{2}, Sacrifice another creature or artifact: Surveil 1.",
			Cost:   Plus(ManaCost("{2}"), SacrificeAnotherN(1, "another creature or artifact", Or(Creature(), Artifact()))),
			Effect: Do(Surveil{N: 1}),
		}},
	})
}

// lrellFierceVisionarySacrificeAtEnd schedules the sacrifice of the
// token(s) the trigger made (the tail's IDs). An on-disk key.
var lrellFierceVisionarySacrificeAtEnd = CardTail("lrell-fierce-visionary/sacrifice-at-end",
	trekSacrificeTokensAtNextEndStep("L'Rell, Fierce Visionary — sacrifice the Klingon Warrior"))
