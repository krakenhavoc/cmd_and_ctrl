package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shaile, Dean of Radiance // Embrose, Dean of Shadow — a modal double-faced
// card (Strixhaven: School of Mages, #5502).
//
//	Shaile, Dean of Radiance   {1}{W}   1/1
//	Legendary Creature — Bird Cleric
//	  Flying, vigilance
//	  {T}: Put a +1/+1 counter on each creature that entered the battlefield
//	  under your control this turn.
//
//	Embrose, Dean of Shadow    {2}{B}{B}   4/4
//	Legendary Creature — Human Warlock
//	  {T}: Put a +1/+1 counter on another target creature, then Embrose deals
//	  2 damage to that creature.
//	  Whenever a creature you control with a +1/+1 counter on it dies, draw a
//	  card.
//
// Shaile's set is every creature on the battlefield that entered it this turn
// under its controller's control now (the per-object entry tally, a creature
// that entered on an opponent's turn did not enter this turn); it is read as
// the ability resolves and Shaile herself counts if she entered this turn.
//
// Embrose: the counter and then the damage, from Embrose, to a creature other
// than Embrose; a target that is gone makes both do nothing (CR 608.2b). The
// draw is Gladehart Cavalry's dies trigger: the dead creature's counters are
// read as it last existed on the battlefield (CR 603.10a) and "you control" is
// the controller it died under, Embrose herself included.
//
// No simplification.
func init() {
	const oracle = "b0907441-30fe-436a-b29c-585b57eabfc3"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Shaile, Dean of Radiance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []effects.ActivatedAbility{{
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Label:   "{T}: Put a +1/+1 counter on each creature that entered the battlefield under your control this turn.",
			Cost:    effects.TapCost(),
			Effect: effects.PutAPlusOneCounterOnEachYouControl(func(g *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.IsCreature() && g.EnteredThisTurn(c.InstanceID)
			}),
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Embrose, Dean of Shadow",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Put a +1/+1 counter on another target creature, then Embrose deals 2 damage to that creature.",
			Cost:    effects.TapCost(),
			Targets: effects.Another(effects.TargetCreature("another target creature")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 2}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller &&
					g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
			}, "Embrose, Dean of Shadow — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
