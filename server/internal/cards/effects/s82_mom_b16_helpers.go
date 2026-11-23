package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82Mom16TargetOpponentOrBattle is "target opponent or battle"
// (Ayara, Widow of the Realm): a player other than the caster, or a
// battle on the battlefield.
func S82Mom16TargetOpponentOrBattle() *game.TargetSpec {
	return &game.TargetSpec{
		Mode:    "any",
		Label:   "target opponent or battle",
		Players: true,
		PlayerOK: func(_ *game.Game, caster uuid.UUID, p *game.Player) bool {
			return p.ID != caster
		},
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsBattle()
		},
		Min: 1, Max: 1,
	}
}

// s82Mom16CardsOpponentsDrewThisTurn is the number of cards the
// modifier's controller's opponents have drawn this turn (Heliod, the
// Warped Eclipse). A player who has left the game is no opponent.
func S82Mom16CardsOpponentsDrewThisTurn(q game.CostQuery) int {
	n := 0
	for _, p := range q.Game.Seats {
		if p == nil || p.Eliminated || p.ID == q.Source.Controller {
			continue
		}
		n += len(q.Game.DrawnThisTurn[p.ID])
	}
	return n
}

// s82Mom16ReturnTargetToHand is "return target <card> from a graveyard to
// your hand": the first target still legal as the ability resolves
// (CR 608.2b).
func S82Mom16ReturnTargetToHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
}

// s82Mom16PutTargetOntoTheBattlefield is "put target <card> from a
// graveyard onto the battlefield under your control".
func S82Mom16PutTargetOntoTheBattlefield(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
}
