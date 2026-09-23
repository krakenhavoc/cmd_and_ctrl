package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// amass_cards_test.go — #1236, the catalog half of CR 701.47. The
// engine half (find-or-create, the multi-Army prompt, the CR 614
// windows, the layer-4 subtype) is in game/amass_test.go; everything
// here is a real card, cast or triggered through the stack.
//
// Four proofs, chosen for what each one adds:
//
//	Orcish Bowmasters    the row's named card — amass off a trigger,
//	                     twice in a turn, on the same Army
//	Dreadhorde Invasion  find-or-create across turns: one token, then
//	                     counters
//	Eternal Skylord      the SUBTYPE reaching another card's static
//	Widespread Brutality "the Army you amassed" (CR 701.47c) and the
//	                     "non-Army" read side

const (
	orcishBowmastersOracle   = "ea5103f5-27e0-4eb1-902c-7f34652d6bf3"
	dreadhordeInvasionOracle = "01deabbb-af6a-4998-99a7-35b7cfa9ef77"
	eternalSkylordOracle     = "324a30cc-bddd-463c-9f28-c94c93a26780"
	widespreadBrutalityOracl = "3a06f4c1-2af1-495f-914f-1ac4e26f87d4"
)

// armyOf returns the single Army creature `controller` controls, or
// fails. Every test below expects exactly one, which is the point of
// find-or-create.
func armyOf(t *testing.T, g *game.Game, controller uuid.UUID) *game.Card {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	var found *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && game.IsArmy(*c) {
			if found != nil {
				t.Fatalf("two Armies on the battlefield; amass must find the one already there")
			}
			found = c
		}
	}
	if found == nil {
		t.Fatal("no Army on the battlefield — the amass created nothing")
	}
	return found
}

// --- ArmyToken, the template the keyword derives ---------------------

// TestArmyTokenIsTheKeywordsPrintedToken pins the derivation against
// CR 701.47a's words: "a 0/0 black [subtype] Army creature token". The
// template is built from the subtype rather than looked up in
// tokens_table.go (ADR 0087 decision 2), so this is the only thing
// standing between a typo and a colourless 1/1.
func TestArmyTokenIsTheKeywordsPrintedToken(t *testing.T) {
	for _, subtype := range []string{"Zombie", "Orc", "Goblin", "Sliver"} {
		tok := ArmyToken(subtype)
		if tok.TypeLine != "Token Creature — "+subtype+" Army" {
			t.Errorf("%s: type line = %q", subtype, tok.TypeLine)
		}
		if tok.Power != 0 || tok.Toughness != 0 || !tok.PrintedPTKnown {
			t.Errorf("%s: %d/%d known=%v, want a KNOWN 0/0 so CR 704.5f can kill an amass 0",
				subtype, tok.Power, tok.Toughness, tok.PrintedPTKnown)
		}
		if len(tok.Colors) != 1 || tok.Colors[0] != "B" {
			t.Errorf("%s: colors = %v, want black", subtype, tok.Colors)
		}
		if len(tok.Keywords) != 0 {
			t.Errorf("%s: the Army token is vanilla, got keywords %v", subtype, tok.Keywords)
		}
	}
	// CR 701.47d: the War of the Spark cards printed "amass N" with no
	// subtype and were errata'd to "amass Zombies N".
	if got := ArmyToken("").TypeLine; got != "Token Creature — Zombie Army" {
		t.Errorf("a subtype-less amass makes %q, want the CR 701.47d Zombie Army", got)
	}
}

// --- Orcish Bowmasters ------------------------------------------------

// TestOrcishBowmastersShootsAndAmassesOnEntry is the card's first
// trigger condition: it enters, pings a target, and amasses Orcs 1.
func TestOrcishBowmastersShootsAndAmassesOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	before := opp.Life

	castCatalogSpell(t, g, "Orcish Bowmasters", "Creature — Orc Archer", orcishBowmastersOracle, nil)
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	if opp.Life != before-1 {
		t.Errorf("opponent life = %d, want %d — the Bowmasters deal 1 damage", opp.Life, before-1)
	}
	army := armyOf(t, g, me.ID)
	if !army.HasSubtype("Orc") {
		t.Errorf("the Army is %q, want an Orc Army", army.TypeLine)
	}
	if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

// TestOrcishBowmastersGrowsTheSameArmyOnEveryDraw is the card's second
// trigger condition and the whole reason amass needed find-or-create:
// a wheel is one Army with N counters on it, not N 0/0 Armies.
func TestOrcishBowmastersGrowsTheSameArmyOnEveryDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]

	castCatalogSpell(t, g, "Orcish Bowmasters", "Creature — Orc Archer", orcishBowmastersOracle, nil)
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	first := armyOf(t, g, me.ID).InstanceID

	// Two draws by an opponent, neither in their own draw step (it is
	// still the active player's turn).
	for i := 0; i < 2; i++ {
		if err := g.DrawCard(opp.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
		passPriorityAroundTable(t, g)
		b16PickPlayer(t, g, me.ID, opp.ID)
		passPriorityAroundTable(t, g)
	}

	army := armyOf(t, g, me.ID)
	if army.InstanceID != first {
		t.Error("a second Army was created; amass must find the one already out")
	}
	if got := countersOn(g, first, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 — one per trigger", got)
	}
	if p := army.CurrentPower(); p != 3 {
		t.Errorf("the Army is a %d/%d, want 3/3", p, army.CurrentToughness())
	}
}

// TestOrcishBowmastersIgnoresTheDrawStepDraw is the declared
// simplification, pinned so it cannot quietly become something else:
// a draw during the drawing player's OWN draw step does not fire the
// Bowmasters. Printed, only the first such draw is exempt.
func TestOrcishBowmastersIgnoresTheDrawStepDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]

	castCatalogSpell(t, g, "Orcish Bowmasters", "Creature — Orc Archer", orcishBowmastersOracle, nil)
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	before := countersOn(g, armyOf(t, g, me.ID).InstanceID, game.CounterPlusOne)

	// Walk to the next player's draw step; the turn-based draw is
	// theirs, in their own draw step.
	advanceToDrawStepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)

	if got := countersOn(g, armyOf(t, g, me.ID).InstanceID, game.CounterPlusOne); got != before {
		t.Errorf("+1/+1 counters = %d, want %d — the draw-step draw is exempt", got, before)
	}
}

// advanceToDrawStepOfSeat walks to the named seat's draw step.
func advanceToDrawStepOfSeat(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if g.Turn.Step == game.StepDraw && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("never reached seat %d's draw step", seat)
}

// --- Dreadhorde Invasion ---------------------------------------------

// TestDreadhordeInvasionMakesOneArmyAndGrowsIt is find-or-create
// across turns, on the card that makes it visible: the first upkeep
// mints the Army, the second puts a second counter on the same one,
// and the controller loses a life each time.
func TestDreadhordeInvasionMakesOneArmyAndGrowsIt(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Dreadhorde Invasion",
		TypeLine:   "Enchantment",
		OracleID:   dreadhordeInvasionOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	life := me.Life

	advanceToUpkeepOfSeat(t, g, seat)
	passPriorityAroundTable(t, g)
	first := armyOf(t, g, me.ID)
	if !first.HasSubtype("Zombie") {
		t.Errorf("the Army is %q, want a Zombie Army", first.TypeLine)
	}
	if me.Life != life-1 {
		t.Errorf("life = %d, want %d", me.Life, life-1)
	}
	firstID := first.InstanceID

	advanceToUpkeepOfSeat(t, g, seat)
	passPriorityAroundTable(t, g)
	second := armyOf(t, g, me.ID)
	if second.InstanceID != firstID {
		t.Error("the second upkeep built a new Army instead of growing the first")
	}
	if got := countersOn(g, firstID, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if me.Life != life-2 {
		t.Errorf("life = %d, want %d", me.Life, life-2)
	}
}

// advanceToUpkeepOfSeat walks to the named seat's NEXT upkeep.
func advanceToUpkeepOfSeat(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == seat {
			return
		}
	}
	t.Fatalf("never reached seat %d's upkeep", seat)
}

// --- Eternal Skylord --------------------------------------------------

// TestEternalSkylordsArmyFlies is CR 701.47a's last sentence reaching
// a static on another card: the Army is a Zombie TOKEN, so "Zombie
// tokens you control have flying" finds it. Nothing in the Skylord's
// file knows what an Army is.
func TestEternalSkylordsArmyFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	castCatalogSpell(t, g, "Eternal Skylord", "Creature — Zombie Wizard", eternalSkylordOracle, nil)
	passPriorityAroundTable(t, g)

	army := armyOf(t, g, me.ID)
	if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if !game.HasKeyword(army, "flying") {
		t.Error("the amassed Zombie Army has no flying — the Skylord's static did not find it")
	}
}

// TestAnAmassedArmyPicksUpTheSubtypeItWasMissing is the "in addition
// to its other types" half through two real cards: the Bowmasters
// makes an ORC Army, and an Eternal Skylord amassing Zombies onto it
// makes it an Orc Zombie Army — which the Skylord's own static then
// grants flying.
func TestAnAmassedArmyPicksUpTheSubtypeItWasMissing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]

	castCatalogSpell(t, g, "Orcish Bowmasters", "Creature — Orc Archer", orcishBowmastersOracle, nil)
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	orcArmy := armyOf(t, g, me.ID).InstanceID

	castCatalogSpell(t, g, "Eternal Skylord", "Creature — Zombie Wizard", eternalSkylordOracle, nil)
	passPriorityAroundTable(t, g)

	army := armyOf(t, g, me.ID)
	if army.InstanceID != orcArmy {
		t.Fatal("the Skylord built a second Army instead of amassing onto the Orc one")
	}
	if !army.HasSubtype("Orc") {
		t.Error("the Army stopped being an Orc — the subtype is ADDED, not set")
	}
	if !army.HasSubtype("Zombie") {
		t.Error(`the Army did not become a Zombie ("it's also a Zombie")`)
	}
	if !game.HasKeyword(army, "flying") {
		t.Error("the now-Zombie Army has no flying")
	}
}

// --- Widespread Brutality --------------------------------------------

// TestWidespreadBrutalitySweepsWithTheAmassedArmy is CR 701.47c and
// the read side in one card: the Army it just amassed is a 2/2, it
// deals 2 to each non-Army creature, and it does not deal damage to
// itself or to any other Army.
func TestWidespreadBrutalitySweepsWithTheAmassedArmy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]

	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, PrintedPTKnown: true, Owner: opp.ID, Controller: opp.ID,
	})
	theirArmy := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Army", TypeLine: "Token Creature — Zombie Army",
		Power: 0, Toughness: 0, PrintedPTKnown: true,
		Counters: map[string]int{game.CounterPlusOne: 4},
		Owner:    opp.ID, Controller: opp.ID,
	})

	castCatalogSpell(t, g, "Widespread Brutality", "Sorcery", widespreadBrutalityOracl, nil)
	passPriorityAroundTable(t, g)

	army := armyOf(t, g, me.ID)
	if p := army.CurrentPower(); p != 2 {
		t.Fatalf("the amassed Army is a %d/%d, want 2/2", p, army.CurrentToughness())
	}
	if findBattlefieldCardByID(g, bear) != nil {
		t.Error("the 2/2 Bear survived 2 damage from the amassed Army")
	}
	other := findBattlefieldCardByID(g, theirArmy)
	if other == nil {
		t.Fatal("the opponent's Army was killed — \"each NON-Army creature\" excludes it")
	}
	if other.DamageMarked != 0 {
		t.Errorf("the opponent's Army took %d damage; it is an Army", other.DamageMarked)
	}
	if army.DamageMarked != 0 {
		t.Errorf("the amassed Army dealt %d damage to itself", army.DamageMarked)
	}
}

// --- Dreadhorde Invasion's second ability ----------------------------

// TestDreadhordeInvasionsBigArmyGainsLifelink is the card's other
// clause, which only ever fires on an Army the card itself grew: the
// Army is a Zombie TOKEN (CR 701.47a made it one), and six counters
// later it is the 6-power attacker the clause is written for.
func TestDreadhordeInvasionsBigArmyGainsLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Dreadhorde Invasion",
		TypeLine:   "Enchantment",
		OracleID:   dreadhordeInvasionOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	army := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zombie Army", TypeLine: "Token Creature — Zombie Army",
		Power: 0, Toughness: 0, PrintedPTKnown: true,
		Counters: map[string]int{game.CounterPlusOne: 6},
		Owner:    me.ID, Controller: me.ID,
	})

	attackWith(t, g, opp.ID, army)
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCardByID(g, army)
	if c == nil {
		t.Fatal("the attacking Army is gone")
	}
	if !game.HasKeyword(c, "lifelink") {
		t.Error("the 6/6 Zombie Army attacked without gaining lifelink")
	}
}
