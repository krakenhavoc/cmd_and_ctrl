package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Korozda Guildmage — Creature — Elf Shaman {B}{G}, 2/2 (Commander 2016, #5537):
//
//	"{1}{B}{G}: Target creature gets +1/+1 and gains intimidate until end of
//	 turn.
//	 {2}{B}{G}, Sacrifice a nontoken creature: Create X 1/1 green Saproling
//	 creature tokens, where X is the sacrificed creature's toughness."
//
// The first is the pump-and-keyword helper (PumpAndGrantFirstTarget). The
// second pays the sacrifice as a cost (CR 602.2b) and reads the creature's
// toughness as it last existed on the battlefield (CR 608.2h), counters and
// bonuses included; a creature of toughness 0 or less makes no tokens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b209b4ae-ceba-4a98-b307-222d3c78f8a1",
		Name:         "Korozda Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{B}{G}: Target creature gets +1/+1 and gains intimidate until end of turn.",
				Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "intimidate")), game.AnswerPump),
				Cost:    effects.ManaCost("{1}{B}{G}"),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.PumpAndGrantFirstTarget(1, 1, "Korozda Guildmage — +1/+1 and intimidate", "intimidate"),
			},
			{
				Label:   "{2}{B}{G}, Sacrifice a nontoken creature: Create X 1/1 green Saproling creature tokens, where X is the sacrificed creature's toughness.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost: effects.Plus(effects.ManaCost("{2}{B}{G}"), game.AbilityCost{
					SacrificeOther: effects.SacrificeSpec("a nontoken creature", effects.Creature(), effects.Not(effects.IsTokenPredicate())),
				}),
				Effect: func(g *game.Game, item *game.StackItem) error {
					fed, ok := effects.B17PermanentSacrificedToPay(g, item)
					if !ok {
						return nil
					}
					x := effects.DepartedCreatureToughness(g, fed)
					if x <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: x}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
