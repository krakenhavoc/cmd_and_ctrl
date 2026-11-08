package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdm_r_helpers.go — shared bodies for Tarkir: Dragonstorm cards from R
// (#4463), each also used by the earlier card that first had it.

// tdmCastAnotherSpellThisTurn is "if you've cast another spell this turn"
// as a self cost reduction's condition (Focus the Mind, Rally the
// Monastery). It reads the caster's cast tally as the cost is determined
// (CR 601.2f). The spell being cast is not in the tally yet, because a
// spell becomes cast only at CR 601.2i, so any spell there is another one.
func tdmCastAnotherSpellThisTurn(q game.CostQuery) bool {
	return q.Game != nil && q.Game.CastTallyFor(q.Controller).Total > 0
}

// tdmGreenEqualToPower is "{T}: Add an amount of {G} equal to this
// creature's power" as a mana ability's amount (Topiary Lecturer, Rainveil
// Rejuvenator): the source's power as the ability is activated, layers
// brought current first.
func tdmGreenEqualToPower(g *game.Game, _ uuid.UUID, source uuid.UUID) game.ManaAmount {
	return game.ManaAmount{Options: []string{"G"}, N: b31CurrentPowerOf(g, source)}
}

// tdmMayDiscardThenDrawOne is "you may discard a card. If you do, draw a
// card." as a trigger body: the discard is chosen as it resolves, up to
// one card, and the draw happens only if a card was discarded.
func tdmMayDiscardThenDrawOne(question string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return b39MayDiscardThenDraw(1, false, question, drawOneIfDiscarded)(NewContext(g, item))
	}
}

// tdmCreaturesYouControl is every creature the player controls as the
// effect resolves (CR 611.2c fixes the set then).
func tdmCreaturesYouControl(g *game.Game, player uuid.UUID) []uuid.UUID {
	var mine []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			mine = append(mine, c.InstanceID)
		}
	}
	return mine
}
