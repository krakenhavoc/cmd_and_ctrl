package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b850_helpers.go — shared pieces of S82 OLD batch 850 (Ice Age, #5636).

// S82OldB850DepletionCounter is the counter kind the Ice Age depletion lands
// print. Same word as MmqDepletionCounter: it is one counter kind.
const S82OldB850DepletionCounter = "depletion"

// S82OldB850DepletionLand is the whole Spec of an Ice Age depletion land:
//
//	"This land doesn't untap during your untap step if it has a depletion
//	 counter on it.
//	 At the beginning of your upkeep, remove a depletion counter from this land.
//	 {T}: Add {a} or {b}. Put a depletion counter on this land."
//
// The counter is the mana ability's rider (it is part of the mana ability, so
// it does not use the stack, CR 605.3b). The untap clause reads the counter at
// the untap step (CR 502.3); the upkeep trigger comes after it, so a land
// tapped for mana sits out exactly one untap step.
func S82OldB850DepletionLand(oracleID, name, a, b string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{{
			Label: "doesn't untap during your untap step if it has a depletion counter on it",
			Restricts: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target != nil && source != nil && target.InstanceID == source.InstanceID &&
					target.Counters[S82OldB850DepletionCounter] > 0
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep(name+" — remove a depletion counter from this land", RemoveACounterFromThis(S82OldB850DepletionCounter)),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{" + a + "|" + b + "}",
			Label:    "{T}: Add {" + a + "} or {" + b + "}. Put a depletion counter on this land.",
			Rider: func(g *game.Game, _ uuid.UUID, source uuid.UUID) error {
				return g.AddCounterForEffect(source, S82OldB850DepletionCounter, 1)
			},
		}},
	}
}

// TargetsThisCreature is "target spell that targets this creature"
// (Mistfolk): any spell on the stack with the clause's own source among its
// targets (game.TargetSpec.ItemTargetsSource).
func TargetsThisCreature() *game.TargetSpec {
	spec := TargetSpell("target spell that targets this creature")
	spec.ItemTargetsSource = true
	return spec
}
