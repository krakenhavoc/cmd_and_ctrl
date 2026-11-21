package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wurmcoil Larva — Artifact Creature — Phyrexian Wurm {3}{B}{B}, 3/3
// (Modern Horizons 3):
//
//	"Deathtouch, lifelink
//	 When this creature dies, create a 1/2 black Phyrexian Wurm artifact
//	 creature token with deathtouch and a 2/1 black Phyrexian Wurm artifact
//	 creature token with lifelink."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "72a39730-acc3-470c-960e-3a91c5eae4be",
		Name:            "Wurmcoil Larva",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Tokens: 2, CreatureTokens: 2},
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Wurmcoil Larva — create a 1/2 deathtouch Wurm and a 2/1 lifelink Wurm", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, key := range []string{
					"1/2 black Phyrexian Wurm artifact with deathtouch",
					"2/1 black Phyrexian Wurm artifact with lifelink",
				} {
					if err := (effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard(key), N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
