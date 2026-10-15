package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Born to Drive — {2}{W} Enchantment — Aura:
//
//	"Enchant artifact or creature
//	 As long as enchanted permanent is a creature, it gets +1/+1 for each
//	 creature and/or Vehicle you control.
//	 Channel — {2}{W}, Discard this card: Create two 1/1 colorless Pilot
//	 creature tokens with "This token crews Vehicles as though its power
//	 were 2 greater.""
//
// The bonus is one layer 7c modifier recounted on every layer pass, behind
// the condition "enchanted permanent is a creature" (an uncrewed Vehicle
// gets nothing until it becomes one). "Creature and/or Vehicle" counts each
// permanent once, so a crewed Vehicle is one, and "you control" is the
// Aura's controller, not the host's. The count reads the live battlefield
// (the Coat of Arms rule), so a creature that enters in the same pass is
// counted.
//
// Channel is an activated ability that works from the hand (Otawara's
// shape): the discard is its cost, and the two tokens are CrewPilotToken
// (ADR 0186 section 5), crew-only.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ddd84d2-1a6d-43f5-a42d-3c66eee5f7e5",
		Name:         "Born to Drive",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant artifact or creature", Or(Artifact(), Creature())),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{{
			Label:    "As long as enchanted permanent is a creature, it gets +1/+1 for each creature and/or Vehicle you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return AttachedToSource(target, g, source) && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := creaturesAndVehiclesControlledBy(g, source.Controller)
				c.Power += n
				c.Toughness += n
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "Channel — {2}{W}, Discard this card: Create two 1/1 colorless Pilot creature tokens with \"This token crews Vehicles as though its power were 2 greater.\"",
			Cost:    game.AbilityCost{Mana: "{2}{W}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Template: CrewPilotToken(), N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}

// creaturesAndVehiclesControlledBy counts the permanents `player`
// controls that are creatures and/or Vehicles, each once. It walks
// g.Battlefield.Cards directly because it runs inside a layer recompute
// pass, where the caller may hold only the read lock (see
// sharedCreatureTypeCount).
func creaturesAndVehiclesControlledBy(g *game.Game, player uuid.UUID) int {
	if g == nil || g.Battlefield == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == player && (c.IsCreature() || c.HasSubtype("Vehicle")) {
			n++
		}
	}
	return n
}
