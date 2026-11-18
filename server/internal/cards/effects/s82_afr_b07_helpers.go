package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b07_helpers.go — shared reads for Adventures in the Forgotten
// Realms batch 7 ("if a creature died this turn" and life gain).

// ACreatureDiedThisTurnSpell is the cost predicate "if a creature died this
// turn" (Purple Worm): any creature, any controller, tokens included, read
// as the spell is announced (CR 601.2f).
func ACreatureDiedThisTurnSpell() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Game != nil && q.Game.TurnTally.CreaturesDied > 0
	}
}

// s82AfrB07CreatureDiedCast is "cast this spell only if a creature died this
// turn" (Grim Wanderer's Tragic Backstory).
func s82AfrB07CreatureDiedCast(g *game.Game, _ uuid.UUID, _ game.Card) bool {
	return g.TurnTally.CreaturesDied > 0
}

// s82AfrB07NotNamedEbondeathDied is "a creature not named Ebondeath,
// Dracolich died this turn": the table-wide tally of creatures that died,
// less those that had that name as they died.
func s82AfrB07NotNamedEbondeathDied(g *game.Game, _ uuid.UUID, card game.Card) bool {
	return g.DiedNotNamedThisTurn(card.Name) > 0
}

// s82AfrB07SkeletonFilter is "each Skeleton you control".
var s82AfrB07SkeletonFilter = TribeFilter{Tribes: []string{"Skeleton"}, YoursOnly: true}

// s82AfrB07SkeletonPump is Skeletal Swarming's "gets +X/+0, where X is the
// number of other Skeletons you control", worked out for each Skeleton in
// turn (the one getting the bonus is not "other" to itself).
func s82AfrB07SkeletonPump() game.StaticAbility {
	f := s82AfrB07SkeletonFilter
	return game.StaticAbility{
		Label:     "Each Skeleton you control gets +X/+0, where X is the number of other Skeletons you control.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: f.Matches,
		Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
			n := 0
			for _, o := range g.BattlefieldCardsForEffect() {
				o := o
				if o.InstanceID != target.InstanceID && f.Matches(&o, g, source) {
					n++
				}
			}
			c.Power += n
		},
	}
}

// CondAnyCreatureDied is "if a creature died this turn": any creature, any
// controller, tokens included (Bulette). Unlike CondCreatureDied it is not
// limited to creatures that died under the controller's control.
type CondAnyCreatureDied struct{}

func (CondAnyCreatureDied) Holds(ctx *Context) bool {
	return ctx.Game.TurnTally.CreaturesDied > 0
}

func (CondAnyCreatureDied) condition() (game.FootprintStep, bool) {
	return condStep(game.FootprintCondAnyCreatureDied, 0), true
}

// s82AfrB07AnyCreatureDiedWhen is the trigger-time half of the same "if": a
// creature of any controller died this turn.
func s82AfrB07AnyCreatureDiedWhen(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.TurnTally.CreaturesDied > 0
}
