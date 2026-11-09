package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ether — Artifact {3}{U} (Final Fantasy, #4462):
//
//	"{T}, Exile this artifact: Add {U}. When you next cast an instant or
//	 sorcery spell this turn, copy that spell. You may choose new
//	 targets for the copy."
//
// Brass Infiniscope's shape: a mana ability (CR 605) whose rider
// schedules a delayed triggered ability (CR 603.7), here Galvanic
// Iteration's copy of the next instant or sorcery you cast (WhenYouNextCast
// / copyTheSpellBody). The delayed trigger belongs to the ability, so it
// outlives Ether, which is exiled as the cost; it ends at cleanup
// whether or not it fired (CR 603.7b), and it does not care whether the
// {U} was spent on that spell (the ruling). The copy is created, not
// cast. "Exile this artifact" is the engine's battlefield ExileSelf for a
// mana ability (the permanent leaves for exile as part of the cost, not
// a sacrifice).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "566533af-5e67-463f-ac33-dfcf1ec735c9",
		Name:         "Ether",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Zones:    []game.ZoneKind{game.ZoneBattlefield},
			Cost:     ManaAbilityCost{Tap: true, ExileSelf: true},
			Produced: "{U}",
			Label:    "{T}, Exile this artifact: Add {U}. When you next cast an instant or sorcery spell this turn, copy that spell. You may choose new targets for the copy.",
			Rider:    etherCopyTheNextSpell,
		}},
	})
}

// etherCopyTheNextSpell schedules the delayed copy trigger.
func etherCopyTheNextSpell(g *game.Game, controller, source uuid.UUID) error {
	g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   controller,
		SourceCardID: source,
		Label:        "Ether — copy that spell",
		On:           []game.EventKind{game.EventCast},
		Condition:    youNextCastCondition,
		CondParams:   game.EffectParams{Filter: game.CastFilter{Types: []string{"Instant", "Sorcery"}}},
		Body:         copyTheSpellBody,
	})
	return nil
}
