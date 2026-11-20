package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsmi "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/mi"
	cardso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/o"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweepFHand puts a card with a mana cost in p's hand.
func sweepFHand(p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return id
}

// sweepFSettle passes priority and answers trigger-order prompts until
// the stack is empty.
func sweepFSettle(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.StackFullyEmpty(g) {
			return
		}
		if !effectstest.AnswerAnyTriggerOrderPrompt(t, g, who) {
			return
		}
	}
}

// --- Aetherflux Reservoir ---------------------------------------------

func TestSweepFAetherfluxReservoirGainsForEachSpellCastThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Aetherflux Reservoir", "Artifact", effectstest.SweepFAetherfluxOracle, 0, 0)
	before := me.Life

	effectstest.CastCatalogSpell(t, g, "First Spell", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Fatalf("first spell: life %d, want %d (one spell cast this turn)", me.Life, before+1)
	}
	effectstest.CastCatalogSpell(t, g, "Second Spell", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Errorf("second spell: life %d, want %d (1 then 2)", me.Life, before+3)
	}
}

func TestSweepFAetherfluxReservoirPaysFiftyLifeForFiftyDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	res := effectstest.B12Push(g, me.ID, "Aetherflux Reservoir", "Artifact", effectstest.SweepFAetherfluxOracle, 0, 0)
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
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppBefore-50 {
		t.Errorf("opponent life %d → %d, want 50 damage", oppBefore, opp.Life)
	}
}

// --- Kutzil, Malamet Exemplar ------------------------------------------

func TestSweepFKutzilDrawsForACreatureAboveItsBasePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Kutzil, Malamet Exemplar", "Legendary Creature — Cat Warrior", effectstest.SweepFKutzilOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	hand := me.Hand.Size()

	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d: a 3/3 with base power 2 connected, want one draw", hand, me.Hand.Size())
	}
}

func TestSweepFKutzilIgnoresACreatureAtItsBasePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Kutzil, Malamet Exemplar", "Legendary Creature — Cat Warrior", effectstest.SweepFKutzilOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// Power SET to 5 is base power 5: not greater than its base power.
	g.WithWriteLock(func() {
		_ = g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(bear),
			[]game.Mod{game.SetBasePowerMod(5)}, game.IndefiniteDuration(), "test")
	})
	hand := me.Hand.Size()

	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d: a creature whose power was only set connected, want no draw", hand, me.Hand.Size())
	}
}

func TestSweepFKutzilStopsOpponentsCastingOnYourTurn(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.SweepFKutzilOracle)
	if !ok {
		t.Fatal("Kutzil is not registered")
	}
	if len(spec.CastRestrictions) != 1 {
		t.Errorf("cast restrictions = %d, want the one during-your-turn ban", len(spec.CastRestrictions))
	}
}

// --- Wishclaw Talisman -------------------------------------------------

func TestSweepFWishclawTalismanTutorsThenChangesHands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	tal := effectstest.CastCatalogSpell(t, g, "Wishclaw Talisman", "Artifact", effectstest.SweepFWishclawOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, tal, "wish"); got != 3 {
		t.Fatalf("wish counters on entry = %d, want 3", got)
	}
	needle := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Needle", TypeLine: "Sorcery"})

	effectstest.SweepFAddMana(g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, tal, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := effectstest.CounterCount(g, tal, "wish"); got != 2 {
		t.Errorf("wish counters after paying = %d, want 2", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, needle)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	if !effectstest.InHandOf(me, needle) {
		t.Error("the searched card is not in hand")
	}
	if c := effectstest.LayeredCard(t, g, tal); c.Controller != opp.ID {
		t.Errorf("Talisman controller = %v, want the chosen opponent", c.Controller)
	}
}

func TestSweepFWishclawTalismanOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	tal := effectstest.B12Push(g, opp.ID, "Wishclaw Talisman", "Artifact", effectstest.SweepFWishclawOracle, 0, 0)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(tal, "wish", 3) })
	effectstest.SweepFAddMana(g, opp.ID, "{C}")
	if err := g.ActivateCatalogAbility(opp.ID, tal, 0, game.ActivateAbilityParams{}); err == nil {
		t.Errorf("activated on %s's turn by its non-active controller", me.Name)
	}
}

// --- Mindbreak Trap ----------------------------------------------------

func TestSweepFMindbreakTrapExilesTheTargetSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a := effectstest.CastCatalogSpell(t, g, "Spell A", "Instant", "", nil)
	b := effectstest.CastCatalogSpell(t, g, "Spell B", "Instant", "", nil)
	effectstest.CastCatalogSpell(t, g, "Mindbreak Trap", "Instant — Trap", effectstest.SweepFMindbreakOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, a) || !effectstest.InExile(g, b) {
		t.Errorf("exiled A=%v B=%v, want both spells exiled", effectstest.InExile(g, a), effectstest.InExile(g, b))
	}
}

func TestSweepFMindbreakTrapCostNeedsOneOpponentWithThreeSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	check := func() bool {
		var ok bool
		g.ReadSnapshot(func() { ok = cardsmi.AnOpponentCastThreeOrMoreSpellsThisTurn(g, me.ID) })
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	talent := effectstest.PushClass(g, me.ID, "Caretaker's Talent", "Enchantment — Class", effectstest.SweepFCaretakerOracle)
	effectstest.ToMainForCost(t, g)
	hand := me.Hand.Size()

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 green Bear"), 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d: two tokens entering together draw one card", hand, me.Hand.Size())
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 green Bear"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d: it triggers only once each turn", hand, me.Hand.Size())
	}

	effectstest.ToMainForCost(t, g)
	if err := effectstest.LevelUpTo(t, g, me.ID, talent, 0, "{W}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	effectstest.AnswerAnyPendingTargetPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Bear"); got != 4 {
		t.Fatalf("%d Bears after level 2, want 4 (a token copy of a token)", got)
	}

	if err := effectstest.LevelUpTo(t, g, me.ID, talent, 1, "{3}{W}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && c.Controller == me.ID {
			if p := effectstest.LayeredCard(t, g, c.InstanceID).CurrentPower(); p != 4 {
				t.Errorf("a Bear token is power %d at level 3, want 4", p)
			}
		}
	}
}

// --- Innkeeper's Talent ------------------------------------------------

func TestSweepFInnkeepersTalentLevelThreeDoublesYourCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	talent := effectstest.PushClass(g, me.ID, "Innkeeper's Talent", "Enchantment — Class", effectstest.SweepFInnkeeperOracle)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ox", "Creature — Ox", 2, 2)
	effectstest.ToMainForCost(t, g)
	if err := effectstest.LevelUpTo(t, g, me.ID, talent, 0, "{G}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mine, game.CounterPlusOne, 1) })
	if got := effectstest.CounterCount(g, mine, game.CounterPlusOne); got != 1 {
		t.Fatalf("level 2 doubled a counter: %d", got)
	}
	if err := effectstest.LevelUpTo(t, g, me.ID, talent, 1, "{3}{G}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(mine, game.CounterPlusOne, 1) })
	if got := effectstest.CounterCount(g, mine, game.CounterPlusOne); got != 3 {
		t.Errorf("level 3: %d counters, want 1 + 2", got)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 1) })
	if got := effectstest.CounterCount(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("a counter the opponent's permanent got by its controller's placement: %d, want 1", got)
	}
}

func TestSweepFInnkeepersTalentPutsACounterAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushClass(g, me.ID, "Innkeeper's Talent", "Enchantment — Class", effectstest.SweepFInnkeeperOracle)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAnyPendingTargetPrompts(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters at the beginning of combat = %d, want 1", got)
	}
}

func TestSweepFSpringheartNantukoCopiesTheEnchantedCreatureWhenPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.SweepFBestowed(g, me.ID, bear)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SweepFAddMana(g, me.ID, "{C}{G}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Bear"); got != 2 {
		t.Errorf("%d Bears, want the original and a token copy", got)
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Insect"); got != 0 {
		t.Errorf("%d Insects, want none once a copy was made", got)
	}
}

func TestSweepFSpringheartNantukoMakesAnInsectWhenDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.SweepFBestowed(g, me.ID, bear)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Bear"); got != 1 {
		t.Errorf("%d Bears, want no copy", got)
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Insect"); got != 1 {
		t.Errorf("%d Insects, want one", got)
	}
}

func TestSweepFSpringheartNantukoUnattachedMakesAnInsect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Springheart Nantuko", "Enchantment Creature — Insect Monk", effectstest.SweepFSpringheartOracle, 1, 1)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Insect"); got != 1 {
		t.Errorf("%d Insects, want one with no prompt", got)
	}
}

// --- Veyran, Voice of Duality -----------------------------------------

func TestSweepFVeyranDoublesMagecraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	veyran := effectstest.B12Push(g, me.ID, "Veyran, Voice of Duality", "Legendary Creature — Efreet Wizard", effectstest.SweepFVeyranOracle, 2, 2)
	effectstest.B12Push(g, me.ID, "Archmage Emeritus", "Creature — Human Wizard", effectstest.SweepFArchmageOracle, 2, 2)
	effectstest.ToMainForCost(t, g)
	hand := me.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Spell", "Instant", "", nil)
	sweepFSettle(t, g, me.ID)
	// castCatalogSpell put the spell in hand first, so the hand is back
	// to `hand` before the draws.
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2: Archmage Emeritus's magecraft triggers twice", got)
	}
	if p := effectstest.LayeredCard(t, g, veyran).CurrentPower(); p != 4 {
		t.Errorf("Veyran power %d, want 4: its own magecraft triggers twice", p)
	}
}

// --- Isochron Scepter --------------------------------------------------

func TestSweepFIsochronScepterImprintsAndCastsACopyFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := sweepFHand(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	big := sweepFHand(me, "Big Instant", "Instant", "{2}{U}{U}", "")
	scepter := effectstest.CastCatalogSpell(t, g, "Isochron Scepter", "Artifact", effectstest.SweepFIsochronOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
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
	effectstest.AnswerChooseCards(t, g, me.ID, bolt)
	if !effectstest.InExile(g, bolt) {
		t.Fatal("the imprinted card is not in exile")
	}

	effectstest.SweepFAddMana(g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, scepter, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	cp := effectstest.CopyOf(t, g, "Lightning Bolt")
	before := opp.Life
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("casting the copy free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("opponent life %d → %d, want the copy to deal 3", before, opp.Life)
	}
	if !effectstest.InExile(g, bolt) {
		t.Error("the imprinted card left exile")
	}
}

// --- High Fae Trickster ------------------------------------------------

func TestSweepFHighFaeTricksterLetsYouCastAsThoughFlash(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.SweepFHighFaeOracle)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	third := effectstest.SweepFTop(me, "Third", "Sorcery")
	second := effectstest.SweepFTop(me, "Second", "Sorcery")
	first := effectstest.SweepFTop(me, "First", "Sorcery")

	effectstest.CastCatalogSpell(t, g, "Expressive Iteration", "Sorcery", effectstest.SweepFExpressiveOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, second)
	effectstest.AnswerChooseCards(t, g, me.ID, first)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.InHandOf(me, second) {
		t.Error("the card chosen for the hand is not there")
	}
	if me.Library.Cards[0].InstanceID != first {
		t.Error("the card chosen for the bottom is not on the bottom")
	}
	if !effectstest.InExile(g, third) {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	second := effectstest.SweepFTop(me, "Second", "Sorcery")
	first := effectstest.SweepFTop(me, "First", "Sorcery")

	effectstest.CastCatalogSpell(t, g, "Expressive Iteration", "Sorcery", effectstest.SweepFExpressiveOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, first)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, first) || effectstest.InExile(g, second) || !me.Library.Contains(second) {
		t.Errorf("hand=%v exiled=%v library=%v, want one in hand and the other on the bottom",
			effectstest.InHandOf(me, first), effectstest.InExile(g, second), me.Library.Contains(second))
	}
}

// --- Tainted Pact ------------------------------------------------------

func TestSweepFTaintedPactStopsOnTheSecondCardWithAName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	alpha2 := effectstest.SweepFTop(me, "Alpha", "Sorcery")
	beta := effectstest.SweepFTop(me, "Beta", "Sorcery")
	alpha1 := effectstest.SweepFTop(me, "Alpha", "Sorcery")
	under := me.Library.Cards[len(me.Library.Cards)-4].InstanceID

	effectstest.CastCatalogSpell(t, g, "Tainted Pact", "Instant", effectstest.SweepFTaintedPactOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false) // Alpha
	effectstest.AnswerMayChoice(t, g, me.ID, false) // Beta
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Fatal("asked about a second Alpha: the process ends there")
	}
	for _, id := range []uuid.UUID{alpha1, beta, alpha2} {
		if !effectstest.InExile(g, id) {
			t.Errorf("card %s not exiled", id)
		}
	}
	if !me.Library.Contains(under) {
		t.Error("exiled past the duplicate")
	}
}

func TestSweepFTaintedPactTakesTheChosenCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gamma := effectstest.SweepFTop(me, "Gamma", "Sorcery")
	beta := effectstest.SweepFTop(me, "Beta", "Sorcery")
	alpha := effectstest.SweepFTop(me, "Alpha", "Sorcery")

	effectstest.CastCatalogSpell(t, g, "Tainted Pact", "Instant", effectstest.SweepFTaintedPactOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false) // Alpha
	effectstest.AnswerMayChoice(t, g, me.ID, true)  // Beta
	if !effectstest.InExile(g, alpha) || !effectstest.InHandOf(me, beta) || !me.Library.Contains(gamma) {
		t.Errorf("alpha exiled=%v beta in hand=%v gamma in library=%v",
			effectstest.InExile(g, alpha), effectstest.InHandOf(me, beta), me.Library.Contains(gamma))
	}
}

// --- Blade of Selves ---------------------------------------------------

func TestSweepFBladeOfSelvesGivesTheEquippedCreatureMyriad(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	blade := effectstest.B12Push(g, me.ID, "Blade of Selves", "Artifact — Equipment", effectstest.SweepFBladeOracle, 0, 0)
	knight := effectstest.PushTribalCreature(g, me.ID, "Knight", "Creature — Human Knight", 3, 3, "vigilance")
	effectstest.ToMainForCost(t, g)
	effectstest.SweepFAddMana(g, me.ID, "{C}{C}{C}{C}")
	if err := effectstest.ActivateOn(t, g, me.ID, blade, 0, knight); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.DeclareAttack(t, g, opp1.ID, knight)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	evendo := effectstest.B12Push(g, me.ID, "Evendo, Waking Haven", "Land — Planet", effectstest.SweepFEvendoOracle, 0, 0)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	effectstest.B12Creature(g, me.ID, "Elk", "Creature — Elk", 2, 2)
	effectstest.ToMainForCost(t, g)

	effectstest.SweepFAddMana(g, me.ID, "{G}")
	if err := g.ActivateManaAbility(me.ID, evendo, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("the 12+ ability activated with no charge counters")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(evendo, game.CounterCharge, 12) })
	effectstest.ActivateManaFor(t, g, me.ID, evendo, 1, game.ManaAbilityParams{})
	if got := me.ManaPool.Total(); got != 3 {
		t.Errorf("pool %d after {G}, {T}: want three {G} for three creatures", got)
	}
}

// --- Demonic Consultation ----------------------------------------------

func TestSweepFDemonicConsultationFindsTheNamedCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	after := effectstest.SweepFTop(me, "After", "Sorcery")
	wanted := effectstest.SweepFTop(me, "Wanted", "Sorcery")
	other := effectstest.SweepFTop(me, "Other", "Sorcery")
	var six []uuid.UUID
	for i := 0; i < 5; i++ {
		six = append(six, effectstest.SweepFTop(me, "Junk", "Sorcery"))
	}
	// A Wanted among the top six is exiled with them, not found.
	six = append(six, effectstest.SweepFTop(me, "Wanted", "Sorcery"))

	effectstest.CastCatalogSpell(t, g, "Demonic Consultation", "Instant", effectstest.SweepFConsultationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "Wanted")
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range six {
		if !effectstest.InExile(g, id) {
			t.Errorf("a top-six card %s was not exiled", id)
		}
	}
	if !effectstest.InExile(g, other) {
		t.Error("a revealed card that was not named was not exiled")
	}
	if !effectstest.InHandOf(me, wanted) {
		t.Error("the named card is not in hand")
	}
	if !me.Library.Contains(after) {
		t.Error("a card below the named one left the library")
	}
}

// --- Approach of the Second Sun ----------------------------------------

func TestSweepFApproachGoesSeventhThenWins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life

	first := effectstest.CastCatalogSpell(t, g, "Approach of the Second Sun", "Sorcery", effectstest.SweepFApproachOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
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

	effectstest.CastCatalogSpell(t, g, "Approach of the Second Sun", "Sorcery", effectstest.SweepFApproachOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Errorf("outcome %+v, want the caster to win with the second Approach", g.Outcome)
	}
}

// --- Cyberdrive Awakener -----------------------------------------------

func TestSweepFCyberdriveAwakenerAnimatesNoncreatureArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Cyberdrive Awakener", "Artifact Creature — Construct", effectstest.SweepFCyberdriveOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.LayeredCard(t, g, rock)
	if !c.IsCreature() || c.CurrentPower() != 4 {
		t.Errorf("Rock creature=%v power=%d, want a 4/4 artifact creature", c.IsCreature(), c.CurrentPower())
	}
	if !game.HasKeyword(&c, "flying") {
		t.Error("the animated artifact does not have flying from the static")
	}

	late := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Late Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID,
	})
	if effectstest.LayeredCard(t, g, late).IsCreature() {
		t.Error("an artifact that arrived after the trigger was animated (CR 611.2c)")
	}
}

func TestSweepFEtaliExilesUntilNonlandFromEachLibraryAndOffersThemFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var lands, spells []uuid.UUID
	for _, p := range g.Seats {
		spells = append(spells, effectstest.SweepFTop(p, "Spell of "+p.Name, "Sorcery"))
		lands = append(lands, effectstest.SweepFTop(p, "Land of "+p.Name, "Land"))
	}
	effectstest.ImportToBattlefield(t, g, effectstest.EtaliRow(), me)
	effectstest.PassPriorityAroundTable(t, g)

	for i := range g.Seats {
		if !effectstest.InExile(g, lands[i]) || !effectstest.InExile(g, spells[i]) {
			t.Errorf("seat %d: land exiled=%v spell exiled=%v, want both", i, effectstest.InExile(g, lands[i]), effectstest.InExile(g, spells[i]))
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	etali := effectstest.ImportToBattlefield(t, g, effectstest.EtaliRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	// #4250: the entry trigger's free casts are offered inside its
	// resolution; this test declines them.
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		effectstest.DeclineResolutionCast(t, g, me.ID)
	}
	effectstest.ToMainForCost(t, g)
	effectstest.SweepFAddMana(g, me.ID, "{C}{C}{C}{C}{C}{C}{C}{C}{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, etali, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("transform: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, etali); c.Name != "Etali, Primal Sickness" {
		t.Fatalf("face %q after the transform, want the back face", c.Name)
	}
	effectstest.DealCombatDamageToPlayer(g, etali, opp.ID, 11)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Counters[game.CounterPoison]; got != 11 {
		t.Errorf("poison counters %d, want 11", got)
	}
}

// --- Ojer Taq, Deepest Foundation --------------------------------------

func ojerTaqRow() cards.Card {
	return effectstest.TransformRow(cardso.OjerTaqOracleID,
		"Ojer Taq, Deepest Foundation", "Legendary Creature — God", "{4}{W}{W}",
		"Temple of Civilization", "Land", "", "", []string{"W"})
}

func TestSweepFOjerTaqTriplesCreatureTokensOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Ojer Taq, Deepest Foundation", "Legendary Creature — God", cardso.OjerTaqOracleID, 6, 6)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TokenCard("2/2 green Bear"), 1)
		_ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Bear"); got != 3 {
		t.Errorf("%d Bears, want three times one", got)
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasures, want one: a noncreature token is not tripled", got)
	}
}

func TestSweepFOjerTaqReturnsAsTempleOfCivilization(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	god := effectstest.ImportToBattlefield(t, g, ojerTaqRow(), me)
	effectstest.OjerKill(t, g, god)
	effectstest.PassPriorityAroundTable(t, g)
	temple := effectstest.OjerOnBattlefield(g, cardso.OjerTaqOracleID)
	if temple == nil || temple.Name != "Temple of Civilization" || !temple.Tapped {
		t.Fatalf("returned %+v, want a tapped Temple of Civilization", temple)
	}
}
