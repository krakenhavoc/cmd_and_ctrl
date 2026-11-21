package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b215_helpers.go — shared shapes for GAP batch 215 (#5948).

// GapB215GreatestPowerAmongCreaturesAndGraveyardCards is "the greatest power
// among creatures you control and creature cards in your graveyard"
// (Ambitious Dragonborn): the layered power of each creature the player
// controls and the printed power of each creature card in their graveyard
// (a card there has no layers; a */* card reads 0). Zero with none. Caller
// must hold g.mu.
func GapB215GreatestPowerAmongCreaturesAndGraveyardCards(g *game.Game, player uuid.UUID) int {
	best := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			if p := c.CurrentPower(); p > best {
				best = p
			}
		}
	}
	if p := g.PlayerByIDForEffect(player); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() && c.Power > best {
				best = c.Power
			}
		}
	}
	return best
}

// GapB215WardXExperience is "ward {X}, where X is the number of experience
// counters you have" (Minthara, Merciless Soul), as the ward's cost.
func GapB215WardXExperience() WardCost { return WardCost{ManaFromExperience: true} }
