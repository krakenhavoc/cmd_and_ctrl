package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mobilized District — Land:
//
//	"{T}: Add {C}.
//	 {4}: This land becomes a 3/3 Citizen creature with vigilance until
//	 end of turn. It's still a land. This ability costs {1} less to
//	 activate for each legendary creature and planeswalker you control."
//
// A manland built like Restless Spire, with the ability's own cost
// clause (ActivatedAbility.CostModifiers, #1296): {1} less for each
// legendary permanent you control that is a creature or a planeswalker,
// counted as the cost is determined (CR 602.2b, 601.2f). A legendary
// creature planeswalker is one permanent and counts once. The engine's
// floor keeps the cost at {0} generic. The animation sets the base P/T,
// adds the Creature type and the Citizen subtype, and grants vigilance
// until end of turn; the land keeps its colour (colourless) and its mana
// ability.
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	Register(Spec{
		OracleID:     "eb2094cf-b4be-4f52-8615-e179ef7c741d",
		Name:         "Mobilized District",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}: This land becomes a 3/3 Citizen creature with vigilance until end of turn. It's still a land. This ability costs {1} less to activate for each legendary creature and planeswalker you control.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{4}"),
			CostModifiers: []game.CostModifier{
				CostsLessEach(PermanentsYouControl(And(Legendary(), Or(Creature(), Planeswalker()))),
					"This ability costs {1} less to activate for each legendary creature and planeswalker you control."),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				return ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(g.TextReaderForEffect(item).TextSubtypes("Citizen")...),
						game.SetBasePowerMod(3),
						game.SetBaseToughnessMod(3),
						game.AddKeywordsMod("vigilance"),
					},
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Mobilized District — a 3/3 Citizen creature with vigilance until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
