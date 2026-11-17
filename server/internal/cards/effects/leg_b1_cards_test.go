package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// leg_b1_cards_test.go — Legends, batch 1 (#5639). Each test plays the
// real catalog entry and asserts what its text says.

const (
	legB1PalladiaOracle  = "d7121ac7-e425-46cd-b006-1af391d97f87"
	legB1VaevictisOracle = "4d2cc2a8-e08a-420a-8922-c63e39129e23"
	legB1BolasOracle     = "e26bf6a9-b31c-4bc0-b55b-c01f2f69be6b"
	legB1XiraOracle      = "4a6e367c-7bc9-44a3-8ede-f2d0651abad3"
	legB1AdunOracle      = "aefd49a4-ecec-42cb-b882-e5db7f81eff8"
	legB1TuknirOracle    = "0b6fa658-a344-4202-866b-e19997e32ff2"
	legB1TorWaukiOracle  = "9439dfa0-daf8-4dac-af32-755fc2984126"
	legB1CaleriaOracle   = "74284bbc-b011-4e16-b8d5-7ddacc77f9d7"
	legB1RamsesOracle    = "6c9a9071-8159-4607-9465-de796f1dc2cc"
	legB1JacquesOracle   = "c030ca14-33cb-40b3-a1f5-b6d0cb0efd49"
	legB1KarakasOracle   = "59119143-c0fa-49dd-adf0-e2fd3029c48b"

	legB1ElderDragonType = "Legendary Creature — Elder Dragon"
)

func TestLegB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		legB1PalladiaOracle:  "Palladia-Mors",
		legB1VaevictisOracle: "Vaevictis Asmadi",
		legB1BolasOracle:     "Nicol Bolas",
		legB1XiraOracle:      "Xira Arien",
		legB1AdunOracle:      "Adun Oakenshield",
		legB1TuknirOracle:    "Tuknir Deathlock",
		legB1TorWaukiOracle:  "Tor Wauki",
		legB1CaleriaOracle:   "Lady Caleria",
		legB1RamsesOracle:    "Ramses Overdark",
		legB1JacquesOracle:   "Jacques le Vert",
		legB1KarakasOracle:   "Karakas",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// legB1Push seeds a permanent on the battlefield, catalog row or not,
// untapped and not summoning sick.
func legB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// legB1Activate activates the source's ability `index` for `who` with
// the given card targets, without resolving it.
func legB1Activate(g *game.Game, who, source uuid.UUID, index int, targets ...game.TargetRef) error {
	return g.ActivateCatalogAbility(who, source, index, game.ActivateAbilityParams{Strict: true, Targets: targets})
}

// legB1Attacking marks a creature as attacking `defender`.
func legB1Attacking(g *game.Game, id, defender uuid.UUID) {
	g.WithWriteLock(func() {
		if c := battlefieldCardFor(g, id); c != nil {
			c.AttackingTarget = defender
		}
	})
}

// --- the Elder Dragons' upkeep -----------------------------------------------------

func TestLegB1ElderDragonsAreSacrificedUnlessTheirUpkeepIsPaid(t *testing.T) {
	for _, tc := range []struct{ name, oracle, cost string }{
		{"Palladia-Mors", legB1PalladiaOracle, "{R}{G}{W}"},
		{"Vaevictis Asmadi", legB1VaevictisOracle, "{B}{R}{G}"},
		{"Nicol Bolas", legB1BolasOracle, "{U}{B}{R}"},
	} {
		t.Run(tc.name+" paid", func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			dragon := legB1Push(g, me.ID, tc.name, legB1ElderDragonType, tc.oracle, "", 7, 7)
			walkToUpkeepOfSeat0(t, g)
			passPriorityAroundTable(t, g)
			dcMana(t, g, me.ID, tc.cost)
			answerPayUnless(t, g, me.ID, true)
			if !g.Battlefield.Contains(dragon) {
				t.Fatalf("paying %s did not keep %s", tc.cost, tc.name)
			}
			if n := me.ManaPool.Total(); n != 0 {
				t.Errorf("%d mana left in the pool after paying %s", n, tc.cost)
			}
		})
		t.Run(tc.name+" declined", func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			dragon := legB1Push(g, me.ID, tc.name, legB1ElderDragonType, tc.oracle, "", 7, 7)
			walkToUpkeepOfSeat0(t, g)
			passPriorityAroundTable(t, g)
			answerPayUnless(t, g, me.ID, false)
			if g.Battlefield.Contains(dragon) {
				t.Fatalf("declining the upkeep did not sacrifice %s", tc.name)
			}
			if !me.Graveyard.Contains(dragon) {
				t.Errorf("%s is not in its owner's graveyard", tc.name)
			}
		})
	}
}

func TestLegB1ElderDragonsHaveTheirPrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		keywords     []string
	}{
		{"Palladia-Mors", legB1PalladiaOracle, []string{"flying", "trample"}},
		{"Vaevictis Asmadi", legB1VaevictisOracle, []string{"flying"}},
		{"Nicol Bolas", legB1BolasOracle, []string{"flying"}},
		{"Xira Arien", legB1XiraOracle, []string{"flying"}},
		{"Tuknir Deathlock", legB1TuknirOracle, []string{"flying"}},
	} {
		g := newCatalogGame(t)
		id := legB1Push(g, g.Seats[0].ID, tc.name, "Legendary Creature", tc.oracle, "", 2, 2)
		for _, kw := range tc.keywords {
			if !hasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s does not have %s", tc.name, kw)
			}
		}
	}
}

// --- Vaevictis Asmadi ----------------------------------------------------------------

func TestLegB1VaevictisAsmadiPumpsOnceForEachColour(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	dragon := legB1Push(g, me.ID, "Vaevictis Asmadi", legB1ElderDragonType, legB1VaevictisOracle, "", 7, 7)
	for i, mana := range []string{"{B}", "{R}", "{G}"} {
		dcMana(t, g, me.ID, mana)
		if err := legB1Activate(g, me.ID, dragon, i); err != nil {
			t.Fatalf("activate %s: %v", mana, err)
		}
		passPriorityAroundTable(t, g)
	}
	if p, tough := ptOf(t, g, dragon); p != 10 || tough != 7 {
		t.Errorf("Vaevictis Asmadi is %d/%d after three pumps, want 10/7", p, tough)
	}
	dcMana(t, g, me.ID, "{G}")
	if err := legB1Activate(g, me.ID, dragon, 0); err == nil {
		t.Error("the {B} pump was activated with only {G} in the pool")
	}
}

// --- Nicol Bolas ----------------------------------------------------------------------

func TestLegB1NicolBolasMakesTheDamagedOpponentDiscardTheirHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	bolas := legB1Push(g, me.ID, "Nicol Bolas", legB1ElderDragonType, legB1BolasOracle, "", 7, 7)
	bear := legB1Push(g, opp.ID, "Bear", "Creature — Bear", "", "", 2, 2)
	for _, p := range []*game.Player{me, opp, other} {
		pushCatalogHandCard(p, "One", "Instant", "")
		pushCatalogHandCard(p, "Two", "Sorcery", "")
	}
	myHand, oppHand, otherHand := me.Hand.Size(), opp.Hand.Size(), other.Hand.Size()
	oppGraveyard := opp.Graveyard.Size()

	dealNoncombatDamage(g, me.ID, bolas, bear)
	passPriorityAroundTable(t, g)
	if n := opp.Hand.Size(); n != oppHand {
		t.Fatalf("damage to a creature changed a hand: %d cards, want %d", n, oppHand)
	}

	dealCombatDamage(g, me.ID, bolas, opp.ID)
	passPriorityAroundTable(t, g)
	if n := opp.Hand.Size(); n != 0 {
		t.Errorf("the damaged opponent kept %d cards", n)
	}
	if n := opp.Graveyard.Size(); n != oppGraveyard+oppHand {
		t.Errorf("the damaged opponent's graveyard has %d cards, want %d", n, oppGraveyard+oppHand)
	}
	if me.Hand.Size() != myHand || other.Hand.Size() != otherHand {
		t.Errorf("hands of players not dealt damage changed: me %d (was %d), other %d (was %d)",
			me.Hand.Size(), myHand, other.Hand.Size(), otherHand)
	}
}

// --- Xira Arien -------------------------------------------------------------------------

func TestLegB1XiraArienMakesTargetPlayerDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	seedLibrary(opp, "Top")
	xira := legB1Push(g, me.ID, "Xira Arien", "Legendary Creature — Insect Wizard", legB1XiraOracle, "", 1, 2)
	before := opp.Hand.Size()
	dcMana(t, g, me.ID, "{B}{R}{G}")
	if err := legB1Activate(g, me.ID, xira, 0, playerTarget(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != before+1 {
		t.Errorf("the target has %d cards, want %d", got, before+1)
	}
	if !findCardAnywhere(t, g, xira).Tapped {
		t.Error("Xira Arien did not tap")
	}
}

// --- Adun Oakenshield -------------------------------------------------------------------

func TestLegB1AdunOakenshieldReturnsACreatureCardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	adun := legB1Push(g, me.ID, "Adun Oakenshield", "Legendary Creature — Human Knight", legB1AdunOracle, "", 1, 2)
	bear := pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	rite := pushGraveyardCardTyped(me, "Rite", "Sorcery")
	theirs := pushGraveyardCardTyped(opp, "Their Bear", "Creature — Bear")

	dcMana(t, g, me.ID, "{B}{R}{G}")
	for _, bad := range []uuid.UUID{rite, theirs} {
		if err := legB1Activate(g, me.ID, adun, 0, cardTarget(bad)...); err == nil {
			t.Fatal("a noncreature card, or a card in another graveyard, was accepted as the target")
		}
	}
	if err := legB1Activate(g, me.ID, adun, 0, cardTarget(bear)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature card did not come back to hand")
	}
}

// --- Tuknir Deathlock -------------------------------------------------------------------

func TestLegB1TuknirDeathlockGivesTargetCreaturePlusTwoPlusTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	tuknir := legB1Push(g, me.ID, "Tuknir Deathlock", "Legendary Creature — Human Wizard", legB1TuknirOracle, "", 2, 2)
	bear := legB1Push(g, opp.ID, "Bear", "Creature — Bear", "", "", 2, 2)
	dcMana(t, g, me.ID, "{R}{G}")
	if err := legB1Activate(g, me.ID, tuknir, 0, cardTarget(bear)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
}

// --- Tor Wauki, Lady Caleria -------------------------------------------------------------

func TestLegB1ArchersShootOnlyAttackingOrBlockingCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		damage                 int
	}{
		{"Tor Wauki", "Legendary Creature — Human Archer", legB1TorWaukiOracle, 2},
		{"Lady Caleria", "Legendary Creature — Elf Archer", legB1CaleriaOracle, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me, opp, _ := blbB1Seats(g)
			archer := legB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, "", 3, 3)
			attacker := legB1Push(g, opp.ID, "Attacker", "Creature — Ogre", "", "", 5, 5)
			idle := legB1Push(g, opp.ID, "Idle", "Creature — Ogre", "", "", 5, 5)
			legB1Attacking(g, attacker, me.ID)
			if err := legB1Activate(g, me.ID, archer, 0, cardTarget(idle)...); err == nil {
				t.Fatal("a creature neither attacking nor blocking was accepted as the target")
			}
			if err := legB1Activate(g, me.ID, archer, 0, cardTarget(attacker)...); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
			if got := findCardAnywhere(t, g, attacker).DamageMarked; got != tc.damage {
				t.Errorf("the attacker has %d damage, want %d", got, tc.damage)
			}
		})
	}
}

// --- Ramses Overdark ----------------------------------------------------------------------

func TestLegB1RamsesOverdarkDestroysOnlyAnEnchantedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	ramses := legB1Push(g, me.ID, "Ramses Overdark", "Legendary Creature — Human Assassin", legB1RamsesOracle, "", 4, 3)
	plain := legB1Push(g, opp.ID, "Plain", "Creature — Bear", "", "", 2, 2)
	enchanted := legB1Push(g, opp.ID, "Enchanted", "Creature — Bear", "", "", 2, 2)
	b15Aura(g, opp.ID, enchanted)
	if err := legB1Activate(g, me.ID, ramses, 0, cardTarget(plain)...); err == nil {
		t.Fatal("a creature with no Aura was accepted as the target")
	}
	if err := legB1Activate(g, me.ID, ramses, 0, cardTarget(enchanted)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(enchanted) {
		t.Error("the enchanted creature survived")
	}
	if !g.Battlefield.Contains(plain) {
		t.Error("the creature with no Aura was destroyed")
	}
}

// --- Jacques le Vert ------------------------------------------------------------------------

func TestLegB1JacquesLeVertGivesYourGreenCreaturesPlusZeroPlusTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	jacques := legB1Push(g, me.ID, "Jacques le Vert", "Legendary Creature — Human Warrior", legB1JacquesOracle, "{1}{R}{G}{W}", 3, 2)
	green := legB1Push(g, me.ID, "Green", "Creature — Bear", "", "{1}{G}", 2, 2)
	red := legB1Push(g, me.ID, "Red", "Creature — Ogre", "", "{1}{R}", 2, 2)
	theirs := legB1Push(g, opp.ID, "Their Green", "Creature — Bear", "", "{1}{G}", 2, 2)
	for _, tc := range []struct {
		name  string
		id    uuid.UUID
		p, tt int
	}{
		{"Jacques le Vert", jacques, 3, 4},
		{"my green creature", green, 2, 4},
		{"my red creature", red, 2, 2},
		{"an opponent's green creature", theirs, 2, 2},
	} {
		if p, tough := ptOf(t, g, tc.id); p != tc.p || tough != tc.tt {
			t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.p, tc.tt)
		}
	}
}

// --- Karakas ----------------------------------------------------------------------------------

func TestLegB1KarakasTapsForWhite(t *testing.T) {
	g := newCatalogGame(t)
	land := legB1Push(g, g.Seats[0].ID, "Karakas", "Legendary Land", legB1KarakasOracle, "", 0, 0)
	if got := ltProduced(t, g, land); !slices.Equal(got, []string{"{W}"}) {
		t.Errorf("Karakas adds %v, want {W}", got)
	}
}

func TestLegB1KarakasReturnsALegendaryCreatureToItsOwnersHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	land := legB1Push(g, me.ID, "Karakas", "Legendary Land", legB1KarakasOracle, "", 0, 0)
	legend := legB1Push(g, opp.ID, "Their Legend", "Legendary Creature — Human", "", "", 2, 2)
	plain := legB1Push(g, opp.ID, "Plain", "Creature — Human", "", "", 2, 2)
	if err := legB1Activate(g, me.ID, land, 0, cardTarget(plain)...); err == nil {
		t.Fatal("a nonlegendary creature was accepted as the target")
	}
	if err := legB1Activate(g, me.ID, land, 0, cardTarget(legend)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(legend) {
		t.Error("the legendary creature is not in its owner's hand")
	}
	if !g.Battlefield.Contains(plain) {
		t.Error("the nonlegendary creature moved")
	}
}
