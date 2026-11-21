package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "407d9d09-3db6-49b0-9e80-4e1538156e44",
		Name:         "Quest for the Necropolis",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Quest for the Necropolis — put a quest counter on this enchantment",
				Do(CounterOnThis{Kind: quest, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{5}{B}, Sacrifice this enchantment: Put target creature card from a graveyard onto the battlefield under your control. This ability costs {1} less to activate for each quest counter on this enchantment. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{5}{B}"), SacrificeThis()),
			SorcerySpeed: true,
			CostModifiers: []game.CostModifier{
				CostsLessEach(CountersOnSource(quest),
					"This ability costs {1} less to activate for each quest counter on this enchantment."),
			},
			Targets: TargetCardInGraveyard("target creature card from a graveyard", Creature()),
			Effect:  gapB55PutTargetCreatureCardOntoTheBattlefield,
		}},
	})
}
