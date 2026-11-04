package game

import (
	"testing"

	"github.com/google/uuid"
)

// permission_pt_test.go — #4964: "creature spells with power or toughness
// N or less" on a cast permission's filter, and the reading a lookup lends
// a card whose power and toughness its own ability defines (CR 604.3).

// The clause is an OR over the two numbers, and only over creatures.
func TestPermissionFilterPowerOrToughnessAtMost(t *testing.T) {
	f := PermissionFilter{CreatureOnly: true, PowerOrToughnessCapped: true, PowerOrToughnessAtMost: 1}
	creature := func(p, tt int) Card {
		return Card{TypeLine: "Creature — Bear", Power: p, Toughness: tt}
	}
	for _, tc := range []struct {
		name string
		card Card
		want bool
	}{
		{"1/1", creature(1, 1), true},
		{"2/1", creature(2, 1), true},
		{"1/2", creature(1, 2), true},
		{"2/2", creature(2, 2), false},
		{"0/9", creature(0, 9), true},
		{"9/0", creature(9, 0), true},
		{"3/3", creature(3, 3), false},
		{"instant", Card{TypeLine: "Instant"}, false},
		{"land", Card{TypeLine: "Basic Land — Forest"}, false},
		{"a 1/1 artifact that is not a creature", Card{TypeLine: "Artifact", Power: 1, Toughness: 1}, false},
	} {
		if got := f.Matches(tc.card); got != tc.want {
			t.Errorf("%s: Matches = %v, want %v", tc.name, got, tc.want)
		}
	}
	// Zero is a legal N, which is why the flag says the clause is there.
	zero := PermissionFilter{CreatureOnly: true, PowerOrToughnessCapped: true}
	if !zero.Matches(creature(5, 0)) || zero.Matches(creature(1, 1)) {
		t.Error(`"power or toughness 0 or less" admitted the wrong creatures`)
	}
	// And no clause at all admits a 6/6.
	if !(PermissionFilter{CreatureOnly: true}).Matches(creature(6, 6)) {
		t.Error("a filter without the clause stopped matching a 6/6")
	}
}

// A `*` the engine cannot compute is not "1 or less": unknown narrows the
// permission, never widens it.
func TestPermissionFilterPowerOrToughnessRefusesAnUncomputableStar(t *testing.T) {
	f := PermissionFilter{CreatureOnly: true, PowerOrToughnessCapped: true, PowerOrToughnessAtMost: 1}
	star := Card{TypeLine: "Creature — Lhurgoyf", VariableToughness: true}
	if f.Matches(star) {
		t.Error("a */* creature whose numbers the engine does not know matched")
	}
	// The importer's 0 stand-in is not a printed 0/0 either way round.
	if f.Matches(Card{TypeLine: "Creature — Lhurgoyf", Power: 4, VariableToughness: true}) {
		t.Error("a 4/* creature matched on its stand-in toughness")
	}
}

// withPTAnywhereLocked lends the reading a card's own characteristic-
// defining ability gives it where it sits (CR 604.3, CR 113.6a), and the
// filter reads that rather than the printed numbers. The ability here is
// Tarmogoyf's: power N, toughness N+1, for N card types among graveyards.
func TestPermissionFilterReadsALentCharacteristicDefiningPT(t *testing.T) {
	const key = "test-pt-cda-4964"
	prev := CatalogStaticAbilities
	t.Cleanup(func() { CatalogStaticAbilities = prev })
	types := 1
	CatalogStaticAbilities = func(k string) []StaticAbility {
		if k != key {
			return nil
		}
		return []StaticAbility{{
			Layer:    Layer7PT,
			SubLayer: SubLayer7A_CDA,
			Apply: func(ch *Characteristic, target *Card, g *Game, source *Card) {
				ch.Power, ch.Toughness = types, types+1
			},
		}}
	}

	g := NewGame()
	f := PermissionFilter{CreatureOnly: true, PowerOrToughnessCapped: true, PowerOrToughnessAtMost: 1}
	goyf := Card{
		InstanceID: uuid.New(), Name: "Test Goyf", TypeLine: "Creature — Lhurgoyf",
		OracleID: key, Power: 0, Toughness: 1, VariableToughness: true,
	}

	// Unlent, a card with such an ability is unknown and does not match:
	// a lookup that forgets to lend is narrow, never generous.
	if f.Matches(goyf) {
		t.Error("an unlent card with a power-defining ability matched")
	}

	var lent Card
	g.WithWriteLock(func() { lent = g.withPTAnywhereLocked(goyf) })
	if !f.Matches(lent) {
		t.Error("a lent 1/2 did not match `power or toughness 1 or less`")
	}
	types = 2
	g.WithWriteLock(func() { lent = g.withPTAnywhereLocked(goyf) })
	if f.Matches(lent) {
		t.Error("a lent 2/3 matched `power or toughness 1 or less`")
	}

	// A card with no such ability is never lent anything and reads its
	// printed numbers.
	bear := Card{InstanceID: uuid.New(), TypeLine: "Creature — Bear", Power: 2, Toughness: 2}
	var plain Card
	g.WithWriteLock(func() { plain = g.withPTAnywhereLocked(bear) })
	if plain.ptAnywhere != nil {
		t.Error("a card with no power-defining ability was lent a reading")
	}
	if f.Matches(plain) {
		t.Error("a 2/2 matched")
	}

	// And lendPTLocked lends only for a permission that asks the question.
	var untouched Card
	g.WithWriteLock(func() {
		untouched = g.lendPTLocked(&CastPermission{Filter: PermissionFilter{CreatureOnly: true}}, goyf)
	})
	if untouched.ptAnywhere != nil {
		t.Error("a permission with no power-or-toughness clause was lent a reading")
	}
}
