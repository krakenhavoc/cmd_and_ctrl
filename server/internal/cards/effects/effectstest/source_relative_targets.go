package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// source_relative_targets_test.go — #2146. A target described relative
// to its source ("with lesser power") is judged against the source's
// effective power at announce (the offered set and CR 601.2c) and again
// at resolution (CR 608.2b), and against its last-known information
// once the source has left (CR 608.2h).

const (
	RangersOfIthilienOracle  = "4d5fd9f0-ecb3-4b91-ae55-f02336d7cf37"
	HammerDropperOracle      = "088c9fa7-65ec-48c6-90cc-9087bc9df43e"
	BladeInstructorOracle    = "77cff683-013c-4568-8ae0-f86ef0590fe1"
	BargingSergeantOracle    = "af422a4b-087c-4753-9288-c31fb3457740"
	UnlivingPsychopathOracle = "038463c8-d2e1-4e7b-829d-67a744ae2660"
)

// --- mentor: the source dead, last-known information ------------------

// mentorAttack puts the mentor and its allies on the battlefield, declares
// them all as attackers and returns at the mentor's target prompt.
func MentorAttack(t *testing.T, g *game.Game, mentor uuid.UUID, allies ...uuid.UUID) *game.PendingChoice {
	t.Helper()
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	AdvanceToDeclareAttackersOf(t, g, seat)
	for _, id := range append([]uuid.UUID{mentor}, allies...) {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	LockInAttacks(t, g)
	prompt := LatestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("mentor asked for no target")
	}
	return prompt
}

func MentorGame(t *testing.T, name, oracle string, power int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Soldier", OracleID: oracle,
		Power: power, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	return g, id
}

func Ally(g *game.Game, name string, power int) uuid.UUID {
	me := g.Seats[g.Turn.ActiveSeat]
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Soldier", Power: power, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
}
