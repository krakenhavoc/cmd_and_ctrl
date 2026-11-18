package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b12_test.go — Midnight Hunt, batch 12 (#5497): eight transforming
// cards. Each test plays the real catalog entry through the engine.

type midB12Row struct {
	oracle      string
	front, back game.Face
}

func midB12Rows() []midB12Row {
	return []midB12Row{
		{"1a7e367e-1d8d-463d-a35f-4bad44a5ebea",
			game.Face{Name: "Mystic Skull", TypeLine: "Artifact", ManaCost: "{2}"},
			game.Face{Name: "Mystic Monstrosity", TypeLine: "Artifact Creature — Construct", Power: 5, Toughness: 6}},
		{"bfa60dd1-fcc8-4126-ab7a-4695c74f5bce",
			game.Face{Name: "Bereaved Survivor", TypeLine: "Creature — Human Peasant", ManaCost: "{2}{W}", Colors: []string{"W"}, Power: 2, Toughness: 1},
			game.Face{Name: "Dauntless Avenger", TypeLine: "Creature — Human Soldier", Colors: []string{"W"}, Power: 3, Toughness: 2}},
		{"c92d6f26-5856-42dc-9ee5-21292d78a1fb",
			game.Face{Name: "Deathbonnet Sprout", TypeLine: "Creature — Fungus", ManaCost: "{G}", Colors: []string{"G"}, Power: 1, Toughness: 1},
			game.Face{Name: "Deathbonnet Hulk", TypeLine: "Creature — Fungus Horror", Colors: []string{"G"}, Power: 3, Toughness: 3}},
		{"26ff3da5-4e02-41d1-a52f-3f8e87b01a77",
			game.Face{Name: "Ecstatic Awakener", TypeLine: "Creature — Human Wizard", ManaCost: "{B}", Colors: []string{"B"}, Power: 1, Toughness: 1},
			game.Face{Name: "Awoken Demon", TypeLine: "Creature — Demon", Colors: []string{"B"}, Power: 4, Toughness: 4}},
		{"b32253d6-5f00-40a8-a7e5-dc4655cd288c",
			game.Face{Name: "Mysterious Tome", TypeLine: "Artifact — Book", ManaCost: "{2}{U}", Colors: []string{"U"}},
			game.Face{Name: "Chilling Chronicle", TypeLine: "Artifact — Book", Colors: []string{"U"}}},
		{"b33390dd-f2f8-4059-83a9-45c0edccd5e4",
			game.Face{Name: "Heirloom Mirror", TypeLine: "Artifact", ManaCost: "{1}{B}", Colors: []string{"B"}},
			game.Face{Name: "Inherited Fiend", TypeLine: "Creature — Demon", Colors: []string{"B"}, Power: 4, Toughness: 4}},
		{"7493d78c-26c4-402e-bec5-6449773d0344",
			game.Face{Name: "Smoldering Egg", TypeLine: "Creature — Dragon Egg", ManaCost: "{1}{R}", Colors: []string{"R"}, Power: 0, Toughness: 4},
			game.Face{Name: "Ashmouth Dragon", TypeLine: "Creature — Dragon", Colors: []string{"R"}, Power: 4, Toughness: 4}},
		{"b85e7ab2-6f57-4b65-8c16-034b342996f0",
			game.Face{Name: "Ambitious Farmhand", TypeLine: "Creature — Human Peasant", ManaCost: "{1}{W}", Colors: []string{"W"}, Power: 1, Toughness: 1},
			game.Face{Name: "Seasoned Cathar", TypeLine: "Creature — Human Knight", Colors: []string{"W"}, Power: 3, Toughness: 3}},
	}
}

func midB12Row_(name string) midB12Row {
	for _, r := range midB12Rows() {
		if r.front.Name == name {
			return r
		}
	}
	panic("no row " + name)
}

func (r midB12Row) card(owner uuid.UUID, face int) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: r.oracle, Layout: game.LayoutTransform,
		Owner: owner, Controller: owner, Faces: []game.Face{r.front, r.back},
	}
	c.SetFace(face)
	return c
}

func (r midB12Row) put(g *game.Game, owner uuid.UUID, face int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, r.card(owner, face))
}

func midB12Name(g *game.Game, id uuid.UUID) string {
	c, ok := battlefieldCard(g, id)
	if !ok {
		return "(gone)"
	}
	return c.Name
}

// midB12Activate pays `mana` into the pool and activates ability idx Strict.
func midB12Activate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, idx int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if mana != "" {
		dcMana(t, g, me.ID, mana)
	}
	params.Strict = true
	err := g.ActivateCatalogAbility(me.ID, id, idx, params)
	if err == nil {
		passPriorityAroundTable(t, g)
	}
	return err
}

func TestMidB12EveryFaceIsRegisteredFull(t *testing.T) {
	for _, r := range midB12Rows() {
		for i, f := range []game.Face{r.front, r.back} {
			oracle := r.oracle
			if i == 1 {
				oracle += "#1"
			}
			spec, ok := Lookup(oracle)
			if !ok || spec.Name != f.Name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
				t.Errorf("%s: %+v %v", f.Name, spec, ok)
			}
		}
	}
}

// --- Mystic Skull ---------------------------------------------------------

func TestMidB12MysticSkullTransformsForFiveAndTapAndTheBackGrantsLandsAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	r := midB12Row_("Mystic Skull")
	forest := gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	id := r.put(g, me.ID, 0)
	if n := grantedManaRowsOf(g, forest); n != 0 {
		t.Fatalf("the front face granted %d rows", n)
	}
	if err := midB12Activate(t, g, me, id, 0, "{C}{C}{C}{C}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("four mana paid for a {5} transform")
	}
	if err := midB12Activate(t, g, me, id, 0, "{C}{C}{C}{C}{C}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := midB12Name(g, id); got != "Mystic Monstrosity" {
		t.Fatalf("face = %q", got)
	}
	if n := grantedManaRowsOf(g, forest); n != 1 {
		t.Errorf("the Forest has %d granted rows, want 1", n)
	}
	if n := grantedManaRowsOf(g, id); n != 0 {
		t.Errorf("the Monstrosity is not a land but has %d granted rows", n)
	}
}

func TestMidB12MysticSkullTapsForAnyColorForOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := midB12Row_("Mystic Skull").put(g, me.ID, 0)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	dcMana(t, g, me.ID, "{C}")
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, ch := range g.PendingChoices {
		if ch != nil && ch.Kind == game.PendingChoiceMana && ch.Chooser == me.ID {
			if err := g.ResolveManaChoice(ch.ID, me.ID, "U"); err != nil {
				t.Fatalf("choose: %v", err)
			}
		}
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want [U] (the {C} paid the {1})", got)
	}
}

// --- Mysterious Tome ------------------------------------------------------

func TestMidB12MysteriousTomeDrawsThenTransformsAndChroniclePaysToTap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	r := midB12Row_("Mysterious Tome")
	ssCard(me.Library, me.ID, "Top", "Sorcery", 0, 0)
	id := r.put(g, me.ID, 0)
	hand := me.Hand.Size()
	if err := midB12Activate(t, g, me, id, 0, "{C}{C}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
	if got := midB12Name(g, id); got != "Chilling Chronicle" {
		t.Fatalf("face = %q", got)
	}
	// Untap it for the second activation: tap target nonland permanent.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	land := gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	tgt := func(x uuid.UUID) game.ActivateAbilityParams {
		return game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: x}}}
	}
	if err := midB12Activate(t, g, me, id, 0, "{C}", tgt(land)); err == nil {
		t.Fatal("a land is not a legal target")
	}
	if err := midB12Activate(t, g, me, id, 0, "{C}", tgt(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if c, _ := battlefieldCard(g, bear); !c.Tapped {
		t.Error("the Bear was not tapped")
	}
	if got := midB12Name(g, id); got != "Mysterious Tome" {
		t.Errorf("face = %q, want it turned back over", got)
	}
}

// --- Ambitious Farmhand ---------------------------------------------------

func TestMidB12AmbitiousFarmhandFetchesABasicPlainsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Sacred Foundry", "Land — Mountain Plains"))
	r := midB12Row_("Ambitious Farmhand")
	c := r.card(me.ID, 0)
	c.Power, c.Toughness = 1, 1
	me.Hand.PushTop(c)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	ch := searchChoiceFor(g, me.ID)
	if ch == nil {
		t.Fatal("no search prompt")
	}
	for _, bad := range []string{"Forest", "Sacred Foundry"} {
		if searchOptionNamed(g, ch, bad) != uuid.Nil {
			t.Errorf("the search offered %s", bad)
		}
	}
	hand := me.Hand.Size()
	answerSearchNamed(t, g, me.ID, "Plains")
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
}

func TestMidB12AmbitiousFarmhandCovenGatesTheTransform(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := midB12Row_("Ambitious Farmhand").put(g, me.ID, 0)
	b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	if err := midB12Activate(t, g, me, id, 0, "{W}{W}{W}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("two different powers activated coven")
	}
	b12Creature(g, me.ID, "Another Two", "Creature — Bear", 2, 2)
	if err := midB12Activate(t, g, me, id, 0, "{W}{W}{W}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("two creatures sharing a power still make only two different powers")
	}
	b12Creature(g, me.ID, "Three", "Creature — Bear", 3, 3)
	if err := midB12Activate(t, g, me, id, 0, "{W}{W}{W}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with powers 1, 2, 3: %v", err)
	}
	if got := midB12Name(g, id); got != "Seasoned Cathar" {
		t.Errorf("face = %q", got)
	}
}

// --- Bereaved Survivor ----------------------------------------------------

func TestMidB12BereavedSurvivorTransformsWhenAnotherCreatureOfYoursDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := midB12Row_("Bereaved Survivor").put(g, me.ID, 0)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	usgB3Destroy(t, g, theirs)
	if got := midB12Name(g, id); got != "Bereaved Survivor" {
		t.Fatalf("an opponent's creature dying turned it over: %q", got)
	}
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	usgB3Destroy(t, g, mine)
	if got := midB12Name(g, id); got != "Dauntless Avenger" {
		t.Errorf("face = %q, want Dauntless Avenger", got)
	}
}

func TestMidB12BereavedSurvivorDyingItselfTransformsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := midB12Row_("Bereaved Survivor").put(g, me.ID, 0)
	usgB3Destroy(t, g, id)
	if got := midB12Name(g, id); got != "(gone)" {
		t.Errorf("got %q", got)
	}
}

func TestMidB12DauntlessAvengerReturnsASmallCreatureTappedAndAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := midB12Row_("Bereaved Survivor").put(g, me.ID, 1)
	small := ssCard(me.Graveyard, me.ID, "Small", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == small {
				me.Graveyard.Cards[i].ManaCost = "{1}{G}"
			}
		}
	})
	big := ssCard(me.Graveyard, me.ID, "Big", "Creature — Bear", 4, 4)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == big {
				me.Graveyard.Cards[i].ManaCost = "{2}{G}"
			}
		}
	})
	declareAttack(t, g, opp.ID, id)
	if latestPickTarget(g, me.ID) != nil {
		answerTriggerTargets(t, g, me.ID, small)
	}
	passPriorityAroundTable(t, g)
	if pick := latestOptionPickFor(g, me.ID); pick != nil {
		answerOptionPick(t, g, me.ID, 0)
	}
	c, ok := battlefieldCard(g, small)
	if !ok {
		t.Fatal("the small creature did not return")
	}
	if !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("tapped=%v attacking=%v", c.Tapped, c.AttackingTarget)
	}
	if !me.Graveyard.Contains(big) {
		t.Error("the mana value 3 creature left the graveyard")
	}
}

// --- Ecstatic Awakener ----------------------------------------------------

func TestMidB12EcstaticAwakenerSacrificesAnotherToDrawAndTransform(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := midB12Row_("Ecstatic Awakener").put(g, me.ID, 0)
	ssCard(me.Library, me.ID, "Top", "Sorcery", 0, 0)
	if err := midB12Activate(t, g, me, id, 0, "{B}{B}{B}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("it can sacrifice only ANOTHER creature")
	}
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	hand := me.Hand.Size()
	params := game.ActivateAbilityParams{}
	params.SacrificeIDs = []uuid.UUID{fodder}
	if err := midB12Activate(t, g, me, id, 0, "{B}{B}{B}", params); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, me.Hand.Size())
	}
	if got := midB12Name(g, id); got != "Awoken Demon" {
		t.Errorf("face = %q", got)
	}
	if _, ok := battlefieldCard(g, fodder); ok {
		t.Error("the sacrificed creature is still on the battlefield")
	}
}

// --- Heirloom Mirror // Inherited Fiend -----------------------------------

func midB12Mirror(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID) error {
	t.Helper()
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	var discard uuid.UUID
	if len(me.Hand.Cards) > 0 {
		discard = me.Hand.Cards[0].InstanceID
	}
	return midB12Activate(t, g, me, id, 0, "{C}", game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{discard}})
}

func TestMidB12HeirloomMirrorLoopsThenTransformsOnTheThirdRitualCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 4; i++ {
		ssCard(me.Library, me.ID, "Lib", "Sorcery", 0, 0)
		ssCard(me.Hand, me.ID, "Discard fodder", "Sorcery", 0, 0)
	}
	id := midB12Row_("Heirloom Mirror").put(g, me.ID, 0)
	life := lifeOf(g, me.ID)
	for i := 1; i <= 2; i++ {
		if err := midB12Mirror(t, g, me, id); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		if n := counterCount(g, id, "ritual"); n != i {
			t.Fatalf("ritual counters after %d = %d", i, n)
		}
		if got := midB12Name(g, id); got != "Heirloom Mirror" {
			t.Fatalf("turned over early: %q", got)
		}
	}
	if err := midB12Mirror(t, g, me, id); err != nil {
		t.Fatalf("third: %v", err)
	}
	if got := midB12Name(g, id); got != "Inherited Fiend" {
		t.Fatalf("face = %q", got)
	}
	if c, _ := battlefieldCard(g, id); c.Counters["ritual"] != 0 {
		t.Errorf("ritual counters left: %d", c.Counters["ritual"])
	}
	if got := lifeOf(g, me.ID); got != life-3 {
		t.Errorf("life %d -> %d, want -3", life, got)
	}
	if me.Graveyard.Size() < 6 {
		t.Errorf("graveyard %d, want at least 3 discards and 3 mills", me.Graveyard.Size())
	}
}

func TestMidB12HeirloomMirrorNeedsADiscardAndSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	ssCard(me.Library, me.ID, "Lib", "Sorcery", 0, 0)
	id := midB12Row_("Heirloom Mirror").put(g, me.ID, 0)
	if err := midB12Activate(t, g, me, id, 0, "{C}", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with no card to discard")
	}
	ssCard(me.Hand, me.ID, "Fodder", "Sorcery", 0, 0)
	advanceTo(t, g, game.StepBeginCombat)
	dcMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated outside a main phase")
	}
}

func TestMidB12InheritedFiendExilesACreatureCardAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := midB12Row_("Heirloom Mirror").put(g, me.ID, 1)
	bear := ssCard(opp.Graveyard, opp.ID, "Bear", "Creature — Bear", 2, 2)
	spell := ssCard(opp.Graveyard, opp.ID, "Bolt", "Instant", 0, 0)
	tgt := func(x uuid.UUID) game.ActivateAbilityParams {
		return game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: x}}}
	}
	if err := midB12Activate(t, g, me, id, 0, "{B}{B}{B}", tgt(spell)); err == nil {
		t.Fatal("an instant card is not a legal target")
	}
	if err := midB12Activate(t, g, me, id, 0, "{B}{B}{B}", tgt(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if opp.Graveyard.Contains(bear) {
		t.Error("the Bear is still in the graveyard")
	}
	if n := counterCount(g, id, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters = %d", n)
	}
}

// --- Deathbonnet Sprout // Hulk -------------------------------------------

// midB12NextUpkeep steps off the current step and on to seat 0's next upkeep.
func midB12NextUpkeep(t *testing.T, g *game.Game) {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
}

func TestMidB12DeathbonnetSproutMillsThenTransformsAtThreeCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	r := midB12Row_("Deathbonnet Sprout")
	id := r.put(g, me.ID, 0)
	ssCard(me.Graveyard, me.ID, "G1", "Creature — Bear", 1, 1)
	ssCard(me.Graveyard, me.ID, "G2", "Creature — Bear", 1, 1)
	ssCard(me.Library, me.ID, "Milled", "Sorcery", 0, 0)
	midB12NextUpkeep(t, g)
	if got := midB12Name(g, id); got != "Deathbonnet Sprout" {
		t.Fatalf("two creature cards turned it over: %q", got)
	}
	ssCard(me.Graveyard, me.ID, "G3", "Creature — Bear", 1, 1)
	ssCard(me.Library, me.ID, "Milled 2", "Sorcery", 0, 0)
	midB12NextUpkeep(t, g)
	if got := midB12Name(g, id); got != "Deathbonnet Hulk" {
		t.Errorf("face = %q", got)
	}
}

func TestMidB12DeathbonnetSproutCountsTheCardItJustMilled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := midB12Row_("Deathbonnet Sprout").put(g, me.ID, 0)
	ssCard(me.Graveyard, me.ID, "G1", "Creature — Bear", 1, 1)
	ssCard(me.Graveyard, me.ID, "G2", "Creature — Bear", 1, 1)
	ssCard(me.Library, me.ID, "Milled", "Creature — Bear", 1, 1)
	midB12NextUpkeep(t, g)
	if got := midB12Name(g, id); got != "Deathbonnet Hulk" {
		t.Errorf("the milled creature card is the third: %q", got)
	}
}

func TestMidB12DeathbonnetHulkExilesACardAndGrowsOnlyForACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := midB12Row_("Deathbonnet Sprout").put(g, me.ID, 1)
	spell := ssCard(opp.Graveyard, opp.ID, "Bolt", "Instant", 0, 0)
	midB12NextUpkeep(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no graveyard pick")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(spell) {
		t.Error("the instant was not exiled")
	}
	if n := counterCount(g, id, game.CounterPlusOne); n != 0 {
		t.Errorf("a noncreature card earned %d counters", n)
	}

	bear := ssCard(opp.Graveyard, opp.ID, "Bear", "Creature — Bear", 2, 2)
	midB12NextUpkeep(t, g)
	pick = latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no graveyard pick")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := counterCount(g, id, game.CounterPlusOne); n != 1 {
		t.Errorf("a creature card earned %d counters, want 1", n)
	}
}

// --- Smoldering Egg // Ashmouth Dragon ------------------------------------

func midB12Cast(t *testing.T, g *game.Game, me *game.Player, typeLine, cost string, n int, params game.CastSpellParams) {
	t.Helper()
	advanceToMain(t, g)
	if err := socB19CastWithPool(g, me, game.Card{Name: "Spell", TypeLine: typeLine, ManaCost: cost}, "R", n, params); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestMidB12SmolderingEggCountsTheManaSpentAndTransformsAtSeven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := midB12Row_("Smoldering Egg").put(g, me.ID, 0)
	midB12Cast(t, g, me, "Sorcery", "{2}{R}", 3, game.CastSpellParams{})
	if n := counterCount(g, id, "ember"); n != 3 {
		t.Fatalf("ember counters = %d, want 3", n)
	}
	// A creature spell counts nothing.
	midB12Cast(t, g, me, "Creature — Bear", "{1}{R}", 2, game.CastSpellParams{})
	if n := counterCount(g, id, "ember"); n != 3 {
		t.Fatalf("a creature spell moved the counters: %d", n)
	}
	midB12Cast(t, g, me, "Instant", "{3}{R}", 4, game.CastSpellParams{})
	if got := midB12Name(g, id); got != "Ashmouth Dragon" {
		t.Fatalf("face = %q", got)
	}
	if c, _ := battlefieldCard(g, id); c.Counters["ember"] != 0 {
		t.Errorf("ember counters left: %d", c.Counters["ember"])
	}
}

func TestMidB12SmolderingEggStaysShortOfSeven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := midB12Row_("Smoldering Egg").put(g, me.ID, 0)
	midB12Cast(t, g, me, "Instant", "{3}{R}", 4, game.CastSpellParams{})
	midB12Cast(t, g, me, "Instant", "{1}{R}", 2, game.CastSpellParams{})
	if got := midB12Name(g, id); got != "Smoldering Egg" {
		t.Errorf("six counters turned it over: %q", got)
	}
	if n := counterCount(g, id, "ember"); n != 6 {
		t.Errorf("ember counters = %d, want 6", n)
	}
}

func TestMidB12AshmouthDragonDealsTwoWhenYouCastAnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	midB12Row_("Smoldering Egg").put(g, me.ID, 1)
	before := lifeOf(g, opp.ID)
	advanceToMain(t, g)
	if err := socB19CastWithPool(g, me, game.Card{Name: "Spell", TypeLine: "Instant", ManaCost: "{R}"}, "R", 1, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	answerTriggerTargets2(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != before-2 {
		t.Errorf("life %d -> %d, want -2", before, got)
	}
}
