package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lattice Library — Enchantment {X}{G}{G} (Secrets of Strixhaven
// Commander, #4834):
//
//	"This enchantment enters with X study counters on it.
//	 When this enchantment enters and whenever you cast your first spell
//	 with {X} in its mana cost each turn, create a 0/0 green and blue
//	 Fractal creature token. Put a number of +1/+1 counters on it equal to
//	 the number of study counters on this enchantment."
//
// One printed triggered ability with two trigger events, so one row
// watching both. "Your first spell with {X} in its mana cost each turn"
// is the turn's cast tally (CastTally.XSpells == 1, bumped before
// EventCast): a first X spell cast before the Library was on the
// battlefield, the Library itself included, uses the turn's trigger up
// (the ruling). The study counters are counted as the ability resolves,
// as the Library last existed if it has left (CR 608.2h); the Fractal is
// Deekah's create-then-put shape (socB8FractalWithCounters).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                   "0030578c-afaf-4072-a62b-2510bdceb64a",
		Name:                       "Lattice Library",
		Completeness:               effects.CompletenessFull,
		XMatters:                   true,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{effects.XCounters("study")},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventCast}, latticeLibraryTriggers,
				"Lattice Library — create a Fractal with a +1/+1 counter for each study counter",
				effects.SocB8FractalWithCounters("lattice-library/before-fractal", effects.CountersOnThisAsItResolves("study"))),
		},
	})
}

// latticeLibraryTriggers is "when this enchantment enters and whenever
// you cast your first spell with {X} in its mana cost each turn".
func latticeLibraryTriggers(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventETB:
		return ev.CardID == source.InstanceID
	case game.EventCast:
		return ev.Actor == source.Controller && g.CastTallyFor(source.Controller).XSpells == 1
	}
	return false
}
