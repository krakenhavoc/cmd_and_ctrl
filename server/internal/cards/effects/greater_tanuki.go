package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greater Tanuki — Enchantment Creature — Dog {4}{G}{G}, 6/5 (Duskmourn
// Commander, #5368):
//
//	"Trample
//	 Channel — {2}{G}, Discard this card: Search your library for a basic land
//	 card, put it onto the battlefield tapped, then shuffle."
//
// Channel is an activated ability that works from the hand (Boseiju, Who
// Endures): the discard is a cost, paid with the ability already on the stack
// (CR 602.2b), and the search is the shared basic-land fetch Topiary Stomper
// uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b2d19232-9a80-4ea2-9aa7-c47b4719106c",
		Name:            "Greater Tanuki",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []ActivatedAbility{{
			Label:   "Channel — {2}{G}, Discard this card: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
			Cost:    game.AbilityCost{Mana: "{2}{G}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Purpose: game.Purpose{Lands: 1},
			Effect:  cmdr2B15SearchBasicLand("Greater Tanuki — a basic land card, onto the battlefield tapped", true),
		}},
	})
}
