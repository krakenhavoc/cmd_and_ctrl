package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inverter of Truth — {2}{B}{B} Creature — Eldrazi 6/6 (Oath of the Gatewatch,
// S82 OLD b296):
//
//	"Devoid (This card has no color.)
//	 Flying
//	 When this creature enters, exile all cards from your library face down,
//	 then shuffle all cards from your graveyard into your library."
//
// Necropotence's face-down exile (CR 406.3) sized to the whole library, then
// Finale of Revelation's graveyard shuffle: the graveyard cards become the new
// library, and an empty graveyard still shuffles.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "941350ba-30e2-4549-865f-ca1bd7626782",
		Name:            "Inverter of Truth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Inverter of Truth — exile all cards from your library face down, then shuffle your graveyard into your library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if p := g.PlayerByIDForEffect(item.Controller); p != nil && p.Library != nil && p.Library.Size() > 0 {
						if err := (effects.ExileTopFaceDown{Player: item.Controller, N: p.Library.Size()}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.FinaleShuffleGraveyardIntoLibrary(ctx, item.Controller)
				}),
		},
	})
}
