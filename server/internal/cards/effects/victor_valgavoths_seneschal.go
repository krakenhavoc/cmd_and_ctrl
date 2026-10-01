package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Victor, Valgavoth's Seneschal — Legendary Creature — Human Warlock
// {1}{W}{B}, 3/3:
//
//	"Eerie — Whenever an enchantment you control enters and whenever
//	 you fully unlock a Room, surveil 2 if this is the first time this
//	 ability has resolved this turn. If it's the second time, each
//	 opponent discards a card. If it's the third time, put a creature
//	 card from a graveyard onto the battlefield under your control."
//
// Eerie is one ability with two conditions (Eerie, rooms.go). The count
// is the per-object resolution tally (Game.ResolvedThisTurn, keyed on
// the stack label and the source's epoch, and including the resolution
// in progress), so a trigger that was countered does not count and a
// Victor that left and came back starts again. A fourth resolution does
// nothing. The discard is each opponent's own choice; the reanimation
// is chosen on resolution from every graveyard (it does not target).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92bf8f05-cbfd-42d8-8bc1-d7f2e46f9872",
		Name:         "Victor, Valgavoth's Seneschal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Eerie(victorEerieLabel, victorEerie),
		},
	})
}

const victorEerieLabel = "Victor, Valgavoth's Seneschal — surveil 2, then each opponent discards, then reanimate (eerie)"

func victorEerie(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	switch g.ResolvedThisTurn(item.SourceCardID, victorEerieLabel) {
	case 1:
		return Surveil{N: 2}.Apply(ctx)
	case 2:
		return eachOpponentDiscardsOne(g, item)
	case 3:
		return victorReanimate(ctx)
	}
	return nil
}

// victorReanimate puts a creature card of the controller's choice from
// any graveyard onto the battlefield under their control.
func victorReanimate(ctx *Context) error {
	var candidates []uuid.UUID
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() {
				candidates = append(candidates, c.InstanceID)
			}
		}
	}
	me := ctx.Controller()
	put := func(c *Context, id uuid.UUID) error {
		return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: me}.Apply(c)
	}
	switch len(candidates) {
	case 0:
		return nil
	case 1:
		return put(ctx, candidates[0])
	}
	item := ctx.Item
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  me,
		Source:   ctx.Source(),
		Question: "Put a creature card from a graveyard onto the battlefield under your control",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return put(NewContext(g, item), picked[0])
		},
	})
	return nil
}
