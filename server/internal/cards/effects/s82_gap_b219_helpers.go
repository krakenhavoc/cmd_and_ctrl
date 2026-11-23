package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b219_helpers.go — helpers for GAP batch 219 (#5948), Commander
// Legends: Battle for Baldur's Gate red creatures.

// GapB219GoadedCreatureOfAnOpponent is the blocker scope of "Goaded creatures your
// opponents control" (Bothersome Quasit): a creature on the battlefield that a
// player other than the rule source's controller controls and that is goaded as
// the block is declared, by a resolved goad or a continuous one (CR 701.15).
func GapB219GoadedCreatureOfAnOpponent(_ *game.Game, c, source *game.Card) bool {
	return source != nil && c.IsCreature() && c.Controller != source.Controller && c.Goaded()
}

// GapB219LandsDefendingPlayerControls counts the lands controlled by the
// defending player (CR 508.5) of the creature `attacker` is attacking, zero when
// it isn't attacking: Elturel Survivors' X.
func GapB219LandsDefendingPlayerControls(g *game.Game, attacker *game.Card) int {
	if attacker == nil || attacker.AttackingTarget == uuid.Nil {
		return 0
	}
	defender := g.DefendingPlayerForAttackForEffect(attacker.AttackingTarget)
	if defender == uuid.Nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == defender && c.HasCardType("land") {
			n++
		}
	}
	return n
}

// GapB219OpponentsDealtDamageThisTurn counts the opponents of `you` who have been
// dealt damage this turn, combat or not, by any source (the turn tally's
// per-player damage taken): Gnoll War Band's discount.
func GapB219OpponentsDealtDamageThisTurn(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, seat := range g.Seats {
		if seat == nil || seat.Eliminated || seat.ID == you {
			continue
		}
		if g.TurnTallyFor(seat.ID).DamageTaken > 0 {
			n++
		}
	}
	return n
}

// GapB219NamedCard is the predicate for "creatures you control named <name>"'s name
// half: the card's effective name, compared case-folded (CR 201.2).
func GapB219NamedCard(name string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return strings.EqualFold(c.Effective().Name, name)
	}
}
