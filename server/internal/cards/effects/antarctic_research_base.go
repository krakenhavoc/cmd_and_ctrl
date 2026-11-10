package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Antarctic Research Base — Plane — Earth (Doctor Who, #4870):
//
//	"When you planeswalk to Antarctic Research Base and at the beginning
//	 of your upkeep, investigate.
//	 Whenever chaos ensues, put X +1/+1 counters on target creature you
//	 control, where X is the number of artifacts you control. It becomes
//	 a Plant in addition to its other types."
//
// A plane (ADR 0216 §6): its abilities function only while it is face up
// (CR 901.7), controlled by the planar controller (CR 901.6). The first is
// one ability with two trigger conditions; investigating creates a Clue
// (CR 701.16a). The chaos ability (CR 311.7) counts the artifacts as it
// resolves, and the Plant type has no stated duration, so it lasts while
// the creature stays on the battlefield (CR 611.2a). Per the ruling, with
// no artifacts the target still becomes a Plant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d65ff607-1ff5-42bd-8db2-f082cc2f0b5e",
		Name:         "Antarctic Research Base",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("Antarctic Research Base — investigate", Do(Investigate{N: 1})),
			Targeting(WheneverChaosEnsues("Antarctic Research Base — put X +1/+1 counters on target creature you control (X = artifacts you control); it becomes a Plant",
				antarcticResearchBaseChaos), TargetCreature("target creature you control", YouControl())),
		},
	})
}

func antarcticResearchBaseChaos(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	x := len(MatchingBattlefield(ctx, And(Artifact(), YouControl())))
	if err := (AddCounter{Target: t, Kind: game.CounterPlusOne, N: x}).Apply(ctx); err != nil {
		return err
	}
	return ScopedEffectFor{
		Target:   t,
		Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtype("Plant"))},
		Duration: g.PinnedTo(game.IndefiniteDuration(), t),
		Label:    "Antarctic Research Base — a Plant in addition to its other types",
	}.Apply(ctx)
}
