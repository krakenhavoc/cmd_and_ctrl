package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaya, Spirits' Justice — Legendary Planeswalker — Kaya {2}{W}{B},
// loyalty 3 (Murders at Karlov Manor, #4469):
//
//	"Whenever one or more creatures you control and/or creature cards in
//	 your graveyard are put into exile, you may choose a creature card
//	 from among them. Until end of turn, target token you control becomes
//	 a copy of it, except it has flying.
//	 +2: Surveil 2, then exile a card from a graveyard.
//	 +1: Create a 1/1 white and black Spirit creature token with flying.
//	 −2: Exile target creature you control. For each other player, exile
//	 up to one target creature that player controls."
//
// The trigger is once per batch of exiles (CR 603.2c) off the zone-move
// events: a creature its controller controlled as it left the
// battlefield (last-known information, CR 603.10a) or a creature card
// that left its controller's graveyard. "Them" are the exiles of that
// batch, read off the turn's log as the trigger resolves; only those
// still in exile and creature cards there can be chosen (the ruling), so
// a token is never among them. The target token becomes a copy of the
// card's printed values (CR 707.2) with flying added, until end of turn.
//
// The −2 is one clause for your own creature and one "up to one" clause
// per other player in turn order (perOpponentClauses), all chosen by
// you; every still-legal target is exiled in one simultaneous event, so
// the trigger above sees your creature among the exiles.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        kayaSpiritsJusticeOracle,
		Name:            "Kaya, Spirits' Justice",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Triggered:       []game.TriggeredAbility{kayaSpiritsJusticeTrigger()},
		Activated: []ActivatedAbility{
			{
				Label: "+2: Surveil 2, then exile a card from a graveyard.",
				Cost:  LoyaltyCost(2),
				Effect: Do(Instruction(func(ctx *Context) error {
					return Surveil{Player: ctx.Controller(), N: 2, ThenTail: ctx.Tail(kayaSpiritsJusticeThenExile)}.Apply(ctx)
				})),
			},
			{
				Label:   "+1: Create a 1/1 white and black Spirit creature token with flying.",
				Cost:    LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: kayasGuileSpirit(), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−2: Exile target creature you control. For each other player, exile up to one target creature that player controls.",
				Cost:  LoyaltyCost(-2),
				Targets: Clauses(TargetCreature("target creature you control", YouControl()),
					perOpponentClauses(game.MaxPlayers-1, "creature", Creature())),
				Effect: kayaSpiritsJusticeExileTargets,
			},
		},
	})
}

const (
	kayaSpiritsJusticeOracle       = "0a094c5f-cefb-4ba1-90c8-dd78ae8efe95"
	kayaSpiritsJusticeTriggerLabel = "Kaya, Spirits' Justice — a token you control becomes a copy of an exiled creature card"
)

// kayaSpiritsJusticeTrigger is the first ability: once per batch of
// exiles, targeting a token you control.
func kayaSpiritsJusticeTrigger() game.TriggeredAbility {
	t := OncePerBatch(On(game.EventZoneMove, kayaSpiritsJusticeExiled, kayaSpiritsJusticeTriggerLabel, kayaSpiritsJusticeCopy))
	t.Targets = TargetPermanent("target token you control", IsTokenPredicate(), YouControl())
	return t
}

// kayaSpiritsJusticeExiled is the trigger condition: `ev` put a creature
// the source's controller controlled, or a creature card from their
// graveyard, into exile.
func kayaSpiritsJusticeExiled(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return kayaExiledYourCreature(ev, source.Controller, g)
}

// kayaExiledYourCreature reports whether `ev` is an exile of a creature
// `you` controlled as it left the battlefield (its last-known
// characteristics), or of a creature card from `you`'s graveyard.
func kayaExiledYourCreature(ev game.Event, you uuid.UUID, g *game.Game) bool {
	if ev.Kind != game.EventZoneMove || ev.NewZone != game.ZoneExile || you == uuid.Nil {
		return false
	}
	switch ev.OldZone {
	case game.ZoneBattlefield:
		info, ok := g.LastKnownPermanentForEffect(ev.CardID)
		return ok && info.Controller == you && slices.Contains(info.Characteristic.Types, "Creature")
	case game.ZoneGraveyard:
		c, ok := g.LookupCardForEffect(ev.CardID)
		return ok && c.Owner == you && c.IsCreature()
	}
	return false
}

// kayaSpiritsJusticeCopy offers the creature cards of the triggering
// batch's exiles that are still in exile; the pick's continuation makes
// the target token a copy.
func kayaSpiritsJusticeCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || item.Trigger == nil {
		return nil
	}
	candidates := kayaExiledCreatureCards(g, item.Controller, item.Trigger.Event.Batch)
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Kaya, Spirits' Justice — you may choose a creature card exiled this way; your token becomes a copy of it",
		Cards:    candidates,
		Min:      0,
		Max:      1,
		ThenKey:  kayaSpiritsJusticeCopyPick,
		Carry:    []uuid.UUID{targets[0].ID},
		LiveZone: game.ZoneExile,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// kayaExiledCreatureCards is "them": the cards the batch `batch` put into
// exile that the trigger counts, still in exile and creature cards there,
// once each, in the order they were exiled.
func kayaExiledCreatureCards(g *game.Game, you uuid.UUID, batch uint64) []uuid.UUID {
	events := g.EventsThisTurn()
	var out []uuid.UUID
	for i := len(events) - 1; i >= 0; i-- {
		ev := events[i]
		if ev.Batch < batch {
			break
		}
		if ev.Batch != batch || !kayaExiledYourCreature(ev, you, g) || slices.Contains(out, ev.CardID) {
			continue
		}
		if g.Exile == nil || !g.Exile.Contains(ev.CardID) {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && c.IsCreature() && !c.IsToken() {
			out = append(out, ev.CardID)
		}
	}
	slices.Reverse(out)
	return out
}

// kayaSpiritsJusticeCopyPick is the copy: the carried token becomes a
// copy of the chosen card until end of turn, except it has flying.
var kayaSpiritsJusticeCopyPick = game.RegisterChooseCardsThen("choose-cards/kaya-spirits-justice-copy", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 || r.Asked == nil {
		return nil
	}
	return BecomeCopy{
		Targets: []uuid.UUID{r.Carry[0]},
		Of:      r.Picked[0],
		Except:  func(v *game.PrintedValues) { v.AddKeyword("flying") },
		Label:   "Kaya, Spirits' Justice — a copy of the exiled card, with flying",
	}.Apply(NewContext(g, r.Asked))
})

// kayaSpiritsJusticeThenExile is the +2's "then exile a card from a
// graveyard": one card from any graveyard, if any graveyard has one.
var kayaSpiritsJusticeThenExile = CardTail("then/kaya-spirits-justice-exile-from-a-graveyard", func(ctx *Context, _ game.TailOutcome) error {
	g := ctx.Game
	var candidates []uuid.UUID
	for _, p := range g.Seats {
		if p != nil {
			candidates = append(candidates, graveyardCardsOfPlayer(g, p.ID)...)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Kaya, Spirits' Justice — exile a card from a graveyard",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		ThenKey:  kayaSpiritsJusticeExilePick,
	})
	return nil
})

// kayaSpiritsJusticeExilePick exiles the chosen graveyard card.
var kayaSpiritsJusticeExilePick = game.RegisterChooseCardsThen("choose-cards/kaya-spirits-justice-exile", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return g.ExileCardForEffect(r.Picked[0])
})

// kayaSpiritsJusticeExileTargets is the −2: every still-legal target,
// once each, exiled as one simultaneous event (CR 608.2b).
func kayaSpiritsJusticeExileTargets(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || slices.Contains(ids, t.ID) {
			continue
		}
		ids = append(ids, t.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, game.KeyedTail{})
}
