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
// No simplification.
func init() {
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
				return removeAnyNumberOfEachKind(ctx, legal[0].ID, kinds, "Rhys, the Evermore")
			},
		}},
	})
}

// removeAnyNumberOfEachKind asks, kind by kind, how many counters of
// that kind to remove from `id` — 0 up to as many as it has as the
// question is asked — and removes that many before asking the next.
func removeAnyNumberOfEachKind(ctx *Context, id uuid.UUID, kinds []string, name string) error {
	if len(kinds) == 0 {
		return nil
	}
	kind, rest := kinds[0], kinds[1:]
	c, ok := ctx.Game.LookupCardForEffect(id)
	if z := ctx.Game.FindCardZoneForEffect(id); !ok || z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	n := c.Counters[kind]
	if n <= 0 {
		return removeAnyNumberOfEachKind(ctx, id, rest, name)
	}
	opts := make([]game.ChoiceOption, n+1)
	for i := range opts {
		opts[i] = game.ChoiceOption{Label: fmt.Sprintf("Remove %d", i)}
	}
	return PickOption{
		Question: fmt.Sprintf("%s — remove how many %s counters?", name, kind),
		Options:  opts,
		Then: func(ctx *Context, index int) error {
			if index > 0 {
				if err := (AddCounter{Target: id, Kind: kind, N: -index}).Apply(ctx); err != nil {
					return err
				}
			}
			return removeAnyNumberOfEachKind(ctx, id, rest, name)
		},
	}.Apply(ctx)
}
