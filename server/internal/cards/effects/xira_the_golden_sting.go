package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Xira, the Golden Sting — Legendary Creature — Insect Assassin
// {1}{B}{R}{G}, 3/3:
//
//	"Flying, haste
//	 Whenever Xira attacks, put an egg counter on another target creature
//	 without an egg counter on it. When that creature dies, if it has an egg
//	 counter on it, draw a card and create a 1/1 black Insect creature token
//	 with flying."
//
// The attack trigger names its target as it goes on the stack (CR 603.3d)
// and puts the counter as it resolves. The second sentence is part of the
// same ability: a delayed triggered ability the resolution creates (CR
// 603.7a) about that creature as the object it is now (CR 603.7c, 400.7),
// whose controller is Xira's (CR 603.7d). It has no end but the creature's
// death (CR 603.7b: it triggers once), so it outlives the turn and Xira
// herself. Its intervening "if" (CR 603.4) reads the counters the creature
// had as it last existed on the battlefield (CR 608.2h): a creature whose
// egg counter was removed in the meantime pays nothing, and one that was
// given the counter by something else still doesn't, since the ability only
// watches the creature Xira marked. The counter's placement waits for any
// replacement prompt before the delayed ability is created.
//
// No simplification.
func init() {
	trig := WheneverThisAttacks("Xira, the Golden Sting — put an egg counter on another target creature without an egg counter on it",
		Do(Instruction(xiraPutEggCounter), Instruction(xiraWatchForTheDeath)))
	trig.Targets = Another(TargetCreature("another target creature without an egg counter on it", noEggCounter()))
	Register(Spec{
		OracleID:        "18d0e6e1-3824-465d-9245-21e0174f551f",
		Name:            "Xira, the Golden Sting",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered:       []game.TriggeredAbility{trig},
	})
}

// noEggCounter is "without an egg counter on it".
func noEggCounter() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Counters["egg"] == 0 }
}

// xiraPutEggCounter is "put an egg counter on another target creature".
func xiraPutEggCounter(ctx *Context) error {
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	return AddCounter{Target: target, Kind: "egg", N: 1}.Apply(ctx)
}

// xiraWatchForTheDeath creates "when that creature dies, if it has an egg
// counter on it, draw a card and create a 1/1 black Insect creature token
// with flying", with no end but the creature's death.
func xiraWatchForTheDeath(ctx *Context) error {
	// The announced target, not a re-judged one: it now has the egg
	// counter its own clause forbids, so it is no longer a legal target
	// but is still the creature this ability is about.
	if len(ctx.Item.Targets) == 0 {
		return nil
	}
	target := ctx.Item.Targets[0].ID
	if target == uuid.Nil {
		return nil
	}
	c, ok := battlefieldCardForEffect(ctx.Game, target)
	if !ok {
		return nil
	}
	end := game.IndefiniteDuration()
	ctx.Game.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
		Controller:   ctx.Controller(),
		SourceCardID: ctx.Source(),
		Label:        "Xira, the Golden Sting — draw a card and create a 1/1 black Insect creature token with flying",
		On:           []game.EventKind{game.EventLTB},
		Condition:    xiraEggDiedCondition,
		CondParams:   game.EffectParams{Object: game.ObjectRef{ID: target, Epoch: c.ObjectEpoch}},
		Body:         xiraEggDiedBody,
		Duration:     &end,
	})
	return nil
}

// xiraEggDiedCondition is the marked creature's death with its egg counter
// still on it, read from its last-known record (CR 603.4, 608.2h).
var xiraEggDiedCondition = game.DelayedEachCondition("dies/that-creature-with-an-egg-counter",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, p game.EffectParams) []uuid.UUID {
		ids := thatCreatureDied(ev, dt, g, p)
		if len(ids) == 0 {
			return nil
		}
		info, ok := g.PermanentForEffect(p.Object)
		if !ok || info.Counters["egg"] <= 0 {
			return nil
		}
		return ids
	})

var xiraEggDiedBody = game.SimpleDelayedBody("xira-the-golden-sting/draw-and-insect",
	func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
			return err
		}
		return CreateToken{Template: TokenCard("1/1 black Insect with flying"), N: 1}.Apply(ctx)
	})
