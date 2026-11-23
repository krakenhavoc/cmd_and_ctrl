package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eagles of the North — Creature — Bird Soldier {5}{W}, 3/3 (The Hobbit):
//
//	"Flying
//	 When this creature enters, creatures you control get +1/+0 and gain
//	 first strike until end of turn.
//	 Plainscycling {1} ({1}, Discard this card: Search your library for a
//	 Plains card, reveal it, put it into your hand, then shuffle.)"
//
// Both halves of the enters trigger apply to the same creatures, fixed as
// it resolves (CR 611.2c): the Eagles themselves, and anything that
// arrives later this turn is left out. Plainscycling is typecycling
// (CR 702.29e): a Plains card is any land with the Plains subtype.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e519cebf-0f09-4da9-abef-82928738514d",
		Name:            "Eagles of the North",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Plainscycling", "{1}", "a Plains card", effects.IsLandWithSubtype("Plains")),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Eagles of the North — creatures you control get +1/+0 and gain first strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					mine := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: mine, Power: 1, Label: "Eagles of the North — +1/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"first strike"},
						Label: "Eagles of the North — first strike until end of turn"}.Apply(ctx)
				}),
		},
	})
}
