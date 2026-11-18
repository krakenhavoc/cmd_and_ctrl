package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Primal Adversary — Creature — Wolf {2}{G}, 4/3:
//
//	"Trample
//	 When this creature enters, you may pay {1}{G} any number of times.
//	 When you pay this cost one or more times, put that many +1/+1
//	 counters on this creature, then up to that many target lands you
//	 control become 3/3 Wolf creatures with haste that are still lands."
//
// The Adversary cycle's shape (Bloodthirsty Adversary). The lands'
// change has no printed duration, so it lasts for the rest of the game
// (CR 611.2a), pinned to each land (CR 400.7). It adds the creature type
// and the Wolf subtype to a land that keeps its land types, and sets its
// base power and toughness to 3/3 with haste (CR 613.4b).
func init() {
	Register(Spec{
		OracleID:        "087d6330-c9de-42b7-a804-1202bc6aa682",
		Name:            "Primal Adversary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Primal Adversary — pay {1}{G} any number of times; counters, then lands become 3/3 Wolves", primalAdversaryEnters),
		},
	})
}

var primalAdversaryBody = game.ReflexiveBody("primal-adversary/counters-and-wolves",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := NewContext(g, item)
		if err := (CounterOnThis{Kind: game.CounterPlusOne, N: p.Amount}).Apply(ctx); err != nil {
			return err
		}
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (ScopedEffectFor{
				Target: t.ID,
				Mods: []game.Mod{
					game.AddTypesMod("Creature"),
					game.AddSubtypesMod(ctx.TextSubtypes("Wolf")...),
					game.SetBasePowerMod(3),
					game.SetBaseToughnessMod(3),
					game.AddKeywordsMod("haste"),
				},
				Duration: g.PinnedTo(game.IndefiniteDuration(), t.ID),
				Label:    "Primal Adversary — becomes a 3/3 Wolf creature with haste",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetPermanent("up to "+strconv.Itoa(p.Amount)+" target lands you control",
			And(Land(), YouControl())).WithCount(0, p.Amount)
	})

var primalAdversaryPaid = PayUnlessBranches("pay-unless/primal-adversary-times",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return ReflexiveTrigger{
			Label:  "Primal Adversary — " + strconv.Itoa(r.Number) + " +1/+1 counters, then up to that many lands become 3/3 Wolves",
			Body:   primalAdversaryBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

func primalAdversaryEnters(g *game.Game, item *game.StackItem) error {
	return MayPayX{
		Symbol: "{1}{G}",
		Label:  "Primal Adversary",
		Buys:   "put X +1/+1 counters on it, then up to X lands you control become 3/3 Wolves with haste",
		Goal:   XAsHighAsYouCan,
		Then:   primalAdversaryPaid,
	}.Apply(NewContext(g, item))
}
