package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runeforge Champion — Creature — Dwarf Warrior {2}{W}, 2/3 (Kaldheim):
//
//	"When this creature enters, you may search your library and/or graveyard
//	 for a Rune card, reveal it, and put it into your hand. If you search your
//	 library this way, shuffle.
//	 You may pay {1} rather than pay the mana cost for Rune spells you cast."
//
// The search is "you may", across the library and/or graveyard (the searcher
// names the zones first, and the library is shuffled only if it was searched).
// The static grants every Rune spell its controller casts a {1} alternative
// cost (CR 118.9), through the granted-alternative-cost seam with a Rune
// subtype filter (PermissionFilter.SpellSubtype).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                "2705a288-3b9d-4452-a320-28b86331aa09",
		Name:                    "Runeforge Champion",
		Completeness:            effects.CompletenessFull,
		Purpose:                 game.Purpose{Tutors: 1},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{effects.PayOneForRuneSpellsYouCast()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Runeforge Champion — you may search your library and/or graveyard for a Rune card, reveal it, and put it into your hand; if you search your library this way, shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     func(c game.Card) bool { return c.HasSubtype("Rune") },
						Dest:          game.ZoneHand,
						Limit:         1,
						Optional:      true,
						Reveal:        true,
						Shuffle:       true,
						AlsoGraveyard: true,
						Reason:        "Runeforge Champion — a Rune card",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
