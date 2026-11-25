package game

import "testing"

// protection_6253_test.go pins #6253's two qualities (CR 702.16a, 702.16b,
// 702.16f): "protection from creatures your opponents control" (Crypsis)
// and "protection from non-<type> creatures" (Spare from Evil).

func TestProtectionFromOpponentsCreaturesParsesAndBinds(t *testing.T) {
	q, ok := ParseProtectionQuality(ProtectionFromOpponentsCreatures)
	if !ok || q.Kind != ProtectionQualityOpponentsCreatures || q.Kind.String() != "opponents_creatures" {
		t.Fatalf("parsed %+v, %v", q, ok)
	}
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	hero := pushTypedTestCard(g, Card{Name: "Hero", TypeLine: "Creature — Human", Owner: me, Controller: me,
		Keywords: []string{ProtectionFromOpponentsCreatures}})
	theirBear := pushTypedTestCard(g, Card{Name: "Their Bear", TypeLine: "Creature — Bear", Owner: opp, Controller: opp})
	myBear := pushTypedTestCard(g, Card{Name: "My Bear", TypeLine: "Creature — Bear", Owner: me, Controller: me})
	theirRock := pushTypedTestCard(g, Card{Name: "Their Rock", TypeLine: "Artifact", Owner: opp, Controller: opp})

	h := layeredBattlefieldCard(t, g, hero)
	tb, mb, tr := layeredBattlefieldCard(t, g, theirBear), layeredBattlefieldCard(t, g, myBear), layeredBattlefieldCard(t, g, theirRock)
	if !ProtectedFrom(&h, SourceCharacteristics(&tb)) {
		t.Error("an opponent's creature is not stopped")
	}
	if ProtectedFrom(&h, SourceCharacteristics(&mb)) {
		t.Error("my own creature is stopped")
	}
	if ProtectedFrom(&h, SourceCharacteristics(&tr)) {
		t.Error("an opponent's noncreature artifact is stopped")
	}
	// Blocking (CR 702.16f): their creature can't block it, mine can.
	if r := blockRefusalFor(g, &h, &tb); r.Reason != BlockReasonProtection {
		t.Errorf("their creature blocking: refusal %+v, want protection", r)
	}
	if r := blockRefusalFor(g, &h, &mb); r.Reason == BlockReasonProtection {
		t.Error("my creature was refused as a blocker by protection")
	}
	// Targeting (CR 702.16b): an ability from their creature can't target
	// it, one from mine can.
	g.WithWriteLock(func() {
		if canBeTargetedBy(g, &h, ZoneBattlefield, SourceObject(opp, &tb)) {
			t.Error("their creature's ability targets it")
		}
		if !canBeTargetedBy(g, &h, ZoneBattlefield, SourceObject(me, &mb)) {
			t.Error("my creature's ability can't target it")
		}
	})
}

func TestProtectionFromNonSubtypeCreatures(t *testing.T) {
	tok := ProtectionFromNonSubtypeCreatures("Human")
	q, ok := ParseProtectionQuality(tok)
	if !ok || q.Kind != ProtectionQualityNonSubtypeCreatures || q.Value != "Human" || q.Kind.String() != "non_subtype_creatures" {
		t.Fatalf("parsed %+v, %v", q, ok)
	}
	if _, ok := ParseProtectionQuality("protection from non-Blorp creatures"); ok {
		t.Error("a quality naming no creature type parsed")
	}
	cases := []struct {
		name string
		src  Characteristic
		want bool
	}{
		{"non-Human creature", Characteristic{Types: []string{"Creature"}, Subtypes: []string{"Zombie"}}, true},
		{"Human creature", Characteristic{Types: []string{"Creature"}, Subtypes: []string{"Human", "Soldier"}}, false},
		{"changeling", Characteristic{Types: []string{"Creature"}, AllCreatureTypes: true}, false},
		{"noncreature", Characteristic{Types: []string{"Artifact"}}, false},
	}
	for _, c := range cases {
		src := c.src
		if got := q.Matches(&src); got != c.want {
			t.Errorf("%s: matches = %v, want %v", c.name, got, c.want)
		}
	}
	// Spare from Evil's acceptance: a non-Human can't block, a Human can.
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	att := pushTypedTestCard(g, Card{Name: "Attacker", TypeLine: "Creature — Bear", Owner: me, Controller: me, Keywords: []string{tok}})
	zombie := pushTypedTestCard(g, Card{Name: "Zombie", TypeLine: "Creature — Zombie", Owner: opp, Controller: opp})
	human := pushTypedTestCard(g, Card{Name: "Human", TypeLine: "Creature — Human Soldier", Owner: opp, Controller: opp})
	a, z, h := layeredBattlefieldCard(t, g, att), layeredBattlefieldCard(t, g, zombie), layeredBattlefieldCard(t, g, human)
	if r := blockRefusalFor(g, &a, &z); r.Reason != BlockReasonProtection {
		t.Errorf("a Zombie blocking: refusal %+v, want protection", r)
	}
	if r := blockRefusalFor(g, &a, &h); r.Reason == BlockReasonProtection {
		t.Error("a Human was refused as a blocker by protection")
	}
}

func blockRefusalFor(g *Game, attacker, blocker *Card) BlockRefusal {
	var r BlockRefusal
	g.ReadSnapshot(func() { r = g.BlockPairRefusalLocked(attacker, blocker) })
	return r
}
