package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Beckett Mariner, Impetuous Ensign — Legendary Creature — Human Officer
// {1}{R}, 2/1 (Star Trek, #4973):
//
//	"First strike
//	 Reckless Behavior — Whenever Beckett Mariner becomes tapped, put a
//	 promotion counter on her. Then if there are three or more promotion
//	 counters on her, remove those counters and she deals 2 damage to
//	 each opponent."
//
// "Reckless Behavior" is a flavor word (CR 207.2d). "Becomes tapped" is
// the tap event and the attack that taps her (ThisBecameTapped, #4534),
// so a vigilance attack is not one. The count is read once the promotion
// counter LANDS (AddCounterByKeyedForEffect's continuation), so the third
// tap removes all of them and deals the damage in the same resolution.
// If she has left the battlefield by resolution no counter is put on her,
// nothing is removed, and the check reads the counters she last had there
// (CR 608.2h); the damage is dealt by her as she last existed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1033a0cb-1a00-46c4-83d5-68a6eac85feb",
		Name:            "Beckett Mariner, Impetuous Ensign",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Beckett Mariner — put a promotion counter on her, then at three remove them and deal 2 damage to each opponent",
				beckettMarinerPromote),
		},
	})
}

// beckettMarinerCounter is the promotion counter's kind.
const beckettMarinerCounter = "promotion"

// beckettMarinerPromote puts the promotion counter on her and runs the
// check once it has landed; with her gone (or a new object), the check
// alone, against what she last had.
func beckettMarinerPromote(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !b15OnBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return beckettMarinerCheck(ctx)
	}
	return g.AddCounterByKeyedForEffect(uuid.Nil, item.SourceCardID, beckettMarinerCounter, 1,
		ctx.ItemTail(beckettMarinerCheckTail))
}

// beckettMarinerCheckTail is the "Then if …" after the counter settles.
// An on-disk identity: never renamed, never reused.
var beckettMarinerCheckTail = CardTail("beckett-mariner/promotion-check", func(ctx *Context, _ game.TailOutcome) error {
	return beckettMarinerCheck(ctx)
})

// beckettMarinerCheck is "if there are three or more promotion counters
// on her, remove those counters and she deals 2 damage to each opponent".
func beckettMarinerCheck(ctx *Context) error {
	info, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	n := info.Counters[beckettMarinerCounter]
	if n < 3 {
		return nil
	}
	if !info.Left {
		if err := ctx.Game.AddCounterForEffect(ctx.Item.SourceCardID, beckettMarinerCounter, -n); err != nil {
			return err
		}
	}
	return DamageEachOpponent{N: 2}.Apply(ctx)
}
