package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Taeko, the Patient Avalanche — Legendary Creature — Turtle Ninja {3}{U},
// 4/5 (Foundations Jumpstart, #5929):
//
//	"Taeko enters tapped.
//	 Whenever another creature you control leaves the battlefield, if it
//	 didn't die, scry 1 and put a +1/+1 counter on Taeko.
//	 Whenever Taeko attacks, you may pay {U/B}. When you do, target attacking
//	 creature can't be blocked this turn."
//
// The entry is a self replacement (CR 614.1c). The leave trigger is Aang, at
// the Crossroads's "another creature you control leaves the battlefield",
// narrowed by the intervening "if it didn't die": the event's destination is
// anywhere but a graveyard (CR 700.4), a token that is sacrificed or destroyed
// having gone there first. The scry and the counter are one resolution. The
// attack trigger is a real "you may pay" with a reflexive trigger
// (CR 603.12) whose target is chosen as it goes on the stack and is judged
// again as it resolves, so a creature that left combat in between is no
// longer a legal target; the restriction lasts until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d3eb043-8ce6-461a-b8bc-67f0ba6cd580",
		Name:         "Taeko, the Patient Avalanche",
		Completeness: CompletenessFull,
		// The leave trigger's scry (ADR 0126 §6).
		Purpose:      game.Purpose{Scry: 1},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ev.NewZone != game.ZoneGraveyard && anotherCreatureYouControlLeft(ev, source, lki, g)
			}, "Taeko — scry 1 and put a +1/+1 counter on her (a creature left without dying)",
				Do(Scry{N: 1}, CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			WheneverThisAttacks("Taeko — you may pay {U/B}; when you do, target attacking creature can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{U/B}",
						Question: "Taeko, the Patient Avalanche — pay {U/B} to make target attacking creature unblockable this turn?",
						Then:     taekoPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

var taekoPaid = PayUnlessBranches("pay-unless/bns-b52-taeko-unblockable",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label: "Taeko — target attacking creature can't be blocked this turn",
			Body:  taekoUnblockableBody,
		}.Apply(ctx)
	}, nil)

var taekoUnblockableBody = game.ReflexiveBody("bns-b52/taeko-cant-be-blocked",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := b16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Taeko — can't be blocked"}.Apply(ctx)
	}),
	constTargets(func() *game.TargetSpec {
		return TargetCreature("target attacking creature", AttackingCreature())
	}))
