package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unpredictable Cyclone — Enchantment {3}{R}{R}:
//
//	"If a cycling ability of another nonland card would cause you to draw a
//	 card, instead exile cards from the top of your library until you exile
//	 a card that shares a card type with the cycled card. You may cast that
//	 card without paying its mana cost. Then put the exiled cards that
//	 weren't cast this way on the bottom of your library in a random order.
//	 Cycling {2}"
//
// A draw replacement (CR 614.1a) that reads which card's cycling ability is
// making the draw (ReplacementEvent.DrawCycled, #6169): a draw made while
// a cycling ability resolves, by that ability's controller. It applies only
// when that player controls this enchantment, the cycled card is not this
// card (its own cycling draws normally) and it is not a land card. The
// substitute is the "exile until" shape (game.ExileUntilCastFreeForEffect):
// cards are exiled from the top until one shares a card type with the
// cycled card, that card may be cast free during the resolution (CR
// 608.2g), and the rest (with the hit, if it isn't cast) go to the bottom in
// a random order. A typecycling ability draws nothing, so it is never
// replaced.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     UnpredictableCycloneOracleID,
		Name:         "Unpredictable Cyclone",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
		Replacements: []game.ReplacementEffect{{
			Watches:     []game.EventKind{game.EventDrawCard},
			AppliesTo:   unpredictableCycloneApplies,
			DrawInstead: game.RegisterDrawInstead("unpredictable-cyclone", unpredictableCycloneInstead),
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Unpredictable Cyclone: exile until a card that shares a card type with the cycled card instead",
		}},
	})
}

// UnpredictableCycloneOracleID is the card's oracle_id.
const UnpredictableCycloneOracleID = "5b67a944-ab0b-4155-8bc0-becb1b38b3bb"

// unpredictableCycloneApplies is "a cycling ability of another nonland card
// would cause you to draw a card".
func unpredictableCycloneApplies(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
	if src == nil || ev.Kind != game.RepEventDraw || ev.DrawInstruction || ev.DrawCycled == uuid.Nil {
		return false
	}
	if ev.DrawPlayer != src.Controller || ev.DrawCycled == src.InstanceID {
		return false
	}
	cycled, ok := g.LookupCardForEffect(ev.DrawCycled)
	return ok && !cycled.IsLand()
}

// unpredictableCycloneInstead walks the library to a card sharing a card
// type with the cycled card and offers its free cast; the rest go to the
// bottom in a random order.
func unpredictableCycloneInstead(g *game.Game, drawer, source uuid.UUID, done game.DrawDone) error {
	cycled, ok := g.LookupCardForEffect(done.Cycled(g))
	if ok {
		hit := func(c game.Card) bool { return effects.SharesACardType(c, cycled) }
		if err := g.ExileUntilCastFreeForEffect(drawer, source, hit); err != nil {
			return err
		}
	}
	return done.Run(g)
}
