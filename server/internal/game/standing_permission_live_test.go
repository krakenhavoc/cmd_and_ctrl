package game

import (
	"testing"

	"github.com/google/uuid"
)

// standing_permission_live_test.go — #4458: the live parts of a standing
// permission (ExiledWithSourceThisTurn, MaxSpellManaValueCountOf).

func TestPermissionFilterExiledWithSourceThisTurnReadsTheLinkTurn(t *testing.T) {
	src := PermissionCardRef{ID: uuid.New(), Epoch: 1}
	f := PermissionFilter{ExiledWithSource: true, ExiledWith: src, ExiledWithSourceThisTurn: true, ExiledOnTurn: 7}
	if !f.Matches(Card{ExiledWith: src, ExiledWithTurn: 7}) {
		t.Error("a card linked this turn must match")
	}
	if f.Matches(Card{ExiledWith: src, ExiledWithTurn: 6}) {
		t.Error("a card linked on an earlier turn must not match")
	}
	bare := f
	bare.ExiledOnTurn = 0
	if bare.Matches(Card{ExiledWith: src}) {
		t.Error("the flag with no derived turn matches nothing")
	}
	if !f.faceBlind() {
		t.Error("the clause is about the object, not its face")
	}
}

func TestStampExiledWithRecordsTheTurn(t *testing.T) {
	ref := PermissionCardRef{ID: uuid.New(), Epoch: 1}
	id := uuid.New()
	exile := &Zone{Kind: ZoneExile, Cards: []Card{{InstanceID: id}}}
	stampExiledWithLocked(exile, id, ref, AbilityRef{}, 4)
	if exile.Cards[0].ExiledWithTurn != 4 {
		t.Errorf("ExiledWithTurn = %d, want 4", exile.Cards[0].ExiledWithTurn)
	}
}

func TestFillDerivedLiveCountsElvesAndFaeriesOnce(t *testing.T) {
	g := &Game{Battlefield: &Zone{Kind: ZoneBattlefield}}
	g.Turn.Seq = 3
	me, them := uuid.New(), uuid.New()
	g.Battlefield.Cards = append(g.Battlefield.Cards, []Card{
		{InstanceID: uuid.New(), Controller: me, TypeLine: "Creature — Elf Faerie"},
		{InstanceID: uuid.New(), Controller: me, TypeLine: "Creature — Elf"},
		{InstanceID: uuid.New(), Controller: me, TypeLine: "Kindred Enchantment — Faerie"},
		{InstanceID: uuid.New(), Controller: me, TypeLine: "Creature — Goblin"},
		{InstanceID: uuid.New(), Controller: them, TypeLine: "Creature — Elf"},
	}...)
	perm := CastPermission{
		Filter:                   PermissionFilter{ExiledWithSource: true, ExiledWithSourceThisTurn: true},
		MaxSpellManaValueCountOf: []string{"Elf", "Faerie"},
	}
	g.fillDerivedLiveLocked(&perm, me)
	if perm.MaxSpellManaValue == nil || *perm.MaxSpellManaValue != 3 {
		t.Errorf("cap = %v, want 3 (each permanent once, yours only)", perm.MaxSpellManaValue)
	}
	if perm.Filter.ExiledOnTurn != 3 {
		t.Errorf("ExiledOnTurn = %d, want the current turn 3", perm.Filter.ExiledOnTurn)
	}
	plain := CastPermission{}
	g.fillDerivedLiveLocked(&plain, me)
	if plain.MaxSpellManaValue != nil || plain.Filter.ExiledOnTurn != 0 {
		t.Error("a permission printing neither clause is untouched")
	}
}
