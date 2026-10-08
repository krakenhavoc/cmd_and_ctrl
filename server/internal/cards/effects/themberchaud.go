package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Themberchaud — Legendary Creature — Dragon {4}{R}{R}{R}, 5/5:
//
//	"Trample
//	 When Themberchaud enters, he deals X damage to each other creature
//	 without flying and each player, where X is the number of Mountains
//	 you control.
//	 You may exert Themberchaud as he attacks. When you do, he gains
//	 flying until end of turn. (An exerted creature won't untap during
//	 your next untap step.)"
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with a linked "when
// you do" (CR 607.2h) that grants flying before blockers.
//
// The enters trigger counts Mountains as it resolves (CR 608.2h),
// reading effective subtypes, so a land an effect made a Mountain
// counts and a Mountain Themberchaud's own damage can't touch. It hits
// every creature without flying except Themberchaud himself, and every
// player still in the game, the controller included, all as one damage
// instance. A Themberchaud that left and came back is a new object and
// is not "other" for the old trigger (CR 400.7).
//
// No simplification.
func init() {
	const label = "Themberchaud — X damage to each other creature without flying and each player, where X is the number of Mountains you control"
	Register(Spec{
		OracleID:        "a422a5b0-1aff-4d07-bcc5-edfbda09dc44",
		Name:            "Themberchaud",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		// ADR 0126 §6: the damage is the number of Mountains, counted as the trigger resolves, so the sweep is declared as the destruction it usually is, an upper bound (Chain Reaction's shape). Partial: fliers and Themberchaud are spared.
		Purpose:       game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := mountainsControlledBy(g, item.Controller)
				self := item.SourceCardID
				if sourceIsNewObject(g, item) {
					self = uuid.Nil
				}
				other := func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != self }
				return damageEachMatchingAndEachPlayerThen(ctx, And(Creature(), WithoutKeyword("flying"), other), x, nil)
			}),
			ExertedGets("Themberchaud — he gains flying until end of turn", 0, 0, "flying"),
		},
	})
}

// mountainsControlledBy counts the Mountains `controller` controls, by
// effective subtype.
func mountainsControlledBy(g *game.Game, controller uuid.UUID) int {
	return b12OtherMountainsControlled(g, controller, uuid.Nil)
}
