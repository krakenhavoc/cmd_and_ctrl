package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emiel the Blessed — Legendary Creature — Unicorn {2}{W}{W}, 4/4
// (Jumpstart; the Selesnya deck request's commander, #4893):
//
//	"{3}: Exile another target creature you control, then return it to
//	 the battlefield under its owner's control.
//	 Whenever another creature you control enters, you may pay {G/W}. If
//	 you do, put a +1/+1 counter on it. If it's a Unicorn, put two +1/+1
//	 counters on it instead. ({G/W} can be paid with either {G} or {W}.)"
//
// The blink is The Mind Stone's "under its owner's control" flicker on
// another creature you control, so a borrowed creature goes home. The
// returned creature is a new object and triggers the second ability.
//
// The entry trigger asks for the hybrid {G/W} as it resolves (either
// colour pays it). "It" is the object that entered: the paid branch
// reads it from the trigger, and a creature that left (or left and came
// back as a new object) before the answer gets nothing. Whether it is a
// Unicorn is read as the counters are put on it, so a creature that
// became a Unicorn in response gets two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b11c250c-f191-4c52-ba02-a9176f163447",
		Name:         "Emiel the Blessed",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}: Exile another target creature you control, then return it to the battlefield under its owner's control.",
			Cost:    ManaCost("{3}"),
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect:  flickerFirstLegalTargetUnderOwner,
		}},
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Emiel the Blessed — pay {G/W} to put a +1/+1 counter on it (two if it's a Unicorn)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
						return nil
					}
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{G/W}",
						Question: "Emiel the Blessed — pay {G/W} to put a +1/+1 counter on the creature that entered (two if it's a Unicorn)?",
						Then:     emielTheBlessedPaid,
					}.Apply(ctx)
				}),
		},
	})
}

// emielTheBlessedPaid is the paid half, keyed: the trigger's event names
// the creature, and its epoch tells it from a new object (CR 400.7).
var emielTheBlessedPaid = PayUnlessBranches("pay-unless/emiel-the-blessed-counters",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		obj := ctx.Trigger().Object
		if obj == nil {
			return nil
		}
		if info, ok := ctx.Game.PermanentForEffect(obj.Ref()); !ok || info.Left {
			return nil
		}
		c, ok := ctx.Game.LookupCardForEffect(obj.Ref().ID)
		if !ok {
			return nil
		}
		n := 1
		if c.HasSubtype(ctx.TextSubtype("Unicorn")) {
			n = 2
		}
		return AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
	}, nil)
