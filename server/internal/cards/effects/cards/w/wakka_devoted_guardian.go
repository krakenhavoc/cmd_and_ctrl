package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	damage := effects.WheneverThisDealsCombatDamageToAPlayer(
		"Wakka, Devoted Guardian — destroy up to one artifact that player controls and put a +1/+1 counter on Wakka",
		wakkaDestroyAndGrow)
	damage.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		return effects.TargetPermanent("up to one target artifact that player controls",
			effects.Artifact(), effects.ControlledBy(tc.Event.Target)).WithCount(0, 1)
	}
	effects.Register(effects.Spec{
		OracleID:        "fbd37284-35ff-4409-b6d5-f813c06c4931",
		Name:            "Wakka, Devoted Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{
			damage,
			effects.On(game.EventBeginEndStep,
				effects.AllOf(effects.ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.ACounterWasPutOnThisObjectThisTurn(g, source.InstanceID)
				}),
				"Blitzball Captain — a +1/+1 counter on each other creature you control",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.ACounterWasPutOnThisObjectThisTurn(g, item.SourceCardID) {
						return nil
					}
					return effects.PutCounterOnEachOtherCreatureYouControl(g, item)
				}),
		},
	})
}

func wakkaDestroyAndGrow(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
