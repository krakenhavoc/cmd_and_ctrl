package effects

import "github.com/google/uuid"

// Runebound Wolf — Creature — Wolf {1}{R}, 2/2 (Innistrad: Crimson Vow,
// #5495):
//
//	"{3}{R}, {T}: This creature deals damage equal to the number of Wolves
//	 and Werewolves you control to target opponent."
//
// The Wolves and Werewolves are counted once, as the ability resolves
// (CR 608.2h), by effective subtype, the Runebound Wolf itself included
// (it is a Wolf). The Wolf is the damage source, so lifelink or protection
// on it applies. If it has left, the count is its controller's board as it
// is now and the damage is its last-known source's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "955e4c54-14c4-49bb-94e1-c6cc800c6d91",
		Name:         "Runebound Wolf",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}, {T}: This creature deals damage equal to the number of Wolves and Werewolves you control to target opponent.",
			Cost:    Plus(ManaCost("{3}{R}"), TapCost()),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect: sourceDealsComputedDamageToEachLegalTarget(func(ctx *Context) int {
				return vowB07WolvesAndWerewolves(ctx.Game, ctx.Controller(), ctx.TextSubtype("Wolf"), ctx.TextSubtype("Werewolf"), uuid.Nil)
			}),
		}},
	})
}
