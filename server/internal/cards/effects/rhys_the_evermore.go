package effects

import (
	"fmt"
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rhys, the Evermore — Legendary Creature — Elf Warrior {1}{W}, 2/2:
//
//	"Flash
//	 When Rhys enters, another target creature you control gains persist
//	 until end of turn.
//	 {W}, {T}: Remove any number of counters from target creature you
//	 control. Activate only as a sorcery."
//
// "Any number" is asked at resolution, one question per kind of counter
// on the creature (0 to however many it has), so a player can strip the
// -1/-1 counter persist left and keep a +1/+1 counter, or remove nothing.
// Persist is the engine's (#2075).
//
// The kinds still to ask about are words, and the step record holds them
// (StepRecord.Strs, #4228, CR 608.2c) with the creature and how far the
// questions have got; each question's continuation is a registered key, so a
// table waiting on one is a restore point and the rest of the questions follow
// the answer.
//
// No simplification.
func init() {
	rhysRemoved = OptionPickThen("option-pick/rhys-the-evermore-remove", rhysRemovedBody)
	Register(Spec{
		OracleID:        "9f013b5d-fff0-4c50-8621-158b6dc34834",
		Name:            "Rhys, the Evermore",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Rhys, the Evermore — another creature you control gains persist until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return grantEachLegalTargetUntilEOT(NewContext(g, item), game.KeywordPersist,
						"Rhys, the Evermore — persist until end of turn")
				}), Another(TargetCreature("another target creature you control", YouControl()))),
		},
		Activated: []ActivatedAbility{{
			Label:        "{W}, {T}: Remove any number of counters from target creature you control. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{W}"), TapCost()),
			Targets:      TargetCreature("target creature you control", YouControl()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				legal := ctx.LegalTargets()
				if len(legal) == 0 {
					return nil
				}
				c, ok := g.LookupCardForEffect(legal[0].ID)
				if !ok {
					return nil
				}
				kinds := make([]string, 0, len(c.Counters))
				for kind, n := range c.Counters {
					if n > 0 {
						kinds = append(kinds, kind)
					}
				}
				sort.Strings(kinds)
				ctx.RecordStepRecord(rhysRecord, game.StepRecord{IDs: []uuid.UUID{legal[0].ID}, Ns: []int{0}, Strs: kinds})
				return rhysAskNext(ctx, item.ID)
			},
		}},
	})
}

// rhysRecord files the creature, the kinds of counter it had when the ability
// resolved (sorted) and how many of them have been asked about (#4228). The
// kinds are words, so they live in the step record's Strs, where a table
// restored in the middle of the questions still finds them.
const rhysRecord = "rhys-the-evermore/kinds"

// rhysAskNext asks, for the next kind of counter still to ask about, how many
// of them to remove (0 up to as many as the creature has as the question is
// asked), and files that the kind is asked. `itemID` is the resolving ability,
// whose record this is: the answer's continuation runs with no item of its own.
// With no kind left, or the creature gone, the ability is done.
func rhysAskNext(ctx *Context, itemID uuid.UUID) error {
	g := ctx.Game
	holder := &game.StackItem{ID: itemID}
	rec, ok := g.StepRecordForEffect(holder, rhysRecord)
	if !ok || len(rec.IDs) == 0 || len(rec.Ns) == 0 {
		return nil
	}
	id, next := rec.IDs[0], rec.Ns[0]
	for ; next < len(rec.Strs); next++ {
		c, ok := g.LookupCardForEffect(id)
		if z := g.FindCardZoneForEffect(id); !ok || z == nil || z.Kind != game.ZoneBattlefield {
			return nil
		}
		kind := rec.Strs[next]
		n := c.Counters[kind]
		if n <= 0 {
			continue
		}
		rec.Ns = []int{next + 1}
		g.RecordStepForEffect(holder, rhysRecord, rec)
		opts := make([]game.ChoiceOption, n+1)
		for i := range opts {
			opts[i] = game.ChoiceOption{Label: fmt.Sprintf("Remove %d", i)}
		}
		return PickOption{
			Question: fmt.Sprintf("Rhys, the Evermore — remove how many %s counters?", kind),
			Options:  opts,
			ThenKey:  rhysRemoved,
			Carry:    []uuid.UUID{itemID},
		}.Apply(ctx)
	}
	return nil
}

// rhysRemoved takes the counters the answer names off the kind just asked
// about, then asks the next kind. It runs from the answered prompt, also on a
// table restored while the question was open.
// It is registered in init: it asks the next question, so it refers to itself.
var rhysRemoved game.OptionPickThen

func rhysRemovedBody(ctx *Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 {
		return nil
	}
	itemID := r.Carry[0]
	rec, ok := ctx.Game.StepRecordForEffect(&game.StackItem{ID: itemID}, rhysRecord)
	if !ok || len(rec.IDs) == 0 || len(rec.Ns) == 0 || rec.Ns[0] < 1 || rec.Ns[0] > len(rec.Strs) {
		return nil
	}
	if r.Index > 0 {
		kind := rec.Strs[rec.Ns[0]-1]
		if err := (AddCounter{Target: rec.IDs[0], Kind: kind, N: -r.Index}).Apply(ctx); err != nil {
			return err
		}
	}
	return rhysAskNext(ctx, itemID)
}
