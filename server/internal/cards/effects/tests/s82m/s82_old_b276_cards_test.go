package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 276 (Eldritch Moon): Galvanic Bombardment, Assembled
// Alphas, Nahiri's Wrath, Make Mischief, Permeating Mass, Stensia Banquet,
// Take Inventory, Emrakul's Influence, Ulvenwald Observer, Courageous
// Outrider, Spirit of the Hunt, Fogwalker.

const (
	oldB276Galvanic  = "13d8347f-adb9-41b5-a7b0-a6153bcf2d6c"
	oldB276Alphas    = "151a681d-a350-4d1d-86cd-b80d58602810"
	oldB276Wrath     = "420091b5-b966-4233-877d-e9143481e97e"
	oldB276Mischief  = "5de50aa5-3b30-4d7d-9203-bc8fc0f8fb34"
	oldB276Mass      = "5e719329-612e-4ad5-ac95-6eb5328738ab"
	oldB276Banquet   = "089e7086-3f8e-4b10-a660-fcf4fc99604e"
	oldB276Inventory = "925b3d1b-98cd-44f5-8aa0-e2b17df2808a"
	oldB276Influence = "92eb07ca-ec57-48c3-98ce-9e878bd968c1"
	oldB276Observer  = "dff3cfff-69f9-4afd-af28-b75504d30ffa"
	oldB276Outrider  = "00915480-ffa8-4930-a158-b1c1f75b7876"
	oldB276Spirit    = "3a81a3a6-4940-4d69-94c1-d431023b275c"
	oldB276Fogwalker = "c597cfae-3f92-4d31-8bc1-5a2da995f1bf"
)

func TestOldB276CardsAreFullWithNoCaveats(t *testing.T) {
	for _, oracle := range []string{oldB276Galvanic, oldB276Alphas, oldB276Wrath, oldB276Mischief, oldB276Mass, oldB276Banquet,
		oldB276Inventory, oldB276Influence, oldB276Observer, oldB276Outrider, oldB276Spirit, oldB276Fogwalker} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB276GalvanicBombardmentCountsItsNamesakesInTheGraveyard(t *testing.T) {
	for _, copies := range []int{0, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		for range copies {
			effectstest.GraveyardCardOf(g, me, "Galvanic Bombardment", "Instant", "{R}", oldB276Galvanic)
		}
		effectstest.GraveyardCardOf(g, me, "Other Bolt", "Instant", "{R}", "")
		ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 9, 9)
		effectstest.AdvanceToMain(t, g)
		effectstest.CastCatalogSpell(t, g, "Galvanic Bombardment", "Instant", oldB276Galvanic, effectstest.CardRefs(ogre))
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.P1g2WantDamage(t, g, ogre, 2+copies, "the Ogre")
	}
}

func TestOldB276AssembledAlphasDamagesTheCreatureAndItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	alphas := effectstest.B12Push(g, opp.ID, "Assembled Alphas", "Creature — Wolf", oldB276Alphas, 5, 5)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 9)
	life := me.Life
	effectstest.WoeS2Block(g, opp.ID, alphas, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, attacker); got != 3 {
		t.Errorf("the blocked creature has %d damage, want 3", got)
	}
	if me.Life != life-3 {
		t.Errorf("its controller's life %d → %d, want -3", life, me.Life)
	}
}

func TestOldB276NahirisWrathDiscardsXAndDealsTheirTotalManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me.Hand.Cards = nil
	spell := effectstest.HandCardOf(g, me, "Nahiri's Wrath", "Sorcery", "{2}{R}", oldB276Wrath)
	d1 := effectstest.HandCardOf(g, me, "Two", "Instant", "{1}{R}", "")
	d2 := effectstest.HandCardOf(g, me, "Three", "Instant", "{2}{R}", "")
	keep := effectstest.HandCardOf(g, me, "Keep", "Instant", "{R}", "")
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Ogre", 9, 9)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Ogre", 9, 9)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Ogre", 9, 9)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{R}")
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 2, DiscardIDs: []uuid.UUID{d1, d2},
		Targets: effectstest.CardRefs(a, b, c)})
	if err == nil {
		t.Fatal("three targets accepted with X = 2")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 2, DiscardIDs: []uuid.UUID{d1, d2},
		Targets: effectstest.CardRefs(a, b)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(d1) || !me.Graveyard.Contains(d2) || !me.Hand.Contains(keep) {
		t.Error("the two named cards were not the ones discarded")
	}
	effectstest.P1g2WantDamage(t, g, a, 5, "A")
	effectstest.P1g2WantDamage(t, g, b, 5, "B")
	effectstest.P1g2WantDamage(t, g, c, 0, "C")
}

func TestOldB276MakeMischiefPingsAndMakesADevilThatPingsOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life := opp.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Make Mischief", "Sorcery", oldB276Mischief,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d → %d, want -1", life, opp.Life)
	}
	devil := effectstest.BattlefieldIDNamed(g, me.ID, "Devil")
	if devil == uuid.Nil {
		t.Fatal("no Devil token")
	}
	if p, tough := effectstest.EffectivePower(t, g, devil), effectstest.EffectiveToughness(t, g, devil); p != 1 || tough != 1 {
		t.Errorf("Devil is %d/%d, want 1/1", p, tough)
	}
	effectstest.KillCreature(t, g, me.ID, devil)
	g.RunStateChecksForTest()
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d after the Devil died, want %d", opp.Life, life-2)
	}
}

func TestOldB276PermeatingMassCopiesTheCreatureItDamages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mass := effectstest.B12Push(g, me.ID, "Permeating Mass", "Creature — Spirit", oldB276Mass, 1, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 9)
	effectstest.DealNoncombatDamage(g, me.ID, mass, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.EffectiveCardForTest(t, g, victim); c.Name == "Permeating Mass" {
		t.Fatal("noncombat damage copied the creature")
	}
	effectstest.DealCombatDamage(g, me.ID, mass, victim)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.EffectiveCardForTest(t, g, victim)
	if c.Name != "Permeating Mass" || !c.HasSubtype("Spirit") {
		t.Errorf("the creature is %q %v, want a copy of Permeating Mass", c.Name, c.Effective().Subtypes)
	}
}

func TestOldB276StensiaBanquetCountsVampiresAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Creature(g, me.ID, "V1", "Creature — Vampire", 1, 1)
	effectstest.B12Creature(g, me.ID, "V2", "Creature — Vampire Rogue", 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, opp.ID, "TheirV", "Creature — Vampire", 1, 1)
	effectstest.SeedLibrary(me, "A", "B")
	life, hand := opp.Life, me.Hand.Size()
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Stensia Banquet", "Sorcery", oldB276Banquet,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d → %d, want -2 (two Vampires you control)", life, opp.Life)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestOldB276TakeInventoryDrawsOnePlusItsNamesakes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	effectstest.GraveyardCardOf(g, me, "Take Inventory", "Sorcery", "{1}{U}", oldB276Inventory)
	effectstest.GraveyardCardOf(g, me, "Take Inventory", "Sorcery", "{1}{U}", oldB276Inventory)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Take Inventory", "Sorcery", oldB276Inventory, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards, want 1 + 2 namesakes = 3", got)
	}
}

func TestOldB276EmrakulsInfluenceDrawsForABigEldraziCreatureSpellOnly(t *testing.T) {
	for _, tc := range []struct {
		name, typ, cost string
		draws           int
	}{
		{"Big Eldrazi", "Creature — Eldrazi", "{6}{G}", 2},
		{"Small Eldrazi", "Creature — Eldrazi", "{5}{G}", 0},
		{"Big Giant", "Creature — Giant", "{7}", 0},
		{"Big Eldrazi Sorcery", "Sorcery", "{7}", 0},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.PushCatalogPermanent(g, me.ID, "Emrakul's Influence", "Enchantment", oldB276Influence, false)
		effectstest.SeedLibrary(me, "A", "B", "C")
		effectstest.AdvanceToMain(t, g)
		hand := me.Hand.Size()
		effectstest.CastWithCost(t, g, tc.name, tc.typ, tc.cost, "")
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != tc.draws {
			t.Errorf("%s: drew %d, want %d", tc.name, got, tc.draws)
		}
	}
}

func TestOldB276UlvenwaldObserverDrawsForABigCreatureDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Ulvenwald Observer", "Creature — Treefolk", oldB276Observer, false)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 5, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 5, 5)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	effectstest.KillCreature(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("a toughness-3 creature of yours or an opponent's creature drew (%d → %d)", hand, me.Hand.Size())
	}
	effectstest.KillCreature(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card for the toughness-4 creature", hand, me.Hand.Size())
	}
}

func TestOldB276OutriderTakesOneHumanFromTheTopFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	human := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: human, Name: "Human Soldier", TypeLine: "Creature — Human Soldier", Owner: me.ID, Controller: me.ID})
	effectstest.CastHoldCreature(t, g, "Courageous Outrider", "Creature — Human Scout", oldB276Outrider, 3, 4)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMax != 1 || len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != human {
		t.Fatalf("want a take-one prompt offering only the Human: %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{human}); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(human) {
		t.Fatal("the Human card did not go to hand")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom {
		t.Fatalf("the rest are not ordered on the bottom: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, order.ScryCards, nil); err != nil {
		t.Fatal(err)
	}
	_ = ids
}

func TestOldB276SpiritOfTheHuntBoostsOtherWolvesAndWerewolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	wolf := effectstest.B12Creature(g, me.ID, "Wolf", "Creature — Wolf", 2, 2)
	were := effectstest.B12Creature(g, me.ID, "Were", "Creature — Human Werewolf", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Wolf", "Creature — Wolf", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastAndResolveCreature(t, g, "Spirit of the Hunt", "Creature — Wolf Spirit", oldB276Spirit)
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{wolf: 5, were: 5, bear: 2, theirs: 2} {
		if got := effectstest.EffectiveToughness(t, g, id); got != want {
			t.Errorf("toughness %d, want %d", got, want)
		}
	}
}

func TestOldB276FogwalkerFreezesAnOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastHoldCreature(t, g, "Fogwalker", "Creature — Spirit", oldB276Fogwalker, 1, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p != nil && effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("your own creature is offered")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HmlUntapSkips(t, g, bear); len(got) != 1 {
		t.Errorf("the opposing creature has %d next-untap marks, want 1", len(got))
	}
	if got := effectstest.HmlUntapSkips(t, g, mine); len(got) != 0 {
		t.Errorf("your own creature has %d next-untap marks, want 0", len(got))
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, effectstest.BattlefieldIDNamed(g, me.ID, "Fogwalker"), "skulk") {
		t.Error("no skulk")
	}
}
