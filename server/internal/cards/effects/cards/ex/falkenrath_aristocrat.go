package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Falkenrath Aristocrat — Creature — Vampire Noble {2}{B}{R} 4/1:
//
//	"Flying, haste
//	 Sacrifice a creature: This creature gains indestructible until end of
//	 turn. If the sacrificed creature was a Human, put a +1/+1 counter on this
//	 creature."
//
// Falkenrath Torturer's shape: the sacrificed creature is read as it last
// existed (a changeling counts as a Human).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cd2a87c4-d1c3-4a9e-b2df-e6fe251421c4",
		Name:            "Falkenrath Aristocrat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gains indestructible until end of turn. If the sacrificed creature was a Human, put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.SacrificeN(1, "a creature", effects.Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				src := ctx.Source()
				if !effects.OnBattlefield(g, src) {
					return nil
				}
				if err := (effects.GrantKeywordUntilEOT{Target: src, Keywords: []string{"indestructible"}}).Apply(ctx); err != nil {
					return err
				}
				if info, ok := ctx.SacrificedPermanent(); ok &&
					(info.Characteristic.AllCreatureTypes || effects.HasFold(info.Characteristic.Subtypes, "Human")) {
					return effects.AddCounter{Target: src, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
