package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// the Goblins that entered under the controller's control this turn.
//
// "THAT ENTERED … THIS TURN" IS A TALLY OF ENTRIES, not a look at the
// board (CR 603.10 / CR 608.2h, #811). The count reads the engine's
// per-turn subtype tally (Game.EnteredWithSubtypeThisTurn), which
// records each permanent's subtypes and the player it entered under
// at its EventETB, so:
//
//   - a Goblin token that entered and has since died still counts, as
//     printed — the case the card is actually played for, and the one
//     the old event-log walk lost, because CR 704.5d had already taken
//     the token out of the graveyard;
//   - a creature that entered as something else and has since BECOME a
//     Goblin does not count, because it did not enter as one;
//   - a Goblin that entered under an opponent's control and that you
//     have since gained control of does not count for you;
//   - a changeling counts (CR 702.73a).
//
// A tap ability on a creature, so summoning sickness applies.
//
// A creature that is a Goblin as it enters only through another
// permanent's static ability — a Maskwood Nexus already on the
// battlefield — counts: the tally reads the permanent with the layer
// pass applied (#2940).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43fea418-db6f-4953-89d1-6873b2ca41a8",
		Name:         "Hobgoblin Bandit Lord",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Goblin"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}: This creature deals damage equal to the number of Goblins that entered the battlefield under your control this turn to any target.",
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Effect:  effects.B34DamageChosenTargetPerGoblinEnteredThisTurn,
		}},
	})
}
