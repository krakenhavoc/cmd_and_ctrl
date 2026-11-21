package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quest for the Necropolis — Enchantment {B} (Modern Horizons 3):
//
//	"Landfall — Whenever a land you control enters, put a quest counter on
//	 this enchantment.
//	 {5}{B}, Sacrifice this enchantment: Put target creature card from a
//	 graveyard onto the battlefield under your control. This ability costs
//	 {1} less to activate for each quest counter on this enchantment.
//	 Activate only as a sorcery."
//
// The counter is Diary of Dreams' (CounterOnThis on a trigger) and the
// cost clause its own cost modifier (CostsLessEach over the counters on
// the source), counted as the cost is determined (CR 602.2b, 601.2f): {1}
// less per counter, never below {B} (the coloured part stays). The
// sacrifice is a cost; the card is read from any graveyard as the ability
// resolves and enters under the activator's control (Nighteyes the
// Desecrator's body). Sorcery speed is the ability's.
//
// No simplification.
func init() {
	const quest = "quest"
	effects.Register(effects.Spec{
		OracleID:     "407d9d09-3db6-49b0-9e80-4e1538156e44",
		Name:         "Quest for the Necropolis",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Quest for the Necropolis — put a quest counter on this enchantment",
				effects.Do(effects.CounterOnThis{Kind: quest, N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{5}{B}, Sacrifice this enchantment: Put target creature card from a graveyard onto the battlefield under your control. This ability costs {1} less to activate for each quest counter on this enchantment. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{5}{B}"), effects.SacrificeThis()),
			SorcerySpeed: true,
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(effects.CountersOnSource(quest),
					"This ability costs {1} less to activate for each quest counter on this enchantment."),
			},
			Targets: effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
			Effect:  effects.GapB55PutTargetCreatureCardOntoTheBattlefield,
		}},
	})
}
