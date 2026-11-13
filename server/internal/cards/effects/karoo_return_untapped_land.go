package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// karoo_return_untapped_land.go — the Visions land cycle (#5631): Coral
// Atoll, Dormant Volcano, Everglades, Jungle Basin and Karoo all print
//
//	"This land enters tapped.
//	 When this land enters, sacrifice it unless you return an untapped
//	 <basic land type> you control to its owner's hand.
//	 {T}: Add {C}{<colour>}."
//
// with a different land type and colour. The tapped entry is the CR 614
// self-replacement the Ravnica karoos use (SelfEntersTapped,
// bounce_lands.go). The trigger is returnLandsOrSacrifice
// (return_lands_or_sacrifice.go) with one land: the choice is made as it
// resolves, declining or having no untapped land of the type sacrifices
// this land, and it stays only if the returned land really reached a
// hand. The land type is a word of the card's text, read through it
// (CR 612.2), and it is the type, not the name, so an untapped Tropical
// Island pays Coral Atoll's cost. The land itself entered tapped, so it
// can never pay for itself. {T}: Add {C}{X} is one mana ability
// (CR 605.1a) adding two mana.

// untappedLandOfType is the payment match of "return an untapped <type>
// you control": an untapped land with the land type `word`, the word
// read through the source's text when the source is still here.
func untappedLandOfType(word string) func(source *game.Card, land game.Card) bool {
	return func(source *game.Card, land game.Card) bool {
		sub := word
		if source != nil {
			sub = source.TextSubtype(word)
		}
		return !land.Tapped && land.HasSubtype(sub)
	}
}

// karooReturningUntapped is one land of the cycle: `landType` is the
// basic land type its trigger asks for, `color` the colour letter its
// mana ability adds beside {C}, and `ret` the card's own registered
// return (keys are per card).
func karooReturningUntapped(oracleID, name, landType, color string, ret *returnLandsOrSacrifice) Spec {
	produced := "{C}{" + color + "}"
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{ReturnsLandOnEntry: true},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: produced,
			Label:    "{T}: Add " + produced + ".",
		}},
		Triggered: []game.TriggeredAbility{
			ret.trigger(name + " — sacrifice it unless you return an untapped " + landType + " you control to its owner's hand"),
		},
	}
}
