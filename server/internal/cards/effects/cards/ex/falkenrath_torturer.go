package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Falkenrath Torturer — Creature — Vampire {2}{B}, 2/1 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"Sacrifice a creature: This creature gains flying until end of turn.
//	 If the sacrificed creature was a Human, put a +1/+1 counter on this
//	 creature."
//
// The sacrifice is a cost and may be the Torturer itself. "Was a Human"
// is read off the creature as it last existed on the battlefield
// (CR 608.2h), changelings included (CR 702.73a). The flying is granted
// and the counter placed only while the Torturer is still on the
// battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f578a5e0-77e2-4c35-90b2-6628ed9b218e",
		Name:         "Falkenrath Torturer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gains flying until end of turn. If the sacrificed creature was a Human, put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.SacrificeN(1, "a creature", effects.Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				src := ctx.Source()
				if !effects.OnBattlefield(g, src) {
					return nil
				}
				if err := (effects.GrantKeywordUntilEOT{Target: src, Keywords: []string{"flying"}}).Apply(ctx); err != nil {
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
