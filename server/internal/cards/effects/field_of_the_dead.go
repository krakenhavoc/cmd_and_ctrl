package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Field of the Dead — Land (EDHREC rank 486):
//
//	"This land enters tapped.
//	 {T}: Add {C}.
//	 Whenever this land or another land you control enters, if you
//	 control seven or more lands with different names, create a 2/2
//	 black Zombie creature token."
//
// The landfall engine that got banned in Standard: in a
// singleton-ish Commander mana base the seventh differently named
// land is turn six or seven, and every land after it is a Zombie.
//
// "This land or another land you control" is the landfall clause
// including the Field's own entry, and the Field is on the
// battlefield by the time its own ETB event fires, so its name
// counts toward the seven — as printed.
//
// The intervening "if" (CR 603.4) is checked TWICE: in AppliesTo
// when the land enters, and again inside the effect when the trigger
// resolves, so a land destroyed in response to the trigger stops the
// Zombie exactly as in paper. That is a step past the catalog's
// older intervening-if posture (Garruk's Uprising checks once), and
// the reason this card is not flagged stronger than printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aa959340-c869-4caa-92c7-572bd8d23eef",
		Name:         "Field of the Dead",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsLand() && b04LandNamesControlled(g, source.Controller) >= 7
			}, "Field of the Dead — create a 2/2 Zombie", func(g *game.Game, item *game.StackItem) error {
				if b04LandNamesControlled(g, item.Controller) < 7 {
					return nil // CR 603.4: the "if" is re-checked on resolution.
				}
				return CreateToken{
					Controller: item.Controller,
					Template:   b04ZombieToken(),
					N:          1,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
