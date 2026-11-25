package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// spend_as_though.go — #6202: the narrower "spend mana as though"
// statics (CR 609.4b). The one-spell "any type for its mana cost"
// (North Star) is a NextSpellPromise{ManaCostAnyType: true}, granted
// with GrantNextSpellPromise.

// YouMaySpendColorAsColor is "You may spend <from> mana as though it
// were <to> mana." (Sunglasses of Urza: "W", "R"). Uppercase WUBRG
// letters.
func YouMaySpendColorAsColor(from, to string) []game.AnyColorSpendStatic {
	return []game.AnyColorSpendStatic{{
		Label: "You may spend " + colorWords([]string{from}) + " mana as though it were " + colorWords([]string{to}) + " mana.",
		Whose: game.AnyColorSpendYou,
		From:  from,
		To:    to,
	}}
}

// YouMaySpendColorAsAnyOthersOnlyColorless is "You may spend <from>
// mana as though it were mana of any color. You may spend other mana
// only as though it were colorless mana." (Celestial Dawn: "W").
func YouMaySpendColorAsAnyOthersOnlyColorless(from string) []game.AnyColorSpendStatic {
	return []game.AnyColorSpendStatic{{
		Label:               "You may spend " + colorWords([]string{from}) + " mana as though it were mana of any color. You may spend other mana only as though it were colorless mana.",
		Whose:               game.AnyColorSpendYou,
		From:                from,
		OthersOnlyColorless: true,
	}}
}
