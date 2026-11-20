package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rupture Spire — Land (Conflux, #5587):
//
//	"This land enters tapped.
//	 When this land enters, sacrifice it unless you pay {1}.
//	 {T}: Add one mana of any color."
//
// Command Bridge's land with a mana payment in place of the tap: the
// enters-tapped replacement (CR 614.1c), then a trigger that asks for
// {1} and sacrifices the land on a decline, through the shared
// upkeepSacrificeSource branch. A Spire that left in response, or left
// and came back (CR 400.7), asks nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7eadffcb-1e15-44c1-b1db-78c71b8ec1ce",
		Name:         "Rupture Spire",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rupture Spire — sacrifice it unless you pay {1}",
				func(g *game.Game, item *game.StackItem) error {
					if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
						return nil
					}
					return UpkeepPayUnless{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Rupture Spire — pay {1} or sacrifice Rupture Spire?",
						Else:     upkeepSacrificeSource,
					}.Apply(NewContext(g, item))
				}),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
