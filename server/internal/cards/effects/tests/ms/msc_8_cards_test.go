package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_8_cards_test.go — Marvel Super Heroes Commander, eighth batch
// (#4455). Each test plays the real catalog entry and asserts what its
// text says.

const (
	msc8BigWheelOracle   = "b0941bc2-f482-4821-a96b-720b08c7334a"
	msc8AirForceOracle   = "5d7eea87-b60f-4ab9-b485-30aadda0cd79"
	msc8AtlasOracle      = "efd5f51f-931d-46a9-9773-f74cfc318c23"
	msc8AuntMayOracle    = "bdb769b5-c861-4ec1-ad6b-e2a56ca05fd1"
	msc8ElusiveOracle    = "f334087a-deda-4e9e-ab58-68f7e226d7ce"
	msc8MotorcycleOracle = "19e60344-57a8-41e0-8eca-87dbedf5a81a"
	msc8SpiderGirlOracle = "7a6c4216-7aff-41d4-8d79-bd1ac0b263ae"
	msc8BushmasterOracle = "21067a25-f749-4b79-bc6d-820a91af2834"
	msc8SkyboundOracle   = "e9566303-15cf-4444-a044-fc19540f21b6"
	msc8BerthaOracle     = "32654ffa-b869-4f8b-a153-4242e8f9d92e"
	msc8ReformedOracle   = "c51c429d-d8d7-40d6-9760-72c26acaed42"
)

func TestMsc8EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		msc8BigWheelOracle:   "Big Wheel",
		msc8AirForceOracle:   "Ant-Man's Air Force",
		msc8AtlasOracle:      "Atlas, Sizable Stooge",
		msc8AuntMayOracle:    "Aunt May",
		msc8ElusiveOracle:    "Ant-Man, Elusive Avenger",
		msc8MotorcycleOracle: "Captain America's Motorcycle",
		msc8SpiderGirlOracle: "Amazing Spider-Girl",
		msc8BushmasterOracle: "Bushmaster, Coiled Henchman",
		msc8SkyboundOracle:   "Captain America, Skybound",
		msc8BerthaOracle:     "Big Bertha",
		msc8ReformedOracle:   "Ant-Man, Reformed Rogue",
	}
	for oracle, name := range want {
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

// msc8Cast puts a card in the active player's hand and casts it with no
// payment, resolving the spell and leaving any trigger on its prompt.
func msc8Cast(t *testing.T, g *game.Game, me *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = me.ID, me.ID
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandAndResolve(t, g, me, c)
	return c.InstanceID
}

func TestMsc8BigWheelLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	pitch := effectstest.HandCard(me, "Pitch", "Sorcery")
	msc8Cast(t, g, me, game.Card{Name: "Big Wheel", TypeLine: "Artifact — Vehicle", OracleID: msc8BigWheelOracle, Power: 4, Toughness: 4})
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discarded card is not in the graveyard")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand is %d after discard and draw, want %d", got, hand)
	}

	// Declining the discard draws nothing.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.SeedLibrary(me2, "One", "Two", "Three")
	effectstest.HandCard(me2, "Pitch", "Sorcery")
	msc8Cast(t, g2, me2, game.Card{Name: "Big Wheel", TypeLine: "Artifact — Vehicle", OracleID: msc8BigWheelOracle, Power: 4, Toughness: 4})
	hand2 := me2.Hand.Size()
	effectstest.AnswerDiscard(t, g2, me2.ID)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := me2.Hand.Size(); got != hand2 {
		t.Errorf("declining changed the hand: %d, want %d", got, hand2)
	}
}

func TestMsc8AntMansAirForceShrinksATargetWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flyer := effectstest.B12Push(g, me.ID, "Ant-Man's Air Force", "Creature — Insect", msc8AirForceOracle, 2, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, flyer)
	effectstest.PickIfAsked(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, victim); got != 2 {
		t.Errorf("the target has power %d, want 2", got)
	}
	effectstest.AssertKeywords(t, g, flyer, "flying")
}

func TestMsc8AtlasGainsLifeForEachBigCreatureWhenAttackingAndBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	atlas := effectstest.B12Push(g, me.ID, "Atlas, Sizable Stooge", "Legendary Creature — Human Villain", msc8AtlasOracle, 5, 4)
	effectstest.B12Creature(g, me.ID, "Four Power", "Creature — Ogre", 4, 4)
	effectstest.B12Creature(g, me.ID, "Two Power", "Creature — Bear", 2, 2)
	life := me.Life
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, atlas)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("attacking gained %d life, want 2", got)
	}

	// Blocking: the opponent attacks, the Atlas controller blocks.
	g2 := effectstest.NewCatalogGame(t)
	attacker, defender := g2.Seats[0], g2.Seats[1]
	atlas2 := effectstest.B12Push(g2, defender.ID, "Atlas, Sizable Stooge", "Legendary Creature — Human Villain", msc8AtlasOracle, 5, 4)
	bear := effectstest.B12Creature(g2, attacker.ID, "Their Bear", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g2, defender.ID, bear)
	effectstest.AdvanceTo(t, g2, game.StepDeclareBlockers)
	life2 := defender.Life
	if err := g2.DeclareBlocker(atlas2, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g2)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := defender.Life - life2; got != 1 {
		t.Errorf("blocking gained %d life, want 1", got)
	}
}

func TestMsc8AuntMayGainsLifeAndBuffsSpiders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Aunt May", "Legendary Creature — Human Citizen", msc8AuntMayOracle, 0, 2)
	life := me.Life
	spider := msc8Cast(t, g, me, game.Card{Name: "Web Crawler", TypeLine: "Creature — Spider Hero", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 1 {
		t.Errorf("a Spider entering gained %d life, want 1", got)
	}
	if got := effectstest.CountersOn(g, spider, game.CounterPlusOne); got != 1 {
		t.Errorf("the Spider has %d +1/+1 counters, want 1", got)
	}

	life = me.Life
	bear := msc8Cast(t, g, me, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 1 {
		t.Errorf("a Bear entering gained %d life, want 1", got)
	}
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 0 {
		t.Errorf("the Bear has %d counters, want 0", got)
	}
}

func TestMsc8AntManElusiveAvengerIsSkulkAndMakesTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ant := effectstest.B12Push(g, me.ID, "Ant-Man, Elusive Avenger", "Legendary Creature — Human Rogue Hero", msc8ElusiveOracle, 1, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big Wall", "Creature — Wall", 2, 8)
	small := effectstest.B12Creature(g, opp.ID, "Small Wall", "Creature — Wall", 1, 8)
	effectstest.BrAttack(t, g, ant)
	effectstest.BrRefusal(t, g.DeclareBlocker(big, ant), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(small, ant); err != nil {
		t.Errorf("a creature with equal power should block: %v", err)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	ant2 := effectstest.B12Push(g2, me2.ID, "Ant-Man, Elusive Avenger", "Legendary Creature — Human Rogue Hero", msc8ElusiveOracle, 1, 2)
	before := effectstest.CountNamed(g2, "Treasure")
	effectstest.DealCombatDamageToPlayer(g2, ant2, g2.Seats[1].ID, 3)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.CountNamed(g2, "Treasure") - before; got != 3 {
		t.Errorf("created %d Treasures, want 3", got)
	}
}

func TestMsc8CaptainAmericasMotorcyclePumpsATargetWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	msc8Cast(t, g, me, game.Card{Name: "Captain America's Motorcycle", TypeLine: "Artifact — Vehicle", OracleID: msc8MotorcycleOracle, Power: 3, Toughness: 1})
	effectstest.PickIfAsked(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("the Bear has power %d, want 4", got)
	}
}

func TestMsc8AmazingSpiderGirlWebSlings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	girl := effectstest.HandCardFull(me, "Amazing Spider-Girl", "Legendary Creature — Spider Human Hero", "{3}{W}{W}", msc8SpiderGirlOracle, []string{"W"})
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == girl {
			me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 5, 4
		}
	}
	tapped := effectstest.PushTappedVanillaCreature(g, me.ID, "Tapped Bear", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}{W}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.CastSpell(me.ID, girl, game.CastSpellParams{
		Strict:          true,
		AlternativeCost: "web_slinging",
		AltCostIDs:      []uuid.UUID{tapped},
	}); err != nil {
		t.Fatalf("web-slinging cast: %v", err)
	}
	if !me.Hand.Contains(tapped) {
		t.Error("the tapped creature did not return to hand")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(girl) {
		t.Fatal("Amazing Spider-Girl did not resolve")
	}
	effectstest.AssertKeywords(t, g, girl, "flying", "vigilance")
}

func TestMsc8BushmasterGivesDeathtouchToCreaturesWithCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bush := effectstest.B12Push(g, me.ID, "Bushmaster, Coiled Henchman", "Legendary Creature — Snake Human Villain", msc8BushmasterOracle, 2, 1)
	effectstest.AssertKeywords(t, g, bush, "deathtouch")
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	if effectstest.EffectiveAbilitiesContain(t, g, plain, "deathtouch") {
		t.Error("a creature with no counter has deathtouch")
	}
	g.WithWriteLock(func() {
		if err := (effects.AddCounter{Target: plain, Kind: game.CounterPlusOne, N: 1}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatalf("AddCounter: %v", err)
		}
	})
	if !effectstest.EffectiveAbilitiesContain(t, g, plain, "deathtouch") {
		t.Error("a creature with a +1/+1 counter lacks deathtouch")
	}
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = effects.AddCounter{Target: theirs, Kind: game.CounterPlusOne, N: 1}.Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "deathtouch") {
		t.Error("an opponent's creature got deathtouch")
	}
}

func TestMsc8CaptainAmericaSkyboundBattalion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cap := effectstest.B12Push(g, me.ID, "Captain America, Skybound", "Legendary Creature — Human Soldier Hero", msc8SkyboundOracle, 3, 5)
	a := effectstest.B12Creature(g, me.ID, "Ally A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Ally B", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, cap, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{cap, a, b} {
		if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("attacker %s has %d counters, want 1", id, got)
		}
		if !effectstest.EffectiveAbilitiesContain(t, g, id, "indestructible") {
			t.Errorf("attacker %s is not indestructible", id)
		}
	}
	if got := effectstest.CountersOn(g, idle, game.CounterPlusOne); got != 0 {
		t.Errorf("a creature that stayed home got %d counters", got)
	}

	// Two attackers are not enough.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	cap2 := effectstest.B12Push(g2, me2.ID, "Captain America, Skybound", "Legendary Creature — Human Soldier Hero", msc8SkyboundOracle, 3, 5)
	c2 := effectstest.B12Creature(g2, me2.ID, "Ally", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g2, g2.Seats[1].ID, cap2, c2)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.CountersOn(g2, cap2, game.CounterPlusOne); got != 0 {
		t.Errorf("two attackers triggered battalion: %d counters", got)
	}
}

func TestMsc8BigBerthaEntersWithXCountersAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.CastXSpell(t, g, "Big Bertha", "Legendary Creature — Mutant Hero", msc8BerthaOracle, "{X}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Big Bertha did not reach the battlefield")
	}
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 3 {
		t.Fatalf("entered with %d counters, want 3", got)
	}
	effectstest.AssertKeywords(t, g, id, "vigilance")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := effectstest.CrtActivateRaw(t, g, me.ID, id, 0, nil, "G", "G"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 4 {
		t.Errorf("after activating: %d counters, want 4", got)
	}
}

func TestMsc8AntManReformedRogueFollowsTheColoursYouCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ant := effectstest.B12Push(g, me.ID, "Ant-Man, Reformed Rogue", "Legendary Creature — Human Rogue Hero", msc8ReformedOracle, 2, 3)
	effectstest.ToMainForCost(t, g)
	green := effectstest.HandCardFull(me, "Green Sorcery", "Sorcery", "{G}", "", []string{"G"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, green, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast green: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, ant); got != 3 {
		t.Errorf("after a green spell power is %d, want 3", got)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, ant, "trample") {
		t.Error("no trample after a green spell")
	}

	blue := effectstest.HandCardFull(me, "Blue Sorcery", "Sorcery", "{U}", "", []string{"U"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, blue, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast blue: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, ant); got != 2 {
		t.Errorf("after a blue spell power is %d, want 2", got)
	}
	wall := effectstest.B12Creature(g, g.Seats[1].ID, "Wall", "Creature — Wall", 0, 8)
	effectstest.BrAttack(t, g, ant)
	effectstest.BrRefusal(t, g.DeclareBlocker(wall, ant), game.BlockReasonCantBeBlocked)

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	ant2 := effectstest.B12Push(g2, me2.ID, "Ant-Man, Reformed Rogue", "Legendary Creature — Human Rogue Hero", msc8ReformedOracle, 2, 3)
	effectstest.SeedLibrary(me2, "One", "Two")
	hand := me2.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g2, ant2, g2.Seats[1].ID, 2)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := me2.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}
