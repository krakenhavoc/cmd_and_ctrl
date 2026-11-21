package testsl

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestLegB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.LegB1PalladiaOracle:  "Palladia-Mors",
		effectstest.LegB1VaevictisOracle: "Vaevictis Asmadi",
		effectstest.LegB1BolasOracle:     "Nicol Bolas",
		effectstest.LegB1XiraOracle:      "Xira Arien",
		effectstest.LegB1AdunOracle:      "Adun Oakenshield",
		effectstest.LegB1TuknirOracle:    "Tuknir Deathlock",
		effectstest.LegB1TorWaukiOracle:  "Tor Wauki",
		effectstest.LegB1CaleriaOracle:   "Lady Caleria",
		effectstest.LegB1RamsesOracle:    "Ramses Overdark",
		effectstest.LegB1JacquesOracle:   "Jacques le Vert",
		effectstest.LegB1KarakasOracle:   "Karakas",
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

// legB1Activate activates the source's ability `index` for `who` with
// the given card targets, without resolving it.
func legB1Activate(g *game.Game, who, source uuid.UUID, index int, targets ...game.TargetRef) error {
	return g.ActivateCatalogAbility(who, source, index, game.ActivateAbilityParams{Strict: true, Targets: targets})
}

// legB1Attacking marks a creature as attacking `defender`.
func legB1Attacking(g *game.Game, id, defender uuid.UUID) {
	g.WithWriteLock(func() {
		if c := effectstest.BattlefieldCardFor(g, id); c != nil {
			c.AttackingTarget = defender
		}
	})
}

// --- the Elder Dragons' upkeep -----------------------------------------------------

func TestLegB1ElderDragonsAreSacrificedUnlessTheirUpkeepIsPaid(t *testing.T) {
	for _, tc := range []struct{ name, oracle, cost string }{
		{"Palladia-Mors", effectstest.LegB1PalladiaOracle, "{R}{G}{W}"},
		{"Vaevictis Asmadi", effectstest.LegB1VaevictisOracle, "{B}{R}{G}"},
		{"Nicol Bolas", effectstest.LegB1BolasOracle, "{U}{B}{R}"},
	} {
		t.Run(tc.name+" paid", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			dragon := effectstest.LegB1Push(g, me.ID, tc.name, effectstest.LegB1ElderDragonType, tc.oracle, "", 7, 7)
			effectstest.WalkToUpkeepOfSeat0(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.DcMana(t, g, me.ID, tc.cost)
			effectstest.AnswerPayUnless(t, g, me.ID, true)
			if !g.Battlefield.Contains(dragon) {
				t.Fatalf("paying %s did not keep %s", tc.cost, tc.name)
			}
			if n := me.ManaPool.Total(); n != 0 {
				t.Errorf("%d mana left in the pool after paying %s", n, tc.cost)
			}
		})
		t.Run(tc.name+" declined", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			dragon := effectstest.LegB1Push(g, me.ID, tc.name, effectstest.LegB1ElderDragonType, tc.oracle, "", 7, 7)
			effectstest.WalkToUpkeepOfSeat0(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerPayUnless(t, g, me.ID, false)
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
		{"Palladia-Mors", effectstest.LegB1PalladiaOracle, []string{"flying", "trample"}},
		{"Vaevictis Asmadi", effectstest.LegB1VaevictisOracle, []string{"flying"}},
		{"Nicol Bolas", effectstest.LegB1BolasOracle, []string{"flying"}},
		{"Xira Arien", effectstest.LegB1XiraOracle, []string{"flying"}},
		{"Tuknir Deathlock", effectstest.LegB1TuknirOracle, []string{"flying"}},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.LegB1Push(g, g.Seats[0].ID, tc.name, "Legendary Creature", tc.oracle, "", 2, 2)
		for _, kw := range tc.keywords {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s does not have %s", tc.name, kw)
			}
		}
	}
}

// --- Vaevictis Asmadi ----------------------------------------------------------------

func TestLegB1VaevictisAsmadiPumpsOnceForEachColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	dragon := effectstest.LegB1Push(g, me.ID, "Vaevictis Asmadi", effectstest.LegB1ElderDragonType, effectstest.LegB1VaevictisOracle, "", 7, 7)
	for i, mana := range []string{"{B}", "{R}", "{G}"} {
		effectstest.DcMana(t, g, me.ID, mana)
		if err := legB1Activate(g, me.ID, dragon, i); err != nil {
			t.Fatalf("activate %s: %v", mana, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.PtOf(t, g, dragon); p != 10 || tough != 7 {
		t.Errorf("Vaevictis Asmadi is %d/%d after three pumps, want 10/7", p, tough)
	}
	effectstest.DcMana(t, g, me.ID, "{G}")
	if err := legB1Activate(g, me.ID, dragon, 0); err == nil {
		t.Error("the {B} pump was activated with only {G} in the pool")
	}
}

// --- Nicol Bolas ----------------------------------------------------------------------

func TestLegB1NicolBolasMakesTheDamagedOpponentDiscardTheirHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	bolas := effectstest.LegB1Push(g, me.ID, "Nicol Bolas", effectstest.LegB1ElderDragonType, effectstest.LegB1BolasOracle, "", 7, 7)
	bear := effectstest.LegB1Push(g, opp.ID, "Bear", "Creature — Bear", "", "", 2, 2)
	for _, p := range []*game.Player{me, opp, other} {
		effectstest.PushCatalogHandCard(p, "One", "Instant", "")
		effectstest.PushCatalogHandCard(p, "Two", "Sorcery", "")
	}
	myHand, oppHand, otherHand := me.Hand.Size(), opp.Hand.Size(), other.Hand.Size()
	oppGraveyard := opp.Graveyard.Size()

	effectstest.DealNoncombatDamage(g, me.ID, bolas, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := opp.Hand.Size(); n != oppHand {
		t.Fatalf("damage to a creature changed a hand: %d cards, want %d", n, oppHand)
	}

	effectstest.DealCombatDamage(g, me.ID, bolas, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.SeedLibrary(opp, "Top")
	xira := effectstest.LegB1Push(g, me.ID, "Xira Arien", "Legendary Creature — Insect Wizard", effectstest.LegB1XiraOracle, "", 1, 2)
	before := opp.Hand.Size()
	effectstest.DcMana(t, g, me.ID, "{B}{R}{G}")
	if err := legB1Activate(g, me.ID, xira, 0, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != before+1 {
		t.Errorf("the target has %d cards, want %d", got, before+1)
	}
	if !effectstest.FindCardAnywhere(t, g, xira).Tapped {
		t.Error("Xira Arien did not tap")
	}
}

// --- Adun Oakenshield -------------------------------------------------------------------

func TestLegB1AdunOakenshieldReturnsACreatureCardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	adun := effectstest.LegB1Push(g, me.ID, "Adun Oakenshield", "Legendary Creature — Human Knight", effectstest.LegB1AdunOracle, "", 1, 2)
	bear := effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	rite := effectstest.PushGraveyardCardTyped(me, "Rite", "Sorcery")
	theirs := effectstest.PushGraveyardCardTyped(opp, "Their Bear", "Creature — Bear")

	effectstest.DcMana(t, g, me.ID, "{B}{R}{G}")
	for _, bad := range []uuid.UUID{rite, theirs} {
		if err := legB1Activate(g, me.ID, adun, 0, effectstest.CardTarget(bad)...); err == nil {
			t.Fatal("a noncreature card, or a card in another graveyard, was accepted as the target")
		}
	}
	if err := legB1Activate(g, me.ID, adun, 0, effectstest.CardTarget(bear)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature card did not come back to hand")
	}
}

// --- Tuknir Deathlock -------------------------------------------------------------------

func TestLegB1TuknirDeathlockGivesTargetCreaturePlusTwoPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	tuknir := effectstest.LegB1Push(g, me.ID, "Tuknir Deathlock", "Legendary Creature — Human Wizard", effectstest.LegB1TuknirOracle, "", 2, 2)
	bear := effectstest.LegB1Push(g, opp.ID, "Bear", "Creature — Bear", "", "", 2, 2)
	effectstest.DcMana(t, g, me.ID, "{R}{G}")
	if err := legB1Activate(g, me.ID, tuknir, 0, effectstest.CardTarget(bear)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
}

// --- Tor Wauki, Lady Caleria -------------------------------------------------------------

func TestLegB1ArchersShootOnlyAttackingOrBlockingCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		damage                 int
	}{
		{"Tor Wauki", "Legendary Creature — Human Archer", effectstest.LegB1TorWaukiOracle, 2},
		{"Lady Caleria", "Legendary Creature — Elf Archer", effectstest.LegB1CaleriaOracle, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp, _ := effectstest.BlbB1Seats(g)
			archer := effectstest.LegB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, "", 3, 3)
			attacker := effectstest.LegB1Push(g, opp.ID, "Attacker", "Creature — Ogre", "", "", 5, 5)
			idle := effectstest.LegB1Push(g, opp.ID, "Idle", "Creature — Ogre", "", "", 5, 5)
			legB1Attacking(g, attacker, me.ID)
			if err := legB1Activate(g, me.ID, archer, 0, effectstest.CardTarget(idle)...); err == nil {
				t.Fatal("a creature neither attacking nor blocking was accepted as the target")
			}
			if err := legB1Activate(g, me.ID, archer, 0, effectstest.CardTarget(attacker)...); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.FindCardAnywhere(t, g, attacker).DamageMarked; got != tc.damage {
				t.Errorf("the attacker has %d damage, want %d", got, tc.damage)
			}
		})
	}
}

// --- Ramses Overdark ----------------------------------------------------------------------

func TestLegB1RamsesOverdarkDestroysOnlyAnEnchantedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	ramses := effectstest.LegB1Push(g, me.ID, "Ramses Overdark", "Legendary Creature — Human Assassin", effectstest.LegB1RamsesOracle, "", 4, 3)
	plain := effectstest.LegB1Push(g, opp.ID, "Plain", "Creature — Bear", "", "", 2, 2)
	enchanted := effectstest.LegB1Push(g, opp.ID, "Enchanted", "Creature — Bear", "", "", 2, 2)
	effectstest.B15Aura(g, opp.ID, enchanted)
	if err := legB1Activate(g, me.ID, ramses, 0, effectstest.CardTarget(plain)...); err == nil {
		t.Fatal("a creature with no Aura was accepted as the target")
	}
	if err := legB1Activate(g, me.ID, ramses, 0, effectstest.CardTarget(enchanted)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(enchanted) {
		t.Error("the enchanted creature survived")
	}
	if !g.Battlefield.Contains(plain) {
		t.Error("the creature with no Aura was destroyed")
	}
}

// --- Jacques le Vert ------------------------------------------------------------------------

func TestLegB1JacquesLeVertGivesYourGreenCreaturesPlusZeroPlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	jacques := effectstest.LegB1Push(g, me.ID, "Jacques le Vert", "Legendary Creature — Human Warrior", effectstest.LegB1JacquesOracle, "{1}{R}{G}{W}", 3, 2)
	green := effectstest.LegB1Push(g, me.ID, "Green", "Creature — Bear", "", "{1}{G}", 2, 2)
	red := effectstest.LegB1Push(g, me.ID, "Red", "Creature — Ogre", "", "{1}{R}", 2, 2)
	theirs := effectstest.LegB1Push(g, opp.ID, "Their Green", "Creature — Bear", "", "{1}{G}", 2, 2)
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
		if p, tough := effectstest.PtOf(t, g, tc.id); p != tc.p || tough != tc.tt {
			t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.p, tc.tt)
		}
	}
}

// --- Karakas ----------------------------------------------------------------------------------

func TestLegB1KarakasTapsForWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.LegB1Push(g, g.Seats[0].ID, "Karakas", "Legendary Land", effectstest.LegB1KarakasOracle, "", 0, 0)
	if got := effectstest.LtProduced(t, g, land); !slices.Equal(got, []string{"{W}"}) {
		t.Errorf("Karakas adds %v, want {W}", got)
	}
}

func TestLegB1KarakasReturnsALegendaryCreatureToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	land := effectstest.LegB1Push(g, me.ID, "Karakas", "Legendary Land", effectstest.LegB1KarakasOracle, "", 0, 0)
	legend := effectstest.LegB1Push(g, opp.ID, "Their Legend", "Legendary Creature — Human", "", "", 2, 2)
	plain := effectstest.LegB1Push(g, opp.ID, "Plain", "Creature — Human", "", "", 2, 2)
	if err := legB1Activate(g, me.ID, land, 0, effectstest.CardTarget(plain)...); err == nil {
		t.Fatal("a nonlegendary creature was accepted as the target")
	}
	if err := legB1Activate(g, me.ID, land, 0, effectstest.CardTarget(legend)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(legend) {
		t.Error("the legendary creature is not in its owner's hand")
	}
	if !g.Battlefield.Contains(plain) {
		t.Error("the nonlegendary creature moved")
	}
}
