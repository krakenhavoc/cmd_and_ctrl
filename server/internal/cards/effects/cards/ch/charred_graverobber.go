package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charred Graverobber — Creature — Skeleton Mercenary {2}{B}, 3/1 (Outlaws
// of Thunder Junction Commander, #5368):
//
//	"When this creature enters, return target outlaw card from your
//	 graveyard to your hand.
//	 Escape—{3}{B}{B}, Exile four other cards from your graveyard. (You
//	 may cast this card from your graveyard for its escape cost.)
//	 This creature escapes with a +1/+1 counter on it."
//
// Along the Crooked Way's targeted enters trigger over "outlaw card" (the
// b40Outlaw batch of five subtypes, any card type) in your graveyard, and
// Loathsome Chimera's escape: the counter lands only when the escape cost
// was paid. When the Graverobber escapes it is on the battlefield, so the
// card it returns is another one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "960cff3b-678c-404e-a643-ac5e91e6b50c",
		Name:             "Charred Graverobber",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.EscapeWithCounters("{3}{B}{B}", 4, 1)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Charred Graverobber — return target outlaw card from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
					}),
				effects.TargetCardInGraveyard("target outlaw card from your graveyard", effects.YouOwn(), effects.B40Outlaw()),
			),
		},
	})
}
