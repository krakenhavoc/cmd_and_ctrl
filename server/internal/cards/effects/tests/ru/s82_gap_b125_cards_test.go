package testsru

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b125_cards_test.go — set GAP, batch 125 (#5948). Each test plays
// the real catalog entry through the engine and asserts what its text says.

const (
	gapB125MeriadocOracle  = "fec6a4a7-4ded-4062-b99b-c1826dea7113"
	gapB125ChannelerOracle = "30f373d8-cd8c-4fb2-9383-9969fc547df9"
	gapB125SpiderOracle    = "07fb2903-b239-43d0-864b-c4437b7a5230"
	gapB125WatchdogsOracle = "82259c84-773d-4670-9b09-d73f5a3be7ea"
	gapB125QuickbeamOracle = "9ad903c5-1528-44d6-b4b6-f05a2399d91f"
	gapB125RadagastOracle  = "dfa47d0d-17f0-4a12-89d3-0d3c7f3db8c6"
	gapB125WoseOracle      = "eb2ff7f2-b7fd-4984-a4e8-cf57b619c1a5"
	gapB125GimliOracle     = "9f3941d8-b4d2-4faf-881a-f15973b4368d"
	gapB125StriderOracle   = "118f2cb6-2af9-4acc-b386-a8e8dded5793"
	gapB125ArwenOracle     = "5ffbbce1-fe38-40dc-a7c3-2055ae86c6c5"
	gapB125ElrondOracle    = "28aefa3a-16d3-4207-bb8c-588b722a4041"
	gapB125LegolasOracle   = "ddc12487-f577-4def-acdb-491f7ba6dc51"
)

func TestGapB125EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB125MeriadocOracle:  "Meriadoc Brandybuck",
		gapB125ChannelerOracle: "Mirkwood Channeler",
		gapB125SpiderOracle:    "Mirkwood Spider",
		gapB125WatchdogsOracle: "Mushroom Watchdogs",
		gapB125QuickbeamOracle: "Quickbeam, Upstart Ent",
		gapB125RadagastOracle:  "Radagast the Brown",
		gapB125WoseOracle:      "Wose Pathfinder",
		gapB125GimliOracle:     "Gimli, Mournful Avenger",
		gapB125StriderOracle:   "Strider, Ranger of the North",
		gapB125ArwenOracle:     "Arwen Undómiel",
		gapB125ElrondOracle:    "Elrond, Master of Healing",
		gapB125LegolasOracle:   "Legolas, Counter of Kills",
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

// gapB125Main is a fresh table at the active seat's main phase, with that
// seat and the next.
func gapB125Main(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// gapB125SettleWithPick passes priority until the stack is empty, answering
// `chooser`'s target picks with `ref` and saying `may` to their "you may"
// prompts on the way. Any other prompt stops it, for the test to answer.
func gapB125SettleWithPick(t *testing.T, g *game.Game, chooser uuid.UUID, ref game.TargetRef, may bool) {
	t.Helper()
	for i := 0; i < 48; i++ {
		if p := effectstest.LatestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, ref); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if c := effectstest.LatestTriggerPrompt(g, chooser); c != nil {
			if err := g.ResolveTriggerPrompt(c.ID, chooser, may); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			continue
		}
		if len(g.PendingTriggers) > 0 {
			g.RunStateChecksForTest()
			if effectstest.LatestPickTarget(g, chooser) != nil || effectstest.LatestTriggerPrompt(g, chooser) != nil {
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

// gapB125ChoosableCards is the card list of the open choose-cards prompt.
func gapB125ChoosableCards(g *game.Game, who uuid.UUID) []uuid.UUID {
	if c := effectstest.ChooseCardsChoiceFor(g, who); c != nil {
		return c.ChooseCards
	}
	return nil
}

func gapB125Foods(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.HasSubtype("Food") {
			n++
		}
	}
	return n
}

func gapB125FoodIDs(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.HasSubtype("Food") {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func gapB125Scry(t *testing.T, g *game.Game, me *game.Player, n int) {
	t.Helper()
	g.WithWriteLock(func() { g.ScryForEffect(me.ID, uuid.Nil, n) })
	effectstest.AnswerScryKeepAll(t, g, me.ID)
}

func gapB125Ref(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func gapB125Counters(g *game.Game, id uuid.UUID) int {
	return effectstest.CounterCount(g, id, game.CounterPlusOne)
}

func gapB125Library(p *game.Player, cards ...game.Card) {
	for i := len(cards) - 1; i >= 0; i-- {
		cards[i].InstanceID = uuid.New()
		cards[i].Owner, cards[i].Controller = p.ID, p.ID
		p.Library.PushTop(cards[i])
	}
}

// --- Meriadoc Brandybuck ---------------------------------------------------

func TestGapB125MeriadocMakesOneFoodForHalflingsAttackingAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mer := effectstest.B12Push(g, me.ID, "Meriadoc Brandybuck", "Legendary Creature — Halfling Citizen", gapB125MeriadocOracle, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Hobbit", "Creature — Halfling", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, mer, other)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB125Foods(g, me.ID); got != 1 {
		t.Errorf("Foods = %d, want exactly 1 for two Halflings attacking one player", got)
	}
}

func TestGapB125MeriadocIgnoresNonHalflingAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Meriadoc Brandybuck", "Legendary Creature — Halfling Citizen", gapB125MeriadocOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB125Foods(g, me.ID); got != 0 {
		t.Errorf("Foods = %d, want 0 when no Halfling attacks", got)
	}
}

// --- Mirkwood Channeler ----------------------------------------------------

func TestGapB125ChannelerPumpsAnElfByItsForests(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Mirkwood Channeler", "Creature — Elf Druid", gapB125ChannelerOracle, 3, 3)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf Warrior", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	gapB125SettleWithPick(t, g, me.ID, gapB125Ref(elf), false)
	if p, tough := effectstest.EffectivePTOf(t, g, elf); p != 4 || tough != 4 {
		t.Errorf("the Elf is %d/%d, want 4/4 (three Forests)", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, elf, "trample") {
		t.Error("the Elf lacks trample")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("a non-Elf was affected")
	}
}

func TestGapB125ChannelerCannotTargetANonElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	channeler := effectstest.B12Push(g, me.ID, "Mirkwood Channeler", "Creature — Elf Druid", gapB125ChannelerOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, gapB125Ref(bear)); err == nil {
		t.Error("a Bear was accepted as the target of \"target Elf you control\"")
	}
	gapB125SettleWithPick(t, g, me.ID, gapB125Ref(channeler), false)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Errorf("the Bear is %d power, want 2: it is not an Elf", p)
	}
	// With no Forests X is 0, but the Channeler itself is the only Elf and still gains trample.
	if !effectstest.HasEffectiveKeyword(t, g, channeler, "trample") {
		t.Error("the lone Elf (the Channeler) did not gain trample")
	}
}

// --- Mirkwood Spider -------------------------------------------------------

func TestGapB125SpiderGivesAnAttackingLegendDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spider := effectstest.B12Push(g, me.ID, "Mirkwood Spider", "Creature — Spider", gapB125SpiderOracle, 1, 1)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, spider, "deathtouch") {
		t.Fatal("the Spider lacks deathtouch")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(spider, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	gapB125SettleWithPick(t, g, me.ID, gapB125Ref(legend), false)
	if !effectstest.HasEffectiveKeyword(t, g, legend, "deathtouch") {
		t.Error("the legendary creature did not gain deathtouch")
	}
	if effectstest.HasEffectiveKeyword(t, g, plain, "deathtouch") {
		t.Error("a nonlegendary creature gained deathtouch")
	}
}

// --- Mushroom Watchdogs ----------------------------------------------------

func TestGapB125WatchdogsSacrificeAFoodForACounterAndVigilance(t *testing.T) {
	g, me, _ := gapB125Main(t)
	dog := effectstest.B12Push(g, me.ID, "Mushroom Watchdogs", "Creature — Dog", gapB125WatchdogsOracle, 2, 2)
	if err := effectstest.ConB1Activate(t, g, me, dog, 0, "", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with no Food to sacrifice")
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.FoodToken(), 1); err != nil {
			t.Fatal(err)
		}
	})
	if err := effectstest.ConB1Activate(t, g, me, dog, 0, "", game.ActivateAbilityParams{SacrificeIDs: gapB125FoodIDs(g, me.ID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB125Counters(g, dog); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, dog, "vigilance") {
		t.Error("the Watchdogs lack vigilance")
	}
	if gapB125Foods(g, me.ID) != 0 {
		t.Error("the Food was not sacrificed")
	}
}

func TestGapB125WatchdogsActivateOnlyAsASorcery(t *testing.T) {
	g, me, opp := gapB125Main(t)
	dog := effectstest.B12Push(g, me.ID, "Mushroom Watchdogs", "Creature — Dog", gapB125WatchdogsOracle, 2, 2)
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.FoodToken(), 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	_ = opp
	if err := effectstest.ConB1Activate(t, g, me, dog, 0, "", game.ActivateAbilityParams{SacrificeIDs: gapB125FoodIDs(g, me.ID)}); err == nil {
		t.Error("activated outside a main phase")
	}
	if gapB125Foods(g, me.ID) != 1 {
		t.Error("the Food was sacrificed for a refused activation")
	}
}

// --- Quickbeam, Upstart Ent --------------------------------------------------

func TestGapB125QuickbeamPumpsTwoCreaturesWhenATreefolkEnters(t *testing.T) {
	g, me, _ := gapB125Main(t)
	effectstest.B12Push(g, me.ID, "Quickbeam, Upstart Ent", "Legendary Creature — Treefolk", gapB125QuickbeamOracle, 5, 6)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	bb := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Sapling", "Creature — Treefolk", "", 1, 1)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(a, bb)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, bb} {
		if pw, tough := effectstest.EffectivePTOf(t, g, id); pw != 4 || tough != 4 {
			t.Errorf("a chosen creature is %d/%d, want 4/4", pw, tough)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
			t.Error("a chosen creature lacks trample")
		}
	}
	if pw, _ := effectstest.EffectivePTOf(t, g, c); pw != 2 {
		t.Errorf("the unchosen creature is %d power, want 2", pw)
	}
}

func TestGapB125QuickbeamIgnoresAnEnteringNonTreefolk(t *testing.T) {
	g, me, _ := gapB125Main(t)
	effectstest.B12Push(g, me.ID, "Quickbeam, Upstart Ent", "Legendary Creature — Treefolk", gapB125QuickbeamOracle, 5, 6)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Bear Cub", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, a); p != 2 {
		t.Errorf("A is %d power, want 2: a Bear entering does not trigger Quickbeam", p)
	}
}

// --- Radagast the Brown ------------------------------------------------------

func gapB125CastCreature(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, mana string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, mana)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func TestGapB125RadagastLooksAtXAndTakesACreatureSharingNoType(t *testing.T) {
	g, me, _ := gapB125Main(t)
	effectstest.B12Push(g, me.ID, "Radagast the Brown", "Legendary Creature — Avatar Wizard", gapB125RadagastOracle, 2, 5)
	effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	// The top three cards: an Elf (shares a type), a Goblin (does not), a land.
	var elfCard, goblin uuid.UUID
	gapB125Library(me,
		game.Card{Name: "Elf Card", TypeLine: "Creature — Elf"},
		game.Card{Name: "Goblin Card", TypeLine: "Creature — Goblin"},
		game.Card{Name: "Deep Card", TypeLine: "Creature — Goblin"},
	)
	elfCard, goblin = me.Library.Cards[len(me.Library.Cards)-1].InstanceID, me.Library.Cards[len(me.Library.Cards)-2].InstanceID
	deep := me.Library.Cards[len(me.Library.Cards)-3].InstanceID
	gapB125CastCreature(t, g, me, "Mana Value Two", "Creature — Bear", "{1}{G}", "CG")
	effectstest.PassPriorityAroundTable(t, g)
	// X = 2: the Elf and the Goblin are looked at; the third card is not.
	picks := gapB125ChoosableCards(g, me.ID)
	for _, id := range picks {
		if id == elfCard {
			t.Error("a creature sharing a type with one of mine was offered")
		}
		if id == deep {
			t.Error("a card below the top X was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, goblin)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(goblin) {
		t.Error("the Goblin was not put into hand")
	}
	if me.Hand.Contains(elfCard) {
		t.Error("the Elf was put into hand")
	}
	bottom := me.Library.Cards[0].InstanceID
	if bottom != elfCard && me.Library.Cards[1].InstanceID != elfCard {
		t.Error("the rest did not go to the bottom of the library")
	}
}

func TestGapB125RadagastDoesNotTriggerForATokenOrAnotherPlayersCreature(t *testing.T) {
	g, me, opp := gapB125Main(t)
	effectstest.B12Push(g, me.ID, "Radagast the Brown", "Legendary Creature — Avatar Wizard", gapB125RadagastOracle, 2, 5)
	gapB125Library(me, game.Card{Name: "Top", TypeLine: "Creature — Goblin"})
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Human"), 1); err != nil {
			t.Fatal(err)
		}
		if err := g.CreateTokenForEffect(opp.ID, effects.TokenCard("1/1 white Human"), 1); err != nil {
			t.Fatal(err)
		}
	})
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil || me.Hand.Size() != hand {
		t.Error("Radagast triggered for a token or an opponent's creature")
	}
}

// --- Wose Pathfinder ---------------------------------------------------------

func TestGapB125WosePathfinderTapsForManaAndPumpsAnotherCreature(t *testing.T) {
	g, me, _ := gapB125Main(t)
	wose := effectstest.B12Push(g, me.ID, "Wose Pathfinder", "Creature — Human Shaman", gapB125WoseOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.ConB1Activate(t, g, me, wose, 0, "GCCCCCC", game.ActivateAbilityParams{Targets: effectstest.CardTarget(wose)}); err == nil {
		t.Error("the Pathfinder targeted itself")
	}
	if err := effectstest.ConB1Activate(t, g, me, wose, 0, "GCCCCCC", game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear lacks trample")
	}
	if !effectstest.HmlTapped(t, g, wose) {
		t.Error("the Pathfinder did not tap")
	}
}

func TestGapB125WosePathfinderHasAnAnyColorManaAbility(t *testing.T) {
	spec, ok := effects.Lookup(gapB125WoseOracle)
	if !ok || len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W|U|B|R|G}" || !spec.ManaAbilities[0].Cost.Tap {
		t.Fatalf("mana abilities = %+v", spec.ManaAbilities)
	}
}

// --- Gimli, Mournful Avenger -------------------------------------------------

func gapB125Die(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
}

func TestGapB125GimliGrowsAndFightsOnTheThirdDeath(t *testing.T) {
	g, me, opp := gapB125Main(t)
	gimli := effectstest.B12Push(g, me.ID, "Gimli, Mournful Avenger", "Legendary Creature — Dwarf Warrior", gapB125GimliOracle, 3, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 1)
	var fodder []uuid.UUID
	for i := 0; i < 3; i++ {
		fodder = append(fodder, effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1))
	}
	for i, id := range fodder {
		gapB125Die(t, g, id)
		g.RunStateChecksForTest()
		gapB125SettleWithPick(t, g, me.ID, gapB125Ref(victim), false)
		if got := gapB125Counters(g, gimli); got != i+1 {
			t.Fatalf("after death %d Gimli has %d counters, want %d", i+1, got, i+1)
		}
		if i < 2 && !g.Battlefield.Contains(victim) {
			t.Fatalf("Gimli fought on resolution %d", i+1)
		}
	}
	if g.Battlefield.Contains(victim) {
		t.Error("Gimli did not fight on the third resolution")
	}
}

func TestGapB125GimliIsIndestructibleAfterTwoOfYourCreaturesDied(t *testing.T) {
	g, me, _ := gapB125Main(t)
	gimli := effectstest.B12Push(g, me.ID, "Gimli, Mournful Avenger", "Legendary Creature — Dwarf Warrior", gapB125GimliOracle, 3, 2)
	if effectstest.HasEffectiveKeyword(t, g, gimli, "indestructible") {
		t.Fatal("Gimli is indestructible before any creature died")
	}
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	gapB125Die(t, g, a)
	g.RunStateChecksForTest()
	gapB125SettleWithPick(t, g, me.ID, game.TargetRef{}, false)
	if effectstest.HasEffectiveKeyword(t, g, gimli, "indestructible") {
		t.Error("Gimli is indestructible after one death")
	}
	bb := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	gapB125Die(t, g, bb)
	g.RunStateChecksForTest()
	gapB125SettleWithPick(t, g, me.ID, game.TargetRef{}, false)
	if !effectstest.HasEffectiveKeyword(t, g, gimli, "indestructible") {
		t.Error("Gimli is not indestructible after two deaths")
	}
}

func TestGapB125GimliIgnoresAnOpponentsCreatureDying(t *testing.T) {
	g, me, opp := gapB125Main(t)
	gimli := effectstest.B12Push(g, me.ID, "Gimli, Mournful Avenger", "Legendary Creature — Dwarf Warrior", gapB125GimliOracle, 3, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	gapB125Die(t, g, theirs)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB125Counters(g, gimli); got != 0 {
		t.Errorf("Gimli has %d counters, want 0", got)
	}
}

// --- Strider, Ranger of the North --------------------------------------------

func TestGapB125StriderPumpsAndGivesFirstStrikeAtPowerFour(t *testing.T) {
	g, me, _ := gapB125Main(t)
	effectstest.B12Push(g, me.ID, "Strider, Ranger of the North", "Legendary Creature — Human Ranger", gapB125StriderOracle, 4, 4)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Ogre", 3, 3)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	gapB125SettleWithPick(t, g, me.ID, gapB125Ref(big), false)
	if p, tough := effectstest.EffectivePTOf(t, g, big); p != 4 || tough != 4 {
		t.Errorf("Big is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, big, "first strike") {
		t.Error("a creature that reached power 4 lacks first strike")
	}
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	gapB125SettleWithPick(t, g, me.ID, gapB125Ref(small), false)
	if p, _ := effectstest.EffectivePTOf(t, g, small); p != 2 || effectstest.HasEffectiveKeyword(t, g, small, "first strike") {
		t.Errorf("Small is power %d with first strike %v, want 2 without", p, effectstest.HasEffectiveKeyword(t, g, small, "first strike"))
	}
}

// --- Arwen Undómiel ------------------------------------------------------------

func TestGapB125ArwenPutsACounterWheneverYouScryAndScriesForSix(t *testing.T) {
	g, me, _ := gapB125Main(t)
	arwen := effectstest.B12Push(g, me.ID, "Arwen Undómiel", "Legendary Creature — Elf Noble", gapB125ArwenOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	gapB125Library(me, game.Card{Name: "A"}, game.Card{Name: "B"}, game.Card{Name: "C"})
	if err := effectstest.ConB1Activate(t, g, me, arwen, 0, "GUCCCC", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.ScryChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 2 {
		t.Fatal("the activation did not scry 2")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	gapB125SettleWithPick(t, g, me.ID, gapB125Ref(bear), false)
	if got := gapB125Counters(g, bear); got != 1 {
		t.Errorf("the Bear has %d counters, want 1", got)
	}
}

// --- Elrond, Master of Healing ---------------------------------------------------

func TestGapB125ElrondCountersUpToXCreaturesForTheCardsLookedAt(t *testing.T) {
	g, me, _ := gapB125Main(t)
	effectstest.B12Push(g, me.ID, "Elrond, Master of Healing", "Legendary Creature — Elf Noble", gapB125ElrondOracle, 4, 4)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	bb := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	gapB125Library(me, game.Card{Name: "1"}, game.Card{Name: "2"}, game.Card{Name: "3"})
	gapB125Scry(t, g, me, 2)
	// Two cards looked at: up to two targets, and a third is refused.
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the scry trigger")
	}
	third := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(a, bb, third)); err == nil {
		t.Error("three targets were accepted for a scry 2")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(a, bb)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if gapB125Counters(g, a) != 1 || gapB125Counters(g, bb) != 1 || gapB125Counters(g, third) != 0 {
		t.Errorf("counters = %d, %d, %d, want 1, 1, 0", gapB125Counters(g, a), gapB125Counters(g, bb), gapB125Counters(g, third))
	}
}

func TestGapB125ElrondDrawsWhenAnOpponentTargetsACounteredCreature(t *testing.T) {
	g, me, opp := gapB125Main(t)
	effectstest.B12Push(g, me.ID, "Elrond, Master of Healing", "Legendary Creature — Elf Noble", gapB125ElrondOracle, 4, 4)
	counters := effectstest.B12Creature(g, me.ID, "Counters", "Creature — Bear", 2, 2)
	bare := effectstest.B12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := (effects.AddCounter{Target: counters, Kind: game.CounterPlusOne, N: 1}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
			t.Fatal(err)
		}
	})
	hand := me.Hand.Size()
	effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetCard(bare))
	gapB125SettleWithPick(t, g, me.ID, game.TargetRef{}, true)
	if me.Hand.Size() != hand {
		t.Errorf("Elrond drew for a creature with no counter (hand %d -> %d)", hand, me.Hand.Size())
	}
	effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetCard(counters))
	gapB125SettleWithPick(t, g, me.ID, game.TargetRef{}, true)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1 for a creature with a +1/+1 counter", hand, me.Hand.Size())
	}
}

// --- Legolas, Counter of Kills -----------------------------------------------------

func TestGapB125LegolasUntapsOnAScryOnlyOnceEachTurn(t *testing.T) {
	g, me, _ := gapB125Main(t)
	legolas := effectstest.B12Push(g, me.ID, "Legolas, Counter of Kills", "Legendary Creature — Elf Archer", gapB125LegolasOracle, 2, 3)
	gapB125Library(me, game.Card{Name: "1"}, game.Card{Name: "2"}, game.Card{Name: "3"}, game.Card{Name: "4"})
	// Untapped: a scry does nothing for him.
	gapB125Scry(t, g, me, 1)
	gapB125SettleWithPick(t, g, me.ID, game.TargetRef{}, true)
	effectstest.B16Tap(g, legolas)
	gapB125Scry(t, g, me, 1)
	gapB125SettleWithPick(t, g, me.ID, game.TargetRef{}, true)
	if effectstest.HmlTapped(t, g, legolas) {
		t.Fatal("Legolas stayed tapped after a scry")
	}
	effectstest.B16Tap(g, legolas)
	gapB125Scry(t, g, me, 1)
	gapB125SettleWithPick(t, g, me.ID, game.TargetRef{}, true)
	if !effectstest.HmlTapped(t, g, legolas) {
		t.Error("Legolas untapped a second time in one turn")
	}
}

func TestGapB125LegolasDecliningDoesNotUseUpTheTurn(t *testing.T) {
	g, me, _ := gapB125Main(t)
	legolas := effectstest.B12Push(g, me.ID, "Legolas, Counter of Kills", "Legendary Creature — Elf Archer", gapB125LegolasOracle, 2, 3)
	gapB125Library(me, game.Card{Name: "1"}, game.Card{Name: "2"}, game.Card{Name: "3"})
	effectstest.B16Tap(g, legolas)
	gapB125Scry(t, g, me, 1)
	gapB125SettleWithPick(t, g, me.ID, game.TargetRef{}, false)
	if !effectstest.HmlTapped(t, g, legolas) {
		t.Fatal("Legolas untapped though the player declined")
	}
	gapB125Scry(t, g, me, 1)
	gapB125SettleWithPick(t, g, me.ID, game.TargetRef{}, true)
	if effectstest.HmlTapped(t, g, legolas) {
		t.Error("declining used up the once-a-turn allowance")
	}
}

func TestGapB125LegolasGrowsWhenAnOpponentsCreatureDies(t *testing.T) {
	g, me, opp := gapB125Main(t)
	legolas := effectstest.B12Push(g, me.ID, "Legolas, Counter of Kills", "Legendary Creature — Elf Archer", gapB125LegolasOracle, 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	gapB125Die(t, g, mine)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB125Counters(g, legolas); got != 0 {
		t.Errorf("Legolas has %d counters after my own creature died, want 0", got)
	}
	gapB125Die(t, g, theirs)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB125Counters(g, legolas); got != 1 {
		t.Errorf("Legolas has %d counters after an opponent's creature died, want 1", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, legolas, "reach") {
		t.Error("Legolas lacks reach")
	}
}
