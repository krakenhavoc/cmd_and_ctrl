package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boneyard Mycodrax — Creature — Fungus {2}{B}, */* (Commander 2020, #5513):
//
//	"Boneyard Mycodrax's power and toughness are each equal to the number of
//	 other creature cards in your graveyard.
//	 Scavenge {4}{B} ({4}{B}, Exile this card from your graveyard: Put a
//	 number of +1/+1 counters equal to this card's power on target creature.
//	 Scavenge only as a sorcery.)"
//
// The characteristic-defining ability (CR 604.3, layer 7a) works in every
// zone: on the battlefield it counts the creature cards in its owner's
// graveyard, and in the graveyard it leaves itself out ("other"). Scavenge is
// the keyword (scavenge.go): its cost exiles the card, and "this card's
// power" is read off the card in exile as the ability resolves, where the
// same ability counts the creature cards still in the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f852adf6-3e54-4a1d-92e1-7dbc6d2228ce",
		Name:         "Boneyard Mycodrax",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Boneyard Mycodrax's power and toughness are each equal to the number of other creature cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := c20OtherCreatureCardsInYourGraveyard(g, source)
				c.Power, c.Toughness = n, n
			},
		}},
		Activated: []ActivatedAbility{Scavenge("{4}{B}")},
	})
}

// c20OtherCreatureCardsInYourGraveyard counts the creature cards in the
// graveyard of the source's controller, leaving out the source itself: no
// difference on the battlefield or in exile, and "other" while the card
// sits in that graveyard.
func c20OtherCreatureCardsInYourGraveyard(g *game.Game, source *game.Card) int {
	n := b11CreatureCardsInGraveyard(g, source.Controller)
	p := g.PlayerByIDForEffect(source.Controller)
	if p != nil && p.Graveyard != nil && p.Graveyard.Contains(source.InstanceID) {
		n--
	}
	return max(n, 0)
}
