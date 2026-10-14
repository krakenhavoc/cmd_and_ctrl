package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sin, Unending Cataclysm — Legendary Creature {5}{G}{U}, 5/5:
//
//	"Flying, trample
//	 As Sin enters, remove all counters from any number of artifacts,
//	 creatures, and enchantments. Sin enters with X +1/+1 counters on
//	 it, where X is twice the number of counters removed this way.
//	 When Sin dies, put its counters on target creature you control,
//	 then shuffle this card into its owner's library."
//
// From the Tidus, Yuna's Guardian deck request (#3246); the seam is
// #1893, ADR 0170.
//
// The as-enters sentence is CR 614.1c's replacement with a choice of
// other permanents inside it (AsEntersChoosePermanents, AnyController):
// any number of artifacts, creatures and enchantments, whoever controls
// them, chosen while Sin is not yet on the battlefield (CR 614.12, so it
// can never name itself, CR 614.13a). Every counter of every kind comes
// off each one inline, and X counts what really came off, read back
// from the board: a counter that can't be removed (#1824) stays and is
// not counted. The +1/+1 counters ride the entry
// (ReplacementEvent.AddCounterAtETB), so Sin enters with them, and a
// Hardened Scales or Doubling Season applies to them as it does to any
// "enters with" counters (CR 614.1c).
//
// Only permanents with a counter on them are offered. Choosing one with
// none is legal but does nothing, so leaving it off the list changes no
// outcome.
//
// The dies trigger reads "its counters" from Sin's last-known
// information (ctx.TriggeringPermanent, CR 603.10), every kind, as
// Resourceful Defense does. The shuffle follows the card only while it
// is still in its owner's graveyard: a Sin that left the graveyard in
// response is a new object (CR 400.7), and the ability does not find
// it. With no creature to target the ability is removed from the stack
// (CR 603.3d), and with its only target gone it does nothing at all
// (CR 608.2b), so in neither case is Sin shuffled away.
//
// No simplification.
func init() {
	const name = "Sin, Unending Cataclysm"
	game.RegisterEntryChoiceSuggestion(name, sinSuggest)
	Register(Spec{
		OracleID:        "48534bbf-06b6-46a0-8a86-b14f38caee58",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Replacements: []game.ReplacementEffect{
			AsEntersChoosePermanents(name,
				"Sin — as it enters, remove all counters from any number of artifacts, creatures, and enchantments. It enters with twice that many +1/+1 counters.",
				&game.EntryCardChoice{
					Action:        game.EntryCardChoose,
					AnyController: true,
					AnyNumber:     true,
					Matches:       sinCandidate,
					Then:          sinRemoveAllCounters,
					SuggestKey:    name,
				}),
		},
		Triggered: []game.TriggeredAbility{
			withTargets(
				WhenThisDies("Sin — put its counters on target creature you control, then shuffle it into its owner's library", sinDies),
				TargetCreature("target creature you control", YouControl()),
			),
		},
	})
}

// withTargets is a triggered ability with a target clause.
func withTargets(t game.TriggeredAbility, spec *game.TargetSpec) game.TriggeredAbility {
	t.Targets = spec
	return t
}

// sinCandidate is "an artifact, creature, or enchantment" with a counter
// on it.
func sinCandidate(c game.Card) bool {
	if !c.IsArtifact() && !c.IsCreature() && !c.IsEnchantment() {
		return false
	}
	for _, n := range c.Counters {
		if n > 0 {
			return true
		}
	}
	return false
}

// sinRemoveAllCounters is the entry's Then: every counter off every
// permanent named, and twice the number that really came off onto Sin
// as it enters.
func sinRemoveAllCounters(g *game.Game, ev *game.ReplacementEvent, picked []uuid.UUID) error {
	removed := 0
	for _, id := range picked {
		before, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		kinds := make([]string, 0, len(before.Counters))
		was := 0
		for kind, n := range before.Counters {
			if n > 0 {
				kinds = append(kinds, kind)
				was += n
			}
		}
		sort.Strings(kinds)
		for _, kind := range kinds {
			if _, err := g.AddCounterMustSettleNowForEffect(id, kind, -before.Counters[kind]); err != nil {
				return err
			}
		}
		now := 0
		if after, ok := g.LookupCardForEffect(id); ok {
			for _, n := range after.Counters {
				if n > 0 {
					now += n
				}
			}
		}
		if was > now {
			removed += was - now
		}
	}
	if removed > 0 {
		ev.AddCounterAtETB(game.CounterPlusOne, 2*removed)
	}
	return nil
}

// sinSuggest is the answer a seat nearly always wants: every candidate
// an opponent controls (their counters go, and Sin grows by two for
// each), and those of its own whose counters all hurt it (a -1/-1 or a
// stun counter: removing them is a gain twice over). A permanent of its
// own with a counter it wants is left alone.
func sinSuggest(g *game.Game, chooser uuid.UUID, candidates []uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range cardsStillHere(g, candidates) {
		if c.Controller != chooser || allCountersHarmful(c.Counters) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// allCountersHarmful reports a permanent whose every counter is one its
// controller would rather not have.
func allCountersHarmful(counters map[string]int) bool {
	seen := false
	for kind, n := range counters {
		if n <= 0 {
			continue
		}
		seen = true
		if !game.HarmfulCardCounter(kind) {
			return false
		}
	}
	return seen
}

// sinDies is the dies trigger's resolution: Sin's last-known counters
// onto the target, then Sin from its owner's graveyard into its owner's
// library, shuffled.
func sinDies(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	if left, ok := ctx.TriggeringPermanent(); ok && len(left.Counters) > 0 {
		for _, t := range targets {
			if err := b14PutCounters(ctx, t.ID, left.Counters); err != nil {
				return err
			}
		}
	}
	id := item.SourceCardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	owner := c.Owner
	return g.TuckToLibraryThenForEffect(id, game.TuckOptions{}, func(g *game.Game, _ bool) error {
		return g.ShuffleLibraryForEffect(owner)
	})
}
