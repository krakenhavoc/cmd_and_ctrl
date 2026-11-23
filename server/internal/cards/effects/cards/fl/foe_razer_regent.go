package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foe-Razer Regent — Creature — Dragon {5}{G}{G}, 4/5 (Commander Starter
// Decks):
//
//	"Flying
//	 When this creature enters, you may have it fight target creature you
//	 don't control.
//	 Whenever a creature you control fights, put two +1/+1 counters on it at
//	 the beginning of the next end step."
//
// The enters trigger is Somberwald Stag's. The fight trigger fires once per
// fighter you control (a fight emits one EventFight for each fighter, stamped
// with that fighter's controller, so two of your creatures fighting each other
// is two triggers, CR 701.14), and each schedules a delayed trigger
// (Protean Hydra's) pinned to the fighter it was scheduled for: it gets the
// counters at the next end step only if it is still that object on the
// battlefield.
//
// No simplification.
func init() {
	const name = "Foe-Razer Regent"
	effects.Register(effects.Spec{
		OracleID:        "e7dd3e2c-34a5-4097-b6d1-2ccd097fe57f",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters(name+" — it fights target creature you don't control",
					effects.B30SourceFightsFirstLegalTarget),
				effects.TargetCreature("target creature you don't control", effects.OpponentControls())),
				name+" — have it fight target creature you don't control?"),
			effects.On(game.EventFight, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			}, name+" — put two +1/+1 counters on it at the beginning of the next end step",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					fighter := item.Trigger.Event.CardID
					c, ok := g.LookupCardForEffect(fighter)
					if !ok || !effects.OnBattlefield(g, fighter) {
						return nil
					}
					return effects.ScheduleDelayedTrigger{
						Label:  name + " — put two +1/+1 counters on it",
						Cards:  []uuid.UUID{fighter},
						Body:   foeRazerRegentCountersBody,
						Params: game.EffectParams{Object: game.ObjectRef{ID: fighter, Epoch: c.ObjectEpoch}},
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var foeRazerRegentCountersBody = game.DelayedBody("foe-razer-regent/two-counters", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch || !effects.OnBattlefield(g, id) {
		return nil
	}
	return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}.Apply(effects.NewContext(g, item))
})
