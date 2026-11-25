package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// named_card_copy.go — the catalog side of "Create a copy of the card with
// the chosen name. You may cast the copy." (#6155, CR 707.13, CR 707.12).
// The engine half is game/named_card_copy.go: the copy is built from the
// card's printed record (conjurable_cards.go, the engine's Oracle card
// reference), created outside the game, and offered for a cast during the
// resolution at its own costs.

// CheckNamedCardCopies is the boot guard for a card that copies cards by
// name: every name must have a printed record, filed with the engine
// before any game can ask for it. It returns names unchanged.
func CheckNamedCardCopies(card string, names []string) []string {
	registerConjurables()
	for _, n := range names {
		if _, ok := game.ConjurableTemplate(n); !ok {
			panic(fmt.Sprintf("effects: %s copies %q by name, which has no record in conjurable_cards.go", card, n))
		}
	}
	return names
}

// CopyNamedCardAndOfferCast creates a copy of the card named Name outside
// the game for ctx's controller and offers its cast while the effect is
// still resolving, paying its own costs ("You may cast the copy. (You
// still pay its costs.)"). A copy nobody casts ceases to exist
// (CR 707.10a).
func CopyNamedCardAndOfferCast(ctx *Context, sourceName, name string) (uuid.UUID, error) {
	return ctx.Game.CopyNamedCardAndOfferCastForEffect(game.CardCopyOffer{
		Player:           ctx.Controller(),
		Source:           ctx.Source(),
		SourceName:       sourceName,
		Timing:           game.TimingFlash,
		DuringResolution: true,
	}, name)
}
