package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "7b8f1458-c638-4907-9363-f7cc22d04605",
		Name:            "Chameleon Colossus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling, "protection from black"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}{G}: This creature gets +X/+X until end of turn, where X is its power.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}{G}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				c, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				x := c.PowerForComparison()
				return BoostUntilEOT{
					Target:    item.SourceCardID,
					Power:     x,
					Toughness: x,
					Label:     "Chameleon Colossus — +X/+X, X its power",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
