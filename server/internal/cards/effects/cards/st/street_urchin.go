package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Street Urchin — Legendary Enchantment — Background {1}{R} (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Commander creatures you own have "{1}, Sacrifice another creature or an
//	 artifact: This creature deals 1 damage to any target.""
//
// A granted activated ability (ADR 0093), Broadside Bombardiers' sacrifice
// cost ("another" excludes the commander itself by object) with Ballista
// Watcher's one damage to any target, dealt by the commander. Choose a
// Background is deck construction (ADR 0144).
//
// No simplification.
const streetUrchinGrant = "street-urchin/ping"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "f52759f9-7aa5-4493-8fe3-7be9ec340dde",
		Name:         "Street Urchin",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: streetUrchinGrant,
			Activated: []effects.ActivatedAbility{{
				Label: "{1}, Sacrifice another creature or an artifact: This creature deals 1 damage to any target.",
				Cost: effects.Plus(effects.ManaCost("{1}"),
					effects.SacrificeAnotherN(1, "another creature or artifact", effects.Or(effects.Creature(), effects.Artifact()))),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect:  effects.SourceDealsDamageToEachLegalTarget(1),
			}},
			Text: "{1}, Sacrifice another creature or an artifact: This creature deals 1 damage to any target.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(streetUrchinGrant)},
	})
}
