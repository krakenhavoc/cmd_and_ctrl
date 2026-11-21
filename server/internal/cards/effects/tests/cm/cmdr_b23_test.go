package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b23_test.go — Commander deck cards, batch 23 (#5368).

const (
	cmdrB23BulwarkOracle    = "8ad0d89b-beef-40f7-8ce8-65589e3970be"
	cmdrB23BountifulOracle  = "2eb69a8f-9456-4852-b868-85ae609d3441"
	cmdrB23CryptOracle      = "4fe8af73-c84a-44bd-9739-ee5c8b027874"
	cmdrB23DeceptiveOracle  = "1831fe12-dbe0-437f-8fc8-f01bbb701fe1"
	cmdrB23ForebodingOracle = "bcfe1653-e602-4d38-abe7-bfcc7f203f9d"
	cmdrB23PiaOracle        = "31cb87bf-1020-4fd5-95f7-8ecb197d823b"
	cmdrB23SaheeliOracle    = "20f602b3-814f-406c-b7eb-2d1fc98f8cf1"
	cmdrB23TemmetOracle     = "3dc43ffe-f6b0-4cdc-86b3-01c0f41df0ae"
	cmdrB23ProphetOracle    = "eefe2d53-4856-465d-a900-77836ef9ae07"
	cmdrB23OmnitoolOracle   = "e91240bc-4a13-4131-9379-5ddac6555f4a"
)

func TestCmdrB23CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdrB23BulwarkOracle:    "Walking Bulwark",
		cmdrB23BountifulOracle:  "Bountiful Landscape",
		cmdrB23CryptOracle:      "Crypt of Agadeem",
		cmdrB23DeceptiveOracle:  "Deceptive Landscape",
		cmdrB23ForebodingOracle: "Foreboding Landscape",
		cmdrB23PiaOracle:        "Pia Nalaar, Chief Mechanic",
		cmdrB23SaheeliOracle:    "Saheeli, Radiant Creator",
		cmdrB23TemmetOracle:     "Temmet, Naktamun's Will",
		cmdrB23ProphetOracle:    "Prophet of the Scarab",
		cmdrB23OmnitoolOracle:   "Adaptive Omnitool",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

func TestCmdrB23WalkingBulwarkLetsADefenderAttackAndHitForItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bulwark := effectstest.B12Push(g, me.ID, "Walking Bulwark", "Artifact Creature — Golem", cmdrB23BulwarkOracle, 0, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, bulwark, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("targeted a creature without defender")
	}
	effectstest.B16Activate(t, g, me.ID, bulwark, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bulwark}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bulwark, "haste") || !effectstest.HasEffectiveKeyword(t, g, bulwark, "defender") {
		t.Error("the Bulwark should have haste and keep defender")
	}
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, bulwark)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want 3 combat damage (its toughness)", life, opp.Life)
	}
	if p := effectstest.EffectivePower(t, g, bulwark); p != 0 {
		t.Errorf("power %d, want 0: only the damage assignment changes", p)
	}
}

func TestCmdrB23WalkingBulwarkSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bulwark := effectstest.B12Push(g, me.ID, "Walking Bulwark", "Artifact Creature — Golem", cmdrB23BulwarkOracle, 0, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.ApaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, bulwark, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bulwark}}}); err == nil {
		t.Fatal("activated outside a main phase")
	}
}

func TestCmdrB23LandscapesFetchAndCycle(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, cycling string
		fetch                 []string
		refuse                string
	}{
		{"Bountiful Landscape", cmdrB23BountifulOracle, "{G}{U}{R}", []string{"Forest", "Island", "Mountain"}, "Swamp"},
		{"Deceptive Landscape", cmdrB23DeceptiveOracle, "{W}{B}{G}", []string{"Plains", "Swamp", "Forest"}, "Island"},
		{"Foreboding Landscape", cmdrB23ForebodingOracle, "{B}{G}{U}", []string{"Swamp", "Forest", "Island"}, "Mountain"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			land := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			for _, sub := range append(append([]string{}, tc.fetch...), tc.refuse) {
				effectstest.StapleLibraryCard(me, "Basic "+sub, "Basic Land — "+sub)
			}
			if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("fetch: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.SearchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt")
			}
			if effectstest.SearchOptionNamed(g, c, "Basic "+tc.refuse) != uuid.Nil {
				t.Errorf("the search offered a basic %s", tc.refuse)
			}
			for _, sub := range tc.fetch {
				if effectstest.SearchOptionNamed(g, c, "Basic "+sub) == uuid.Nil {
					t.Errorf("the search did not offer a basic %s", sub)
				}
			}
			effectstest.AnswerSearchNamed(t, g, me.ID, "Basic "+tc.fetch[0])
			effectstest.PassPriorityAroundTable(t, g)
			found := false
			for _, bc := range g.Battlefield.Cards {
				if bc.Name == "Basic "+tc.fetch[0] {
					found = true
					if !bc.Tapped {
						t.Error("the fetched land should enter tapped")
					}
				}
			}
			if !found {
				t.Error("the fetched land did not reach the battlefield")
			}

			effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
			id := effectstest.PushCatalogHandCard(me, tc.name, "Land", tc.oracle)
			if err := g.AddManaForEffect(me.ID, uuid.Nil, tc.cycling); err != nil {
				t.Fatalf("AddManaForEffect: %v", err)
			}
			if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("cycle: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Graveyard.Contains(id) {
				t.Error("the cycled land should be in the graveyard")
			}
		})
	}
}

func TestCmdrB23CryptOfAgadeemCountsBlackCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crypt := effectstest.PushCatalogPermanent(g, me.ID, "Crypt of Agadeem", "Land", cmdrB23CryptOracle, false)
	for _, c := range []game.Card{
		{Name: "Black A", TypeLine: "Creature — Zombie", Colors: []string{"B"}},
		{Name: "Black B", TypeLine: "Creature — Zombie", Colors: []string{"B", "G"}},
		{Name: "Green", TypeLine: "Creature — Elf", Colors: []string{"G"}},
		{Name: "Black Spell", TypeLine: "Sorcery", Colors: []string{"B"}},
	} {
		c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
		me.Graveyard.PushTop(c)
	}
	// An opponent's graveyard doesn't count.
	opp := g.Seats[1]
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Creature — Zombie", Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.ApaMana(me, "C", "C")
	if err := g.ActivateManaAbility(me.ID, crypt, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "B" || got[1] != "B" {
		t.Errorf("pool = %v, want {B}{B}", got)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == crypt {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	effectstest.TapForMana(t, g, me.ID, crypt)
	if got := effectstest.SortedPool(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool = %v, want {B}", got)
	}
}

func TestCmdrB23CryptOfAgadeemEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B31PlayLand(t, g, "Crypt of Agadeem", "Land", cmdrB23CryptOracle)
	if c, ok := effectstest.BattlefieldCard(g, id); !ok || !c.Tapped {
		t.Errorf("Crypt of Agadeem on the battlefield: %v, tapped %v; want tapped", ok, c.Tapped)
	}
}

func TestCmdrB23PiaNalaarGetsEnergyAndMakesAnAetherjet(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pia := effectstest.B12Push(g, me.ID, "Pia Nalaar, Chief Mechanic", "Legendary Creature — Human Artificer", cmdrB23PiaOracle, 2, 4)
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	golem2 := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Golem 2", TypeLine: "Artifact Creature — Golem",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.AttackWith(t, g, opp.ID, pia, golem, golem2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Counters[game.CounterEnergy]; got != 2 {
		t.Fatalf("energy %d, want 2 (one trigger for two artifact creatures hitting one player; Pia is no artifact)", got)
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PayAmountPrompt(g, me.ID)
	if c == nil || c.PayAmount == nil {
		t.Fatal("no pay-any-amount prompt")
	}
	if c.PayAmount.Min != 1 {
		t.Errorf("the payment floor is %d, want 1", c.PayAmount.Min)
	}
	if err := g.ResolvePayAmount(c.ID, me.ID, 2); err != nil {
		t.Fatalf("pay: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Counters[game.CounterEnergy]; got != 0 {
		t.Errorf("energy %d after paying 2, want 0", got)
	}
	ids := effectstest.BattlefieldIDsNamed(g, "Nalaar Aetherjet")
	if len(ids) != 1 {
		t.Fatalf("%d Nalaar Aetherjets, want 1", len(ids))
	}
	jet, _ := effectstest.BattlefieldCard(g, ids[0])
	if jet.Power != 2 || jet.Toughness != 2 || !jet.HasSubtype("Vehicle") || !jet.IsArtifact() || jet.IsCreature() {
		t.Errorf("Aetherjet = %d/%d %q, want a 2/2 Vehicle artifact", jet.Power, jet.Toughness, jet.TypeLine)
	}
	if !effectstest.HasEffectiveKeyword(t, g, ids[0], "flying") {
		t.Error("Nalaar Aetherjet should have flying")
	}
}

func TestCmdrB23SaheeliGetsEnergyForArtifactSpellsAndCopiesAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sah := effectstest.B12Push(g, me.ID, "Saheeli, Radiant Creator", "Legendary Creature — Human Artificer", cmdrB23SaheeliOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Trinket", "Artifact", "{1}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastFromHandForTest(t, g, me, "Plain Sorcery", "Sorcery", "{1}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Counters[game.CounterEnergy]; got != 1 {
		t.Fatalf("energy %d, want 1 (only the artifact spell counts)", got)
	}
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(me.ID, game.CounterEnergy, 2) })
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Counters[game.CounterEnergy]; got != 0 {
		t.Errorf("energy %d, want 0 after paying three", got)
	}
	var token game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && c.IsToken() {
			token = c
		}
	}
	if token.InstanceID == uuid.Nil {
		t.Fatal("no token copy of the Bear")
	}
	if p, tough := effectstest.EffectivePower(t, g, token.InstanceID), effectstest.EffectiveToughness(t, g, token.InstanceID); p != 5 || tough != 5 {
		t.Errorf("token %d/%d, want 5/5", p, tough)
	}
	if c := effectstest.EffectiveCardForTest(t, g, token.InstanceID); !c.IsArtifact() || !c.IsCreature() || !c.HasSubtype("Bear") {
		t.Errorf("token type line %q, want artifact creature Bear", c.TypeLine)
	}
	if !effectstest.HasEffectiveKeyword(t, g, token.InstanceID, "haste") {
		t.Error("the token should have haste")
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if _, still := effectstest.BattlefieldCard(g, token.InstanceID); still {
		t.Error("the token was not sacrificed at the end step")
	}
	_ = sah
}

func TestCmdrB23TemmetLootsOnAttackAndPumpsZombiesOnDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	temmet := effectstest.B12Push(g, me.ID, "Temmet, Naktamun's Will", "Legendary Creature — Zombie Wizard", cmdrB23TemmetOracle, 4, 4)
	zombie := effectstest.B12Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, temmet, zombie)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d, want %d after the draw", got, hand+1)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d after the discard", got, hand)
	}
	if p := effectstest.EffectivePower(t, g, zombie); p != 3 {
		t.Errorf("Zombie power %d, want 3 (drew a card)", p)
	}
	if p := effectstest.EffectivePower(t, g, temmet); p != 5 {
		t.Errorf("Temmet power %d, want 5", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("Bear power %d, want 2", p)
	}
}

func TestCmdrB23ProphetOfTheScarabDrawsTheGreaterOfZombiesOrGraveyardZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Zombie A", "Creature — Zombie", 2, 2)
	for i := 0; i < 4; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Zombie", TypeLine: "Creature — Zombie", Owner: me.ID, Controller: me.ID})
	}
	hand := me.Hand.Size()
	effectstest.CastAndResolveCreature(t, g, "Prophet of the Scarab", "Creature — Zombie Wizard", cmdrB23ProphetOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d, want 4 (graveyard Zombies beat 2 on the battlefield)", got)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.B12Creature(g2, me2.ID, "Zombie A", "Creature — Zombie", 2, 2)
	effectstest.B12Creature(g2, me2.ID, "Zombie B", "Creature — Zombie", 2, 2)
	hand2 := me2.Hand.Size()
	effectstest.CastAndResolveCreature(t, g2, "Prophet of the Scarab", "Creature — Zombie Wizard", cmdrB23ProphetOracle)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := me2.Hand.Size() - hand2; got != 3 {
		t.Errorf("drew %d, want 3 (two Zombies and the Prophet)", got)
	}
}

func TestCmdrB23AdaptiveOmnitoolPumpsPerArtifactAndDigsOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tool := effectstest.PushCatalogPermanent(g, me.ID, "Adaptive Omnitool", "Artifact — Equipment", cmdrB23OmnitoolOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, me.ID, "Trinket", "Artifact")
	effectstest.BrAttach(t, g, tool, bear)
	// Omnitool and Trinket: +2/+2.
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("equipped Bear %d/%d, want 4/4", p, tough)
	}
	g.WithWriteLock(func() { me.Library.Cards = nil })
	lib := effectstest.TrekB48Library(me,
		game.Card{Name: "Land", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Gem", TypeLine: "Artifact"},
		game.Card{Name: "Spell", TypeLine: "Instant"},
		game.Card{Name: "Four", TypeLine: "Instant"},
		game.Card{Name: "Five", TypeLine: "Instant"},
		game.Card{Name: "Six", TypeLine: "Instant"},
		game.Card{Name: "Seventh", TypeLine: "Artifact"},
	)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.LatestChooseCards(g, me.ID)
	if ch == nil {
		t.Fatal("no pick for the artifact")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[0]}); err == nil {
		t.Error("a land was accepted")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[6]}); err == nil {
		t.Error("the seventh card (not looked at) was accepted")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(lib[1]) {
		t.Error("the artifact is not in hand")
	}
	if n := len(me.Library.Cards); n != 6 || me.Library.Cards[n-1].InstanceID != lib[6] {
		t.Errorf("library has %d cards; want the seventh on top", n)
	}
}
