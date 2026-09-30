package effects

import (
	"strings"

	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "91babc5f-ebeb-4894-8c35-b0f60249cb80",
		Name:            "Quintorius Kand",
		Completeness:    CompletenessFull,
		Discovers:       true,
		StartingLoyalty: 4,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source != nil && ev.Actor == source.Controller && ev.OldZone == game.ZoneExile
			}, "Quintorius Kand — 2 damage to each opponent and you gain 2 life", damageEachOpponentThenGainLife(2)),
		},
		Activated: []ActivatedAbility{
			{
				Label:  "+1: Create a 3/2 red and white Spirit creature token.",
				Cost:   LoyaltyCost(1),
				Effect: Do(CreateToken{Template: TokenCard("3/2 red and white Spirit"), N: 1}),
			},
			{
				Label:  "−3: Discover 4.",
				Cost:   LoyaltyCost(-3),
				Effect: DiscoverN(4),
			},
			{
				Label:   "−6: Exile any number of target cards from your graveyard. Add {R} for each card exiled this way. You may play those cards this turn.",
				Cost:    LoyaltyCost(-6),
				Targets: TargetCardInGraveyard("any number of target cards from your graveyard", YouOwn()).WithCount(1, 0),
				Effect:  quintoriusMinusSix,
			},
		},
	})
}

func quintoriusMinusSix(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	controller, source := item.Controller, item.SourceCardID
	return g.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
		if len(exiled) == 0 {
			return nil
		}
		if err := g.AddManaForEffect(controller, source, strings.Repeat("{R}", len(exiled))); err != nil {
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
}
