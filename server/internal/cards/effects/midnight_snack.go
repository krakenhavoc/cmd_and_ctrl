package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Midnight Snack — Enchantment {2}{B} (Foundations, #4465):
//
//	"Raid — At the beginning of your end step, if you attacked this turn,
//	 create a Food token.
//	 {2}{B}, Sacrifice this enchantment: Target opponent loses X life,
//	 where X is the amount of life you gained this turn."
//
// The Food is Searslicer Goblin's raid end-step shape. X is read as the
// ability resolves (CR 608.2h) off the turn's life-gained tally, which
// counts every gain, lifelink and Food included, and not a loss; the
// Enchantment is already sacrificed by then, and the tally does not care.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ddcaee9-a11e-477a-bd19-f5ee10512549",
		Name:         "Midnight Snack",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, AllOf(ByYou, tmcB6YouAttackedThisTurn),
				"Midnight Snack — Raid: create a Food token",
				Do(If{Cond: avatarB11CondYouAttacked{}, Then: []Applier{
					CreateToken{Template: FoodToken(), N: 1},
				}})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice this enchantment: Target opponent loses X life, where X is the amount of life you gained this turn.",
			Cost:    Plus(ManaCost("{2}{B}"), SacrificeThis()),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				victim, ok := firstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				return GainLife{Player: victim, Amount: -b15LifeGainedThisTurn(g, ctx.Controller())}.Apply(ctx)
			},
		}},
	})
}
