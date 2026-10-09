package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lodestone Needle // Guidestone Compass — a transforming artifact
// (#2124, ADR 0137; explore #2720):
//
//	Lodestone Needle — Artifact {1}{U}
//	  "Flash
//	   When this artifact enters, tap up to one target artifact or
//	   creature and put two stun counters on it.
//	   Craft with artifact {2}{U}"
//	Guidestone Compass — Artifact
//	  "{1}, {T}: Target creature you control explores. Activate only as
//	   a sorcery."
//
// The front is Meat Locker's tap-and-stun on an enters trigger, then
// Braided Net's craft. The back is the Map token's ability without the
// sacrifice, through the CR 701.44 explore action.
//
// No simplification.
const lodestoneNeedleOracleID = "f429ca15-c578-4c1d-a604-fd5509c92dba"

func init() {
	Register(Spec{
		OracleID:        lodestoneNeedleOracleID,
		Name:            "Lodestone Needle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("Lodestone Needle — tap up to one target artifact or creature and put two stun counters on it",
				lodestoneNeedleTapAndStun),
			TargetPermanent("up to one target artifact or creature", Or(Artifact(), Creature())).WithCount(0, 1),
		)},
		Activated: []ActivatedAbility{
			Craft("Craft with artifact {2}{U}", "{2}{U}", CraftWith("artifact")),
		},
	})

	Register(Spec{
		OracleID:     lodestoneNeedleOracleID + "#1",
		Name:         "Guidestone Compass",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{1}, {T}: Target creature you control explores. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}"), TapCost()),
			Targets:      targetCreatureYouControl(),
			SorcerySpeed: true,
			Effect:       targetCreatureYouControlExplores,
		}},
	})
}

// lodestoneNeedleTapAndStun taps the target, if one was chosen and it
// is still legal, and puts two stun counters on it.
func lodestoneNeedleTapAndStun(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
		return err
	}
	return AddCounter{Target: t.ID, Kind: game.CounterStun, N: 2}.Apply(ctx)
}
