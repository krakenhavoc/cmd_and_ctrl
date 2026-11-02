package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Storm, Force of Nature — Legendary Creature — Mutant Hero {1}{G}{U}{R},
// 3/4:
//
//	"Flying, vigilance
//	 Ceaseless Tempest — Whenever Storm deals combat damage to a player,
//	 the next instant or sorcery spell you cast this turn has storm. (When
//	 you cast it, copy it for each spell cast before it this turn. You may
//	 choose new targets for the copies.)"
//
// The trigger sets up CR 603.7b's "when you next cast" watch, the one
// Doublecast uses: it waits for the controller's next instant or sorcery
// cast this turn, fires on that one only (a creature spell leaves it
// waiting) and ends at cleanup if none comes. A second connection in the
// same turn (a second combat) sets up a second watch, and the spell is
// then stormed once for each, since a spell may have storm twice
// (CR 702.40b).
//
// The spell gains the keyword's effect by way of the trigger the watch
// puts on the stack as it is cast, above the spell: its resolution is
// storm's (stormCopies), copying the spell for each spell cast before it
// from the recorded cast order, with new targets offered, from last-known
// information if the spell was countered in response (CR 608.2h). The
// copies are not cast, so they do not storm again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7c47876d-143f-4e4c-b151-90c099e452a1",
		Name:            "Storm, Force of Nature",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Storm, Force of Nature — the next instant or sorcery spell you cast this turn has storm",
				func(g *game.Game, item *game.StackItem) error {
					return WhenYouNextCast(
						"Storm, Force of Nature — storm: copy that spell for each spell cast before it",
						game.CastFilter{Types: []string{"Instant", "Sorcery"}},
						stormTheSpellBody,
					).Apply(NewContext(g, item))
				}),
		},
	})
}

// stormTheSpellBody is the delayed trigger's body: storm (CR 702.40a) for
// the spell the watch fired on, which rides the item as its payload.
var stormTheSpellBody = game.SimpleDelayedBody("storm/the-spell-just-cast", stormTheSpellYouJustCast)

func stormTheSpellYouJustCast(g *game.Game, item *game.StackItem) error {
	cast := NewContext(g, item).PayloadCards()
	if len(cast) == 0 {
		return nil
	}
	return stormCopies(g, item, cast[0])
}
