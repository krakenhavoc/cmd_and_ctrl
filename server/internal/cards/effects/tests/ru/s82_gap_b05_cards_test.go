package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b05_cards_test.go — S82 GAP batch 5 (Mystery Booster Commander, #5948).

const (
	gapB05GusthaOracle    = "a0a8c04e-3ce7-4945-a6af-7e40a990c82c"
	gapB05SashOracle      = "2ff470ee-4532-4ded-a684-a2c0d1253fb3"
	gapB05TolabowOracle   = "56d71eda-4037-45fb-9e98-1725147af6fc"
	gapB05EkthiOracle     = "e298a2a4-0e04-4af9-8638-7c01668d97d0"
	gapB05GrakkOracle     = "56277f60-67b0-41fa-a363-c64e2d794d01"
	gapB05NiveaOracle     = "2a5376aa-8996-476e-9ee8-1426b0944091"
	gapB05SelumaOracle    = "dbbf6379-01b4-400d-a277-5887e6be1d0f"
	gapB05YumeOracle      = "f6841a6f-7dd3-4976-9891-539189c0e148"
	gapB05WhtzOracle      = "73ba7ed0-0f08-4ba4-9f51-42c48f1b62ea"
	gapB05OvercookedOracl = "2c42983e-f45a-4d10-afe5-d2234dfae1ee"
	gapB05ThomilOracle    = "fd769cfb-668c-4557-a907-c6c17539c96a"
	gapB05ArzakonOracle   = "a95d29b9-c30b-4e63-afcb-dec7ed522c5a"
)

func gapB05Hand(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true}})
	return id
}

func gapB05Card(g *game.Game, id uuid.UUID) game.Card {
	var c game.Card
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if p := effectstest.FindBattlefieldCardForTest(g, id); p != nil {
			c = *p
		}
	})
	return c
}

// gapB05PT is a permanent's power and toughness after the layers.
func gapB05PT(g *game.Game, id uuid.UUID) (int, int) {
	c := gapB05Card(g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

func TestGapB05AllAreFull(t *testing.T) {
	for _, o := range []string{gapB05GusthaOracle, gapB05SashOracle, gapB05TolabowOracle, gapB05EkthiOracle, gapB05GrakkOracle,
		gapB05NiveaOracle, gapB05SelumaOracle, gapB05OvercookedOracl, gapB05ThomilOracle, gapB05ArzakonOracle, gapB05WhtzOracle, gapB05YumeOracle} {
		spec, ok := effects.Lookup(o)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, completeness %v, caveats %v", o, ok, spec.Completeness, spec.Caveats)
		}
	}
}

// Gustha: only the first spell cast during an opponent's turn draws three and
// puts two back, in the order the controller picks.
func TestGapB05GusthaBrainstormsOnTheFirstSpellOfAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gustha := effectstest.B12Push(g, me.ID, "Gustha Ebbasdotter", "Legendary Creature — Human Wizard", gapB05GusthaOracle, 3, 3)
	if c := gapB05Card(g, gustha); !game.HasKeyword(&c, game.ProtectionFromColor("R")) {
		t.Error("Gustha has protection from red")
	}
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("a spell on your own turn triggered Gustha")
	}

	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	hand := me.Hand.Size()
	effectstest.B22CastInstantAs(t, g, me, "Brainstorm")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Fatalf("drew %d, want 3", got)
	}
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("put-back prompt %+v, want exactly two cards", pick)
	}
	first, second := me.Hand.Cards[0].InstanceID, me.Hand.Cards[1].InstanceID
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{first, second}); err != nil {
		t.Fatal(err)
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil {
		t.Fatal("no order prompt")
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, nil, []uuid.UUID{second, first}); err != nil {
		t.Fatal(err)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("net cards %d, want 1 (three drawn, two put back)", got)
	}
	effectstest.B22CastInstantAs(t, g, me, "Opt")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("the second spell that turn triggered Gustha again")
	}
}

// Ekthi: other Phyrexians get +1/+1 and every Equipment you control has living
// weapon, including one that enters after Ekthi and one an opponent controls
// not at all.
func TestGapB05EkthiPumpsPhyrexiansAndGivesEquipmentLivingWeapon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ekthi := effectstest.B12Push(g, me.ID, "Ekthi, Contaminator Priest", "Legendary Creature — Phyrexian Cleric", gapB05EkthiOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Phyrexian Ally", "Creature — Phyrexian", 2, 2)
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Phyrexian", "Creature — Phyrexian", 2, 2)
	if p, th := gapB05PT(g, ekthi); p != 3 || th != 3 {
		t.Errorf("Ekthi is %d/%d, want 3/3: it pumps only the others", p, th)
	}
	if p, th := gapB05PT(g, mine); p != 3 || th != 3 {
		t.Errorf("your Phyrexian is %d/%d, want 3/3", p, th)
	}
	if p, th := gapB05PT(g, human); p != 2 || th != 2 {
		t.Errorf("a Human is %d/%d, want 2/2", p, th)
	}
	if p, th := gapB05PT(g, theirs); p != 2 || th != 2 {
		t.Errorf("an opponent's Phyrexian is %d/%d, want 2/2", p, th)
	}

	sword := effectstest.CastCatalogSpell(t, g, "Plain Sword", "Artifact — Equipment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	germ := effectstest.FindBattlefieldByName(g, "Phyrexian Germ")
	if germ == uuid.Nil {
		t.Fatal("no Germ: the Equipment had no living weapon")
	}
	if p, th := gapB05PT(g, germ); p != 1 || th != 1 {
		t.Errorf("the Germ is %d/%d, want 1/1 (a Phyrexian under Ekthi)", p, th)
	}
	if sc, _ := effectstest.BattlefieldCard(g, sword); sc.AttachedTo.ID != germ {
		t.Error("the Equipment is not attached to the Germ")
	}
	// An opponent's Equipment gets none.
	g.WithWriteLock(func() { g.Turn.Seq++ })
	pushed := effectstest.B12Permanent(g, opp.ID, "Their Sword", "Artifact — Equipment")
	_ = pushed
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Phyrexian Germ"); n != 1 {
		t.Errorf("%d Germs, want 1", n)
	}
}

// Tolabow: each instant or sorcery cast makes a 1/1 blue and red Otter with
// prowess; a creature spell makes none.
func TestGapB05TolabowMakesAnOtterPerInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Tolabow, Loch Rascal", "Legendary Creature — Otter", gapB05TolabowOracle, 3, 2)
	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Otter"); n != 2 {
		t.Fatalf("%d Otters, want 2", n)
	}
	otter := effectstest.FindBattlefieldByName(g, "Otter")
	c := gapB05Card(g, otter)
	if c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 2 || !game.HasKeyword(&c, "prowess") {
		t.Errorf("Otter is %+v, want a 1/1 blue and red token with prowess", c)
	}
}

// Grakk can't attack or block, and draws once each turn for an Aura cast on an
// opponent's creature.
func TestGapB05GrakkCantAttackOrBlockAndDrawsOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grakk := effectstest.B12Push(g, me.ID, "Grakk the Pacifist", "Legendary Creature — Orc", gapB05GrakkOracle, 4, 3)
	effectstest.AssertRestrictions(t, g, grakk, game.CantAttackOrBlock)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.AdvanceToMain(t, g)

	hand := me.Hand.Size()
	aura := effectstest.CastCatalogSpell(t, g, "Pacifism", "Enchantment — Aura", "", []game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	_ = aura
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("an Aura on your own creature drew (%d → %d)", hand, got)
	}
	effectstest.CastCatalogSpell(t, g, "Pacifism", "Enchantment — Aura", "", []game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("an Aura on their creature: hand %d → %d, want +1", hand, got)
	}
	effectstest.CastCatalogSpell(t, g, "Pacifism", "Enchantment — Aura", "", []game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("the second Aura that turn drew again (%d)", got)
	}
}

// Seluma returns an Angel creature card from your graveyard when it connects,
// and only an Angel.
func TestGapB05SelumaReturnsAnAngelOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seluma := effectstest.B12Push(g, me.ID, "Seluma, Light of Aysen", "Legendary Creature — Angel Warrior", gapB05SelumaOracle, 4, 4)
	angel := effectstest.B17GraveyardCard(me, "Dead Angel", "Creature — Angel", "{3}{W}")
	bear := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	theirAngel := effectstest.B17GraveyardCard(opp, "Their Angel", "Creature — Angel", "{3}{W}")
	effectstest.DealCombatDamageToPlayer(g, seluma, opp.ID, 4)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, angel) {
		t.Error("your Angel is not offered")
	}
	if effectstest.HasID(p.PickTargetCards, bear) || effectstest.HasID(p.PickTargetCards, theirAngel) {
		t.Error("a non-Angel or an opponent's Angel is offered")
	}
	effectstest.PickCard(t, g, me.ID, angel)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(angel) {
		t.Error("the Angel did not return to the battlefield")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the Bear left the graveyard")
	}
}

// Overcooked: nobody gains life; the end step makes a Food, or with two nonland
// permanents entered this turn conjures a Food Fight onto the battlefield.
func TestGapB05OvercookedStopsLifeGainAndMakesFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Overcooked", "Enchantment", gapB05OvercookedOracl, 0, 0)
	g.WithWriteLock(func() {
		for _, p := range []*game.Player{me, opp} {
			before := p.Life
			_ = g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, 3)
			if p.Life != before {
				t.Errorf("%s gained life under Overcooked", p.Name)
			}
		}
	})
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Food"); n != 1 {
		t.Fatalf("%d Food tokens, want 1", n)
	}
	if n := effectstest.B16CountNamed(g, "Food Fight"); n != 0 {
		t.Errorf("%d Food Fights without celebration", n)
	}
}

func TestGapB05OvercookedConjuresFoodFightWithCelebration(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Overcooked", "Enchantment", gapB05OvercookedOracl, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Rock A", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Rock B", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.Celebration(g, me.ID) {
		t.Fatal("two nonland permanents entered, expected celebration")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Food"); n != 0 {
		t.Errorf("%d Food tokens with celebration, want none", n)
	}
	fight := effectstest.FindBattlefieldByName(g, "Food Fight")
	if fight == uuid.Nil {
		t.Fatal("no Food Fight conjured")
	}
	if c, _ := effectstest.BattlefieldCard(g, fight); c.Controller != me.ID || c.Owner != me.ID || c.IsToken() {
		t.Errorf("Food Fight is %+v, want a nontoken card you own and control", c)
	}
	if spec, ok := effects.Lookup("00571e7f-4e40-4a07-a529-34580e942338"); !ok || len(spec.Grants) == 0 {
		t.Error("the conjured card is not the catalog's Food Fight")
	}
}

// Nivea untaps a creature of yours that becomes blocked and removes it from
// combat (the "may" is answered yes), and does nothing for another player's.
func TestGapB05NiveaUntapsAndRemovesABlockedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Nivea, Beloved Battlemage", "Legendary Creature — Human Wizard Warrior", gapB05NiveaOracle, 3, 1)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Giant", 5, 5)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 9)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	effectstest.LockInBlocks(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, attacker)
	if c.Tapped {
		t.Error("the blocked creature is still tapped")
	}
	if c.AttackingTarget != uuid.Nil {
		t.Error("the blocked creature is still attacking")
	}
}

func TestGapB05NiveaTurnsIntoAkroma(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	nivea := effectstest.B12Push(g, me.ID, "Nivea, Beloved Battlemage", "Legendary Creature — Human Wizard Warrior", gapB05NiveaOracle, 3, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(me, "W", "W", "W", "W", "W", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, nivea, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with seven mana")
	}
	effectstest.ApaMana(me, "W")
	effectstest.B16Activate(t, g, me.ID, nivea, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(nivea) {
		t.Error("Nivea was not sacrificed")
	}
	akroma := effectstest.FindBattlefieldByName(g, "Akroma, Angel of Wrath")
	if akroma == uuid.Nil {
		t.Fatal("no Akroma token")
	}
	c := gapB05Card(g, akroma)
	if !c.IsToken() || c.Power != 6 || c.Toughness != 6 || !game.HasKeyword(&c, "flying") ||
		!game.HasKeyword(&c, game.ProtectionFromColor("B")) || !game.HasKeyword(&c, game.ProtectionFromColor("R")) {
		t.Errorf("Akroma token is %+v", c)
	}
}

// Thomil: +2 makes a Zombie, 0 adds BBB only for a creature really sacrificed,
// -5 makes a Lord of the Pit whose upkeep sacrifices another creature or hurts.
func TestGapB05ThomilPlusTwoAndZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thomil := effectstest.PushWalkerForTest(g, me.ID, "Thomil, the Destroyer", gapB05ThomilOracle, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, thomil, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	zombie := effectstest.FindBattlefieldByName(g, "Zombie")
	if zombie == uuid.Nil {
		t.Fatal("no Zombie token")
	}
	if c, _ := effectstest.BattlefieldCard(g, thomil); c.Counters[game.CounterLoyalty] != 6 {
		t.Errorf("loyalty %d, want 6", c.Counters[game.CounterLoyalty])
	}
	// The 0: decline first (the pick is empty), then sacrifice the Zombie. One
	// loyalty ability per planeswalker each turn, so each try has its own.
	first := effectstest.PushWalkerForTest(g, me.ID, "Thomil Stand-in A", gapB05ThomilOracle, 4)
	effectstest.B16Activate(t, g, me.ID, first, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 || !g.Battlefield.Contains(zombie) {
		t.Fatalf("declining sacrificed or added mana (pool %d)", me.ManaPool.Total())
	}
	second := effectstest.PushWalkerForTest(g, me.ID, "Thomil Stand-in B", gapB05ThomilOracle, 4)
	effectstest.B16Activate(t, g, me.ID, second, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, zombie)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(zombie) {
		t.Error("the Zombie was not sacrificed")
	}
	got := effectstest.PoolColors(me)
	if len(got) != 3 || got[0] != "B" || got[1] != "B" || got[2] != "B" {
		t.Errorf("pool %v, want {B}{B}{B}", got)
	}
}

func TestGapB05ThomilMinusFiveMakesALordOfThePit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thomil := effectstest.PushWalkerForTest(g, me.ID, "Thomil, the Destroyer", gapB05ThomilOracle, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, thomil, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(thomil) {
		t.Error("Thomil at 0 loyalty stays")
	}
	lord := effectstest.FindBattlefieldByName(g, "Lord of the Pit")
	if lord == uuid.Nil {
		t.Fatal("no Lord of the Pit")
	}
	c := gapB05Card(g, lord)
	if c.Power != 7 || c.Toughness != 7 || !game.HasKeyword(&c, "flying") || !game.HasKeyword(&c, "trample") || !c.IsToken() {
		t.Errorf("Lord of the Pit is %+v", c)
	}
	// Alone, it deals 7 damage to its controller at their upkeep.
	life := me.Life
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-7 {
		t.Errorf("life %d → %d, want -7", life, me.Life)
	}
	if !g.Battlefield.Contains(lord) {
		t.Error("the Lord sacrificed itself")
	}
	// With another creature, it sacrifices that creature instead.
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	answerLordSacrifice(t, g, me.ID, bear)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not sacrificed")
	}
	if !g.Battlefield.Contains(lord) {
		t.Error("the Lord was sacrificed")
	}
}

func answerLordSacrifice(t *testing.T, g *game.Game, who, pick uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		if c := effectstest.LatestChooseCardsFor(g, who); c != nil {
			if err := g.ResolveChooseCards(c.ID, who, []uuid.UUID{pick}); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			return
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	t.Fatal("no sacrifice prompt")
}

// Arzakon: +2 hits any other target for 3 (never itself), -3 makes a Black Lotus
// that taps for three mana of one colour, -7 is a wheel.
func TestGapB05ArzakonPlusTwoDealsThreeToAnotherTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arz := effectstest.PushWalkerForTest(g, me.ID, "Arzakon", gapB05ArzakonOracle, 4)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, arz, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: arz}}}); err == nil {
		t.Fatal("Arzakon targeted himself")
	}
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, arz, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d → %d, want -3", life, opp.Life)
	}
}

func TestGapB05ArzakonMinusThreeMakesABlackLotus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arz := effectstest.PushWalkerForTest(g, me.ID, "Arzakon", gapB05ArzakonOracle, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, arz, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	lotus := effectstest.FindBattlefieldByName(g, "Black Lotus")
	if lotus == uuid.Nil {
		t.Fatal("no Black Lotus token")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, lotus).SummonedThisTurn = false })
	if err := g.ActivateManaAbility(me.ID, lotus, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Black Lotus mana ability: %v", err)
	}
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || pick.ManaAmounts["W"] != 3 {
		t.Fatalf("pick = %+v, want one pick of three", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.PoolColors(me); len(got) != 3 || got[0] != "R" {
		t.Errorf("pool %v, want three {R}", got)
	}
	if effectstest.OnBattlefieldNow(g, lotus) {
		t.Error("the Lotus is sacrificed for its mana")
	}
}

func TestGapB05ArzakonMinusSevenShufflesAndDrawsSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arz := effectstest.PushWalkerForTest(g, me.ID, "Arzakon", gapB05ArzakonOracle, 7)
	effectstest.AdvanceToMain(t, g)
	gy := effectstest.B17GraveyardCard(me, "Old Card", "Instant", "{U}")
	oppGy := effectstest.B17GraveyardCard(opp, "Their Old Card", "Instant", "{U}")
	effectstest.B16Activate(t, g, me.ID, arz, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range []*game.Player{me, opp} {
		if p.Hand.Size() != 7 {
			t.Errorf("%s has %d cards in hand, want 7", p.Name, p.Hand.Size())
		}
	}
	if me.Graveyard.Contains(gy) || opp.Graveyard.Contains(oppGy) {
		t.Error("a graveyard was not shuffled away")
	}
}

// Sash and Waistcoat: a face-down creature attacking an opponent of Sash's
// controller can't be blocked; any player may manifest, as a sorcery.
func TestGapB05SashManifestIsAnyPlayerAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sash := effectstest.B12Push(g, me.ID, "Sash and Waistcoat, Unmen", "Legendary Creature — Insect Rogue", gapB05SashOracle, 2, 2)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.SeedLibrary(opp, "X", "Y", "Z")
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(opp, "C", "C", "C")
	if err := g.ActivateCatalogAbility(opp.ID, sash, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the opponent manifested on my turn")
	}
	if n := countFaceDown(g); n != 0 {
		t.Fatalf("%d face-down creatures before any manifest", n)
	}
	effectstest.ApaMana(me, "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, sash, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if n := countFaceDown(g); n != 1 {
		t.Fatalf("%d face-down creatures, want 1", n)
	}
	// On their own turn the opponent may, from their own library, under their control.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.ApaMana(opp, "C", "C", "C")
	effectstest.B16Activate(t, g, opp.ID, sash, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if n := countFaceDown(g); n != 2 {
		t.Fatalf("%d face-down creatures, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.FaceDown && c.Controller != c.Owner {
			t.Error("a manifest came from someone else's library")
		}
	}
}

func countFaceDown(g *game.Game) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.FaceDown {
			n++
		}
	}
	return n
}

func TestGapB05SashMakesAFaceDownAttackerUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sash and Waistcoat, Unmen", "Legendary Creature — Insect Rogue", gapB05SashOracle, 2, 2)
	faceUp := effectstest.B12Creature(g, me.ID, "Face Up", "Creature — Bear", 2, 2)
	var down uuid.UUID
	g.WithWriteLock(func() {
		id, err := g.ManifestForEffect(me.ID)
		if err != nil {
			t.Fatalf("manifest: %v", err)
		}
		down = id
		effectstest.FindBattlefieldCardForTest(g, id).SummonedThisTurn = false
	})
	if down == uuid.Nil {
		effectstest.SeedLibrary(me, "A")
		g.WithWriteLock(func() {
			down, _ = g.ManifestForEffect(me.ID)
			effectstest.FindBattlefieldCardForTest(g, down).SummonedThisTurn = false
		})
	}
	effectstest.DeclareAttack(t, g, opp.ID, faceUp, down)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, down); !game.Restricted(&c, game.CantBeBlocked) {
		t.Error("the face-down attacker can be blocked")
	}
	if c := effectstest.LayeredCard(t, g, faceUp); game.Restricted(&c, game.CantBeBlocked) {
		t.Error("the face-up attacker can't be blocked")
	}
}

// Whtz draws and gains life for {3} and {T}; with a 200-card starting deck the
// {3} is gone.
func TestGapB05WhtzDrawsAndGainsLifeAndCostsLessWithABigDeck(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	whtz := effectstest.B12Push(g, me.ID, "Whtz, the Bibliophile", "Legendary Creature — Homunculus", gapB05WhtzOracle, 1, 3)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, whtz, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana in a normal-sized deck")
	}
	effectstest.ApaMana(me, "C", "C", "C")
	hand, life := me.Hand.Size(), me.Life
	effectstest.B16Activate(t, g, me.ID, whtz, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life+1 {
		t.Errorf("hand %d → %d, life %d → %d, want +1 each", hand, me.Hand.Size(), life, me.Life)
	}
	if !effectstest.B16Tapped(t, g, whtz) {
		t.Error("Whtz did not tap")
	}
	// A 200-card starting deck: only the tap.
	g.WithWriteLock(func() {
		me.StartingDeckSize = 200
		effectstest.FindBattlefieldCardForTest(g, whtz).Tapped = false
	})
	hand, life = me.Hand.Size(), me.Life
	effectstest.B16Activate(t, g, me.ID, whtz, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life+1 {
		t.Errorf("free activation: hand %d → %d, life %d → %d, want +1 each", hand, me.Hand.Size(), life, me.Life)
	}
}

// Yume gives the first spell of each of your turns assist; when another player
// really pays, both draw. The second spell has no assist.
func TestGapB05YumeGivesTheFirstSpellAssistAndBothDrawWhenSomeonePays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, bob := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Yume, Chronicler of Valor", "Legendary Creature — Human Druid", gapB05YumeOracle, 1, 2)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.SeedLibrary(bob, "X", "Y")
	first := gapB05Hand(me, "Big Sorcery", "Sorcery", "{3}{R}")
	second := gapB05Hand(me, "Other Sorcery", "Sorcery", "{3}{R}")
	effectstest.ApaMana(me, "R")
	effectstest.ApaMana(bob, "C", "C", "C")
	myHand, bobHand := me.Hand.Size(), bob.Hand.Size()
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, AssistPlayer: bob.ID}); err != nil {
		t.Fatalf("cast with assist: %v", err)
	}
	c := latestPendingAssist(g)
	if c == nil || c.Chooser != bob.ID {
		t.Fatalf("the cast did not wait for Bob: %+v", g.PendingChoices)
	}
	if err := g.ResolvePayAmount(c.ID, bob.ID, 3); err != nil {
		t.Fatalf("Bob helps: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != myHand-1+1 { // the spell left, one card drawn
		t.Errorf("caster's hand %d → %d, want the spell out and one card in", myHand, me.Hand.Size())
	}
	if bob.Hand.Size() != bobHand+1 {
		t.Errorf("helper's hand %d → %d, want +1", bobHand, bob.Hand.Size())
	}
	// The second spell has no assist.
	effectstest.ApaMana(me, "R", "R", "R", "R")
	effectstest.ApaMana(bob, "C", "C", "C")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true, AssistPlayer: bob.ID}); err == nil {
		t.Fatal("the second spell of the turn took an assist")
	}
}

// Without anyone paying (no assist asked), nobody draws, and a spell on another
// player's turn has no assist.
func TestGapB05YumeDrawsNothingWithoutAHelperAndGivesNoAssistOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, bob := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Yume, Chronicler of Valor", "Legendary Creature — Human Druid", gapB05YumeOracle, 1, 2)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.SeedLibrary(bob, "X", "Y")
	spell := gapB05Hand(me, "Plain Instant", "Instant", "{R}")
	effectstest.ApaMana(me, "R")
	myHand, bobHand := me.Hand.Size(), bob.Hand.Size()
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != myHand-1 || bob.Hand.Size() != bobHand {
		t.Errorf("hands: caster %d → %d, helper %d → %d; nobody should draw", myHand, me.Hand.Size(), bobHand, bob.Hand.Size())
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	other := gapB05Hand(me, "Instant On Their Turn", "Instant", "{1}{R}")
	effectstest.ApaMana(me, "R", "R")
	effectstest.ApaMana(bob, "C")
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{Strict: true, AssistPlayer: bob.ID}); err == nil {
		t.Error("a spell cast on an opponent's turn took an assist")
	}
}

func latestPendingAssist(g *game.Game) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Assist != nil {
			return c
		}
	}
	return nil
}
