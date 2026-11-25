package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// sadisticObsessionGrant is the bundle the Aura gives the enchanted creature.
// An on-disk identity: never renamed.
const sadisticObsessionGrant = "sadistic-obsession/put-a-minus-counter"

// Sadistic Obsession — Enchantment — Aura {3}{B} (Modern Horizons):
//
//	"Enchant creature
//	 Enchanted creature has "{B}, {T}: Put a -1/-1 counter on target
//	 creature.""
//
// The granted ability is a layer 6 bundle on the enchanted creature (ADR 0093,
// Ringing Strike Mastery's shape). It is the enchanted creature's, so its
// controller activates it, with the creature's own {T} (a summoning-sick one
// cannot, CR 302.6), and the source of the counter is that creature. A target
// that has left by resolution is illegal and nothing happens (CR 608.2b); the
// counter may go on the enchanted creature itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c55af05-7fbe-4078-9463-689b21c0d3e4",
		Name:         "Sadistic Obsession",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: sadisticObsessionGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{B}, {T}: Put a -1/-1 counter on target creature.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.R35NightmareSowerCounter,
			}},
			Text: "{B}, {T}: Put a -1/-1 counter on target creature.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(sadisticObsessionGrant)},
	})
}
