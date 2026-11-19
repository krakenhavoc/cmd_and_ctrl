package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f578a5e0-77e2-4c35-90b2-6628ed9b218e",
		Name:         "Falkenrath Torturer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a creature: This creature gains flying until end of turn. If the sacrificed creature was a Human, put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    SacrificeN(1, "a creature", Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				src := ctx.Source()
				if !onBattlefield(g, src) {
					return nil
				}
				if err := (GrantKeywordUntilEOT{Target: src, Keywords: []string{"flying"}}).Apply(ctx); err != nil {
					return err
				}
				if info, ok := ctx.SacrificedPermanent(); ok &&
					(info.Characteristic.AllCreatureTypes || hasFold(info.Characteristic.Subtypes, "Human")) {
					return AddCounter{Target: src, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
