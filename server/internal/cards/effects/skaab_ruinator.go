package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skaab Ruinator — Creature — Zombie Horror {1}{U}{U}, 5/6 (Duskmourn
// Commander):
//
//	"As an additional cost to cast this spell, exile three creature cards
//	 from your graveyard.
//	 Flying
//	 You may cast this card from your graveyard."
//
// The additional cost is the mandatory exile-cards slot of a cast
// (ExileFromGraveyardCost, Abhorrent Oculus' shape): the caster names three
// creature cards in their own graveyard, paid with the spell already on the
// stack. The spell itself is never one of them, which matters when it is cast
// from that same graveyard. The permission to cast it from the graveyard is
// CastableZones (CR 601.2a), and the cost is charged wherever it is cast
// from. Flying is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "83734824-1cad-4649-87ae-a5c338771c4c",
		Name:            "Skaab Ruinator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AdditionalCost:  ExileFromGraveyardCost(3, "three creature cards", func(c game.Card) bool { return c.IsCreature() }),
		CastableZones:   []game.ZoneKind{game.ZoneGraveyard},
	})
}
