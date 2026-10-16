package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// top1000_sweep_f_test.go — the sixth slice of the missing top-1,000
// cards (#3654).

const (
	sweepFAetherfluxOracle   = "5f15125e-ae89-471e-b055-226a783bc7c1"
	sweepFKutzilOracle       = "3233af43-7826-4895-b67e-03c6102c2cd5"
	sweepFWishclawOracle     = "81c70ae7-3c18-4c9b-8505-e4db9e0e6518"
	sweepFMindbreakOracle    = "48712d36-5c13-4ded-87a4-68106da4a5cd"
	sweepFCaretakerOracle    = "c97957a2-8310-4cff-8aad-871b7901d124"
	sweepFInnkeeperOracle    = "915b75d1-b5e3-4aca-a54c-beaccef0a2e4"
	sweepFSpringheartOracle  = "8a3ad2ef-8bcb-40c0-85de-f03328c2b644"
	sweepFVeyranOracle       = "d57a6b19-bd06-4b07-8ee1-0b45612ba9e4"
	sweepFIsochronOracle     = "ddd649f4-dbaf-48ec-8ff7-95581257772d"
	sweepFHighFaeOracle      = "28c1a4a0-f58c-4d07-a918-51ff8a74f050"
	sweepFExpressiveOracle   = "c7aecca5-2f67-4245-ab2d-e723d8b23a67"
	sweepFTaintedPactOracle  = "1a85ba2b-ae10-4917-954a-7709b75a9740"
	sweepFBladeOracle        = "9c329f3d-4d0a-4012-8aa6-92afdd120a52"
	sweepFEvendoOracle       = "83161d59-2520-4741-9328-e2a4a8b5d5bc"
	sweepFConsultationOracle = "9a1412db-45ad-46ea-8f12-a85d203113d8"
	sweepFApproachOracle     = "e4125377-34c0-4b54-bdf8-4e88f5d24565"
	sweepFCyberdriveOracle   = "126a5d0b-e7d6-458e-87af-47d845678014"
	sweepFArchmageOracle     = "8305d576-21d8-4ce7-8eda-a7cd9793aca5"
)

// sweepFTop puts a card on TOP of p's library.
func sweepFTop(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// sweepFHand puts a card with a mana cost in p's hand.
func sweepFHand(p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return id
}

func sweepFAddMana(g *game.Game, who uuid.UUID, mana string) {
	g.WithWriteLock(func() { _ = g.AddManaForEffect(who, uuid.Nil, mana) })
}

// sweepFSettle passes priority and answers trigger-order prompts until
// the stack is empty.
func sweepFSettle(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		passPriorityAroundTable(t, g)
		if stackFullyEmpty(g) {
			return
		}
		if !answerAnyTriggerOrderPrompt(t, g, who) {
			return
		}
	}
}

// --- Aetherflux Reservoir ---------------------------------------------

func TestSweepFAetherfluxReservoirGainsForEachSpellCastThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Aetherflux Reservoir", "Artifact", sweepFAetherfluxOracle, 0, 0)
	before := me.Life

	castCatalogSpell(t, g, "First Spell", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Fatalf("first spell: life %d, want %d (one spell cast this turn)", me.Life, before+1)
	}
	castCatalogSpell(t, g, "Second Spell", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("second spell: life %d, want %d (1 then 2)", me.Life, before+3)
	}
}

func TestSweepFAetherfluxReservoirPaysFiftyLifeForFiftyDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	res := b12Push(g, me.ID, "Aetherflux Reservoir", "Artifact", sweepFAetherfluxOracle, 0, 0)
	target := game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}

	me.Life = 49
	if err := g.ActivateCatalogAbility(me.ID, res, 0, target); err == nil {
		t.Fatal("activated with 49 life: 50 life is a cost it can't pay (CR 119.4)")
	}
	me.Life = 60
	oppBefore := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, res, 0, target); err != nil {
		t.Fatalf("activate with 60 life: %v", err)
	}
	if me.Life != 10 {
		t.Errorf("life after paying = %d, want 10", me.Life)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != oppBefore-50 {
		t.Errorf("opponent life %d → %d, want 50 damage", oppBefore, opp.Life)
	}
}

// --- Kutzil, Malamet Exemplar ------------------------------------------

func TestSweepFKutzilDrawsForACreatureAboveItsBasePower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Kutzil, Malamet Exemplar", "Legendary Creature — Cat Warrior", sweepFKutzilOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	hand := me.Hand.Size()

	dealCombatDamageToPlayer(g, bear, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d: a 3/3 with base power 2 connected, want one draw", hand, me.Hand.Size())
	}
}

func TestSweepFKutzilIgnoresACreatureAtItsBasePower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Kutzil, Malamet Exemplar", "Legendary Creature — Cat Warrior", sweepFKutzilOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// Power SET to 5 is base power 5: not greater than its base power.
	g.WithWriteLock(func() {
		_ = g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(bear),
			[]game.Mod{game.SetBasePowerMod(5)}, game.IndefiniteDuration(), "test")
	})
	hand := me.Hand.Size()

	dealCombatDamageToPlayer(g, bear, opp.ID, 5)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d: a creature whose power was only set connected, want no draw", hand, me.Hand.Size())
	}
}

func TestSweepFKutzilStopsOpponentsCastingOnYourTurn(t *testing.T) {
	spec, ok := Lookup(sweepFKutzilOracle)
	if !ok {
		t.Fatal("Kutzil is not registered")
	}
	if len(spec.CastRestrictions) != 1 {
		t.Errorf("cast restrictions = %d, want the one during-your-turn ban", len(spec.CastRestrictions))
	}
}

// --- Wishclaw Talisman -------------------------------------------------

func TestSweepFWishclawTalismanTutorsThenChangesHands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	tal := castCatalogSpell(t, g, "Wishclaw Talisman", "Artifact", sweepFWishclawOracle, nil)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, tal, "wish"); got != 3 {
		t.Fatalf("wish counters on entry = %d, want 3", got)
	}
	needle := pushLibraryCardForTest(me, game.Card{Name: "Needle", TypeLine: "Sorcery"})

	sweepFAddMana(g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, tal, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := counterCount(g, tal, "wish"); got != 2 {
		t.Errorf("wish counters after paying = %d, want 2", got)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, needle)
	answerChoosePlayer(t, g, me.ID, opp)
	if !inHandOf(me, needle) {
		t.Error("the searched card is not in hand")
	}
	if c := layeredCard(t, g, tal); c.Controller != opp.ID {
		t.Errorf("Talisman controller = %v, want the chosen opponent", c.Controller)
	}
}

func TestSweepFWishclawTalismanOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	tal := b12Push(g, opp.ID, "Wishclaw Talisman", "Artifact", sweepFWishclawOracle, 0, 0)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(tal, "wish", 3) })
	sweepFAddMana(g, opp.ID, "{C}")
	if err := g.ActivateCatalogAbility(opp.ID, tal, 0, game.ActivateAbilityParams{}); err == nil {
		t.Errorf("activated on %s's turn by its non-active controller", me.Name)
	}
}

// --- Mindbreak Trap ----------------------------------------------------

func TestSweepFMindbreakTrapExilesTheTargetSpells(t *testing.T) {
	g := newCatalogGame(t)
	a := castCatalogSpell(t, g, "Spell A", "Instant", "", nil)
	b := castCatalogSpell(t, g, "Spell B", "Instant", "", nil)
	castCatalogSpell(t, g, "Mindbreak Trap", "Instant — Trap", sweepFMindbreakOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	if !inExile(g, a) || !inExile(g, b) {
		t.Errorf("exiled A=%v B=%v, want both spells exiled", inExile(g, a), inExile(g, b))
	}
}

func TestSweepFMindbreakTrapCostNeedsOneOpponentWithThreeSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	check := func() bool {
		var ok bool
		g.ReadSnapshot(func() { ok = anOpponentCastThreeOrMoreSpellsThisTurn(g, me.ID) })
		return ok
	}
	if check() {
		t.Fatal("offered with no spells cast")
	}
	g.WithWriteLock(func() {
		g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{
			me.ID: {Total: 5}, opp.ID: {Total: 2}, opp2.ID: {Total: 2},
		}
	})
	if check() {
		t.Error("offered with no single opponent at three (the 2025-09-19 ruling)")
	}
	g.WithWriteLock(func() { g.SpellsCastThisTurn[opp2.ID] = game.CastTally{Total: 3} })
	if !check() {
		t.Error("not offered once an opponent cast three spells")
	}
}

// --- Caretaker's Talent ------------------------------------------------

func TestSweepFCaretakersTalentDrawsOnceAndLevelsUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	talent := pushClass(g, me.ID, "Caretaker's Talent", "Enchantment — Class", sweepFCaretakerOracle)
	toMainForCost(t, g)
	hand := me.Hand.Size()

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 green Bear"), 2) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d: two tokens entering together draw one card", hand, me.Hand.Size())
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 green Bear"), 1) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d: it triggers only once each turn", hand, me.Hand.Size())
	}

	toMainForCost(t, g)
	if err := levelUpTo(t, g, me.ID, talent, 0, "{W}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	answerAnyPendingTargetPrompts(t, g)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Bear"); got != 4 {
		t.Fatalf("%d Bears after level 2, want 4 (a token copy of a token)", got)
	}

	if err := levelUpTo(t, g, me.ID, talent, 1, "{3}{W}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && c.Controller == me.ID {
			if p := layeredCard(t, g, c.InstanceID).CurrentPower(); p != 4 {
				t.Errorf("a Bear token is power %d at level 3, want 4", p)
			}
		}
	}
}

// --- Innkeeper's Talent ------------------------------------------------

func TestSweepFInnkeepersTalentLevelThreeDoublesYourCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	talent := pushClass(g, me.ID, "Innkeeper's Talent", "Enchantment — Class", sweepFInnkeeperOracle)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	toMainForCost(t, g)
	if err := levelUpTo(t, g, me.ID, talent, 0, "{G}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mine, game.CounterPlusOne, 1) })
	if got := counterCount(g, mine, game.CounterPlusOne); got != 1 {
		t.Fatalf("level 2 doubled a counter: %d", got)
	}
	if err := levelUpTo(t, g, me.ID, talent, 1, "{3}{G}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mine, game.CounterPlusOne, 1) })
	if got := counterCount(g, mine, game.CounterPlusOne); got != 3 {
		t.Errorf("level 3: %d counters, want 1 + 2", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 1) })
	if got := counterCount(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("a counter the opponent's permanent got by its controller's placement: %d, want 1", got)
	}
}

func TestSweepFInnkeepersTalentPutsACounterAtCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushClass(g, me.ID, "Innkeeper's Talent", "Enchantment — Class", sweepFInnkeeperOracle)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	answerAnyPendingTargetPrompts(t, g)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters at the beginning of combat = %d, want 1", got)
	}
}

// --- Springheart Nantuko -----------------------------------------------

// sweepFBestowed puts the Nantuko onto the battlefield as an Aura
// attached to host.
func sweepFBestowed(g *game.Game, owner, host uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Springheart Nantuko", TypeLine: "Enchantment — Aura",
		OracleID: sweepFSpringheartOracle, Owner: owner, Controller: owner,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host},
	})
}

func TestSweepFSpringheartNantukoCopiesTheEnchantedCreatureWhenPaid(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sweepFBestowed(g, me.ID, bear)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	sweepFAddMana(g, me.ID, "{C}{G}")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Bear"); got != 2 {
		t.Errorf("%d Bears, want the original and a token copy", got)
	}
	if got := countBattlefieldNamed(g, me.ID, "Insect"); got != 0 {
		t.Errorf("%d Insects, want none once a copy was made", got)
	}
}

func TestSweepFSpringheartNantukoMakesAnInsectWhenDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sweepFBestowed(g, me.ID, bear)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Bear"); got != 1 {
		t.Errorf("%d Bears, want no copy", got)
	}
	if got := countBattlefieldNamed(g, me.ID, "Insect"); got != 1 {
		t.Errorf("%d Insects, want one", got)
	}
}

func TestSweepFSpringheartNantukoUnattachedMakesAnInsect(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Springheart Nantuko", "Enchantment Creature — Insect Monk", sweepFSpringheartOracle, 1, 1)
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Insect"); got != 1 {
		t.Errorf("%d Insects, want one with no prompt", got)
	}
}

// --- Veyran, Voice of Duality -----------------------------------------

func TestSweepFVeyranDoublesMagecraft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	veyran := b12Push(g, me.ID, "Veyran, Voice of Duality", "Legendary Creature — Efreet Wizard", sweepFVeyranOracle, 2, 2)
	b12Push(g, me.ID, "Archmage Emeritus", "Creature — Human Wizard", sweepFArchmageOracle, 2, 2)
	toMainForCost(t, g)
	hand := me.Hand.Size()

	castCatalogSpell(t, g, "Spell", "Instant", "", nil)
	sweepFSettle(t, g, me.ID)
	// castCatalogSpell put the spell in hand first, so the hand is back
	// to `hand` before the draws.
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2: Archmage Emeritus's magecraft triggers twice", got)
	}
	if p := layeredCard(t, g, veyran).CurrentPower(); p != 4 {
		t.Errorf("Veyran power %d, want 4: its own magecraft triggers twice", p)
	}
}

// --- Isochron Scepter --------------------------------------------------

func TestSweepFIsochronScepterImprintsAndCastsACopyFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := sweepFHand(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	big := sweepFHand(me, "Big Instant", "Instant", "{2}{U}{U}", "")
	scepter := castCatalogSpell(t, g, "Isochron Scepter", "Artifact", sweepFIsochronOracle, nil)
	passPriorityAroundTable(t, g)

	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no imprint prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.ChooseCards {
		offered[id] = true
	}
	if !offered[bolt] || offered[big] {
		t.Errorf("imprint offers %v, want the mana value 1 instant and not the 4", c.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, bolt)
	if !inExile(g, bolt) {
		t.Fatal("the imprinted card is not in exile")
	}

	sweepFAddMana(g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, scepter, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	cp := copyOf(t, g, "Lightning Bolt")
	before := opp.Life
	if err := castCopy(g, me.ID, cp.InstanceID, true, playerTarget(opp.ID)); err != nil {
		t.Fatalf("casting the copy free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("opponent life %d → %d, want the copy to deal 3", before, opp.Life)
	}
	if !inExile(g, bolt) {
		t.Error("the imprinted card left exile")
	}
}

// --- High Fae Trickster ------------------------------------------------

func TestSweepFHighFaeTricksterLetsYouCastAsThoughFlash(t *testing.T) {
	spec, ok := Lookup(sweepFHighFaeOracle)
	if !ok {
		t.Fatal("High Fae Trickster is not registered")
	}
	if len(spec.CastTimings) != 1 || spec.CastTimings[0].Timing != game.TimingFlash {
		t.Errorf("cast timings %+v, want one flash rule", spec.CastTimings)
	}
	kw := map[string]bool{}
	for _, k := range spec.PrintedKeywords {
		kw[k] = true
	}
	if !kw["flash"] || !kw["flying"] {
		t.Errorf("keywords %v, want flash and flying", spec.PrintedKeywords)
	}
}

// --- Expressive Iteration ----------------------------------------------

func TestSweepFExpressiveIterationHandBottomExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	third := sweepFTop(me, "Third", "Sorcery")
	second := sweepFTop(me, "Second", "Sorcery")
	first := sweepFTop(me, "First", "Sorcery")

	castCatalogSpell(t, g, "Expressive Iteration", "Sorcery", sweepFExpressiveOracle, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, second)
	answerChooseCards(t, g, me.ID, first)
	passPriorityAroundTable(t, g)

	if !inHandOf(me, second) {
		t.Error("the card chosen for the hand is not there")
	}
	if me.Library.Cards[0].InstanceID != first {
		t.Error("the card chosen for the bottom is not on the bottom")
	}
	if !inExile(g, third) {
		t.Fatal("the third card is not exiled")
	}
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(third)
		if perm := g.CastPermissionOnCardForEffect(c, game.ZoneExile); perm == nil || perm.Player != me.ID {
			t.Errorf("no play permission over the exiled card: %+v", perm)
		}
	})
}

func TestSweepFExpressiveIterationTwoCardsExileNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	second := sweepFTop(me, "Second", "Sorcery")
	first := sweepFTop(me, "First", "Sorcery")

	castCatalogSpell(t, g, "Expressive Iteration", "Sorcery", sweepFExpressiveOracle, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, first)
	passPriorityAroundTable(t, g)
	if !inHandOf(me, first) || inExile(g, second) || !me.Library.Contains(second) {
		t.Errorf("hand=%v exiled=%v library=%v, want one in hand and the other on the bottom",
			inHandOf(me, first), inExile(g, second), me.Library.Contains(second))
	}
}

// --- Tainted Pact ------------------------------------------------------

func TestSweepFTaintedPactStopsOnTheSecondCardWithAName(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alpha2 := sweepFTop(me, "Alpha", "Sorcery")
	beta := sweepFTop(me, "Beta", "Sorcery")
	alpha1 := sweepFTop(me, "Alpha", "Sorcery")
	under := me.Library.Cards[len(me.Library.Cards)-4].InstanceID

	castCatalogSpell(t, g, "Tainted Pact", "Instant", sweepFTaintedPactOracle, nil)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false) // Alpha
	answerMayChoice(t, g, me.ID, false) // Beta
	if latestConfirmFor(g, me.ID) != nil {
		t.Fatal("asked about a second Alpha: the process ends there")
	}
	for _, id := range []uuid.UUID{alpha1, beta, alpha2} {
		if !inExile(g, id) {
			t.Errorf("card %s not exiled", id)
		}
	}
	if !me.Library.Contains(under) {
		t.Error("exiled past the duplicate")
	}
}

func TestSweepFTaintedPactTakesTheChosenCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gamma := sweepFTop(me, "Gamma", "Sorcery")
	beta := sweepFTop(me, "Beta", "Sorcery")
	alpha := sweepFTop(me, "Alpha", "Sorcery")

	castCatalogSpell(t, g, "Tainted Pact", "Instant", sweepFTaintedPactOracle, nil)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false) // Alpha
	answerMayChoice(t, g, me.ID, true)  // Beta
	if !inExile(g, alpha) || !inHandOf(me, beta) || !me.Library.Contains(gamma) {
		t.Errorf("alpha exiled=%v beta in hand=%v gamma in library=%v",
			inExile(g, alpha), inHandOf(me, beta), me.Library.Contains(gamma))
	}
}

// --- Blade of Selves ---------------------------------------------------

func TestSweepFBladeOfSelvesGivesTheEquippedCreatureMyriad(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	blade := b12Push(g, me.ID, "Blade of Selves", "Artifact — Equipment", sweepFBladeOracle, 0, 0)
	knight := pushTribalCreature(g, me.ID, "Knight", "Creature — Human Knight", 3, 3, "vigilance")
	toMainForCost(t, g)
	sweepFAddMana(g, me.ID, "{C}{C}{C}{C}")
	if err := activateOn(t, g, me.ID, blade, 0, knight); err != nil {
		t.Fatalf("equip: %v", err)
	}
	passPriorityAroundTable(t, g)

	declareAttack(t, g, opp1.ID, knight)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	answerMayChoice(t, g, me.ID, true)
	attacking := map[uuid.UUID]int{}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Knight" && c.InstanceID != knight {
			attacking[c.AttackingTarget]++
		}
	}
	if attacking[opp2.ID] != 1 || attacking[opp3.ID] != 1 || attacking[opp1.ID] != 0 {
		t.Errorf("token copies attack %v, want one at each non-defending opponent", attacking)
	}
}

// --- Evendo, Waking Haven ----------------------------------------------

func TestSweepFEvendoTapsForEachCreatureAtTwelve(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	evendo := b12Push(g, me.ID, "Evendo, Waking Haven", "Land — Planet", sweepFEvendoOracle, 0, 0)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	b12Creature(g, me.ID, "Elk", "Creature — Elk", 2, 2)
	toMainForCost(t, g)

	sweepFAddMana(g, me.ID, "{G}")
	if err := g.ActivateManaAbility(me.ID, evendo, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the 12+ ability activated with no charge counters")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(evendo, game.CounterCharge, 12) })
	activateManaFor(t, g, me.ID, evendo, 1, game.ManaAbilityParams{})
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("pool %d after {G}, {T}: want three {G} for three creatures", got)
	}
}

// --- Demonic Consultation ----------------------------------------------

func TestSweepFDemonicConsultationFindsTheNamedCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	after := sweepFTop(me, "After", "Sorcery")
	wanted := sweepFTop(me, "Wanted", "Sorcery")
	other := sweepFTop(me, "Other", "Sorcery")
	var six []uuid.UUID
	for i := 0; i < 5; i++ {
		six = append(six, sweepFTop(me, "Junk", "Sorcery"))
	}
	// A Wanted among the top six is exiled with them, not found.
	six = append(six, sweepFTop(me, "Wanted", "Sorcery"))

	castCatalogSpell(t, g, "Demonic Consultation", "Instant", sweepFConsultationOracle, nil)
	passPriorityAroundTable(t, g)
	actNameFor(t, g, me.ID, "Wanted")
	passPriorityAroundTable(t, g)

	for _, id := range six {
		if !inExile(g, id) {
			t.Errorf("a top-six card %s was not exiled", id)
		}
	}
	if !inExile(g, other) {
		t.Error("a revealed card that was not named was not exiled")
	}
	if !inHandOf(me, wanted) {
		t.Error("the named card is not in hand")
	}
	if !me.Library.Contains(after) {
		t.Error("a card below the named one left the library")
	}
}

// --- Approach of the Second Sun ----------------------------------------

func TestSweepFApproachGoesSeventhThenWins(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life

	first := castCatalogSpell(t, g, "Approach of the Second Sun", "Sorcery", sweepFApproachOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != before+7 {
		t.Errorf("life %d, want %d", me.Life, before+7)
	}
	n := len(me.Library.Cards)
	if n < 7 || me.Library.Cards[n-7].InstanceID != first {
		t.Fatal("Approach is not seventh from the top")
	}
	if g.Outcome != nil {
		t.Fatal("the first Approach won the game")
	}

	castCatalogSpell(t, g, "Approach of the Second Sun", "Sorcery", sweepFApproachOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Errorf("outcome %+v, want the caster to win with the second Approach", g.Outcome)
	}
}

// --- Cyberdrive Awakener -----------------------------------------------

func TestSweepFCyberdriveAwakenerAnimatesNoncreatureArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID,
	})
	castCatalogSpell(t, g, "Cyberdrive Awakener", "Artifact Creature — Construct", sweepFCyberdriveOracle, nil)
	passPriorityAroundTable(t, g)

	c := layeredCard(t, g, rock)
	if !c.IsCreature() || c.CurrentPower() != 4 {
		t.Errorf("Rock creature=%v power=%d, want a 4/4 artifact creature", c.IsCreature(), c.CurrentPower())
	}
	if !game.HasKeyword(&c, "flying") {
		t.Error("the animated artifact does not have flying from the static")
	}

	late := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Late Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID,
	})
	if layeredCard(t, g, late).IsCreature() {
		t.Error("an artifact that arrived after the trigger was animated (CR 611.2c)")
	}
}

// --- Etali, Primal Conqueror -------------------------------------------

func etaliRow() cards.Card {
	return transformRow(etaliOracleID,
		"Etali, Primal Conqueror", "Legendary Creature — Elder Dinosaur", "{5}{R}{R}",
		"Etali, Primal Sickness", "Legendary Creature — Phyrexian Elder Dinosaur", "11", "11", []string{"R"})
}

func TestSweepFEtaliExilesUntilNonlandFromEachLibraryAndOffersThemFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var lands, spells []uuid.UUID
	for _, p := range g.Seats {
		spells = append(spells, sweepFTop(p, "Spell of "+p.Name, "Sorcery"))
		lands = append(lands, sweepFTop(p, "Land of "+p.Name, "Land"))
	}
	importToBattlefield(t, g, etaliRow(), me)
	passPriorityAroundTable(t, g)

	for i := range g.Seats {
		if !inExile(g, lands[i]) || !inExile(g, spells[i]) {
			t.Errorf("seat %d: land exiled=%v spell exiled=%v, want both", i, inExile(g, lands[i]), inExile(g, spells[i]))
		}
	}
	g.ReadSnapshot(func() {
		for _, id := range spells {
			c, _ := g.LookupCardForEffect(id)
			perm := g.CastPermissionOnCardForEffect(c, game.ZoneExile)
			if perm == nil || perm.Player != me.ID || perm.Cost != "{0}" {
				t.Errorf("%s: permission %+v, want a free cast for Etali's controller", c.Name, perm)
			}
		}
		for _, id := range lands {
			c, _ := g.LookupCardForEffect(id)
			if perm := g.CastPermissionOnCardForEffect(c, game.ZoneExile); perm != nil {
				t.Errorf("%s: a land was offered: %+v", c.Name, perm)
			}
		}
	})
}

func TestSweepFEtaliTransformsAndPoisons(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	etali := importToBattlefield(t, g, etaliRow(), me)
	passPriorityAroundTable(t, g)
	toMainForCost(t, g)
	sweepFAddMana(g, me.ID, "{C}{C}{C}{C}{C}{C}{C}{C}{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, etali, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("transform: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, etali); c.Name != "Etali, Primal Sickness" {
		t.Fatalf("face %q after the transform, want the back face", c.Name)
	}
	dealCombatDamageToPlayer(g, etali, opp.ID, 11)
	passPriorityAroundTable(t, g)
	if got := opp.Counters[game.CounterPoison]; got != 11 {
		t.Errorf("poison counters %d, want 11", got)
	}
}

// --- Ojer Taq, Deepest Foundation --------------------------------------

func ojerTaqRow() cards.Card {
	return transformRow(ojerTaqOracleID,
		"Ojer Taq, Deepest Foundation", "Legendary Creature — God", "{4}{W}{W}",
		"Temple of Civilization", "Land", "", "", []string{"W"})
}

func TestSweepFOjerTaqTriplesCreatureTokensOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Ojer Taq, Deepest Foundation", "Legendary Creature — God", ojerTaqOracleID, 6, 6)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 green Bear"), 1)
		_ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1)
	})
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Bear"); got != 3 {
		t.Errorf("%d Bears, want three times one", got)
	}
	if got := countBattlefieldNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasures, want one: a noncreature token is not tripled", got)
	}
}

func TestSweepFOjerTaqReturnsAsTempleOfCivilization(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	god := importToBattlefield(t, g, ojerTaqRow(), me)
	ojerKill(t, g, god)
	passPriorityAroundTable(t, g)
	temple := ojerOnBattlefield(g, ojerTaqOracleID)
	if temple == nil || temple.Name != "Temple of Civilization" || !temple.Tapped {
		t.Fatalf("returned %+v, want a tapped Temple of Civilization", temple)
	}
}
