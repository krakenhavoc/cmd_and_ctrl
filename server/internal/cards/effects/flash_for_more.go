package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// flash_for_more.go — "You may cast this spell as though it had flash if
// you pay {N} more to cast it" (Asinine Antics, Oakshade Stalker,
// Harbinger of the Tides, Mystical Tether; #3172, #5207).
//
// CR 601.3c: "If an effect allows a player to cast a spell as though it
// had flash only if an alternative or additional cost is paid, that
// player may begin to cast that spell as though it had flash." The {N}
// is an optional additional cost the caster announces with the spell (CR
// 601.2b), and announcing it is what opens the instant-speed window for
// THAT cast (AdditionalCost.AsThoughFlash, Quantum Reduction's shape).
// The flash is a timing permission and the {N} an additional cost, so
// neither replaces the mana cost: a flash cast may still claim any one
// alternative cost (CR 118.9a), and still pays the {N} on top of it (CR
// 601.2f), a "without paying its mana cost" cast included.
//
// Until #5207 this was an alternative-cost offer at the printed cost plus
// {N}, which CR 118.9a then refused to combine with a second alternative
// cost; coverage's TestNoFlashIfClauseIsAnAlternativeCost now forbids
// that shape.

// FlashForMoreKey is the optional cost's identity on the card.
const FlashForMoreKey = "flash-for-more"

// FlashForMore is the optional cost. `more` is the surcharge as printed
// ("{2}").
//
//	OptionalCosts: []game.AdditionalCost{FlashForMore("{2}")},
func FlashForMore(more string) game.AdditionalCost {
	return game.AdditionalCost{
		Optional:      true,
		Key:           FlashForMoreKey,
		Label:         "Pay " + more + " more to cast it as though it had flash",
		ManaCost:      more,
		AsThoughFlash: true,
	}
}
