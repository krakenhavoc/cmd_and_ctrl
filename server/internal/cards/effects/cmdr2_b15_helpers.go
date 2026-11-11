package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b15_helpers.go — shared bodies of the Commander-deck wave 2 batch 15
// (land recursion and self-mill for lands, #5368).

// cmdr2B15MillThenMayReturnLandTrigger is "When this creature enters, mill N
// cards, then you may return a land card from your graveyard to your hand"
// (Eccentric Farmer; Pothole Mole prints the same sentence): the pick is made
// after the mill, so a land just milled is a legal choice.
func cmdr2B15MillThenMayReturnLandTrigger(card string, n int, label string) game.TriggeredAbility {
	return WhenThisEnters(card+" — "+label,
		Do(MillCards{N: n}, Instruction(func(ctx *Context) error {
			return ReturnChosenFromGraveyard{
				Player:   ctx.Controller(),
				Question: card + " — you may return a land card from your graveyard to your hand",
				Match:    game.Card.IsLand,
				Max:      1,
				ToHand:   true,
			}.Apply(ctx)
		})))
}

// cmdr2B15RevealTopTakeAllMatching is "Reveal the top N cards of your library.
// Put all <kind> cards revealed this way into your hand and the rest into your
// graveyard" (Mulch, Winding Way): no prompt, the rest are put, not milled.
func cmdr2B15RevealTopTakeAllMatching(g *game.Game, player, source uuid.UUID, n int, label string, match CardPredicate) error {
	revealed := g.RevealTopOfLibraryForEffect(player, source, n, label)
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  revealed,
		Match:  match,
		All:    true,
		Label:  label,
		Key:    takeRestGraveyard,
	}.applyKeyed(NewContext(g, nil), player, source)
}

// cmdr2B15SearchBasicLand is "Search your library for a basic land card, put it
// onto the battlefield [tapped], then shuffle" (Topiary Stomper, Greater Tanuki's
// channel, Fertile Footsteps). A basic land card is read by supertype, so a snow
// basic qualifies.
func cmdr2B15SearchBasicLand(reason string, tapped bool) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return SearchLibrary{
			Player:        item.Controller,
			Predicate:     IsBasicLand,
			Dest:          game.ZoneBattlefield,
			Limit:         1,
			Shuffle:       true,
			TappedOnEntry: tapped,
			Reason:        reason,
		}.Apply(NewContext(g, item))
	}
}

// cmdr2B15AnOpponentControlsMoreLands is Verge Rangers' condition, "an opponent
// controls more lands than you", read live by the permission's liveness test.
func cmdr2B15AnOpponentControlsMoreLands(g *game.Game, controller, _ uuid.UUID) bool {
	mine := len(g.LandsControlledForEffect(controller))
	for _, p := range g.Seats {
		if p == nil || p.ID == controller || p.Eliminated {
			continue
		}
		if len(g.LandsControlledForEffect(p.ID)) > mine {
			return true
		}
	}
	return false
}
