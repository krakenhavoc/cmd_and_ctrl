package testsru

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b128_cards_test.go — set GAP, batch 128 (#5948). Each test plays
// the real catalog entry through the engine and asserts what its text says.

const (
	gapB128BillOracle      = "e461ef8d-59ca-440a-8773-da1de3b6b195"
	gapB128EagleOracle     = "4446e59c-29f7-4722-92be-2625c1d3dab9"
	gapB128CavalierOracle  = "884d08bf-7e97-40cc-864e-445ae0bd1646"
	gapB128FarmerOracle    = "4b1603f3-a246-449b-a109-99adb49e5ea4"
	gapB128FrodoOracle     = "ae50be07-c8bc-45db-b47d-701c5d9ee633"
	gapB128GandalfOracle   = "7a80212a-aae3-483b-abb6-7271a62dff8a"
	gapB128HobbitOracle    = "547da0ef-f1e1-4b04-ba83-4845552b0dc4"
	gapB128ProtectorOracle = "cf6bb391-4fde-43c4-a0c5-990630d0ddcf"
	gapB128SaradocOracle   = "3c7f7a67-6a71-4f40-8554-2e724e6916a6"
	gapB128ShirriffOracle  = "14f36157-f55f-495d-9ffe-f38135a59750"
)

func TestGapB128EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB128BillOracle:      "Bill the Pony",
		gapB128EagleOracle:     "Eagle of Deliverance",
		gapB128CavalierOracle:  "East-Mark Cavalier",
		gapB128FarmerOracle:    "Eastfarthing Farmer",
		gapB128FrodoOracle:     "Frodo, Determined Hero",
		gapB128GandalfOracle:   "Gandalf, White Rider",
		gapB128HobbitOracle:    "Nimble Hobbit",
		gapB128ProtectorOracle: "Protector of Gondor",
		gapB128SaradocOracle:   "Saradoc, Master of Buckland",
		gapB128ShirriffOracle:  "Shire Shirriff",
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

func gapB128Ref(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func gapB128Main(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func gapB128Foods(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.HasSubtype("Food") {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func gapB128Tokens(g *game.Game, controller uuid.UUID, name string) int {
	n, _ := effectstest.CountTokensNamed(g, controller, name)
	return n
}

func gapB128Die(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
}

// gapB128Answer answers the first pick_target prompt open for `chooser`, if
// any, with `ref`.
func gapB128Answer(t *testing.T, g *game.Game, chooser uuid.UUID, ref game.TargetRef) {
	t.Helper()
	if p := effectstest.LatestPickTarget(g, chooser); p != nil {
		if err := g.ResolvePickTarget(p.ID, chooser, ref); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
}

// gapB128Settle passes priority until the stack is empty, saying `may` to
// "you may" prompts on the way.
func gapB128Settle(t *testing.T, g *game.Game, chooser uuid.UUID, may bool) {
	t.Helper()
	for i := 0; i < 48; i++ {
		if c := effectstest.LatestTriggerPrompt(g, chooser); c != nil {
			if err := g.ResolveTriggerPrompt(c.ID, chooser, may); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			continue
		}
		if len(g.PendingTriggers) > 0 {
			g.RunStateChecksForTest()
			if effectstest.LatestTriggerPrompt(g, chooser) != nil {
				continue
			}
		}
		if len(g.PendingChoices) > 0 || effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				continue
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack did not settle")
}

func gapB128Library(p *game.Player, cards ...game.Card) {
	for i := len(cards) - 1; i >= 0; i-- {
		cards[i].InstanceID = uuid.New()
		cards[i].Owner, cards[i].Controller = p.ID, p.ID
		p.Library.PushTop(cards[i])
	}
}

// --- Bill the Pony -----------------------------------------------------------------

func TestGapB128BillMakesTwoFoodsAndSacrificesOneForToughnessDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wall := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 5)
	bill := effectstest.CastHoldCreature(t, g, "Bill the Pony", "Legendary Creature — Horse", gapB128BillOracle, 1, 4)
	foods := gapB128Foods(g, me.ID)
	if len(foods) != 2 {
		t.Fatalf("Foods = %d, want 2", len(foods))
	}
	effectstest.DeclareAttack(t, g, opp.ID, wall)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bill, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(wall), SacrificeIDs: foods[:1]}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(gapB128Foods(g, me.ID)) != 1 {
		t.Error("a Food was not sacrificed")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 5 {
		t.Errorf("the 0/5 dealt %d combat damage, want 5 (its toughness)", got)
	}
}

func TestGapB128BillRefusesWithoutAFoodOrAnOpponentsCreature(t *testing.T) {
	g, me, opp := gapB128Main(t)
	bill := effectstest.B12Push(g, me.ID, "Bill the Pony", "Legendary Creature — Horse", gapB128BillOracle, 1, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, bill, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bill)}); err == nil {
		t.Error("activated with no Food to sacrifice")
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.FoodToken(), 1) })
	if err := g.ActivateCatalogAbility(me.ID, bill, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(theirs), SacrificeIDs: gapB128Foods(g, me.ID)}); err == nil {
		t.Error("targeted an opponent's creature")
	}
}

// --- Eagle of Deliverance ---------------------------------------------------------------

func TestGapB128EagleCountersAnotherCreatureAndDrawsIfItIsSmall(t *testing.T) {
	g, me, _ := gapB128Main(t)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastHoldCreature(t, g, "Eagle of Deliverance", "Creature — Bird Soldier", gapB128EagleOracle, 5, 5)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB128Ref(small)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, small, game.CounterIndestructible) != 1 {
		t.Error("the small creature did not get an indestructible counter")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1 for a power-2 target", hand, me.Hand.Size())
	}
	// A power-5 target draws nothing.
	g2, me2, _ := gapB128Main(t)
	effectstest.B12Creature(g2, me2.ID, "Small", "Creature — Bear", 2, 2)
	big2 := effectstest.B12Creature(g2, me2.ID, "Big", "Creature — Ogre", 5, 5)
	hand2 := me2.Hand.Size()
	effectstest.CastHoldCreature(t, g2, "Eagle of Deliverance", "Creature — Bird Soldier", gapB128EagleOracle, 5, 5)
	gapB128Answer(t, g2, me2.ID, gapB128Ref(big2))
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.CounterCount(g2, big2, game.CounterIndestructible) != 1 {
		t.Error("the big creature did not get an indestructible counter")
	}
	if me2.Hand.Size() != hand2 {
		t.Errorf("hand %d -> %d, want no draw for a power-5 target", hand2, me2.Hand.Size())
	}
}

// --- East-Mark Cavalier -------------------------------------------------------------------

func gapB128BlockedCombat(t *testing.T, g *game.Game, me, opp *game.Player, attacker, blocker uuid.UUID) {
	t.Helper()
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapB128CavalierDestroysAGoblinOrOrcItDamages(t *testing.T) {
	for _, tc := range []struct {
		subtype string
		dies    bool
	}{{"Goblin", true}, {"Orc", true}, {"Bear", false}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		cav := effectstest.B12Push(g, me.ID, "East-Mark Cavalier", "Creature — Human Knight", gapB128CavalierOracle, 2, 2)
		blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — "+tc.subtype, 0, 9)
		gapB128BlockedCombat(t, g, me, opp, cav, blocker)
		if got := !g.Battlefield.Contains(blocker); got != tc.dies {
			t.Errorf("a %s blocker destroyed = %v, want %v", tc.subtype, got, tc.dies)
		}
	}
}

// --- Eastfarthing Farmer -----------------------------------------------------------------------

func TestGapB128FarmerPumpsByTheFoodsYouControl(t *testing.T) {
	g, me, _ := gapB128Main(t)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.FoodToken(), 1) })
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Eastfarthing Farmer", "Creature — Halfling Peasant", gapB128FarmerOracle, 2, 3)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the reflexive trigger")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB128Ref(bear)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(gapB128Foods(g, me.ID)) != 2 {
		t.Fatalf("Foods = %d, want 2", len(gapB128Foods(g, me.ID)))
	}
	if pw, tough := effectstest.EffectivePTOf(t, g, bear); pw != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4 (two Foods)", pw, tough)
	}
}

// --- Frodo, Determined Hero --------------------------------------------------------------------

func TestGapB128FrodoAttachesAnEquipmentOfManaValueTwoOrThree(t *testing.T) {
	g, me, _ := gapB128Main(t)
	cheap := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cheap Sword", TypeLine: "Artifact — Equipment", ManaCost: "{2}", Owner: me.ID, Controller: me.ID,
	})
	pricey := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Pricey Sword", TypeLine: "Artifact — Equipment", ManaCost: "{4}", Owner: me.ID, Controller: me.ID,
	})
	frodo := effectstest.CastHoldCreature(t, g, "Frodo, Determined Hero", "Legendary Creature — Halfling Warrior", gapB128FrodoOracle, 2, 2)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB128Ref(pricey)); err == nil {
		t.Error("an Equipment with mana value 4 was accepted")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB128Ref(cheap)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	gapB128Settle(t, g, me.ID, true)
	c, ok := effectstest.BattlefieldCard(g, cheap)
	if !ok || c.AttachedTo.ID != frodo {
		t.Errorf("the Equipment is attached to %v, want Frodo", c.AttachedTo.ID)
	}
}

func TestGapB128FrodoLeavesTheEquipmentIfYouDecline(t *testing.T) {
	g, me, _ := gapB128Main(t)
	cheap := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cheap Sword", TypeLine: "Artifact — Equipment", ManaCost: "{3}", Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastHoldCreature(t, g, "Frodo, Determined Hero", "Legendary Creature — Halfling Warrior", gapB128FrodoOracle, 2, 2)
	gapB128Answer(t, g, me.ID, gapB128Ref(cheap))
	gapB128Settle(t, g, me.ID, false)
	if c, _ := effectstest.BattlefieldCard(g, cheap); c.AttachedTo.ID != uuid.Nil {
		t.Error("the Equipment was attached though the player declined")
	}
}

func TestGapB128FrodoTakesNoDamageOnYourTurnOnly(t *testing.T) {
	g, me, _ := gapB128Main(t)
	frodo := effectstest.B12Push(g, me.ID, "Frodo, Determined Hero", "Legendary Creature — Halfling Warrior", gapB128FrodoOracle, 2, 2)
	src := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, frodo, 1) })
	if got := effectstest.DamageMarkedOn(g, frodo); got != 0 {
		t.Errorf("Frodo took %d damage on his controller's turn, want 0", got)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, frodo, 1) })
	if got := effectstest.DamageMarkedOn(g, frodo); got != 1 {
		t.Errorf("Frodo took %d damage on an opponent's turn, want 1", got)
	}
}

// --- Gandalf, White Rider --------------------------------------------------------------------------

func TestGapB128GandalfPumpsCreaturesAndScriesOnEachSpell(t *testing.T) {
	g, me, opp := gapB128Main(t)
	effectstest.B12Push(g, me.ID, "Gandalf, White Rider", "Legendary Creature — Avatar Wizard", gapB128GandalfOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	gapB128Library(me, game.Card{Name: "1"}, game.Card{Name: "2"}, game.Card{Name: "3"})
	effectstest.CastHoldCreature(t, g, "Spell One", "Creature — Bear", "", 1, 1)
	if pw, _ := effectstest.EffectivePTOf(t, g, mine); pw != 3 {
		t.Errorf("my creature has %d power after a spell, want 3", pw)
	}
	if pw, _ := effectstest.EffectivePTOf(t, g, theirs); pw != 2 {
		t.Errorf("their creature has %d power, want 2", pw)
	}
	if c := effectstest.ScryChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 1 {
		t.Error("the spell did not scry 1")
	}
}

func TestGapB128GandalfGoesFifthFromTheTopWhenItDies(t *testing.T) {
	g, me, _ := gapB128Main(t)
	gandalf := effectstest.B12Push(g, me.ID, "Gandalf, White Rider", "Legendary Creature — Avatar Wizard", gapB128GandalfOracle, 3, 3)
	gapB128Library(me, game.Card{Name: "1"}, game.Card{Name: "2"}, game.Card{Name: "3"}, game.Card{Name: "4"}, game.Card{Name: "5"}, game.Card{Name: "6"})
	gapB128Die(t, g, gandalf)
	g.RunStateChecksForTest()
	gapB128Settle(t, g, me.ID, true)
	n := len(me.Library.Cards)
	if n < 5 || me.Library.Cards[n-5].InstanceID != gandalf {
		t.Errorf("Gandalf is not fifth from the top of the library (library of %d)", n)
	}
	if me.Graveyard.Contains(gandalf) {
		t.Error("Gandalf is still in the graveyard")
	}
}

func TestGapB128GandalfStaysInTheGraveyardIfYouDecline(t *testing.T) {
	g, me, _ := gapB128Main(t)
	gandalf := effectstest.B12Push(g, me.ID, "Gandalf, White Rider", "Legendary Creature — Avatar Wizard", gapB128GandalfOracle, 3, 3)
	gapB128Library(me, game.Card{Name: "1"}, game.Card{Name: "2"}, game.Card{Name: "3"}, game.Card{Name: "4"}, game.Card{Name: "5"})
	gapB128Die(t, g, gandalf)
	g.RunStateChecksForTest()
	gapB128Settle(t, g, me.ID, false)
	if !me.Graveyard.Contains(gandalf) {
		t.Error("Gandalf left the graveyard though the player declined")
	}
}

// --- Nimble Hobbit -----------------------------------------------------------------------------------

func TestGapB128NimbleHobbitSacrificesAFoodToTapACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hobbit := effectstest.B12Push(g, me.ID, "Nimble Hobbit", "Creature — Halfling Peasant", gapB128HobbitOracle, 1, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.FoodToken(), 1) })
	food := gapB128Foods(g, me.ID)
	effectstest.DeclareAttack(t, g, opp.ID, hobbit)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.B24OpenPayPrompt(g, me)
	if prompt == nil {
		t.Fatal("no sacrifice-or-pay prompt")
	}
	option := 0
	if err := g.ResolvePayUnlessWithOption(prompt.ID, me.ID, true, &option, food); err != nil {
		t.Fatalf("sacrifice the Food: %v", err)
	}
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if len(gapB128Foods(g, me.ID)) != 0 {
		t.Error("the Food was not sacrificed")
	}
	if !effectstest.HmlTapped(t, g, theirs) {
		t.Error("the opponent's creature was not tapped")
	}
}

func TestGapB128NimbleHobbitPaysThreeToTapACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hobbit := effectstest.B12Push(g, me.ID, "Nimble Hobbit", "Creature — Halfling Peasant", gapB128HobbitOracle, 1, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(hobbit, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.B24OpenPayPrompt(g, me)
	if prompt == nil {
		t.Fatal("no sacrifice-or-pay prompt")
	}
	option := 1
	if err := g.ResolvePayUnlessWithOption(prompt.ID, me.ID, true, &option, nil); err != nil {
		t.Fatalf("pay {2}{W}: %v", err)
	}
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, theirs) {
		t.Error("the opponent's creature was not tapped")
	}
}

func TestGapB128NimbleHobbitAsksNothingWithNoFoodAndNoMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hobbit := effectstest.B12Push(g, me.ID, "Nimble Hobbit", "Creature — Halfling Peasant", gapB128HobbitOracle, 1, 3)
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, hobbit)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B24OpenPayPrompt(g, me) != nil {
		t.Error("a payment was offered with no Food and no mana")
	}
}

// --- Protector of Gondor ----------------------------------------------------------------------------------

func TestGapB128ProtectorMakesAHumanSoldier(t *testing.T) {
	g, me, _ := gapB128Main(t)
	effectstest.CastHoldCreature(t, g, "Protector of Gondor", "Creature — Human Soldier", gapB128ProtectorOracle, 3, 3)
	if n := gapB128Tokens(g, me.ID, "Human Soldier"); n != 1 {
		t.Errorf("%d Human Soldier tokens, want 1", n)
	}
}

// --- Saradoc, Master of Buckland -------------------------------------------------------------------------

func TestGapB128SaradocMakesAHalflingForSmallNontokenCreatures(t *testing.T) {
	g, me, _ := gapB128Main(t)
	effectstest.B12Push(g, me.ID, "Saradoc, Master of Buckland", "Legendary Creature — Halfling Citizen", gapB128SaradocOracle, 2, 4)
	effectstest.CastHoldCreature(t, g, "Small", "Creature — Bear", "", 2, 2)
	if n := gapB128Tokens(g, me.ID, "Halfling"); n != 1 {
		t.Fatalf("%d Halfling tokens after a power-2 creature, want 1", n)
	}
	effectstest.CastHoldCreature(t, g, "Large", "Creature — Ogre", "", 3, 3)
	if n := gapB128Tokens(g, me.ID, "Halfling"); n != 1 {
		t.Errorf("%d Halfling tokens after a power-3 creature, want still 1", n)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Human Soldier"), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB128Tokens(g, me.ID, "Halfling"); n != 1 {
		t.Errorf("%d Halfling tokens after a token entered, want still 1 (nontoken only)", n)
	}
}

func TestGapB128SaradocHimselfMakesAHalflingWhenHeEnters(t *testing.T) {
	g, me, _ := gapB128Main(t)
	effectstest.CastHoldCreature(t, g, "Saradoc, Master of Buckland", "Legendary Creature — Halfling Citizen", gapB128SaradocOracle, 2, 4)
	if n := gapB128Tokens(g, me.ID, "Halfling"); n != 1 {
		t.Errorf("%d Halfling tokens when Saradoc entered, want 1", n)
	}
}

func TestGapB128SaradocTapsTwoOtherHalflingsForPowerAndLifelink(t *testing.T) {
	g, me, _ := gapB128Main(t)
	saradoc := effectstest.B12Push(g, me.ID, "Saradoc, Master of Buckland", "Legendary Creature — Halfling Citizen", gapB128SaradocOracle, 2, 4)
	h1 := effectstest.B12Creature(g, me.ID, "H1", "Creature — Halfling", 1, 1)
	h2 := effectstest.B12Creature(g, me.ID, "H2", "Creature — Halfling", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, saradoc, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{saradoc, h1}, Strict: true}); err == nil {
		t.Error("Saradoc paid for his own ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, saradoc, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{h1, bear}, Strict: true}); err == nil {
		t.Error("a non-Halfling paid")
	}
	if err := g.ActivateCatalogAbility(me.ID, saradoc, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{h1, h2}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pw, _ := effectstest.EffectivePTOf(t, g, saradoc); pw != 4 {
		t.Errorf("Saradoc has %d power, want 4", pw)
	}
	if !effectstest.HasEffectiveKeyword(t, g, saradoc, "lifelink") {
		t.Error("Saradoc lacks lifelink")
	}
	if !effectstest.HmlTapped(t, g, h1) || !effectstest.HmlTapped(t, g, h2) {
		t.Error("the Halflings were not tapped")
	}
}

// --- Shire Shirriff --------------------------------------------------------------------------------------------

func TestGapB128ShirriffExilesACreatureUntilItLeaves(t *testing.T) {
	g, me, opp := gapB128Main(t)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.FoodToken(), 1) })
	food := gapB128Foods(g, me.ID)[0]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	shirriff := effectstest.CastHoldCreature(t, g, "Shire Shirriff", "Creature — Halfling Soldier", gapB128ShirriffOracle, 2, 2)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); c != nil {
		if !effectstest.HasID(c.ChooseCards, food) || effectstest.HasID(c.ChooseCards, shirriff) {
			t.Errorf("the sacrifice offers %v, want only the token", c.ChooseCards)
		}
		effectstest.AnswerOwnPermanents(t, g, me.ID, food)
	}
	gapB128Answer(t, g, me.ID, gapB128Ref(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(food) {
		t.Error("the token was not sacrificed")
	}
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the opponent's creature was not exiled")
	}
	gapB128Die(t, g, shirriff)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	returned := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Theirs" && c.Controller == opp.ID {
			returned = true
		}
	}
	if !returned {
		t.Error("the exiled creature did not return when the Shirriff left")
	}
}

func TestGapB128ShirriffWithNoTokenDoesNothing(t *testing.T) {
	g, me, opp := gapB128Main(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Shire Shirriff", "Creature — Halfling Soldier", gapB128ShirriffOracle, 2, 2)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("the Shirriff asked for a sacrifice or a target with no token to sacrifice")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("a creature was exiled without a sacrifice")
	}
}
