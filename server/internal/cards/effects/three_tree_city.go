package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Three Tree City — Legendary Land:
//
//	"As Three Tree City enters, choose a creature type.
//	 {T}: Add {C}.
//	 {2}, {T}: Choose a color. Add an amount of mana of that color
//	 equal to the number of creatures you control of the chosen
//	 type."
//
// Nykthos for a tribe: the same {2}, {T} colour pick, with the amount
// read off a creature count instead of devotion. Both of the pieces
// it needs are catalogue vocabulary — `ChooseCreatureTypeAsEnters` is
// the CR 614.12 prompt whose answer lands on Card.NamedTribe, and
// `ProducedOneColor` is "N mana of any one color" with N computed at
// activation.
//
// "N mana of any ONE color" is one pick minting N tokens, never N
// separate any-colour slots — the difference between a Three Tree
// City with four Elves paying {G}{G}{G}{G} and it paying {W}{U}{B}{R}.
// That is the whole reason the produced-mana grammar carries a
// per-colour amount.
//
// The count reads EFFECTIVE subtypes, so a changeling counts as the
// named type (CR 702.73a) and a creature some layer-4 effect made an
// Elf counts too. Until the controller answers the entry prompt the
// named type is empty, the count is zero and the ability adds nothing
// — never "everything", which is the dangerous direction.
//
// Two abilities, not one with a mode, for Cavern of Souls' reason:
// the colourless half is free and unrestricted so the auto-tapper can
// plan around it, and the second costs {2} and is a deliberate click.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da3b17a2-e1e1-44e9-b9b1-ae54a92037db",
		Name:         "Three Tree City",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Three Tree City"),
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:         ManaAbilityCost{Tap: true, Mana: "{2}"},
				ProducedFunc: ProducedOneColor(creaturesYouControlOfTheChosenType),
				Label:        "Choose a color: add that much mana of it for your creatures of the chosen type",
			},
		},
	})
}

// creaturesYouControlOfTheChosenType counts the creatures `controller`
// controls whose effective subtypes carry the type named on `source`.
// No type named yet is a count of zero.
func creaturesYouControlOfTheChosenType(g *game.Game, controller, source uuid.UUID) int {
	tribe := g.NamedTribeOf(source)
	if tribe == "" {
		return 0
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsCreature() && c.HasSubtype(tribe) {
			n++
		}
	}
	return n
}
