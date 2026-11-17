package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Viral Spawning — Sorcery {2}{G} (Phyrexia: All Will Be One, #5479):
//
//	"Create a 3/3 green Phyrexian Beast creature token with toxic 1.
//	 (Players dealt combat damage by it also get a poison counter.)
//	 Corrupted — As long as an opponent has three or more poison counters
//	 and this card is in your graveyard, it has flashback {2}{G}. (You may
//	 cast this card from your graveyard for its flashback cost. Then exile
//	 it.)"
//
// The flashback is the card's own alternative cost (effects.Flashback),
// with its offer gated on corrupted (seedcoreCorrupted: an opponent has
// three or more poison counters) as the caster would cast it, so it is
// offered while the condition holds and withdrawn when it stops. The cast
// path is the ordinary one from the graveyard, and flashback's exile on
// leaving the stack is the keyword's (CR 702.34a).
//
// No simplification.
func init() {
	flashback := Flashback("{2}{G}")
	flashback.Condition = func(g *game.Game, controller uuid.UUID) bool {
		return seedcoreCorrupted(g, controller, uuid.Nil)
	}
	Register(Spec{
		OracleID:         "9945c7eb-97a9-42df-a051-1ea514973280",
		Name:             "Viral Spawning",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 1, CreatureTokens: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{flashback},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return s82OneB20CreateBeasts(ctx.Game, item, 1)
		},
	})
}
