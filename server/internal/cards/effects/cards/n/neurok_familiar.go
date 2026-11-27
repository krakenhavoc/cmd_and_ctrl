package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Neurok Familiar — Creature — Bird {1}{U}, 1/1 (Mirrodin, S82 OLD,
// #5606):
//
//	"Flying
//	 When this creature enters, reveal the top card of your library. If it's
//	 an artifact card, put it into your hand. Otherwise, put it into your
//	 graveyard."
//
// The top card is revealed to everyone, then it goes to your hand if it is
// an artifact card, and into your graveyard (a one-card mill, CR 701.17a)
// if it is anything else. A revealed token is neither: it is not a card
// (CR 108.2) and stays where it is (CR 111.8). An empty library reveals
// nothing.
//
// No simplification.
func init() {
	const name = "Neurok Familiar"
	effects.Register(effects.Spec{
		OracleID:        "8e11d1f7-fa98-4cd2-9912-081e9c9e37ee",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — reveal the top card of your library; an artifact card goes to your hand, otherwise to your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					revealed := g.RevealTopOfLibraryForEffect(item.Controller, item.SourceCardID, 1, name+" — reveal the top card of your library")
					if len(revealed) == 0 {
						return nil
					}
					c, ok := g.LookupCardForEffect(revealed[0])
					if !ok || effects.IsToken(c) {
						return nil
					}
					if c.IsArtifact() {
						return g.BounceToHandForEffect(revealed[0])
					}
					return effects.MillCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
