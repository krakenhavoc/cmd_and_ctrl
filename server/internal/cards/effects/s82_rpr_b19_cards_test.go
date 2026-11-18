package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b19_cards_test.go — S82 reprint pass, batch 19 (#5908). Each
// test plays the real catalog entry and asserts what its text says.

const (
	rprB19SearchOracle    = "9cdc9f99-c6fa-40cd-90b0-e47d43a8cc3c"
	rprB19TajicOracle     = "8951d797-6356-4f27-a6de-e013216274d6"
	rprB19TenzaOracle     = "f8180ed7-84e9-49f7-8256-dcc21f9d334b"
	rprB19CircleOracle    = "fd1b4523-2399-4771-89bf-1d65f5d67056"
	rprB19CourierOracle   = "86f94f8e-f45e-42fc-8ad6-e9bdfaaa997d"
	rprB19SanctumOracle   = "38c83332-2d21-450c-ae97-34109b565f59"
	rprB19XyrisOracle     = "8687948c-456b-495b-9b31-f818c65624b6"
	rprB19ZaxaraOracle    = "a9946ea0-bf0e-4c9c-a3c6-21337a5029ab"
	rprB19VandalOracle    = "b872f147-00eb-4cf8-af8c-4144356b2089"
	rprB19KillShotOracle  = "e1ce32c7-a882-4d3a-9fd6-e5e2943382fb"
	rprB19SwindlerOracle  = "81674138-08d0-4126-9e6d-27f6e7dd2403"
	rprB19LegendaryHuman  = "Legendary Creature — Human Soldier"
	rprB19KnightTokenName = "Knight"
)

func TestRprB19EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		rprB19SearchOracle: "Search for Tomorrow", rprB19TajicOracle: "Tajic, Blade of the Legion",
		rprB19TenzaOracle: "Tenza, Godo's Maul", rprB19CircleOracle: "The Circle of Loyalty",
		rprB19CourierOracle: "Transguild Courier", rprB19SanctumOracle: "Tyrite Sanctum",
		rprB19XyrisOracle: "Xyris, the Writhing Storm", rprB19ZaxaraOracle: "Zaxara, the Exemplary",
		rprB19VandalOracle: "Faerie Vandal", rprB19KillShotOracle: "Kill Shot",
		rprB19SwindlerOracle: "Tavern Swindler",
	}
	for oracle, name := range want {
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

// rprB19Cast puts a spell in `who`'s hand, floats `mana` and casts it under
// the strict mana gate.
func rprB19Cast(t *testing.T, g *game.Game, who *game.Player, name, typeLine, cost, oracle, mana string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	id := handCardFull(who, name, typeLine, cost, oracle, nil)
	if mana != "" {
		if err := g.AddManaForEffect(who.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	params.Strict = true
	return id, g.CastSpell(who.ID, id, params)
}

// --- Search for Tomorrow -----------------------------------------------

func TestRprB19SearchForTomorrowFetchesAnUntappedBasicAndCanBeSuspended(t *testing.T) {
	spec, _ := Lookup(rprB19SearchOracle)
	if len(spec.SpecialActions) != 1 || spec.SpecialActions[0].Kind != game.SpecialActionSuspend ||
		spec.SpecialActions[0].Counters != 2 || spec.SpecialActions[0].Cost != "{G}" {
		t.Errorf("special actions %+v, want suspend 2—{G}", spec.SpecialActions)
	}
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	forest := pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	pushLibraryCardForTest(me, game.Card{Name: "Wastes of Doom", TypeLine: "Land"})
	if _, err := rprB19Cast(t, g, me, "Search for Tomorrow", "Sorcery", "{2}{G}", rprB19SearchOracle, "{G}", game.CastSpellParams{}); err == nil {
		t.Fatal("cast with one mana under the strict gate")
	}
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	if _, err := rprB19Cast(t, g, me, "Search for Tomorrow", "Sorcery", "{2}{G}", rprB19SearchOracle, "{G}{G}{G}", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, forest)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the basic land did not reach the battlefield")
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == forest && c.Tapped {
			t.Error("the land entered tapped")
		}
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Wastes of Doom" {
			t.Error("a nonbasic land was fetched")
		}
	}
}

// --- Tajic, Blade of the Legion ----------------------------------------

func TestRprB19TajicGetsPlusFiveWithTwoOtherAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	tajic := b12Push(g, me.ID, "Tajic, Blade of the Legion", rprB19LegendaryHuman, rprB19TajicOracle, 2, 2)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, tajic, "indestructible") {
		t.Error("Tajic is not indestructible")
	}
	declareAttack(t, g, opp.ID, tajic, a, b)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, tajic); got != 7 {
		t.Errorf("Tajic has power %d, want 7", got)
	}
	if got := effectiveToughness(t, g, tajic); got != 7 {
		t.Errorf("Tajic has toughness %d, want 7", got)
	}
	if got := effectivePower(t, g, a); got != 2 {
		t.Errorf("Bear A has power %d, want 2", got)
	}
}

func TestRprB19TajicDoesNothingWithOnlyOneOtherAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	tajic := b12Push(g, me.ID, "Tajic, Blade of the Legion", rprB19LegendaryHuman, rprB19TajicOracle, 2, 2)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, tajic, a)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, tajic); got != 2 {
		t.Errorf("Tajic has power %d, want 2", got)
	}
}

// --- Tenza, Godo's Maul ------------------------------------------------

func TestRprB19TenzaPumpsAnyCreatureAndMoreWhenLegendaryAndTramplesWhenRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	maul := seedEquipment(g, me.ID, "Tenza, Godo's Maul", rprB19TenzaOracle)
	bear := seedBear(g, me.ID)
	equipTo(t, g, me.ID, maul, bear)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("plain Bear has power %d, want 3", got)
	}
	if hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("a colourless nonlegendary Bear has trample")
	}
	legend := seedLegendaryCreature(g, me.ID, "Legend")
	equipTo(t, g, me.ID, maul, legend)
	if got := effectivePower(t, g, legend); got != 5 {
		t.Errorf("legendary creature has power %d, want 5", got)
	}
	if hasEffectiveKeyword(t, g, legend, "trample") {
		t.Error("a colourless legendary creature has trample")
	}
	red := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Bear", TypeLine: "Creature — Bear", Colors: []string{"R"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	equipTo(t, g, me.ID, maul, red)
	if got := effectivePower(t, g, red); got != 3 {
		t.Errorf("red nonlegendary creature has power %d, want 3", got)
	}
	if !hasEffectiveKeyword(t, g, red, "trample") {
		t.Error("a red creature has no trample")
	}
	if got := effectivePower(t, g, legend); got != 2 {
		t.Errorf("the creature it left has power %d, want 2", got)
	}
}

// --- The Circle of Loyalty ---------------------------------------------

func TestRprB19CircleOfLoyaltyAnthemTokenAndAffinity(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	// Affinity for Knights: two Knights make {4}{W}{W} cost {2}{W}{W}.
	b12Creature(g, me.ID, "Knight A", "Creature — Human Knight", 2, 2)
	b12Creature(g, me.ID, "Knight B", "Creature — Human Knight", 2, 2)
	if _, err := rprB19Cast(t, g, me, "The Circle of Loyalty", "Legendary Artifact", "{4}{W}{W}", rprB19CircleOracle, "{W}{W}{W}", game.CastSpellParams{}); err == nil {
		t.Fatal("cast for three mana with two Knights, want it to cost four")
	}
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	circle, err := rprB19Cast(t, g, me, "The Circle of Loyalty", "Legendary Artifact", "{4}{W}{W}", rprB19CircleOracle, "{W}{W}{W}{W}", game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast for {2}{W}{W} with two Knights: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(circle) {
		t.Fatal("the Circle did not resolve")
	}
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("Bear has power %d, want 3 under the anthem", got)
	}
	// Casting a legendary spell makes a Knight; a nonlegendary one does not.
	before := len(g.Battlefield.Cards)
	if _, err := rprB19Cast(t, g, me, "Plain Spell", "Sorcery", "{0}", "", "", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast plain: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(g.Battlefield.Cards) != before {
		t.Error("a nonlegendary spell made a Knight")
	}
	if _, err := rprB19Cast(t, g, me, "Legend Spell", "Legendary Artifact", "{0}", "", "", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast legendary: %v", err)
	}
	passPriorityAroundTable(t, g)
	knight := battlefieldTokenNamed(t, g, rprB19KnightTokenName)
	if !hasEffectiveKeyword(t, g, knight.InstanceID, "vigilance") {
		t.Error("the Knight token has no vigilance")
	}
	if got := effectivePower(t, g, knight.InstanceID); got != 3 {
		t.Errorf("Knight token has power %d, want 3 (2/2 plus the anthem)", got)
	}
}

func TestRprB19CircleOfLoyaltyActivationMakesAKnight(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	circle := b12Push(g, me.ID, "The Circle of Loyalty", "Legendary Artifact", rprB19CircleOracle, 0, 0)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}{W}{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, circle, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	battlefieldTokenNamed(t, g, rprB19KnightTokenName)
	if err := g.ActivateCatalogAbility(me.ID, circle, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated again while tapped")
	}
}

// --- Transguild Courier ------------------------------------------------

func TestRprB19TransguildCourierIsAllColors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	courier := b12Push(g, me.ID, "Transguild Courier", "Artifact Creature — Golem", rprB19CourierOracle, 3, 3)
	var c game.Card
	g.ReadSnapshot(func() {
		for _, x := range g.Battlefield.Cards {
			if x.InstanceID == courier {
				c = x
			}
		}
	})
	for _, col := range []string{"W", "U", "B", "R", "G"} {
		if !c.HasColor(col) {
			t.Errorf("Transguild Courier is not %s", col)
		}
	}
}

// --- Tyrite Sanctum ----------------------------------------------------

func rprB19Sanctum(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	return g, me, b12Push(g, me.ID, "Tyrite Sanctum", "Land", rprB19SanctumOracle, 0, 0)
}

func TestRprB19TyriteSanctumMakesALegendaryCreatureAGodWithACounter(t *testing.T) {
	g, me, sanctum := rprB19Sanctum(t)
	legend := seedLegendaryCreature(g, me.ID, "Legend")
	bear := seedBear(g, me.ID)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("targeted a nonlegendary creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, sanctum, 0, game.ActivateAbilityParams{Targets: cardRefs(legend)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	var c game.Card
	g.ReadSnapshot(func() {
		for _, x := range g.Battlefield.Cards {
			if x.InstanceID == legend {
				c = x
			}
		}
	})
	if !c.HasSubtype("God") {
		t.Error("the creature did not become a God")
	}
	if !c.HasSubtype("Human") {
		t.Error("the creature lost its other types")
	}
	if got := counterOn(g, legend, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestRprB19TyriteSanctumSacrificesForAnIndestructibleCounterOnAGod(t *testing.T) {
	g, me, sanctum := rprB19Sanctum(t)
	god := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "A God", TypeLine: "Legendary Enchantment Creature — God",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	bear := seedBear(g, me.ID)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}{G}{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, sanctum, 1, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("targeted a non-God")
	}
	if err := g.ActivateCatalogAbility(me.ID, sanctum, 1, game.ActivateAbilityParams{Targets: cardRefs(god)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sanctum) {
		t.Error("the Sanctum was not sacrificed")
	}
	if got := counterOn(g, god, game.CounterIndestructible); got != 1 {
		t.Errorf("%d indestructible counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, god, "indestructible") {
		t.Error("the God is not indestructible")
	}
}

// --- Xyris, the Writhing Storm -----------------------------------------

func TestRprB19XyrisMakesASnakeForEachExtraOpponentDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Xyris, the Writhing Storm", "Legendary Creature — Snake Leviathan", rprB19XyrisOracle, 3, 5)
	count := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.IsToken() && c.Name == "Snake" {
				n++
			}
		}
		return n
	}
	// Outside their own draw step every card an opponent draws counts.
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := count(); got != 1 {
		t.Fatalf("%d Snakes after an opponent drew off-step, want 1", got)
	}
	// Your own draws never count.
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := count(); got != 1 {
		t.Errorf("%d Snakes after my own draw, want 1", got)
	}
}

func TestRprB19XyrisCombatDamageMakesBothPlayersDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	xyris := b12Push(g, me.ID, "Xyris, the Writhing Storm", "Legendary Creature — Snake Leviathan", rprB19XyrisOracle, 3, 5)
	myHand, oppHand := len(me.Hand.Cards), len(opp.Hand.Cards)
	declareAttack(t, g, opp.ID, xyris)
	advanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - myHand; got != 3 {
		t.Errorf("I drew %d cards, want 3", got)
	}
	if got := len(opp.Hand.Cards) - oppHand; got != 3 {
		t.Errorf("the damaged player drew %d cards, want 3", got)
	}
}

// --- Zaxara, the Exemplary ---------------------------------------------

func TestRprB19ZaxaraMakesAHydraWithXCountersOnlyForXSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Zaxara, the Exemplary", "Legendary Creature — Nightmare Hydra", rprB19ZaxaraOracle, 2, 3)
	hydras := func() []game.Card {
		var out []game.Card
		for _, c := range g.Battlefield.Cards {
			if c.IsToken() && c.Name == "Hydra" {
				out = append(out, c)
			}
		}
		return out
	}
	if _, err := rprB19Cast(t, g, me, "Plain Sorcery", "Sorcery", "{1}", "", "{G}", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast plain: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(hydras()) != 0 {
		t.Fatal("a spell without {X} made a Hydra")
	}
	if _, err := rprB19Cast(t, g, me, "X Sorcery", "Sorcery", "{X}{G}", "", "{G}{G}{G}{G}", game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast X spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	hs := hydras()
	if len(hs) != 1 {
		t.Fatalf("%d Hydras, want 1", len(hs))
	}
	if got := counterOn(g, hs[0].InstanceID, game.CounterPlusOne); got != 3 {
		t.Errorf("the Hydra has %d +1/+1 counters, want X = 3", got)
	}
	if got, _ := socB14PT(t, g, hs[0].InstanceID); got != 3 {
		t.Errorf("the Hydra has power %d, want 3", got)
	}
}

func TestRprB19ZaxaraTapsForTwoManaOfOneColor(t *testing.T) {
	spec, _ := Lookup(rprB19ZaxaraOracle)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != OneColorOfAmount(2) || !spec.ManaAbilities[0].Cost.Tap {
		t.Errorf("mana abilities %+v, want {T}: two mana of any one colour", spec.ManaAbilities)
	}
	if !hasKeywordInSpec(spec, "deathtouch") {
		t.Error("Zaxara has no deathtouch")
	}
}

func hasKeywordInSpec(spec Spec, kw string) bool {
	for _, k := range spec.PrintedKeywords {
		if k == kw {
			return true
		}
	}
	return false
}

// --- Faerie Vandal -----------------------------------------------------

func TestRprB19FaerieVandalGrowsOnTheSecondCardDrawnEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	vandal := b12Push(g, me.ID, "Faerie Vandal", "Creature — Faerie Rogue", rprB19VandalOracle, 1, 2)
	for _, kw := range []string{"flash", "flying"} {
		if !hasEffectiveKeyword(t, g, vandal, kw) {
			t.Errorf("Faerie Vandal has no %s", kw)
		}
	}
	draws := 0
	for g.TurnTallyFor(me.ID).CardsDrawn < 2 {
		g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
		passPriorityAroundTable(t, g)
		if draws++; draws > 4 {
			t.Fatal("never drew a second card")
		}
	}
	if got := counterOn(g, vandal, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after the second draw, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := counterOn(g, vandal, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after a third draw, want still 1", got)
	}
}

// --- Kill Shot ---------------------------------------------------------

func TestRprB19KillShotDestroysOnlyAnAttackingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	attacker := b12Creature(g, me.ID, "Attacking Ogre", "Creature — Ogre", 3, 3)
	idle := b12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, attacker)
	if _, err := rprB19Cast(t, g, opp, "Kill Shot", "Instant", "{2}{W}", rprB19KillShotOracle, "{W}{W}{W}", game.CastSpellParams{Targets: cardTarget(idle)}); err == nil {
		t.Fatal("a creature that is not attacking was accepted as the target")
	}
	g.WithWriteLock(func() { opp.ManaPool = game.ManaPool{} })
	if _, err := rprB19Cast(t, g, opp, "Kill Shot", "Instant", "{2}{W}", rprB19KillShotOracle, "{W}{W}{W}", game.CastSpellParams{Targets: cardTarget(attacker)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking creature was not destroyed")
	}
	if !g.Battlefield.Contains(idle) {
		t.Error("the idle creature left the battlefield")
	}
}

// --- Tavern Swindler ---------------------------------------------------

func TestRprB19TavernSwindlerPaysThreeLifeAndGainsSixOnAWonFlip(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	swindler := b12Push(g, me.ID, "Tavern Swindler", "Creature — Human Rogue", rprB19SwindlerOracle, 2, 2)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, swindler, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != life-3 {
		t.Fatalf("life %d after paying, want %d", me.Life, life-3)
	}
	passPriorityAroundTable(t, g)
	c := pr7bPrompt(t, g, game.PendingChoiceCoinCall)
	if err := g.ResolveCoinCall(c.ID, me.ID, "heads"); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	won := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventFlipCoin && ev.Actor == me.ID {
			won = ev.Won
		}
	}
	want := life - 3
	if won {
		want += 6
	}
	if me.Life != want {
		t.Errorf("life %d (flip won: %v), want %d", me.Life, won, want)
	}
	if err := g.ActivateCatalogAbility(me.ID, swindler, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated again while tapped")
	}
}
