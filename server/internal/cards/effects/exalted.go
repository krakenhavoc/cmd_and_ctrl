package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Exalted (CR 702.90a) — "Whenever a creature you control attacks
// alone, that creature gets +1/+1 until end of turn." Ignoble
// Hierarch's and Noble Hierarch's shared trigger.
//
// Not in the closed PrintedKeywords / HasKeyword table
// (game/keywords.go): that table exists for keywords another
// permanent can GRANT as a bare string, and nothing in the catalog
// grants exalted to anything yet. It is a reminder-text triggered
// ability like any other, wired up directly on each source the way
// Bilbo's Ring wires up its own "attacks alone" trigger — the two
// share the "how many creatures attacked" read (attackedAlone here,
// attachedCreatureAttackedAlone in bilbos_ring.go, both CR 508.3's
// ruling: alone means the only creature in the whole declaration, not
// merely the only one this player controls).
//
// The pump is BoostUntilEOT — the ATTACKING creature's, not the
// exalted source's, since "attacks alone" can be any creature you
// control and the source itself is very often a mana dork that never
// attacks. Multiple exalted sources each fire and each add their own
// +1/+1 (CR 702.90b) — no dedup, because two calls of this
// constructor are two separate triggered abilities.
func Exalted() game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches: []game.EventKind{game.EventAttack},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return attackDeclaredByYou(ev, source.Controller) && attackedAlone(g)
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			attacker := ev.CardID
			return game.NewTriggeredItem(source, "Exalted — +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Target:    attacker,
						Power:     1,
						Toughness: 1,
						Label:     "Exalted — +1/+1",
					}.Apply(NewContext(g, item))
				})
		},
	}
}

// attackedAlone reports whether exactly one creature carries
// AttackingTarget right now — CR 508.3's "attacks alone", counted the
// same way attachedCreatureAttackedAlone (bilbos_ring.go) counts it
// for an equipped creature: DeclareAttackers stamps every attacker's
// AttackingTarget before announcing any of their EventAttack events,
// so a lone declaration reads as genuinely alone rather than "alone
// so far".
func attackedAlone(g *game.Game) bool {
	n := 0
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].AttackingTarget != uuid.Nil {
			n++
		}
	}
	return n == 1
}
