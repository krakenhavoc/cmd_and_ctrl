package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boneyard Desecrator — Creature — Zombie Mercenary {3}{B}, 3/4 (Outlaws
// of Thunder Junction, #4468):
//
//	"Menace
//	 {1}{B}, Sacrifice another creature: Put a +1/+1 counter on this
//	 creature. If an outlaw was sacrificed this way, create a Treasure
//	 token. (Assassins, Mercenaries, Pirates, Rogues, and Warlocks are
//	 outlaws.)"
//
// The sacrifice is the ordinary "another creature" cost
// (SacrificeAnotherN). "If an outlaw was sacrificed this way" reads the
// cost's record of the sacrificed creature as it last existed on the
// battlefield (CR 608.2h, Context.SacrificedPermanent, Agency Coroner's
// read), so a changeling, or a creature something had made a Rogue,
// counts, and the Treasure does not depend on the counter: it comes even
// when the Desecrator has left. The counter goes only on this object
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c3adc698-a0c5-4c82-8347-536745c9e928",
		Name:            "Boneyard Desecrator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice another creature: Put a +1/+1 counter on this creature. If an outlaw was sacrificed this way, create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, SourceCounters: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.PutACounterOnThis(game.CounterPlusOne)(ctx.Game, ctx.Item)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					if !effects.OtjB1SacrificedWasOutlaw(ctx) {
						return nil
					}
					return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
				}),
			),
		}},
	})
}
