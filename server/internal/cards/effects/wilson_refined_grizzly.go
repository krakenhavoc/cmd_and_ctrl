package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wilson, Refined Grizzly — 2/2 Legendary Creature — Bear Warrior for
// {1}{G} (EDHREC rank 3988):
//
//	"This spell can't be countered.
//	 Reach, vigilance, trample
//	 Ward {2}
//	 Choose a Background (You can have a Background as a second
//	 commander.)"
//
// Four keywords and an uncounterable rider on a two-mana 2/2, which
// is why the Bear is a commander rather than a creature: as a
// two-drop general with ward {2} that resists counterspells, it is a
// Voltron shell that is genuinely hard to interact with.
//
// It is in the batch because it is the densest keyword stack in the
// batch and because the uncounterable rider and ward together are the
// two halves of "hard to answer" that are implemented in completely
// different places — one on the Spec, one as a triggered ability.
//
// # "This spell can't be countered" is the card's OWN rider
//
// It goes on the Spec, not in the layer system: a continuous effect
// cannot reach the stack, and this clause is about the Bear while it
// is a spell. Countering it is still a LEGAL play — the counterspell
// resolves and does nothing (CR 701.6a), which is observably
// different from the counterspell fizzling for lack of a target.
//
// # Ward {2}
//
// A triggered ability, not a keyword the block checker reads: when an
// opponent's spell or ability targets the Bear, that player is asked
// for {2} and the spell or ability is countered if they decline. It
// fires on the TARGETING, so it triggers even for a spell that is
// later countered, and it does nothing about the Bear's own
// controller targeting it.
//
// # Declared simplification (weaker than printed): no Background
//
// "Choose a Background" is a deckbuilding permission — it lets the
// Bear share the command zone with a Background enchantment as a
// second commander (CR 702.124b's partner family). The engine has one
// commander slot and no partner machinery, so the Bear is a
// single commander and a Background cannot ride along with it.
//
// That is strictly LESS than the card offers: the deck loses a second
// command-zone card and whatever colours it would have added, and
// nothing about it lets a deck do something paper would not. It costs
// the Bear nothing on the battlefield — every printed ability above
// works — it only costs the deck the second commander.
func init() {
	Register(Spec{
		OracleID:        "d2766fd7-5cf9-4037-9f34-9ae3982c613a",
		Name:            "Wilson, Refined Grizzly",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The Background half isn't available — Wilson can be your commander, but not alongside a Background as a second one."},
		CantBeCountered: true,
		PrintedKeywords: []string{"reach", "vigilance", "trample"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Wilson, Refined Grizzly — ward {2}"),
		},
	})
}
