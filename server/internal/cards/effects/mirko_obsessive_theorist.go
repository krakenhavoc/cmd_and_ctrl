package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mirko, Obsessive Theorist — Legendary Creature — Vampire Detective
// {1}{U}{B}, 1/3:
//
//	"Flying, vigilance
//	 Whenever you surveil, put a +1/+1 counter on Mirko.
//	 At the beginning of your end step, you may return target creature card
//	 with power less than Mirko's from your graveyard to the battlefield
//	 with a finality counter on it. (If it would die, exile it instead.)"
//
// The counter is the batch's surveil trigger. The end-step trigger is
// targeted and optional: the target is a creature card in your graveyard
// whose power is less than Mirko's (RelativeToSource, LesserPower: judged
// against Mirko's effective power as the target is chosen and again as the
// trigger resolves, last-known if Mirko has left), and the "may" is asked as
// it resolves. The card returns with the finality counter as it enters
// (Coalstoke Gearhulk's ReturnFromGraveyardWithCountersKeyedForEffect), and
// one that left the graveyard in response stays where it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1692f19e-4906-4808-965f-5f0284004f55",
		Name:            "Mirko, Obsessive Theorist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			cmdr2B27WheneverYouSurveil("Mirko, Obsessive Theorist — put a +1/+1 counter on Mirko",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			Optional(Targeting(
				AtYourEndStep("Mirko, Obsessive Theorist — you may return target creature card with power less than Mirko's from your graveyard to the battlefield with a finality counter on it",
					cmdr2B27MirkoReturn),
				RelativeToSource(TargetCardInGraveyard("target creature card with power less than Mirko's power from your graveyard",
					Creature(), YouOwn()), LesserPower())),
				"Return a creature card from your graveyard to the battlefield with a finality counter on it?"),
		},
	})
}

// cmdr2B27MirkoReturn is the end-step trigger's effect: the target returns
// with a finality counter, if it is still a legal target in your graveyard.
func cmdr2B27MirkoReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) == 0 {
		return nil
	}
	id := legal[0].ID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ReturnFromGraveyardWithCountersKeyedForEffect(id, item.Controller, false,
		map[string]int{game.CounterFinality: 1}, game.KeyedTail{})
}
