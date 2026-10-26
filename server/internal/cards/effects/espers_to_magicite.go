package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Espers to Magicite — Instant {3}{B} (deck request #3655):
//
//	"Exile each opponent's graveyard. When you do, choose up to one
//	 target creature card exiled this way. Create a token that's a copy
//	 of that card, except it's an artifact and it loses all other card
//	 types."
//
// The exile is immediate. "When you do" is a reflexive trigger (CR
// 603.12), raised only if at least one creature card was exiled, and it
// carries the exiled creature cards as its payload. The token copies
// the chosen card's copiable values (CR 707.2), then is only an
// artifact: no creature type, no power and toughness.
//
// The card is a TARGET of the reflexive trigger (CR 603.12, CR 115.1):
// chosen as the trigger goes on the stack, from the creature cards the
// spell exiled (the trigger's payload, TargetSpec.AmongPayload), and
// re-checked as it resolves (CR 608.2b), so the table can respond between
// the choice and the copy. Up to one: a trigger with nothing chosen does
// nothing.
func init() {
	Register(Spec{
		OracleID:     "06f7ec6d-d70e-44d8-bf43-ae667d244ab4",
		Name:         "Espers to Magicite",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			creature := map[uuid.UUID]bool{}
			for _, opp := range ctx.Opponents() {
				p := ctx.Game.PlayerByIDForEffect(opp)
				if p == nil || p.Graveyard == nil {
					continue
				}
				for _, c := range p.Graveyard.Cards {
					ids = append(ids, c.InstanceID)
					creature[c.InstanceID] = c.IsCreature()
				}
			}
			// The clause behind "exile" reads what really reached exile
			// (a replacement may have sent a card elsewhere), so it runs
			// from the exile's continuation.
			return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(espersToMagiciteExiled))
		},
	})
}

// espersCopy makes the artifact-only token copy of the card the
// reflexive trigger targeted, if it is still a legal target (CR 608.2b).
func espersCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return CreateTokenCopy{
			Controller: item.Controller,
			Copy:       t.ID,
			N:          1,
			Except: func(t *game.Card) {
				t.TypeLine = "Token Artifact"
				t.Power, t.Toughness = 0, 0
			},
		}.Apply(ctx)
	}
	return nil
}

// espersToMagiciteExiled raises the copy trigger over the creature cards that reached exile (#4253 slice 9).
var espersToMagiciteExiled = CardTail("espers-to-magicite/copy-a-creature-exiled", func(ctx *Context, o game.TailOutcome) error {
	g, item, exiled := ctx.Game, ctx.Item, o.IDs
	creature := map[uuid.UUID]bool{}
	for _, id := range exiled {
		if c, ok := g.LookupCardForEffect(id); ok {
			creature[id] = c.IsCreature()
		}
	}

	var creatures []uuid.UUID
	for _, id := range exiled {
		if creature[id] {
			creatures = append(creatures, id)
		}
	}
	// No creature card exiled: the reflexive trigger would
	// have nothing to choose, so it is not raised.
	if len(creatures) == 0 {
		return nil
	}
	return ReflexiveTrigger{
		Label: "Espers to Magicite — copy a creature card exiled this way as an artifact token",
		Cards: creatures,
		Body:  espersCopyBody,
	}.Apply(NewContext(g, item))
})
