package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, from); err != nil {
		t.Fatalf("DeclareAttacker at %s: %v", from, err)
	}
	if err := g.DeclareAttacker(attacker, to); err != nil {
		t.Fatalf("re-point to %s: %v", to, err)
	}
	effectstest.LockInAttacks(t, g)
}

// TestB28BrimazAttackTokenFollowsARepointedAttacker — the reader the
// issue names second (brimaz_king_of_oreskos.go:56, `defender :=
// ev.Target`): the Cat Soldier is put onto the battlefield attacking
// whoever Brimaz attacks, so a stale defender sent it at the wrong
// player.
func TestB28BrimazAttackTokenFollowsARepointedAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	brimaz := effectstest.B12Push(g, me.ID, "Brimaz, King of Oreskos", "Legendary Creature — Cat Soldier", effectstest.B28BrimazOracle, 3, 4)

	repointAttack(t, g, brimaz, first.ID, second.ID)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)

	if got := effectstest.B28TokensNamed(g, me.ID, "Cat Soldier"); got != 1 {
		t.Fatalf("a re-point is one attack, so one Cat Soldier: %d", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Cat Soldier" && effects.IsToken(c) && c.AttackingTarget != second.ID {
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
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, "Kazuul, Tyrant of the Cliffs", "Legendary Creature — Ogre Warrior", effectstest.B17KazuulOracle, 5, 4)
	raider := effectstest.B16Creature(g, opp.ID, "Raider", "Creature — Human", 2, 2, "R")
	effectstest.AdvanceToMainOf(t, g, 1)

	repointAttack(t, g, raider, me.ID, other.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Error("the attacker left Kazuul's controller: no Ogre offer is owed")
	}
	if n := len(effectstest.BattlefieldIDsNamed(g, "Ogre")); n != 0 {
		t.Errorf("%d Ogres for an attack that ended elsewhere, want 0", n)
	}

	// Re-pointed TOWARDS Kazuul's controller: the offer appears.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2, other2 := g2.Seats[0], g2.Seats[1], g2.Seats[2]
	effectstest.B12Push(g2, me2.ID, "Kazuul, Tyrant of the Cliffs", "Legendary Creature — Ogre Warrior", effectstest.B17KazuulOracle, 5, 4)
	raider2 := effectstest.B16Creature(g2, opp2.ID, "Raider", "Creature — Human", 2, 2, "R")
	effectstest.AdvanceToMainOf(t, g2, 1)

	repointAttack(t, g2, raider2, other2.ID, me2.ID)
	effectstest.PassPriorityAroundTable(t, g2)
	if !effectstest.HasPayUnlessFor(g2, opp2.ID) {
		t.Fatal("the attacker ended on Kazuul's controller: one Ogre offer is owed")
	}
	effectstest.AnswerPayUnless(t, g2, opp2.ID, false)
	effectstest.PassPriorityAroundTable(t, g2)
	ogres := effectstest.BattlefieldIDsNamed(g2, "Ogre")
	if len(ogres) != 1 {
		t.Fatalf("%d Ogres for a declined offer, want 1", len(ogres))
	}
	if effectstest.ControllerOf(t, g2, ogres[0]) != me2.ID {
		t.Error("the Ogre belongs to Kazuul's controller")
	}
	if effectstest.HasPayUnlessFor(g2, opp2.ID) {
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
			g := effectstest.NewCatalogGame(t)
			me, victim, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
			curse := effectstest.CastCatalogSpell(t, g, "Curse of Opulence", effectstest.CurseAuraTypeLine, effectstest.CurseOpulenceOracle,
				[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim.ID}})
			effectstest.PassPriorityAroundTable(t, g)
			if host := effectstest.AttachmentHostOf(t, g, curse); host.Kind != game.TargetPlayer || host.ID != victim.ID {
				t.Fatalf("Curse AttachedTo = %+v, want player %s", host, victim.ID)
			}
			attacker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.TestCreatureTypeLine,
				Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
			})

			from, to := victim.ID, bystander.ID
			if tc.towards {
				from, to = bystander.ID, victim.ID
			}
			repointAttack(t, g, attacker, from, to)
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.CountBattlefieldNamed(g, me.ID, "Gold"); n != tc.wantGolds {
				t.Errorf("%d Gold, want %d — the Curse follows the defender the attacker ended on", n, tc.wantGolds)
			}
		})
	}
}

// TestHellriderPingsTheDefenderTheAttackerEndedOn — the plainest of
// the body readers (hellrider.go:51). Two attackers at two seats, one
// of them re-pointed: two pings, both where the creatures ended.
func TestHellriderPingsTheDefenderTheAttackerEndedOn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Hellrider", effectstest.HellriderOracle, "Creature — Devil", 3, 3)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	firstBefore, secondBefore := first.Life, second.Life

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, first.ID); err != nil {
		t.Fatalf("DeclareAttacker a: %v", err)
	}
	if err := g.DeclareAttacker(b, first.ID); err != nil {
		t.Fatalf("DeclareAttacker b: %v", err)
	}
	if err := g.DeclareAttacker(b, second.ID); err != nil {
		t.Fatalf("re-point b: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me, first, second := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Adeline, Resplendent Cathar", "Legendary Creature — Human Knight",
		"38515f89-348b-4cf3-b7bd-1f6fe4ce2fba", false)
	a := effectstest.PushCatalogPermanent(g, me.ID, "Bear A", "Creature — Bear", "", false)
	b := effectstest.PushCatalogPermanent(g, me.ID, "Bear B", "Creature — Bear", "", false)
	before := effectstest.CountBattlefieldNamed(g, me.ID, "Human")

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, first.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	if err := g.DeclareAttacker(a, second.ID); err != nil {
		t.Fatalf("re-point a: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Human") - before; got != len(g.Seats)-1 {
		t.Errorf("a re-pointed declaration made %d Humans, want %d (one per opponent, one batch)",
			got, len(g.Seats)-1)
	}
}
