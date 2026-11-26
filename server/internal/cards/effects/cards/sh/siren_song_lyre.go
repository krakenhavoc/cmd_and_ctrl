package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Siren Song Lyre — Artifact — Equipment {2} (Born of the Gods, S82 OLD b379,
// #5559):
//
//	"Equipped creature has "{2}, {T}: Tap target creature."
//	 Equip {2}"
//
// Blazing Torch's shape: a granted activated ability on the equipped creature
// (the {T} is that creature's own, so summoning sickness applies to it), and
// the ordinary equip ability.
//
// No simplification.
func init() {
	const grant = "siren-song-lyre/tap"
	const text = "{2}, {T}: Tap target creature."
	effects.Register(effects.Spec{
		OracleID:     "400da7e6-43f8-4539-8959-ec60a3cc431b",
		Name:         "Siren Song Lyre",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   text,
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.TapTheFirstLegalTarget,
			}},
			Text: text,
		}},
		Static:    []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
