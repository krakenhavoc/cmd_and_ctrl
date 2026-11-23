package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guildpact Paragon — Artifact Creature — Construct 5/5, the back face of
// Invasion of Ravnica (March of the Machine, S82 batch 23):
//
//	"Whenever you cast a spell that's exactly two colors, look at the top
//	 six cards of your library. You may reveal a card that's exactly two
//	 colors from among them and put it into your hand. Put the rest on the
//	 bottom of your library in a random order."
//
// The cast trigger reads the spell's colours as it is cast (exactly two,
// CR 105.2). The look-and-take is Adaptive Omnitool's, with the one card
// that is exactly two colours optional and revealed; the rest go to the
// bottom in a random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     effects.InvasionOfRavnicaOracleID + "#1",
		Name:         "Guildpact Paragon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.ExactlyTwoColors(),
				"Guildpact Paragon — look at the top six cards, you may reveal a card that's exactly two colors and put it into your hand, the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					return effects.TakeFromLibraryToHand{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 6),
						Match:    effects.ExactlyTwoColors(),
						Max:      1,
						Optional: true,
						Label:    "Guildpact Paragon — you may reveal a card that's exactly two colors and put it into your hand",
						Key:      effects.TakeRevealRestBottomRandom,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
