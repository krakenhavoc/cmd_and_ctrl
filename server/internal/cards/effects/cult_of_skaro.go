package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cult of Skaro — Legendary Artifact Creature — Dalek {1}{U}{B}{R}, 4/4:
//
//	"Whenever Cult of Skaro attacks, choose one at random —
//	 • Thay — Put a +1/+1 counter on each artifact creature you control.
//	 • Caan — Draw two cards.
//	 • Sec — Create a 3/3 black Dalek artifact creature token with menace.
//	 • Jast — Each opponent loses 4 life."
//
// The first card on ChooseOneAtRandom (CR 700.2, #3399): the engine
// picks the bullet as the trigger goes on the stack, from the seeded
// random stream, and none of the four needs a target. Thay counts the
// artifact creatures as it resolves, Cult of Skaro itself included.
//
// No simplification.
func init() {
	const name = "Cult of Skaro"
	t := WheneverThisAttacks(name+" — choose one at random: counters, draw two, a Dalek, or drain 4",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOneAtRandom(
		ModeDoing("Thay — Put a +1/+1 counter on each artifact creature you control.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.Controller != item.Controller || !c.IsCreature() || !c.IsArtifact() {
						continue
					}
					if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}),
		ModeDoing("Caan — Draw two cards.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
			}),
		ModeDoing("Sec — Create a 3/3 black Dalek artifact creature token with menace.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{
					Controller: item.Controller,
					Template:   TokenCard("3/3 black Dalek artifact with menace"),
					N:          1,
				}.Apply(ctx)
			}),
		ModeDoing("Jast — Each opponent loses 4 life.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return EachOpponentLosesLife{N: 4}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "1cf40a30-5b42-4dce-ae4f-699bc879e180",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
