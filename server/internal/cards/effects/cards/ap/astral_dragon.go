package cardsap

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Astral Dragon — Creature — Dragon {6}{U}{U}, 4/4 (Commander Legends: Battle
// for Baldur's Gate, #5948):
//
//	"Flying
//	 Project Image — When this creature enters, create two tokens that are
//	 copies of target noncreature permanent, except they're 3/3 Dragon
//	 creatures in addition to their other types, and they have flying."
//
// Will of the Temur's Dragon copy (CreateTokenCopy with an Except clause, CR
// 707.9): the type line gains Creature and Dragon, the printed size becomes
// 3/3 and flying is added. Two tokens come from one resolution, copying the
// target as it last existed if it has left (CR 608.2b makes the ability do
// nothing when its only target is gone).
//
// No simplification.
func init() {
	const label = "Astral Dragon — create two tokens that are copies of target noncreature permanent, except they're 3/3 Dragon creatures with flying"
	effects.Register(effects.Spec{
		OracleID:        "a20e2e1a-90d8-4073-96e9-efce86f0178e",
		Name:            "Astral Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					dragon := ctx.TextSubtype("Dragon")
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return effects.CreateTokenCopy{
							Controller: ctx.Controller(),
							Copy:       t.ID,
							N:          2,
							Except: func(tok *game.Card) {
								tok.TypeLine = effects.SocB19WithCreatureAndSubtypeInAddition(tok.TypeLine, dragon)
								tok.Power, tok.Toughness = 3, 3
								tok.VariableToughness = false
								tok.PrintedPTKnown = true
								if !slices.Contains(tok.Keywords, "flying") {
									tok.Keywords = append(tok.Keywords, "flying")
								}
							},
						}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetPermanent("target noncreature permanent", effects.Noncreature())),
		},
	})
}
