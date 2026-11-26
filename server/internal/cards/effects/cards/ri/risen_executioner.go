package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Risen Executioner — Creature — Zombie Warrior {2}{B}{B}, 4/3 (Dragons of Tarkir):
//
//	"This creature can't block.
//	 Other Zombie creatures you control get +1/+1.
//	 You may cast this creature from your graveyard if you pay {1} more to cast
//	 it for each other creature card in your graveyard."
//
// The graveyard is a printed cast zone; the surcharge is the card's own cost
// increase, charged only when the cast comes from the graveyard and counting the
// OTHER creature cards in the caster's graveyard (the Executioner itself is
// excluded, CR 601.2f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "8587ec98-f35d-476c-ba67-0ef8bcd1f253",
		Name:          "Risen Executioner",
		Completeness:  effects.CompletenessFull,
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsMoreEach(risenExecutionerOtherCreatureCards,
				"Risen Executioner — {1} more for each other creature card in your graveyard", effects.CastFromGraveyard()),
		},
		Static: []game.StaticAbility{
			effects.RestrictSelf(game.CantBlock),
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Zombie"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}

func risenExecutionerOtherCreatureCards(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	p := q.Game.PlayerByIDForEffect(q.Controller)
	if p == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID != q.Card.InstanceID && c.IsCreature() {
			n++
		}
	}
	return n
}
