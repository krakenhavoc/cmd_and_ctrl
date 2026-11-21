package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ulasht, the Hate Seed — Legendary Creature — Hellion Hydra {2}{R}{G}, 0/0
// (Reprint, #5908):
//
//	"Ulasht enters with a +1/+1 counter on it for each other red creature
//	 you control and a +1/+1 counter on it for each other green creature
//	 you control.
//	 {1}, Remove a +1/+1 counter from Ulasht: Choose one —
//	 • Ulasht deals 1 damage to target creature.
//	 • Create a 1/1 green Saproling creature token."
//
// The entry counters are a CR 614.1c self-replacement counted as it enters
// (b19EntersWithCountersCounted): a red-and-green creature counts for both
// clauses, and Ulasht is never its own "other". The ability is modal with
// the removal as a cost (CR 602.2b), so removing the last counter lets the
// 0/0 Ulasht die before the ability resolves; the damage is then dealt by
// its last-known self (CR 608.2h) and the Saproling is made either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "796f1e1f-7fec-429d-82ae-125f541d6cc7",
		Name:         "Ulasht, the Hate Seed",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B19EntersWithCountersCounted(game.CounterPlusOne, effects.RprB26OnePerOtherRedAndGreenCreature,
				"Ulasht enters with a +1/+1 counter for each other red creature and each other green creature you control"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Remove a +1/+1 counter from Ulasht: Choose one — Ulasht deals 1 damage to target creature; or create a 1/1 green Saproling creature token.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Modes: effects.ChooseOne(
				effects.ModeWithPurpose(effects.ModeDoing("Ulasht deals 1 damage to target creature.",
					effects.TargetCreature("target creature"),
					effects.SocB20DamageToTheModeTarget(1)), effects.ForTargets(effects.DamageToTarget(0, 1))),
				effects.ModeDoing("Create a 1/1 green Saproling creature token.", nil,
					func(_ *game.StackItem, ctx *effects.Context, _ int) error {
						return effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1}.Apply(ctx)
					}),
			),
		}},
	})
}
