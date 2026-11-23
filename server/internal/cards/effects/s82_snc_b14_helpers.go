package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82SncB14CostsLessMana is "This spell costs <unit> less to cast if <when>"
// where the unit carries coloured mana (Even the Score's {U}{U}{U}). It is
// CostsLess's reduction with a Unit, spent by reduceByManaCost (CR 118.7):
// each coloured symbol takes one mana of its colour off the cost, or one
// generic when the cost has none left. Panics at init on a string the cost
// parser rejects or that carries {X}.
func S82SncB14CostsLessMana(unit, label string, when ...CostPredicate) game.CostModifier {
	u, err := game.ParseCost(unit)
	if err != nil || u.XSlots > 0 || u.ManaValue() == 0 {
		panic(fmt.Sprintf("effects.s82SncB14CostsLessMana: %q is not a fixed mana amount (%v)", unit, err))
	}
	m := CostsLess(1, label, when...)
	m.Unit = &u
	return m
}

// s82SncB14AnOpponentDrewAtLeast is "an opponent has drawn <n> or more cards
// this turn": ONE opponent's own tally (not the table's total) reaches n. A
// player who has left the game is no opponent. The caster is q.Controller.
func S82SncB14AnOpponentDrewAtLeast(n int) CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil {
			return false
		}
		for _, p := range q.Game.Seats {
			if p == nil || p.Eliminated || p.ID == q.Controller {
				continue
			}
			if len(q.Game.DrawnThisTurn[p.ID]) >= n {
				return true
			}
		}
		return false
	}
}

// s82SncB14ArtifactsPutIntoGraveyardsThisTurn is "the number of artifacts
// that were put into graveyards from the battlefield this turn" (Structural
// Assault): every controller's, tokens included, counted off the turn's
// EventLTB records with each permanent's types as it last existed
// (CR 603.10a). Read after the destruction, so the artifacts the spell just
// destroyed are in the count along with any that left earlier this turn.
func S82SncB14ArtifactsPutIntoGraveyardsThisTurn(g *game.Game) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == uuid.Nil {
			continue
		}
		c, _ := g.LookupCardForEffect(ev.CardID)
		if LeftAsType(ev, c, "artifact") {
			n++
		}
	}
	return n
}
