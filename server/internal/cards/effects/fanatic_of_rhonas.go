package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fanatic of Rhonas — Creature — Snake Druid {1}{G}, 1/4:
//
//	"{T}: Add {G}.
//	 Ferocious — {T}: Add {G}{G}{G}{G}. Activate only if you control a
//	 creature with power 4 or greater.
//	 Eternalize {2}{G}{G} ({2}{G}{G}, Exile this card from your
//	 graveyard: Create a token that's a copy of it, except it's a 4/4
//	 black Zombie Snake Druid with no mana cost. Eternalize only as a
//	 sorcery.)"
//
// Two mana abilities and a graveyard payoff. The plain {T}: Add {G} and
// the Ferocious {T}: Add {G}{G}{G}{G} are the same ability slot in two
// clauses on the card, so they are two separate ManaAbility entries
// rather than one with a condition on the amount — activating either
// one taps the creature, so a controller who has met Ferocious picks
// which return they want, not both. Condition (S32) is "Activate only
// if ..." exactly: checked before any cost is validated, so failing it
// taps nothing (CR 602.5).
//
// Eternalize is the embalm.go body (CR 702.129a): the token is a 4/4
// black Zombie Snake Druid with no mana cost, keeping the Snake Druid
// subtypes ("in addition to its other types" — it already has them,
// so the clause is silent rather than duplicating). The eternalized
// token's mana abilities are the original's, carried by its oracle ID.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7973820b-fdaf-46ec-9e3e-d4c0e77b5067",
		Name:         "Fanatic of Rhonas",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{G}",
				Label:    "Add {G}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{G}{G}{G}{G}",
				Label:    "Ferocious — Add {G}{G}{G}{G}",
				Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
					return b24ControlsCreatureWithPowerAtLeast(g, controller, 4)
				},
			},
		},
		Activated: []ActivatedAbility{
			Eternalize("{2}{G}{G}"),
		},
	})
}
