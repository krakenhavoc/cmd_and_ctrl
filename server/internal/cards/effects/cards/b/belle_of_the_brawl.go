package cardsb

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Belle of the Brawl — Creature — Human Knight {2}{B}, 3/2 (Throne of
// Eldraine, #5516):
//
//	"Menace
//	 Whenever this creature attacks, other Knights you control get +1/+0
//	 until end of turn."
//
// The set of Knights is fixed as the trigger resolves (CR 611.2c); "other"
// is object identity. "Knight" is a word of the card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6bc4a078-4d30-452d-9e33-ff09f943903e",
		Name:            "Belle of the Brawl",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Belle of the Brawl — other Knights you control get +1/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					word := ctx.TextSubtype("Knight")
					self := item.SourceCardID
					return effects.BoostUntilEOT{
						Match: func(_ *game.Game, you uuid.UUID, c game.Card) bool {
							return c.InstanceID != self && c.Controller == you && c.IsCreature() && c.HasSubtype(word)
						},
						Power: 1,
						Label: "Belle of the Brawl — other Knights +1/+0",
					}.Apply(ctx)
				}),
		},
	})
}
