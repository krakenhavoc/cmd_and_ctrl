package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aragorn, Company Leader — Legendary Creature — Human Ranger
// {1}{G}{W}, 3/3:
//
//	"Vigilance, deathtouch
//	 Whenever the Ring tempts you, if you chose a creature other than
//	 Aragorn as your Ring-bearer, put your choice of a counter from
//	 among first strike, vigilance, deathtouch, and lifelink on Aragorn.
//	 Whenever you put one or more counters on Aragorn, put one of each
//	 of those kinds of counters on up to one other target creature."
//
// The first ability is Faramir's intervening "if"
// (IfYouChoseAnotherRingBearer) with a four-way pick: the engine
// reads keyword counters itself (CR 122.1b), so a first strike counter
// is first strike. The counter is a counter put by you, so it triggers
// the second ability.
//
// The second is the #2150 shape: ONE trigger for a whole batch of
// counters put on Aragorn at once, whatever their kinds, copying every
// kind in the batch to one target. See counter_batch.go for how the
// batch and the kinds are read. A batch of +1/+1 and a keyword counter
// is one trigger and one target, not two. Counters put on Aragorn by
// another player's effect do not trigger it ("you put").
//
// "Up to one" means the trigger goes on the stack with no target and
// does nothing, as printed. One of each kind is one counter per kind,
// however many were put on Aragorn.
func init() {
	Register(Spec{
		OracleID:        "1b841e8f-9484-4737-bd37-dfa11bd06883",
		Name:            "Aragorn, Company Leader",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			IfYouChoseAnotherRingBearer(WheneverTheRingTemptsYou(
				"Aragorn, Company Leader — put your choice of a first strike, vigilance, deathtouch or lifelink counter on Aragorn",
				aragornPickACounter)),
			Targeting(
				WheneverYouPutOneOrMoreCountersOnThis(
					"Aragorn, Company Leader — put one of each of those kinds of counters on up to one other target creature",
					aragornCopyTheKinds),
				Another(TargetCreature("up to one other target creature")).WithCount(0, 1),
			),
		},
	})
}

var aragornCounterKinds = []string{game.CounterFirstStrike, game.CounterVigilance, game.CounterDeathtouch, game.CounterLifelink}

func aragornPickACounter(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	options := make([]game.ChoiceOption, len(aragornCounterKinds))
	for i, kind := range aragornCounterKinds {
		options[i] = game.ChoiceOption{Label: "A " + kind + " counter"}
	}
	return PickOption{
		Question: "Aragorn, Company Leader — put your choice of a counter on Aragorn",
		Options:  options,
		Then: func(ctx *Context, index int) error {
			if index < 0 || index >= len(aragornCounterKinds) {
				return nil
			}
			if !onBattlefield(ctx.Game, item.SourceCardID) || sourceIsNewObject(ctx.Game, item) {
				return nil
			}
			return ctx.Game.AddCounterByForEffect(item.Controller, item.SourceCardID, aragornCounterKinds[index], 1)
		},
	}.Apply(ctx)
}

func aragornCopyTheKinds(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := NewContext(g, item)
	kinds := CountersYouPutInTheSameBatch(g, item.Trigger.Event, item.Controller)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		for _, kind := range kinds {
			if err := g.AddCounterByForEffect(item.Controller, t.ID, kind, 1); err != nil {
				return err
			}
		}
	}
	return nil
}
