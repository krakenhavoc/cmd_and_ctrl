package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hidetsugu Consumes All // Vessel of the All-Consuming — a transforming
// Saga (Kamigawa: Neon Dynasty, #5674).
//
// Front face, Enchantment — Saga {1}{B}{R}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.)
//	 I — Destroy each nonland permanent with mana value 1 or less.
//	 II — Exile all graveyards.
//	 III — Exile this Saga, then return it to the battlefield transformed
//	       under your control."
//
// Back face, Enchantment Creature — Ogre Shaman, 3/3:
//
//	"Trample
//	 Whenever this creature deals damage, put a +1/+1 counter on it.
//	 Whenever this creature deals damage to a player, if it has dealt 10
//	 or more damage to that player this turn, they lose the game."
//
// Chapter I goes through the destroy pipeline, every player's permanents
// alike, so an indestructible permanent survives; a token's mana value is
// 0 unless it copies something. Chapter II is Rest in Peace's entry
// sweep, every seat's graveyard. Chapter III is Fable of the
// Mirror-Breaker's: an exile and a return, so the Vessel is a new object
// with no lore counters (CR 400.7, 712.14a).
//
// The counter ability triggers once each time the Vessel deals damage,
// however many recipients that damage had (CR 603.2c, 510.2): trample
// damage split between a blocker and a player is one counter, and the
// first-strike and regular steps of a double striker are two
// (ThatMuchDamageDealtAtLeast(1), one group per damage instance).
//
// The last ability is an intervening "if" (CR 603.4) over the turn's
// record of what this OBJECT dealt to that player, combat and noncombat
// together (TurnTally.DamagedPlayers). It is checked as the damage is
// dealt, that damage included, and again as the trigger resolves,
// against the object that dealt it even if the Vessel has left since
// (CR 608.2h). A Vessel that left and came back is a new object that has
// dealt nothing (CR 400.7). The loss is CR 104.3e's, so a player who
// can't lose the game doesn't.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     HidetsuguConsumesAllOracleID,
		Name:         "Hidetsugu Consumes All",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.ChapterTrigger(1, effects.SagaChapterLabel("Hidetsugu Consumes All", 1, "destroy each nonland permanent with mana value 1 or less"),
				effects.Do(effects.DestroyAllMatching{Match: effects.And(effects.Nonland(), effects.ManaValueLE(1))})),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel("Hidetsugu Consumes All", 2, "exile all graveyards"),
				effects.B02ExileAllGraveyards),
			effects.ChapterTrigger(3, effects.SagaChapterLabel("Hidetsugu Consumes All", 3, "exile it, then return it transformed"),
				effects.ChapterExileAndReturnTransformed),
		},
	})

	const loseLabel = "Vessel of the All-Consuming — that player loses the game"
	effects.Register(effects.Spec{
		OracleID:        HidetsuguConsumesAllOracleID + "#1",
		Name:            "Vessel of the All-Consuming",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WithAtBatchEnd(effects.On(game.EventDealDamage, vesselOfTheAllConsumingDealtDamage,
				"Vessel of the All-Consuming — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				effects.ThatMuchDamageDealtAtLeast(1)),
			effects.InterveningIf(effects.CondBoard{
				Name: "it has dealt 10 or more damage to that player this turn",
				Test: func(ctx *effects.Context) bool {
					ref, ok := ctx.SourceRef()
					return ok && vesselOfTheAllConsumingDealtTen(ctx.Game, ref, ctx.Trigger().Event.Target)
				},
			}, effects.On(game.EventDealDamage,
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return effects.ThisDealtDamageToAPlayer(ev, source, lki, g) &&
						vesselOfTheAllConsumingDealtTen(g, game.ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}, ev.Target)
				},
				loseLabel,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.LoseTheGame{Player: ctx.Trigger().Event.Target}.Apply(ctx)
				})),
		},
	})
}

// hidetsuguConsumesAllOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const HidetsuguConsumesAllOracleID = "5b8be856-c08d-4356-925d-fce843565189"

// vesselOfTheAllConsumingDealtDamage is one slice of the damage the
// Vessel dealt, to a player or a permanent, combat or not.
func vesselOfTheAllConsumingDealtDamage(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventDealDamage && ev.Amount > 0 && ev.Source == source.InstanceID
}

// vesselOfTheAllConsumingDealtTen is the intervening "if": the Vessel
// object `ref` has dealt 10 or more damage to `player` this turn.
func vesselOfTheAllConsumingDealtTen(g *game.Game, ref game.ObjectRef, player uuid.UUID) bool {
	return g.DamageDealtToPlayerThisTurnByObjectForEffect(ref, player) >= 10
}
