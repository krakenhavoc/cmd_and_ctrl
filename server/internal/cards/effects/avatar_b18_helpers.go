package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b18_helpers.go — shared pieces of Avatar: The Last Airbender
// batches 18 and 19 (#4459). Append-only.

// avB18LessonSpell is a cost predicate for "Lesson spells" (Uncle Iroh):
// Lesson is a spell subtype on the printed type line ("Sorcery —
// Lesson"), read off the spell as it is priced.
func avB18LessonSpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.HasSubtype("Lesson") }
}

// avB18YouSacrificedAnother is "whenever you sacrifice another
// permanent" (Sandbender Scavengers, Zhao): Pirate Peddlers' reading, an
// EventSacrifice with the source's controller as the Actor, one per
// permanent, never the source itself.
func avB18YouSacrificedAnother(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Actor == source.Controller && ev.CardID != uuid.Nil && ev.CardID != source.InstanceID
}

// avB18ThisVehicleDied is "when this Vehicle dies" (Fire Nation
// Warship): "dies" is a creature put into a graveyard from the
// battlefield (CR 700.4), so the Vehicle triggers only when it was a
// creature as it last existed there (CR 603.10a), crewed. A Vehicle
// destroyed while it was not crewed does not die.
func avB18ThisVehicleDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !cardDied(ev, source) {
		return false
	}
	_, ok := diedCreature(ev, g)
	return ok
}

// avB18CanAttackAsThoughNoDefenderWhile is "as long as <cond>, this
// creature can attack as though it didn't have defender" (Platypus-Bear,
// Starfleet Spanner Jockey): a self-only layer 6 static writing an
// unrestricted DefenderWaiver (ADR 0106's 2026-10-16 amendment, #1936)
// while cond holds for the source, read live at declaration. The keyword
// stays (CR 702.3b is waived, not removed).
func avB18CanAttackAsThoughNoDefenderWhile(label string, cond func(g *game.Game, source *game.Card) bool) game.StaticAbility {
	return game.StaticAbility{
		Label: label,
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && selfOnly(target, g, source) && cond(g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			w := game.DefenderWaiver{}
			if source != nil {
				w.Source, w.SourceName = source.InstanceID, source.Name
			}
			c.DefenderWaivers = append(c.DefenderWaivers, w)
		},
	}
}
