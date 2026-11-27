package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b420_cards_test.go — Return to Ravnica, batch 420 (S82 older sets,
// #5566): six cards of card draw and library work. Every test plays the real
// catalog entry through the engine.

const (
	oldB420Vine       = "0f539127-535f-4e0d-abaa-e884521098d2"
	oldB420Sphinx     = "690df8d2-9538-4c51-b730-cee09ad0ef1c"
	oldB420Keyrune    = "8b8c0285-424f-4a82-bfb6-17752cfa71dd"
	oldB420Orders     = "93036f4e-f553-4cc7-a747-6cd087b0994c"
	oldB420Charm      = "e7592863-4aee-47b3-9abf-4b863ab0b6ec"
	oldB420Thoughtfl  = "faad0f02-da5b-47f4-9634-0ea6ee2ff4f6"
	oldB420BearType   = "Creature — Bear"
	oldB420ForestType = "Basic Land — Forest"
)

func TestOldB420EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB420Vine: "Gatecreeper Vine", oldB420Sphinx: "Sphinx of the Chimes", oldB420Keyrune: "Izzet Keyrune",
		oldB420Orders: "Jarad's Orders", oldB420Charm: "Azorius Charm", oldB420Thoughtfl: "Thoughtflare",
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

func TestOldB420GatecreeperVineFetchesABasicLandOrAGate(t *testing.T) {
	for _, want := range []string{"Forest", "Dimir Gate"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.SeedSearchLibrary(me,
			game.Card{Name: "Sol Ring", TypeLine: "Artifact"},
			game.Card{Name: "Nonbasic", TypeLine: "Land"},
			game.Card{Name: "Dimir Gate", TypeLine: "Land — Gate"},
			game.Card{Name: "Forest", TypeLine: oldB420ForestType})
		effectstest.CastAndResolveCreature(t, g, "Gatecreeper Vine", "Creature — Plant", oldB420Vine)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.SearchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("no search prompt")
		}
		if effectstest.SearchOptionNamed(g, c, "Sol Ring") != uuid.Nil || effectstest.SearchOptionNamed(g, c, "Nonbasic") != uuid.Nil {
			t.Error("the search offered a card that is neither a basic land nor a Gate")
		}
		effectstest.AnswerSearchNamed(t, g, me.ID, want)
		effectstest.PassPriorityAroundTable(t, g)
		found := false
		for _, card := range me.Hand.Cards {
			if card.Name == want {
				found = true
			}
		}
		if !found {
			t.Errorf("%s is not in hand", want)
		}
	}
}

func TestOldB420GatecreeperVineMayDeclineTheSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: oldB420ForestType})
	hand := me.Hand.Size()
	effectstest.CastAndResolveCreature(t, g, "Gatecreeper Vine", "Creature — Plant", oldB420Vine)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		effectstest.AnswerSearchFailToFind(t, g, me.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d (nothing was fetched)", me.Hand.Size(), hand)
	}
}

func TestOldB420SphinxOfTheChimesDiscardsAPairToDrawFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sphinx := effectstest.B12Push(g, me.ID, "Sphinx of the Chimes", "Creature — Sphinx", oldB420Sphinx, 5, 6)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	var ids []uuid.UUID
	for _, n := range []string{"Opt", "Shock", "Opt", "Forest", "Forest"} {
		typ := "Instant"
		if n == "Forest" {
			typ = oldB420ForestType
		}
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: typ, Owner: me.ID, Controller: me.ID})
		ids = append(ids, id)
	}
	for i := 0; i < 8; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Lib", TypeLine: "Instant"})
	}
	activate := func(pick ...uuid.UUID) error {
		return g.ActivateCatalogAbility(me.ID, sphinx, 0, game.ActivateAbilityParams{DiscardIDs: pick, Strict: true})
	}
	if err := activate(ids[0], ids[1]); err == nil {
		t.Fatal("two differently named cards paid the cost")
	}
	if err := activate(ids[3], ids[4]); err == nil {
		t.Fatal("two lands paid a cost of nonland cards")
	}
	hand, yard := me.Hand.Size(), me.Graveyard.Size()
	if err := activate(ids[0], ids[2]); err != nil {
		t.Fatalf("two Opts: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand-2+4 {
		t.Errorf("hand %d, want %d", got, hand-2+4)
	}
	if got := me.Graveyard.Size(); got != yard+2 {
		t.Errorf("graveyard %d, want %d", got, yard+2)
	}
	if !effectstest.HasEffectiveKeyword(t, g, sphinx, "flying") {
		t.Error("the Sphinx has no flying")
	}
}

func TestOldB420ThoughtflareDrawsFourThenDiscardsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 8; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Lib", TypeLine: "Instant"})
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Thoughtflare", "Instant", oldB420Thoughtfl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DiscardOwed(g, me.ID); got != 2 {
		t.Fatalf("%d cards owed, want 2", got)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+4-2 {
		t.Errorf("hand %d, want %d (the spell itself left it)", got, hand+4-2)
	}
}

func TestOldB420IzzetKeyruneTapsForUOrRAndAnimatesAsATwoOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	key := effectstest.B12Push(g, me.ID, "Izzet Keyrune", "Artifact", oldB420Keyrune, 0, 0)
	effectstest.AdvanceToMain(t, g)
	if p, _ := effectstest.PtOf(t, g, key); p != 0 {
		t.Fatalf("the Keyrune started as a creature with power %d", p)
	}
	effectstest.FloatMana(t, g, me, "{U}{R}")
	if err := g.ActivateCatalogAbility(me.ID, key, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("animate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, key); p != 2 || tough != 1 {
		t.Errorf("the Keyrune is %d/%d, want 2/1", p, tough)
	}
	colors := effectstest.EffectiveColors(t, g, key)
	if len(colors) != 2 {
		t.Errorf("colours %v, want blue and red", colors)
	}
	_ = opp
}

func TestOldB420IzzetKeyruneLootsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	key := effectstest.B12Creature(g, me.ID, "Izzet Keyrune", "Artifact Creature — Elemental", 2, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == key {
				g.Battlefield.Cards[i].OracleID = oldB420Keyrune
			}
		}
	})
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Lib", TypeLine: "Instant"})
	}
	hand := me.Hand.Size()
	effectstest.Pr7bAttack(t, g, opp.ID, key)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DiscardOwed(g, me.ID); got != 1 {
		t.Fatalf("%d cards owed after the loot, want 1", got)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand)
	}
}

func TestOldB420JaradsOrdersPutsOneCreatureInHandAndTheOtherInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: oldB420BearType},
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf"},
		game.Card{Name: "Elk", TypeLine: "Creature — Elk"})
	yard := me.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Jarad's Orders", "Sorcery", oldB420Orders, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Bolt") != uuid.Nil {
		t.Error("the search offered a noncreature card")
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Bear", "Wolf")
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.RevealPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick between the two creatures")
	}
	if err := g.ResolveRevealPick(pick.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	inHand := func(name string) bool {
		for _, card := range me.Hand.Cards {
			if card.Name == name {
				return true
			}
		}
		return false
	}
	inYard := func(name string) bool {
		for _, card := range me.Graveyard.Cards {
			if card.Name == name {
				return true
			}
		}
		return false
	}
	if !inHand("Wolf") || inHand("Bear") {
		t.Errorf("hand has Wolf %v and Bear %v, want true false", inHand("Wolf"), inHand("Bear"))
	}
	if !inYard("Bear") || me.Graveyard.Size() < yard+1 {
		t.Error("the other creature is not in the graveyard")
	}
}

func TestOldB420JaradsOrdersWithOneCreatureFoundPutsItInHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Bear", TypeLine: oldB420BearType})
	effectstest.CastCatalogSpell(t, g, "Jarad's Orders", "Sorcery", oldB420Orders, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		effectstest.AnswerSearchNamed(t, g, me.ID, "Bear")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.RevealPickFor(g, me.ID); pick != nil {
		if err := g.ResolveRevealPick(pick.ID, me.ID, ids[1:2]); err != nil {
			t.Fatalf("ResolveRevealPick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	found := false
	for _, card := range me.Hand.Cards {
		if card.Name == "Bear" {
			found = true
		}
	}
	if !found {
		t.Error("the only creature found did not reach the hand")
	}
}

func TestOldB420AzoriusCharmModes(t *testing.T) {
	castMode := func(g *game.Game, mode int, targets []game.TargetRef) {
		effectstest.CastCatalogSpellWithModesTargets(t, g, oldB420Charm, []int{mode}, targets)
		effectstest.PassPriorityAroundTable(t, g)
	}

	// Mode 1: lifelink for your creatures only.
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB420BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB420BearType, 2, 2)
	castMode(g, 0, nil)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "lifelink") || effectstest.HasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Error("lifelink did not go to exactly your creatures")
	}

	// Mode 2: draw a card.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Lib", TypeLine: "Instant"})
	}
	hand := me.Hand.Size()
	castMode(g, 1, nil)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestOldB420AzoriusCharmTucksAnAttackingCreatureAndRefusesAnIdleOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB420BearType, 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", oldB420BearType, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	lib := me.Library.Size()
	id := effectstest.HandCardFull(me, "Azorius Charm", "Instant", "{W}{U}", oldB420Charm, []string{"W", "U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{2}, Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("Azorius Charm targeted a creature that is neither attacking nor blocking")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{2}, Targets: effectstest.CardTarget(attacker)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacker is still on the battlefield")
	}
	top, err := me.Library.Top()
	if err != nil || top.InstanceID != attacker {
		t.Errorf("the attacker is not on top of its owner's library (top %v, err %v)", top.Name, err)
	}
	if me.Library.Size() != lib+1 {
		t.Errorf("library %d, want %d", me.Library.Size(), lib+1)
	}
}
