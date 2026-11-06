package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Release to Memory — Instant {3}{W} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Exile target opponent's graveyard. For each creature card exiled this
//	 way, create a 1/1 colorless Spirit creature token."
//
// The whole graveyard is exiled as one event, and the keyed tail is told
// the cards that actually reached exile (CR 400.7): a card a replacement
// sent elsewhere was not "exiled this way". Each of those that is a
// creature card makes one plain 1/1 colorless Spirit (the printed token
// has no abilities, unlike the set's own Spirit), all created at once
// under the caster's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ae456e4-cb35-47be-b654-051e288622fe",
		Name:         "Release to Memory",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			opp, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			p := ctx.Game.PlayerByIDForEffect(opp)
			if p == nil {
				return nil
			}
			ids := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
			for _, c := range p.Graveyard.Cards {
				ids = append(ids, c.InstanceID)
			}
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(releaseToMemorySpirits))
		},
	})
}

// releaseToMemorySpirits is "for each creature card exiled this way,
// create a 1/1 colorless Spirit creature token", told the cards that
// reached exile. An on-disk key.
var releaseToMemorySpirits = CardTail("release-to-memory/spirit-per-creature-card", func(ctx *Context, o game.TailOutcome) error {
	n := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 colorless Spirit"), N: n}.Apply(ctx)
})
