package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mutated Cultist — Creature — Eldrazi Horror {2}{B}, 1/3:
//
//	"Devoid (This card has no color.)
//	 When you cast this spell, remove all counters from up to one target
//	 permanent or opponent. The next spell you cast this turn costs {1}
//	 less to cast for each counter removed this way.
//	 Deathtouch"
//
// The cast trigger (Abundant Maw's WhenYouCastThisSpell) is announced with
// "up to one" permanent or opponent as its target, so it may go on the stack
// with none. Resolving it removes every counter of every kind from the
// permanent (Aether Snap's removal) or from the opponent (energy, poison,
// experience, rad ...), counting the counters that actually came off, and
// promises the next spell its controller casts this turn a {1} reduction
// for each (Hardened Berserker's NextSpellPromise); no counters removed
// promises nothing. The target is re-checked as the trigger resolves
// (CR 608.2b).
func init() {
	cast := WhenYouCastThisSpell("Mutated Cultist — remove all counters from up to one target permanent or opponent; the next spell you cast this turn costs {1} less for each counter removed this way",
		mutatedCultistRemove)
	cast.Targets = mutatedCultistTarget()
	Register(Spec{
		OracleID:        "5168be9b-ffca-41c5-a944-92a1a2e30779",
		Name:            "Mutated Cultist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "deathtouch"},
		Triggered:       []game.TriggeredAbility{cast},
	})
}

// mutatedCultistTarget is "up to one target permanent or opponent".
func mutatedCultistTarget() *game.TargetSpec {
	return (&game.TargetSpec{
		Mode:    "any",
		Label:   "up to one target permanent or opponent",
		Players: true,
		PlayerOK: func(g *game.Game, caster uuid.UUID, p *game.Player) bool {
			return Opponent()(g, caster, p)
		},
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(*game.Game, uuid.UUID, game.Card, game.ZoneKind) bool {
			return true
		},
		Min: 1, Max: 1,
	}).WithCount(0, 1)
}

func mutatedCultistRemove(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	removed := 0
	for _, t := range ctx.LegalTargets() {
		switch t.Kind {
		case game.TargetCard:
			c, ok := g.LookupCardForEffect(t.ID)
			if !ok {
				continue
			}
			for _, n := range c.Counters {
				if n > 0 {
					removed += n
				}
			}
			if err := removeEveryCounterFrom(ctx, t.ID); err != nil {
				return err
			}
		case game.TargetPlayer:
			p := g.PlayerByIDForEffect(t.ID)
			if p == nil {
				continue
			}
			kinds := make([]string, 0, len(p.Counters))
			for kind, n := range p.Counters {
				if n > 0 {
					kinds = append(kinds, kind)
				}
			}
			sort.Strings(kinds)
			for _, kind := range kinds {
				n := p.Counters[kind]
				if err := g.AddPlayerCounterForEffect(t.ID, kind, -n); err != nil {
					return err
				}
				removed += n
			}
		}
	}
	if removed <= 0 {
		return nil
	}
	return GrantNextSpellPromise{From: "Mutated Cultist", Promise: game.NextSpellPromise{
		Reduce: removed,
		Text:   "The next spell you cast this turn costs {1} less to cast for each counter removed this way.",
	}}.Apply(ctx)
}
