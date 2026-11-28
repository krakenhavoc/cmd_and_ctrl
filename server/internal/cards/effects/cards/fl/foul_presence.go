package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foul Presence — Enchantment — Aura {2}{B} (Apocalypse, S82 OLD, #5614):
//
//	"Enchant creature
//	 Enchanted creature gets -1/-1 and has "{T}: Target creature gets -1/-1
//	 until end of turn.""
//
// A layer 7c -1/-1 on the host and a granted activated ability (Quicksilver
// Dagger's shape, ADR 0093): the enchanted creature's controller activates it
// and it needs {T}, so summoning sickness applies to the host (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16f3a09f-e29b-4fb5-95ed-d76ad550c2ad",
		Name:         "Foul Presence",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Grants: []effects.AbilityGrant{{
			Key: "foul-presence/tap-shrink",
			Activated: []effects.ActivatedAbility{{
				Label:   "{T}: Target creature gets -1/-1 until end of turn.",
				Purpose: effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.BoostTargetUntilEOT("Foul Presence — -1/-1 until end of turn", -1, -1),
			}},
			Text: "{T}: Target creature gets -1/-1 until end of turn.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(-1, -1),
			effects.GrantAbilitiesToAttached("foul-presence/tap-shrink"),
		},
	})
}
