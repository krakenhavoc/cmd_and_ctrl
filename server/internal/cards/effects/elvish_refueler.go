package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elvish Refueler — Creature — Elf Druid {2}{G}, 2/3:
//
//	"During your turn, as long as you haven't activated an exhaust
//	 ability this turn, you may activate exhaust abilities as though
//	 they haven't been activated.
//	 Exhaust — {1}{G}: Put a +1/+1 counter on this creature.
//	 (Activate each exhaust ability only once.)"
//
// The only printed card that reads the exhaust record and then tells
// one player to ignore it (#1184), and the reason
// `Game.AbilityExhausted` learned who is asking. The permission is
// not a fact about the ability — the record still says it was
// activated, and every opponent still reads it that way — so it could
// not be answered by the (source, ability) pair the gate had.
//
// It is also NOT a refresh. Nothing is cleared: activating an exhaust
// ability under this permission writes the record a SECOND time, and
// that is what makes the card self-limiting without anything having
// to turn it off. Read the clause again in order —
//
//	"as long as you haven't activated an exhaust ability this turn"
//
// — and the second activation is itself an exhaust ability activated
// this turn, so the permission is off for the rest of the turn the
// moment it is used. One extra activation per turn, on your turn, is
// the whole card.
//
// Its own {1}{G} is an ordinary exhaust ability and is a legal
// beneficiary: spend it, and on a later turn the permission lets you
// spend it again.
//
// The permission covers MANA abilities too, because a mana ability is
// an activated ability (CR 605.1a) and the gate is one predicate for
// both kinds — so Loot, the Pathfinder's spent exhaust mana ability
// is available again under the same conditions.
//
// What it does not touch: the cost, the "Activate only if …"
// condition, the {T} on an already-tapped source. "As though" changes
// the one rule it names and nothing else (CR 609.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b62f9fe-e3f4-4a21-aa6b-4dbd411d0c44",
		Name:         "Elvish Refueler",
		Completeness: CompletenessFull,
		ExhaustPermissions: []game.ExhaustPermission{
			MayActivateExhaustAbilitiesAgain(
				"During your turn, as long as you haven't activated an exhaust ability this turn, "+
					"you may activate exhaust abilities as though they haven't been activated.",
				DuringTheControllersTurn(),
				ControllerHasActivatedNoExhaustAbilityThisTurn()),
		},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — {1}{G}: Put a +1/+1 counter on this creature.",
			Exhaust: true,
			Cost:    ManaCost("{1}{G}"),
			Effect:  plusOneCountersOnThis(1),
		}},
	})
}
