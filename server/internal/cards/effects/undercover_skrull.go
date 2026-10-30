package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Undercover Skrull — {1}{G} Creature — Skrull Shapeshifter Villain 1/1
// (Marvel Super Heroes, #4455):
//
//	"As long as there are two or more creature cards in your graveyard,
//	 this creature gets +2/+2 and is all creature types.
//	 {T}: Add one mana of any color."
//
// Killmonger's graveyard gate on two statics, both read live on every layer
// pass: the +2/+2 is a layer 7c modification and "all creature types" is
// the layer 4 fact on the Characteristic (Maskwood Nexus' applyAllCreatureTypes),
// so the two come and go together as the graveyard fills and empties.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33b1379f-b329-42cd-92ed-4811d44ed8dd",
		Name:         "Undercover Skrull",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "As long as there are two or more creature cards in your graveyard, this creature gets +2/+2.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: skrullTwoCreatureCardsInYourGraveyard,
				Apply:     pumpBy(2, 2),
			},
			{
				Label:     "As long as there are two or more creature cards in your graveyard, this creature is all creature types.",
				Layer:     game.Layer4Type,
				AppliesTo: skrullTwoCreatureCardsInYourGraveyard,
				Apply:     applyAllCreatureTypes,
			},
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}

// skrullTwoCreatureCardsInYourGraveyard is the gate: the source itself,
// while its controller's graveyard holds two or more creature cards.
func skrullTwoCreatureCardsInYourGraveyard(target *game.Card, g *game.Game, source *game.Card) bool {
	return target.InstanceID == source.InstanceID &&
		GraveyardAtLeast(2, func(c game.Card) bool { return c.IsCreature() })(g, source.Controller, source.InstanceID)
}
