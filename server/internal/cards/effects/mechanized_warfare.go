package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mechanized Warfare — Enchantment {1}{R}{R} (EDHREC rank 4210):
//
//	"If a red or artifact source you control would deal damage to an
//	 opponent or a permanent an opponent controls, it deals that much
//	 damage plus 1 instead."
//
// Gratuitous Violence's one-sided cousin, and much wider than it
// looks: "source you control" is every red or artifact source, not
// only creatures, so a Shock does four instead of three, every red
// creature in an attack adds a point, and a Thopter token adds one on
// each of its swings. In a go-wide red deck it is the difference
// between a board stall and lethal.
//
// A CR 614 replacement, and three of its four clauses are the whole
// card:
//
//   - "red or artifact SOURCE YOU CONTROL" — read off the source's
//     post-layer characteristics as the damage event snapshotted them
//     (as it last existed, for a permanent that has already left —
//     #1417, CR 608.2h), so an artifact creature
//     an effect has made blue still qualifies (it is an artifact) and
//     a creature an effect has made red qualifies too. A source you
//     do not control never does, which is what keeps the enchantment
//     one-sided.
//   - "to an OPPONENT or a permanent an opponent controls" — damage
//     to you, to your own creatures, and to a permanent you control is
//     untouched. Combat damage your blocker deals to an attacking
//     creature its controller owns IS boosted, because that attacker
//     is a permanent an opponent controls.
//   - "plus 1", once. Two Mechanized Warfares are +2 in either CR 616
//     order, which falls out of the pipeline rather than needing a
//     guard.
//
// A source outside any zone the engine can look up — a spell that has
// already left the stack — cannot be tested for colour, so it is not
// boosted. That errs weaker and never stronger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c912a43b-8994-434e-84c0-f4cf58abbd42",
		Name:         "Mechanized Warfare",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || src == nil {
					return false
				}
				return b40RedOrArtifactSourceControlledBy(ev, g, src.Controller) &&
					b40DamageHitsAnOpponentOf(ev, g, src.Controller)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.DamageAmount++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Mechanized Warfare: +1 damage",
		}},
	})
}
