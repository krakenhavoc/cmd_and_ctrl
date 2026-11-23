package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// annAmounts is the annihilator numbers a battlefield card has.
func annAmounts(t *testing.T, g *game.Game, id uuid.UUID) []int {
	t.Helper()
	c := effectstest.AnnCard(t, g, id)
	return game.AnnihilatorAmounts(&c)
}

func TestPrintedAnnihilatorPoolCardsCarryTheirNumber(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	for _, c := range []struct {
		name, oracle string
		want         int
	}{
		{"Artisan of Kozilek", effectstest.ArtisanOfKozilekOracle, 2},
		{"Nulldrifter", effectstest.NulldrifterOracle, 1},
		{"Eldrazi Ravager", effectstest.EldraziRavagerOracle, 1},
		{"Flayer of Loyalties", effectstest.FlayerOfLoyaltiesOracle, 2},
		{"It That Betrays", effectstest.ItThatBetraysOracle, 2},
		{"Pathrazer of Ulamog", effectstest.PathrazerOracle, 3},
		{"Ulamog's Crusher", effectstest.UlamogsCrusherOracle, 2},
		{"Breaker of Creation", effectstest.BreakerOfCreationOracle, 2},
		{"Hand of Emrakul", effectstest.HandOfEmrakulOracle, 1},
		{"Spawnsire of Ulamog", effectstest.SpawnsireOracle, 1},
		{"Hideous Taskmaster", effectstest.HideousTaskmasterOracle, 1},
	} {
		id := effectstest.PushAnnCatalogCreature(g, me, c.name, c.oracle, 5, 5)
		if got := annAmounts(t, g, id); len(got) != 1 || got[0] != c.want {
			t.Errorf("%s: annihilator %v, want [%d]", c.name, got, c.want)
		}
	}
}

func TestArtisanOfKozilekReturnsACreatureOnCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	id := effectstest.PutInHand(me, game.Card{Name: "Artisan of Kozilek", TypeLine: "Creature — Eldrazi",
		OracleID: effectstest.ArtisanOfKozilekOracle, ManaCost: "{9}", Power: 10, Toughness: 9})
	effectstest.ToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AnnOnBattlefield(g, bear) {
		t.Error("the creature card did not return to the battlefield")
	}
}

func TestNulldrifterEvokedDrawsTwoAndIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PutInHand(me, game.Card{Name: "Nulldrifter", TypeLine: "Creature — Eldrazi Elemental",
		OracleID: effectstest.NulldrifterOracle, ManaCost: "{7}", Power: 4, Toughness: 4})
	effectstest.ToMain(t, g)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "evoke"}); err != nil {
		t.Fatalf("CastSpell evoke: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want +2 drawn -1 cast", hand, got)
	}
	if effectstest.AnnOnBattlefield(g, id) || !me.Graveyard.Contains(id) {
		t.Error("an evoked Nulldrifter was not sacrificed")
	}
}

func TestEldraziRavagerReturnsFromTheGraveyardForTwoEldrazi(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	ravager := effectstest.PushGraveyardPermanent(me, "Eldrazi Ravager", "Creature — Eldrazi", "{5}{C}")
	for i := range me.Graveyard.Cards {
		if me.Graveyard.Cards[i].InstanceID == ravager {
			me.Graveyard.Cards[i].OracleID = effectstest.EldraziRavagerOracle
		}
	}
	a := effectstest.PushAnnCatalogCreature(g, me.ID, "Eldrazi Spawn", "", 0, 1)
	b := effectstest.PushAnnCatalogCreature(g, me.ID, "Eldrazi Scion", "", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, ravager, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ravager) {
		t.Error("the Ravager did not return to hand")
	}
	if effectstest.AnnOnBattlefield(g, a) || effectstest.AnnOnBattlefield(g, b) {
		t.Error("the two Eldrazi were not sacrificed")
	}
}

func TestFlayerOfLoyaltiesStealsAndMakesA1010Annihilator(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	id := effectstest.PutInHand(me, game.Card{Name: "Flayer of Loyalties", TypeLine: "Creature — Eldrazi",
		OracleID: effectstest.FlayerOfLoyaltiesOracle, ManaCost: "{10}", Power: 10, Toughness: 10})
	effectstest.ToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.AnnCard(t, g, bear)
	if c.Controller != me.ID || c.Tapped {
		t.Errorf("controller %s tapped %v; want mine and untapped", c.Controller, c.Tapped)
	}
	if c.CurrentPower() != 10 || c.CurrentToughness() != 10 {
		t.Errorf("P/T %d/%d, want 10/10", c.CurrentPower(), c.CurrentToughness())
	}
	if !game.HasKeyword(&c, "trample") || !game.HasKeyword(&c, "haste") {
		t.Error("the stolen creature lacks trample or haste")
	}
	if got := game.AnnihilatorAmounts(&c); len(got) != 1 || got[0] != 2 {
		t.Errorf("annihilator %v, want [2]", got)
	}
}

func TestAzlaskCountsColorlessDeathsAndPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	azlask := effectstest.PushAnnCatalogCreature(g, me.ID, "Azlask, the Swelling Scourge", effectstest.AzlaskOracle, 2, 2)
	colorless := effectstest.PushAnnCatalogCreature(g, me.ID, "Eldrazi Drone", "", 2, 2)
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear",
		TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	spawn := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Eldrazi Spawn",
		TypeLine: "Token Creature — Eldrazi Spawn", Power: 0, Toughness: 1, Owner: me.ID, Controller: me.ID})
	for _, id := range []uuid.UUID{colorless, green} {
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatalf("destroy: %v", err)
			}
		})
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effects.ExperienceCounters(g, me.ID); got != 1 {
		t.Fatalf("experience counters = %d, want 1 (the colourless death only)", got)
	}
	effectstest.ToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, azlask, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.AnnCard(t, g, azlask); c.CurrentPower() != 3 {
		t.Errorf("Azlask power %d, want 2+1", c.CurrentPower())
	}
	s := effectstest.AnnCard(t, g, spawn)
	if !game.HasKeyword(&s, "indestructible") || len(game.AnnihilatorAmounts(&s)) != 1 {
		t.Error("the Spawn did not gain indestructible and annihilator 1")
	}
	if a := effectstest.AnnCard(t, g, azlask); len(game.AnnihilatorAmounts(&a)) != 0 {
		t.Error("Azlask is not a Scion or Spawn and must not gain annihilator")
	}
}

func TestEldraziConscriptionAddsASecondAnnihilator(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.PushAnnihilatorCreature(g, me.ID, "annihilator 4")
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Eldrazi Conscription", "Kindred Enchantment — Eldrazi Aura", effectstest.EldraziConscriptionOracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	c := effectstest.AnnCard(t, g, host)
	if c.CurrentPower() != 15 || !game.HasKeyword(&c, "trample") {
		t.Errorf("host is %d power, trample %v; want 15 and trample", c.CurrentPower(), game.HasKeyword(&c, "trample"))
	}
	if got := game.AnnihilatorAmounts(&c); len(got) != 2 {
		t.Errorf("annihilator %v, want the printed 4 and the granted 2", got)
	}
}

func TestItThatBetraysTakesAnOpponentsSacrificedPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushAnnCatalogCreature(g, me.ID, "It That Betrays", effectstest.ItThatBetraysOracle, 11, 11)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spawn",
		TypeLine: "Token Creature — Eldrazi Spawn", Owner: opp.ID, Controller: opp.ID, Toughness: 1})
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Forest", "Basic Land — Forest", "", false)
	for _, s := range []struct{ who, id uuid.UUID }{{opp.ID, land}, {opp.ID, token}, {me.ID, mine}} {
		if err := g.SacrificePermanent(s.who, s.id); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	c, ok := effectstest.BattlefieldCard(g, land)
	if !ok || c.Controller != me.ID {
		t.Fatalf("the opponent's sacrificed land is not mine on the battlefield (on: %v)", ok)
	}
	if effectstest.AnnOnBattlefield(g, mine) {
		t.Error("my own sacrifice came back")
	}
}

func TestPathrazerNeedsThreeBlockersAndCrusherMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0].ID
	p := effectstest.PushAnnCatalogCreature(g, me, "Pathrazer of Ulamog", effectstest.PathrazerOracle, 9, 9)
	if rules := game.CatalogBlockRules(effectstest.AnnCard(t, g, p).OracleID); len(rules) != 1 {
		t.Errorf("Pathrazer declares %d block rules, want the three-blocker minimum", len(rules))
	}
	if statics := game.CatalogStaticAbilities(effectstest.UlamogsCrusherOracle); len(statics) == 0 {
		t.Errorf("Ulamog's Crusher declares %d statics, want attacks each combat", len(statics))
	}
}

func TestBreakerOfCreationGainsLifePerColorlessPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushAnnCatalogCreature(g, me.ID, "Drone", "", 1, 1)
	effectstest.PushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", "", false)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear",
		TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	id := effectstest.PutInHand(me, game.Card{Name: "Breaker of Creation", TypeLine: "Creature — Eldrazi",
		OracleID: effectstest.BreakerOfCreationOracle, ManaCost: "{6}{C}{C}", Power: 8, Toughness: 4})
	effectstest.ToMain(t, g)
	life := me.Life
	colorless := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsColorless() {
			colorless++
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != colorless {
		t.Errorf("gained %d life, want %d", got, colorless)
	}
}

func TestHandOfEmrakulOffersTheSpawnSacrifice(t *testing.T) {
	ac := game.AlternativeCostByKey(effectstest.HandOfEmrakulOracle, "sacrifice")
	if ac == nil || ac.Sacrifice == nil || ac.Sacrifice.Min != 4 || ac.Sacrifice.Max != 4 || ac.ManaCost != "" {
		t.Fatalf("alternative cost = %+v, want sacrifice exactly four and no mana", ac)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	var spawn []uuid.UUID
	for i := 0; i < 4; i++ {
		spawn = append(spawn, effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Eldrazi Spawn",
			TypeLine: "Token Creature — Eldrazi Spawn", Toughness: 1, Owner: me.ID, Controller: me.ID}))
	}
	id := effectstest.PutInHand(me, game.Card{Name: "Hand of Emrakul", TypeLine: "Creature — Eldrazi",
		OracleID: effectstest.HandOfEmrakulOracle, ManaCost: "{9}", Power: 7, Toughness: 7})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "sacrifice", AltCostIDs: spawn, Strict: true}); err != nil {
		t.Fatalf("CastSpell for four Spawn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AnnOnBattlefield(g, id) {
		t.Error("Hand of Emrakul did not resolve")
	}
	for _, s := range spawn {
		if effectstest.AnnOnBattlefield(g, s) {
			t.Error("a Spawn was not sacrificed")
		}
	}
}

func TestNazgulBattleMaceGrantsAndTakesUnlessLifeIsPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.PushAnnCatalogCreature(g, me.ID, "Knight", "", 3, 3)
	mace := effectstest.PushCatalogPermanent(g, me.ID, "Nazgûl Battle-Mace", "Artifact — Equipment", effectstest.NazgulBattleMaceOracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(mace, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	c := effectstest.AnnCard(t, g, host)
	if !game.HasKeyword(&c, "menace") || !game.HasKeyword(&c, "deathtouch") || len(game.AnnihilatorAmounts(&c)) != 1 {
		t.Fatalf("the equipped creature lacks menace, deathtouch or annihilator 1: %v", c.Effective().Abilities)
	}

	kept := effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	taken := effectstest.PushCatalogPermanent(g, opp.ID, "Swamp", "Basic Land — Swamp", "", false)
	life := opp.Life
	if err := g.SacrificePermanent(opp.ID, kept); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, opp.ID, 0) // pay 3 life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 || effectstest.AnnOnBattlefield(g, kept) {
		t.Errorf("paying 3 life: life %d -> %d, card on battlefield %v", life, opp.Life, effectstest.AnnOnBattlefield(g, kept))
	}
	if err := g.SacrificePermanent(opp.ID, taken); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, opp.ID, 1) // don't pay
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, taken); !ok || c.Controller != me.ID {
		t.Error("not paying did not give me the card")
	}
}

func TestSpawnsireMakesTwoSpawn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	id := effectstest.PushAnnCatalogCreature(g, me.ID, "Spawnsire of Ulamog", effectstest.SpawnsireOracle, 7, 11)
	before := effectstest.CardsNamed(g, "Eldrazi Spawn", me.ID)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CardsNamed(g, "Eldrazi Spawn", me.ID) - before; got != 2 {
		t.Errorf("made %d Spawn, want 2", got)
	}
}

func TestHideousTaskmasterTakesOneCreaturePerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[0], g.Seats[1], g.Seats[2]
	ca := effectstest.PushCatalogPermanent(g, a.ID, "Bear A", "Creature — Bear", "", false)
	cb := effectstest.PushCatalogPermanent(g, b.ID, "Bear B", "Creature — Bear", "", false)
	id := effectstest.PutInHand(me, game.Card{Name: "Hideous Taskmaster", TypeLine: "Creature — Eldrazi",
		OracleID: effectstest.HideousTaskmasterOracle, ManaCost: "{6}{R}", Power: 7, Toughness: 2})
	effectstest.ToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// One "up to one" clause per opponent, asked in seat order: seat
	// 1's creature, then seat 2's. Seat 3 controls nothing, so its
	// clause is skipped without a prompt.
	for i, pick := range []uuid.UUID{ca, cb} {
		p := effectstest.LatestPickTarget(g, me.ID)
		if p == nil {
			t.Fatalf("no target prompt for opponent %d", i+1)
		}
		if !effectstest.HasID(p.PickTargetCards, pick) || p.PickTargetMax != 1 || p.PickTargetMin != 0 {
			t.Fatalf("clause %d offers %v (%d..%d), want up to one of that opponent's creatures",
				i+1, p.PickTargetCards, p.PickTargetMin, p.PickTargetMax)
		}
		if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: pick}}); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatalf("a prompt for the opponent with no creatures: %q", p.Reason)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{ca, cb} {
		c := effectstest.AnnCard(t, g, id)
		if c.Controller != me.ID || !game.HasKeyword(&c, "haste") || len(game.AnnihilatorAmounts(&c)) != 1 {
			t.Errorf("%s: controller %s, haste %v, annihilator %v", c.Name, c.Controller,
				game.HasKeyword(&c, "haste"), game.AnnihilatorAmounts(&c))
		}
	}
}

func TestIdolOfFalseGodsGrowsAndBecomesAnAnnihilator(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	idol := effectstest.PushCatalogPermanent(g, me.ID, "Idol of False Gods", "Kindred Artifact — Eldrazi", effectstest.IdolOfFalseGodsOracle, false)
	drone := effectstest.PushAnnCatalogCreature(g, me.ID, "Drone", "", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(drone); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, idol, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters after an Eldrazi died = %d, want 1", got)
	}
	if c := effectstest.AnnCard(t, g, idol); c.IsCreature() {
		t.Error("the Idol is a creature at one counter")
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(idol, game.CounterPlusOne, 7); err != nil {
			t.Fatalf("add counters: %v", err)
		}
	})
	c := effectstest.AnnCard(t, g, idol)
	if !c.IsCreature() || c.CurrentPower() != 8 || c.CurrentToughness() != 8 {
		t.Errorf("at eight counters: creature %v, %d/%d; want an 8/8 creature", c.IsCreature(), c.CurrentPower(), c.CurrentToughness())
	}
	if got := game.AnnihilatorAmounts(&c); len(got) != 1 || got[0] != 2 {
		t.Errorf("annihilator %v, want [2]", got)
	}
}
