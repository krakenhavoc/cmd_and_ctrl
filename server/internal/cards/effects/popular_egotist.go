package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Popular Egotist — Creature — Human Rogue {2}{B}, 3/2 (Duskmourn,
// #4466):
//
//	"{1}{B}, Sacrifice another creature or enchantment: This creature
//	 gains indestructible until end of turn. Tap it.
//	 Whenever you sacrifice a permanent, target opponent loses 1 life and
//	 you gain 1 life."
//
// The ability's sacrifice is a cost (b23SacrificeAnotherCreatureOrEnchantment),
// so "another" keeps the Egotist from paying with itself. The trigger is
// one per permanent sacrificed (Juri's EventSacrifice read), whatever asked
// for the sacrifice: a cost, an effect or a Treasure. It fires as the
// permanent leaves, so it goes above the spell or ability that was paid
// for and resolves even if that is countered, and sacrificing the Egotist
// itself triggers it, with the other permanents sacrificed at the same
// time each triggering it as well (the rulings). The drain is Blood
// Artist's body with the target opponent chosen as the trigger goes on the
// stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e6377683-1c36-4786-bd8d-c62ff2d25767",
		Name:         "Popular Egotist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventSacrifice, b23YouSacrificedAPermanent,
				"Popular Egotist — target opponent loses 1 life and you gain 1 life",
				Do(TargetPlayerLosesLife{N: 1}, GainLife{Amount: 1})),
				TargetPlayer("target opponent", Opponent())),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another creature or enchantment: This creature gains indestructible until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}{B}"), b23SacrificeAnotherCreatureOrEnchantment()),
			Effect:  b23IndestructibleAndTapThis("Popular Egotist — indestructible until end of turn"),
		}},
	})
}
