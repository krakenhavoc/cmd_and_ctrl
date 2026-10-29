package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Black Panther, Vanguard — Legendary Creature — Human Warrior Hero
// {2}{G}{W}, 4/4 (Marvel Super Heroes, #4455):
//
//	"Whenever another nontoken Hero you control enters, choose one —
//	 • Create a 1/1 white Soldier creature token.
//	 • Creatures you control get +1/+1 until end of turn."
//
// Felidar Retreat's shape: the trigger itself does nothing and each bullet
// is a ModeDoing body, so the choice is the CR 603.3c mode_pick prompt.
// "Another" excludes Black Panther himself, "nontoken" excludes a Hero
// token, and the Hero is read off the entering permanent's effective
// subtypes, so a changeling counts. The second bullet pumps the creatures
// you control as it resolves, the Hero that just entered included.
//
// No simplification.
func init() {
	trigger := On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, ok := enteredUnderYourControl(ev, source, g, true)
		return ok && c.IsCreature() && !c.IsToken() && c.HasSubtype(source.TextSubtype("Hero"))
	}, "Black Panther, Vanguard — choose one", func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("Create a 1/1 white Soldier creature token.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1}.Apply(ctx)
			}),
		ModeDoing("Creatures you control get +1/+1 until end of turn.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
					Label: "Black Panther, Vanguard — creatures you control get +1/+1"}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "b0bf1964-c543-4c53-9f09-0dd7c608d5c5",
		Name:         "Black Panther, Vanguard",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
