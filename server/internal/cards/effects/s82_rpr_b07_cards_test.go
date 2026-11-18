package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b07_cards_test.go — S82 reprint pass, batch 7 (#5908): the
// Magic Commander Collection reprints (Knights, Golems, Faeries and two
// seams: Haakon's graveyard-only cast and Ion Storm's two-kind counter cost).

const (
	rprB07HaakonOracle     = "8f089280-e5da-4fda-8bff-dea2971d0292"
	rprB07IgniterOracle    = "12aba1d4-477f-4d36-96bb-c3c7bbd51b75"
	rprB07AllianceOracle   = "b4ca6a20-b1ba-4dee-ae8a-83f8edf360c4"
	rprB07IonStormOracle   = "448e0cf2-c97c-4dc0-96ef-91f1eb75d97c"
	rprB07JuniperOracle    = "18ca77c9-e7d2-4afc-a72b-98772578fad7"
	rprB07KeskitOracle     = "e665acb1-a839-46c6-b5ae-ad189213bee4"
	rprB07ExemplarOracle   = "ecb198fc-5663-4380-b2da-ad54421e966d"
	rprB07LastBreathOracle = "67791323-bf21-4c72-bf57-32197d3d5775"
	rprB07ChargeOracle     = "b4014b7c-aa1d-4feb-ab7a-4921ad150f9a"
	rprB07BearerOracle     = "7b600c0e-5ad2-48a2-a147-346ca78c8629"
	rprB07SplicerOracle    = "9e0c5919-aed6-4d98-85c8-2658be78bdf0"
)

func rprB07Pay(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func rprB07Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return blbB1Push(g, owner, name, typeLine, oracle, power, toughness)
}

func TestRprB07EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rprB07HaakonOracle:     "Haakon, Stromgald Scourge",
		rprB07IgniterOracle:    "Hellkite Igniter",
		rprB07AllianceOracle:   "Improbable Alliance",
		rprB07IonStormOracle:   "Ion Storm",
		rprB07JuniperOracle:    "Juniper Order Ranger",
		rprB07KeskitOracle:     "Keskit, the Flesh Sculptor",
		rprB07ExemplarOracle:   "Knight Exemplar",
		rprB07LastBreathOracle: "Knight of the Last Breath",
		rprB07ChargeOracle:     "Knights' Charge",
		rprB07BearerOracle:     "Liliana's Standard Bearer",
		rprB07SplicerOracle:    "Master Splicer",
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

func TestRprB07HellkiteIgniterPumpsByTheArtifactsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	igniter := rprB07Push(g, me.ID, "Hellkite Igniter", "Creature — Dragon", rprB07IgniterOracle, 5, 5)
	for i := 0; i < 2; i++ {
		rprB07Push(g, me.ID, "My Trinket", "Artifact", "", 0, 0)
	}
	rprB07Push(g, opp.ID, "Their Trinket", "Artifact", "", 0, 0)
	advanceToMain(t, g)
	rprB07Pay(me, "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, igniter, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, igniter)
	if got := c.CurrentPower(); got != 7 {
		t.Errorf("power %d, want 5 + the two artifacts you control", got)
	}
	if got := c.CurrentToughness(); got != 5 {
		t.Errorf("toughness %d, want 5", got)
	}
}

func TestRprB07MasterSplicerMakesAGolemAndLordsGolems(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	theirGolem := rprB07Push(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "", 2, 2)
	splicer := castHoldCreature(t, g, "Master Splicer", "Creature — Phyrexian Human Artificer", rprB07SplicerOracle, 1, 1)
	passPriorityAroundTable(t, g)
	if n := battlefieldNamed(g, me.ID, "Phyrexian Golem"); n != 1 {
		t.Fatalf("Golem tokens %d, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Golem" {
			if got := c.CurrentPower(); got != 4 {
				t.Errorf("token power %d, want 3+1", got)
			}
		}
	}
	if c, _ := battlefieldCard(g, splicer); c.CurrentPower() != 1 {
		t.Errorf("Splicer power %d, want 1 (it is no Golem)", c.CurrentPower())
	}
	if c, _ := battlefieldCard(g, theirGolem); c.CurrentPower() != 2 {
		t.Errorf("an opponent's Golem power %d, want 2 (\"you control\")", c.CurrentPower())
	}
}

func TestRprB07KnightExemplarLordsOnlyYourOtherKnights(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	exemplar := rprB07Push(g, me.ID, "Knight Exemplar", "Creature — Human Knight", rprB07ExemplarOracle, 2, 2)
	mine := rprB07Push(g, me.ID, "My Knight", "Creature — Human Knight", "", 2, 2)
	theirs := rprB07Push(g, opp.ID, "Their Knight", "Creature — Human Knight", "", 2, 2)
	bear := rprB07Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	for id, want := range map[uuid.UUID]int{exemplar: 2, mine: 3, theirs: 2, bear: 2} {
		c, _ := battlefieldCard(g, id)
		if c.CurrentPower() != want {
			t.Errorf("%s power %d, want %d", c.Name, c.CurrentPower(), want)
		}
		wantInd := id == mine
		if got := hasEffectiveKeyword(t, g, id, "indestructible"); got != wantInd {
			t.Errorf("%s indestructible = %v, want %v", c.Name, got, wantInd)
		}
	}
	if !hasEffectiveKeyword(t, g, exemplar, "first strike") {
		t.Error("the Exemplar lost its own first strike")
	}
}

func TestRprB07JuniperOrderRangerPutsACounterOnBothCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	ranger := rprB07Push(g, me.ID, "Juniper Order Ranger", "Creature — Human Knight Ranger", rprB07JuniperOracle, 2, 4)
	// Another creature entering under an opponent's control does nothing.
	rprB07Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, ranger, game.CounterPlusOne); got != 0 {
		t.Fatalf("Ranger counters %d after an opponent's creature, want 0", got)
	}
	bear := castHoldCreature(t, g, "My Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("entering creature counters %d, want 1", got)
	}
	if got := counterCount(g, ranger, game.CounterPlusOne); got != 1 {
		t.Errorf("Ranger counters %d, want 1", got)
	}
}

func TestRprB07KnightsChargeDrainsPerAttackingKnightAndRebuildsTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	charge := rprB07Push(g, me.ID, "Knights' Charge", "Enchantment", rprB07ChargeOracle, 0, 0)
	k1 := rprB07Push(g, me.ID, "Knight One", "Creature — Human Knight", "", 2, 2)
	k2 := rprB07Push(g, me.ID, "Knight Two", "Creature — Human Knight", "", 2, 2)
	bear := rprB07Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	myLife, oppLife := lifeOf(g, me.ID), lifeOf(g, opp.ID)
	declareAttack(t, g, opp.ID, k1, k2, bear)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != oppLife-2 {
		t.Errorf("opponent life %d, want %d (one per Knight, none for the Bear)", got, oppLife-2)
	}
	if got := lifeOf(g, me.ID); got != myLife+2 {
		t.Errorf("your life %d, want %d", got, myLife+2)
	}

	// The activated ability: sacrifice, return every Knight creature card.
	gk := pushGraveyardPermanent(me, "Dead Knight", "Creature — Human Knight", "{1}{W}")
	gb := pushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	advanceTo(t, g, game.StepPostcombatMain)
	rprB07Pay(me, "W", "B", "C", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, charge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, charge); ok {
		t.Error("the Charge was not sacrificed")
	}
	if _, ok := battlefieldCard(g, gk); !ok {
		t.Error("the Knight card did not return")
	}
	if _, ok := battlefieldCard(g, gb); ok {
		t.Error("a non-Knight returned")
	}
}

func TestRprB07ImprobableAllianceMakesAFaerieOnTheSecondDrawOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	rprB07Push(g, me.ID, "Improbable Alliance", "Enchantment", rprB07AllianceOracle, 0, 0)
	advanceToMain(t, g)
	// The harness starts after this turn's draw-step draw, so the next draw
	// is the second card drawn this turn.
	if got := g.TurnTallyFor(me.ID).CardsDrawn; got != 1 {
		t.Fatalf("cards drawn so far %d, want 1", got)
	}
	g.DrawCard(me.ID)
	passPriorityAroundTable(t, g)
	if n := battlefieldNamed(g, me.ID, "Faerie"); n != 1 {
		t.Fatalf("Faeries %d after the second draw, want 1", n)
	}
	g.DrawCard(me.ID)
	passPriorityAroundTable(t, g)
	if n := battlefieldNamed(g, me.ID, "Faerie"); n != 1 {
		t.Errorf("Faeries %d after the third draw, want still 1", n)
	}
}

func TestRprB07ImprobableAllianceLoots(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	alliance := rprB07Push(g, me.ID, "Improbable Alliance", "Enchantment", rprB07AllianceOracle, 0, 0)
	advanceToMain(t, g)
	rprB07Pay(me, "U", "R", "C", "C", "C", "C")
	lib := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, alliance, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool %d, want the {4}{U}{R} paid", me.ManaPool.Total())
	}
	if got := lib - me.Library.Size(); got != 1 {
		t.Errorf("library shrank by %d, want the one card drawn", got)
	}
}

func TestRprB07KnightOfTheLastBreathSacrificesNontokenAndLeavesThreeSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	knight := rprB07Push(g, me.ID, "Knight of the Last Breath", "Creature — Giant Knight", rprB07LastBreathOracle, 4, 4)
	fodder := rprB07Push(g, me.ID, "Fodder", "Creature — Bear", "", 2, 2)
	token := rprB07Push(g, me.ID, "Spirit Token", "Token Creature — Spirit", "", 1, 1)
	advanceToMain(t, g)
	rprB07Pay(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{token}}); err == nil {
		t.Fatal("a token was accepted for \"another nontoken creature\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{knight}}); err == nil {
		t.Fatal("the Knight sacrificed itself for \"another\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := battlefieldNamed(g, me.ID, "Spirit"); n != 1 {
		t.Fatalf("Spirits %d, want 1", n)
	}
	killCreature(t, g, me.ID, knight)
	passPriorityAroundTable(t, g)
	if n := battlefieldNamed(g, me.ID, "Spirit"); n != 4 {
		t.Errorf("Spirits %d after afterlife 3, want 4", n)
	}
}

func TestRprB07LilianasStandardBearerDrawsForCreaturesThatDiedThisTurn(t *testing.T) {
	for _, died := range []int{0, 2} {
		g := newCatalogGame(t)
		me, opp, _ := blbB1Seats(g)
		for i := 0; i < died; i++ {
			c := rprB07Push(g, me.ID, "Casualty", "Creature — Bear", "", 1, 1)
			killCreature(t, g, me.ID, c)
		}
		// An opponent's creature dying is not "under your control".
		o := rprB07Push(g, opp.ID, "Theirs", "Creature — Bear", "", 1, 1)
		killCreature(t, g, opp.ID, o)
		passPriorityAroundTable(t, g)
		advanceToMain(t, g)
		before := handSize(me)
		castHoldCreature(t, g, "Liliana's Standard Bearer", "Creature — Zombie Knight", rprB07BearerOracle, 3, 1)
		passPriorityAroundTable(t, g)
		if got := handSize(me) - before; got != died {
			t.Errorf("died=%d: drew %d, want %d", died, got, died)
		}
	}
}

func TestRprB07KeskitSacrificesThreeToTakeTwoOfTheTopThree(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	keskit := rprB07Push(g, me.ID, "Keskit, the Flesh Sculptor", "Legendary Creature — Phyrexian Human Artificer", rprB07KeskitOracle, 1, 3)
	a := rprB07Push(g, me.ID, "Fodder A", "Creature — Bear", "", 1, 1)
	b := rprB07Push(g, me.ID, "Fodder B", "Artifact", "", 0, 0)
	c := rprB07Push(g, me.ID, "Fodder C", "Creature — Bear", "", 1, 1)
	d := rprB07Push(g, me.ID, "Spare", "Creature — Bear", "", 1, 1)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, keskit, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, b}}); err == nil {
		t.Fatal("two sacrifices were accepted for three")
	}
	if err := g.ActivateCatalogAbility(me.ID, keskit, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, b, keskit}}); err == nil {
		t.Fatal("Keskit sacrificed itself for \"other\"")
	}
	var top []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		top = append(top, id)
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Top", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
	}
	hand, gy := handSize(me), me.Graveyard.Size()
	if err := g.ActivateCatalogAbility(me.ID, keskit, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, b, c}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if _, ok := battlefieldCard(g, id); ok {
			t.Errorf("%s was not sacrificed", id)
		}
	}
	if _, ok := battlefieldCard(g, d); !ok {
		t.Error("the spare was sacrificed")
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt")
	}
	if pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Errorf("pick bounds %d..%d, want exactly 2", pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[0], top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := handSize(me) - hand; got != 2 {
		t.Errorf("hand +%d, want +2", got)
	}
	// Three sacrificed permanents and the card put into the graveyard.
	if got := me.Graveyard.Size() - gy; got != 4 {
		t.Errorf("graveyard +%d, want +4 (three sacrificed, one put)", got)
	}
}

func TestRprB07IonStormRemovesOnlyAPlusOneOrChargeCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	storm := rprB07Push(g, me.ID, "Ion Storm", "Enchantment", rprB07IonStormOracle, 0, 0)
	charged := rprB07Push(g, me.ID, "Charged", "Artifact", "", 0, 0)
	other := rprB07Push(g, me.ID, "Other", "Artifact", "", 0, 0)
	setCounters(g, charged, map[string]int{"charge": 1})
	setCounters(g, other, map[string]int{"stun": 1})
	advanceToMain(t, g)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	rprB07Pay(me, "R", "C", "R", "C", "R", "C")
	err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{
		Targets: target, CounterSourceIDs: []uuid.UUID{other}, CounterKind: "stun"})
	if err == nil {
		t.Fatal("a stun counter paid for Ion Storm")
	}
	oppLife := lifeOf(g, opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{
		Targets: target, CounterSourceIDs: []uuid.UUID{charged}, CounterKind: "charge"}); err != nil {
		t.Fatalf("charge counter refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, charged, "charge"); got != 0 {
		t.Errorf("charge counters %d, want 0", got)
	}
	if got := lifeOf(g, opp.ID); got != oppLife-2 {
		t.Errorf("opponent life %d, want %d", got, oppLife-2)
	}
	// A +1/+1 counter pays it as well.
	bear := rprB07Push(g, me.ID, "Grown", "Creature — Bear", "", 1, 1)
	setCounters(g, bear, map[string]int{game.CounterPlusOne: 1})
	if err := g.ActivateCatalogAbility(me.ID, storm, 0, game.ActivateAbilityParams{
		Targets: target, CounterSourceIDs: []uuid.UUID{bear}, CounterKind: game.CounterPlusOne}); err != nil {
		t.Fatalf("+1/+1 counter refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != oppLife-4 {
		t.Errorf("opponent life %d, want %d", got, oppLife-4)
	}
}

func rprB07PushHaakonInGraveyard(me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Graveyard.Cards = append(me.Graveyard.Cards, game.Card{
		InstanceID: id, Name: "Haakon, Stromgald Scourge", TypeLine: "Legendary Creature — Zombie Knight",
		ManaCost: "{1}{B}{B}", OracleID: rprB07HaakonOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	return id
}

func TestRprB07HaakonCastsOnlyFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	advanceToMain(t, g)
	hand := pushHandCardWithManaCost(me, "Haakon, Stromgald Scourge", "Legendary Creature — Zombie Knight", "{1}{B}{B}")
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == hand {
			me.Hand.Cards[i].OracleID = rprB07HaakonOracle
		}
	}
	rprB07Pay(me, "B", "B", "C")
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{Strict: true}); !errors.Is(err, game.ErrCastZoneNotAllowed) {
		t.Fatalf("hand cast: err = %v, want ErrCastZoneNotAllowed", err)
	}
	gy := rprB07PushHaakonInGraveyard(me)
	if err := g.CastSpell(me.ID, gy, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("graveyard cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, gy); !ok {
		t.Error("Haakon did not resolve onto the battlefield")
	}
}

func TestRprB07HaakonOpensKnightSpellsInTheGraveyardWhileOnTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	advanceToMain(t, g)
	knight := pushGraveyardPermanent(me, "Dead Knight", "Creature — Human Knight", "{1}{W}")
	bear := pushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	rprB07Pay(me, "W", "C", "G", "C")
	if err := g.CastSpell(me.ID, knight, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("a Knight was cast from the graveyard with no Haakon on the battlefield")
	}
	haakon := rprB07Push(g, me.ID, "Haakon, Stromgald Scourge", "Legendary Creature — Zombie Knight", rprB07HaakonOracle, 3, 3)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("a non-Knight was cast from the graveyard")
	}
	if err := g.CastSpell(me.ID, knight, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("Knight from the graveyard under Haakon: %v", err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, knight); !ok {
		t.Error("the Knight did not resolve")
	}
	life := lifeOf(g, me.ID)
	killCreature(t, g, me.ID, haakon)
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, me.ID); got != life-2 {
		t.Errorf("life %d after Haakon died, want %d", got, life-2)
	}
}
