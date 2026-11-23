package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tenth District Legionnaire — Creature — Human Soldier {R}{W}, 2/2 (S82 GAP, #5948):
//
//	"Haste
//	 Whenever you cast a spell that targets this creature, put a +1/+1
//	 counter on this creature, then scry 1."
//
// Heroic, as Rosnakht reads it: the cast event, the spell's targets read off
// its stack item (CR 601.2c). One trigger; counter then scry.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "32fad8d3-5351-49ce-a453-53dee403b5ff",
		Name:            "Tenth District Legionnaire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Purpose:         game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastASpellThatTargetsThis,
				"Tenth District Legionnaire — put a +1/+1 counter on this creature, then scry 1 (heroic)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.Scry{N: 1}.Apply(ctx)
				}),
		},
	})
}
