package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Outcaster Trailblazer — Creature — Human Druid {2}{G}, 4/2 (Outlaws
// of Thunder Junction, #4468):
//
//	"When this creature enters, add one mana of any color.
//	 Whenever another creature you control with power 4 or greater
//	 enters, draw a card.
//	 Plot {2}{G} (You may pay {2}{G} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// The enters trigger is not a mana ability, since it triggers on
// something other than a mana ability (CR 605.1b): it uses the stack and
// adds the mana as it resolves, Lotus Cobra's pick of all five colours.
// The draw is Temur Ascendancy's power check without the "may" and
// without the Trailblazer itself: the creature's power as it enters,
// counters and anthems included. Plot is Beastbond Outcaster's special
// action (CR 702.170a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "3b5ff623-c44e-4f2d-8ab7-ee6b11f1deee",
		Name:           "Outcaster Trailblazer",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{2}{G}")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Outcaster Trailblazer — add one mana of any color",
				Do(AddMana{Produced: "{W|U|B|R|G}"})),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsCreature() && c.CurrentPower() >= 4
			}, "Outcaster Trailblazer — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
