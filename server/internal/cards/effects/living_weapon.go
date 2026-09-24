package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// living_weapon.go — CR 702.91a, "Living weapon": "When this
// Equipment enters the battlefield, create a 0/0 black Phyrexian
// Germ creature token, then attach this Equipment to it."
//
// Not an engine seam. The keyword is a printed triggered ability
// spelled out in reminder text, and every part of it already exists:
// WhenThisEnters puts it on the stack so it can be responded to, the
// token table holds the Germ, and game.AttachSourceForEffect is the
// same attach the equip ability resolves with. The shared helper is
// here so all three living-weapon cards get one implementation —
// Batterskull, Batterbone and Kaldra Compleat.
//
// # Why it is one effect and not two lines
//
// A 0/0 with nothing attached to it dies to the CR 704.5f
// state-based action. The Germ only survives because the Equipment
// lands on it before state-based actions are next checked, and both
// halves run inside this one ability's resolution — so the token and
// the attach are never separated by a priority boundary.
//
// That is also why the attach rides
// game.CreateTokensThenForEffect's continuation rather than the next
// line of Go. Token creation goes through the CR 701.7b replacement
// window, and a window with two different applicable effects in it
// (a Doubling Season and an Academy Manufactor) pauses for a CR 616
// ordering prompt; the IDs are not known until it is answered.
// Reading CreateTokensForEffect's slice on the next line would work
// on an empty board and quietly make an unattached Germ on a board
// with a token doubler on it.
//
// # The two ways it does less than the reminder text says
//
// Both are the rules, not simplifications. If the Equipment has left
// the battlefield in response to the trigger, nothing is attached
// (CR 701.3b) and the Germ dies on the next state-based check — the
// ability still resolved. The same goes for an Equipment that left and
// came back: it is a new object (CR 400.7), and the trigger names the
// one that entered (StackItem.SourceObject, #1418). And if some effect stops the token being
// created, there is nothing to attach to and the ability does as
// much as it can.

// livingWeaponGermToken is the token every living-weapon card makes.
// The key is spelled the way the cards print it.
const livingWeaponGermToken = "0/0 black Phyrexian Germ"

// LivingWeapon is the living weapon keyword ability for `cardName`.
// The label names the card, so a stack with two of them on it says
// which is which.
func LivingWeapon(cardName string) game.TriggeredAbility {
	return WhenThisEnters(cardName+" — living weapon: create a Germ and attach this to it",
		livingWeaponCreateGermAndAttach)
}

// livingWeaponCreateGermAndAttach is the resolution. A package-level
// func rather than a closure per card: it captures nothing, reads
// everything off the item it is handed, and is therefore safe across
// the Clone an undo restores.
func livingWeaponCreateGermAndAttach(g *game.Game, item *game.StackItem) error {
	return g.CreateTokensThenForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups: []game.TokenGroup{{
			Template: TokenCard(livingWeaponGermToken),
			Count:    1,
		}},
	}, func(g *game.Game, created []uuid.UUID) error {
		if len(created) == 0 {
			return nil
		}
		return g.AttachSourceForEffect(item, game.TargetRef{Kind: game.TargetCard, ID: created[0]})
	})
}
