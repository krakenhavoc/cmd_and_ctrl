package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kalastria Highborn — Creature — Vampire Shaman {B}{B}, 2/2 (Foundations,
// #4465):
//
//	"Whenever this creature or another Vampire you control dies, you may
//	 pay {B}. If you do, target player loses 2 life and you gain 2 life."
//
// One trigger per Vampire that dies (the ruling), each choosing its
// target player as it goes on the stack. The {B} is paid on resolution
// (MayPay, registered branch), and the whole ability does nothing when the
// target is no longer legal (CR 608.2b): the pay prompt is never asked, so
// the controller gains nothing. "Vampire" is read as the creature last
// existed on the battlefield (CR 603.10a, leftAsSubtype), so a changeling
// or a creature a lord made a Vampire counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77e319da-d782-4258-bf0e-626755e5adc9",
		Name:         "Kalastria Highborn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				if !ok || leftUnderControlOf(ev, dead) != source.Controller {
					return false
				}
				return ev.CardID == source.InstanceID || leftAsSubtype(ev, dead, source.TextSubtype("Vampire"))
			},
			Targets: TargetPlayer("target player"),
			Key:     "Kalastria Highborn — you may pay {B}: target player loses 2 life and you gain 2 life",
			Purpose: game.Purpose{DeathPayoff: true},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if len(ctx.LegalTargets()) == 0 {
					return nil
				}
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{B}",
					Question: "Kalastria Highborn — pay {B}? If you do, target player loses 2 life and you gain 2 life.",
					Then:     kalastriaHighbornPaid,
				}.Apply(ctx)
			},
		}},
	})
}

// kalastriaHighbornPaid is the "if you do" half: the target loses 2 life
// and the controller gains 2. An on-disk key: never renamed, never reused.
var kalastriaHighbornPaid = PayUnlessBranches("pay-unless/kalastria-highborn-drain",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return Do(TargetPlayerLosesLife{N: 2}, GainLife{Amount: 2})(ctx.Game, ctx.Item)
	}, nil)
