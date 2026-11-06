package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// dragonstorm.go — the Tarkir: Dragonstorm cycle's shared second ability
// (Encroaching Dragonstorm, Breaching Dragonstorm, #4463):
//
//	"When a Dragon you control enters, return this enchantment to its
//	 owner's hand."
//
// Any Dragon entering under your control (effective subtypes, so a
// changeling Dragon counts, and a Dragon token too) returns the
// enchantment if it is still on the battlefield when the trigger
// resolves. Append-only: a new Dragonstorm calls this.
func WhenADragonYouControlEntersReturnThis(name string) game.TriggeredAbility {
	return On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return b23DragonYouControlEntered(ev, source, g)
	}, name+" — return it to its owner's hand", func(g *game.Game, item *game.StackItem) error {
		if !onBattlefield(g, item.SourceCardID) {
			return nil
		}
		return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
	})
}
