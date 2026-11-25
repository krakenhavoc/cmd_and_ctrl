package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b137_helpers.go — shared bodies for the Theros Beyond Death cards of
// S82 set OLD, batch 137 (#5514).

// OldB137PurphorosEntered is the end of Purphoros, Bronze-Blooded's ability:
// "Sacrifice it at the beginning of the next end step", scheduled for the
// card the put-onto-the-battlefield brought in. A declined or empty put makes
// no delayed trigger. An on-disk key: never renamed, never reused.
var OldB137PurphorosEntered = PutFromHandTail("purphoros-bronze-blooded/entered", func(ctx *Context, res PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	c, ok := BattlefieldCardForEffect(ctx.Game, res.Entered)
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label:  "Purphoros, Bronze-Blooded — sacrifice that creature",
		Body:   Cmdr2S12SoulstokeSacrifice,
		Params: game.EffectParams{Object: game.ObjectRef{ID: res.Entered, Epoch: c.ObjectEpoch}},
	}.Apply(ctx)
})

// OldB137NyleaRevealTop is Nylea, Keen-Eyed's second ability as it resolves:
// "Reveal the top card of your library. If it's a creature card, put it into
// your hand. Otherwise, you may put it into your graveyard."
func OldB137NyleaRevealTop(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	revealed := g.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), 1, "Nylea, Keen-Eyed — reveal the top card of your library")
	if len(revealed) == 0 {
		return nil
	}
	top, ok := g.LookupCardForEffect(revealed[0])
	if !ok {
		return nil
	}
	if top.IsCreature() {
		return g.BounceToHandForEffect(top.InstanceID)
	}
	return MayChoice{
		Player:   item.Controller,
		Question: "Nylea, Keen-Eyed — put " + top.Name + " into your graveyard?",
		YesLabel: "Put it into your graveyard",
		NoLabel:  "Leave it on top",
		Then:     oldB137NyleaToGraveyard,
		Carry:    []uuid.UUID{top.InstanceID},
	}.Apply(ctx)
}

// oldB137NyleaToGraveyard is the "yes": the revealed card goes to the
// graveyard if it is still in the library. Keyed: never renamed, never reused.
var oldB137NyleaToGraveyard = MayYes("confirm/nylea-keen-eyed-to-graveyard", PutCarriedLibraryCardIntoGraveyard)

// OldB137ExileFromHandOfOwner is "then that player exiles a card from their
// hand": `player` picks one card from their own hand and it is exiled face up
// (Mindleech Ghoul's exile, for one named player). An empty hand exiles
// nothing.
func OldB137ExileFromHandOfOwner(g *game.Game, source, player uuid.UUID, label string) {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil || p.Hand.Size() == 0 {
		return
	}
	cards := make([]uuid.UUID, 0, p.Hand.Size())
	for _, c := range p.Hand.Cards {
		cards = append(cards, c.InstanceID)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     source,
		Question:   label,
		Cards:      cards,
		Min:        1,
		Max:        1,
		ThenKey:    stdB40ExileFromHand,
		LiveZone:   game.ZoneHand,
	})
}

// OldB137AshiokNightmareToken is the 2/3 blue and black Nightmare of Ashiok,
// Nightmare Muse's +1, for CreateToken.
func OldB137AshiokNightmareToken() game.Card { return TokenFromCatalog(printedOldB137NightmareToken) }

// printedOldB137NightmareToken is that Nightmare as PRINTED: "Whenever this
// creature attacks or blocks, each opponent exiles the top two cards of their
// library." "Each opponent" is the token controller's. The slug is an on-disk
// identity: never renamed, never reused.
func printedOldB137NightmareToken() TokenTemplate {
	return TokenTemplate{
		Slug: "nightmare-attacks-or-blocks-exile-two",
		Card: game.Card{
			Name:      "Nightmare",
			TypeLine:  "Token Creature — Nightmare",
			Power:     2,
			Toughness: 3,
			Colors:    []string{"B", "U"},
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return AttackDeclared(ev, source) || SelfBlocksOnce(ev, source)
				},
				"Nightmare — each opponent exiles the top two cards of their library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, opp := range ctx.Opponents() {
						if err := (MillToZone{Player: opp, N: 2, To: game.ZoneExile}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
					return nil
				}),
		},
		Text: "Whenever this creature attacks or blocks, each opponent exiles the top two cards of their library.",
	}
}

// OldB137AshiokBounce is Ashiok, Nightmare Muse's −3 as it resolves: "Return
// target nonland permanent to its owner's hand, then that player exiles a card
// from their hand." The owner is read before the move; the exile is asked once
// the return has settled (a commander's command-zone question may pause it).
func OldB137AshiokBounce(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, found := g.LookupCardForEffect(id)
	if !found {
		return nil
	}
	return BounceToHand{Target: id, ThenTail: WithCarry(ctx.ItemTail(oldB137AshiokExile), c.Owner)}.Apply(ctx)
}

// oldB137AshiokExile is "then that player exiles a card from their hand",
// the owner riding the tail as its Carry. An on-disk key: never renamed,
// never reused.
var oldB137AshiokExile = CardTail("then/ashiok-nightmare-muse-exile", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	OldB137ExileFromHandOfOwner(ctx.Game, ctx.Source(), o.Tail.Carry[0], "Ashiok, Nightmare Muse — exile a card from your hand")
	return nil
})

// OldB137AshiokCastFromExile is Ashiok, Nightmare Muse's −7: "You may cast up
// to three spells from among face-up cards your opponents own from exile
// without paying their mana costs." One free, cast-only permission over every
// nonland face-up card an opponent owns in exile, with three casts between
// them (Doom Reigns Supreme's grant), offered during the resolution
// (CR 608.2g); declining withdraws it, so nothing can be cast afterwards.
func OldB137AshiokCastFromExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var cards []game.Card
	var ids []uuid.UUID
	if g.Exile != nil {
		for _, c := range g.Exile.Cards {
			if c.FaceDown || c.IsLand() || c.IsToken() || c.Owner == item.Controller {
				continue
			}
			if p := g.PlayerByIDForEffect(c.Owner); p == nil || p.Eliminated {
				continue
			}
			cards = append(cards, c)
			ids = append(ids, c.InstanceID)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     ctx.Controller(),
		Zone:       game.ZoneExile,
		Duration:   g.UntilEndOfTurnDuration(),
		Cost:       "{0}",
		Timing:     game.TimingFlash,
		CastOnly:   true,
		CastsLeft:  3,
		Source:     ctx.Source(),
		SourceName: "Ashiok, Nightmare Muse",
		Label:      "Ashiok, Nightmare Muse — cast it without paying its mana cost",
	}, cards)
	OfferResolutionCastOver(g, ctx.Controller(), ctx.Source(), "Ashiok, Nightmare Muse",
		game.MayCastKeywordFreeCast, DeclineLeaveInExile, ids...)
	return nil
}
