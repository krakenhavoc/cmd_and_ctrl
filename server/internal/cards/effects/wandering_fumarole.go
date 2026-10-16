package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wandering Fumarole — Land:
//
//	"This land enters tapped.
//	 {T}: Add {U} or {R}.
//	 {2}{U}{R}: Until end of turn, this land becomes a 1/4 blue and red
//	 Elemental creature with '{0}: Switch this creature's power and
//	 toughness until end of turn.' It's still a land."
//
// Raging Ravine's animation (one effect, pinned, at one timestamp; base
// P/T set in 7b) with an activated ability granted by the same effect
// (ADR 0093). The granted switch is layer 7d (CR 613.4d, #3996), after
// the 7b base, every pump and every counter, so a 1/4 with a +1/+1
// counter switches to 5/2. The granted ability belongs to the land, so
// "this creature" is the permanent it is granted to.
//
// No simplification.
const wanderingFumaroleSwitch = "wandering-fumarole/switch"

func init() {
	Register(Spec{
		OracleID:     "741c51f1-cfbe-4c29-ac8f-ca6bcd2652f9",
		Name:         "Wandering Fumarole",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U|R}",
			Label:    "Add {U} or {R}",
		}},
		Grants: []AbilityGrant{{
			Key: wanderingFumaroleSwitch,
			Activated: []ActivatedAbility{{
				Label:   "{0}: Switch this creature's power and toughness until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{0}"),
				Effect:  switchThisPT("Wandering Fumarole — switch its power and toughness"),
			}},
			Text: "{0}: Switch this creature's power and toughness until end of turn.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}{R}: Until end of turn, this land becomes a 1/4 blue and red Elemental creature with \"{0}: Switch this creature's power and toughness until end of turn.\" It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{2}{U}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return GrantAbilitiesFor{
					Target: item.SourceCardID,
					Keys:   []string{wanderingFumaroleSwitch},
					Also: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes("Elemental")...),
						game.SetColorsMod(g.TextReaderForEffect(item).TextColors("U", "R")...),
						game.SetBasePowerMod(1),
						game.SetBaseToughnessMod(4),
					},
					Label: "Wandering Fumarole — a 1/4 blue and red Elemental creature until end of turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
