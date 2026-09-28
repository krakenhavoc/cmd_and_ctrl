package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// combat_death_triggers_test.go — #1661: "whenever an attacking /
// blocking creature dies" reads the dying creature's combat state off
// its leaves-the-battlefield event (CR 603.10a), because the exit has
// already cleared it from the card. Kardur, Doomscourge (the caveat
// this cleared), Death Tyrant and Ares, God of War (new), and Garna,
// Bloodfist of Keld (moved off an event-log walk onto the same read).

const (
	deathTyrantOracle  = "fa944633-e078-41bb-9c7d-4f6b669c27a5"
	aresGodOfWarOracle = "8a09d4a2-4796-4a15-ab75-a3a0b717f62d"
)

// lifeSnapshot is every seat's life total, by seat index.
func lifeSnapshot(g *game.Game) []int {
	out := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = p.Life
	}
	return out
}

// assertKardurDrains checks that every opponent of seat 0 lost `n` and
// seat 0 gained `n` since `before`.
func assertKardurDrains(t *testing.T, g *game.Game, before []int, n int) {
	t.Helper()
	for i, p := range g.Seats {
		want := before[i] - n
		if i == 0 {
			want = before[i] + n
		}
		if p.Life != want {
			t.Errorf("seat %d life %d → %d, want %d (%d Kardur drain(s))", i, before[i], p.Life, want, n)
		}
	}
}

// pushKardur seats Kardur for seat 0 without casting him, so his ETB
// requirement does not shape the combat under test.
func pushKardur(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Kardur, Doomscourge", "Legendary Creature — Demon Berserker", kardurDoomscourgeOracle, 4, 3)
}

func TestKardurDrainsWhenAnAttackerDiesToCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushKardur(g, me.ID)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 3, 4)
	before := lifeSnapshot(g)

	declareAttack(t, g, opp.ID, bear)
	b35Block(t, g, wall, bear)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Fatal("setup: the Bear should have died to the Wall")
	}
	assertKardurDrains(t, g, before, 1)
}

// The death happens after blockers, to an instant, not to combat
// damage: the attack is still read off the event.
func TestKardurDrainsWhenAnAttackerIsDestroyedByAnInstantAfterBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushKardur(g, me.ID)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	before := lifeSnapshot(g)

	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	blade := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: blade, Name: "Doom Blade", TypeLine: "Instant",
		OracleID: doomBladeOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, blade, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("CastSpell(Doom Blade): %v", err)
	}
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(bear) {
		t.Fatal("setup: Doom Blade did not kill the attacker")
	}
	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("setup: the kill happened in %s, want declare_blockers", g.Turn.Step)
	}
	assertKardurDrains(t, g, before, 1)
}

// CR 603.10a: Kardur's trigger looks back in time, so a wipe that
// kills him alongside two attackers — one of them Kardur himself —
// still drains once per attacking creature.
func TestKardurSeesAttackersDieInTheSameWipeAsHim(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kardur := pushKardur(g, me.ID)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "My Idle Elf", "Creature — Elf", 1, 1)
	before := lifeSnapshot(g)

	declareAttack(t, g, opp.ID, kardur, bear)
	g.WithWriteLock(func() {
		if n := g.DestroyPermanentsForEffect([]uuid.UUID{kardur, bear, idle}); n != 3 {
			t.Fatalf("wipe destroyed %d, want 3", n)
		}
	})
	b30SettleOrder(t, g)

	assertKardurDrains(t, g, before, 2)
}

// A death that is not an attacking creature's is silent — outside
// combat, and a creature that sat out the combat it died during.
func TestKardurIgnoresANonAttackingCreatureDying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushKardur(g, me.ID)
	idle := b12Creature(g, me.ID, "My Idle Elf", "Creature — Elf", 1, 1)
	outside := b12Creature(g, me.ID, "My Other Elf", "Creature — Elf", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	before := lifeSnapshot(g)

	b18Kill(t, g, outside)
	assertKardurDrains(t, g, before, 0)

	declareAttack(t, g, opp.ID, bear)
	b18Kill(t, g, idle)
	assertKardurDrains(t, g, before, 0)
}

// The mirror of an attacker dying: an opponent attacks, the blocker
// that dies is NOT an attacking creature, so Kardur is silent for it.
// Death Tyrant's two clauses are crossed the same way — MY blocker is
// not "a blocking creature an opponent controls" and THEIR attacker is
// not "an attacking creature you control" — so it makes nothing, while
// Kardur does drain for their attacker.
func TestKardurAndDeathTyrantOnAnOpponentsAttackTradingWithMyBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushKardur(g, me.ID)
	b12Push(g, me.ID, "Death Tyrant", "Creature — Beholder Skeleton", deathTyrantOracle, 4, 6)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(theirs, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	before := lifeSnapshot(g)
	b35Block(t, g, mine, theirs)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Fatal("setup: the two Bears should have traded")
	}
	// One drain, for their attacker. My blocker's death is not one.
	assertKardurDrains(t, g, before, 1)
	if n := countTokensControlled(g, me.ID, "Zombie"); n != 0 {
		t.Errorf("Death Tyrant made %d Zombies for crossed clauses, want 0", n)
	}
}

// Death Tyrant's two clauses, each on its own: my attacker dying, and
// an opponent's blocker dying. Kardur is silent for the blocker.
func TestDeathTyrantMakesAZombieForMyAttackerAndTheirBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushKardur(g, me.ID)
	b12Push(g, me.ID, "Death Tyrant", "Creature — Beholder Skeleton", deathTyrantOracle, 4, 6)
	big := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 3, 3)
	small := b12Creature(g, me.ID, "My Squire", "Creature — Human", 1, 1)
	chump := b12Creature(g, opp.ID, "Their Chump", "Creature — Human", 1, 1)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 2, 4)
	before := lifeSnapshot(g)

	declareAttack(t, g, opp.ID, big, small)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(chump, big); err != nil {
		t.Fatalf("DeclareBlocker(chump): %v", err)
	}
	if err := g.DeclareBlocker(wall, small); err != nil {
		t.Fatalf("DeclareBlocker(wall): %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	// Kardur's drain and two Negative Energy Cones trigger together
	// for one controller: a CR 603.3b order prompt first.
	b30SettleOrder(t, g)

	if g.Battlefield.Contains(chump) {
		t.Fatal("setup: their chump blocker should have died")
	}
	if g.Battlefield.Contains(small) {
		t.Fatal("setup: my Squire should have died to the Wall")
	}
	if n := countTokensControlled(g, me.ID, "Zombie"); n != 2 {
		t.Errorf("Death Tyrant made %d Zombies, want 2 — one for my attacker, one for their blocker", n)
	}
	// Kardur: my Squire was attacking (drain), their chump was blocking (none).
	assertKardurDrains(t, g, before, 1)
}

// CR 506.4: a creature removed from combat is not an attacking
// creature when it later dies. A control change is the engine's
// removal-from-combat route.
func TestKardurIgnoresAnAttackerRemovedFromCombatBeforeItDied(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushKardur(g, me.ID)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	before := lifeSnapshot(g)

	declareAttack(t, g, opp.ID, bear)
	stealForTest(t, g, bear, opp.ID)
	if c, _ := g.LookupCardForEffect(bear); c.AttackingTarget != uuid.Nil {
		t.Fatal("setup: the steal did not remove the Bear from combat")
	}
	b18Kill(t, g, bear)
	if g.Battlefield.Contains(bear) {
		t.Fatal("setup: the Bear should have died")
	}
	assertKardurDrains(t, g, before, 0)
}

// Ares returns an attacking creature YOU control to its owner's hand;
// a non-attacker's death, and an opponent's attacker's, stay put.
func TestAresReturnsYourDeadAttackerToHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ares := b12Push(g, me.ID, "Ares, God of War", "Legendary Creature — God Warrior Villain", aresGodOfWarOracle, 4, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "My Idle Elf", "Creature — Elf", 1, 1)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 3, 4)

	declareAttack(t, g, opp.ID, ares, bear)
	b35Block(t, g, wall, bear)
	b18Kill(t, g, idle)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(bear) {
		t.Error("the Bear died attacking and should be back in its owner's hand")
	}
	if !me.Graveyard.Contains(idle) {
		t.Error("the idle Elf never attacked and should have stayed in the graveyard")
	}
}

// "Ares attacks each combat if able" is the CR 508.1d requirement.
func TestAresMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	b12Push(g, g.Seats[0].ID, "Ares, God of War", "Legendary Creature — God Warrior Villain", aresGodOfWarOracle, 4, 3)
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.PassPriority(); err == nil {
		t.Fatal("passing with Ares home succeeded, want the attack requirement to refuse it")
	}
}

// "That card" is one object (CR 400.7): exiled from the graveyard in
// response it stays exiled, and one that left the graveyard and came
// back is a new object Ares's trigger never saw.
func TestAresReturnsOnlyTheObjectThatDied(t *testing.T) {
	for _, tc := range []struct {
		name  string
		leave func(t *testing.T, g *game.Game, owner, id uuid.UUID)
	}{
		{"exiled", func(t *testing.T, g *game.Game, _, id uuid.UUID) {
			g.WithWriteLock(func() {
				if err := g.ExileCardForEffect(id); err != nil {
					t.Fatalf("exile from graveyard: %v", err)
				}
			})
		}},
		{"left and came back", func(t *testing.T, g *game.Game, owner, id uuid.UUID) {
			if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
				t.Fatalf("graveyard to exile: %v", err)
			}
			if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner}, id); err != nil {
				t.Fatalf("exile back to graveyard: %v", err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			ares := b12Push(g, me.ID, "Ares, God of War", "Legendary Creature — God Warrior Villain", aresGodOfWarOracle, 4, 3)
			bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

			declareAttack(t, g, opp.ID, ares, bear)
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
			tc.leave(t, g, me.ID, bear)
			passPriorityAroundTable(t, g)
			if me.Hand.Contains(bear) {
				t.Error("Ares returned a card that is not the object that died")
			}
		})
	}
}

// Death Tyrant's graveyard ability returns it tapped, and it has menace.
func TestDeathTyrantReturnsFromGraveyardTapped(t *testing.T) {
	g := newCatalogGame(t)
	id, me := activateFromGraveyard(t, g, "Death Tyrant", "Creature — Beholder Skeleton",
		deathTyrantOracle, 4, 6, "{B}{B}{B}{B}{B}{B}", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) || me.Graveyard.Contains(id) {
		t.Fatal("Death Tyrant did not return from the graveyard to the battlefield")
	}
	if c, _ := g.LookupCardForEffect(id); !c.Tapped {
		t.Error("Death Tyrant must return tapped")
	}
	if !slices.Contains(effectiveAbilities(t, g, id), "menace") {
		t.Error("Death Tyrant has no menace on the battlefield")
	}
}

// Garna's "if it was attacking" is the same read, and it now sees an
// attacker nobody declared: a token created tapped and attacking. The
// event-log walk it replaced looked for an EventAttack and drew
// nothing here.
func TestGarnaSeesATokenPutOntoTheBattlefieldAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Garna, Bloodfist of Keld", "Legendary Creature — Human Berserker", b35GarnaOracle, 4, 3)

	// Nothing is declared: the Cat is the only attacker, and it never
	// had an EventAttack for the old log walk to find.
	advanceTo(t, g, game.StepDeclareAttackers)
	g.WithWriteLock(func() {
		if err := g.CreateTokensAttackingForEffect(me.ID, b31TappedCatLifelinkToken(), 1, opp.ID); err != nil {
			t.Fatalf("CreateTokensAttackingForEffect: %v", err)
		}
	})
	var cat uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.AttackingTarget == opp.ID {
			cat = c.InstanceID
		}
	}
	if cat == uuid.Nil {
		t.Fatal("setup: no attacking Cat token")
	}
	hand := me.Hand.Size()
	before := lifeOfOpponents(g)
	b18Kill(t, g, cat)
	if me.Hand.Size() != hand+1 {
		t.Errorf("an attacking token dying should draw: hand %d → %d", hand, me.Hand.Size())
	}
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d: %d → %d, want unchanged for an attacker", i+1, b, got)
		}
	}
}
