package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:        "ce70f16c-ea45-4f1d-aafb-34a5d05d31e7",
		Name:            "Kuroki, Thief of Talents",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourEndStep("Kuroki, Thief of Talents — target opponent may draw four cards; if they do, look at their hand and you may cast a spell from it for free; if they don't, put two +1/+1 counters on Kuroki",
				effects.Do(effects.Instruction(kurokiAsk), effects.Instruction(kurokiLookAndCast))),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

const kurokiDrewKey = "kuroki-thief-of-talents/drew"

// kurokiAsk is the opponent's "may draw four cards", a confirm that holds the
// resolution open: everything printed after it waits for the answer (CR
// 608.2c), and the cast is a step of the same resolution (kurokiLookAndCast),
// which is what lets it be offered mid-resolution (CR 608.2g).
func kurokiAsk(ctx *effects.Context) error {
	opp, ok := effects.FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return effects.MayChoice{
		Player:   opp,
		Question: "Kuroki, Thief of Talents — draw four cards? If you do, its controller looks at your hand and may cast a spell from it for free; if you don't, Kuroki gets two +1/+1 counters.",
		Draws:    4,
		Then:     kurokiAnswer,
	}.Apply(ctx)
}

// kurokiAnswer is the opponent's answer, keyed: yes draws four and files that
// for the step after; no puts two +1/+1 counters on Kuroki.
var kurokiAnswer = effects.MayAnswer("confirm/kuroki-thief-of-talents", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if r.Accepted {
		if err := (effects.DrawCards{Player: r.Chooser, N: 4}).Apply(ctx); err != nil {
			return err
		}
		ctx.RecordStep(kurokiDrewKey, []uuid.UUID{r.Chooser}, 1)
		return nil
	}
	if !effects.OnBattlefield(ctx.Game, r.Source) || ctx.IsNewSourceObject(r.Source) {
		return nil
	}
	return effects.AddCounter{Target: r.Source, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
})

// kurokiLookAndCast is "look at that player's hand and you may cast a spell
// from their hand without paying its mana cost", only if they drew.
func kurokiLookAndCast(ctx *effects.Context) error {
	rec := ctx.StepRecord(kurokiDrewKey)
	if len(rec.IDs) != 1 || len(rec.Ns) != 1 || rec.Ns[0] != 1 {
		return nil
	}
	g := ctx.Game
	controller := ctx.Controller()
	hand := g.LookAtHandForEffect(controller, rec.IDs[0])
	perm := g.FreeCastFromHandPermission(controller, ctx.Source(), "Kuroki, Thief of Talents")
	candidates := effects.FreeCastCandidates(g, &perm, hand)
	if len(candidates) == 0 {
		return nil
	}
	g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      controller,
		Source:       ctx.Source(),
		Reason:       "Kuroki, Thief of Talents",
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Leave it in their hand",
	})
	return nil
}
