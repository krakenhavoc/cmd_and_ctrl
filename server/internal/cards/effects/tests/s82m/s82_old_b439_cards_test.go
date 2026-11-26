package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b439_cards_test.go — Avacyn Restored, batch 439 (#5570, S82 OLD).
// Every test plays the real catalog entry through the engine.

const (
	oldB439Triumph = "87e88311-dbd2-42f7-a166-138f1d271686"
	oldB439Bower   = "f91ef713-5289-4901-9291-0a5afe6f997b"
	oldB439Grave   = "cc7d902e-c86e-4771-ba66-2f63e847cff1"
	oldB439Eaten   = "d4c04cec-3a0d-4e86-95ed-8d367fe7220d"
	oldB439Bear    = "Creature — Bear"
)

func TestOldB439CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB439Triumph: "Triumph of Cruelty", oldB439Bower: "Bower Passage", oldB439Grave: "Grave Exchange", oldB439Eaten: "Eaten by Spiders",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB439TriumphOfCrueltyDiscardsOnlyWithTheGreatestPower(t *testing.T) {
	for _, tc := range []struct {
		name      string
		myPower   int
		discarded bool
	}{{"greatest", 5, true}, {"tied", 4, true}, {"outpowered", 3, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.PushCatalogPermanent(g, me.ID, "Triumph of Cruelty", "Enchantment", oldB439Triumph, false)
			effectstest.BrCreature(g, me.ID, "Mine", oldB439Bear, tc.myPower, 2)
			effectstest.BrCreature(g, opp.ID, "Theirs", oldB439Bear, 4, 2)
			effectstest.PushCatalogHandCard(opp, "Card", "Sorcery", "")
			effectstest.WalkToUpkeepOfSeat0(t, g)
			effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}, true)
			owed := effectstest.DiscardOwed(g, opp.ID)
			if tc.discarded {
				if owed != 1 {
					t.Fatalf("the opponent owes %d discards, want 1", owed)
				}
				before := opp.Hand.Size()
				effectstest.DiscardFromHand(t, g, opp.ID)
				if opp.Hand.Size() != before-1 {
					t.Errorf("hand %d, want %d", opp.Hand.Size(), before-1)
				}
			} else if owed != 0 {
				t.Errorf("the opponent owes %d discards without the greatest power", owed)
			}
		})
	}
}

func TestOldB439BowerPassageStopsFliersBlockingCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Bower Passage", "Enchantment", oldB439Bower, false)
	flier := effectstest.BrCreature(g, opp.ID, "Flier", oldB439Bear, 2, 2, "flying")
	ground := effectstest.BrCreature(g, opp.ID, "Ground", oldB439Bear, 2, 2)
	attacker := effectstest.BrCreature(g, me.ID, "Attacker", oldB439Bear, 2, 2)
	effectstest.BrAttack(t, g, attacker)
	effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{{Blocker: flier, Attacker: attacker}}), game.BlockReasonCantBlockAttacker)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: ground, Attacker: attacker}}); err != nil {
		t.Fatalf("a creature without flying blocking: %v", err)
	}
}

func TestOldB439BowerPassageLeavesOtherPlayersAttackersAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Bower Passage", "Enchantment", oldB439Bower, false)
	flier := effectstest.BrCreature(g, opp.ID, "Flier", oldB439Bear, 2, 2, "flying")
	attacker := effectstest.BrCreature(g, me.ID, "Attacker", oldB439Bear, 2, 2)
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: flier, Attacker: attacker}}); err != nil {
		t.Fatalf("the Passage bound an attacker its controller does not control: %v", err)
	}
}

func TestOldB439EatenBySpidersDestroysTheFlierAndItsEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	flier := effectstest.BrCreature(g, opp.ID, "Flier", oldB439Bear, 2, 2, "flying")
	blade := effectstest.BrCreature(g, opp.ID, "Blade", "Artifact — Equipment", 0, 0)
	other := effectstest.BrCreature(g, opp.ID, "Other Blade", "Artifact — Equipment", 0, 0)
	effectstest.BrAttach(t, g, blade, flier)
	effectstest.CastCatalogSpell(t, g, "Eaten by Spiders", "Instant", oldB439Eaten, effectstest.CardTargets(flier))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flier) || g.Battlefield.Contains(blade) {
		t.Error("the flier or its Equipment survived")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("an unattached Equipment was destroyed")
	}
}

func TestOldB439EatenBySpidersRefusesACreatureWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.BrCreature(g, g.Seats[1].ID, "Bear", oldB439Bear, 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Eaten by Spiders", "Instant", oldB439Eaten, effectstest.CardTargets(bear)); err == nil {
		t.Error("Eaten by Spiders targeted a creature without flying")
	}
}

func TestOldB439GraveExchangeReturnsACreatureAndMakesAPlayerSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", oldB439Bear, "", 2, 2)
	victim := effectstest.BrCreature(g, opp.ID, "Victim", oldB439Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Grave Exchange", "Sorcery", oldB439Grave, []game.TargetRef{
		{Kind: game.TargetCard, ID: dead}, {Kind: game.TargetPlayer, ID: opp.ID},
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, victim)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the player did not sacrifice their creature")
	}
}
