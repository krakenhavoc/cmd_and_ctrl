package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wakka, Devoted Guardian — Legendary Creature — Human Warrior
// {2}{G}{W}, 4/4:
//
//	"Reach, trample
//	 Whenever Wakka deals combat damage to a player, destroy up to one
//	 target artifact that player controls and put a +1/+1 counter on
//	 Wakka.
//	 Blitzball Captain — At the beginning of your end step, if a counter
//	 was put on Wakka this turn, put a +1/+1 counter on each other
//	 creature you control."
//
// "That player" is the damaged player, a fact about this trigger, so the
// clause is built per trigger from its event (TargetsFrom). The target
// is "up to one": with no artifact picked, or the pick gone, Wakka still
// gets its counter.
//
// The end-step condition is an intervening if (CR 603.4), asked as the
// step begins and again on resolution, and read off this turn's event
// log (ACounterWasPutOnThisObjectThisTurn): any counter of any kind, put
// by anyone, on this Wakka object. Its own combat-damage counter counts,
// which is the card's point. Mikaeus's body puts the counters.
//
// No simplification.
func init() {
	damage := WheneverThisDealsCombatDamageToAPlayer(
		"Wakka, Devoted Guardian — destroy up to one artifact that player controls and put a +1/+1 counter on Wakka",
		wakkaDestroyAndGrow)
	damage.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		return TargetPermanent("up to one target artifact that player controls",
			Artifact(), ControlledBy(tc.Event.Target)).WithCount(0, 1)
	}
	Register(Spec{
		OracleID:        "fbd37284-35ff-4409-b6d5-f813c06c4931",
		Name:            "Wakka, Devoted Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{
			damage,
			On(game.EventBeginEndStep,
				AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ACounterWasPutOnThisObjectThisTurn(g, source.InstanceID)
				}),
				"Blitzball Captain — a +1/+1 counter on each other creature you control",
				func(g *game.Game, item *game.StackItem) error {
					if !ACounterWasPutOnThisObjectThisTurn(g, item.SourceCardID) {
						return nil
					}
					return putCounterOnEachOtherCreatureYouControl(g, item)
				}),
		},
	})
}

func wakkaDestroyAndGrow(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
