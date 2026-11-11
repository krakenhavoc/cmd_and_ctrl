package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arbor Adherent — Creature — Dog Druid {3}{G}, 2/4 (Tarkir: Dragonstorm
// Commander):
//
//	"{T}: Add one mana of any color.
//	 {T}: Add X mana of any one color, where X is the greatest toughness among
//	 other creatures you control."
//
// Two mana abilities, each its own row. The first is the ordinary one mana of
// any color. The second is one colour pick for the whole amount (Astral
// Cornucopia's ProducedOneColor), with X read off the battlefield when the
// ability is activated (CR 106.4, 605.3): the greatest toughness among the
// OTHER creatures its controller has, as they are then (negative or zero
// toughness adds nothing), so the Adherent's own 4 never counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9640225a-8f31-495f-b9de-53ed348ff24b",
		Name:         "Arbor Adherent",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
			},
			{
				Cost:           ManaAbilityCost{Tap: true},
				ProducedAmount: ProducedOneColor(arborAdherentX),
				Label:          "Add X mana of any one color, where X is the greatest toughness among other creatures you control",
			},
		},
	})
}

// arborAdherentX is the greatest toughness among the other creatures the
// ability's controller controls.
func arborAdherentX(g *game.Game, controller, source uuid.UUID) int {
	g.RecomputeLayersIfStaleLocked()
	best := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || c.InstanceID == source || !c.IsCreature() {
			continue
		}
		if t := c.CurrentToughness(); t > best {
			best = t
		}
	}
	return best
}
