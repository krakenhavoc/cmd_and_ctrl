package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Krark, the Thumbless — Legendary Creature — Goblin Wizard {1}{R}, 2/2:
//
//	"Whenever you cast an instant or sorcery spell, flip a coin. If you
//	 lose the flip, return that spell to its owner's hand. If you win
//	 the flip, copy that spell, and you may choose new targets for the
//	 copy.
//	 Partner (You can have two commanders if both have partner.)"
//
// The trigger goes on the stack above the spell, so the table can
// respond to it. The flip is the engine's call-and-win/loss coin
// (game.FlipCoinForEffect): a lost flip bounces the spell without
// countering it (ReturnSpellToHand), a won flip copies it
// (CopySpell, CR 707.10) and offers the CR 707.10c re-target. A spell
// that has already left the stack by the time the coin lands is
// neither returned nor, on a win, lost to the copy: "copy that spell"
// names it and does not target it, so the copy is made from last-known
// information (CR 608.2h), as Doublecast's is.
//
// Partner is a deck-building rule (CR 702.124), enforced where decks
// are built, not on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "18e421c0-2879-4ad8-b027-397f168dfbc2",
		Name:         "Krark, the Thumbless",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()), "Krark — flip a coin: lose returns the spell, win copies it", krarkFlip),
		},
	})
}

func krarkFlip(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	spell := ctx.Trigger().Event.CardID
	if spell == uuid.Nil {
		return nil
	}
	controller, source := ctx.Controller(), ctx.Source()
	self, kind := item.ID, item.Kind
	g.FlipCoinForEffect(game.CoinFlipSpec{
		Flipper:  controller,
		Source:   source,
		Question: "Krark, the Thumbless — call the coin flip",
		Then: func(g *game.Game, result game.CoinFlipResult) error {
			c := NewContext(g, &game.StackItem{
				ID: self, Kind: kind,
				Controller: controller, Owner: controller, SourceCardID: source,
			})
			if len(result.Won) > 0 && result.Won[0] {
				return CopySpell{
					StackID:          spell,
					Controller:       controller,
					ChooseNewTargets: true,
					FromLastKnown:    true,
				}.Apply(c)
			}
			return ReturnSpellToHand{StackID: spell}.Apply(c)
		},
	})
	return nil
}
