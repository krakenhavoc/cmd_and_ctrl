package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fallaji Vanguard — Creature — Human Soldier {2}{R}{W}, 2/3 (The Brothers'
// War):
//
//	"First strike
//	 Whenever this creature or another creature you control enters, target
//	 creature gets +2/+0 until end of turn."
//
// Jaws of Defeat's trigger: every creature entering under your control,
// tokens and the Vanguard itself included (CreatureEnteredUnderYourControl),
// puts its own trigger on the stack with its own target. The target is any
// creature, chosen as the trigger goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e60b23ba-2258-4da1-b873-452ee215846b",
		Name:            "Fallaji Vanguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, CreatureEnteredUnderYourControl,
				"Fallaji Vanguard — target creature gets +2/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || !onBattlefield(g, t.ID) {
						return nil
					}
					return BoostUntilEOT{Target: t.ID, Power: 2, Label: "Fallaji Vanguard — +2/+0"}.Apply(ctx)
				}),
				TargetCreature("target creature")),
		},
	})
}
