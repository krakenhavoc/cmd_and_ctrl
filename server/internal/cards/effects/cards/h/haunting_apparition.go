package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Haunting Apparition — Creature — Spirit {1}{U}{B}, 1+*/2 (Mirage, #5633):
//
//	"Flying
//	 As this creature enters, choose an opponent.
//	 Haunting Apparition's power is equal to 1 plus the number of green
//	 creature cards in the chosen player's graveyard."
//
// The opponent is named as it enters (CR 614.12) and stored on the permanent,
// as Nyxathid does. The power is a characteristic-defining ability (CR 604.3,
// layer 7a) reading the chosen player's graveyard live; the layer cache drops
// whenever a card enters or leaves any graveyard. Until the choice is made
// nobody's graveyard counts, so the power is the printed 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ced9130c-2865-49df-abdd-84069aa1b180",
		Name:            "Haunting Apparition",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AsEnters:        effects.ChoosePlayerAsEnters("Haunting Apparition", effects.Opponents),
		Static: []game.StaticAbility{{
			Label:     "Haunting Apparition's power is equal to 1 plus the number of green creature cards in the chosen player's graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = 1 + hauntingApparitionGreenCreatureCards(g, effects.ChosenPlayerOf(g, source.InstanceID), source.TextColor("G"))
			},
		}},
	})
}

// hauntingApparitionGreenCreatureCards counts the green creature cards in a
// player's graveyard; zero for nobody.
func hauntingApparitionGreenCreatureCards(g *game.Game, player uuid.UUID, green string) int {
	if player == uuid.Nil {
		return 0
	}
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() && c.HasColor(green) {
			n++
		}
	}
	return n
}
