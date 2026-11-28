package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Magnetic Web — Artifact {2} (Tempest, S82 OLD batch 760, #5628):
//
//	"If a creature with a magnet counter on it attacks, all creatures with
//	 magnet counters on them attack if able.
//	 Whenever a creature with a magnet counter on it attacks, all creatures
//	 with magnet counters on them block that creature this turn if able.
//	 {1}, {T}: Put a magnet counter on target creature."
//
// The first sentence is a requirement on declaring attackers (CR 508.1d), not
// a trigger: the Web gives every creature with a magnet counter one
// requirement that exists only in a declaration in which another such creature
// attacks (game.AttackRequirement.IfAnotherWithCounterAttacks, Viashino Bey's
// gate). The second is a trigger per attacking creature with a magnet counter
// whose effect writes Provoke's "blocks that attacker this turn if able" on
// each creature with a magnet counter. Both read the counter live, through the
// layers, and both lapse with the Web (CR 613.1f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c62c8006-5650-43cb-8247-da097aa42376",
		Name:         "Magnetic Web",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82OldB760AttackIfAnotherWithCounterAttacks("magnet",
				"If a creature with a magnet counter on it attacks, all creatures with magnet counters on them attack if able."),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.Counters["magnet"] > 0
			}, "Magnetic Web — all creatures with magnet counters on them block that creature this turn if able",
				effects.S82OldB760EachWithCounterBlocksTheAttacker("magnet", "Magnetic Web — blocks the attacking creature if able")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Put a magnet counter on target creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.RenewCountersOnTarget("magnet", 1),
		}},
	})
}
