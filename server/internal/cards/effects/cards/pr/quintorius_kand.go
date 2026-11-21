package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quintorius Kand — Legendary Planeswalker — Quintorius {3}{R}{W},
// loyalty 4:
//
//	"Whenever you cast a spell from exile, Quintorius Kand deals 2
//	 damage to each opponent and you gain 2 life.
//	 +1: Create a 3/2 red and white Spirit creature token.
//	 −3: Discover 4.
//	 −6: Exile any number of target cards from your graveyard. Add {R}
//	 for each card exiled this way. You may play those cards this turn."
//
// The static trigger reads the cast event's source zone, so a
// discovered card cast out of exile (ADR 0099) — the −3's own — sets it
// off. The −6 counts the cards that really reached exile (CR 400.7) for
// both the mana and the play permission, which is a per-object grant for
// the rest of the turn (ADR 0066) that also lets a land be played.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "91babc5f-ebeb-4894-8c35-b0f60249cb80",
		Name:            "Quintorius Kand",
		Completeness:    effects.CompletenessFull,
		Discovers:       true,
		StartingLoyalty: 4,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && ev.Actor == source.Controller && ev.OldZone == game.ZoneExile
			}, "Quintorius Kand — 2 damage to each opponent and you gain 2 life", effects.DamageEachOpponentThenGainLife(2)),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Create a 3/2 red and white Spirit creature token.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("3/2 red and white Spirit"), N: 1}),
			},
			{
				Label:   "−3: Discover 4.",
				Purpose: game.Purpose{Discover: 4},
				Cost:    effects.LoyaltyCost(-3),
				Effect:  effects.DiscoverN(4),
			},
			{
				Label:   "−6: Exile any number of target cards from your graveyard. Add {R} for each card exiled this way. You may play those cards this turn.",
				Cost:    effects.LoyaltyCost(-6),
				Targets: effects.TargetCardInGraveyard("any number of target cards from your graveyard", effects.YouOwn()).WithCount(1, 0),
				Effect:  quintoriusMinusSix,
			},
		},
	})
}

func quintoriusMinusSix(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, effects.NewContext(g, item).ItemTail(quintoriusExiled))
}

// quintoriusExiled adds the mana and grants the casts over what reached exile (#4253 slice 9).
var quintoriusExiled = effects.CardTail("quintorius-kand/mana-and-cast-from-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	g, exiled, controller, source := ctx.Game, o.IDs, ctx.Controller(), ctx.Source()

	if len(exiled) == 0 {
		return nil
	}
	if err := g.AddManaWithOptionsForEffect(controller, source, "", game.AddManaOptions{Amount: effects.ManaN(len(exiled), "R")}); err != nil {
		return err
	}
	var cards []game.Card
	for _, id := range exiled {
		if c, ok := g.LookupCardForEffect(id); ok {
			cards = append(cards, c)
		}
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player: controller,
		Zone:   game.ZoneExile,
		Source: source,
		Label:  "Quintorius Kand — you may play this card this turn",
	}, cards)
	return nil
})
