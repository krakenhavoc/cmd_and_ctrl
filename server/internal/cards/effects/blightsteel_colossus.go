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
// "From anywhere" also means a DISCARD, and a discard is its own
// ReplacementEventKind (RepEventDiscard, not RepEventMove — ADR 0013
// §5g), pre-filtered by its own Watches key (EventDiscardCard, not
// EventZoneMove). A card that watched only the move kind would let a
// discarded Blightsteel Colossus fall straight into the graveyard, so
// both are declared and AppliesTo checks either Kind: mill and a
// battlefield death both already carry RepEventMove, and discard
// carries RepEventDiscard with the same NewZone/CardID shape (the two
// "share every field the exit path reads", per replacements.go's
// isExitMove).
//
// The card is genuinely shuffled into its owner's library, not merely
// placed on top — ADR 0013 §5ah's ShuffleDestinationLibrary, set here
// alongside the redirected NewZone, is what closes the S17-era caveat
// this card used to carry: the replacement pipeline could rewrite a
// move's destination but had no hook to run a shuffle once the card
// had actually landed, so it shipped on top of the library, a real
// information leak (everyone knows the top card) that "reveal" alone
// does not excuse — the printed card reveals WHICH card, never WHERE
// it ends up.
func init() {
	Register(Spec{
		OracleID:        "e80772e2-8623-4094-81a2-70828b2b151c",
		Name:            "Blightsteel Colossus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "infect", "indestructible"},
		Replacements: []game.ReplacementEffect{{
			Watches:         []game.EventKind{game.EventZoneMove, game.EventDiscardCard},
			SelfReplacement: true,
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				fromMove := ev.Kind == game.RepEventMove || ev.Kind == game.RepEventDiscard
				return fromMove && ev.CardID == src.InstanceID && ev.NewZone == game.ZoneGraveyard
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.NewZone = game.ZoneLibrary
				ev.NewZoneOwner = uuid.Nil
				ev.ShuffleDestinationLibrary = true
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Blightsteel Colossus: shuffled into its owner's library instead",
		}},
	})
}
