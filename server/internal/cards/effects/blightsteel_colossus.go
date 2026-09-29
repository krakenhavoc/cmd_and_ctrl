package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blightsteel Colossus — Artifact Creature — Phyrexian Golem {12},
// 11/11:
//
//	"Trample, infect, indestructible
//	 If Blightsteel Colossus would be put into a graveyard from
//	 anywhere, reveal Blightsteel Colossus and shuffle it into its
//	 owner's library instead."
//
// All three keywords ride PrintedKeywords — infect (CR 702.90, the
// engine's damage tail, ADR 0056) is live since S25 and needs no
// per-card machinery beyond the token itself.
//
// The self-replacement is Stone of Erech's shape one instruction over:
// a `RepEventMove` whose destination is rewritten before the move
// happens, with no `OldZone` check at all — that omission IS "from
// anywhere" (#539 made every zone exit open this same window, not
// only a battlefield death).
//
// Declared simplification, weaker than printed: the card lands on TOP
// of its owner's library rather than being shuffled in. The
// replacement pipeline rewrites a move's destination zone before the
// card arrives there; nothing in it can shuffle a library AFTER the
// fact, and there is no hook to run one once this card has actually
// landed. A library-top landing is public knowledge no player would
// have had if it had been mixed in — a real, if narrow, information
// leak, and it is the reason this ships as a caveat rather than as
// the printed card. "Reveal" is not modelled separately; it is
// cosmetic once the destination is already known to be the library's
// top.
func init() {
	Register(Spec{
		OracleID:        "e80772e2-8623-4094-81a2-70828b2b151c",
		Name:            "Blightsteel Colossus",
		Completeness:    CompletenessCaveats,
		PrintedKeywords: []string{"trample", "infect", "indestructible"},
		Caveats: []string{
			"When this would go to a graveyard from anywhere, it's put on top of its owner's library instead of being shuffled in.",
		},
		Replacements: []game.ReplacementEffect{{
			Watches:         []game.EventKind{game.EventZoneMove},
			SelfReplacement: true,
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.CardID == src.InstanceID && ev.NewZone == game.ZoneGraveyard
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.NewZone = game.ZoneLibrary
				ev.NewZoneOwner = uuid.Nil
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Blightsteel Colossus: into its owner's library instead",
		}},
	})
}
