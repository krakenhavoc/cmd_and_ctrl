package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b075_test.go — Kaldheim Commander, batch 75 (#5506): +1/+1 and other
// counters, activated abilities, Auras and graveyard recursion. Each test plays
// the real catalog entry and asserts what its text says.

const (
	oldB075Hero     = "83f1408e-8270-49b4-9d3f-eacfd3645543"
	oldB075Scarlade = "b7ebd207-43e3-49eb-86ca-b5987ef9e22f"
	oldB075Crown    = "db961314-fe78-4362-902f-aa0dbf6c487d"
	oldB075Admirers = "ea97fc76-3a32-414c-aa12-87be1a0eb9f3"
)

func oldB075Elf(g *game.Game, p *game.Player, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Elf Warrior",
		Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
}

func TestOldB075EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB075Hero: "Hero of Bretagard", oldB075Scarlade: "Lys Alana Scarblade", oldB075Crown: "Crown of Skemfar", oldB075Admirers: "Masked Admirers",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s is %v with caveats %v, want Full with none", name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Hero of Bretagard ------------------------------------------------------

func TestOldB075HeroGetsACounterPerPermanentMyAbilityExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hero := effectstest.PushCatalogPermanent(g, me.ID, "Hero of Bretagard", "Creature — Human Warrior", oldB075Hero, false)
	a := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Their A")
	b := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Their B")
	effectstest.CastXSpell(t, g, "Curse of the Swine", "Sorcery", effectstest.B07CurseOfTheSwineOracle, "{X}{U}{U}", 2,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Fatal("the spell did not exile both creatures")
	}
	if n := effectstest.CountersOn(g, hero, game.CounterPlusOne); n != 2 {
		t.Errorf("%d +1/+1 counters, want 2 (one trigger, that many)", n)
	}
}

func TestOldB075HeroIgnoresAnOpponentsExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hero := effectstest.PushCatalogPermanent(g, me.ID, "Hero of Bretagard", "Creature — Human Warrior", oldB075Hero, false)
	mine := effectstest.PushCreatureToBattlefieldForTest(g, me.ID, "My Bear")
	_ = caster
	effectstest.CastCatalogSpell(t, g, "Swords to Plowshares", "Instant", "b1544f21-7e98-461b-aed5-e748b0168c52",
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(mine) {
		t.Fatal("Swords did not exile the creature")
	}
	if n := effectstest.CountersOn(g, hero, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters from an opponent's exile, want 0", n)
	}
}

func TestOldB075HeroCountsACardForetoldFromMyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hero := effectstest.PushCatalogPermanent(g, me.ID, "Hero of Bretagard", "Creature — Human Warrior", oldB075Hero, false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	card := effectstest.HandCardForTest(me, "Saw It Coming", "Instant", effectstest.SawItComingOracle)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.PerformSpecialAction(me.ID, card, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("foretell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, hero, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after a foretell, want 1", n)
	}
}

func TestOldB075HeroGainsFlyingAndAngelAtFiveCountersThenIndestructibleAndGodAtTen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hero := effectstest.PushCatalogPermanent(g, me.ID, "Hero of Bretagard", "Creature — Human Warrior", oldB075Hero, false)
	hasType := func(sub string) bool {
		var ok bool
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == hero {
					ok = c.HasSubtype(sub)
				}
			}
		})
		return ok
	}
	effectstest.AddCountersForTest(g, hero, game.CounterPlusOne, 4)
	if effectstest.HasKeywordOnBattlefield(t, g, hero, "flying") || hasType("Angel") {
		t.Fatal("four counters already gave flying or the Angel type")
	}
	effectstest.AddCountersForTest(g, hero, "charge", 1) // counters of any kind count
	if !effectstest.HasKeywordOnBattlefield(t, g, hero, "flying") || !hasType("Angel") || !hasType("Human") {
		t.Error("five counters did not give flying and the Angel type (keeping Human)")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, hero, "indestructible") || hasType("God") {
		t.Error("five counters gave indestructible or God")
	}
	effectstest.AddCountersForTest(g, hero, game.CounterPlusOne, 5)
	if !effectstest.HasKeywordOnBattlefield(t, g, hero, "indestructible") || !hasType("God") || !hasType("Angel") {
		t.Error("ten counters did not give indestructible and the God type")
	}
}

// --- Lys Alana Scarblade ----------------------------------------------------

func TestOldB075ScarbladeDiscardsAnElfCardForMinusXWhereXIsYourElves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blade := effectstest.PushCatalogPermanent(g, me.ID, "Lys Alana Scarblade", "Creature — Elf Assassin", oldB075Scarlade, false)
	oldB075Elf(g, me, "Elf")
	oldB075Elf(g, opp, "Their Elf") // not yours
	elfCard := effectstest.PutInHand(me, game.Card{Name: "Elf Card", TypeLine: "Creature — Elf Druid"})
	bearCard := effectstest.PutInHand(me, game.Card{Name: "Bear Card", TypeLine: "Creature — Bear"})
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Ogre", 4, 4)
	effectstest.AdvanceToMain(t, g)
	// A non-Elf card is not a legal discard.
	if err := g.ActivateCatalogAbility(me.ID, blade, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(victim), DiscardIDs: []uuid.UUID{bearCard}}); err == nil {
		t.Fatal("discarded a non-Elf card to pay")
	}
	if err := g.ActivateCatalogAbility(me.ID, blade, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(victim), DiscardIDs: []uuid.UUID{elfCard}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(elfCard) {
		t.Error("the Elf card was not discarded")
	}
	// Two Elves you control: the Scarblade and the other Elf.
	if p, tough := effectstest.PtOf(t, g, victim); p != 2 || tough != 2 {
		t.Errorf("the target is %d/%d, want 2/2 (-2/-2)", p, tough)
	}
	if !effectstest.TappedNow(g, blade) {
		t.Error("the Scarblade did not tap")
	}
}

// --- Crown of Skemfar -------------------------------------------------------

func TestOldB075CrownPumpsPerElfYouControlAndGivesReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	oldB075Elf(g, me, "Elf A")
	oldB075Elf(g, me, "Elf B")
	oldB075Elf(g, opp, "Their Elf")
	effectstest.Enchant(t, g, "Crown of Skemfar", oldB075Crown, host)
	if p, tough := effectstest.PtOf(t, g, host); p != 4 || tough != 4 {
		t.Errorf("the enchanted Bear is %d/%d, want 4/4 (two Elves you control)", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, host, "reach") {
		t.Error("the enchanted creature lacks reach")
	}
	oldB075Elf(g, me, "Elf C")
	if p, _ := effectstest.PtOf(t, g, host); p != 5 {
		t.Errorf("the enchanted Bear has %d power after another Elf, want 5", p)
	}
}

func TestOldB075CrownReturnsFromTheGraveyardForTwoAndAGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crown := effectstest.PushGraveyardCardTyped(me, "Crown of Skemfar", "Enchantment — Aura")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == crown {
				me.Graveyard.Cards[i].OracleID = oldB075Crown
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, crown, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {G}")
	}
	effectstest.FloatMana(t, g, me, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, crown, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(crown) || me.Graveyard.Contains(crown) {
		t.Error("the Crown did not return to the hand")
	}
}

// --- Masked Admirers --------------------------------------------------------

func TestOldB075AdmirersDrawsWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	effectstest.EnterFromHand(t, g, me.ID, "Masked Admirers", "Creature — Elf Shaman", oldB075Admirers)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), before+1)
	}
}

func oldB075AdmirersInGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	id := effectstest.PushGraveyardCardTyped(me, "Masked Admirers", "Creature — Elf Shaman")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == id {
				me.Graveyard.Cards[i].OracleID = oldB075Admirers
			}
		}
	})
	return id
}

func TestOldB075AdmirersReturnsFromTheGraveyardWhenYouCastACreatureAndPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	admirers := oldB075AdmirersInGraveyard(g, me)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{G}{G}{G}")
	effectstest.CastFromHandForTest(t, g, me, "Bear", "Creature — Bear", "{G}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(admirers) || me.Graveyard.Contains(admirers) {
		t.Error("Masked Admirers did not return to the hand")
	}
}

func TestOldB075AdmirersStaysWhenYouDeclineOrCastANoncreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	admirers := oldB075AdmirersInGraveyard(g, me)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{G}{G}{G}{R}")
	effectstest.CastFromHandForTest(t, g, me, "Shock", "Instant", "{R}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a noncreature spell triggered Masked Admirers")
	}
	effectstest.CastFromHandForTest(t, g, me, "Bear", "Creature — Bear", "{G}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(admirers) {
		t.Error("Masked Admirers left the graveyard without payment")
	}
}
