package effects

import "github.com/google/uuid"

// Wanderbrine Trapper — Creature — Merfolk Scout {W}, 2/1 (Lorwyn
// Eclipsed, #4458):
//
//	"{1}, {T}, Tap another untapped creature you control: Tap target
//	 creature an opponent controls."
//
// Captain Janeway's cost shape: the second tap is not the {T} symbol (CR
// 302.6), so a creature that arrived this turn may pay it, and it does not
// target. The Trapper's own {T} is the symbol, so it must not be summoning
// sick. The target is judged as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "915d70ac-46ce-41d0-a137-f17e7aaea87d",
		Name:         "Wanderbrine Trapper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}, Tap another untapped creature you control: Tap target creature an opponent controls.",
			Cost: Plus(ManaCost("{1}"), TapCost(),
				TapAnotherUntapped("another untapped creature you control", Creature())),
			Targets: TargetCreature("target creature an opponent controls", OpponentControls()),
			Effect: Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return TapTarget{Target: id}.Apply(ctx)
			})),
		}},
	})
}
