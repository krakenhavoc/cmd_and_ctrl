package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Waterbender's Restoration — Instant — Lesson {U}{U}:
//
//	"As an additional cost to cast this spell, waterbend {X}. (While
//	 paying a waterbend cost, you can tap your artifacts and
//	 creatures to help. Each one pays for {1}.)
//	 Exile X target creatures you control. Return those cards to the
//	 battlefield under their owner's control at the beginning of the
//	 next end step."
//
// The canonical delayed blink: the creatures leave now and come back
// a step boundary later as new objects, which is what makes the card
// both an ETB engine and a fog against targeted removal.
//
// The waterbend cost is the card. X is not a number the caster picks
// for free — it is the size of a cost they have to pay, in mana or by
// tapping their own board, and every permanent tapped to buy another
// blink is a permanent that isn't blocking. Ship the clause without
// the cost and you have a two-mana mass blink.
//
// Which is exactly how it shipped in S22 (#255) and what issue #259
// tracked: with no tap-permanents-as-a-cost component in the engine
// and no {X} in the printed mana cost, there was nothing to charge,
// so the target clause read "any number of target creatures you
// control" for free — the only card in this catalog knowingly
// STRONGER than paper.
//
// It is costed now. `Waterbend("{X}")` layers an extra {X} onto the
// {U}{U}, the caster announces X at cast time, taps artifacts and
// creatures to pay for {1} apiece (or pays it in mana, or both), and
// `CountFromX` ties the target count to that same X: blinking three
// creatures costs {U}{U} plus three. No simplifications remain.
func init() {
	Register(Spec{
		OracleID:     "285046f6-b3c4-4eb7-8712-9dffebabc762",
		Name:         "Waterbender's Restoration",
		Completeness: CompletenessFull,
		TapCost:      Waterbend("{X}"),
		Targets:      targetsCountedByX(TargetCreature("X target creatures you control", YouControl())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			exiled, err := exileTargetsForDelayedReturn(ctx)
			if err != nil {
				return err
			}
			if len(exiled) == 0 {
				return nil
			}
			return ScheduleDelayedTrigger{
				At:     game.StepEnd,
				Label:  "Waterbender's Restoration — return the exiled creatures",
				Cards:  exiled,
				Effect: returnExiledCardsToOwners,
			}.Apply(ctx)
		},
	})
}

// targetsCountedByX marks a clause whose target count is the X
// announced at cast time — "Exile X target creatures you control" —
// rather than a printed constant. The cast path resolves Min and Max
// to the announced value before validating anything against it.
//
// A helper rather than a raw field poke so the one card using it
// reads as a clause instead of as a struct edit, and so a second one
// doesn't have to rediscover that Min and Max must be left alone.
func targetsCountedByX(spec *game.TargetSpec) *game.TargetSpec {
	spec.CountFromX = true
	return spec
}
