package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unruly Krasis — Creature — Shark Octopus Lizard {1}{G}{U}, 4/4 (Ravnica:
// Clue Edition, S82):
//
//	"Trample
//	 Whenever this creature attacks, you may have the base power and
//	 toughness of another target creature you control become X/X until end
//	 of turn, where X is this creature's power.
//	 {3}{G}{U}: Adapt 3."
//
// The attack trigger is a "you may" with a target (Admiral Brass's
// Optional(Targeting(…))). X is the Krasis's power as the ability resolves,
// counters and pumps included, as it last existed if it has gone (CR
// 608.2h); the new base size is layer 7b, so the target's counters and
// pumps still apply on top (Tanazir Quandrix). Adapt is Pteramander's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8420f73d-fdcf-4425-b9a7-c9d359fc028d",
		Name:            "Unruly Krasis",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WheneverThisAttacks("Unruly Krasis — another target creature you control has base power and toughness X/X until end of turn, where X is this creature's power",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if !ok {
						return nil
					}
					for _, id := range effects.LegalTargetCardIDs(ctx) {
						if err := effects.UntilEndOfTurn(ctx, id, nil, "Unruly Krasis — base power and toughness equal to its power",
							game.SetBasePTMods(info.Power, info.Power)...); err != nil {
							return err
						}
					}
					return nil
				}),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
				"Unruly Krasis — have another creature you control become X/X?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{G}{U}: Adapt 3.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{3}{G}{U}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.AdaptN(g, item, 3) },
		}},
	})
}
