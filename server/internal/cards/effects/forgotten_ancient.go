package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forgotten Ancient — Creature — Elemental {3}{G}, 0/3:
//
//	"Whenever a player casts a spell, you may put a +1/+1 counter on
//	 this creature.
//	 At the beginning of your upkeep, you may move any number of +1/+1
//	 counters from this creature onto other creatures."
//
// Any player's spell, the "you may" asked as the trigger goes on the
// stack.
//
// The upkeep move is chosen as it resolves (CR 608.2c): first which
// other creatures receive counters, from every player's board (a
// resolution-time choice, so not targeted), then, creature by creature,
// how many of the Ancient's remaining +1/+1 counters go there. Choosing
// none is the "you may" declined. Each move puts the counters on the
// receiving creature and takes as many off the Ancient (CR 122.5), and
// only while both are still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ef75cf3-30dc-4852-a6d1-1d6391fad022",
		Name:         "Forgotten Ancient",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, AnyPlayer, "Forgotten Ancient — put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
				"Forgotten Ancient — put a +1/+1 counter on it?"),
			AtYourUpkeep("Forgotten Ancient — move any number of +1/+1 counters onto other creatures", forgottenAncientMove),
		},
	})
	forgottenAncientMoveRest = RegisterThenSteps("then-steps/forgotten-ancient-move", forgottenAncientMoveSteps()...)
}

func forgottenAncientMove(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	self := item.SourceCardID
	if sourceIsNewObject(g, item) || !onBattlefield(g, self) || forgottenAncientCounters(g, self) == 0 {
		return nil
	}
	var seats []uuid.UUID
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			seats = append(seats, p.ID)
		}
	}
	return ChoosePermanents{
		Of:       seats,
		Question: "Forgotten Ancient — choose creatures to move +1/+1 counters onto (none to keep them)",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var ids []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.IsCreature() && c.InstanceID != self {
					ids = append(ids, c.InstanceID)
				}
			}
			return ids, 0, len(ids)
		},
		Then: func(ctx *Context, picked game.PromptedPicks) error {
			return forgottenAncientMoveOnto(ctx, self, picked.Cards())
		},
	}.Apply(ctx)
}

// forgottenAncientMoveOnto asks, for the first creature in `onto`, how
// many of the Ancient's +1/+1 counters to move there, moves them, and
// goes on to the rest.
func forgottenAncientMoveOnto(ctx *Context, self uuid.UUID, onto []uuid.UUID) error {
	for len(onto) > 0 {
		to, rest := onto[0], onto[1:]
		have := forgottenAncientCounters(ctx.Game, self)
		if have == 0 || !onBattlefield(ctx.Game, self) {
			return nil
		}
		if !onBattlefield(ctx.Game, to) {
			onto = rest
			continue
		}
		name := "that creature"
		if c, ok := ctx.Game.LookupCardForEffect(to); ok && c.Name != "" {
			name = c.Name
		}
		return ChooseNumber{
			Min:      0,
			Max:      have,
			Question: "Forgotten Ancient — how many +1/+1 counters onto " + name + "?",
			Unit:     game.PayAmountCounters,
			Then: func(ctx *Context, n int) error {
				if n <= 0 || !onBattlefield(ctx.Game, self) || !onBattlefield(ctx.Game, to) {
					n = 0
				}
				// The rest of the move, and the next creature's
				// question, wait for a prompt the counters ask (a
				// CR 616 order; #4162, CR 608.2c).
				ctx.RecordStep(forgottenAncientRecord, append([]uuid.UUID{to}, rest...), n)
				return forgottenAncientMoveRest.Run(ctx)
			},
		}.Apply(ctx)
	}
	return nil
}

// forgottenAncientRecord is the creature the counters move onto, then
// the creatures still to ask about; and how many move (0: none).
const forgottenAncientRecord = "forgotten-ancient/move"

// forgottenAncientMoveRest is one move, then the next creature. It is
// registered in init, because its last step asks about the next
// creature, whose answer runs it again.
var forgottenAncientMoveRest ThenSteps

func forgottenAncientMoveSteps() []Applier {
	return []Applier{
		Instruction(func(ctx *Context) error {
			r := ctx.StepRecord(forgottenAncientRecord)
			if r.N() <= 0 || len(r.IDs) == 0 {
				return nil
			}
			return ctx.Game.AddCounterByForEffect(ctx.Controller(), r.IDs[0], game.CounterPlusOne, r.N())
		}),
		Instruction(func(ctx *Context) error {
			n, self := ctx.StepRecord(forgottenAncientRecord).N(), ctx.Item.SourceCardID
			if n <= 0 || !onBattlefield(ctx.Game, self) {
				return nil
			}
			return ctx.Game.AddCounterForEffect(self, game.CounterPlusOne, -n)
		}),
		Instruction(func(ctx *Context) error {
			r := ctx.StepRecord(forgottenAncientRecord)
			if len(r.IDs) == 0 {
				return nil
			}
			return forgottenAncientMoveOnto(ctx, ctx.Item.SourceCardID, r.IDs[1:])
		}),
	}
}

func forgottenAncientCounters(g *game.Game, id uuid.UUID) int {
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return 0
	}
	return c.Counters[game.CounterPlusOne]
}
