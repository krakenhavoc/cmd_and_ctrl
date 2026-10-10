package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blitzball — Artifact {3}:
//
//	"{T}: Add one mana of any color.
//	 GOOOOAAAALLL! — {T}, Sacrifice this artifact: Draw two cards.
//	 Activate only if an opponent was dealt combat damage by a
//	 legendary creature this turn."
//
// The mana ability is the ordinary five-colour pipe. The goal ability's
// condition reads the turn tally's record of which creature OBJECTS
// dealt combat damage to each opponent (`CreaturesThatDealtCombatDamageTo
// ThisTurn`, #2149) and looks each one up with its last-known
// information, so a legend that traded in combat after connecting still
// counts, and one that was flickered since is a different object and
// does not. "Legendary" is read off the characteristic the creature had
// when last seen. Sacrificing Blitzball is a cost, so the draw happens
// whatever happens to it next.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "75719f87-ce1a-40d6-ac5a-5ebf0c936c84",
		Name:         "Blitzball",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:     "GOOOOAAAALLL! — {T}, Sacrifice this artifact: Draw two cards. Activate only if an opponent was dealt combat damage by a legendary creature this turn.",
			Cost:      Plus(TapCost(), SacrificeThis()),
			Purpose:   game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Condition: opponentDealtCombatDamageByALegend,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}

// opponentDealtCombatDamageByALegend is Blitzball's activation gate.
func opponentDealtCombatDamageByALegend(g *game.Game, controller, _ uuid.UUID) bool {
	for _, seat := range g.Seats {
		if seat == nil || seat.Eliminated || seat.ID == controller {
			continue
		}
		for _, ref := range g.CreaturesThatDealtCombatDamageToThisTurn(seat.ID) {
			info, ok := g.PermanentForEffect(ref)
			if !ok {
				continue
			}
			for _, s := range info.Characteristic.Supertypes {
				if s == "Legendary" {
					return true
				}
			}
		}
	}
	return false
}
