package teststo

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b48_cards_test.go — Star Trek Commander, batch 48 (#4974). Each
// test plays the real catalog entry and asserts what its text says.

const (
	trekB48SiskoOracle   = "f1416577-97d0-4d6c-afd2-455798556d4c"
	trekB48ToskOracle    = "48c2dc82-7bed-4730-8b8b-1970b743414f"
	trekB48EscapeOracle  = "629133e2-af20-4107-bc82-11ba32d5e2a6"
	trekB48CrypticOracle = "a3e51a35-09df-4189-b131-08a21e6a557d"
	trekB48AugurOracle   = "9e774188-caa1-4b62-b9cc-547f8e444b7f"
	trekB48SatyrOracle   = "ead4c01a-e2ae-49c9-95d9-651d81005907"
	trekB48MendingOracle = "b997a982-09e1-4ae1-a972-3c2e5f101e8c"
	trekB48PestOracle    = "072db429-3838-42ba-89b6-e26cbce22af9"
	trekB48RiteOracle    = "99957aa7-b654-41b5-9b96-7bce47545d25"
	trekB48SunOracle     = "52232aad-e4ca-4551-8c8d-c69a3b6b49c7"
)

func TestTrekB48EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB48SiskoOracle:   "Benjamin Sisko, Besieged",
		trekB48ToskOracle:    "Tenacious Tosk",
		trekB48EscapeOracle:  "Chronomantic Escape",
		trekB48CrypticOracle: "Cryptic Command",
		trekB48AugurOracle:   "Augur of Autumn",
		trekB48SatyrOracle:   "Satyr Wayfinder",
		trekB48MendingOracle: "Faithful Mending",
		trekB48PestOracle:    "Pest Control",
		trekB48RiteOracle:    "Rite of Oblivion",
		trekB48SunOracle:     "The Immortal Sun",
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

// --- Benjamin Sisko, Besieged ---------------------------------------------------

func TestTrekB48SiskoPlaysLandsAndCastsCreaturesFromARevealedTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.ToMain(t, g)
	sisko := effectstest.PushTopSource(t, g, me.ID, "Benjamin Sisko, Besieged", "Legendary Creature — Human Officer", trekB48SiskoOracle)
	if !effectstest.TopVisibleTo(g, me.ID, opp.ID) {
		t.Error("the top card is not revealed to the table")
	}
	bear := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast a creature off the top: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	land := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("play a land off the top: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land is not on the battlefield")
	}
	sorcery := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{0}"})
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Error("a sorcery was cast off the top")
	}
	effectstest.RemoveFromBattlefield(g, sisko)
	top := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Bear 2", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneLibrary).Granted() {
		t.Error("the permission outlived Sisko")
	}
}

// --- Tenacious Tosk ---------------------------------------------------------------

func TestTrekB48ToskCantBeDoubleBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tosk := effectstest.B12Push(g, me.ID, "Tenacious Tosk", "Creature — Tosk Rogue", trekB48ToskOracle, 4, 4)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, tosk)
	effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: tosk}, {Blocker: b, Attacker: tosk},
	}), game.BlockReasonTooManyBlockers)
	if err := g.DeclareBlocker(a, tosk); err != nil {
		t.Fatalf("one blocker is legal: %v", err)
	}
}

func TestTrekB48ToskGrowsWhenItHitsAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tosk := effectstest.B12Push(g, me.ID, "Tenacious Tosk", "Creature — Tosk Rogue", trekB48ToskOracle, 4, 4)
	life := opp.Life
	effectstest.BrAttack(t, g, tosk)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-4)
	}
	if c := effectstest.FindBattlefieldCardForTest(g, tosk); c == nil || c.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the Tosk has no +1/+1 counter after hitting a player: %+v", c)
	}
}

func TestTrekB48ToskBlockedGetsNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tosk := effectstest.B12Push(g, me.ID, "Tenacious Tosk", "Creature — Tosk Rogue", trekB48ToskOracle, 4, 4)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	effectstest.BrAttack(t, g, tosk)
	if err := g.DeclareBlocker(wall, tosk); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, tosk); c == nil || c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("a blocked Tosk got a counter: %+v", c)
	}
}

// --- Chronomantic Escape ----------------------------------------------------------

func TestTrekB48ChronomanticEscapeStopsAttacksOnYouUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if len(g.Seats) < 3 {
		t.Skip("needs three seats")
	}
	mySeat := g.Turn.ActiveSeat
	me := g.Seats[mySeat]
	oppSeat := (mySeat + 1) % len(g.Seats)
	opp, other := g.Seats[oppSeat], g.Seats[(mySeat+2)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 3},
	})
	id := effectstest.CastFromHand(t, g, game.Card{Name: "Chronomantic Escape", TypeLine: "Sorcery",
		ManaCost: "{4}{W}{W}", OracleID: trekB48EscapeOracle}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.InExile(g, id) {
		t.Fatal("Chronomantic Escape is not exiled")
	}
	if n := effectstest.CounterCountInExile(g, id, game.CounterTime); n != 3 {
		t.Fatalf("%d time counters, want 3", n)
	}
	for _, p := range g.Seats {
		if p.ID != me.ID && !effectstest.TrekB48Restricted(p, me.ID) {
			t.Errorf("%s is not restricted", p.Name)
		}
	}
	if effectstest.TrekB48Restricted(me, me.ID) {
		t.Error("the caster restricted themselves")
	}

	// A creature that enters after the spell resolved is refused too.
	raider := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Raider", TypeLine: "Creature — Test",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.TrekB48ToAttackersOf(t, g, oppSeat)
	err := g.Clone().DeclareAttacker(raider, me.ID)
	var pe *game.PlayerCantAttackError
	if !errors.As(err, &pe) {
		t.Fatalf("attack at the caster: err = %v, want *PlayerCantAttackError", err)
	}
	if err := g.Clone().DeclareAttacker(raider, walker); err != nil {
		t.Errorf("attack at the caster's planeswalker is legal (the ruling): %v", err)
	}
	if err := g.Clone().DeclareAttacker(raider, other.ID); err != nil {
		t.Errorf("attack at another player: %v", err)
	}

	// The caster's next turn ends it: on the opponent's following turn
	// the attack is legal again.
	effectstest.AdvanceToUpkeepOf(t, g, mySeat)
	effectstest.TrekB48ToAttackersOf(t, g, oppSeat)
	if err := g.Clone().DeclareAttacker(raider, me.ID); err != nil {
		t.Errorf("attack at the caster after their next turn: %v", err)
	}
}

func TestTrekB48ChronomanticEscapeHasSuspendThree(t *testing.T) {
	spec, _ := effects.Lookup(trekB48EscapeOracle)
	if len(spec.SpecialActions) != 1 {
		t.Fatalf("%d special actions, want suspend", len(spec.SpecialActions))
	}
	sa := spec.SpecialActions[0]
	if sa.Kind != game.SpecialActionSuspend || sa.Counters != 3 || sa.Cost != "{2}{W}" {
		t.Errorf("special action %+v, want suspend 3—{2}{W}", sa)
	}
}

// --- Cryptic Command --------------------------------------------------------------

const (
	trekB48CrypticCounter = iota
	trekB48CrypticBounce
	trekB48CrypticTap
	trekB48CrypticDraw
)

func TestTrekB48CrypticCommandCountersAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the creature: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.CastModal(t, g, "Cryptic Command", "Instant", trekB48CrypticOracle,
		[]int{trekB48CrypticCounter, trekB48CrypticDraw},
		[]game.TargetRef{effectstest.ModeRef(game.TargetCard, spell, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spell) || !me.Graveyard.Contains(spell) {
		t.Error("the spell was not countered")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (one card drawn)", got, hand+1)
	}
}

func TestTrekB48CrypticCommandBouncesAndTapsOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, g.Seats[len(g.Seats)-1].ID, "Bear B", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastModal(t, g, "Cryptic Command", "Instant", trekB48CrypticOracle,
		[]int{trekB48CrypticBounce, trekB48CrypticTap},
		[]game.TargetRef{effectstest.ModeRef(game.TargetCard, rock, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the permanent was not returned to its owner's hand")
	}
	for _, id := range []uuid.UUID{a, b} {
		if c := effectstest.FindBattlefieldCardForTest(g, id); c == nil || !c.Tapped {
			t.Errorf("an opponent's creature is untapped: %+v", c)
		}
	}
	if c := effectstest.FindBattlefieldCardForTest(g, mine); c == nil || c.Tapped {
		t.Error("the caster's own creature was tapped")
	}
}

func TestTrekB48CrypticCommandNeedsTwoDifferentModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	if _, err := effectstest.CastModalErr(t, g, "Cryptic Command", "Instant", trekB48CrypticOracle,
		[]int{trekB48CrypticDraw}, nil); err == nil {
		t.Error("one mode was accepted")
	}
	if _, err := effectstest.CastModalErr(t, g, "Cryptic Command", "Instant", trekB48CrypticOracle,
		[]int{trekB48CrypticDraw, trekB48CrypticDraw}, nil); err == nil {
		t.Error("the same mode twice was accepted")
	}
}

// With its only target gone the spell does nothing, the draw included.
func TestTrekB48CrypticCommandWithAnIllegalTargetDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastModal(t, g, "Cryptic Command", "Instant", trekB48CrypticOracle,
		[]int{trekB48CrypticBounce, trekB48CrypticDraw},
		[]game.TargetRef{effectstest.ModeRef(game.TargetCard, bear, 0, 0)})
	effectstest.RemoveFromBattlefield(g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d (no draw once every target is illegal)", got, hand)
	}
}

// --- Augur of Autumn --------------------------------------------------------------

func TestTrekB48AugurPlaysLandsAndCastsCreaturesOnlyWithCoven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Augur of Autumn", "Creature — Human Druid", trekB48AugurOracle, 2, 3)
	if !effectstest.TopVisibleTo(g, me.ID, me.ID) || effectstest.TopVisibleTo(g, me.ID, opp.ID) {
		t.Error("the look should be private to the Augur's controller")
	}
	land := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("play a land off the top: %v", err)
	}

	// Two powers (2 and 2): no coven.
	effectstest.B12Creature(g, me.ID, "Two", "Creature — Test", 2, 2)
	bear := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if effectstest.GrantedPermissionOn(g, me.ID, bear, game.ZoneLibrary).Granted() {
		t.Error("a creature on top is open without coven")
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Error("a creature was cast off the top without coven")
	}
	effectstest.B12Creature(g, me.ID, "One", "Creature — Test", 1, 1)
	if effectstest.GrantedPermissionOn(g, me.ID, bear, game.ZoneLibrary).Granted() {
		t.Error("two different powers opened the creature permission")
	}

	// Powers 1, 2 and 3: coven.
	effectstest.B12Creature(g, me.ID, "Three", "Creature — Test", 3, 3)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast a creature off the top with coven: %v", err)
	}
	sorcery := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{0}"})
	if effectstest.GrantedPermissionOn(g, me.ID, sorcery, game.ZoneLibrary).Granted() {
		t.Error("a sorcery on top is open")
	}
}

// --- Satyr Wayfinder --------------------------------------------------------------

func TestTrekB48SatyrWayfinderTakesALandAndBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := effectstest.TrekB48Library(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Spell", TypeLine: "Sorcery"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Fifth", TypeLine: "Instant"},
	)
	effectstest.CastFromHand(t, g, game.Card{Name: "Satyr Wayfinder", TypeLine: "Creature — Satyr",
		OracleID: trekB48SatyrOracle, ManaCost: "{1}{G}", Power: 1, Toughness: 1}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.LatestChooseCards(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 || ch.ChooseMax != 1 {
		t.Fatalf("want an optional pick of one land: %+v", ch)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[0]}); err == nil {
		t.Error("a nonland card was accepted")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(lib[1]) {
		t.Error("the chosen land is not in hand")
	}
	for _, i := range []int{0, 2, 3} {
		if !me.Graveyard.Contains(lib[i]) {
			t.Errorf("revealed card %d is not in the graveyard", i)
		}
	}
	if !me.Library.Contains(lib[4]) {
		t.Error("the fifth card left the library")
	}
}

// --- Faithful Mending -------------------------------------------------------------

func TestTrekB48FaithfulMendingGainsDrawsDiscardsAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life, hand := me.Life, me.Hand.Size()
	id := effectstest.CastFromHand(t, g, game.Card{Name: "Faithful Mending", TypeLine: "Instant",
		ManaCost: "{W}{U}", OracleID: trekB48MendingOracle}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d after the draw, want %d", got, hand+2)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after the discard, want %d", got, hand)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("Faithful Mending is not in the graveyard")
	}

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("the flashed-back spell is not in exile")
	}
	if me.Life != life+4 {
		t.Errorf("life %d after flashback, want %d", me.Life, life+4)
	}
}

// --- Pest Control -----------------------------------------------------------------

func TestTrekB48PestControlDestroysCheapNonlandPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	push := func(owner uuid.UUID, name, typeLine, cost string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
			ManaCost: cost, Power: 1, Toughness: 1, Owner: owner, Controller: owner})
	}
	one := push(me.ID, "Elf", "Creature — Elf", "{G}")
	x := push(opp.ID, "Hydra", "Creature — Hydra", "{X}{G}")
	zero := push(opp.ID, "Ornament", "Artifact", "{0}")
	two := push(opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	land := push(opp.ID, "Forest", "Basic Land — Forest", "")
	food := effectstest.PushToken(g, opp.ID, effects.FoodToken())
	effectstest.CastFromHand(t, g, game.Card{Name: "Pest Control", TypeLine: "Sorcery", ManaCost: "{W}{B}",
		OracleID: trekB48PestOracle}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"mana value 1": one, "{X}{G}": x, "mana value 0": zero, "token": food} {
		if g.Battlefield.Contains(id) {
			t.Errorf("the %s permanent survived", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"mana value 2": two, "land": land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("the %s permanent was destroyed", name)
		}
	}
}

func TestTrekB48PestControlCyclesForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	id, _ := effectstest.CycleFromHand(t, g, "Pest Control", "Sorcery", trekB48PestOracle, "{C}{C}")
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Error("the cycled card is not in the graveyard")
	}
	// The cycled card was added to the hand and left it; one was drawn.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (cycled one, drew one)", got, hand+1)
	}
}

// --- Rite of Oblivion -------------------------------------------------------------

func TestTrekB48RiteOfOblivionSacrificesThenExilesAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	fodder := effectstest.B12Permanent(g, me.ID, "Fodder", "Artifact")
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Test", 3, 3)
	id := effectstest.CastFromHand(t, g, game.Card{Name: "Rite of Oblivion", TypeLine: "Sorcery", ManaCost: "{W}{B}",
		OracleID: trekB48RiteOracle}, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	})
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice was not paid as the spell was cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !effectstest.InExile(g, victim) {
		t.Error("the target was not exiled")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("Rite of Oblivion is not in the graveyard")
	}

	// The flashback cast still needs the sacrifice: with no nonland
	// permanent to give up, it is refused.
	victim2 := effectstest.B12Creature(g, opp.ID, "Victim 2", "Creature — Test", 3, 3)
	land := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback",
		SacrificeIDs: []uuid.UUID{land}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim2}}})
	if err == nil {
		t.Fatal("a land paid the nonland sacrifice")
	}
	fodder2 := effectstest.B12Permanent(g, me.ID, "Fodder 2", "Enchantment")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback",
		SacrificeIDs: []uuid.UUID{fodder2}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim2}}}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, victim2) || !effectstest.InExile(g, id) {
		t.Error("the flashback cast did not exile its target and itself")
	}
}

func TestTrekB48RiteOfOblivionRefusesALandTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	fodder := effectstest.B12Permanent(g, me.ID, "Fodder", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rite of Oblivion", TypeLine: "Sorcery", ManaCost: "{W}{B}",
		OracleID: trekB48RiteOracle, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "WB")
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		SacrificeIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("a land target: err %v, want ErrIllegalTarget", err)
	}
}

// --- The Immortal Sun -------------------------------------------------------------

func TestTrekB48ImmortalSunAnthemAndCostReduction(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.ToMainForCost(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Test", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Test", 2, 2)

	bear := effectstest.Spm5Hand(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	effectstest.FloatForTest(g, me, "G")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a {1}{G} spell was cast off one {G} without the Sun")
	}

	effectstest.B12Push(g, me.ID, "The Immortal Sun", "Legendary Artifact", trekB48SunOracle, 0, 0)
	if p := effectstest.EffectivePower(t, g, mine); p != 3 {
		t.Errorf("your creature has power %d, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a {1}{G} spell off one {G} with the Sun: %v", err)
	}
}

func TestTrekB48ImmortalSunStopsLoyaltyAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	sorin := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sorin, Lord of Innistrad", OracleID: effectstest.NghdSorinLordOracle,
		TypeLine: "Legendary Planeswalker — Sorin", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	effectstest.B12Push(g, me.ID, "The Immortal Sun", "Legendary Artifact", trekB48SunOracle, 0, 0)
	err := g.ActivateCatalogAbility(me.ID, sorin, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrCantActivate) {
		t.Fatalf("a loyalty ability under the Sun: err = %v, want ErrCantActivate", err)
	}
	if got := effectstest.LoyaltyOf(g, sorin); got != 3 {
		t.Errorf("loyalty %d after a refused activation, want 3", got)
	}
}

func TestTrekB48ImmortalSunDrawsAnExtraCardInYourDrawStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mySeat := g.Turn.ActiveSeat
	me := g.Seats[mySeat]
	effectstest.B12Push(g, me.ID, "The Immortal Sun", "Legendary Artifact", trekB48SunOracle, 0, 0)
	effectstest.AdvanceToUpkeepOf(t, g, (mySeat+1)%len(g.Seats))
	other := g.Seats[g.Turn.ActiveSeat]
	otherHand := other.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepDraw)
	effectstest.PassPriorityAroundTable(t, g)
	if got := other.Hand.Size(); got != otherHand+1 {
		t.Errorf("an opponent drew %d in their draw step, want 1", got-otherHand)
	}

	effectstest.AdvanceToUpkeepOf(t, g, mySeat)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepDraw)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("drew %d in the draw step, want 2", got-hand)
	}
}
