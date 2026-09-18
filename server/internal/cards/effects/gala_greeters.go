package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// b15GalaGreetersLabel is the stack label the per-turn tally keys
// on — every resolution of this trigger is one mode used up.
const b15GalaGreetersLabel = "Gala Greeters — alliance"

// Gala Greeters — Creature — Elf Druid {1}{G}, 1/1 (EDHREC rank
// 1706):
//
//	"Alliance — Whenever another creature you control enters, choose
//	 one that hasn't been chosen this turn —
//	 • Put a +1/+1 counter on this creature.
//	 • Create a tapped Treasure token.
//	 • You gain 2 life."
//
// A two-drop that pays a counter, a Treasure and two life for the
// first three creatures each turn. Corpse Knight's condition
// (b13AnotherCreatureYouControlEntered); the modes are each one
// primitive.
//
// #764 made it a REAL mode choice. TriggeredAbility.Modes is the same
// game.ModeSpec a modal spell declares, and the choice happens as the
// ability is put on the stack (CR 603.3c) through a mode_pick prompt
// to the controller — before targets, before anyone gets priority,
// and a full priority round before resolution, which is the timing a
// chain of confirms at resolution could not have reached. The trigger
// targets nothing, so the prompt is the only thing between the
// harvest and the stack.
//
// Declared simplification, STRONGER than printed and the reason this
// card is still CompletenessCaveats: "that hasn't been chosen this
// turn" is a per-source tally over the turn, and the engine has no
// seam for one (ADR 0065 "Out of scope"). Every bullet is offered
// every time, so a controller with three triggers in a turn may take
// the Treasure three times rather than one of each. The old shape —
// the first unused mode in printed order — was weaker than printed
// and made the choice for the player; this one asks the question the
// card asks and does not yet enforce the restriction on the answer.
func init() {
	Register(Spec{
		OracleID:     "cce081eb-8820-415a-a7b2-3c5b9d4a2601",
		Name:         "Gala Greeters",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You choose the mode each time, but \"that hasn't been chosen this turn\" isn't enforced — the same mode can be chosen twice in a turn."},
		Triggered: []game.TriggeredAbility{
			b15GalaGreetersTrigger(),
		},
	})
}

// b15GalaGreetersTrigger is the alliance trigger with its three
// bullets. Each bullet declares its own body (ModeDoing), so the
// engine dispatches the chosen one at resolution in announce order
// (CR 700.2c) and the card file writes no switch.
func b15GalaGreetersTrigger() game.TriggeredAbility {
	t := WheneverAnotherCreatureEntersUnderYourControl(b15GalaGreetersLabel,
		func(g *game.Game, item *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("Put a +1/+1 counter on this creature.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				if !b15OnBattlefield(ctx.Game, item.SourceCardID) {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: 1}.Apply(ctx)
			}),
		ModeDoing("Create a tapped Treasure token.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return b13CreateTappedTreasures(ctx, item.Controller, 1)
			}),
		ModeDoing("You gain 2 life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
			}),
	)
	return t
}
