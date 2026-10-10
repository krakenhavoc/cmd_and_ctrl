package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// per_type_permission_test.go — #2167: the per-type budget's pure half
// (which types a card may spend) and the stored permission's spend.

func perTypeCard(typeLine string) Card {
	return Card{InstanceID: uuid.New(), Name: typeLine, TypeLine: typeLine}
}

func TestPermissionTypeChoicesReadTheCardAsPlayed(t *testing.T) {
	perm := &CastPermission{PerType: PermanentPermissionTypes}
	cases := []struct {
		typeLine string
		want     []string
	}{
		{"Creature — Bear", []string{"creature"}},
		{"Artifact Creature — Golem", []string{"artifact", "creature"}},
		// A land is played, so it spends "land" and nothing else.
		{"Artifact Land", []string{"land"}},
		{"Land Creature — Forest Dryad", []string{"land"}},
		{"Sorcery", nil},
		{"Kindred Enchantment — Elf", []string{"enchantment"}},
	}
	for _, c := range cases {
		if got := perm.PermissionTypeChoices(perTypeCard(c.typeLine)); !slices.Equal(got, c.want) {
			t.Errorf("%s: choices = %v, want %v", c.typeLine, got, c.want)
		}
	}
	nonland := &CastPermission{PerType: NonlandPermissionTypes, PerTypeUsed: []string{"creature"}}
	if got := nonland.PermissionTypeChoices(perTypeCard("Artifact Creature — Golem")); !slices.Equal(got, []string{"artifact"}) {
		t.Errorf("with creature spent: %v, want [artifact]", got)
	}
	if got := nonland.PermissionTypeChoices(perTypeCard("Artifact Land")); got != nil {
		t.Errorf("a land under a nonland budget: %v, want none", got)
	}
	if got := (&CastPermission{}).PermissionTypeChoices(perTypeCard("Creature")); got != nil {
		t.Errorf("no budget: %v, want nil", got)
	}
}

func TestPerTypeBudgetClosesCoversCard(t *testing.T) {
	bear := perTypeCard("Creature — Bear")
	perm := &CastPermission{Player: uuid.New(), Zone: ZoneGraveyard, Scope: ScopeStanding, PerType: PermanentPermissionTypes}
	if !perm.CoversCard(bear, ZoneGraveyard) {
		t.Fatal("an unspent creature budget does not cover a creature")
	}
	perm.PerTypeUsed = []string{"creature"}
	if perm.CoversCard(bear, ZoneGraveyard) {
		t.Fatal("a spent creature budget still covers a creature")
	}
}

// CR 712.11c: the face being cast is judged, so a creature whose
// Adventure is a sorcery is still open under a nonland budget once the
// creature use is spent.
func TestPerTypeBudgetAsksEveryCastableFace(t *testing.T) {
	c := Card{
		InstanceID: uuid.New(), Name: "Giant", TypeLine: "Creature — Giant", Layout: LayoutAdventure,
		Faces: []Face{
			{Name: "Giant", TypeLine: "Creature — Giant"},
			{Name: "Stomp", TypeLine: "Instant — Adventure"},
		},
	}
	perm := &CastPermission{Player: uuid.New(), Zone: ZoneExile, PerType: NonlandPermissionTypes,
		PerTypeUsed: []string{"creature"}, Cards: []PermissionCardRef{{ID: c.InstanceID}}}
	if !perm.CoversCard(c, ZoneExile) {
		t.Fatal("the Adventure half's instant use does not open the card")
	}
	perm.PerTypeUsed = append(perm.PerTypeUsed, "instant")
	if perm.CoversCard(c, ZoneExile) {
		t.Fatal("both halves spent, and the card is still open")
	}
}

func TestSettlePermissionType(t *testing.T) {
	grant := &CastPermission{PerType: PermanentPermissionTypes}
	golem := perTypeCard("Artifact Creature — Golem")
	if _, err := settlePermissionTypeLocked(grant, golem, ""); !errors.Is(err, ErrPermissionTypeRequired) {
		t.Errorf("two types, none named: %v", err)
	}
	if got, err := settlePermissionTypeLocked(grant, golem, "creature"); err != nil || got != "creature" {
		t.Errorf("named creature: %q, %v", got, err)
	}
	if got, err := settlePermissionTypeLocked(grant, perTypeCard("Creature"), ""); err != nil || got != "creature" {
		t.Errorf("one type, none named: %q, %v", got, err)
	}
	if _, err := settlePermissionTypeLocked(nil, golem, "artifact"); !errors.Is(err, ErrPermissionTypeNotOffered) {
		t.Errorf("a type with no budget: %v", err)
	}
	spent := &CastPermission{PerType: PermanentPermissionTypes, PerTypeUsed: []string{"artifact", "creature"}}
	if _, err := settlePermissionTypeLocked(spent, golem, ""); !errors.Is(err, ErrPermissionTypeSpent) {
		t.Errorf("both spent: %v", err)
	}
}

// A stored permission's spend is copy-on-write: a clone taken before it
// (an undo point) keeps the budget it had.
func TestStoredPerTypeSpendLeavesAnEarlierCloneAlone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := uuid.New()
	golem := perTypeCard("Artifact Creature — Golem")
	golem.Owner = me.ID
	g.Exile.PushTop(golem)
	g.mu.Lock()
	defer g.mu.Unlock()
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player: me.ID, Zone: ZoneExile, CastOnly: true, PerType: NonlandPermissionTypes,
		Source: src, Label: "test",
	}, []Card{golem})
	before := g.cloneLocked()
	grant := me.CastPermissions[0]
	g.spendPermissionTypeLocked(me.ID, golem, &grant, "creature")
	if got := me.CastPermissions[0].PerTypeUsed; !slices.Equal(got, []string{"creature"}) {
		t.Fatalf("spent = %v, want [creature]", got)
	}
	if got := before.Seats[0].CastPermissions[0].PerTypeUsed; len(got) != 0 {
		t.Fatalf("the earlier clone's budget changed: %v", got)
	}
}

// A derived permission's spend is per holder and per granting object.
func TestDerivedPerTypeSpendIsPerObject(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.mu.Lock()
	defer g.mu.Unlock()
	source := Card{InstanceID: uuid.New(), ObjectEpoch: 3}
	g.Battlefield.PushTop(source)
	grant := &CastPermission{Player: me.ID, Scope: ScopeStanding, Zone: ZoneGraveyard,
		PerType: PermanentPermissionTypes, Source: source.InstanceID}
	g.spendPermissionTypeLocked(me.ID, Card{}, grant, "creature")

	stamped := CastPermission{PerType: PermanentPermissionTypes}
	g.fillDerivedPerTypeUsedLocked(&stamped, me.ID, &source)
	if !slices.Equal(stamped.PerTypeUsed, []string{"creature"}) {
		t.Fatalf("same object: used = %v, want [creature]", stamped.PerTypeUsed)
	}
	newer := source
	newer.ObjectEpoch++
	g.fillDerivedPerTypeUsedLocked(&stamped, me.ID, &newer)
	if len(stamped.PerTypeUsed) != 0 {
		t.Fatalf("a new object: used = %v, want none", stamped.PerTypeUsed)
	}
	g.fillDerivedPerTypeUsedLocked(&stamped, g.Seats[1].ID, &source)
	if len(stamped.PerTypeUsed) != 0 {
		t.Fatalf("another holder: used = %v, want none", stamped.PerTypeUsed)
	}
}
