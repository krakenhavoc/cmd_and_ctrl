package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forbidden Crypt — Enchantment {3}{B}{B}:
//
//	"If you would draw a card, return a card from your graveyard to
//	 your hand instead. If you can't, you lose the game.
//	 If a card would be put into your graveyard from anywhere, exile
//	 that card instead."
//
// Both halves are mandatory CR 614 replacements (#2168). The draw is
// cancelled (nothing is drawn) and its body asks which card to take
// from the graveyard; with none to take, its controller loses the game
// (the lose-the-game outcome other effects use, so a "can't lose"
// effect still holds). A "draw three" asks three times in a row
// (CR 121.6b), and the graveyard is read fresh for each.
//
// The second half is Rest in Peace's replacement narrowed to the
// controller's own graveyard (GraveyardBecomesExile.YoursOnly): what
// would be milled, discarded or destroyed into YOUR graveyard is
// exiled, and an opponent's graveyard is untouched.
//
// A draw a doubler made into two is two draws, and the replacement runs for
// each.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b34b211-985b-4934-bb3b-8e6672685ac2",
		Name:         "Forbidden Crypt",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			{
				Watches: []game.EventKind{game.EventDrawCard},
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
					return ev.Kind == game.RepEventDraw && ev.DrawCount > 0 && src != nil && ev.DrawPlayer == src.Controller
				},
				DrawInstead: game.RegisterDrawInstead("forbidden-crypt", ChooseFromGraveyardToHandInstead(
					"Forbidden Crypt — return a card from your graveyard to your hand instead of drawing",
					func(g *game.Game, drawer, source uuid.UUID) error {
						// "If you can't, you lose the game."
						_, err := g.LoseTheGameForEffect(drawer, source)
						return err
					})),
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Forbidden Crypt: return a card from your graveyard to your hand instead of drawing",
			},
			GraveyardBecomesExile{
				YoursOnly: true,
				Label:     "Forbidden Crypt: exile instead of your graveyard",
			}.Build(),
		},
	})
}
