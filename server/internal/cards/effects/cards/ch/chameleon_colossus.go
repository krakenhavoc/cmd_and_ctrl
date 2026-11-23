package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chameleon Colossus — Creature — Shapeshifter {2}{G}{G}, 4/4:
//
//	"Changeling (This card is every creature type.)
//	 Protection from black
//	 {2}{G}{G}: This creature gets +X/+X until end of turn, where X is
//	 its power."
//
// Changeling and protection are printed keywords. X is read as the
// ability resolves, from the creature's power then (the layered view), and
// both halves of the boost use that one number.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7b8f1458-c638-4907-9363-f7cc22d04605",
		Name:            "Chameleon Colossus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling, "protection from black"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}{G}: This creature gets +X/+X until end of turn, where X is its power.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{G}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				x := c.PowerForComparison()
				return effects.BoostUntilEOT{
					Target:    item.SourceCardID,
					Power:     x,
					Toughness: x,
					Label:     "Chameleon Colossus — +X/+X, X its power",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
