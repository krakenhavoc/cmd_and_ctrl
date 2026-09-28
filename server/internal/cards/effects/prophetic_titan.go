package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prophetic Titan — {4}{U}{R} 4/4 Creature — Giant Wizard:
//
//	"Delirium — When this creature enters, choose one. If there are
//	 four or more card types among cards in your graveyard, choose
//	 both instead.
//	 • This creature deals 4 damage to any target.
//	 • Look at the top four cards of your library. Put one of them into
//	   your hand and the rest on the bottom of your library in a random
//	   order."
//
// The forced count of #1655: with delirium the controller must take
// BOTH bullets — one is not an answer — so InsteadIf raises the
// minimum with the maximum. The count is read as the trigger goes on
// the stack (CR 603.3c) and carried on the mode_pick prompt; a card
// type that leaves the graveyard in response changes nothing.
//
// The damage comes from the Titan (item.SourceCardID), last known
// information if it has left. "Put one of them" is mandatory, a look
// rather than a reveal. No simplification.
func init() {
	t := WhenThisEnters("Prophetic Titan — choose one; both with delirium",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("This creature deals 4 damage to any target.",
			TargetAny(),
			DealFixedDamageToModesTarget(4)),
		ModeDoing("Look at the top four cards of your library. Put one of them into your hand and the rest on the bottom of your library in a random order.",
			nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				player := item.Controller
				return TakeFromLibraryToHand{
					Player: player,
					Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
					Max:    1,
					Label:  "Prophetic Titan — put one into your hand",
					Then:   TakeRestOnBottomInRandomOrder,
				}.Apply(ctx)
			}),
	).InsteadIf(2, DeliriumForModes)
	Register(Spec{
		OracleID:     "224f5f1a-4f31-4935-bf5a-910fd0a666a1",
		Name:         "Prophetic Titan",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
