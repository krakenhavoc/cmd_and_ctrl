package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kuroki, Thief of Talents — Legendary Creature — Spirit {4}{B}, 4/4 (Mystery
// Booster Commander, #5948):
//
//	"Menace
//	 At the beginning of your end step, target opponent may draw four cards.
//	 If they do, look at that player's hand and you may cast a spell from
//	 their hand without paying its mana cost. If they don't, put two +1/+1
//	 counters on Kuroki."
//
// The opponent's "may" is a keyed may-pay with a free "yes" (NoPayment), asked
// of the targeted player as the trigger resolves. On yes they draw four, then
// Kuroki's controller looks at that hand (CR 701.20e: they alone learn it) and
// may cast a spell from it for free through Silent-Blade Oni's resolution
// cast (the card is cast out of its owner's hand and goes to its owner's
// graveyard; X is 0 and additional costs are paid). On no, two +1/+1 counters
// go on Kuroki, which only the same permanent still on the battlefield gets
// (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ce70f16c-ea45-4f1d-aafb-34a5d05d31e7",
		Name:            "Kuroki, Thief of Talents",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourEndStep("Kuroki, Thief of Talents — target opponent may draw four cards; if they do, look at their hand and you may cast a spell from it for free; if they don't, put two +1/+1 counters on Kuroki",
				kurokiEndStep),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

func kurokiEndStep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	opp, ok := firstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return MayPay{
		Chooser:  opp,
		Pay:      NoPayment(),
		Question: "Kuroki, Thief of Talents — draw four cards? If you do, its controller looks at your hand and may cast a spell from it for free; if you don't, Kuroki gets two +1/+1 counters.",
		Then:     kurokiAnswer,
	}.Apply(ctx)
}

// kurokiAnswer is the keyed pair of branches: the opponent drew, or did not.
var kurokiAnswer = PayUnlessBranches("pay-unless/kuroki-thief-of-talents",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if err := (DrawCards{Player: r.Chooser, N: 4}).Apply(ctx); err != nil {
			return err
		}
		controller := ctx.Controller()
		if len(r.Carry) > 0 {
			controller = r.Carry[0]
		}
		g := ctx.Game
		hand := g.LookAtHandForEffect(controller, r.Chooser)
		perm := g.FreeCastFromHandPermission(controller, r.Source, "Kuroki, Thief of Talents")
		candidates := freeCastCandidates(g, &perm, hand)
		if len(candidates) == 0 {
			return nil
		}
		g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
			Chooser:      controller,
			Source:       r.Source,
			Reason:       "Kuroki, Thief of Talents",
			Cards:        candidates,
			Permission:   &perm,
			Keyword:      game.MayCastKeywordFreeCast,
			AcceptLabel:  "Cast it free",
			DeclineLabel: "Leave it in their hand",
		})
		return nil
	},
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if !onBattlefield(ctx.Game, r.Source) || ctx.isNewSourceObject(r.Source) {
			return nil
		}
		return AddCounter{Target: r.Source, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
	})
