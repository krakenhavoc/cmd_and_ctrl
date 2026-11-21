package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2519ccc0-c206-41fb-aa15-2ef88025ff96",
		Name:         "Prehistoric Pet",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedByComparing(effects.OnSelf(), effects.GreaterPowerThanAttacker(), "creatures with greater power"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}{W}, {T}: Return another target creature you control to its owner's hand. Activate only during your turn.",
			Cost:      effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets:   effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			Condition: effects.DuringYourTurn(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard && t.ID != item.SourceCardID {
						return effects.BounceToHand{Target: t.ID}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
