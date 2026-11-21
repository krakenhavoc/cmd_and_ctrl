package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
const LodestoneNeedleOracleID = "f429ca15-c578-4c1d-a604-fd5509c92dba"

func init() {
	effects.Register(effects.Spec{
		OracleID:        LodestoneNeedleOracleID,
		Name:            "Lodestone Needle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.WhenThisEnters("Lodestone Needle — tap up to one target artifact or creature and put two stun counters on it",
				lodestoneNeedleTapAndStun),
			effects.TargetPermanent("up to one target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())).WithCount(0, 1),
		)},
		Activated: []effects.ActivatedAbility{
			effects.Craft("Craft with artifact {2}{U}", "{2}{U}", effects.CraftWith("artifact")),
		},
	})

	effects.Register(effects.Spec{
		OracleID:     LodestoneNeedleOracleID + "#1",
		Name:         "Guidestone Compass",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}, {T}: Target creature you control explores. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets:      effects.TargetCreatureYouControl(),
			SorcerySpeed: true,
			Effect:       effects.TargetCreatureYouControlExplores,
		}},
	})
}

// lodestoneNeedleTapAndStun taps the target, if one was chosen and it
// is still legal, and puts two stun counters on it.
func lodestoneNeedleTapAndStun(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
		return err
	}
	return effects.AddCounter{Target: t.ID, Kind: game.CounterStun, N: 2}.Apply(ctx)
}
