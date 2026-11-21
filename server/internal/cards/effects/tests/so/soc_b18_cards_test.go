package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b18_cards_test.go — Secrets of Strixhaven Commander, eighteenth batch
// (#4834). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	socB18JadziOracle     = "c63b3511-d597-4c6e-9094-eabdeaa8a394"
	socB18MaestroOracle   = "485233bf-dc4b-4f31-8dfb-5054aee70c08"
	socB18ProfessorOracle = "0859d16e-2a1f-4b48-8687-99af8a8bdb38"
	socB18DellianOracle   = "8c5fe985-69e1-4a03-99fc-a6fd2314a9a7"
	socB18StealOracle     = "de403543-e61a-44fc-85e1-1e621d90ca8b"
	socB18FeasterOracle   = "0c4277ec-aa12-41f5-8896-c83cfc754d6e"
	socB18ZaffaiOracle    = "3f8a8650-0e38-4926-8f48-526351e6a863"
)

func TestSocB18EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB18JadziOracle:        "Jadzi, Steward of Fate",
		socB18JadziOracle + "#1": "Oracle's Gift",
		socB18MaestroOracle:      "Molten-Core Maestro",
		socB18ProfessorOracle:    "Postmortem Professor",
		socB18DellianOracle:      "Professor Dellian Fel",
		socB18StealOracle:        "Steal the Show",
		socB18FeasterOracle:      "Tragedy Feaster",
		socB18ZaffaiOracle:       "Zaffai and the Tempests",
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

// --- Jadzi, Steward of Fate // Oracle's Gift --------------------------------

func socB18CastJadzi(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB18JadziOracle, nil, "Jadzi, Steward of Fate", "Legendary Creature — Human Wizard", "{2}{U}", 2, 4,
		"Oracle's Gift", "Sorcery", "{X}{X}{U}")
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Jadzi: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return c.InstanceID
}

func TestSocB18JadziEntersPreparedAndLootsTwoOverTheDrawnCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drawn := effectstest.SeedLibrary(me, "Drawn One", "Drawn Two", "Deeper")
	jadzi := socB18CastJadzi(t, g, me)
	if !g.Battlefield.Contains(jadzi) {
		t.Fatal("Jadzi did not enter")
	}
	if !g.IsPreparedForEffect(jadzi) {
		t.Fatal("Jadzi did not enter prepared")
	}
	if _, ok := effectstest.PrepareCopyOf(g, "Oracle's Gift"); !ok {
		t.Fatal("no Oracle's Gift copy in exile")
	}
	if !me.Hand.Contains(drawn[0]) || !me.Hand.Contains(drawn[1]) {
		t.Fatalf("the two cards were not drawn before the discard (hand %d)", me.Hand.Size())
	}
	if n := effectstest.DiscardOwed(g, me.ID); n != 2 {
		t.Fatalf("owes %d discards, want 2", n)
	}
	// A card just drawn is a legal discard (CR 608.2c: draw, then discard).
	effectstest.AnswerDiscard(t, g, me.ID, drawn[0], drawn[1])
	if !me.Graveyard.Contains(drawn[0]) || !me.Graveyard.Contains(drawn[1]) {
		t.Error("the two drawn cards are not in the graveyard")
	}
}

// socB18Fractal is a nontoken Fractal creature already on the table.
func socB18Fractal(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.B12Creature(g, owner, name, "Creature — Fractal", 1, 1)
}

func socB18Tokens(g *game.Game, owner uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Fractal" && c.Controller == owner {
			out = append(out, c)
		}
	}
	return out
}

func socB18CastOraclesGift(t *testing.T, g *game.Game, me *game.Player, jadzi uuid.UUID, x int, params game.CastSpellParams) error {
	t.Helper()
	cp, ok := effectstest.PrepareCopyOf(g, "Oracle's Gift")
	if !ok {
		t.Fatal("no Oracle's Gift copy in exile")
	}
	params.FromZone = string(game.ZoneExile)
	params.XValue = x
	return g.CastSpell(me.ID, cp.InstanceID, params)
}

func TestSocB18OraclesGiftMakesXFractalsThenXCountersOnEachFractalYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	jadzi := socB18CastJadzi(t, g, me)
	effectstest.DiscardFromHand(t, g, me.ID)
	mine := socB18Fractal(g, me.ID, "My Fractal")
	theirs := socB18Fractal(g, opp.ID, "Their Fractal")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	if err := socB18CastOraclesGift(t, g, me, jadzi, 2, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Oracle's Gift: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	tokens := socB18Tokens(g, me.ID)
	if len(tokens) != 2 {
		t.Fatalf("%d Fractal tokens, want X = 2", len(tokens))
	}
	for _, c := range tokens {
		if p, tough := c.CurrentPower(), c.CurrentToughness(); p != 2 || tough != 2 {
			t.Errorf("a new Fractal is %d/%d, want 2/2", p, tough)
		}
	}
	if c, ok := effectstest.BattlefieldCard(g, mine); !ok || c.CurrentPower() != 3 {
		t.Errorf("an old Fractal you control: %+v, want power 1+2", c)
	}
	if c, ok := effectstest.BattlefieldCard(g, theirs); !ok || c.CurrentPower() != 1 {
		t.Errorf("an opponent's Fractal got counters: power %d, want 1", c.CurrentPower())
	}
	if c, ok := effectstest.BattlefieldCard(g, bear); !ok || c.CurrentPower() != 2 {
		t.Errorf("a non-Fractal got counters: power %d, want 2", c.CurrentPower())
	}
	if g.IsPreparedForEffect(jadzi) {
		t.Error("Jadzi is still prepared after the copy was cast")
	}
}

func TestSocB18OraclesGiftForZeroMakesNothingAndAddsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	jadzi := socB18CastJadzi(t, g, me)
	effectstest.DiscardFromHand(t, g, me.ID)
	mine := socB18Fractal(g, me.ID, "My Fractal")
	if err := socB18CastOraclesGift(t, g, me, jadzi, 0, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB18Tokens(g, me.ID)); n != 0 {
		t.Errorf("%d tokens for X = 0", n)
	}
	if c, ok := effectstest.BattlefieldCard(g, mine); !ok || c.CurrentPower() != 1 {
		t.Errorf("an old Fractal changed for X = 0: %+v", c)
	}
}

func TestSocB18OraclesGiftCostsTwiceXPlusBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	jadzi := socB18CastJadzi(t, g, me)
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.FloatForTest(g, me, "CCCU") // {2}{2}{U} is five
	if err := socB18CastOraclesGift(t, g, me, jadzi, 2, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("X = 2 was cast for four mana")
	}
	effectstest.FloatForTest(g, me, "C")
	if err := socB18CastOraclesGift(t, g, me, jadzi, 2, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("X = 2 for five mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB18Tokens(g, me.ID)); n != 2 {
		t.Errorf("%d Fractals, want 2", n)
	}
}

// --- Molten-Core Maestro -------------------------------------------------

func socB18Maestro(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.B12Push(g, me.ID, "Molten-Core Maestro", "Creature — Goblin Bard", socB18MaestroOracle, 2, 2)
}

func socB18CastWithMana(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	effectstest.CastFromHandForTest(t, g, me, name, typeLine, cost, "", game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB18MaestroGrowsOnAnInstantOrSorceryAndAddsManaOnlyAtFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	m := socB18Maestro(g, me)

	socB18CastWithMana(t, g, me, "Cheap Instant", "Instant", "{1}{R}", "{C}{R}")
	c, _ := effectstest.BattlefieldCard(g, m)
	if c.CurrentPower() != 3 {
		t.Fatalf("power %d after a two-mana spell, want 3", c.CurrentPower())
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Fatalf("%d mana added for a two-mana spell", n)
	}

	socB18CastWithMana(t, g, me, "Four Sorcery", "Sorcery", "{3}{R}", "{C}{C}{C}{R}")
	c, _ = effectstest.BattlefieldCard(g, m)
	if c.CurrentPower() != 4 {
		t.Fatalf("power %d after a four-mana spell, want 4", c.CurrentPower())
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Fatalf("%d mana added for a four-mana spell", n)
	}

	socB18CastWithMana(t, g, me, "Five Sorcery", "Sorcery", "{4}{R}", "{C}{C}{C}{C}{R}")
	c, _ = effectstest.BattlefieldCard(g, m)
	if c.CurrentPower() != 5 {
		t.Fatalf("power %d after a five-mana spell, want 5", c.CurrentPower())
	}
	// {R} equal to the power AFTER the counter went on.
	if n := me.ManaPool.Total(); n != 5 {
		t.Fatalf("%d mana added, want 5 (the Maestro's power)", n)
	}
	if n := me.ManaPool.CountColor("R"); n != 5 {
		t.Errorf("%d of it is {R}, want all 5", n)
	}
}

func TestSocB18MaestroIgnoresACreatureSpellAndAnOpponentsSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	m := socB18Maestro(g, me)

	socB18CastWithMana(t, g, me, "Big Creature", "Creature — Bear", "{4}{R}", "{C}{C}{C}{C}{R}")
	if c, _ := effectstest.BattlefieldCard(g, m); c.CurrentPower() != 2 {
		t.Errorf("power %d after a creature spell, want 2", c.CurrentPower())
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana added for a creature spell", n)
	}

	opp := g.Seats[1]
	castCatalogSpellFor(t, g, opp, "Their Instant", "Instant")
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, m); c.CurrentPower() != 2 {
		t.Errorf("power %d after an opponent's instant, want 2", c.CurrentPower())
	}
}

func TestSocB18MaestroNeedsFiveManaSpentNotFiveManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	m := socB18Maestro(g, me)
	// A mana value of six, cast for nothing: no mana was spent.
	id := effectstest.HandCardFull(me, "Free Six", "Sorcery", "{4}{R}{R}", "", []string{"R"})
	g.WithWriteLock(func() { g.FreeCastFromHandForEffect(me.ID, uuid.Nil, id, "Test") })
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, m)
	if c.CurrentPower() != 3 {
		t.Errorf("power %d, want 3 (the counter still goes on)", c.CurrentPower())
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana added although none was spent", n)
	}
}

// castCatalogSpellFor casts a plain instant from `p`'s hand with no mana
// restrictions, for a seat that is not the active one.
func castCatalogSpellFor(t *testing.T, g *game.Game, p *game.Player, name, typeLine string) {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
}

// --- Postmortem Professor -----------------------------------------------

func socB18Professor(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.B12Push(g, me.ID, "Postmortem Professor", "Creature — Zombie Warlock", socB18ProfessorOracle, 2, 2)
}

func TestSocB18ProfessorDrainsEachOpponentWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prof := socB18Professor(g, me)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(prof, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		switch s.ID {
		case me.ID:
			if s.Life != lives[s.ID]+1 {
				t.Errorf("my life %d, want %d", s.Life, lives[s.ID]+1)
			}
		case opp.ID:
			// The attacked player may also have taken the combat damage by
			// now; never less than the drain.
			if s.Life > lives[s.ID]-1 {
				t.Errorf("the attacked opponent is at %d, want at most %d", s.Life, lives[s.ID]-1)
			}
		default:
			if s.Life != lives[s.ID]-1 {
				t.Errorf("seat %s life %d, want %d", s.ID, s.Life, lives[s.ID]-1)
			}
		}
	}
}

func TestSocB18ProfessorCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prof := socB18Professor(g, opp)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	attackerOwner := g.Seats[g.Turn.ActiveSeat]
	if attackerOwner.ID != me.ID {
		t.Fatal("setup: seat 0 is not the active player")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(prof, bear); err == nil {
		t.Fatal("the Professor blocked")
	}
}

func TestSocB18ProfessorReturnsFromTheGraveyardByExilingAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[0]
	prof := effectstest.PushCatalogGraveyardCard(me, "Postmortem Professor", "Creature — Zombie Warlock", socB18ProfessorOracle, 2, 2)
	sorcery := effectstest.GyCard(me, "Dead Sorcery", "Sorcery", "{1}{R}", "")
	creature := effectstest.GyCard(me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}"); err != nil {
		t.Fatal(err)
	}
	// A creature card is not an instant or sorcery card.
	if err := g.ActivateCatalogAbility(me.ID, prof, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{creature}}); err == nil {
		t.Fatal("exiled a creature card to pay for it")
	}
	if err := g.ActivateCatalogAbility(me.ID, prof, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without exiling a card")
	}
	if err := g.ActivateCatalogAbility(me.ID, prof, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{sorcery}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(sorcery) {
		t.Error("the sorcery was not exiled as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, prof)
	if !ok {
		t.Fatal("the Professor is not on the battlefield")
	}
	if c.Tapped {
		t.Error("the Professor returned tapped")
	}
	if me.Graveyard.Contains(prof) {
		t.Error("the Professor is still in the graveyard")
	}
	if !me.Graveyard.Contains(creature) {
		t.Error("the creature card left the graveyard")
	}
}

// --- Professor Dellian Fel -------------------------------------------------

func socB18ActivateDellian(t *testing.T, g *game.Game, me *game.Player, dellian uuid.UUID, idx int, targets ...uuid.UUID) error {
	t.Helper()
	var refs []game.TargetRef
	for _, id := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return g.ActivateCatalogAbility(me.ID, dellian, idx, game.ActivateAbilityParams{Targets: refs})
}

func TestSocB18DellianPlusTwoGainsThreeAndZeroDrawsAndLoses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	d := effectstest.PushWalkerForTest(g, me.ID, "Professor Dellian Fel", socB18DellianOracle, 5)
	life := me.Life
	if err := socB18ActivateDellian(t, g, me, d, 0); err != nil {
		t.Fatalf("+2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d after +2, want %d", me.Life, life+3)
	}
	if got := effectstest.LoyaltyOf(g, d); got != 7 {
		t.Errorf("loyalty %d after +2, want 7", got)
	}
}

func TestSocB18DellianZeroDrawsACardAndLosesOneLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	d := effectstest.PushWalkerForTest(g, me.ID, "Professor Dellian Fel", socB18DellianOracle, 5)
	life, hand := me.Life, me.Hand.Size()
	if err := socB18ActivateDellian(t, g, me, d, 1); err != nil {
		t.Fatalf("0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d (was %d), hand %d (was %d); want -1 life and +1 card", me.Life, life, me.Hand.Size(), hand)
	}
	if got := effectstest.LoyaltyOf(g, d); got != 5 {
		t.Errorf("loyalty %d after 0, want 5", got)
	}
}

func TestSocB18DellianMinusThreeDestroysTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	d := effectstest.PushWalkerForTest(g, me.ID, "Professor Dellian Fel", socB18DellianOracle, 5)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := socB18ActivateDellian(t, g, me, d, 2, land); err == nil {
		t.Fatal("targeted a land")
	}
	if err := socB18ActivateDellian(t, g, me, d, 2, victim); err != nil {
		t.Fatalf("−3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature survived")
	}
	if got := effectstest.LoyaltyOf(g, d); got != 2 {
		t.Errorf("loyalty %d after −3, want 2", got)
	}
}

func TestSocB18DellianEmblemDrainsAnOpponentForTheLifeGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	d := effectstest.PushWalkerForTest(g, me.ID, "Professor Dellian Fel", socB18DellianOracle, 6)
	if err := socB18ActivateDellian(t, g, me, d, 3); err != nil {
		t.Fatalf("−6: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.EmblemsOf(g, me.ID)); n != 1 {
		t.Fatalf("%d emblems, want 1", n)
	}
	if opp.Life != 40 {
		t.Fatalf("setup: opponent at %d", opp.Life)
	}
	// Gain 4: the emblem's trigger asks for a target opponent.
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 4); err != nil {
			t.Errorf("gain: %v", err)
		}
	})
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 36 {
		t.Errorf("the opponent is at %d, want 36 (lost the 4 gained)", opp.Life)
	}
	// An opponent gaining life does not trigger it.
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 5); err != nil {
			t.Errorf("gain: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 40+4 || opp.Life != 41 {
		t.Errorf("life me %d opp %d, want 44 and 41", me.Life, opp.Life)
	}
}

// --- Steal the Show -----------------------------------------------------

func socB18Steal(t *testing.T, g *game.Game, me *game.Player, modes []int, targets []game.TargetRef) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Steal the Show", TypeLine: "Sorcery", ManaCost: "{2}{R}",
		OracleID: socB18StealOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets}); err != nil {
		t.Fatalf("cast Steal the Show: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB18StealTheShowDamagesByInstantsAndSorceriesInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.GyCard(me, "Old Instant", "Instant", "{R}", "")
	effectstest.GyCard(me, "Old Sorcery", "Sorcery", "{1}{R}", "")
	effectstest.GyCard(me, "Old Creature", "Creature — Bear", "{1}{G}", "")
	// An opponent's graveyard does not count.
	effectstest.GyCard(opp, "Their Instant", "Instant", "{U}", "")
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	socB18Steal(t, g, me, []int{1}, []game.TargetRef{effectstest.ModeRef(game.TargetCard, victim, 0, 0)})
	c, ok := effectstest.BattlefieldCard(g, victim)
	if !ok || c.DamageMarked != 2 {
		t.Errorf("damage %d on the target (alive %v), want 2", c.DamageMarked, ok)
	}
}

func TestSocB18StealTheShowDamagesAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.GyCard(me, "Old Instant", "Instant", "{R}", "")
	effectstest.GyCard(me, "Old Sorcery", "Sorcery", "{1}{R}", "")
	effectstest.GyCard(me, "Old Instant 2", "Instant", "{R}", "")
	walker := effectstest.PushWalkerForTest(g, opp.ID, "Their Walker", "", 5)
	socB18Steal(t, g, me, []int{1}, []game.TargetRef{effectstest.ModeRef(game.TargetCard, walker, 0, 0)})
	if got := effectstest.LoyaltyOf(g, walker); got != 2 {
		t.Errorf("loyalty %d, want 2 (3 damage)", got)
	}
}

func TestSocB18StealTheShowDiscardsAnyNumberThenDrawsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(opp, "N1", "N2", "N3", "N4")
	opp.Hand.Cards = nil
	h := effectstest.FillHand(opp, "H1", "H2", "H3")
	socB18Steal(t, g, me, []int{0}, []game.TargetRef{effectstest.ModeRef(game.TargetPlayer, opp.ID, 0, 0)})
	c := effectstest.DiscardChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the target player was not asked")
	}
	if c.ChooseMin != 0 || c.ChooseMax != 3 {
		t.Errorf("bounds [%d,%d], want any number from 0 to 3", c.ChooseMin, c.ChooseMax)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, h[0], h[2])
	g.SettleResolution()
	if !opp.Graveyard.Contains(h[0]) || !opp.Graveyard.Contains(h[2]) || !opp.Hand.Contains(h[1]) {
		t.Error("the wrong cards were discarded")
	}
	if got := opp.Hand.Size(); got != 3 {
		t.Errorf("hand %d, want 3 (discarded 2, drew 2)", got)
	}
}

func TestSocB18StealTheShowCanDiscardNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "N1", "N2")
	me.Hand.Cards = nil
	h := effectstest.FillHand(me, "H1", "H2")
	socB18Steal(t, g, me, []int{0}, []game.TargetRef{effectstest.ModeRef(game.TargetPlayer, me.ID, 0, 0)})
	effectstest.AnswerDiscard(t, g, me.ID)
	g.SettleResolution()
	if !me.Hand.Contains(h[0]) || !me.Hand.Contains(h[1]) || me.Hand.Size() != 2 {
		t.Errorf("hand %d after discarding nothing, want the same 2 cards", me.Hand.Size())
	}
}

func TestSocB18StealTheShowBothModesCountTheCardsJustDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "N1", "N2", "N3")
	me.Hand.Cards = nil
	i1 := effectstest.HandCardFull(me, "Held Instant", "Instant", "{R}", "", []string{"R"})
	s1 := effectstest.HandCardFull(me, "Held Sorcery", "Sorcery", "{R}", "", []string{"R"})
	other := effectstest.HandCardFull(me, "Held Land", "Basic Land — Mountain", "", "", nil)
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	socB18Steal(t, g, me, []int{0, 1}, []game.TargetRef{
		effectstest.ModeRef(game.TargetPlayer, me.ID, 0, 0),
		effectstest.ModeRef(game.TargetCard, victim, 1, 0),
	})
	// The damage waits for the discard: nothing is marked yet.
	if c, _ := effectstest.BattlefieldCard(g, victim); c.DamageMarked != 0 {
		t.Fatalf("damage dealt before the discard was answered: %d", c.DamageMarked)
	}
	effectstest.AnswerDiscard(t, g, me.ID, i1, s1, other)
	g.SettleResolution()
	c, ok := effectstest.BattlefieldCard(g, victim)
	if !ok || c.DamageMarked != 2 {
		t.Errorf("damage %d (alive %v), want 2 for the instant and the sorcery just discarded", c.DamageMarked, ok)
	}
	if me.Hand.Size() != 3 {
		t.Errorf("hand %d, want 3 (discarded 3, drew 3)", me.Hand.Size())
	}
}

// --- Tragedy Feaster ----------------------------------------------------

func TestSocB18FeasterWardsWithADiscard(t *testing.T) {
	g, me, opp, warded := effectstest.WardTable(t, "Tragedy Feaster", "Creature — Demon", socB18FeasterOracle)
	hand := effectstest.FillHand(opp, "Filler A", "Filler B")
	effectstest.BoltWardedCreature(t, g, opp, warded)
	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if n := effectstest.DiscardOwed(g, opp.ID); n != 1 {
		t.Fatalf("the payer owes %d cards, want 1", n)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, hand[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(hand[0]) || !effectstest.BoltResolved(g, warded) {
		t.Error("the ward was paid: the discard should have happened and the Bolt resolved")
	}
	_ = me
}

func TestSocB18FeasterHasTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	f := effectstest.PushCatalogPermanent(g, me.ID, "Tragedy Feaster", "Creature — Demon", socB18FeasterOracle, false)
	if !effectstest.HasEffectiveKeyword(t, g, f, "trample") {
		t.Error("no trample")
	}
}

func socB18ToEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepEnd)
}

func TestSocB18FeasterSacrificesAPermanentOfYourChoiceWhenYouGainedNoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	f := effectstest.PushCatalogPermanent(g, me.ID, "Tragedy Feaster", "Creature — Demon", socB18FeasterOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	socB18ToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no sacrifice was asked for")
	}
	for _, id := range []uuid.UUID{f, bear, land} {
		if !effectstest.HasID(pick.ChooseCards, id) {
			t.Errorf("permanent %s is not offered", id)
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bear)
	g.SettleResolution()
	if g.Battlefield.Contains(bear) {
		t.Error("the chosen permanent was not sacrificed")
	}
	if !g.Battlefield.Contains(f) || !g.Battlefield.Contains(land) {
		t.Error("something else was sacrificed too")
	}
}

func TestSocB18FeasterSpareYouWhenYouGainedLifeThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Tragedy Feaster", "Creature — Demon", socB18FeasterOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1); err != nil {
			t.Fatalf("gain: %v", err)
		}
	})
	socB18ToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("asked for a sacrifice after gaining life")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("lost a permanent after gaining life")
	}
}

func TestSocB18FeasterLosingLifeIsNotGainingIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	f := effectstest.PushCatalogPermanent(g, me.ID, "Tragedy Feaster", "Creature — Demon", socB18FeasterOracle, false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -3); err != nil {
			t.Fatalf("lose: %v", err)
		}
	})
	socB18ToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no sacrifice was asked for after only losing life")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, f)
	g.SettleResolution()
	if g.Battlefield.Contains(f) {
		t.Error("the Feaster can sacrifice itself and did not")
	}
}

func TestSocB18FeasterDoesNothingAtAnOpponentsEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	f := effectstest.PushCatalogPermanent(g, me.ID, "Tragedy Feaster", "Creature — Demon", socB18FeasterOracle, false)
	effectstest.EndTurn(t, g)
	socB18ToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil || !g.Battlefield.Contains(f) {
		t.Error("the Feaster triggered at an opponent's end step")
	}
}

// --- Zaffai and the Tempests ---------------------------------------------

func socB18Zaffai(g *game.Game, me *game.Player) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, me.ID, "Zaffai and the Tempests", "Legendary Creature — Human Bard Sorcerer", socB18ZaffaiOracle, false)
}

func TestSocB18ZaffaiCastsAnInstantOrSorceryFromHandFreeOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	socB18Zaffai(g, me)
	key := effects.GrantedAltCostFreeInstantOrSorceryOnceYourTurn

	sorcery := effectstest.HandSorcery(me, "Three Sorcery", "{1}{B}{B}")
	instant := effectstest.HandCardFull(me, "Held Instant", "Instant", "{R}", "", []string{"R"})
	creature := effectstest.HandCardFull(me, "Held Bear", "Creature — Bear", "{1}{G}", "", []string{"G"})
	for _, id := range []uuid.UUID{sorcery, instant} {
		if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, id, game.ZoneHand), key) == nil {
			t.Fatal("an instant or sorcery in hand is not offered the free cast")
		}
	}
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, creature, game.ZoneHand), key) != nil {
		t.Error("a creature card was offered the free cast")
	}
	offer := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, sorcery, game.ZoneHand), key)
	if offer.ManaCost != "" || offer.Life != 0 {
		t.Errorf("the offer charges %q and %d life, want nothing", offer.ManaCost, offer.Life)
	}

	// Strict with an empty pool: the free cast is the only way to pay.
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast a three-mana sorcery for nothing without the offer")
	}
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true, AlternativeCost: key}); err != nil {
		t.Fatalf("free cast from hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(sorcery) {
		t.Error("the sorcery did not resolve")
	}

	// Once during each of your turns.
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, instant, game.ZoneHand), key) != nil {
		t.Error("a second free cast was offered this turn")
	}
	if err := g.CastSpell(me.ID, instant, game.CastSpellParams{Strict: true, AlternativeCost: key}); err == nil {
		t.Error("the cast gate accepted a second free cast this turn")
	}
}

func TestSocB18ZaffaiNotDuringAnOpponentsTurnAndComesBackNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.ToMain(t, g)
	socB18Zaffai(g, me)
	key := effects.GrantedAltCostFreeInstantOrSorceryOnceYourTurn
	instant := effectstest.HandCardFull(me, "Held Instant", "Instant", "{R}", "", []string{"R"})

	effectstest.AdvanceToMainOf(t, g, (seat+1)%len(g.Seats))
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, instant, game.ZoneHand), key) != nil {
		t.Fatal("Zaffai offered its free cast during an opponent's turn")
	}
	effectstest.AdvanceToMainOf(t, g, seat)
	if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, instant, game.ZoneHand), key) == nil {
		t.Fatal("the free cast did not come back on my next turn")
	}
}

func TestSocB18ZaffaiFreeCastOfAnXSpellHasXZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	socB18Zaffai(g, me)
	key := effects.GrantedAltCostFreeInstantOrSorceryOnceYourTurn
	effectstest.SeedLibrary(me, "A", "B", "C")
	xs := effectstest.HandCardFull(me, "X Sorcery", "Sorcery", "{X}{R}", "", []string{"R"})
	// CR 107.3b: the only legal choice for X is 0, so a cast naming another
	// X is refused or runs with X = 0, never with 3.
	if err := g.CastSpell(me.ID, xs, game.CastSpellParams{Strict: true, AlternativeCost: key, XValue: 3}); err == nil {
		for _, it := range g.StackMeta {
			if it != nil && it.XValue != 0 {
				t.Errorf("a free cast ran with X = %d", it.XValue)
			}
		}
	}
}
