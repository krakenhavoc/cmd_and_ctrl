package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Case of the Filched Falcon — Enchantment — Case {U} (Murders at Karlov
// Manor, #4469):
//
//	"When this Case enters, investigate. (Create a Clue token. It's an
//	 artifact with "{2}, Sacrifice this token: Draw a card.")
//	 To solve — You control three or more artifacts. (If unsolved, solve
//	 at the beginning of your end step.)
//	 Solved — {2}{U}, Sacrifice this Case: Put four +1/+1 counters on
//	 target noncreature artifact. It becomes a 0/0 Bird creature with
//	 flying in addition to its other types."
//
// The Case (ADR 0071) is Case of the Stashed Skeleton's skeleton: ToSolve
// for the condition, an activated ability gated on Solved() for the
// payoff. The solve condition counts the artifacts its controller
// controls at the end step and again on resolution (CR 603.4).
//
// The payoff runs in printed order. The counters go on first, while the
// artifact is still not a creature, so a replacement that reads "a
// creature" does not see them; then the indefinite type change (CR
// 611.2a, pinned to that object, Tezzeret, Cruel Captain's Robot shape)
// adds Creature, Bird, base 0/0 and flying. An artifact that became a
// creature in response is no longer a legal target, and the ability does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33b31fb0-71ce-408a-a256-2a5a3b085cb6",
		Name:         "Case of the Filched Falcon",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Case of the Filched Falcon — investigate", Do(Investigate{N: 1})),
			ToSolve("Case of the Filched Falcon — to solve: you control three or more artifacts", youControlThreeOrMoreArtifacts),
		},
		Activated: []ActivatedAbility{{
			Label:      "{2}{U}, Sacrifice this Case: Put four +1/+1 counters on target noncreature artifact. It becomes a 0/0 Bird creature with flying in addition to its other types.",
			Cost:       Plus(ManaCost("{2}{U}"), SacrificeThis()),
			ActiveWhen: Solved(),
			Targets:    TargetPermanent("target noncreature artifact", Artifact(), Noncreature()),
			Effect: Do(
				Instruction(func(ctx *Context) error {
					id := FirstLegalBattlefieldTarget(ctx)
					if id == uuid.Nil {
						return nil
					}
					return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 4}.Apply(ctx)
				}),
				Instruction(caseOfTheFilchedFalconBecomesABird),
			),
		}},
	})
}

// caseOfTheFilchedFalconBecomesABird is "It becomes a 0/0 Bird creature
// with flying in addition to its other types": no duration is stated,
// so the change is indefinite, pinned to the target object.
func caseOfTheFilchedFalconBecomesABird(ctx *Context) error {
	id := FirstLegalBattlefieldTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	bird := []game.Mod{
		game.AddTypesMod("Creature"),
		game.AddSubtypesMod(ctx.TextSubtypes("Bird")...),
		game.SetBasePowerMod(0),
		game.SetBaseToughnessMod(0),
		game.AddKeywordsMod("flying"),
	}
	duration := ctx.Game.PinnedTo(game.IndefiniteDuration(), id)
	return ScopedEffectFor{Target: id, Mods: bird, Duration: duration,
		Label: "Case of the Filched Falcon — becomes a 0/0 Bird creature with flying"}.Apply(ctx)
}

// youControlThreeOrMoreArtifacts is the Case's solve condition.
func youControlThreeOrMoreArtifacts(g *game.Game, controller, _ uuid.UUID) bool {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsArtifact() {
			n++
		}
	}
	return n >= 3
}
