package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_triggers_test.go — S22: "whenever ~ attacks" triggers on
// the new EventAttack. Seat 0 is the active player throughout; its
// creatures attack seat 1 (and seat 2 for the multi-defender case).
//
// Attack triggers land during the declare-attackers step, so these
// tests stop at StepDeclareAttackers rather than walking on to
// combat damage the way the S19 sub-PR 7 tests do — the whole point
// of the event is that the trigger is answerable before blockers.

const (
	HellriderOracle   = "f02557b0-f422-48cb-875e-2c814c39967d"
	SunTitanOracle    = "b2e950fb-cb7e-40a0-a311-5bbdd0477b29"
	KrenkoKingpinOrcl = "e8065e1d-e937-4b56-8011-78f0d07328a0"
)

// declareAttack advances to the declare-attackers step, declares each
// creature against defender, and locks the declaration in — the
// priority wrap where the engine announces it and harvests the attack
// triggers off the final assignment (#859). The cursor stays in the
// step whenever the declaration produced a trigger; a declaration
// with none wraps on, exactly as an uneventful attack does.
func DeclareAttack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker %s: %v", id, err)
		}
	}
	LockInAttacks(t, g)
}

// triggersOnStackFrom counts the triggered-ability items in
// StackMeta sourced from cardID. Attack triggers fire once per
// attacking creature, so the count is the assertion.
func TriggersOnStackFrom(g *game.Game, cardID uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == cardID {
			n++
		}
	}
	return n
}

// pushGraveyardPermanent seeds a permanent card with a real mana
// cost into a player's graveyard — Sun Titan's mana-value clause
// needs a parseable cost, which pushGraveyardCardForTest does not
// set.
func PushGraveyardPermanent(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		ManaCost:   manaCost,
		Power:      2,
		Toughness:  2,
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}
