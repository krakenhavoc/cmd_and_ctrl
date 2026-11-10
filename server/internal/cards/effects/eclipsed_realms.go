package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eclipsed Realms — Land (Lorwyn Eclipsed, #4458):
//
//	"As this land enters, choose Elemental, Elf, Faerie, Giant, Goblin,
//	 Kithkin, Merfolk, or Treefolk.
//	 {T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a spell
//	 of the chosen type or activate an ability of a source of the chosen
//	 type."
//
// Secluded Courtyard without the word "creature": the choice is
// Dawn-Blessed Pennant's ChooseCreatureTypeFromAsEnters over the eight
// printed types, landing on Card.NamedTribe, and the spend restriction
// is one ManaRestrictAnyOf tag: a cast of a spell with the type, or an
// activation whose source has it (CR 106.6), in whatever zone the source
// is (the ruling: a Kithkin card in your hand or graveyard). A
// changeling spell or source has every creature type (CR 702.73a). A
// payment that is neither a cast nor an activation (an unlock, turning a
// permanent face up) cannot spend it. Until the type is named the
// coloured mana is unspendable, the weaker direction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5715ed43-395c-4877-99a7-8e28e7bf9dce",
		Name:         "Eclipsed Realms",
		Completeness: CompletenessFull,
		AsEnters: ChooseCreatureTypeFromAsEnters("Eclipsed Realms",
			"Elemental", "Elf", "Faerie", "Giant", "Goblin", "Kithkin", "Merfolk", "Treefolk"),
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:             ManaAbilityCost{Tap: true},
				Produced:         "{W|U|B|R|G}",
				Label:            "Add one mana of any color (spells and sources of the chosen type only)",
				RestrictionsFunc: eclipsedRealmsRestrictions,
			},
		},
	})
}

// eclipsedRealmsRestrictions is "spend this mana only to cast a spell of
// the chosen type or activate an ability of a source of the chosen type".
func eclipsedRealmsRestrictions(g *game.Game, _ uuid.UUID, source uuid.UUID) []string {
	tribe := game.ManaRestrictSubtype(g.NamedTribeOf(source))
	return []string{game.ManaRestrictAnyOf(
		[]string{game.ManaRestrictCast, tribe},
		[]string{game.ManaRestrictActivate, tribe},
	)}
}
