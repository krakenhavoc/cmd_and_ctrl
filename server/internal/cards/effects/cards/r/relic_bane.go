package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Relic Bane — Enchantment — Aura {1}{B}{B} (Mirrodin, #5606):
//
//	"Enchant artifact
//	 Enchanted artifact has "At the beginning of your upkeep, you lose 2
//	 life.""
//
// Pillory of the Sleepless' granted upkeep trigger on an artifact: "you" is the
// enchanted artifact's controller, who loses the life, and the ability leaves
// with the Aura.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3609d300-af9f-4ddf-ad0e-8bebeca15a04",
		Name:         "Relic Bane",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant artifact", effects.Artifact()),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []effects.AbilityGrant{{
			Key: relicBaneGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Relic Bane — you lose 2 life", effects.Do(effects.YouLoseLife(2))),
			},
			Text: "At the beginning of your upkeep, you lose 2 life.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(relicBaneGrant)},
	})
}

// relicBaneGrant names the granted trigger's bundle: an on-disk identity.
const relicBaneGrant = "relic-bane/upkeep-lose-2-life"
