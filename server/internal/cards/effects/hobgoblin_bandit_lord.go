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
// One declared simplification, weaker than printed and inherited from
// the tally: a creature that was a Goblin at the moment it entered
// ONLY because of another permanent's static ability — a Maskwood
// Nexus already on the battlefield, a Conspiracy naming Goblin — is
// not counted. The layer cache is rebuilt after an entry rather than
// during it, so the tally reads the permanent's own types and a
// granted one is invisible to it.
func init() {
	Register(Spec{
		OracleID:     "43fea418-db6f-4953-89d1-6873b2ca41a8",
		Name:         "Hobgoblin Bandit Lord",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A creature that was only a Goblin because of another permanent (a Maskwood Nexus, a Conspiracy) when it entered isn't counted.",
		},
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
