package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b69_test.go — Ravnica Remastered artifacts and creatures, batch 69
// (#5948). Each test plays the real catalog entry through the engine.

const (
	gapB69BottledCloisterOracle = "1fd5e5e7-f23d-4891-86c4-268195ea09fc"
	gapB69CivicSaberOracle      = "79441673-6e43-41bc-936c-958dbb9d92da"
	gapB69SealOracle            = "9a2d60bc-3024-4a1b-ad74-81645fd227c4"
	gapB69SwordOracle           = "cc33444e-da8c-4af5-b3ed-552e91b0e656"
	gapB69VoyagerStaffOracle    = "4573f044-7011-412e-81bd-ae8d0c21283b"
	gapB69JunktrollerOracle     = "6ef46fd7-82db-45d3-b02e-7312384577f3"
	gapB69DreadmalkinOracle     = "118cf682-a532-42e9-8d91-6ce6f0d31ace"
	gapB69LordOfTheVoidOracle   = "626b85f8-8889-46e6-9659-460bac79ae9b"
	gapB69WoebringerOracle      = "21ea9501-9a0f-43a3-a2cc-e9648325c45e"
)

func TestGapB69EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB69BottledCloisterOracle: "Bottled Cloister",
		gapB69CivicSaberOracle:      "Civic Saber",
		gapB69SealOracle:            "Seal of the Guildpact",
		gapB69SwordOracle:           "Sword of the Paruns",
		gapB69VoyagerStaffOracle:    "Voyager Staff",
		gapB69JunktrollerOracle:     "Junktroller",
		gapB69DreadmalkinOracle:     "Dreadmalkin",
		gapB69LordOfTheVoidOracle:   "Lord of the Void",
		gapB69WoebringerOracle:      "Woebringer Demon",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestGapB69CivicSaberAddsPowerPerColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	saber := seedEquipment(g, me.ID, "Civic Saber", gapB69CivicSaberOracle)
	two := apaPush(g, me.ID, me.ID, game.Card{Name: "Two", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"W", "B"}})
	none := apaPush(g, me.ID, me.ID, game.Card{Name: "None", TypeLine: "Artifact Creature — Golem", Power: 2, Toughness: 2})
	mono := apaPush(g, me.ID, me.ID, game.Card{Name: "Mono", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})

	equipTo(t, g, me.ID, saber, two)
	if p, tg := effectivePower(t, g, two), effectiveToughness(t, g, two); p != 4 || tg != 2 {
		t.Errorf("a two-color creature is %d/%d, want 4/2", p, tg)
	}
	untap(g, saber)
	fillPool(me, 1)
	equipTo(t, g, me.ID, saber, none)
	if p := effectivePower(t, g, none); p != 2 {
		t.Errorf("a colorless creature has power %d, want 2", p)
	}
	if p := effectivePower(t, g, two); p != 2 {
		t.Errorf("the old host kept power %d, want 2", p)
	}
	fillPool(me, 1)
	equipTo(t, g, me.ID, saber, mono)
	if p := effectivePower(t, g, mono); p != 3 {
		t.Errorf("a one-color creature has power %d, want 3", p)
	}
}

func TestGapB69JunktrollerTucksAGraveyardCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := b12Push(g, me.ID, "Junktroller", "Artifact Creature — Golem", gapB69JunktrollerOracle, 0, 6)
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	advanceToMain(t, g)
	if err := eoeCDActivate(g, me.ID, golem, 0, eoeCDCard(dead)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(dead) {
		t.Fatal("the card is still in the graveyard")
	}
	if cards := opp.Library.Cards; len(cards) == 0 || cards[0].InstanceID != dead {
		t.Error("the card is not on the bottom of its owner's library")
	}
	if !findBattlefieldCardForTest(g, golem).Tapped {
		t.Error("the Junktroller should be tapped")
	}
	// A tapped Junktroller cannot activate again.
	if err := eoeCDActivate(g, me.ID, golem, 0, eoeCDCard(dead)); err == nil {
		t.Error("a tapped Junktroller activated")
	}
}

func TestGapB69VoyagerStaffBlinksAtTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	staff := pushCatalogPermanent(g, me.ID, "Voyager Staff", "Artifact", gapB69VoyagerStaffOracle, false)
	bear := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	advanceToMain(t, g)
	fillPool(me, 2)
	if err := eoeCDActivate(g, me.ID, staff, 0, eoeCDCard(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if findBattlefieldCardForTest(g, staff) != nil {
		t.Error("the Staff should be sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, bear) != nil {
		t.Fatal("the bear is still on the battlefield")
	}
	if z := g.FindCardZoneForEffect(bear); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("the bear is in %v, want exile", z)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	back := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" {
			back = true
			if c.Controller != opp.ID {
				t.Errorf("the bear came back under %s, want its owner", c.Controller)
			}
		}
	}
	if !back {
		t.Error("the bear did not return at the beginning of the end step")
	}
}

func TestGapB69DreadmalkinSacrificesAnotherCreatureOrPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cat := b12Push(g, me.ID, "Dreadmalkin", "Creature — Zombie Cat", gapB69DreadmalkinOracle, 1, 1)
	fodder := apaPush(g, me.ID, me.ID, game.Card{Name: "Fodder", TypeLine: "Creature — Goblin", Power: 1, Toughness: 1})
	walker := apaPush(g, me.ID, me.ID, game.Card{Name: "Walker", TypeLine: "Planeswalker — Test", Counters: map[string]int{"loyalty": 3}})
	land := apaPush(g, me.ID, me.ID, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	advanceToMain(t, g)
	fillPoolColored(me, "B", 9)
	activate := func(ids ...uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, cat, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: ids})
	}
	for _, bad := range []uuid.UUID{cat, land} {
		if err := activate(bad); err == nil {
			t.Errorf("sacrificing %s paid the cost", bad)
		}
	}
	if err := activate(fodder); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := eoeCDPT(t, g, cat); p != 3 || tg != 3 {
		t.Errorf("Dreadmalkin is %d/%d, want 3/3", p, tg)
	}
	if findBattlefieldCardForTest(g, fodder) != nil {
		t.Error("the fodder survived")
	}
	if err := activate(walker); err != nil {
		t.Fatalf("a planeswalker is a legal sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := eoeCDPT(t, g, cat); p != 5 {
		t.Errorf("Dreadmalkin has power %d, want 5", p)
	}
	if findBattlefieldCardForTest(g, walker) != nil {
		t.Error("the planeswalker survived")
	}
}

func TestGapB69SwordOfTheParunsFollowsTheHostsTappedState(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sword := seedEquipment(g, me.ID, "Sword of the Paruns", gapB69SwordOracle)
	host := apaPush(g, me.ID, me.ID, game.Card{Name: "Host", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	other := apaPush(g, me.ID, me.ID, game.Card{Name: "Other", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	fillPool(me, 3)
	equipTo(t, g, me.ID, sword, host)

	pt := func(id uuid.UUID) (int, int) { return effectivePower(t, g, id), effectiveToughness(t, g, id) }
	// Untapped host: untapped creatures you control get +0/+2.
	if p, tg := pt(host); p != 2 || tg != 4 {
		t.Errorf("untapped host %d/%d, want 2/4", p, tg)
	}
	if p, tg := pt(other); p != 2 || tg != 4 {
		t.Errorf("untapped other %d/%d, want 2/4", p, tg)
	}
	if p, tg := pt(theirs); p != 2 || tg != 2 {
		t.Errorf("an opponent's creature %d/%d, want 2/2", p, tg)
	}
	// Tap the host: tapped creatures you control get +2/+0, untapped ones lose the bonus.
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(host) })
	if p, tg := pt(host); p != 4 || tg != 2 {
		t.Errorf("tapped host %d/%d, want 4/2", p, tg)
	}
	if p, tg := pt(other); p != 2 || tg != 2 {
		t.Errorf("untapped other while the host is tapped %d/%d, want 2/2", p, tg)
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(other) })
	if p, tg := pt(other); p != 4 || tg != 2 {
		t.Errorf("tapped other %d/%d, want 4/2", p, tg)
	}
}

func TestGapB69SwordOfTheParunsTapsOrUntapsTheHost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	sword := seedEquipment(g, me.ID, "Sword of the Paruns", gapB69SwordOracle)
	host := apaPush(g, me.ID, me.ID, game.Card{Name: "Host", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	fillPool(me, 3)
	equipTo(t, g, me.ID, sword, host)
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if !findBattlefieldCardForTest(g, host).Tapped {
		t.Fatal("the host should have been tapped")
	}
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	if findBattlefieldCardForTest(g, host).Tapped {
		t.Fatal("the host should have been untapped")
	}
	fillPool(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	if findBattlefieldCardForTest(g, host).Tapped {
		t.Fatal("declining changed the host")
	}
}

func TestGapB69WoebringerDemonMakesEachUpkeepPlayerSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	demon := pushCatalogPermanent(g, me.ID, "Woebringer Demon", "Creature — Demon", gapB69WoebringerOracle, false)
	a := apaPush(g, opp.ID, opp.ID, game.Card{Name: "A", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	b := apaPush(g, opp.ID, opp.ID, game.Card{Name: "B", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	prompt := chooseCardsChoiceFor(g, opp.ID)
	if prompt == nil {
		t.Fatal("the upkeep player was not asked to sacrifice")
	}
	if len(prompt.ChooseCards) != 2 {
		t.Errorf("offered %v, want their two creatures only", prompt.ChooseCards)
	}
	answerChooseCards(t, g, opp.ID, b)
	if findBattlefieldCardForTest(g, b) != nil || findBattlefieldCardForTest(g, a) == nil {
		t.Error("the wrong creature was sacrificed")
	}
	if findBattlefieldCardForTest(g, demon) == nil {
		t.Error("the Demon was sacrificed though the player had a creature")
	}
}

func TestGapB69WoebringerDemonSacrificesItselfWhenThePlayerCant(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	demon := pushCatalogPermanent(g, me.ID, "Woebringer Demon", "Creature — Demon", gapB69WoebringerOracle, false)
	_ = opp
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, demon) != nil {
		t.Error("the Demon stayed though the upkeep player had no creature")
	}
}

func TestGapB69WoebringerDemonMayBeItsOwnControllersSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	demon := pushCatalogPermanent(g, me.ID, "Woebringer Demon", "Creature — Demon", gapB69WoebringerOracle, false)
	advanceToUpkeepOf(t, g, 0)
	// Walk a full turn so seat 0's upkeep comes round again.
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if prompt := chooseCardsChoiceFor(g, me.ID); prompt != nil {
		answerChooseCards(t, g, me.ID, demon)
	}
	if findBattlefieldCardForTest(g, demon) != nil {
		t.Error("the Demon was not sacrificed: it was its controller's only creature")
	}
}

func TestGapB69BottledCloisterExilesTheHandAndReturnsItWithADraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cloister := pushCatalogPermanent(g, me.ID, "Bottled Cloister", "Artifact", gapB69BottledCloisterOracle, false)
	h1 := handCardForTest(me, "One", "Sorcery", "")
	h2 := handCardForTest(me, "Two", "Sorcery", "")
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 {
		t.Fatalf("hand has %d cards after an opponent's upkeep, want 0", me.Hand.Size())
	}
	for _, id := range []uuid.UUID{h1, h2} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Fatalf("card %s is in %v, want exile", id, z)
		}
	}
	// Our own upkeep (an opponent's did not trigger for it) returns them, then draws.
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{h1, h2} {
		if !me.Hand.Contains(id) {
			t.Errorf("card %s did not return to hand", id)
		}
	}
	if me.Hand.Size() < 3 {
		t.Errorf("hand has %d cards, want the two back and a draw", me.Hand.Size())
	}
	_ = cloister
}

func TestGapB69BottledCloisterReturnsTheHandOnItsOwnersUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Bottled Cloister", "Artifact", gapB69BottledCloisterOracle, false)
	h1 := handCardForTest(me, "One", "Sorcery", "")
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	// Seat 1's upkeep exiled the hand; ours returned it with a draw.
	if !me.Hand.Contains(h1) {
		t.Error("the exiled card did not come back on its controller's upkeep")
	}
	if z := g.FindCardZoneForEffect(h1); z == nil || z.Kind != game.ZoneHand {
		t.Errorf("card is in %v, want hand", z)
	}
}

func TestGapB69LordOfTheVoidTakesACreatureFromTheTopSeven(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lord := apaPush(g, me.ID, me.ID, game.Card{Name: "Lord of the Void", OracleID: gapB69LordOfTheVoidOracle, TypeLine: "Creature — Demon", Power: 7, Toughness: 7, Keywords: []string{"flying"}})
	var rows []struct{ name, typeLine string }
	for i := 0; i < 5; i++ {
		rows = append(rows, struct{ name, typeLine string }{"Filler", "Sorcery"})
	}
	rows = append(rows, struct{ name, typeLine string }{"Deep Creature", "Creature — Elf"}, struct{ name, typeLine string }{"Seventh", "Creature — Elf"}, struct{ name, typeLine string }{"Eighth", "Creature — Elf"})
	top := hornTopFive(opp, rows)
	for i := range opp.Library.Cards {
		opp.Library.Cards[i].Power, opp.Library.Cards[i].Toughness = 1, 1
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(lord, opp.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	prompt := chooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no creature prompt after combat damage")
	}
	if len(prompt.ChooseCards) != 2 || prompt.ChooseMin != 1 || prompt.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the two creature cards among the top seven, exactly one", prompt.ChooseCards, prompt.ChooseMin, prompt.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, top[5])
	onField := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Deep Creature" {
			onField = c.Controller == me.ID
		}
	}
	if !onField {
		t.Error("the chosen creature is not on the battlefield under the Lord's controller")
	}
	for _, id := range append(append([]uuid.UUID{}, top[:5]...), top[6]) {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Errorf("card %s is in %v, want exile", id, z)
		}
	}
	if z := g.FindCardZoneForEffect(top[7]); z == nil || z.Kind != game.ZoneLibrary {
		t.Errorf("the eighth card is in %v, want the library", z)
	}
}

func TestGapB69SealOfTheGuildpactDiscountsPerChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	fillPool(me, 5)
	castCatalogSpell(t, g, "Seal of the Guildpact", "Artifact", gapB69SealOracle, nil)
	passPriorityAroundTable(t, g)
	first := pendingOfKind(g, game.PendingChoiceColor)
	if first == nil {
		t.Fatal("no first color prompt")
	}
	if len(first.ColorOptions) != 5 {
		t.Errorf("first prompt offers %v, want all five", first.ColorOptions)
	}
	answerColor(t, g, me.ID, "W")
	second := pendingOfKind(g, game.PendingChoiceColor)
	if second == nil {
		t.Fatal("no second color prompt")
	}
	for _, o := range second.ColorOptions {
		if o == "W" {
			t.Errorf("the second prompt offers the first color again: %v", second.ColorOptions)
		}
	}
	answerColor(t, g, me.ID, "U")
	var seal uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Seal of the Guildpact" {
			seal = c.InstanceID
		}
	}
	if got := g.ChosenColorOf(seal); got != "WU" {
		t.Fatalf("chosen colors = %q, want WU", got)
	}
	// A white-blue {2}{W}{U} spell costs {2} less: {W}{U} pays for it.
	me.ManaPool = game.ManaPool{}
	me.ManaPool.AddMana(game.ManaToken{Color: "W"})
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	both := gapB69HandCard(me, game.Card{Name: "Both", TypeLine: "Creature — Test", ManaCost: "{2}{W}{U}", Colors: []string{"W", "U"}, Power: 1, Toughness: 1})
	if err := g.CastSpell(me.ID, both, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a spell of both chosen colors should cost {2} less: %v", err)
	}
	passPriorityAroundTable(t, g)
	// A red spell of the same shape is not discounted.
	me.ManaPool = game.ManaPool{}
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	red := gapB69HandCard(me, game.Card{Name: "Red", TypeLine: "Creature — Test", ManaCost: "{1}{R}", Colors: []string{"R"}, Power: 1, Toughness: 1})
	if err := g.CastSpell(me.ID, red, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a spell of neither chosen color is not discounted, but this one is fully paid: %v", err)
	}
	passPriorityAroundTable(t, g)
	me.ManaPool = game.ManaPool{}
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	red2 := gapB69HandCard(me, game.Card{Name: "Red Two", TypeLine: "Creature — Test", ManaCost: "{1}{R}", Colors: []string{"R"}, Power: 1, Toughness: 1})
	if err := g.CastSpell(me.ID, red2, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a red spell was discounted by a W/U Seal")
	}
}

func gapB69HandCard(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}
