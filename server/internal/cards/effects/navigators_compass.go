package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Navigator's Compass — Artifact {1}:
//
//	"When this artifact enters, you gain 3 life.
//	 {T}: Until end of turn, target land you control becomes the basic
//	 land type of your choice in addition to its other types."
//
// ADR 0109 §1 decision 6 (#1881): "in addition to its other types" takes
// nothing away (CR 205.1b, the last sentence of CR 305.7), so it is the
// existing add-subtypes record until end of turn, with the type chosen as
// the ability resolves (CR 608.2). The land keeps its land types and
// abilities and gains the chosen type's mana ability (CR 305.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c994e148-1bb5-4113-a9f2-73e6dab8f72d",
		Name:         "Navigator's Compass",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Navigator's Compass — you gain 3 life", Do(GainLife{Amount: 3})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Until end of turn, target land you control becomes the basic land type of your choice in addition to its other types.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target land you control", Land(), YouControl()),
			Effect:  TargetLandYouControlGainsChosenTypeUntilEOT("Navigator's Compass"),
		}},
	})
}
