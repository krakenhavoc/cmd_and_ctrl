package game

import "github.com/google/uuid"

// builtin_replacements.go holds the engine-owned replacement
// effects — the ones that live in the rules themselves, not on
// any particular card. Registered on every game at NewGame.
//
// Today: just commanderZoneReplacement (CR 903.9), refactored from
// S13.1's inline applyCommanderZoneReplacementLocked. Future
// built-ins (e.g. the "if a card would be exiled by an effect that
// replaces it with something else" meta-replacement) would live
// here too.
//
// Added in S17 sub-PR 2.

// commanderZoneReplacement implements CR 903.9: "if a commander
// would be put into a library, hand, graveyard, or exile from
// anywhere, its owner may put it into the command zone instead."
//
// Personal-sandbox simplification: we apply the replacement
// automatically rather than prompting the owner. Matches today's
// S13.1 behaviour byte-for-byte.
//
// Controlled by the commander's owner (drives CR 616 affected-
// player ordering if another commander-zone-touching replacement
// ever joins — today there's only one).
//
// The commander zone move is flagged by ev.asCommanderMove — set
// by MoveCardByIDAsCommander when the caller passes asCommander =
// true. The other MoveCardByID path (asCommander=false) routes
// commanders to graveyard / exile normally, preserving S13.1's
// semantics for admin moves.
var commanderZoneReplacement = ReplacementEffect{
	Watches: []EventKind{EventZoneMove},
	AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
		if ev.Kind != RepEventMove {
			return false
		}
		if !ev.asCommanderMove {
			return false
		}
		switch ev.NewZone {
		case ZoneGraveyard, ZoneExile, ZoneHand, ZoneLibrary:
			// Eligible destination.
		default:
			return false
		}
		card, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return false
		}
		return card.IsCommander
	},
	Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
		card, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return nil
		}
		ev.NewZone = ZoneCommand
		ev.NewZoneOwner = card.Owner
		return nil
	},
	Controller: func(ev *ReplacementEvent, g *Game, _ *Card) uuid.UUID {
		card, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			return uuid.Nil
		}
		return card.Owner
	},
	Label: "Route commander to command zone",
}
