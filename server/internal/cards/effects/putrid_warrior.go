package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Putrid Warrior — Creature — Zombie Soldier Warrior {W}{B}, 2/2
// (Apocalypse, #5614):
//
//	"Whenever this creature deals damage, choose one —
//	 • Each player loses 1 life.
//	 • Each player gains 1 life."
//
// El-Hajjâj's reading of "deals damage" (thisDealtDamage): one trigger
// for each damage event the Warrior is the source of, combat or not, to
// a player or a permanent. The mode is a real "choose one" on the
// trigger (TriggeredAbility.Modes), picked as it is put on the stack
// (CR 700.2b). "Each player" is every player still in the game, the
// controller included (eachPlayerLifeChanges).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ec24c24-7e87-4c12-b94e-3cd5b184beb5",
		Name:         "Putrid Warrior",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: thisDealtDamage,
			Key:       "Putrid Warrior — each player loses 1 life, or each player gains 1 life",
			Effect:    func(*game.Game, *game.StackItem) error { return nil },
			Modes: ChooseOne(
				ModeDoing("Each player loses 1 life.", nil,
					func(_ *game.StackItem, ctx *Context, _ int) error { return eachPlayerLifeChanges(ctx, -1) }),
				ModeDoing("Each player gains 1 life.", nil,
					func(_ *game.StackItem, ctx *Context, _ int) error { return eachPlayerLifeChanges(ctx, 1) }),
			),
		}},
	})
}
