package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB397MayCastPick is "you may cast that card for as long as it remains
// exiled, and you may spend mana as though it were mana of any color to cast
// that spell", given to the chooser once the picked card has landed in exile
// (Covetous Urge's continuation, under the card's own key). The key is an
// on-disk identity: never renamed, never reused.
func S82OldB397MayCastPick(key, sourceName string) game.RevealedPickThen {
	return RevealedPickThen(key, func(ctx *Context, pick game.RevealedPick) error {
		g := ctx.Game
		for _, chosen := range pick.Chosen {
			if z := g.FindCardZoneForEffect(chosen.InstanceID); z == nil || z.Kind != game.ZoneExile {
				continue
			}
			c, ok := g.LookupCardForEffect(chosen.InstanceID)
			if !ok {
				continue
			}
			g.GrantCastPermissionToCardsForEffect(game.CastPermission{
				Player:     pick.Chooser,
				Zone:       game.ZoneExile,
				CastOnly:   true,
				AnyColor:   true,
				Duration:   game.WhileInZoneDuration(),
				Source:     pick.Source,
				SourceName: sourceName,
				Label:      "You may cast it for as long as it remains exiled; mana of any color can be spent to cast it",
			}, []game.Card{c})
		}
		return nil
	})
}

// S82OldB397RevealNonlandPick is "target opponent reveals their hand. You
// choose a nonland card from that player's graveyard or hand and exile it",
// continuing with then.
func S82OldB397RevealNonlandPick(ctx *Context, then game.RevealedPickThen) error {
	return ChooseFromRevealedHand{
		Player:          TargetedPlayer(ctx),
		Filter:          Nonland(),
		Label:           "nonland card",
		Exile:           true,
		FromGraveyard:   true,
		GraveyardFilter: Nonland(),
		Then:            then,
	}.Apply(ctx)
}

// S82OldB397AshiokExileTopThree is Ashiok, Nightmare Weaver's +2: the top three
// cards of the target opponent's library (the last elements of the zone),
// exiled together and linked to this Ashiok object.
func S82OldB397AshiokExileTopThree(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opp := TargetedPlayerOf(ctx)
	p := g.PlayerByIDForEffect(opp)
	if p == nil || p.Library == nil {
		return nil
	}
	ref, _ := ctx.SourceRef()
	var ids []uuid.UUID
	for i := len(p.Library.Cards) - 1; i >= 0 && len(ids) < 3; i-- {
		ids = append(ids, p.Library.Cards[i].InstanceID)
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileLinkedKeyedForEffect(ref, ids, game.KeyedTail{})
}

// S82OldB397AshiokReturn is Ashiok's −X as it resolves: a creature card with
// mana value X among the cards exiled with this Ashiok, picked by the
// controller.
func S82OldB397AshiokReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	x := ctx.X()
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsCreature() && c.ManaValue() == x {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   item.SourceCardID,
		Question: "Ashiok, Nightmare Weaver — put a creature card with mana value " + strconv.Itoa(x) + " exiled with Ashiok onto the battlefield",
		Cards:    ids,
		Min:      1,
		Max:      1,
		ThenKey:  s82OldB397AshiokPick,
		LiveZone: game.ZoneExile,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// s82OldB397AshiokPick is the keyed answer: the card goes onto the battlefield
// under the asker's control. An on-disk key: never renamed, never reused.
var s82OldB397AshiokPick = AskedPick("choose-cards/ashiok-nightmare-weaver-return", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ReturnFromExile{
		Target:     r.Picked[0],
		Controller: r.Chooser,
		ThenTail:   ctx.Tail(s82OldB397AshiokNightmare),
	}.Apply(ctx)
})

// s82OldB397AshiokNightmare makes the permanent that entered a Nightmare in
// addition to its other types, with no end. An on-disk key: never renamed,
// never reused.
var s82OldB397AshiokNightmare = EnteredTail("ashiok-nightmare-weaver/nightmare", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	return ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Nightmare")...)},
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Ashiok, Nightmare Weaver — that creature is a Nightmare in addition to its other types",
	}.Apply(ctx)
})

// S82OldB397AshiokExileEverything is Ashiok's −10: every card in every
// opponent's hand and graveyard, exiled together and linked to nothing.
func S82OldB397AshiokExileEverything(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, opp := range ctx.Opponents() {
		p := g.PlayerByIDForEffect(opp)
		if p == nil {
			continue
		}
		if p.Hand != nil {
			for _, c := range p.Hand.Cards {
				ids = append(ids, c.InstanceID)
			}
		}
		if p.Graveyard != nil {
			for _, c := range p.Graveyard.Cards {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileLinkedKeyedForEffect(game.ObjectRef{}, ids, game.KeyedTail{})
}
