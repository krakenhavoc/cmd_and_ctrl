package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "e6377683-1c36-4786-bd8d-c62ff2d25767",
		Name:         "Popular Egotist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventSacrifice, effects.B23YouSacrificedAPermanent,
				"Popular Egotist — target opponent loses 1 life and you gain 1 life",
				effects.Do(effects.TargetPlayerLosesLife{N: 1}, effects.GainLife{Amount: 1})),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another creature or enchantment: This creature gains indestructible until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.B23SacrificeAnotherCreatureOrEnchantment()),
			Effect:  effects.B23IndestructibleAndTapThis("Popular Egotist — indestructible until end of turn"),
		}},
	})
}
