package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b07_helpers.go — shared by the Midnight Hunt graveyard batch.

// hasFlashbackOrDisturb is "a card with flashback or disturb": the card
// prints flashback (CR 702.34a) or disturb (CR 702.146a). Flashback is read
// from the printed text the importer stamped, or the catalog's flashback
// cost (game.HasFlashback); disturb is a canonical keyword the importer
// carries, or the catalog's disturb cost. A keyword another effect granted
// is not printed and does not count; a face-down card has neither.
func hasFlashbackOrDisturb(c game.Card) bool {
	if c.FaceDown {
		return false
	}
	return game.HasFlashback(c) || game.HasKeyword(&c, "disturb") ||
		game.AlternativeCostByKey(game.CatalogKey(c), "disturb") != nil
}

// WithFlashbackOrDisturb is hasFlashbackOrDisturb as a card predicate:
// Thraben Exorcism's "creature with disturb" and Phantom Carriage's and
// Shipwreck Sifters' "card with flashback or disturb".
func WithFlashbackOrDisturb() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hasFlashbackOrDisturb(c) }
}

// s82MidB07ExileUpToOneFromAGraveyard is the enters trigger of the two
// scarecrow-and-hippogriff graveyard hate cards: "exile up to one target
// card from a graveyard".
func s82MidB07ExileUpToOneFromAGraveyard(name string) game.TriggeredAbility {
	return Targeting(
		WhenThisEnters(name+" — exile up to one target card from a graveyard", b27ExileChosenTarget),
		TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1))
}
