package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pixie Illusionist — Creature — Faerie Wizard {U}, 1/1:
//
//	"Kicker {3}{G} (You may pay an additional {3}{G} as you cast this spell.)
//	 Flying
//	 If this creature was kicked, it enters with two +1/+1 counters on it.
//	 {T}: Target land you control becomes the basic land type of your choice until end of turn."
//
// ADR 0109 §1 (#1881): CR 305.7 from a resolved ability. The basic land
// type is chosen as the ability resolves (CR 608.2), by its controller,
// through effects.ChooseBasicLandTypeTail; a land that has become an
// illegal target by then is left alone and nothing is asked (CR 608.2b).
// Until end of turn the land's land types are replaced by the chosen one
// (its other subtypes stay, CR 205.1a), it loses the abilities its rules
// text gives it, keeps any another effect granted it, and taps for the
// chosen type's colour (CR 305.6).
//
// The counters are a CR 614.1c entry clause read off the cast
// (CountersPerKick, Vodalian Serpent's shape): kicked, it enters with two;
// put onto the battlefield any other way, with none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "3bd066d3-5b07-4986-9e6c-39d1ae6212ff",
		Name:                       "Pixie Illusionist",
		Completeness:               effects.CompletenessFull,
		PrintedKeywords:            []string{"flying"},
		OptionalCosts:              []game.AdditionalCost{effects.Kicker("{3}{G}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.CountersPerKick(game.CounterPlusOne, 2)},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target land you control becomes the basic land type of your choice until end of turn.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
			Effect:  effects.TargetLandBecomesUntilEOT("Pixie Illusionist"),
		}},
	})
}
