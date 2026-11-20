package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB33EnvSciencesOracle  = "2f889671-a1d6-4034-881b-58c93897d707"
	gapB33IndicateOracle     = "67876214-3777-41b1-935e-75dd5075fa53"
	gapB33CallGraveOracle    = "252ff28e-3ea7-4e2c-a30c-3cb80aa278d0"
	gapB33DeadRingersOracle  = "fb34e487-f7a9-48ad-8f42-00295a80a160"
	gapB33PoxOracle          = "539d7c1a-b5e9-4204-9109-ddb2e8dec168"
	gapB33RainDaggersOracle  = "e2048201-6dc9-4cf5-916f-1d867ae8dbdd"
	gapB33StirringsOracle    = "82f18e7d-5c42-47c4-8e74-3fccc9b7b1f0"
	gapB33NaturalOrderOracle = "8c1fe337-375a-4add-93b6-0ac39ed72b4f"
)

func TestGapB33EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB33EnvSciencesOracle:  "Environmental Sciences",
		gapB33IndicateOracle:     "Indicate",
		gapB33CallGraveOracle:    "Call from the Grave",
		gapB33DeadRingersOracle:  "Dead Ringers",
		gapB33PoxOracle:          "Pox",
		gapB33RainDaggersOracle:  "Rain of Daggers",
		gapB33StirringsOracle:    "Ancient Stirrings",
		gapB33NaturalOrderOracle: "Natural Order",
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

func gapB33Perm(g *game.Game, owner uuid.UUID, name, typeLine, mana string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		ManaCost: mana, Power: 1, Toughness: 1, Owner: owner, Controller: owner})
}

func TestGapB33EnvironmentalSciencesFetchesABasicLandAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me,
		game.Card{Name: "Cradle", TypeLine: "Legendary Land"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	life := me.Life
	castCatalogSpell(t, g, "Environmental Sciences", "Sorcery — Lesson", gapB33EnvSciencesOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Cradle") != uuid.Nil {
		t.Error("a nonbasic land was offered")
	}
	answerSearchNamed(t, g, me.ID, "Forest")
	if !b02bHandHasNamed(me, "Forest") {
		t.Error("the basic land did not reach the hand")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
}

func TestGapB33IndicateOnlyNeedsAPermanentTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	perm := gapB33Perm(g, g.Seats[1].ID, "Rock", "Artifact", "")
	life := me.Life
	castCatalogSpell(t, g, "Indicate", "Sorcery", gapB33IndicateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: perm}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(perm) || me.Life != life {
		t.Errorf("Indicate did something: on battlefield %v, life %d", g.Battlefield.Contains(perm), me.Life)
	}
	if err := castCatalogSpellErrStrict(t, g, "Indicate", "Sorcery", gapB33IndicateOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}); err == nil {
		t.Error("Indicate targeted a player")
	}
}

func TestGapB33RainOfDaggersLosesTwoPerCreatureReallyDestroyed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	a := bnsCreature(g, opp.ID, "A", "{1}", 1, 1)
	b := bnsCreature(g, opp.ID, "B", "{1}", 1, 1)
	tough := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Tough", TypeLine: "Creature — Golem",
		Power: 2, Toughness: 2, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID})
	mine := bnsCreature(g, me.ID, "Mine", "{1}", 1, 1)
	elsewhere := bnsCreature(g, third.ID, "Third", "{1}", 1, 1)
	life := me.Life
	if err := bnsCast(t, g, "Rain of Daggers", "Sorcery", gapB33RainDaggersOracle, "{4}{B}{B}",
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a creature of the target opponent survived")
	}
	if !g.Battlefield.Contains(tough) || !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(elsewhere) {
		t.Error("the indestructible creature or another player's creature was destroyed")
	}
	if me.Life != life-4 {
		t.Errorf("life %d → %d, want -4 (two destroyed)", life, me.Life)
	}
}

func TestGapB33AncientStirringsTakesAColorlessCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var top [5]uuid.UUID
	cards := []struct{ name, typeLine, mana string }{
		{"Bear", "Creature — Bear", "{1}{G}"},
		{"Wastes", "Basic Land", ""},
		{"Sol Ring", "Artifact", "{1}"},
		{"Shock", "Instant", "{R}"},
		{"Eldrazi", "Creature — Eldrazi", "{3}"},
	}
	for i := len(cards) - 1; i >= 0; i-- {
		top[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: top[i], Name: cards[i].name, TypeLine: cards[i].typeLine,
			ManaCost: cards[i].mana, Owner: me.ID, Controller: me.ID})
	}
	castCatalogSpell(t, g, "Ancient Stirrings", "Sorcery", gapB33StirringsOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if len(offered) != 3 || !offered[top[1]] || !offered[top[2]] || !offered[top[4]] || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the land, the artifact and the colorless creature, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[2]) {
		t.Fatal("the colorless card is not in hand")
	}
	if putInLibraryChoiceFor(g, me.ID) == nil {
		t.Fatal("no order prompt for the rest on the bottom")
	}
}

func TestGapB33NaturalOrderSacrificesAGreenCreatureAndFetchesOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	red := gapB33Perm(g, me.ID, "Goblin", "Creature — Goblin", "{R}")
	green := gapB33Perm(g, me.ID, "Elf", "Creature — Elf", "{G}")
	seedSearchLibrary(me,
		game.Card{Name: "Titan", TypeLine: "Creature — Giant", ManaCost: "{4}{G}{G}"},
		game.Card{Name: "Dragon", TypeLine: "Creature — Dragon", ManaCost: "{4}{R}{R}"},
	)
	if _, err := tryCastWithSacrifice(g, "Natural Order", "Sorcery", gapB33NaturalOrderOracle, []uuid.UUID{red}); err == nil {
		t.Fatal("sacrificed a non-green creature")
	}
	castWithSacrifice(t, g, "Natural Order", "Sorcery", gapB33NaturalOrderOracle, green)
	if g.Battlefield.Contains(green) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Dragon") != uuid.Nil {
		t.Error("a non-green creature was offered")
	}
	answerSearchNamed(t, g, me.ID, "Titan")
	found := false
	for _, bc := range gapB33Battlefield(g) {
		if bc.Name == "Titan" {
			found = true
		}
	}
	if !found {
		t.Error("the green creature did not enter the battlefield")
	}
}

func TestGapB33DeadRingersDestroysOnlyAMatchingPair(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	refs := func(ids ...uuid.UUID) []game.TargetRef {
		var out []game.TargetRef
		for _, id := range ids {
			out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
		}
		return out
	}
	// Two green creatures: the same colours, both destroyed.
	a := gapB33Perm(g, opp.ID, "A", "Creature — Bear", "{G}")
	b := gapB33Perm(g, opp.ID, "B", "Creature — Bear", "{1}{G}")
	castCatalogSpell(t, g, "Dead Ringers", "Sorcery", gapB33DeadRingersOracle, refs(a, b))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("two green creatures survived")
	}
	// A green and a red one differ: nothing is destroyed.
	c := gapB33Perm(g, opp.ID, "C", "Creature — Bear", "{G}")
	d := gapB33Perm(g, opp.ID, "D", "Creature — Goblin", "{R}")
	castCatalogSpell(t, g, "Dead Ringers", "Sorcery", gapB33DeadRingersOracle, refs(c, d))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(c) || !g.Battlefield.Contains(d) {
		t.Error("creatures of different colours were destroyed")
	}
	// A green-red one and a green one differ: nothing is destroyed.
	e := gapB33Perm(g, opp.ID, "E", "Creature — Bear", "{R}{G}")
	castCatalogSpell(t, g, "Dead Ringers", "Sorcery", gapB33DeadRingersOracle, refs(c, e))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(c) || !g.Battlefield.Contains(e) {
		t.Error("a multicolour and a mono-colour creature were destroyed")
	}
	// Two colourless creatures share the same (empty) colours.
	f := gapB33Perm(g, opp.ID, "F", "Artifact Creature — Golem", "{2}")
	h := gapB33Perm(g, opp.ID, "H", "Artifact Creature — Golem", "{3}")
	castCatalogSpell(t, g, "Dead Ringers", "Sorcery", gapB33DeadRingersOracle, refs(f, h))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(f) || g.Battlefield.Contains(h) {
		t.Error("two colourless creatures survived")
	}
	// A black creature is not a legal target.
	k := gapB33Perm(g, opp.ID, "K", "Creature — Zombie", "{B}")
	m := gapB33Perm(g, opp.ID, "M", "Creature — Zombie", "{1}{B}")
	if err := castCatalogSpellErrStrict(t, g, "Dead Ringers", "Sorcery", gapB33DeadRingersOracle, refs(k, m)); err == nil {
		t.Error("targeted black creatures")
	}
}

func TestGapB33CallFromTheGravePutsARandomCreatureIntoPlayAndHurts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fallen", TypeLine: "Creature — Bear", ManaCost: "{2}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", ManaCost: "{R}",
		Owner: opp.ID, Controller: opp.ID})
	life := me.Life
	castCatalogSpell(t, g, "Call from the Grave", "Sorcery", gapB33CallGraveOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the only creature in any graveyard did not enter")
	}
	for _, bc := range gapB33Battlefield(g) {
		if bc.InstanceID == dead && bc.Controller != me.ID {
			t.Errorf("it entered under %v, want the caster", bc.Controller)
		}
	}
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want 3 damage (mana value 3)", life, me.Life)
	}
}

func TestGapB33CallFromTheGraveWithNoCreatureDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	castCatalogSpell(t, g, "Call from the Grave", "Sorcery", gapB33CallGraveOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d → %d with no creature to return", life, me.Life)
	}
}

// gapB33SettlePox answers every open choice with its first legal cards until
// none is left, the way a table of players picking arbitrarily would.
func gapB33SettlePox(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 40; i++ {
		answered := false
		for _, p := range g.Seats {
			c := latestChooseCardsFor(g, p.ID)
			if c == nil {
				continue
			}
			n := c.ChooseMax
			if n > len(c.ChooseCards) {
				n = len(c.ChooseCards)
			}
			if err := g.ResolveChooseCards(c.ID, p.ID, append([]uuid.UUID(nil), c.ChooseCards[:n]...)); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			answered = true
		}
		passPriorityAroundTable(t, g)
		if !answered && i > 3 {
			return
		}
	}
}

func TestGapB33PoxTakesAThirdRoundedUpOfEverything(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	me.Life, opp.Life, third.Life = 20, 10, 4
	opp.Hand.Cards = opp.Hand.Cards[:4]     // a third of 4, rounded up: 2
	third.Hand.Cards = third.Hand.Cards[:1] // a third of 1, rounded up: 1
	var oppCreatures, oppLands, thirdCreature []uuid.UUID
	for i := 0; i < 4; i++ {
		oppCreatures = append(oppCreatures, gapB33Perm(g, opp.ID, "OC", "Creature — Bear", "{1}"))
	}
	for i := 0; i < 5; i++ {
		oppLands = append(oppLands, gapB33Perm(g, opp.ID, "OL", "Basic Land — Forest", ""))
	}
	thirdCreature = append(thirdCreature, gapB33Perm(g, third.ID, "TC", "Creature — Bear", "{1}"))
	myHand := me.Hand.Size() // the Pox goes onto the stack, so the count after casting
	if err := bnsCast(t, g, "Pox", "Sorcery", gapB33PoxOracle, "{B}{B}{B}", game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	gapB33SettlePox(t, g)
	if me.Life != 13 || opp.Life != 6 || third.Life != 2 {
		t.Errorf("life = %d / %d / %d, want 13 / 6 / 2 (a third of 20, 10 and 4, rounded up)", me.Life, opp.Life, third.Life)
	}
	if got := opp.Hand.Size(); got != 2 {
		t.Errorf("the opponent's hand is %d, want 2", got)
	}
	if got := third.Hand.Size(); got != 0 {
		t.Errorf("the third seat's hand is %d, want 0", got)
	}
	if want := myHand - (myHand+2)/3; me.Hand.Size() != want {
		t.Errorf("the caster's hand is %d, want %d", me.Hand.Size(), want)
	}
	count := func(ids []uuid.UUID) int {
		n := 0
		for _, id := range ids {
			if g.Battlefield.Contains(id) {
				n++
			}
		}
		return n
	}
	if got := count(oppCreatures); got != 2 {
		t.Errorf("the opponent kept %d of 4 creatures, want 2", got)
	}
	if got := count(oppLands); got != 3 {
		t.Errorf("the opponent kept %d of 5 lands, want 3", got)
	}
	if got := count(thirdCreature); got != 0 {
		t.Errorf("the third seat kept its only creature")
	}
}

func gapB33Battlefield(g *game.Game) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() { out = g.BattlefieldCardsForEffect() })
	return out
}
