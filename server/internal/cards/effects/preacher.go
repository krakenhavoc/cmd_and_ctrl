package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Preacher — Creature — Human Cleric {1}{W}{W}, 1/1 (#2055, #4061):
//
//	"You may choose not to untap this creature during your untap step.
//	 {T}: For as long as this creature remains tapped, gain control of
//	 target creature of an opponent's choice they control."
//
// Two seams meet here. The target is chosen by an opponent from among
// the creatures THEY control (TheyControl(OfAnOpponentsChoice(...)),
// #2055, ADR 0019's amendment of 2026-10-16): the engine asks the
// controller which opponent, when more than one could choose, and that
// opponent for the creature, before anything is paid.
//
// The duration is read from the moment the ability is ACTIVATED (#4061,
// CR 611.2b), per Preacher's ruling: "Whether Preacher is tapped is
// checked continually, starting when the ability is activated. If,
// before the ability resolves, there's any point at which Preacher is
// untapped, the ability has no effect - even if Preacher is tapped
// again by the time the ability resolves." The engine records an untap
// on the ability's stack item (StackItem.SourceUntapped), and
// DurationWhileThisRemainsTapped refuses on it, so a Preacher untapped
// and tapped again in response takes nothing. A Preacher that left, or
// came back as a new object, takes nothing either (#1432).
//
// The creature stays stolen while Preacher stays tapped: the untap
// opt-out is how its controller keeps it, as with Vedalken Shackles.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5ea9176-30c1-4152-9086-322f39245f46",
		Name:         "Preacher",
		Completeness: CompletenessFull,
		UntapOptOuts: []game.UntapOptOut{mayChooseNotToUntapSelf("Preacher — you may choose not to untap this creature")},
		Activated: []ActivatedAbility{{
			Label: "{T}: For as long as this creature remains tapped, gain control of target creature of an opponent's choice they control.",
			Cost:  TapCost(),
			Targets: TheyControl(OfAnOpponentsChoice(
				TargetCreature("target creature of an opponent's choice they control"))),
			Effect: GainControlOfTargetFor("Preacher — control while tapped", DurationWhileThisRemainsTapped),
		}},
	})
}
