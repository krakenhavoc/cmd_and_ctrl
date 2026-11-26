package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Curious Homunculus // Voracious Reader — {1}{U} Creature — Homunculus 1/1
// // Creature — Eldrazi Homunculus 3/4 (Eldritch Moon, S82 OLD, #5541):
//
//	Front: "{T}: Add {C}. Spend this mana only to cast an instant or sorcery
//	        spell.
//	        At the beginning of your upkeep, if there are three or more
//	        instant and/or sorcery cards in your graveyard, transform this
//	        creature."
//	Back:  "Prowess
//	        Instant and sorcery spells you cast cost {1} less to cast."
//
// The mana is Vodalian Arcanist's restricted mana. The upkeep trigger is an
// intervening "if" (CR 603.4): it is checked as the upkeep begins and again as
// the ability resolves, and the graveyard is the controller's own. The back
// face's discount is Baral's.
//
// No simplification.
func init() {
	const oracle = "c38db10f-c0b7-4391-b49a-77079a4d087c"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Curious Homunculus",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{C}",
			Label:        "Add {C}. Spend this mana only to cast an instant or sorcery spell",
			Restrictions: []string{effects.ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIfYou("there are three or more instant and/or sorcery cards in your graveyard",
				threeInstantsOrSorceriesInYourGraveyard,
				effects.AtYourUpkeep("Curious Homunculus — transform this creature", effects.Do(effects.TransformThis{}))),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Voracious Reader",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				effects.YourSpell(), effects.InstantOrSorcerySpell()),
		},
	})
}

// threeInstantsOrSorceriesInYourGraveyard is Curious Homunculus's condition.
func threeInstantsOrSorceriesInYourGraveyard(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	if p == nil || p.Graveyard == nil {
		return false
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsInstant() || c.IsSorcery() {
			n++
		}
	}
	return n >= 3
}
