package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kelpie Guide — Creature — Beast {2}{U}, 2/2 (EDHREC rank 4606):
//
//	"{T}: Untap another target permanent you control.
//	 {T}: Tap target permanent. Activate only if you control eight or
//	 more lands."
//
// An untapper that turns into an Icy Manipulator late. The second
// ability's gate is its activation condition (CR 602.1b, #743),
// ControlsAtLeast(8, lands). Both are {T} abilities on a creature, so
// summoning sickness applies to both (CR 302.6).
//
// One declared simplification, weaker than printed: "another target
// permanent you control" is matched by NAME, the catalog's convention
// for "another" on an activated ability's target clause (Benevolent
// Hydra, Ioreth) — a target predicate is not told which permanent is
// activating. The Guide cannot untap itself, as printed, and cannot
// untap a second Kelpie Guide either, which the printed card can.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8fc1e8b1-3afc-4e9d-a3ae-7bd9bcdcb465",
		Name:         "Kelpie Guide",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Untap another target permanent you control.",
				Cost:    effects.TapCost(),
				Targets: effects.Another(effects.TargetPermanent("another target permanent you control", effects.YouControl())),
				Effect:  effects.B22UntapEachLegalTarget,
			},
			{
				Label:     "{T}: Tap target permanent. Activate only if you control eight or more lands.",
				Cost:      effects.TapCost(),
				Targets:   effects.TargetPermanent("target permanent", effects.Permanent()),
				Condition: effects.ControlsAtLeast(8, effects.MatchLand),
				Effect:    effects.B36TapChosenCreature,
			},
		},
	})
}
