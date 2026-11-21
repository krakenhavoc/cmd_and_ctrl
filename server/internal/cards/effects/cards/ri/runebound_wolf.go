package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "955e4c54-14c4-49bb-94e1-c6cc800c6d91",
		Name:         "Runebound Wolf",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}, {T}: This creature deals damage equal to the number of Wolves and Werewolves you control to target opponent.",
			Cost:    effects.Plus(effects.ManaCost("{3}{R}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
				return effects.VowB07WolvesAndWerewolves(ctx.Game, ctx.Controller(), ctx.TextSubtype("Wolf"), ctx.TextSubtype("Werewolf"), uuid.Nil)
			}),
		}},
	})
}
