package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_b1_helpers.go — shared bodies for Marvel Super Heroes, half B, first
// batch (#4455). Every name carries the mshB1 prefix because the other
// Marvel batches land beside this one.

// mshB1YouAttackAlone is the When of "whenever a creature you control
// attacks alone" (Agent 13, Sharon Carter): one of the source controller's
// creatures was declared as an attacker, and it is the only creature
// declared in the combat (CR 506.5). The engine stamps every attacker of a
// declaration before it announces any attack event, so a lone declaration
// reads as alone when the trigger is harvested.
func mshB1YouAttackAlone(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return attackDeclaredByYou(ev, source.Controller) && game.AttackedAlone(g)
}

// mshB1CounterOnEachOtherYouControl is "put a +1/+1 counter on each other
// <match> you control" (Agent Phil Coulson). The set is snapshotted before
// any counter lands, so a creature a counter trigger makes mid-loop gets
// none, and each recipient is its own placement.
func mshB1CounterOnEachOtherYouControl(match CardPredicate) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		var ids []uuid.UUID
		for _, c := range MatchingBattlefield(ctx, And(match, YouControl(), NotSelf(item.SourceCardID))) {
			ids = append(ids, c.InstanceID)
		}
		for _, id := range ids {
			if !onBattlefield(g, id) {
				continue
			}
			if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
				return err
			}
		}
		return nil
	}
}
