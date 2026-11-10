package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b53_helpers.go — shared seams for Standard batch 53 (#4452).

// std53FaerieControlledBy is 1 when controller controls a Faerie as it is
// read now (effective subtypes, so a changeling counts), else 0: the
// "as you cast this spell" Faerie check of Faerie Fencing.
func std53FaerieControlledBy(g *game.Game, controller uuid.UUID) int {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype("Faerie") {
			return 1
		}
	}
	return 0
}

// takeOneFromAmongThem is "put a <match> card from among them into your
// hand" with no "may" (Free the Fae): mayTakeOneFromAmongThem's candidates
// (cards the mill actually put in the graveyard that match) with the pick's
// floor at one, so the player must take a card when one is eligible.
func takeOneFromAmongThem(ctx *Context, milled []uuid.UUID, match CardPredicate, question string) error {
	player := ctx.Controller()
	var candidates []uuid.UUID
	for _, id := range milled {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || !match(ctx.Game, player, c) {
			continue
		}
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		candidates = append(candidates, id)
	}
	return takeOneIntoYourHand(ctx, candidates, question, 1)
}

// putDiedThisOnBottomOfLibrary is "When this creature dies, put it on the
// bottom of its owner's library" (Murderous Rider, Fell Horseman): the
// card moves only while it is still the graveyard object that died
// (CR 400.7), through the tuck route, so a commander's owner is offered
// the command zone (CR 903.9) and a token, which ceased to exist
// (CR 111.7), is not found.
func putDiedThisOnBottomOfLibrary(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard || !sameGraveyardObject(z, item) {
		return nil
	}
	return PutIntoLibrary{Card: item.SourceCardID, ToBottom: true}.Apply(NewContext(g, item))
}
