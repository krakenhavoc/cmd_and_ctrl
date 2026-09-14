package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hobgoblin Bandit Lord — Creature — Goblin Rogue {1}{R}{R}, 2/3
// (EDHREC rank 3642):
//
//	"Other Goblins you control get +1/+1.
//	 {R}, {T}: This creature deals damage equal to the number of
//	 Goblins that entered the battlefield under your control this
//	 turn to any target."
//
// The Forgotten Realms Goblin lord. The anthem is the tribal builder
// with both printed clauses — "other" and "you control". The
// activation is targeted ("any target") and counts, as it resolves,
// the Goblins that entered under the controller's control this turn
// off the event log back to the turn's upkeep — a Goblin that has
// since died still counts, as printed, and a changeling counts. A
// tap ability on a creature, so summoning sickness applies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43fea418-db6f-4953-89d1-6873b2ca41a8",
		Name:         "Hobgoblin Bandit Lord",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Goblin"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}: This creature deals damage equal to the number of Goblins that entered the battlefield under your control this turn to any target.",
			Cost:    Plus(ManaCost("{R}"), TapCost()),
			Targets: TargetAny(),
			Effect:  b34DamageChosenTargetPerGoblinEnteredThisTurn,
		}},
	})
}
