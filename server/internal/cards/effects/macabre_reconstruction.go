package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Macabre Reconstruction — Sorcery {3}{B} (Murders at Karlov Manor,
// #4469):
//
//	"This spell costs {2} less to cast if a creature card was put into
//	 your graveyard from anywhere this turn.
//	 Return up to two target creature cards from your graveyard to your
//	 hand."
//
// The discount is a self cost modifier read off this turn's events, the
// scan descend's "you descended this turn" makes (graveyard_provenance.go)
// with "creature card" for "permanent card": a card you own arriving in
// your graveyard from any other zone. A token is not a card, so a creature
// token dying does not count. The return is Does Machines' "up to two
// target … cards from your graveyard".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04633ecd-eac0-426a-83a5-8e88262fca30",
		Name:         "Macabre Reconstruction",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("up to two target creature cards from your graveyard", Creature(), YouOwn()).WithCount(0, 2),
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if a creature card was put into your graveyard from anywhere this turn.",
				func(q game.CostQuery) bool {
					return q.Game != nil && creatureCardPutIntoYourGraveyardThisTurn(q.Game, q.Controller)
				}),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
	})
}

// creatureCardPutIntoYourGraveyardThisTurn is "a creature card was put
// into your graveyard from anywhere this turn": an arrival this turn, from
// another zone, of a nontoken creature card `player` owns. The card's types
// are read as it is now, as descendedThisTurn reads them.
//
// Caller holds g.mu.
func creatureCardPutIntoYourGraveyardThisTurn(g *game.Game, player uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.NewZone != game.ZoneGraveyard || ev.OldZone == game.ZoneGraveyard || ev.CardID == uuid.Nil {
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if ok && c.Owner == player && c.IsCreature() && !c.IsToken() {
			return true
		}
	}
	return false
}
