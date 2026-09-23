package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ba Sing Se — Land:
//
//	"This land enters tapped unless you control a basic land.
//	 {T}: Add {G}.
//	 {2}{G}, {T}: Earthbend 2. Activate only as a sorcery.
//	 (Target land you control becomes a 0/0 creature with haste that's
//	 still a land. Put two +1/+1 counters on it. When it dies or is
//	 exiled, return it to the battlefield tapped.)"
//
// The other card #1178 stopped at, and the one that shows where the
// line between the two keywords is: this earthbend activation is NOT
// an exhaust ability. It prints "Activate only as a sorcery" and
// nothing else, so it is repeatable every turn for as long as the land
// survives, and its entry deliberately carries no `Exhaust` bit. The
// set's other earthbend activations do print the keyword (Bitter
// Work); a land that could earthbend once per turn forever and a
// permanent that can do it once ever are different cards, and the
// catalog now says which is which rather than flattening both.
//
// The tapped entry is EntersTappedUnless on the shared
// b11ControlsBasicLand condition — a real CR 614 self-replacement, so
// an effect that watches permanents entering tapped sees it. The
// mana ability is the ordinary tap-for-{G}. The earthbend is the
// keyword action (earthbend.go, #1180) with the shared target clause;
// the count is the only thing this card contributes to it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de1ae205-ca5b-4d26-8194-ca85f1406e53",
		Name:         "Ba Sing Se",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{EntersTappedUnless(b11ControlsBasicLand)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{2}{G}, {T}: Earthbend 2. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{G}"), TapCost()),
			SorcerySpeed: true,
			Targets:      EarthbendTargets(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target := FirstLegalBattlefieldTarget(ctx)
				if target == uuid.Nil {
					return nil
				}
				return Earthbend{Target: target, N: 2}.Apply(ctx)
			},
		}},
	})
}
