package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lwaxana Troi, Charming Guest — Legendary Creature — Betazoid Advisor
// {1}{R}, 2/2 (Star Trek, #4973):
//
//	"You may choose not to untap Lwaxana Troi during your untap step.
//	 {2}, {T}: Choose another target creature you control. For as long as
//	 Lwaxana Troi remains tapped, if a triggered ability of that creature
//	 triggers, that ability triggers an additional time."
//
// The ability's effect is a continuous effect from a resolving ability
// timed by "for as long as Lwaxana Troi remains tapped" (CR 611.2b, the
// Couriers' shape): it ends the moment she untaps, and never starts if
// she is untapped by the time it resolves. While it lasts, the creature
// carries a CR 603.2d trigger doubler (the granted-doubler seam, #4137)
// that names its own triggered abilities, whatever caused them: each one
// triggers an additional time, and the two instances are put on the stack
// separately, with their own choices and targets. Another permanent's
// abilities are not doubled. "You may choose not to untap" is the untap
// opt-out that keeps the effect going.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e56472ec-6ae8-48b2-a4ce-0ed154161c06",
		Name:         "Lwaxana Troi, Charming Guest",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Lwaxana Troi — you may choose not to untap Lwaxana Troi")},
		Grants: []AbilityGrant{{
			Key:             lwaxanaTroiDoublerGrant,
			TriggerDoublers: []game.TriggerDoubler{lwaxanaTroiDoubler()},
			Text:            "If a triggered ability of this creature triggers, that ability triggers an additional time.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Choose another target creature you control. For as long as Lwaxana Troi remains tapped, if a triggered ability of that creature triggers, that ability triggers an additional time.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect: TargetGetsWhileThisRemainsTapped("Lwaxana Troi — that creature's triggered abilities trigger an additional time",
				game.GrantAbilitiesMod(lwaxanaTroiDoublerGrant)),
		}},
	})
}
