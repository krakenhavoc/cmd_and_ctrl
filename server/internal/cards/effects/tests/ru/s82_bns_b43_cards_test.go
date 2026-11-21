package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b43_cards_test.go — Foundations Jumpstart batch 43 (#5929).

const (
	bnsOmenOfTheSea         = "41960d32-ddb5-42be-94b2-3a2e77ca148d"
	bnsRetreatToKazandu     = "3f8e5ff1-af89-427e-924c-19a44f9a3788"
	bnsTrailOfCrumbs        = "185d6356-8f8d-4b32-8259-b4eb0a95866c"
	bnsTrialOfKnowledge     = "6155908a-0079-42c7-ae81-326a6faccb89"
	bnsTrialOfZeal          = "0798eb13-83f1-4b7c-b537-83a9da2915a6"
	bnsCageOfHands          = "18ffd8ad-1556-450d-8bc1-8103a3d0404e"
	bnsCaptureSphere        = "2dc59566-10cb-468a-b6a2-17d6c0186e6d"
	bnsCartoucheOfKnowledge = "488696c2-b0c7-49be-a5a1-e833b6fd8751"
	bnsCartoucheSolidarity  = "b9176173-c4a1-45ae-8f34-bd6b31c141fe"
	bnsCartoucheOfZeal      = "4fa23be3-a192-45ab-bdf6-646d4be305e1"
	bnsChosenByHeliod       = "f6475b40-e694-4aac-a25c-bbd48d45127d"
)

func TestS82BnsB43EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsOmenOfTheSea: "Omen of the Sea", bnsRetreatToKazandu: "Retreat to Kazandu",
		bnsTrailOfCrumbs: "Trail of Crumbs", bnsTrialOfKnowledge: "Trial of Knowledge",
		bnsTrialOfZeal: "Trial of Zeal", bnsCageOfHands: "Cage of Hands",
		bnsCaptureSphere: "Capture Sphere", bnsCartoucheOfKnowledge: "Cartouche of Knowledge",
		bnsCartoucheSolidarity: "Cartouche of Solidarity", bnsCartoucheOfZeal: "Cartouche of Zeal",
		bnsChosenByHeliod: "Chosen by Heliod",
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

func bnsPushLibrary(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear",
				Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
		}
	})
}

func TestS82BnsB43CartouchesPumpAndGrant(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, kw string
		mana             string
	}{
		{"Cartouche of Knowledge", bnsCartoucheOfKnowledge, "flying", ""},
		{"Cartouche of Solidarity", bnsCartoucheSolidarity, "first strike", ""},
		{"Cartouche of Zeal", bnsCartoucheOfZeal, "haste", ""},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bnsPushLibrary(g, me, 5)
		host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)
		if err := effectstest.CastCatalogSpellErr(t, g, tc.name, effectstest.AuraTypeLine, tc.oracle, effectstest.B16TargetCard(theirs)); err == nil {
			t.Errorf("%s enchanted an opponent's creature", tc.name)
		}
		effectstest.CastCatalogSpell(t, g, tc.name, effectstest.AuraTypeLine, tc.oracle, effectstest.B16TargetCard(host))
		effectstest.PassPriorityAroundTable(t, g)
		if p, to := effectstest.EffectivePTOf(t, g, host); p != 3 || to != 3 {
			t.Errorf("%s: host is %d/%d, want 3/3", tc.name, p, to)
		}
		if !effectstest.HasEffectiveKeyword(t, g, host, tc.kw) {
			t.Errorf("%s: host lacks %s", tc.name, tc.kw)
		}
	}
}

func TestS82BnsB43CartoucheOfKnowledgeDrawsOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bnsPushLibrary(g, me, 5)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	before := effectstest.HandSize(me)
	effectstest.CastCatalogSpell(t, g, "Cartouche of Knowledge", effectstest.AuraTypeLine, bnsCartoucheOfKnowledge, effectstest.B16TargetCard(host))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != before+1 {
		t.Errorf("hand = %d, want %d (the cast card left, one was drawn)", got, before+1)
	}
}

func TestS82BnsB43CartoucheOfSolidarityMakesAVigilantWarrior(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Cartouche of Solidarity", effectstest.AuraTypeLine, bnsCartoucheSolidarity, effectstest.B16TargetCard(host))
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Warrior" && c.IsToken() {
			n++
			if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "vigilance") {
				t.Error("the Warrior lacks vigilance")
			}
		}
	}
	if n != 1 {
		t.Errorf("Warrior tokens = %d, want 1", n)
	}
}

func TestS82BnsB43CartoucheOfZealStopsATargetCreatureBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Cartouche of Zeal", effectstest.AuraTypeLine, bnsCartoucheOfZeal, effectstest.B16TargetCard(host))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, theirs)
	if !game.Restricted(&c, game.CantBlock) {
		t.Error("the targeted creature can still block")
	}
	c = effectstest.LayeredCard(t, g, host)
	if game.Restricted(&c, game.CantBlock) {
		t.Error("the enchanted creature was restricted")
	}
}

func TestS82BnsB43ChosenByHeliodDrawsAndAddsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bnsPushLibrary(g, me, 5)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	before := effectstest.HandSize(me)
	effectstest.CastCatalogSpell(t, g, "Chosen by Heliod", effectstest.AuraTypeLine, bnsChosenByHeliod, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if p, to := effectstest.EffectivePTOf(t, g, theirs); p != 2 || to != 4 {
		t.Errorf("enchanted creature is %d/%d, want 2/4", p, to)
	}
	if got := effectstest.HandSize(me); got != before+1 {
		t.Errorf("hand = %d, want %d", got, before+1)
	}
}

func TestS82BnsB43CageOfHandsBarsAttackAndBlockAndReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	cage := effectstest.CastCatalogSpell(t, g, "Cage of Hands", effectstest.AuraTypeLine, bnsCageOfHands, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, theirs)
	if !game.Restricted(&c, game.CantAttack) || !game.Restricted(&c, game.CantBlock) {
		t.Fatal("the enchanted creature can still attack or block")
	}
	if err := g.ActivateCatalogAbility(me.ID, cage, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the return was activated with no mana")
	}
	effectstest.B31AddMana(me, "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, cage, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, cage); ok {
		t.Error("the Aura is still on the battlefield")
	}
	inHand := false
	for _, h := range me.Hand.Cards {
		if h.InstanceID == cage {
			inHand = true
		}
	}
	if !inHand {
		t.Error("the Aura is not in its owner's hand")
	}
	c = effectstest.LayeredCard(t, g, theirs)
	if game.Restricted(&c, game.CantAttack) {
		t.Error("the creature is still barred after the Aura left")
	}
}

func TestS82BnsB43CaptureSphereTapsAndKeepsItTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Capture Sphere", effectstest.AuraTypeLine, bnsCaptureSphere, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.DftCG2Tapped(g, theirs) {
		t.Fatal("the enchanted creature is not tapped")
	}
	effectstest.DftMPPassTurn(t, g)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.DftCG2Tapped(g, theirs) {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

func TestS82BnsB43OmenOfTheSeaScriesThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bnsPushLibrary(g, me, 6)
	effectstest.AdvanceToMain(t, g)
	before := effectstest.HandSize(me)
	effectstest.CastCatalogSpell(t, g, "Omen of the Sea", "Enchantment", bnsOmenOfTheSea, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != before+1 {
		t.Errorf("hand = %d, want %d (the Omen left, one card drawn)", got, before+1)
	}
}

func TestS82BnsB43OmenOfTheSeaSacrificeScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bnsPushLibrary(g, me, 6)
	omen := effectstest.B12Push(g, me.ID, "Omen of the Sea", "Enchantment", bnsOmenOfTheSea, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, omen, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.B31AddMana(me, "U", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, omen, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, omen); ok {
		t.Error("the Omen was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
}

func TestS82BnsB43RetreatToKazanduModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Retreat to Kazandu", "Enchantment", bnsRetreatToKazandu, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1}, 0)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", c.Counters[game.CounterPlusOne])
	}
	life := me.Life
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1}, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
}

func TestS82BnsB43TrialOfKnowledgeDrawsThreeDiscardsOneAndBouncesForACartouche(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bnsPushLibrary(g, me, 8)
	host := effectstest.B12Creature(g, me.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	before := effectstest.HandSize(me)
	trial := effectstest.CastCatalogSpell(t, g, "Trial of Knowledge", "Enchantment", bnsTrialOfKnowledge, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandSize(me); got != before+2 {
		t.Errorf("hand = %d, want %d (cast one, drew three, discarded one)", got, before+2)
	}
	if _, ok := effectstest.BattlefieldCard(g, trial); !ok {
		t.Fatal("the Trial is not on the battlefield")
	}
	effectstest.CastCatalogSpell(t, g, "Cartouche of Solidarity", "Enchantment — Aura Cartouche", bnsCartoucheSolidarity, effectstest.B16TargetCard(host))
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceTriggerOrder {
			if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			break
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, trial); ok {
		t.Error("the Trial stayed after a Cartouche entered")
	}
}

func TestS82BnsB43TrialOfZealDealsThreeToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	trial := effectstest.CastCatalogSpell(t, g, "Trial of Zeal", "Enchantment", bnsTrialOfZeal, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life = %d, want %d", opp.Life, life-3)
	}
	if _, ok := effectstest.BattlefieldCard(g, trial); !ok {
		t.Error("the Trial left with no Cartouche entering")
	}
}

func bnsFoods(g *game.Game, p uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == p && c.HasSubtype("Food") {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func TestS82BnsB43TrailOfCrumbsMakesFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Trail of Crumbs", "Enchantment", bnsTrailOfCrumbs, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(bnsFoods(g, me.ID)); n != 1 {
		t.Errorf("Foods = %d, want 1", n)
	}
}

func TestS82BnsB43TrailOfCrumbsDigsOnlyWhenYouPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Trail of Crumbs", "Enchantment", bnsTrailOfCrumbs, 0, 0)
	bnsPushLibrary(g, me, 4)
	var spell, bear uuid.UUID
	g.WithWriteLock(func() {
		spell = uuid.New()
		bear = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: bear, Name: "Top Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: spell, Name: "Top Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	})
	effectstest.AdvanceToMain(t, g)

	// Declining the payment does nothing.
	food := effectstest.PushToken(g, me.ID, effects.FoodToken())
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, food, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("crack the Food: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a look was offered without paying")
	}

	// Paying looks at the top two; the permanent card goes to hand.
	food = effectstest.PushToken(g, me.ID, effects.FoodToken())
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, food, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("crack the second Food: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ch == nil {
		t.Fatal("no look-at-top prompt after paying")
	}
	for _, id := range ch.ChooseCards {
		if id == spell {
			t.Error("a non-permanent card was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	inHand := false
	for _, h := range me.Hand.Cards {
		if h.InstanceID == bear {
			inHand = true
		}
	}
	if !inHand {
		t.Error("the revealed permanent card is not in hand")
	}
}
