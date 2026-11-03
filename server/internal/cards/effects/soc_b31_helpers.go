package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b31_helpers.go — shared readers for Secrets of Strixhaven batch 31
// (#4456).

// socB31LandOrHasXInManaCost is "a land card or a card with {X} in its mana
// cost" (Paradox Surveyor): the cost printed on the card, parsed, so an {X}
// in an additional or alternative cost does not count.
func socB31LandOrHasXInManaCost() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		if c.IsLand() {
			return true
		}
		cost, err := game.ParseCost(c.ManaCost)
		return err == nil && cost.XSlots > 0
	}
}

// socB31ChooseALandCardName is the resolution of "choose a land card name"
// (Petrified Hamlet): it queues the ordinary card-name prompt for the
// ability's controller, and the answer is stored on the source permanent
// (ChosenName) as Pithing Needle's is. There is no list of real card names to
// check the answer against (game/choose_card_name.go), so "land card" is kept
// where the name is read: the two clauses that read it only reach lands.
type socB31ChooseALandCardName struct{ Label string }

func (s socB31ChooseALandCardName) Apply(ctx *Context) error {
	ctx.Game.QueueCardNameChoiceForEffect(ctx.Controller(), ctx.Source(), s.Label)
	return nil
}

// socB31LandFaceNamed reports whether `c` has a land face called `named`
// (case-insensitive, CR 201.2b): the active face by the card's printed type
// line, every other face by its own. phyrexian_revoker.go's nonlandFaceNamed
// turned the other way round.
func socB31LandFaceNamed(c game.Card, named string) bool {
	if c.FaceDown {
		return false
	}
	named = strings.TrimSpace(named)
	if named == "" {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(c.Name), named) && c.PrintedIsLand() {
		return true
	}
	for _, f := range c.Faces {
		if strings.EqualFold(strings.TrimSpace(f.Name), named) && faceTypeLineIsLand(f.TypeLine) {
			return true
		}
	}
	return false
}
