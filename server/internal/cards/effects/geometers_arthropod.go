package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Geometer's Arthropod — Creature — Fractal Crab {G}{U}, 1/4 (Secrets of
// Strixhaven, #4456):
//
//	"Whenever you cast a spell with {X} in its mana cost, look at the top X
//	 cards of your library. Put one of them into your hand and the rest on
//	 the bottom of your library in a random order."
//
// "{X} in its mana cost" is the spell's own printed mana cost, parsed: an {X}
// in an additional or alternative cost does not count. X is the value
// announced for that spell, read as the trigger resolves from the spell on the
// stack or, if it was countered in response, from its last-known information
// (CR 608.2h, game.SpellXValueForEffect). Only you see the cards you look at;
// the one you keep is mandatory and is chosen from all of them, whatever its
// type (Prophetic Titan's look), and the rest go to the bottom in a random
// order. X of 0 looks at nothing; a library shorter than X is looked at whole.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7f577b46-c69d-4c75-9db5-7e10206b6b59",
		Name:         "Geometer's Arthropod",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, socB26YouCastASpellWithXInItsManaCost,
				"Geometer's Arthropod — look at the top X cards of your library; put one into your hand and the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					x, _ := g.SpellXValueForEffect(ctx.Trigger().Event.CardID)
					if x <= 0 {
						return nil
					}
					player := item.Controller
					return TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, x),
						Max:    1,
						Label:  "Geometer's Arthropod — put one of them into your hand",
						Key:    takeRestBottomRandom,
					}.Apply(ctx)
				}),
		},
	})
}
