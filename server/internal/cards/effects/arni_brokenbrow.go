package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arni Brokenbrow — Legendary Creature — Human Berserker {2}{R}, 3/3:
//
//	"Haste
//	 Boast — {1}: You may have Arni's base power become 1 plus the
//	 greatest power among other creatures you control until end of turn.
//	 (Activate only if this creature attacked this turn and only once
//	 each turn.)"
//
// #2731, ADR 0197 §10. The boast is the Boast constructor (CR 702.142a),
// and its whole effect is a "you may" asked as the ability resolves
// (YouMay): the choice matters because the set is a layer 7b base power,
// not a pump, so with nothing bigger beside Arni a "yes" makes it
// smaller. The new base power is counted on a "yes", on the board the
// ability resolves on (CR 608.2h); with no other creature it is 1. An
// Arni that has left the battlefield, or come back as a new object (CR
// 400.7), is not changed.
//
// The prompt warns, and the bot declines, when the new base power would
// be no higher than Arni's own (arniSetWouldRaise). That reads the
// power Arni's copiable values give it and ignores any other effect
// that already set its base power; the player's answer is never limited
// by it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3dce6c6c-599f-4e35-ae15-728c5e78dee1",
		Name:            "Arni Brokenbrow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{
			YouMay(BoastAnswering(game.AnswerPump,
				"{1}: You may have Arni's base power become 1 plus the greatest power among other creatures you control until end of turn.",
				ManaCost("{1}"), arniSetBasePower),
				MayAtResolution{
					Question: "Arni Brokenbrow — have its base power become 1 plus the greatest power among other creatures you control until end of turn?",
					Helps:    arniSetWouldRaise,
				}),
		},
	})
}

// arniNewBasePower is 1 plus the greatest power among the creatures the
// ability's controller controls other than its source, or 1 when there
// are none. Power as it is now, through the layers (CR 208.3), and a
// negative power counts as itself.
func arniNewBasePower(ctx *Context) int {
	best, any := 0, false
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.InstanceID == ctx.Source() || c.Controller != ctx.Controller() || !c.IsCreature() {
			continue
		}
		if p := c.PowerForComparison(); !any || p > best {
			best, any = p, true
		}
	}
	return 1 + best
}

// arniSourceHere reports whether the object the ability came from is
// still on the battlefield (CR 400.7).
func arniSourceHere(ctx *Context) (game.PermanentInfo, bool) {
	info, ok := ctx.SourcePermanent()
	return info, ok && !info.Left
}

// arniSetBasePower is the "yes": Arni's base power becomes the new
// value until end of turn.
func arniSetBasePower(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if _, ok := arniSourceHere(ctx); !ok {
		return nil
	}
	n := arniNewBasePower(ctx)
	return untilEndOfTurn(ctx, ctx.Source(), nil,
		fmt.Sprintf("Arni Brokenbrow — base power %d until end of turn", n),
		game.SetBasePowerMod(n))
}

// arniSetWouldRaise is the "would 'yes' help?" the prompt carries: the
// new base power is higher than the one Arni's copiable values give it.
// False when Arni has left, since "yes" then changes nothing.
func arniSetWouldRaise(ctx *Context) bool {
	info, ok := arniSourceHere(ctx)
	if !ok {
		return false
	}
	base := info.Power
	if info.Copiable != nil {
		base = info.Copiable.Power
	} else if card, found := ctx.Game.LookupCardForEffect(ctx.Source()); found {
		base = card.Power
	}
	return arniNewBasePower(ctx) > base
}
