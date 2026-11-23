package effectstest

// theme_deck_graveyard_test.go — S29's exit criterion (issue #94):
//
//	"Theme-deck smoke test (Muldrotha-style graveyard deck casts via
//	 flashback + escape)"
//
// A scripted engine test rather than a bot game. The bots cannot take
// any of these casts: legal/cast.go enumerates hand and command-zone
// casts only and never offers an alternative cost, so a bot-driven
// run would pass without ever touching the graveyard. The bot version
// waits on #89.
//
// The harness is the S27 smoke test's (theme_deck_smoke_test.go), for
// the same reasons:
//
//  1. THE CARDS COME OFF THE IMPORT ROAD. Every card is a cards.Card
//     row passed through deck.ToGameCard (themeDeckRow and
//     importThemeCard), so the flashback and escape offers are found
//     by the oracle ID the importer stamped, not by a field the test
//     set.
//
//  2. THE MANA IS REAL. Every cast, from hand or from the graveyard,
//     goes through Strict + AutoTap. A flashback or escape cost the
//     deck cannot pay fails here instead of resolving for free, and
//     the colours matter: Lingering Souls' flashback is black and its
//     printed cost is white, and there is no Plains on the table.
//
//  3. THE TURNS ARE REAL. Three of the theme seat's turns at a
//     four-player table, with the other three seats' turns in
//     between. The graveyard carries from turn to turn because
//     nothing resets it.
//
// The concessions: nine lands start on the battlefield rather than
// being played one per turn (a land drop per turn would make the
// test about the mana curve), and the opening hand of fillers goes
// back into the library so the hand holds exactly the four spells and
// no cleanup discard gets in the way. The graveyard is NOT seeded:
// every card in it got there by being milled, discarded or cast.
//
// WHAT IT ASSERTS
//
//	Flashback  cast from the graveyard for the flashback cost; the
//	           card is exiled as it leaves the stack (CR 702.34a), so
//	           it is gone from the graveyard and cannot be cast again
//	           on a later turn; casting from the graveyard does not
//	           lift a sorcery's timing (CR 117.1a).
//	Escape     cast from the graveyard for the escape cost, exiling
//	           exactly the N other cards named (CR 702.138a); an
//	           escaped sorcery goes back to the graveyard on
//	           resolution and can escape again next turn; cards
//	           already exiled to pay one escape cannot pay the next;
//	           a graveyard too thin to pay is refused and the card
//	           stays put.
//
// WHAT IS DELIBERATELY NOT HERE
//
// Granted flashback and escape (Snapcaster Mage, Past in Flames,
// Underworld Breach), "unless it escaped" (Uro, Kroxa, Phlage) and
// madness: none of them has an engine primitive yet, which is the
// tail #94 splits into its own issues. A flashback spell that is
// countered or fizzles is pinned in the game package
// (flashback_test.go there); this file only drives the resolution
// road.

// The oracle keys for the three cards no other test in this package
// names as a constant. Faithless Looting's is faithlessLootingOracle
// in pirates_test.go.
const (
	GyThemeLingeringSoulsOracle = "0b8c3337-04dd-4798-8203-6d8b8cfb936b"
	GyThemeFruitOfTizerusOracle = "8d6ad0a0-3b71-4fab-8874-470285c39299"
	GyThemeSweetOblivionOracle  = "c023538d-2feb-4af5-a44e-b355a190f081"
)
