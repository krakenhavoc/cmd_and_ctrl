package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "796f1e1f-7fec-429d-82ae-125f541d6cc7",
		Name:         "Ulasht, the Hate Seed",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, rprB26OnePerOtherRedAndGreenCreature,
				"Ulasht enters with a +1/+1 counter for each other red creature and each other green creature you control"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Remove a +1/+1 counter from Ulasht: Choose one — Ulasht deals 1 damage to target creature; or create a 1/1 green Saproling creature token.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{1}"), RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Modes: ChooseOne(
				ModeWithPurpose(ModeDoing("Ulasht deals 1 damage to target creature.",
					TargetCreature("target creature"),
					socB20DamageToTheModeTarget(1)), ForTargets(DamageToTarget(0, 1))),
				ModeDoing("Create a 1/1 green Saproling creature token.", nil,
					func(_ *game.StackItem, ctx *Context, _ int) error {
						return CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1}.Apply(ctx)
					}),
			),
		}},
	})
}
