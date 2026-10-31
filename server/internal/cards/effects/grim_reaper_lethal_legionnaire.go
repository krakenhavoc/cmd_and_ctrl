package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grim Reaper, Lethal Legionnaire — Legendary Creature — Human Villain
// {3}{B}, 3/4 (Marvel Super Heroes, #4455, seam #4575):
//
//	"Whenever Grim Reaper attacks, you may pay {3}{B}. When you do,
//	 return target creature card from your graveyard to the battlefield
//	 tapped and attacking with a finality counter on it. (If a creature
//	 with a finality counter on it would die, exile it instead.)"
//
// Ant-Man's MayPay / CR 603.12 pair: the payment is the whole condition
// and the return is a second triggered ability whose target is chosen
// when it goes on the stack, so the creature is picked after the mana is
// spent. The card enters through the CR 506.3c door (the Chorale of the
// Void and Interceptor path), tapped and attacking whichever player,
// planeswalker or battle its controller picks, and with the counter as an
// entry counter (CR 614.1c), so it is there for every replacement and
// trigger that sees the entry. It was never declared as an attacker, so it
// triggers nothing that watches attacks.
//
// "If a creature with a finality counter on it would die, exile it
// instead" is not on this card: it is the engine's CR 122.1h replacement
// (game/finality_counters.go), which applies to any permanent with the
// counter, whatever put it there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "edf429a8-6a4d-4d2a-b0ca-c0c72a0698a3",
		Name:         "Grim Reaper, Lethal Legionnaire",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(grimReaperPayLabel, grimReaperPayOffer),
		},
	})
}

const grimReaperPayLabel = "Grim Reaper, Lethal Legionnaire — you may pay {3}{B}: return target creature card from your graveyard tapped and attacking with a finality counter"

// grimReaperPayOffer asks for the {3}{B}; the reflexive trigger lives in
// the paid half.
func grimReaperPayOffer(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{3}{B}",
		Question: "Grim Reaper, Lethal Legionnaire — pay {3}{B} to return a creature card from your graveyard tapped and attacking with a finality counter?",
		Then:     grimReaperPaid,
	}.Apply(NewContext(g, item))
}

// grimReaperPaid is the paid half, keyed: the reflexive trigger is queued
// against the asking item.
var grimReaperPaid = PayUnlessBranches("pay-unless/grim-reaper-lethal-legionnaire-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return ReflexiveTrigger{
			Label: "Grim Reaper, Lethal Legionnaire — return target creature card from your graveyard tapped and attacking with a finality counter",
			Body:  grimReaperReturnBody,
		}.Apply(ctx)
	}, nil)

// grimReaperReturnBody is "when you do, return target creature card from
// your graveyard to the battlefield tapped and attacking with a finality
// counter on it".
var grimReaperReturnBody = game.ReflexiveBody("grim-reaper-lethal-legionnaire/return",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, id := range legalTargetIDs(ctx) {
			return putAttackingWithFinality(ctx, id, game.ZoneGraveyard, "Grim Reaper, Lethal Legionnaire")
		}
		return nil
	}),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn())
	}))
