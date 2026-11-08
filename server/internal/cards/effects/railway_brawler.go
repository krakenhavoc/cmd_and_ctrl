package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Railway Brawler — Creature — Rhino Warrior {3}{G}{G}, 5/5 (Outlaws of
// Thunder Junction, #4468):
//
//	"Reach, trample
//	 Whenever another creature you control enters, put X +1/+1 counters
//	 on it, where X is its power.
//	 Plot {3}{G} (You may pay {3}{G} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// The trigger is WheneverAnotherCreatureEntersUnderYourControl. X is the
// entering creature's power as the trigger resolves, with the layers
// applied (CR 608.2h), so an anthem or a counter it entered with counts.
// A creature that has left by then gets nothing, and a power of 0 or less
// puts no counters. Plot is the ordinary special action (CR 702.170).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b653db44-ddae-44e8-8fa6-6af3232cd586",
		Name:            "Railway Brawler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		SpecialActions:  []game.SpecialAction{Plot("{3}{G}")},
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Railway Brawler — put X +1/+1 counters on it, where X is its power", railwayBrawlerCounters),
		},
	})
}

// railwayBrawlerCounters reads the entering creature's power now and puts
// that many +1/+1 counters on it.
func railwayBrawlerCounters(g *game.Game, item *game.StackItem) error {
	id := item.Trigger.Event.CardID
	if !onBattlefield(g, id) {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.CurrentPower() <= 0 {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: c.CurrentPower()}.Apply(NewContext(g, item))
}
