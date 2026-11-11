package effects

import (
	"errors"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b09_helpers.go — Commander wave 2, batches 9 and 10 (#5368).

// cmdr2B09FlorianLook is Florian, Voldaren Scion's postcombat clause:
// "look at the top X cards of your library, where X is the total amount of
// life your opponents lost this turn. Exile one of those cards and put the
// rest on the bottom of your library in a random order. You may play the
// exiled card this turn." X is counted as the trigger resolves; with X of
// zero, or an empty library, nothing is looked at. The exile is mandatory
// (a floor of one), so the player picks which card of the look they keep.
func cmdr2B09FlorianLook(g *game.Game, item *game.StackItem) error {
	player := item.Controller
	x := opponentLifeLostThisTurn(g, player)
	if x <= 0 {
		return nil
	}
	cards := g.LookAtTopOfLibraryForEffect(player, x)
	if len(cards) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   item.SourceCardID,
		Question: "Florian, Voldaren Scion — exile one of those cards (you may play it this turn); the rest go on the bottom in a random order",
		Cards:    cards,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  cmdr2B09FlorianExile,
		Carry:    cards,
	})
	return nil
}

// cmdr2B09FlorianExile is the answer: the picked card is exiled with a
// permission to play it this turn, and the rest still in the library go to
// the bottom in a random order. An on-disk key: never renamed, never reused.
var cmdr2B09FlorianExile = game.RegisterChooseCardsThen("choose-cards/cmdr2-b09-florian-exile", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	picked := r.Picked[0]
	var rest []uuid.UUID
	for _, id := range cardsStillInALibrary(g, r.Carry) {
		if id != picked {
			rest = append(rest, id)
		}
	}
	if err := g.ExileCardWithPermissionForEffect(picked, game.CastPermission{
		Player: r.Chooser,
		Label:  "Florian, Voldaren Scion — play the exiled card this turn",
	}); err != nil {
		return err
	}
	return g.PutOnBottomInRandomOrderForEffect(r.Chooser, game.ZoneLibrary, rest)
})

// cmdr2B09TimotharOffer is Timothar, Baron of Bats' trigger as it resolves:
// "you may pay {1} and exile it. If you do, create a 1/1 black Bat creature
// token with flying. It gains 'When this token deals combat damage to a
// player, sacrifice it and return the exiled card to the battlefield
// tapped.'" The dead Vampire must still be in a graveyard to be exiled, so
// a card that has already left is not offered.
func cmdr2B09TimotharOffer(g *game.Game, item *game.StackItem) error {
	dead := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(dead); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Timothar, Baron of Bats — pay {1} and exile that Vampire card to create a 1/1 Bat?",
		Then:     cmdr2B09TimotharPaid,
		Carry:    []uuid.UUID{dead},
	}.Apply(NewContext(g, item))
}

// cmdr2B09TimotharPaid exiles the card once {1} is paid; the Bat is made in
// the exile's tail, only for a card that really reached exile ("if you do",
// CR 608.2c; a commander may go to the command zone instead, CR 903.9).
var cmdr2B09TimotharPaid = PayUnlessBranches("pay-unless/cmdr2-b09-timothar-exile",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if len(r.Carry) == 0 {
			return nil
		}
		dead := r.Carry[0]
		return ExileTarget{Target: dead, ThenTail: withCarry(ctx.ItemTail(cmdr2B09TimotharExiled), dead)}.Apply(ctx)
	}, nil)

// cmdr2B09TimotharExiled makes the Bat for an exiled card; its own tail
// links the card to the Bat and grants the Bat its trigger. An on-disk key:
// never renamed, never reused.
var cmdr2B09TimotharExiled = CardTail("cmdr2-b09-timothar/after-exile", func(ctx *Context, o game.TailOutcome) error {
	dead := tailCarry0(o)
	if len(o.IDs) != 1 {
		return nil
	}
	return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: ctx.Controller(),
		Source:     ctx.Source(),
		Groups: []game.TokenGroup{{
			Template: ctx.TextToken(TokenCard("1/1 black Bat with flying")),
			Count:    1,
		}},
	}, withCarry(ctx.ItemTail(cmdr2B09TimotharBat), dead))
})

// cmdr2B09TimotharBat links the exiled card to the new Bat (CR 607.2a) and
// gives the Bat the combat-damage trigger. It gains it with no stated
// duration, so for as long as it is on the battlefield (CR 611.2a). An
// on-disk key: never renamed, never reused.
var cmdr2B09TimotharBat = CardTail("cmdr2-b09-timothar/bat", func(ctx *Context, o game.TailOutcome) error {
	dead := tailCarry0(o)
	if len(o.IDs) != 1 || dead == uuid.Nil {
		return nil
	}
	bat := o.IDs[0]
	ref, ok := ctx.Game.PermanentRefForEffect(bat)
	if !ok {
		return nil
	}
	ctx.Game.RelinkExiledWithForEffect(dead, game.PermissionCardRef(ref))
	return GrantAbilitiesFor{
		Target:   bat,
		Keys:     []string{cmdr2B09TimotharBatKey},
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), bat),
		Label:    "Timothar, Baron of Bats — the Bat returns the exiled card",
	}.Apply(ctx)
})

// cmdr2B09TimotharBatKey is the granted ability's bundle key.
const cmdr2B09TimotharBatKey = "timothar-baron-of-bats/bat-returns"

// cmdr2B09BatReturnsTheCard is the Bat's granted effect: "sacrifice it and
// return the exiled card to the battlefield tapped". The card is read off
// the link first (the Bat is gone after the sacrifice), then the Bat is
// sacrificed, then the card returns tapped under the Bat controller's
// control (CR 110.2a); with no linked card in exile only the sacrifice
// happens.
func cmdr2B09BatReturnsTheCard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var cards []uuid.UUID
	if ref, ok := g.SourceObjectForEffect(item); ok {
		for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
			cards = append(cards, c.InstanceID)
		}
	}
	if onBattlefield(g, item.SourceCardID) {
		if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, id := range cards {
		if err := (ReturnFromExile{Target: id, Controller: ctx.Controller(), Tapped: true}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// cmdr2B09ReturnTheEventCardUnderYourControl is "return that card to the
// battlefield under your control" for a trigger whose event names the card
// (Kaya's Ghostform, Athreos, Shroud-Veiled): the card that left the
// battlefield returns as a new object (CR 400.7) under the trigger's
// controller, and a card that has since left its zone is "nothing to
// return".
func cmdr2B09ReturnTheEventCardUnderYourControl(g *game.Game, item *game.StackItem) error {
	_, err := g.ReturnToBattlefieldForEffect(item.Trigger.Event.CardID, item.Controller, false)
	if errors.Is(err, game.ErrCardNotFound) {
		return nil
	}
	return err
}

// cmdr2B09DrawForEachCounterOnIt is "draw a card for each +1/+1 counter on
// it" when the source has left the battlefield (Toothy, Imaginary Friend,
// Bloodtracker): it counts the +1/+1 counters the permanent had as it left
// (last-known information, CR 608.2h).
func cmdr2B09DrawForEachCounterOnIt(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	return DrawCards{Player: item.Controller, N: info.Counters[game.CounterPlusOne]}.Apply(ctx)
}
