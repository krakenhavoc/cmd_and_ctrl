package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Solar Array — Artifact {3}:
//
//	"{T}: Add one mana of any color. When you next cast an artifact
//	 spell this turn, that spell gains sunburst. (If it's a creature,
//	 it enters with a +1/+1 counter on it for each color of mana spent
//	 to cast it. Otherwise, it enters with that many charge counters on
//	 it.)"
//
// A mana ability whose effect also creates a delayed triggered ability
// (CR 603.7): the next artifact spell its controller casts this turn
// gains sunburst (ADR 0109 §11, #1552). The trigger is data — the
// "you next cast" condition and a registered body — so it survives a
// restore point, and it ends in the cleanup step whether or not it
// fired. Tapped while paying for an artifact spell, it is already
// waiting when that spell becomes cast (CR 601.2g-i), so it gives that
// very spell sunburst, as printed. The spell carries the keyword onto
// the permanent (CR 400.7a), where the engine counts it like a printed
// one.
//
// The grant arrives after the costs are paid, so the payment is not
// spread across colours for it. One declared simplification, the
// paid-cost record's: with strict mana off the engine never saw what
// paid, so the spell counts no colours (ADR 0068 §3).
func init() {
	Register(Spec{
		OracleID:     "e8e2f273-5e74-4f16-8d49-2e86e9c9f2dc",
		Name:         "Solar Array",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't track which mana you spent, so the artifact you cast next gets no counters from it."},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color. When you next cast an artifact spell this turn, that spell gains sunburst",
			Rider:    solarArrayNextArtifactGainsSunburst,
		}},
	})
}

// solarArrayNextArtifactGainsSunburst schedules the delayed trigger.
func solarArrayNextArtifactGainsSunburst(g *game.Game, controller, source uuid.UUID) error {
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   controller,
		SourceCardID: source,
		Label:        "Solar Array — that spell gains sunburst",
		On:           []game.EventKind{game.EventCast},
		Condition:    youNextCastCondition,
		CondParams:   game.EffectParams{Filter: game.CastFilter{Types: []string{"Artifact"}}},
		Body:         thatSpellGainsSunburstBody,
	})
	return nil
}
