package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skemfar Shadowsage — Creature — Elf Cleric {3}{B}, 2/5 (EDHREC rank
// 4533):
//
//	"When this creature enters, choose one —
//	 • Each opponent loses X life, where X is the greatest number of
//	   creatures you control that have a creature type in common.
//	 • You gain X life, where X is the greatest number of creatures
//	   you control that have a creature type in common."
//
// The tribal deck's reach and its stabiliser in one card, and the
// choice between them is the whole point: at four life you gain, at
// four opponents' life you drain. In an Elf or Zombie list X is
// routinely eight or more.
//
// A MODAL TRIGGERED ABILITY, which the engine could not express until
// #937 (ADR 0065). The mode is chosen as the ability is PUT ON THE
// STACK (CR 603.3c) — not at resolution — so an opponent responding to
// the trigger already knows which half is coming, exactly as in paper.
// Neither bullet targets, so both are always offered.
//
// X IS COMPUTED AT RESOLUTION, over the creatures the Shadowsage's
// controller has at that moment. "The greatest number of creatures you
// control that have a creature type in common" is the size of the
// LARGEST tribe on your board, read off post-layer subtypes — so a
// changeling counts towards every tribe, a lord that grants a type
// counts, and the Shadowsage itself (an Elf Cleric) counts towards
// both Elf and Cleric. Creatures with no creature type at all
// contribute to nothing.
//
// A board of three Elves and two Clerics, two of the Elves also
// Clerics, is X = 3: the count is per TYPE, not a sum, and the
// biggest single type wins. A creature is counted once per type it
// has, which is why one creature can be in two tallies at once.
//
// Life LOSS on the first bullet, not damage, so no prevention shield
// stops it and no lifelink triggers off it (CR 119.3). It hits EACH
// opponent, and the amount is the same for all of them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "496ecb73-f508-4ac7-b101-23e03404580f",
		Name:         "Skemfar Shadowsage",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{b43ShadowsageTrigger()},
	})
}

func b43ShadowsageTrigger() game.TriggeredAbility {
	t := On(game.EventETB, b06SelfETB, "Skemfar Shadowsage — choose one",
		func(_ *game.Game, _ *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("Each opponent loses X life, where X is the greatest number of creatures you control that have a creature type in common.",
			nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				n := b43LargestTribe(ctx.Game, item.Controller)
				if n <= 0 {
					return nil
				}
				return ctx.Game.LoseLifeEachThenForEffect(item.SourceCardID, ctx.Opponents(), n, nil)
			}),
		ModeDoing("You gain X life, where X is the greatest number of creatures you control that have a creature type in common.",
			nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return GainLife{Player: item.Controller, Amount: b43LargestTribe(ctx.Game, item.Controller)}.Apply(ctx)
			}),
	)
	return t
}

// b43LargestTribe is "the greatest number of creatures you control
// that have a creature type in common" — the size of the biggest
// single creature type on `controller`'s board.
//
// Counted per TYPE, not summed: a creature with two types is in two
// tallies, and the answer is the largest tally, never their total. A
// creature with no creature type contributes to nothing, so a board of
// Eldrazi Spawn tokens and a Golem answers with the Spawn count.
//
// Post-layer subtypes through the battlefield snapshot, so a
// changeling counts towards every type it has and a creature made an
// Elf by a lord counts as one.
func b43LargestTribe(g *game.Game, controller uuid.UUID) int {
	tally := map[string]int{}
	best := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.IsCreature() {
			continue
		}
		for _, s := range c.Effective().Subtypes {
			if !game.IsCreatureType(s) {
				continue
			}
			tally[s]++
			if tally[s] > best {
				best = tally[s]
			}
		}
	}
	return best
}
