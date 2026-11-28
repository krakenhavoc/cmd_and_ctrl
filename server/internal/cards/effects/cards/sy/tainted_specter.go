package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tainted Specter — Creature — Specter {3}{B}, 2/2 (Mirage, #5633):
//
//	"Flying
//	 {1}{B}{B}, {T}: Target player discards a card unless they put a card from
//	 their hand on top of their library. If that player discards a card this way,
//	 this creature deals 1 damage to each creature and each player. Activate only
//	 as a sorcery."
//
// The target player is offered the put-on-top payment (#6187: not a discard);
// declining, or having no card to put, makes them discard a card of their choice,
// and only a card actually discarded brings the damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fc638800-6aec-4a48-ad2c-c9a5ce93c236",
		Name:            "Tainted Specter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{B}{B}, {T}: Target player discards a card unless they put a card from their hand on top of their library. If that player discards a card this way, this creature deals 1 damage to each creature and each player. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}{B}{B}"), effects.TapCost()),
			Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}},
			SorcerySpeed: true,
			Targets:      effects.TargetPlayer("target player"),
			Effect:       effects.OldB795TaintedSpecter,
		}},
	})
}
