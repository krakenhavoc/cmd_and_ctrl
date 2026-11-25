package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b155_test.go — Core Set 2020 auras and planeswalkers, S82 set OLD
// batch 155 (#5518). Every test plays the real catalog entry through the
// engine.

const (
	oldB155Aegis      = "0b0c98ba-3505-40aa-b0f2-e03c9790f51e"
	oldB155Rage       = "b5dbb212-0399-475b-a831-921f7c76cdc4"
	oldB155Indenture  = "df8fe368-00eb-42fa-89d6-794fc61f8f8c"
	oldB155Gauntlets  = "e5741c85-cd09-4ccb-ab39-3f3a62f98939"
	oldB155HardCover  = "8613d71b-4ac9-490d-b883-aefe03bcebce"
	oldB155Ajani      = "15474e82-94b4-47e1-b91c-3deaf4dbffd6"
	oldB155VivienBow  = "3e83dec6-e1a0-4137-aaab-308dfd3122e1"
	oldB155VivienNat  = "5513b150-c08f-4f3e-8e41-fd7b43d00b78"
	oldB155MuYanling  = "5f39accf-62fd-407a-a147-781210569d33"
	oldB155ChandraAoF = "626e8acc-20da-496c-9a79-bfbf7529c01d"
)

func TestOldB155CardsAreRegisteredWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB155Aegis: "Glaring Aegis", oldB155Rage: "Maniacal Rage", oldB155Indenture: "Unholy Indenture",
		oldB155Gauntlets: "Gauntlets of Light", oldB155HardCover: "Hard Cover",
		oldB155Ajani: "Ajani, Inspiring Leader", oldB155VivienBow: "Vivien, Arkbow Ranger",
		oldB155VivienNat: "Vivien, Nature's Avenger", oldB155MuYanling: "Mu Yanling, Celestial Wind",
		oldB155ChandraAoF: "Chandra, Acolyte of Flame",
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

func oldB155Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// oldB155Aura puts an Aura permanent on the battlefield attached to host.
func oldB155Aura(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := effectstest.B12Push(g, owner, name, "Enchantment — Aura", oracle, 0, 0)
	effectstest.AttachForTest(t, g, aura, host)
	return aura
}

// --- Auras --------------------------------------------------------------------

func TestOldB155GlaringAegisPumpsTheHostAndTapsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB155Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Glaring Aegis", effectstest.AuraTypeLine, oldB155Aegis, []game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerTriggerTargets(t, g, me.ID, theirs)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.MkmB1PT(t, g, mine); p != 3 || tough != 5 {
		t.Errorf("the host is %d/%d, want 3/5", p, tough)
	}
	if c, ok := effectstest.BattlefieldCard(g, theirs); !ok || !c.Tapped {
		t.Errorf("their creature: present %v tapped %v, want tapped", ok, c.Tapped)
	}
}

func TestOldB155ManiacalRagePumpsAndStopsBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AuraCast(t, g, "Maniacal Rage", oldB155Rage, blocker)
	if p, tough := effectstest.MkmB1PT(t, g, blocker); p != 4 || tough != 4 {
		t.Errorf("the host is %d/%d, want 4/4", p, tough)
	}
	effectstest.AttackIntoBlocks(t, g, them.ID, attacker)
	if effectstest.BlockRefusal(t, g, blocker, attacker) == "" {
		t.Error("the enchanted creature blocked")
	}
}

func TestOldB155UnholyIndentureReturnsTheCreatureUnderYourControlWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := oldB155Seats(g)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oldB155Aura(t, g, me.ID, "Unholy Indenture", oldB155Indenture, victim)
	effectstest.KillCreature(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, victim)
	if !ok {
		t.Fatal("the creature did not return")
	}
	if c.Controller != me.ID {
		t.Errorf("returned under %v's control, want mine", c.Controller)
	}
	if got := effectstest.CounterCount(g, victim, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestOldB155UnholyIndentureDoesNotReturnAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB155Seats(g)
	tok := effectstest.PushToken(g, me.ID, game.Card{Name: "Soldier", TypeLine: "Token Creature — Soldier", Power: 1, Toughness: 1})
	oldB155Aura(t, g, me.ID, "Unholy Indenture", oldB155Indenture, tok)
	effectstest.KillCreature(t, g, me.ID, tok)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, tok); ok {
		t.Error("a token came back")
	}
}

func TestOldB155GauntletsAssignsDamageByToughnessAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB155Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Glass Cannon", "Creature — Bear", 5, 1)
	oldB155Aura(t, g, me.ID, "Gauntlets of Light", oldB155Gauntlets, bear)
	if p, tough := effectstest.MkmB1PT(t, g, bear); p != 5 || tough != 3 {
		t.Fatalf("the host is %d/%d, want 5/3", p, tough)
	}
	c, _ := effectstest.BattlefieldCard(g, bear)
	if !c.Effective().DamageByToughnessAlways {
		t.Error("the host does not assign combat damage by toughness")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	effectstest.FloatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {2}{W} for two mana")
	}
	effectstest.FloatForTest(g, me, "W")
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the granted untap: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); c.Tapped {
		t.Error("the creature stayed tapped")
	}
}

func TestOldB155HardCoverPumpsAndGrantsALoot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := oldB155Seats(g)
	effectstest.Msc18Library(me, 5)
	bear := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	oldB155Aura(t, g, me.ID, "Hard Cover", oldB155HardCover, bear)
	if p, tough := effectstest.MkmB1PT(t, g, bear); p != 2 || tough != 4 {
		t.Errorf("the host is %d/%d, want 2/4", p, tough)
	}
	effectstest.S58HandCards(me, 1)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the granted loot: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d, want the same (drew one, discarded one)", hand, got)
	}
	if c, _ := effectstest.BattlefieldCard(g, bear); !c.Tapped {
		t.Error("the loot did not tap the creature")
	}
}

// --- Ajani, Inspiring Leader --------------------------------------------------

func TestOldB155AjaniPlusTwoGainsLifeAndCountersUpToOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB155Seats(g)
	ajani := effectstest.S82Walker(g, me.ID, "Ajani, Inspiring Leader", oldB155Ajani, 5)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.S82Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters on the Bear, want 2", got)
	}
	if got := effectstest.LoyaltyOf(g, ajani); got != 7 {
		t.Errorf("loyalty %d, want 7", got)
	}
}

func TestOldB155AjaniPlusTwoWithNoTargetStillGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB155Seats(g)
	ajani := effectstest.S82Walker(g, me.ID, "Ajani, Inspiring Leader", oldB155Ajani, 5)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	effectstest.S82Activate(t, g, me.ID, ajani, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
}

func TestOldB155AjaniMinusThreeExilesAndTheControllerGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB155Seats(g)
	ajani := effectstest.S82Walker(g, me.ID, "Ajani, Inspiring Leader", oldB155Ajani, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	life, myLife := opp.Life, me.Life
	effectstest.S82Activate(t, g, me.ID, ajani, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, theirs) {
		t.Error("the creature was not exiled")
	}
	if opp.Life != life+2 || me.Life != myLife {
		t.Errorf("their life %d → %d (want +2), mine %d → %d (want unchanged)", life, opp.Life, myLife, me.Life)
	}
}

func TestOldB155AjaniUltimateGivesFlyingAndDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB155Seats(g)
	ajani := effectstest.S82Walker(g, me.ID, "Ajani, Inspiring Leader", oldB155Ajani, 10)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, ajani, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"flying", "double strike"} {
		if !effectstest.HasEffectiveKeyword(t, g, mine, kw) {
			t.Errorf("my creature lacks %s", kw)
		}
		if effectstest.HasEffectiveKeyword(t, g, theirs, kw) {
			t.Errorf("their creature has %s", kw)
		}
	}
}

// --- The Viviens --------------------------------------------------------------

func TestOldB155VivienArkbowPlusOneDistributesCountersAndGivesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB155Seats(g)
	viv := effectstest.S82Walker(g, me.ID, "Vivien, Arkbow Ranger", oldB155VivienBow, 4)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, viv, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("%d counters on a target, want 1", got)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
			t.Error("a target has no trample")
		}
	}
}

func TestOldB155VivienArkbowPlusOneOnOneTargetPutsBothCountersOnIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB155Seats(g)
	viv := effectstest.S82Walker(g, me.ID, "Vivien, Arkbow Ranger", oldB155VivienBow, 4)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, viv, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(a), Distribution: map[uuid.UUID]int{a: 2},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, a, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2", got)
	}
}

func TestOldB155VivienArkbowMinusThreeBitesACreatureOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB155Seats(g)
	viv := effectstest.S82Walker(g, me.ID, "Vivien, Arkbow Ranger", oldB155VivienBow, 4)
	biter := effectstest.B12Creature(g, me.ID, "Biter", "Creature — Wolf", 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, viv, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(biter, victim)})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("the victim survived 3 damage")
	}
	if c, ok := effectstest.BattlefieldCard(g, biter); !ok || c.DamageMarked != 0 {
		t.Errorf("the biter took damage back: present %v, damage %d", ok, c.DamageMarked)
	}
	if got := effectstest.LoyaltyOf(g, viv); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

func TestOldB155VivienArkbowMinusFiveOffersOnlyCreatureCardsFromOutsideTheGame(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2,
		effectstest.OutsideCard("Grizzly Bears", "Creature — Bear"), effectstest.OutsideCard("Lightning Bolt", "Instant"))
	me := g.Seats[0]
	viv := effectstest.S82Walker(g, me.ID, "Vivien, Arkbow Ranger", oldB155VivienBow, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, viv, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no outside-the-game pick: %+v", g.PendingChoices)
	}
	if got := effectstest.CandidateNames(g, c); !effectstest.EqualNames(got, []string{"Grizzly Bears"}) {
		t.Errorf("candidates %v, want only the creature card", got)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{effectstest.SideCardID(me, "Grizzly Bears")}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OutsideHandHas(me, "Grizzly Bears") {
		t.Error("the creature card did not come into my hand")
	}
}

func TestOldB155VivienNaturesAvengerPlusOnePutsThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB155Seats(g)
	viv := effectstest.S82Walker(g, me.ID, "Vivien, Nature's Avenger", oldB155VivienNat, 3)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, viv, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(a)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, a, game.CounterPlusOne); got != 3 {
		t.Errorf("%d counters, want 3", got)
	}
}

func TestOldB155VivienNaturesAvengerMinusSixPumpsAndGivesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB155Seats(g)
	viv := effectstest.S82Walker(g, me.ID, "Vivien, Nature's Avenger", oldB155VivienNat, 6)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, viv, 2, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(a)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmB1PT(t, g, a); p != 12 || tough != 12 {
		t.Errorf("the target is %d/%d, want 12/12", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, a, "trample") {
		t.Error("the target has no trample")
	}
}

func TestOldB155VivienNaturesAvengerMinusOneRevealsToTheFirstCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB155Seats(g)
	viv := effectstest.S82Walker(g, me.ID, "Vivien, Nature's Avenger", oldB155VivienNat, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc18Library(me, 3)
	bear := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: bear, Name: "Library Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top Land", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	})
	libBefore := me.Library.Size()
	effectstest.S82Activate(t, g, me.ID, viv, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHand(me, bear) {
		t.Error("the creature card is not in my hand")
	}
	if got := me.Library.Size(); got != libBefore-1 {
		t.Errorf("library %d → %d, want one card fewer (the rest went to the bottom)", libBefore, got)
	}
}

// --- Mu Yanling, Celestial Wind -----------------------------------------------

func TestOldB155MuYanlingPlusOneShrinksUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB155Seats(g)
	mu := effectstest.S82Walker(g, me.ID, "Mu Yanling, Celestial Wind", oldB155MuYanling, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 6, 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, mu, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.CurrentPower(t, g, theirs); p != 1 {
		t.Errorf("their creature has power %d, want 1", p)
	}
}

func TestOldB155MuYanlingMinusThreeBouncesUpToTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB155Seats(g)
	mu := effectstest.S82Walker(g, me.ID, "Mu Yanling, Celestial Wind", oldB155MuYanling, 5)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, mu, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(a, b)})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if _, ok := effectstest.BattlefieldCard(g, id); ok {
			t.Error("a targeted creature stayed on the battlefield")
		}
	}
	if _, ok := effectstest.BattlefieldCard(g, c); !ok {
		t.Error("an untargeted creature was bounced")
	}
}

func TestOldB155MuYanlingUltimatePumpsYourFliersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB155Seats(g)
	mu := effectstest.S82Walker(g, me.ID, "Mu Yanling, Celestial Wind", oldB155MuYanling, 7)
	flier := effectstest.B12Push(g, me.ID, "Flier", "Creature — Bird", "", 1, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, flier).Keywords = []string{"flying"} })
	ground := effectstest.B12Creature(g, me.ID, "Ground", "Creature — Bear", 2, 2)
	theirFlier := effectstest.B12Push(g, opp.ID, "Their Flier", "Creature — Bird", "", 1, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, theirFlier).Keywords = []string{"flying"} })
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, mu, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmB1PT(t, g, flier); p != 6 || tough != 6 {
		t.Errorf("my flier is %d/%d, want 6/6", p, tough)
	}
	if p, _ := effectstest.MkmB1PT(t, g, ground); p != 2 {
		t.Errorf("my ground creature has power %d, want 2", p)
	}
	if p, _ := effectstest.MkmB1PT(t, g, theirFlier); p != 1 {
		t.Errorf("their flier has power %d, want 1", p)
	}
}

// --- Chandra, Acolyte of Flame ------------------------------------------------

func oldB155RedWalker(g *game.Game, owner uuid.UUID, name string, loyalty int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Legendary Planeswalker — Test", ManaCost: "{2}{R}{R}", Colors: []string{"R"},
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
	return id
}

func TestOldB155ChandraFirstZeroAddsLoyaltyToEachRedPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB155Seats(g)
	chandra := effectstest.S82Walker(g, me.ID, "Chandra, Acolyte of Flame", oldB155ChandraAoF, 4)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, chandra).Colors = []string{"R"} })
	red := oldB155RedWalker(g, me.ID, "Red Walker", 3)
	blue := effectstest.S82Walker(g, me.ID, "Blue Walker", "", 3)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, blue).Colors = []string{"U"} })
	theirRed := oldB155RedWalker(g, opp.ID, "Their Red Walker", 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, chandra, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, chandra); got != 5 {
		t.Errorf("Chandra has loyalty %d, want 5", got)
	}
	if got := effectstest.LoyaltyOf(g, red); got != 4 {
		t.Errorf("my red walker has loyalty %d, want 4", got)
	}
	if got := effectstest.LoyaltyOf(g, blue); got != 3 {
		t.Errorf("my blue walker has loyalty %d, want 3", got)
	}
	if got := effectstest.LoyaltyOf(g, theirRed); got != 3 {
		t.Errorf("their red walker has loyalty %d, want 3", got)
	}
}

func TestOldB155ChandraSecondZeroMakesHastyElementalsThatAreSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB155Seats(g)
	chandra := effectstest.S82Walker(g, me.ID, "Chandra, Acolyte of Flame", oldB155ChandraAoF, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, chandra, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	toks := effectstest.S46Tokens(g, me.ID, "Elemental")
	if len(toks) != 2 {
		t.Fatalf("%d Elementals, want 2", len(toks))
	}
	for _, tok := range toks {
		if !effectstest.HasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
			t.Error("an Elemental has no haste")
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.S46Tokens(g, me.ID, "Elemental")); got != 0 {
		t.Errorf("%d Elementals left after the end step, want 0", got)
	}
}

func TestOldB155ChandraMinusTwoCastsACheapSpellFromTheGraveyardAndExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB155Seats(g)
	chandra := effectstest.S82Walker(g, me.ID, "Chandra, Acolyte of Flame", oldB155ChandraAoF, 4)
	bolt := uuid.New()
	pricey := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Test Bolt", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: pricey, Name: "Test Wrath", TypeLine: "Sorcery", ManaCost: "{2}{R}{R}", Owner: me.ID, Controller: me.ID})
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, chandra, 2, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(pricey)}); err == nil {
		t.Fatal("targeted a mana value 4 card")
	}
	effectstest.S82Activate(t, g, me.ID, chandra, 2, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bolt)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatForTest(g, me, "R")
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, bolt)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, bolt, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast Test Bolt from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, bolt) {
		t.Error("the spell cast from the graveyard did not end up in exile")
	}
	if effectstest.InGraveyard(me, bolt) {
		t.Error("the spell is still in the graveyard")
	}
	_ = opp
}
