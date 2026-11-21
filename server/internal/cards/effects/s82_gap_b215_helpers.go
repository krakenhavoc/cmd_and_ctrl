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

// GapB215LifeAtMostHalfStarting is "your life total is less than or equal to
// half your starting life total" (Bane, Lord of Darkness): the source's
// controller's life against the table's starting life, compared as life*2 <=
// starting so an odd total has no rounding.
func GapB215LifeAtMostHalfStarting(g *game.Game, source *game.Card) bool {
	if g == nil || source == nil {
		return false
	}
	for _, p := range g.Seats {
		if p != nil && p.ID == source.Controller {
			return p.Life*2 <= StartingLifeOf(g)
		}
	}
	return false
}

// GapB215SelfIndestructibleAtHalfLife is "As long as your life total is less
// than or equal to half your starting life total, <this> has indestructible":
// a layer 6 keyword grant (CR 613.1f) that re-reads on every life change.
func GapB215SelfIndestructibleAtHalfLife(name string) game.StaticAbility {
	ab := LifeGatedKeyword(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && GapB215LifeAtMostHalfStarting(g, source)
	}, "indestructible")
	ab.Label = "As long as your life total is less than or equal to half your starting life total, " + name + " has indestructible."
	return ab
}
