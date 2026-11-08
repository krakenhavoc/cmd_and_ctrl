package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// otj_cast_from_hand_cards_test.go — "if you haven't cast a spell from
// your hand this turn" (#5235) and the Outlaws of Thunder Junction cards
// that read it (#4468). The tally cell itself is pinned in
// game/cast_tally_from_hand_test.go; these play each card's real catalog
// entry.

const (
	canyonCrabOracle         = "4681facf-21ca-4118-a827-0e3d4233d8e6"
	emergentHauntingOracle   = "9f4d0133-0ecc-4498-8a47-4906321faae2"
	inventiveWingsmithOracle = "96186b7e-3dd3-44e7-ad23-79638787004c"
	jemLightfooteOracle      = "5802aa98-a9b7-4b0d-9d7c-c3938701870d"
	prairieDogOracle         = "70987216-f84d-43b9-8d75-bb21ff37747b"
	wranglerOfTheDamned      = "0d8fded4-0393-43c2-82a1-383f29de65ce"
)

func TestCastFromHandCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		canyonCrabOracle:         "Canyon Crab",
		emergentHauntingOracle:   "Emergent Haunting",
		inventiveWingsmithOracle: "Inventive Wingsmith",
		jemLightfooteOracle:      "Jem Lightfoote, Sky Explorer",
		prairieDogOracle:         "Prairie Dog",
		wranglerOfTheDamned:      "Wrangler of the Damned",
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

// The rows whose steps declare a footprint start with the from-hand
// condition. Canyon Crab's loot is a hand-written Effect, and Emergent
// Haunting's "isn't a creature" declares nothing, so those two rows have
// no footprint and are never reordered.
func TestCastFromHandRowsDeclareTheirCondition(t *testing.T) {
	for _, oracle := range []string{inventiveWingsmithOracle, jemLightfooteOracle, prairieDogOracle, wranglerOfTheDamned} {
		found := false
		for _, row := range game.CatalogTriggers(oracle) {
			if len(row.Footprint) > 1 && row.Footprint[0].Kind == game.FootprintCondition &&
				row.Footprint[0].Cond == game.FootprintCondNoSpellsCastFromHand {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: no row's footprint starts with the from-hand condition", oracle)
		}
	}
}

// otjPush puts a catalog permanent on the battlefield under `owner`, not
// summoning-sick.
func otjPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tgh int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tgh, Owner: owner, Controller: owner,
	})
}

// otjCastFromHand casts a plain sorcery from the active player's hand at
// a main phase and resolves it.
func otjCastFromHand(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Hand Fixture", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
}

// otjCastFromGraveyard casts a plain sorcery from p's graveyard under a
// permission (a flashback-like cast) at a main phase and resolves it.
func otjCastFromGraveyard(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Graveyard Fixture", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
		if !g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{Player: p.ID, Zone: game.ZoneGraveyard}) {
			t.Fatal("setup: no permission over the graveyard card")
		}
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// --- the condition helper ---------------------------------------------------

// The readers agree: nothing cast, then a spell cast from a graveyard
// (still "haven't"), then one from the hand.
func TestCastFromHandReaders(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	src := game.Card{Controller: me.ID}
	read := func() (plain, when, cond bool) {
		g.WithWriteLock(func() {
			plain = CastASpellFromHandThisTurn(g, me.ID)
			when = YouHaventCastASpellFromHand(game.Event{}, &src, game.Characteristic{}, g)
			cond = CondNoSpellFromHand{}.Holds(&Context{Game: g, Item: &game.StackItem{Controller: me.ID}})
		})
		return
	}
	if p, w, c := read(); p || !w || !c {
		t.Fatalf("with nothing cast: cast %v, haven't %v, condition %v", p, w, c)
	}
	otjCastFromGraveyard(t, g, me)
	if p, w, c := read(); p || !w || !c {
		t.Errorf("after a graveyard cast: cast %v, haven't %v, condition %v; want not cast", p, w, c)
	}
	otjCastFromHand(t, g)
	if p, w, c := read(); !p || w || c {
		t.Errorf("after a hand cast: cast %v, haven't %v, condition %v; want cast", p, w, c)
	}
	step, ok := CondNoSpellFromHand{}.condition()
	if !ok || step.Cond != game.FootprintCondNoSpellsCastFromHand {
		t.Errorf("footprint %+v (%v), want %s", step, ok, game.FootprintCondNoSpellsCastFromHand)
	}
}

// --- Jem Lightfoote, Sky Explorer --------------------------------------------

func TestJemLightfooteDrawsWhenNothingWasCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	otjPush(g, me.ID, "Jem Lightfoote, Sky Explorer", "Legendary Creature — Human Scout", jemLightfooteOracle, 3, 3)
	otjCastFromGraveyard(t, g, me)
	hand := me.Hand.Size()
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn (a graveyard cast is not from the hand)", hand, got)
	}
}

func TestJemLightfooteDrawsNothingAfterAHandCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	otjPush(g, me.ID, "Jem Lightfoote, Sky Explorer", "Legendary Creature — Human Scout", jemLightfooteOracle, 3, 3)
	otjCastFromHand(t, g)
	hand := me.Hand.Size()
	advanceToEndStep(t, g)
	if !stackFullyEmpty(g) {
		t.Fatal("the ability triggered after a spell cast from the hand")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d, want no draw", hand, got)
	}
}

// --- Wrangler of the Damned --------------------------------------------------

func TestWranglerOfTheDamnedMakesASpirit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	otjPush(g, me.ID, "Wrangler of the Damned", "Creature — Human Soldier", wranglerOfTheDamned, 1, 4)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	spirits := tokensNamed(g, "Spirit")
	if len(spirits) != 1 {
		t.Fatalf("%d Spirits, want 1", len(spirits))
	}
	s := spirits[0]
	if s.Controller != me.ID || s.Power != 2 || s.Toughness != 2 || !slices.Equal(s.Colors, []string{"W"}) ||
		!game.HasKeyword(&s, "flying") {
		t.Errorf("Spirit %+v, want a 2/2 white Spirit with flying", s)
	}
}

// The intervening if is asked again as the ability resolves (CR 603.4):
// a second Wrangler flashed in from the hand in response turns it off.
func TestWranglerOfTheDamnedRespondingFromHandTurnsItOff(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	otjPush(g, me.ID, "Wrangler of the Damned", "Creature — Human Soldier", wranglerOfTheDamned, 1, 4)
	advanceToEndStep(t, g)
	if stackFullyEmpty(g) {
		t.Fatal("setup: the ability did not trigger")
	}
	flashed := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: flashed, Name: "Wrangler of the Damned", TypeLine: "Creature — Human Soldier",
			OracleID: wranglerOfTheDamned, Power: 1, Toughness: 4, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, flashed, game.CastSpellParams{}); err != nil {
		t.Fatalf("flash in a second Wrangler: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(flashed) {
		t.Fatal("setup: the flashed Wrangler did not resolve")
	}
	if n := len(tokensNamed(g, "Spirit")); n != 0 {
		t.Errorf("%d Spirits after a spell cast from the hand in response, want 0", n)
	}
}

// --- Canyon Crab -------------------------------------------------------------

func TestCanyonCrabLootsAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Drawn")
	otjPush(g, me.ID, "Canyon Crab", "Creature — Crab", canyonCrabOracle, 0, 5)
	hand := me.Hand.Size()
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Fatalf("hand %d → %d after the draw, want one drawn", hand, got)
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d, want draw then discard", hand, got)
	}
}

func TestCanyonCrabPumpsAndStaysQuietAfterAHandCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	crab := otjPush(g, me.ID, "Canyon Crab", "Creature — Crab", canyonCrabOracle, 0, 5)
	otjCastFromHand(t, g)
	fillPoolColored(me, "U", 1)
	b09AddColorless(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, crab, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePTOf(t, g, crab); p != 2 || tg != 3 {
		t.Errorf("P/T %d/%d, want 2/3 after +2/-2", p, tg)
	}
	advanceToEndStep(t, g)
	if !stackFullyEmpty(g) {
		t.Error("the loot triggered after a spell cast from the hand")
	}
}

// --- Prairie Dog -------------------------------------------------------------

func TestPrairieDogGrowsAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dog := otjPush(g, me.ID, "Prairie Dog", "Creature — Squirrel", prairieDogOracle, 2, 2)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, dog, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// Its {4}{W} adds one to every +1/+1 counter placement its controller
// makes on a creature they control this turn, its own end-step counter
// included; a hand cast earlier stops that counter.
func TestPrairieDogActivationAddsOneMoreCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dog := otjPush(g, me.ID, "Prairie Dog", "Creature — Squirrel", prairieDogOracle, 2, 2)
	bear := otjPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	fillPoolColored(me, "W", 1)
	b09AddColorless(me, 4)
	if err := g.ActivateCatalogAbility(me.ID, dog, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(bear, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
	})
	if got := countersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the Bear got %d +1/+1 counters from one, want 2", got)
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, dog, game.CounterPlusOne); got != 2 {
		t.Errorf("the Dog's end-step counter came to %d, want 2", got)
	}

	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	dog = otjPush(g, me.ID, "Prairie Dog", "Creature — Squirrel", prairieDogOracle, 2, 2)
	otjCastFromHand(t, g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, dog, game.CounterPlusOne); got != 0 {
		t.Errorf("%d +1/+1 counters after a hand cast, want 0", got)
	}
}

// --- Inventive Wingsmith -----------------------------------------------------

// One flying counter, then never another: at its next end step it
// already has one, so the ability doesn't trigger.
func TestInventiveWingsmithGetsOneFlyingCounter(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	smith := otjPush(g, me.ID, "Inventive Wingsmith", "Creature — Dwarf Artificer", inventiveWingsmithOracle, 2, 4)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, smith, game.CounterFlying); got != 1 {
		t.Fatalf("%d flying counters, want 1", got)
	}
	if c, ok := battlefieldCard(g, smith); !ok || !game.HasKeyword(&c, "flying") {
		t.Error("a flying counter did not give it flying (CR 122.1b)")
	}
	advanceToStepInTurn(t, g, game.StepCleanup)
	advanceToEndStepOf(t, g, seat)
	if !stackFullyEmpty(g) {
		t.Error("the ability triggered again with a flying counter already on it")
	}
	if got := countersOn(g, smith, game.CounterFlying); got != 1 {
		t.Errorf("%d flying counters at its next end step, want 1", got)
	}
}

// --- Emergent Haunting -------------------------------------------------------

// It becomes a 3/3 blue Spirit creature with flying, still an
// enchantment, and as a creature it never triggers again.
func TestEmergentHauntingBecomesASpiritOnce(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	haunting := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Emergent Haunting", TypeLine: "Enchantment", OracleID: emergentHauntingOracle,
		ManaCost: "{1}{U}", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID,
	})
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, haunting)
	if !ok || !c.IsCreature() || !c.HasCardType("enchantment") || !slices.Contains(c.Effective().Subtypes, "Spirit") ||
		!game.HasKeyword(&c, "flying") || !slices.Equal(c.Effective().Colors, []string{"U"}) {
		t.Fatalf("Haunting %+v, want a blue Spirit enchantment creature with flying", c.Effective())
	}
	if p, tg := effectivePTOf(t, g, haunting); p != 3 || tg != 3 {
		t.Errorf("P/T %d/%d, want 3/3", p, tg)
	}
	advanceToStepInTurn(t, g, game.StepCleanup)
	advanceToEndStepOf(t, g, seat)
	if !stackFullyEmpty(g) {
		t.Error("the ability triggered again while it is a creature")
	}
	if p, tg := effectivePTOf(t, g, haunting); p != 3 || tg != 3 {
		t.Errorf("at its next end step: P/T %d/%d, want 3/3 (the change has no duration)", p, tg)
	}
}

func TestEmergentHauntingSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top")
	haunting := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Emergent Haunting", TypeLine: "Enchantment", OracleID: emergentHauntingOracle,
		ManaCost: "{1}{U}", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	fillPoolColored(me, "U", 1)
	b09AddColorless(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, haunting, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID); c == nil {
		t.Fatal("no surveil prompt")
	}
}

func TestEmergentHauntingAfterAHandCastStaysAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	haunting := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Emergent Haunting", TypeLine: "Enchantment", OracleID: emergentHauntingOracle,
		Owner: me.ID, Controller: me.ID,
	})
	otjCastFromHand(t, g)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, haunting); !ok || c.IsCreature() {
		t.Error("it became a creature after a spell cast from the hand")
	}
}
