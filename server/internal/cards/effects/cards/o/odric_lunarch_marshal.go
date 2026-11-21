package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// odricLunarchKeywords are the thirteen abilities Odric shares, in printed
// order.
var odricLunarchKeywords = []string{
	"first strike", "flying", "deathtouch", "double strike", "haste", "hexproof",
	"indestructible", "lifelink", "menace", "reach", "skulk", "trample", "vigilance",
}

// Odric, Lunarch Marshal — Legendary Creature — Human Soldier {3}{W}, 3/3:
//
//	"At the beginning of each combat, creatures you control gain first
//	 strike until end of turn if a creature you control has first strike.
//	 The same is true for flying, deathtouch, double strike, haste,
//	 hexproof, indestructible, lifelink, menace, reach, skulk, trample, and
//	 vigilance."
//
// Each combat's, anyone's. The thirteen checks are made together as the
// ability resolves, against the keywords the creatures have then through the
// layered view (CR 608.2c, 611.2c): a keyword one creature gains from this
// ability does not make another keyword shared, and the creatures that gain
// them are the ones its controller controls as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bad76170-c773-4be5-9457-20dc9f745cb4",
		Name:         "Odric, Lunarch Marshal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfEachCombat("Odric, Lunarch Marshal — creatures you control gain each keyword a creature you control has until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					mine := effects.And(effects.Creature(), effects.ControlledBy(item.Controller))
					creatures := effects.MatchingBattlefield(ctx, mine)
					var shared []string
					for _, kw := range odricLunarchKeywords {
						for i := range creatures {
							if game.HasKeyword(&creatures[i], kw) {
								shared = append(shared, kw)
								break
							}
						}
					}
					if len(shared) == 0 {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Match: mine, Keywords: shared,
						Label: "Odric, Lunarch Marshal — shared keywords until end of turn"}.Apply(ctx)
				}),
		},
	})
}
