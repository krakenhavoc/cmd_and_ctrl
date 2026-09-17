package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_declaration_test.go — #859 at the catalog level: what the
// cards that read "attacks <player>" see when the attacker is
// re-pointed before the declaration is complete.
//
// The engine side is server/internal/game/attack_declaration_test.go.
// Here the assertions are the observable ones: which player the Cat
// Soldier attacks, whose Ogre offer appears, who gets the Gold, where
// Hellrider's ping lands — every one of them the DEFENDER the
// attacker ended the declaration on, not the one it was first pointed
// at.

// repointAttack declares `attacker` against `from`, re-points it at
// `to` while the declaration is still open, and locks the declaration
// in — the whole #859 gesture in one call.
func repointAttack(t *testing.T, g *game.Game, attacker, from, to uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, from); err != nil {
		t.Fatalf("DeclareAttacker at %s: %v", from, err)
	}
	if err := g.DeclareAttacker(attacker, to); err != nil {
		t.Fatalf("re-point to %s: %v", to, err)
	}
	lockInAttacks(t, g)
}

// TestB28BrimazAttackTokenFollowsARepointedAttacker — the reader the
// issue names second (brimaz_king_of_oreskos.go:56, `defender :=
// ev.Target`): the Cat Soldier is put onto the battlefield attacking
// whoever Brimaz attacks, so a stale defender sent it at the wrong
// player.
func TestB28BrimazAttackTokenFollowsARepointedAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	brimaz := b12Push(g, me.ID, "Brimaz, King of Oreskos", "Legendary Creature — Cat Soldier", b28BrimazOracle, 3, 4)

	repointAttack(t, g, brimaz, first.ID, second.ID)
	passPriorityAroundTable(t, g)

	if got := b28TokensNamed(g, me.ID, "Cat Soldier"); got != 1 {
		t.Fatalf("a re-point is one attack, so one Cat Soldier: %d", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cat Soldier" && IsToken(c) && c.AttackingTarget != second.ID {
			t.Errorf("the token attacks %s, want the defender Brimaz ended on (%s)", c.AttackingTarget, second.ID)
		}
	}
}

// TestB17KazuulFollowsARepointedAttacker — the reader the issue names
// first, b17DefendingPlayer (batch17_helpers.go:73-78). Kazuul's
// "whenever a creature an opponent controls attacks you" is a trigger
// CONDITION on the defender, so a stale target does not merely
// mislabel the trigger: it creates one that should not exist and
// swallows one that should.
func TestB17KazuulFollowsARepointedAttacker(t *testing.T) {
	// Re-pointed AWAY from Kazuul's controller: no offer at all.
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Kazuul, Tyrant of the Cliffs", "Legendary Creature — Ogre Warrior", b17KazuulOracle, 5, 4)
	raider := b16Creature(g, opp.ID, "Raider", "Creature — Human", 2, 2, "R")
	advanceToMainOf(t, g, 1)

	repointAttack(t, g, raider, me.ID, other.ID)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, opp.ID) {
		t.Error("the attacker left Kazuul's controller: no Ogre offer is owed")
	}
	if n := len(battlefieldIDsNamed(g, "Ogre")); n != 0 {
		t.Errorf("%d Ogres for an attack that ended elsewhere, want 0", n)
	}

	// Re-pointed TOWARDS Kazuul's controller: the offer appears.
	g2 := newCatalogGame(t)
	me2, opp2, other2 := g2.Seats[0], g2.Seats[1], g2.Seats[2]
	b12Push(g2, me2.ID, "Kazuul, Tyrant of the Cliffs", "Legendary Creature — Ogre Warrior", b17KazuulOracle, 5, 4)
	raider2 := b16Creature(g2, opp2.ID, "Raider", "Creature — Human", 2, 2, "R")
	advanceToMainOf(t, g2, 1)

	repointAttack(t, g2, raider2, other2.ID, me2.ID)
	passPriorityAroundTable(t, g2)
	if !hasPayUnlessFor(g2, opp2.ID) {
		t.Fatal("the attacker ended on Kazuul's controller: one Ogre offer is owed")
	}
	answerPayUnless(t, g2, opp2.ID, false)
	passPriorityAroundTable(t, g2)
	ogres := battlefieldIDsNamed(g2, "Ogre")
	if len(ogres) != 1 {
		t.Fatalf("%d Ogres for a declined offer, want 1", len(ogres))
	}
	if controllerOf(t, g2, ogres[0]) != me2.ID {
		t.Error("the Ogre belongs to Kazuul's controller")
	}
	if hasPayUnlessFor(g2, opp2.ID) {
		t.Error("one attacker, one offer")
	}
}

// TestCurseOfOpulenceFollowsARepointedAttacker — the enchanted player
// is the trigger's condition (curse_of_opulence.go:55). A creature
// that leaves them pays no Gold, and one that arrives on them does.
func TestCurseOfOpulenceFollowsARepointedAttacker(t *testing.T) {
	// A fresh table per direction: the Curse is a once-per-batch
	// trigger, so re-using one game would let the first attack's
	// bookkeeping colour the second.
	for _, tc := range []struct {
		name      string
		towards   bool
		wantGolds int
	}{
		{"away from the cursed player", false, 0},
		{"towards the cursed player", true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
			curse := castCatalogSpell(t, g, "Curse of Opulence", curseAuraTypeLine, curseOpulenceOracle,
				[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
			passPriorityAroundTable(t, g)
			if host := attachmentHostOf(t, g, curse); host.Kind != game.TargetPlayer || host.ID != victim.ID {
				t.Fatalf("Curse AttachedTo = %+v, want player %s", host, victim.ID)
			}
			attacker := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
				Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
			})

			from, to := victim.ID, bystander.ID
			if tc.towards {
				from, to = bystander.ID, victim.ID
			}
			repointAttack(t, g, attacker, from, to)
			passPriorityAroundTable(t, g)
			if n := countBattlefieldNamed(g, me.ID, "Gold"); n != tc.wantGolds {
				t.Errorf("%d Gold, want %d — the Curse follows the defender the attacker ended on", n, tc.wantGolds)
			}
		})
	}
}

// TestHellriderPingsTheDefenderTheAttackerEndedOn — the plainest of
// the body readers (hellrider.go:51). Two attackers at two seats, one
// of them re-pointed: two pings, both where the creatures ended.
func TestHellriderPingsTheDefenderTheAttackerEndedOn(t *testing.T) {
	g := newCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	pushDiesCreatureForTest(g, me.ID, "Hellrider", hellriderOracle, "Creature — Devil", 3, 3)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	firstBefore, secondBefore := first.Life, second.Life

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, first.ID); err != nil {
		t.Fatalf("DeclareAttacker a: %v", err)
	}
	if err := g.DeclareAttacker(b, first.ID); err != nil {
		t.Fatalf("DeclareAttacker b: %v", err)
	}
	if err := g.DeclareAttacker(b, second.ID); err != nil {
		t.Fatalf("re-point b: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)

	if got := firstBefore - first.Life; got != 1 {
		t.Errorf("the seat one attacker stayed on took %d, want 1", got)
	}
	if got := secondBefore - second.Life; got != 1 {
		t.Errorf("the seat the other attacker moved to took %d, want 1", got)
	}
}

// TestAdelineMakesOneBatchOfHumansAcrossARepoint — the #854 batch
// rule survives the lock-in: two attackers, one of them re-pointed
// mid-declaration, are one declaration and therefore one occurrence.
func TestAdelineMakesOneBatchOfHumansAcrossARepoint(t *testing.T) {
	g := newCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Adeline, Resplendent Cathar", "Legendary Creature — Human Knight",
		"38515f89-348b-4cf3-b7bd-1f6fe4ce2fba", false)
	a := pushCatalogPermanent(g, me.ID, "Bear A", "Creature — Bear", "", false)
	b := pushCatalogPermanent(g, me.ID, "Bear B", "Creature — Bear", "", false)
	before := countBattlefieldNamed(g, me.ID, "Human")

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, first.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	if err := g.DeclareAttacker(a, second.ID); err != nil {
		t.Fatalf("re-point a: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)

	if got := countBattlefieldNamed(g, me.ID, "Human") - before; got != len(g.Seats)-1 {
		t.Errorf("a re-pointed declaration made %d Humans, want %d (one per opponent, one batch)",
			got, len(g.Seats)-1)
	}
}
