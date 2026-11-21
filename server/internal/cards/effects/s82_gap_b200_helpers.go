package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// GapB200RippleTrigger is "Spells you cast have ripple N" (CR 702.60a) as
// the trigger of the permanent that gives it: whenever its controller casts
// a spell, they may reveal the top N cards of their library, may cast the
// spells among them with the same name as that spell without paying their
// mana costs, and put the rest on the bottom of their library in any order.
func GapB200RippleTrigger(name string, n int) game.TriggeredAbility {
	label := name + " — ripple " + strconv.Itoa(n)
	return WheneverYouCast(nil, label, func(g *game.Game, item *game.StackItem) error {
		return MayChoice{
			Question: label + ": reveal the top " + strconv.Itoa(n) + " cards of your library?",
			Then:     gapB200RippleAnswer,
			Word:     strconv.Itoa(n),
		}.Apply(NewContext(g, item))
	})
}

// gapB200RippleAnswer is the "yes": reveal, grant a free cast over each card
// with the spell's name and offer them while the trigger resolves (CR 608.2g,
// each its own "may"), then the rest goes to the bottom. Keyed.
var gapB200RippleAnswer = MayYes("confirm/s82-gap-b200-ripple-reveal", func(ctx *Context, r game.ConfirmAnswer) error {
	g := ctx.Game
	you, source := ctx.Controller(), ctx.Item.SourceCardID
	revealed := g.RevealTopOfLibraryForEffect(you, source, WordInt(r), "ripple")
	if len(revealed) == 0 {
		return nil
	}
	spellName := gapB200RippleSpellName(ctx)
	perm := g.FreeCastFromLibraryPermission(you, source, "Ripple")
	perm.LapseOnPass = game.LapseStaysInHand
	var granted []uuid.UUID
	for _, id := range revealed {
		c, ok := g.LookupCardForEffect(id)
		if !ok || spellName == "" || c.Name != spellName || !perm.OffersASpell(c) {
			continue
		}
		if g.GrantCastPermissionOverCardForEffect(id, perm) {
			granted = append(granted, id)
		}
	}
	if len(granted) > 0 && g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       source,
		Reason:       "Ripple",
		Cards:        granted,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put the rest on the bottom",
		Then:         gapB200RippleBottom,
		Carry:        revealed,
	}) {
		return nil
	}
	return TakeRestOnBottomInAnyOrder(g, TakeFromLibraryResult{Player: you, Source: source, Rest: revealed})
})

// gapB200RippleSpellName is the name of the spell whose cast triggered the
// ripple, as it was cast (CR 603.10).
func gapB200RippleSpellName(ctx *Context) string {
	tc := ctx.Trigger()
	if tc.Object != nil && tc.Object.Name != "" {
		return tc.Object.Name
	}
	if c, ok := ctx.Game.LookupCardForEffect(tc.Event.CardID); ok {
		return c.Name
	}
	return ""
}

// gapB200RippleBottom is "put the rest on the bottom of your library in any
// order", run once the casts are made: the revealed cards that were not cast
// (and are still in the library) go to the bottom. Keyed.
var gapB200RippleBottom = game.RegisterChooseCardsThen("resolution-cast/s82-gap-b200-ripple-bottom",
	func(g *game.Game, r game.CardsChosen) error {
		cast := make(map[uuid.UUID]bool, len(r.Picked))
		for _, id := range r.Picked {
			cast[id] = true
		}
		var rest []uuid.UUID
		for _, id := range r.Carry {
			if !cast[id] {
				rest = append(rest, id)
			}
		}
		if len(rest) == 0 {
			return nil
		}
		return TakeRestOnBottomInAnyOrder(g, TakeFromLibraryResult{Player: r.Chooser, Source: r.Source, Rest: rest})
	})
