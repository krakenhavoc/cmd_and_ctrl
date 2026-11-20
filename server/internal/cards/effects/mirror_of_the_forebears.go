package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirror of the Forebears — Artifact {2} (Commander 2017, #5533):
//
//	"As this artifact enters, choose a creature type.
//	 {1}: Until end of turn, this artifact becomes a copy of target creature
//	 you control of the chosen type, except it's an artifact in addition to
//	 its other types."
//
// The type is chosen as it enters (Kindred Boon's ChooseCreatureTypeAsEnters,
// CR 614.12) and the target clause reads it off the Mirror
// (OfTheChosenType). The copy is Mirage Mirror's until-end-of-turn duration
// copy of the target (CR 707.2, 611.2), with Saheeli, Sublime Artificer's
// "except" adding the Artifact card type to the copied values (CR 707.9b).
// While it is the copy it has the copied card's text, so this ability is
// gone until cleanup puts the Mirror back; a Mirror that left in response is
// a new object and is left alone (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9eaeb54b-73a8-43cc-b4d8-ed8b5eb19646",
		Name:         "Mirror of the Forebears",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Mirror of the Forebears"),
		Activated: []ActivatedAbility{{
			Label:   "{1}: Until end of turn, this artifact becomes a copy of target creature you control of the chosen type, except it's an artifact in addition to its other types.",
			Cost:    ManaCost("{1}"),
			Targets: OfTheChosenType(TargetCreature("target creature you control of the chosen type", YouControl())),
			Effect: selfBecomesCopyOfTargetExcept("Mirror of the Forebears — becomes a copy until end of turn",
				func(v *game.PrintedValues) { v.AddCardType("Artifact") }),
		}},
	})
}
