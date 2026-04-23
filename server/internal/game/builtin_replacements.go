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
// CR 614.10 "may" replacement — Optional=true triggers the apply-
// loop's yes/no prompt path (PendingChoiceOptionalReplacement) so
// the owner decides each time. Sub-PR 6 change: widened AppliesTo
// (dropped the asCommanderMove gate) so the replacement fires for
// every move that puts a commander into graveyard/exile/hand/
// library — spell-driven (Wrath, Path, Swords), SBA-driven (dies
// to damage), bounce (Unsummon), mill. S13.1's flag-gated single-
// path implementation missed every case except the admin "move as
// commander" UI action.
//
// Controlled by the commander's owner (drives both the CR 614.10
// yes/no prompt and the CR 616 multi-replacement order prompt if
// other commander-zone-touching replacements ever join).
var commanderZoneReplacement = ReplacementEffect{
	Watches:        []EventKind{EventZoneMove},
	Optional:       true,
	PromptQuestion: "Send commander to command zone instead?",
	AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
		if ev.Kind != RepEventMove {
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
	Label: "Commander zone replacement",
}
