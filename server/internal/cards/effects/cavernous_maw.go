package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cavernous Maw — Land — Cave (The Lost Caverns of Ixalan, #4470):
//
//	"{T}: Add {C}.
//	 {2}: This land becomes a 3/3 Elemental creature until end of turn.
//	 It's still a Cave land. Activate only if the number of other Caves
//	 you control plus the number of Cave cards in your graveyard is three
//	 or greater."
//
// A manland built like Restless Spire, without a colour or a granted
// ability: the animation adds the creature type and the Elemental
// subtype and sets its base power and toughness until end of turn, and
// the land keeps its land type and its Cave subtype. It stays colourless.
// "Elemental" is a word of the land's text (CR 612.2), read through it.
//
// The activation condition is caveCount leaving the Maw itself out
// ("other Caves"), checked as the ability is activated (CR 602.5b) and
// not again as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "952ab8fe-f7d3-4673-89de-8c6d3f8a081f",
		Name:          "Cavernous Maw",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{painlessColorless()},
		Activated: []ActivatedAbility{{
			Label:   "{2}: This land becomes a 3/3 Elemental creature until end of turn. It's still a Cave land. Activate only if the number of other Caves you control plus the number of Cave cards in your graveyard is three or greater.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{2}"),
			Condition: func(g *game.Game, controller, source uuid.UUID) bool {
				return caveCount(g, controller, source) >= 3
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes("Elemental")...),
						game.SetBasePowerMod(3),
						game.SetBaseToughnessMod(3),
					},
					Duration: DurationUntilEndOfTurn(NewContext(g, item)),
					Label:    "Cavernous Maw — a 3/3 Elemental creature until end of turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
