package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b13_helpers.go — The Brothers' War, S82 batch 13 (#5482): the
// shared bodies of Feldon, Ronom Excavator and Mishra's Research Desk, and
// Terisian Mindbreaker's rounded-up mill.

// s82Bro13ExileChooseOnePlay is "exile the top N cards of your library.
// Choose one of them. Until the end of your next turn, you may play that
// card": every card is exiled face up, the controller picks one, and only it
// gets the window; the rest stay exiled with nothing. N is read as the
// effect resolves (Feldon's is the damage he was dealt).
func s82Bro13ExileChooseOnePlay(ctx *Context, n int) error {
	if n <= 0 {
		return nil
	}
	ids := ctx.Game.LookAtTopOfLibraryForEffect(ctx.Controller(), n)
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(s82Bro13ExileChoose))
}

// s82Bro13ExileChoose is "choose one of them" over the cards that reached
// exile. An on-disk key: never renamed, never reused.
var s82Bro13ExileChoose = CardTail("then/s82-bro-b13-choose-one-play", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	name := "Choose one of them"
	if src := ctx.Source(); src != uuid.Nil {
		if c, ok := ctx.Game.LookupCardForEffect(src); ok {
			name = c.Name
		}
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: name + " — choose one of them (you may play it until the end of your next turn)",
		Cards:    o.IDs,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  s82Bro13ExilePlayGrant,
	})
	return nil
})

// s82Bro13ExilePlayGrant gives the chosen card its window. An on-disk key.
var s82Bro13ExilePlayGrant = game.RegisterChooseCardsThen("choose-cards/s82-bro-b13-play", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player:   r.Chooser,
			Zone:     game.ZoneExile,
			Duration: g.UntilEndOfYourNextTurnDuration(r.Chooser),
			Label:    "Play it until the end of your next turn",
		})
	}
	return nil
})

// s82Bro13MillHalfRoundedUp is "<player> mills half their library, rounded
// up" (Terisian Mindbreaker): b22MillHalf's twin, read as the ability
// resolves.
func s82Bro13MillHalfRoundedUp(ctx *Context, player uuid.UUID) error {
	p := ctx.PlayerByID(player)
	if p == nil || p.Library == nil {
		return nil
	}
	return MillCards{Player: player, N: (p.Library.Size() + 1) / 2}.Apply(ctx)
}

// s82Bro13ArtifactCard and s82Bro13ArtifactCreatureCard are the recipient
// tests of the graveyard grants of Mishra, Tamer of Mak Fawa and Wire
// Surgeons.
func s82Bro13ArtifactCard(card *game.Card) bool { return card.IsArtifact() }

func s82Bro13ArtifactCreatureCard(card *game.Card) bool {
	return card.IsArtifact() && card.IsCreature()
}

// s82Bro13ArtifactOrCreatureCardLeftYourGraveyard is "one or more artifact
// and/or creature cards leave your graveyard during your turn" (Thran
// Vigil), asked of the card as it was in the graveyard (CR 603.10a).
func s82Bro13ArtifactOrCreatureCardLeftYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !IsYourTurn(g, source.Controller) {
		return false
	}
	c, ok := cardThatLeftYourGraveyard(ev, source, g)
	return ok && wasCardInAGraveyard(c, func(c game.Card) bool { return c.IsCreature() || c.IsArtifact() })
}

// s82Bro13HasUnearth reports whether a card, as it is now, has an unearth
// ability: an activated row whose label is the keyword's (CR 702.84a). A
// graveyard grant (Mishra, Tamer of Mak Fawa's) is a row of the graveyard
// only, so it is not read off a permanent.
func s82Bro13HasUnearth(c game.Card) bool {
	rows, _ := game.ActivatedAbilitiesWithOrigins(c)
	for _, r := range rows {
		if strings.HasPrefix(r.Label, "Unearth") {
			return true
		}
	}
	return false
}

// s82Bro13ExiledReturnToHand is "then return that card to its owner's
// hand" over the card that reached exile (Meticulous Excavation). An
// on-disk key: never renamed, never reused.
var s82Bro13ExiledReturnToHand = CardTail("then/s82-bro-b13-exiled-to-hand", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

// s82Bro13ReturnOrExileToHand is Meticulous Excavation's body: "Return
// target permanent you control to its owner's hand. If it has unearth,
// instead exile it, then return that card to its owner's hand." Whether it
// has unearth is read as the ability resolves, from the permanent as it is
// then (CR 608.2b); a target no longer legal does nothing.
func s82Bro13ReturnOrExileToHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(target)
	if !ok {
		return nil
	}
	if s82Bro13HasUnearth(c) {
		return ExileTarget{Target: target, ThenTail: ctx.ItemTail(s82Bro13ExiledReturnToHand)}.Apply(ctx)
	}
	return BounceToHand{Target: target}.Apply(ctx)
}

// s82Bro13CreatureEnteredFromYourGraveyard is "a creature enters from your
// graveyard" (Dredging Claw): any creature, under any control, whose ETB came
// out of the source controller's graveyard (EventETB.EnteredFrom).
func s82Bro13CreatureEnteredFromYourGraveyard(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !EnteredFromYourGraveyard(ev, source, lki, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}
