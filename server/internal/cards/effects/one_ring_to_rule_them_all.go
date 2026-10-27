package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// One Ring to Rule Them All — Enchantment — Saga {2}{B}{B}:
//
//	"(As this Saga enters and after your draw step, add a lore
//	 counter. Sacrifice after III.)
//	 I — The Ring tempts you, then each player mills cards equal to
//	     your Ring-bearer's power.
//	 II — Destroy all nonlegendary creatures.
//	 III — Each opponent loses 1 life for each creature card in that
//	     player's graveyard."
//
// Chapter I reads the power of your Ring-bearer after the tempt, so it
// is the creature just chosen; with no creature, nothing is milled,
// and a negative power mills nothing. Each player mills in APNAP order, one at a time.
// Chapter III counts each opponent's own graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c75add5e-4e9d-4c68-a8ad-b878b1dbc0a4",
		Name:         "One Ring to Rule Them All",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "One Ring to Rule Them All — the Ring tempts you, then each player mills", oneRingTemptThenMill),
			TriggerWithPurpose(ChapterTrigger(2, "One Ring to Rule Them All — destroy all nonlegendary creatures", func(g *game.Game, item *game.StackItem) error {
				return DestroyAllMatching{Match: And(Creature(), Not(Legendary()))}.Apply(NewContext(g, item))
			}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
			ChapterTrigger(3, "One Ring to Rule Them All — each opponent loses life for their creature cards", oneRingDrain),
		},
	})
}

// oneRingTemptThenMill is chapter I.
// Keyed (#4253 slice 10b): the mill is the tempt's tail.
func oneRingTemptThenMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return TheRingTemptsYou{ThenTail: ctx.ItemTail(oneRingMillTail)}.Apply(ctx)
}

// oneRingMillTail is "Then each player mills cards equal to the power of
// your Ring-bearer". An on-disk key: never renamed, never reused.
var oneRingMillTail = RingTail("then/one-ring-to-rule-them-all-mill", func(ctx *Context, _ uuid.UUID) error {
	ctx.Game.RecomputeLayersIfStaleLocked()
	n := 0
	if rb := game.RingBearerOf(ctx.Game, ctx.Controller()); rb != uuid.Nil {
		if c, ok := ctx.Game.LookupCardForEffect(rb); ok {
			n = c.CurrentPower()
		}
	}
	// One library at a time in APNAP order (#4266).
	return millEachPlayerInTurn(ctx, apnapPlayers(ctx.Game), constantMill(n), game.KeyedTail{})
})

// oneRingDrain is chapter III.
func oneRingDrain(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		p := g.PlayerByIDForEffect(opp)
		if p == nil || p.Graveyard == nil {
			continue
		}
		n := 0
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() {
				n++
			}
		}
		if n == 0 {
			continue
		}
		if err := g.ChangePlayerLifeForEffect(ctx.Source(), opp, -n); err != nil {
			return err
		}
	}
	return nil
}
