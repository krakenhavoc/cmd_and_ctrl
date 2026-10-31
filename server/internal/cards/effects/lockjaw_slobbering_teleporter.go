package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lockjaw, Slobbering Teleporter — {1}{U} Legendary Creature — Inhuman Dog
// Hero 1/1 (Marvel Super Heroes Commander, #4455):
//
//	"Vigilance
//	 At the beginning of combat on your turn, if you've cast a noncreature
//	 spell this turn, put a +1/+1 counter on Lockjaw. When you do, Lockjaw
//	 and up to one other target creature you control can't be blocked this
//	 turn."
//
// The trigger has Franklin Richards' intervening if (CR 603.4): it is
// checked as combat begins and again as the ability resolves. The "when you
// do" is a CR 603.12 reflexive trigger, raised only if the counter really
// landed (Red Hulk's shape); it chooses its own "up to one other target" as
// it goes on the stack and makes Lockjaw and that creature unblockable
// until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5efce07e-24c4-4520-a46f-9954b67c4e6e",
		Name:            "Lockjaw, Slobbering Teleporter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, true), franklinCastNoncreature),
				"Lockjaw — put a +1/+1 counter on him; when you do, he and up to one other target creature you control can't be blocked this turn",
				msc13LockjawCounter),
		},
	})
}

func msc13LockjawCounter(g *game.Game, item *game.StackItem) error {
	if g.CastTallyFor(item.Controller).Noncreature == 0 || !b09SourceStillOnBattlefield(g, item) {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return ReflexiveTrigger{
		Label: "Lockjaw — he and up to one other target creature you control can't be blocked this turn",
		Body:  msc13LockjawBody,
	}.Apply(ctx)
}

var msc13LockjawBody = game.ReflexiveBody("lockjaw/unblockable", simpleBody(msc13LockjawUnblockable),
	constTargets(func() *game.TargetSpec {
		return Another(TargetCreature("up to one other target creature you control", YouControl())).WithCount(0, 1)
	}))

// msc13LockjawUnblockable makes the source and each legal target unblockable
// until end of turn.
func msc13LockjawUnblockable(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := []uuid.UUID{item.SourceCardID}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	for _, id := range ids {
		if !onBattlefield(g, id) {
			continue
		}
		if err := (RestrictUntilEOT{
			Target:       id,
			Restrictions: game.CantBeBlocked,
			Label:        "Lockjaw — can't be blocked",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
