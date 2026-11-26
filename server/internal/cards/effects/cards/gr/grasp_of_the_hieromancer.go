package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grasp of the Hieromancer — Enchantment — Aura {1}{W} (Magic Origins):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has "Whenever this creature attacks,
//	 tap target creature defending player controls.""
//
// The trigger is the enchanted creature's own (ADR 0093, Conformer
// Shuriken's shape): its controller controls it and the target clause is
// read off that attack's defending player. A target that has left does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	const grant = "grasp-of-the-hieromancer/tap"
	const text = "Whenever this creature attacks, tap target creature defending player controls."
	effects.Register(effects.Spec{
		OracleID:     "51310af7-796d-4cf0-871e-5055765ffcdd",
		Name:         "Grasp of the Hieromancer",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Triggered: []game.TriggeredAbility{
				effects.WithTargetsFrom(
					effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
						return effects.AttackDeclared(ev, source)
					}, "Grasp of the Hieromancer — tap target creature defending player controls",
						effects.OldB319TapTheTargetCreature),
					effects.TargetOfDefendingPlayer("creature", "target creature defending player controls", effects.Creature())),
			},
			Text: text,
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantAbilitiesToAttached(grant),
		},
	})
}
