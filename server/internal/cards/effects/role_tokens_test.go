package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// role_tokens_test.go — #1945: Role tokens (CR 111.10, 303.7) and the
// CR 704.5z state-based action, through the real proof cards.

const (
	monstrousRageOracle  = "646a2371-54c0-4492-ac2f-20f109d6108c"
	royalTreatmentOracle = "fd0f8fdc-4065-41f4-b8fb-ecb8f185774d"
	cursedCourtierOracle = "17916cf9-1e6c-41bd-96ce-2b050d828838"
	spitefulHexmageOracl = "d07e0d88-8b7f-423d-97b7-cf7b80d8924c"
	faunsbaneTrollOracle = "b707c131-de13-4d4d-839d-b9f47d62f090"
)

// rolesOn lists the Role tokens on the battlefield attached to host.
func rolesOn(g *game.Game, host uuid.UUID) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.IsRole() && c.IsAttachedTo(host) {
				out = append(out, c)
			}
		}
	})
	return out
}

func castRoleSpell(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, name, "Instant", oracle, []game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
}

func TestMonsterRoleGivesPlusOneAndTrampleAndIsAnAuraToken(t *testing.T) {
	g := newCatalogGame(t)
	bear := pushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "Bear")

	castRoleSpell(t, g, "Monstrous Rage", monstrousRageOracle, bear)

	roles := rolesOn(g, bear)
	if len(roles) != 1 || roles[0].Name != "Monster Role" {
		t.Fatalf("roles on the bear = %+v, want one Monster Role", roles)
	}
	if !roles[0].IsToken() || !roles[0].IsAura() || roles[0].Controller != g.Seats[0].ID {
		t.Errorf("Monster Role = %+v, want a token Aura under the caster's control", roles[0])
	}
	// 2/2 +2/+0 until end of turn (Monstrous Rage) +1/+1 (Role).
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("bear is %d/%d, want 5/3", p, tough)
	}
	if !containsString(effectiveAbilities(t, g, bear), "trample") {
		t.Error("the enchanted creature lacks trample")
	}
}

// CR 704.5z: a second Role the same player puts on one creature sends
// the older one to the graveyard (a token, so it ceases to exist).
func TestSecondRoleReplacesTheOlderOneFromTheSamePlayer(t *testing.T) {
	g := newCatalogGame(t)
	bear := pushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "Bear")

	castRoleSpell(t, g, "Monstrous Rage", monstrousRageOracle, bear)
	castRoleSpell(t, g, "Royal Treatment", royalTreatmentOracle, bear)

	roles := rolesOn(g, bear)
	if len(roles) != 1 || roles[0].Name != "Royal Role" {
		t.Fatalf("roles on the bear = %+v, want only the newest (Royal Role)", roles)
	}
	if findBattlefieldByName(g, "Monster Role") != uuid.Nil {
		t.Error("the older Monster Role is still on the battlefield")
	}
	// 2/2 +1/+1 (Royal). Monstrous Rage's +2/+0 is until end of turn and still on.
	if p := effectivePower(t, g, bear); p != 5 {
		t.Errorf("bear power %d, want 5 (2 base +2 Rage +1 Royal)", p)
	}
	if containsString(effectiveAbilities(t, g, bear), "trample") {
		t.Error("the bear kept trample from the Role that left")
	}
}

// CR 704.5z counts Roles "controlled by the same player": two players'
// Roles on one creature both stay.
func TestRolesUnderDifferentControllersCoexist(t *testing.T) {
	g := newCatalogGame(t)
	bear := pushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "Bear")
	opp := g.Seats[1]

	theirs := RoleToken(RoleCursed)
	theirs.InstanceID = uuid.New()
	theirs.Owner, theirs.Controller = opp.ID, opp.ID
	theirs.AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: bear}
	theirs.AttachedAt = 1
	pushBattlefieldCardWithTimestamp(g, theirs)

	castRoleSpell(t, g, "Monstrous Rage", monstrousRageOracle, bear)

	if roles := rolesOn(g, bear); len(roles) != 2 {
		t.Fatalf("roles on the bear = %d, want 2 (one per controller)", len(roles))
	}
}

func TestCursedRoleMakesTheCreatureOneOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	courtier := castCatalogSpell(t, g, "Cursed Courtier", "Creature — Human Noble", cursedCourtierOracle, nil)
	passPriorityAroundTable(t, g)

	if roles := rolesOn(g, courtier); len(roles) != 1 || roles[0].Name != "Cursed Role" || roles[0].Controller != me.ID {
		t.Fatalf("roles on the Courtier = %+v, want its own Cursed Role", roles)
	}
	if p, tough := effectivePower(t, g, courtier), effectiveToughness(t, g, courtier); p != 1 || tough != 1 {
		t.Errorf("Courtier is %d/%d, want 1/1", p, tough)
	}
}

func TestSpitefulHexmageTargetsAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	hexmage := castCatalogSpell(t, g, "Spiteful Hexmage", "Creature — Human Warlock", spitefulHexmageOracl, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, bear)
		passPriorityAroundTable(t, g)
	}

	if roles := rolesOn(g, bear); len(roles) != 1 || roles[0].Name != "Cursed Role" {
		t.Fatalf("roles on the bear = %+v, want a Cursed Role", roles)
	}
	if roles := rolesOn(g, hexmage); len(roles) != 0 {
		t.Errorf("the Hexmage itself got %d Roles", len(roles))
	}
}

// A Role is not made for a host that is not a creature on the
// battlefield (here: a land), and it makes no drain for it.
func TestNoRoleForANonCreatureHost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})

	g.WithWriteLock(func() {
		if err := (CreateRoleToken{Role: RoleWicked, Host: land}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("CreateRoleToken: %v", err)
		}
	})
	if findBattlefieldByName(g, "Wicked Role") != uuid.Nil {
		t.Error("a Role was created attached to a land")
	}
}

// Wicked Role's own trigger: when it is put into a graveyard, each
// opponent loses 1 life. Killing the host puts it there (CR 704.5m).
func TestWickedRoleDrainsWhenItGoesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushCreatureToBattlefieldForTest(g, me.ID, "Bear")

	g.WithWriteLock(func() {
		if err := (CreateRoleToken{Role: RoleWicked, Host: bear}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("CreateRoleToken: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if p := effectivePower(t, g, bear); p != 3 {
		t.Fatalf("bear power %d, want 3 (+1/+0)", p)
	}
	before := []int{g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life, me.Life}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)

	for i := 1; i <= 3; i++ {
		if got := g.Seats[i].Life; got != before[i-1]-1 {
			t.Errorf("seat %d life %d, want %d", i, got, before[i-1]-1)
		}
	}
	if me.Life != before[3] {
		t.Errorf("the Role's controller lost life: %d -> %d", before[3], me.Life)
	}
}

func TestRoyalRoleGivesPlusOnePlusOne(t *testing.T) {
	g := newCatalogGame(t)
	bear := pushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "Bear")
	castRoleSpell(t, g, "Royal Treatment", royalTreatmentOracle, bear)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("bear is %d/%d, want 3/3", p, tough)
	}
	if !containsString(effectiveAbilities(t, g, bear), "hexproof") {
		t.Error("Royal Treatment's hexproof is missing")
	}
}

// --- Faunsbane Troll: the "Sacrifice an Aura attached to this
// creature" cost.

func pushTroll(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	troll := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Faunsbane Troll", OracleID: faunsbaneTrollOracle,
		TypeLine: "Creature — Troll", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if err := (CreateRoleToken{Role: RoleMonster, Host: troll}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("CreateRoleToken: %v", err)
		}
	})
	g.RunStateChecksForTest()
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	return troll
}

func TestFaunsbaneTrollSacrificesItsRoleToFightAndExile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	troll := pushTroll(t, g)
	role := rolesOn(g, troll)[0].InstanceID
	victim := pushBear(g, opp.ID, "Victim", 2)

	if err := g.ActivateCatalogAbility(me.ID, troll, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		SacrificeIDs: []uuid.UUID{role},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(role) {
		t.Error("the Role was not sacrificed")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the fought creature survived 4 damage")
	}
	if !graveyardOrExileHas(g, opp.ID, victim, game.ZoneExile) {
		t.Error("the fought creature did not go to exile instead of dying")
	}
}

// The cost names an Aura attached to THIS creature: a Role on another
// creature of yours is not a legal payment, and the ability is refused
// with nothing paid.
func TestFaunsbaneTrollCostRefusesAnAuraOnSomethingElse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	troll := pushTroll(t, g)
	other := pushCreatureToBattlefieldForTest(g, me.ID, "Bear")
	g.WithWriteLock(func() {
		if err := (CreateRoleToken{Role: RoleCursed, Host: other}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("CreateRoleToken: %v", err)
		}
	})
	g.RunStateChecksForTest()
	stray := rolesOn(g, other)[0].InstanceID
	victim := pushBear(g, opp.ID, "Victim", 2)

	err := g.ActivateCatalogAbility(me.ID, troll, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		SacrificeIDs: []uuid.UUID{stray},
	})
	if err == nil {
		t.Fatal("sacrificing an Aura attached to another creature paid the cost")
	}
	if !g.Battlefield.Contains(stray) {
		t.Error("the refused payment still sacrificed the Aura")
	}
}

// CR 704.5m's "not attached to anything" half applies to a Role token
// although a token has no catalog enchant clause.
func TestUnattachedRoleIsPutIntoTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	r := RoleToken(RoleMonster)
	r.InstanceID = uuid.New()
	r.Owner, r.Controller = me.ID, me.ID
	pushBattlefieldCardWithTimestamp(g, r)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(r.InstanceID) {
		t.Error("an unattached Role stayed on the battlefield")
	}
}

// graveyardOrExileHas reports whether `id` is in `zone` for the owner.
func graveyardOrExileHas(g *game.Game, owner, id uuid.UUID, zone game.ZoneKind) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == zone
}
