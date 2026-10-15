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
// DECLARED SIMPLIFICATION (choice): the card is picked as the reflexive
// trigger resolves, from the creature cards still exiled by this spell,
// rather than being targeted as it goes on the stack. The difference is
// only that the table cannot respond between the choice and the copy.
func init() {
	Register(Spec{
		OracleID:     "06f7ec6d-d70e-44d8-bf43-ae667d244ab4",
		Name:         "Espers to Magicite",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The creature card is chosen as the reflexive trigger resolves, not targeted as it goes on the stack."},
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
			return ctx.Game.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
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
		},
	})
}

// espersCopy picks up to one of the payload's creature cards still in
// exile and makes the artifact-only token copy of it.
func espersCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var candidates []uuid.UUID
	for _, id := range ctx.PayloadCards() {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     item.SourceCardID,
		Question:   "Espers to Magicite — choose up to one creature card exiled this way to copy",
		Cards:      candidates,
		Min:        0,
		Max:        1,
		Zone:       game.ZoneExile,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return CreateTokenCopy{
				Controller: item.Controller,
				Copy:       picked[0],
				N:          1,
				Except: func(t *game.Card) {
					t.TypeLine = "Token Artifact"
					t.Power, t.Toughness = 0, 0
				},
			}.Apply(NewContext(g, item))
		},
	})
	return nil
}
