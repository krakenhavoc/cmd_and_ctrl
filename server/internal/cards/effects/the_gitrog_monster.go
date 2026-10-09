package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Gitrog Monster — Legendary Creature — Frog Horror {3}{B}{G},
// 6/6:
//
//	"Deathtouch
//	 At the beginning of your upkeep, sacrifice The Gitrog Monster
//	 unless you sacrifice a land.
//	 You may play an additional land on each of your turns.
//	 Whenever one or more land cards are put into your graveyard from
//	 anywhere, draw a card."
//
// Deathtouch is PrintedKeywords; the extra land drop is
// `Spec.AdditionalLandPlays` (Exploration's whole card). The other two
// clauses:
//
//   - The upkeep clause is a plain two-way choice, not a CR 118.12
//     "unless you pay" (that primitive is for a MANA payment, and this
//     isn't one) — a `PickOption` between "sacrifice a land" and
//     "sacrifice The Gitrog Monster" when there is a land to offer,
//     and a bare `SacrificePermanent` on itself when there is not.
//     Either branch is a real `PendingChoice`, so the upkeep step
//     cannot advance past it unanswered.
//   - "Land cards put into your graveyard from anywhere" watches both
//     zone-motion event kinds the engine emits for a graveyard
//     landing — `EventLTB` for a battlefield exit (including the
//     card's own upkeep sacrifice) and `EventZoneMove` for everything
//     else (discard, mill) — and reads the card AFTER it lands to ask
//     whether it's a land the controller owns. `OncePerBatch`
//     collapses a simultaneous multi-land dump into the one trigger
//     CR 603.2c's "one or more" describes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "a5e54d2b-aad8-4ddd-af4b-13668913762b",
		Name:                "The Gitrog Monster",
		Completeness:        CompletenessFull,
		PrintedKeywords:     []string{"deathtouch"},
		AdditionalLandPlays: 1,
		Purpose:             game.Purpose{ExtraLandDrops: 1}, // #2678: the bot reads the extra drop
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("The Gitrog Monster — sacrifice this unless you sacrifice a land", gitrogUpkeepEffect),
			OncePerBatch(OnAny([]game.EventKind{game.EventZoneMove, game.EventLTB}, gitrogLandToGraveyard,
				gitrogDrawLabel, Do(DrawCards{N: 1}))),
		},
	})
}

const gitrogDrawLabel = "The Gitrog Monster — draw a card"

// gitrogLandToGraveyard is "a land card is put into your graveyard
// from anywhere" — no OldZone check, which is the "from anywhere"
// itself: it fires whether the land came from the battlefield, hand
// or library.
func gitrogLandToGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.NewZone != game.ZoneGraveyard {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsLand() && c.Owner == source.Controller
}

// gitrogUpkeepEffect is "sacrifice The Gitrog Monster unless you
// sacrifice a land" (sacrifice_unless_land.go).
var gitrogUpkeepEffect = sacrificeThisUnlessYouSacrificeALand("The Gitrog Monster")
