package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prehistoric Pet — {W} Creature — Dinosaur Ninja 1/2 (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"This creature can't be blocked by creatures with greater power.
//	 {1}{W}, {T}: Return another target creature you control to its
//	 owner's hand. Activate only during your turn."
//
// The evasion is Ant-Man, Elusive Avenger's block rule, both powers read live
// as blockers are declared (CR 509.1b). The activation is Dour Port-Mage's
// bounce of another creature you control, gated to its controller's own
// turn (CR 602.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2519ccc0-c206-41fb-aa15-2ef88025ff96",
		Name:         "Prehistoric Pet",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedByComparing(OnSelf(), GreaterPowerThanAttacker(), "creatures with greater power"),
		},
		Activated: []ActivatedAbility{{
			Label:     "{1}{W}, {T}: Return another target creature you control to its owner's hand. Activate only during your turn.",
			Cost:      Plus(ManaCost("{1}{W}"), TapCost()),
			Targets:   Another(TargetCreature("another target creature you control", YouControl())),
			Condition: DuringYourTurn(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard && t.ID != item.SourceCardID {
						return BounceToHand{Target: t.ID}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
