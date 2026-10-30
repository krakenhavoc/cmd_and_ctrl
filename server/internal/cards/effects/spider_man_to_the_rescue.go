package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Man, To the Rescue — {2}{G/W} Legendary Creature — Spider Human
// Hero 3/2 (Marvel Super Heroes, #4455):
//
//	"Flash
//	 Reach, vigilance
//	 No One Dies! — When Spider-Man enters, you may tap him. When you do,
//	 another target nonattacking creature you control gains indestructible
//	 until end of turn."
//
// The "you may" is the parent trigger's optional prompt (reflexive.go's
// rule); tapping him is the doing, and the reflexive trigger (CR 603.12)
// is created only if he was actually tapped, so a Spider-Man that entered
// tapped or left in response gives nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "85dec6cf-6f27-4eb0-834b-5f6fbbe25fc8",
		Name:            "Spider-Man, To the Rescue",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "reach", "vigilance"},
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEnters(spiderManRescueLabel, spiderManTapThenReflexive),
				"No One Dies! — tap Spider-Man? (When you do, another target nonattacking creature you control gains indestructible until end of turn.)"),
		},
	})
}

const spiderManRescueLabel = "Spider-Man, To the Rescue — No One Dies!"

func spiderManTapThenReflexive(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src, ok := ctx.SourcePermanent()
	if !ok || src.Tapped {
		return nil
	}
	if err := (TapTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	return WhenYouDo("Spider-Man, To the Rescue — another target nonattacking creature you control gains indestructible", spiderManRescueBody).Apply(ctx)
}

var spiderManRescueBody = game.ReflexiveBody("spider-man-to-the-rescue/indestructible",
	simpleBody(firstTargetGainsUntilEndOfTurn(0, "Spider-Man, To the Rescue — indestructible until end of turn", "indestructible")),
	constTargets(func() *game.TargetSpec {
		return Another(TargetCreature("another target nonattacking creature you control", YouControl(), Not(AttackingCreature())))
	}))
