package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kazarov, Sengir Pureblood — Legendary Creature — Vampire {5}{B}{B}, 4/4
// (Dominaria):
//
//	"Flying
//	 Whenever a creature an opponent controls is dealt damage, put a +1/+1
//	 counter on Kazarov.
//	 {3}{R}: Kazarov deals 2 damage to target creature."
//
// The trigger waits for the damage to finish (CR 603.2c, ThatMuchDamage) so a
// creature dealt damage by several sources at once is one trigger, not one per
// source; it watches creatures an opponent of Kazarov's controller controls, as
// each was dealt the damage. The ability is a plain damage ability, Kazarov the
// source, its target chosen as it is activated and checked again as it
// resolves (CR 608.2b); it needs no black mana, so any {3}{R} pays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3067c653-ac29-4f1d-822f-7e2e5bb38a62",
		Name:            "Kazarov, Sengir Pureblood",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WithAtBatchEnd(effects.On(game.EventDealDamage, effects.OldB209OpponentCreatureDealtDamage,
				"Kazarov, Sengir Pureblood — put a +1/+1 counter on Kazarov",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})), effects.ThatMuchDamage()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{R}: Kazarov deals 2 damage to target creature.",
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Cost:    effects.ManaCost("{3}{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.RfDamageFirstLegalTarget(2),
		}},
	})
}
