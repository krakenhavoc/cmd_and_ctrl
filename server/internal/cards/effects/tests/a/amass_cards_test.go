package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- ArmyToken, the template the keyword derives ---------------------

// TestArmyTokenIsTheKeywordsPrintedToken pins the derivation against
// CR 701.47a's words: "a 0/0 black [subtype] Army creature token". The
// template is built from the subtype rather than looked up in
// tokens_table.go (ADR 0087 decision 2), so this is the only thing
// standing between a typo and a colourless 1/1.
func TestArmyTokenIsTheKeywordsPrintedToken(t *testing.T) {
	for _, subtype := range []string{"Zombie", "Orc", "Goblin", "Sliver"} {
		tok := effects.ArmyToken(subtype)
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
	if got := effects.ArmyToken("").TypeLine; got != "Token Creature — Zombie Army" {
		t.Errorf("a subtype-less amass makes %q, want the CR 701.47d Zombie Army", got)
	}
}

// --- Orcish Bowmasters ------------------------------------------------

// TestOrcishBowmastersShootsAndAmassesOnEntry is the card's first
// trigger condition: it enters, pings a target, and amasses Orcs 1.
func TestOrcishBowmastersShootsAndAmassesOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	before := opp.Life

	effectstest.CastCatalogSpell(t, g, "Orcish Bowmasters", "Creature — Orc Archer", effectstest.OrcishBowmastersOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != before-1 {
		t.Errorf("opponent life = %d, want %d — the Bowmasters deal 1 damage", opp.Life, before-1)
	}
	army := effectstest.ArmyOf(t, g, me.ID)
	if !army.HasSubtype("Orc") {
		t.Errorf("the Army is %q, want an Orc Army", army.TypeLine)
	}
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

// TestOrcishBowmastersGrowsTheSameArmyOnEveryDraw is the card's second
// trigger condition and the whole reason amass needed find-or-create:
// a wheel is one Army with N counters on it, not N 0/0 Armies.
func TestOrcishBowmastersGrowsTheSameArmyOnEveryDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]

	effectstest.CastCatalogSpell(t, g, "Orcish Bowmasters", "Creature — Orc Archer", effectstest.OrcishBowmastersOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	first := effectstest.ArmyOf(t, g, me.ID).InstanceID

	// Two draws by an opponent, neither in their own draw step (it is
	// still the active player's turn).
	for i := 0; i < 2; i++ {
		if err := g.DrawCard(opp.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}

	army := effectstest.ArmyOf(t, g, me.ID)
	if army.InstanceID != first {
		t.Error("a second Army was created; amass must find the one already out")
	}
	if got := effectstest.CountersOn(g, first, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3 — one per trigger", got)
	}
	if p := army.CurrentPower(); p != 3 {
		t.Errorf("the Army is a %d/%d, want 3/3", p, army.CurrentToughness())
	}
}

// Only the first completed draw in the opponent's draw step is exempt.
func TestOrcishBowmastersIgnoresTheDrawStepDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]

	effectstest.CastCatalogSpell(t, g, "Orcish Bowmasters", "Creature — Orc Archer", effectstest.OrcishBowmastersOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	before := effectstest.CountersOn(g, effectstest.ArmyOf(t, g, me.ID).InstanceID, game.CounterPlusOne)

	// Walk to the next player's draw step; the turn-based draw is
	// theirs, in their own draw step.
	effectstest.AdvanceToDrawStepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountersOn(g, effectstest.ArmyOf(t, g, me.ID).InstanceID, game.CounterPlusOne); got != before {
		t.Errorf("+1/+1 counters = %d, want %d — the draw-step draw is exempt", got, before)
	}
}

// --- Dreadhorde Invasion ---------------------------------------------

// TestDreadhordeInvasionMakesOneArmyAndGrowsIt is find-or-create
// across turns, on the card that makes it visible: the first upkeep
// mints the Army, the second puts a second counter on the same one,
// and the controller loses a life each time.
func TestDreadhordeInvasionMakesOneArmyAndGrowsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Dreadhorde Invasion",
		TypeLine:   "Enchantment",
		OracleID:   effectstest.DreadhordeInvasionOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	life := me.Life

	effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	first := effectstest.ArmyOf(t, g, me.ID)
	if !first.HasSubtype("Zombie") {
		t.Errorf("the Army is %q, want a Zombie Army", first.TypeLine)
	}
	if me.Life != life-1 {
		t.Errorf("life = %d, want %d", me.Life, life-1)
	}
	firstID := first.InstanceID

	effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	second := effectstest.ArmyOf(t, g, me.ID)
	if second.InstanceID != firstID {
		t.Error("the second upkeep built a new Army instead of growing the first")
	}
	if got := effectstest.CountersOn(g, firstID, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if me.Life != life-2 {
		t.Errorf("life = %d, want %d", me.Life, life-2)
	}
}

// --- Eternal Skylord --------------------------------------------------

// TestEternalSkylordsArmyFlies is CR 701.47a's last sentence reaching
// a static on another card: the Army is a Zombie TOKEN, so "Zombie
// tokens you control have flying" finds it. Nothing in the Skylord's
// file knows what an Army is.
func TestEternalSkylordsArmyFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	effectstest.CastCatalogSpell(t, g, "Eternal Skylord", "Creature — Zombie Wizard", effectstest.EternalSkylordOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	army := effectstest.ArmyOf(t, g, me.ID)
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 2 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]

	effectstest.CastCatalogSpell(t, g, "Orcish Bowmasters", "Creature — Orc Archer", effectstest.OrcishBowmastersOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	orcArmy := effectstest.ArmyOf(t, g, me.ID).InstanceID

	effectstest.CastCatalogSpell(t, g, "Eternal Skylord", "Creature — Zombie Wizard", effectstest.EternalSkylordOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	army := effectstest.ArmyOf(t, g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]

	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, PrintedPTKnown: true, Owner: opp.ID, Controller: opp.ID,
	})
	theirArmy := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Army", TypeLine: "Token Creature — Zombie Army",
		Power: 0, Toughness: 0, PrintedPTKnown: true,
		Counters: map[string]int{game.CounterPlusOne: 4},
		Owner:    opp.ID, Controller: opp.ID,
	})

	effectstest.CastCatalogSpell(t, g, "Widespread Brutality", "Sorcery", effectstest.WidespreadBrutalityOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)

	army := effectstest.ArmyOf(t, g, me.ID)
	if p := army.CurrentPower(); p != 2 {
		t.Fatalf("the amassed Army is a %d/%d, want 2/2", p, army.CurrentToughness())
	}
	if effectstest.FindBattlefieldCardByID(g, bear) != nil {
		t.Error("the 2/2 Bear survived 2 damage from the amassed Army")
	}
	other := effectstest.FindBattlefieldCardByID(g, theirArmy)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Dreadhorde Invasion",
		TypeLine:   "Enchantment",
		OracleID:   effectstest.DreadhordeInvasionOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	army := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zombie Army", TypeLine: "Token Creature — Zombie Army",
		Power: 0, Toughness: 0, PrintedPTKnown: true,
		Counters: map[string]int{game.CounterPlusOne: 6},
		Owner:    me.ID, Controller: me.ID,
	})

	effectstest.AttackWith(t, g, opp.ID, army)
	effectstest.PassPriorityAroundTable(t, g)

	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.FindBattlefieldCardByID(g, army)
	if c == nil {
		t.Fatal("the attacking Army is gone")
	}
	if !game.HasKeyword(c, "lifelink") {
		t.Error("the 6/6 Zombie Army attacked without gaining lifelink")
	}
}
