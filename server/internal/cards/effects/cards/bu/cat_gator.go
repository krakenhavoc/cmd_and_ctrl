package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cat-Gator — Creature — Fish Crocodile {6}{B}, 3/2:
//
//	"Lifelink
//	 When this creature enters, it deals damage equal to the number of
//	 Swamps you control to any target."
//
// Consuming Corruption's count on an entry trigger. The Swamps are
// counted once, as the trigger resolves (CR 608.2h), by effective
// subtype. The Cat-Gator is the source, so its lifelink gains you that
// much life (CR 120.3f).
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "349e6b9e-40a8-49bc-9cd3-4a22f5363fa5",
		Name:            "Cat-Gator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Cat-Gator — damage equal to the number of Swamps you control to any target",
				effects.SourceDealsComputedDamageToEachLegalTarget(func(ctx *effects.Context) int {
					return effects.SwampsControlledBy(ctx.Game, ctx.Controller(), ctx.TextSubtype("Swamp"))
				})), effects.TargetAny()),
		},
	})
}
