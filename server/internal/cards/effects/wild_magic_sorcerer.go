package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wild-Magic Sorcerer — Creature — Orc Shaman Sorcerer {3}{R}, 4/3
// (Adventures in the Forgotten Realms Commander):
//
//	"The first spell you cast from exile each turn has cascade."
//
// Maelstrom Nexus's shape (a battlefield trigger on your cast, so cascade is
// the real keyword's effect and its mana-value cap is read off the spell) with
// the per-turn clock of CastTally.FromExile: the tally is bumped as the spell
// becomes cast, before EventCast, so a count of exactly one on a spell cast
// from exile (the event's OldZone) means "this is the first". A spell cast
// from your hand does not use the clock up, and a second spell from exile is
// skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a450ef56-8511-43e0-bdca-0c25f54ba338",
		Name:         "Wild-Magic Sorcerer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			s82AfrB16GrantsCascadeWhen("Wild-Magic Sorcerer",
				func(ev game.Event, source *game.Card, g *game.Game) bool {
					return ev.OldZone == game.ZoneExile && g.SpellsCastThisTurn[source.Controller].FromExile == 1
				}),
		},
	})
}
