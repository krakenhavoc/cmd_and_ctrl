package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #2166: mana you don't lose as steps and phases end. The statics
// (Upwelling, Leyline Tyrant, Kruphix) are derived from the battlefield;
// the mark (Karn, Savage Ventmaw) and the granted statement (The Last
// Agni Kai) are per-mana and per-player state.

const (
	mkUpwellingOracle = "420ff0f4-6056-4b40-9a28-fd4c23d6b81c"
	mkTyrantOracle    = "f92aaa00-6ece-4033-b3df-b2fc2c4718d9"
	mkKruphixOracle   = "b2cefcd6-4b81-479c-86ff-1695b836972c"
	mkVentmawOracle   = "f743d85f-754e-4e70-9907-af54acb9e8d4"
	mkKarnOracle      = "bc8802db-8c94-4eb4-817f-5f4f5408b7a4"
	mkAgniKaiOracle   = "fd1998c0-c18e-4310-82de-b6320be5fe10"
)

func mkAdvanceTo(t *testing.T, g *game.Game, step game.Step) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if g.Turn.Step == step {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("never reached %s", step)
}

func mkAdvance(t *testing.T, g *game.Game) {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
}

func mkAdd(t *testing.T, g *game.Game, p *game.Player, produced string, opts game.AddManaOptions) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddManaWithOptionsForEffect(p.ID, uuid.Nil, produced, opts); err != nil {
			t.Fatalf("AddMana: %v", err)
		}
	})
}

func mkColors(p *game.Player) map[string]int {
	out := map[string]int{}
	for _, tok := range p.ManaPool {
		out[tok.Color]++
	}
	return out
}

func mkWantPool(t *testing.T, p *game.Player, want map[string]int) {
	t.Helper()
	got := mkColors(p)
	if len(got) != len(want) {
		t.Fatalf("pool = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("pool = %v, want %v", got, want)
		}
	}
}

// Control: with no keep rule the pool empties at the next step, which
// is what every test below departs from.
func TestManaStillEmptiesAtTheNextStepByDefault(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	mkAdd(t, g, me, "{R}{G}", game.AddManaOptions{})
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{})
}

func TestUpwellingKeepsEveryPlayersManaAndLeavingEndsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	wall := pushCatalogPermanent(g, opp.ID, "Upwelling", "Enchantment", mkUpwellingOracle, false)
	mkAdd(t, g, me, "{R}{C}", game.AddManaOptions{})
	mkAdd(t, g, opp, "{G}", game.AddManaOptions{})
	mkAdvance(t, g)
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{"R": 1, "C": 1})
	mkWantPool(t, opp, map[string]int{"G": 1})

	removeFromBattlefield(g, wall)
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{})
	mkWantPool(t, opp, map[string]int{})
}

func TestLeylineTyrantKeepsOnlyRedForItsController(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	tyrant := pushCatalogPermanent(g, me.ID, "Leyline Tyrant", "Creature — Dragon", mkTyrantOracle, false)
	mkAdd(t, g, me, "{R}{R}{G}{C}", game.AddManaOptions{})
	mkAdd(t, g, opp, "{R}", game.AddManaOptions{})
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{"R": 2})
	mkWantPool(t, opp, map[string]int{})

	removeFromBattlefield(g, tyrant)
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{})
}

func TestKruphixConvertsLostManaToColorlessAndKeepsRestrictions(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	god := pushCatalogPermanent(g, me.ID, "Kruphix, God of Horizons", "Legendary Enchantment Creature — God", mkKruphixOracle, false)
	mkAdd(t, g, me, "{R}", game.AddManaOptions{Restrictions: []string{game.ManaRestrictNotNonartifactSpell}})
	mkAdd(t, g, me, "{G}{U}", game.AddManaOptions{})
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{"C": 3})
	restricted := 0
	for _, tok := range me.ManaPool {
		if len(tok.Restrictions) == 1 && tok.Restrictions[0] == game.ManaRestrictNotNonartifactSpell {
			restricted++
		}
	}
	if restricted != 1 {
		t.Errorf("restricted tokens = %d, want the one Karn-style mana to keep its restriction", restricted)
	}
	// Colourless mana is "lost" again at the next boundary and becomes
	// colourless again: it is kept for as long as Kruphix is out.
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{"C": 3})

	removeFromBattlefield(g, god)
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{})
}

func TestMarkedManaSurvivesStepsAndExpiresAtCleanup(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	mkAdd(t, g, me, "{R}{R}{G}", game.AddManaOptions{Riders: []game.ManaSpendRider{KeepManaUntilEndOfTurn()}})
	mkAdd(t, g, me, "{B}", game.AddManaOptions{})
	mkAdvance(t, g)
	// The unmarked {B} went at the first boundary; the marked mana stays.
	mkWantPool(t, me, map[string]int{"R": 2, "G": 1})
	mkAdvanceToEnd := func() { mkAdvanceTo(t, g, game.StepEnd) }
	mkAdvanceToEnd()
	mkWantPool(t, me, map[string]int{"R": 2, "G": 1})
	// The cleanup step ends "until end of turn": the mark goes and the
	// mana is lost like any other.
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{})
}

func TestMarkedManaIsConvertedByKruphixWhenTheMarkExpires(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Kruphix, God of Horizons", "Legendary Enchantment Creature — God", mkKruphixOracle, false)
	mkAdd(t, g, me, "{R}", game.AddManaOptions{Riders: []game.ManaSpendRider{KeepManaUntilEndOfTurn()}})
	mkAdvanceTo(t, g, game.StepEnd)
	mkWantPool(t, me, map[string]int{"R": 1})
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{"C": 1})
	for _, tok := range me.ManaPool {
		if tok.Riders != nil {
			t.Errorf("expired mark still on the token: %+v", tok.Riders)
		}
	}
}

func TestSavageVentmawsManaIsMarkedAndKept(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Savage Ventmaw", "Creature — Dragon", mkVentmawOracle, false)
	spec, ok := Lookup(mkVentmawOracle)
	if !ok || len(spec.Triggered) != 1 {
		t.Fatal("Savage Ventmaw is not registered with its attack trigger")
	}
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: me.ID, SourceCardID: uuid.Nil}
		if err := spec.Triggered[0].Effect(g, item); err != nil {
			t.Fatalf("trigger effect: %v", err)
		}
	})
	mkWantPool(t, me, map[string]int{"R": 3, "G": 3})
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{"R": 3, "G": 3})
	mkAdvanceTo(t, g, game.StepEnd)
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{})
}

func TestKarnSizeAndRestrictedKeptUpkeepMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	karn := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Karn, Legacy Reforged", OracleID: mkKarnOracle,
		TypeLine: "Legendary Artifact Creature — Golem", ManaCost: "{5}",
		Owner: me.ID, Controller: me.ID,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{7}",
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCardForTest(g, karn)
	if c.CurrentPower() != 7 || c.CurrentToughness() != 7 {
		t.Errorf("Karn is %d/%d, want 7/7 (greatest mana value among artifacts)", c.CurrentPower(), c.CurrentToughness())
	}

	// Drive the upkeep trigger's effect: two artifacts, so {C}{C}.
	spec, _ := Lookup(mkKarnOracle)
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: me.ID, SourceCardID: karn}
		if err := spec.Triggered[0].Effect(g, item); err != nil {
			t.Fatalf("trigger effect: %v", err)
		}
	})
	mkWantPool(t, me, map[string]int{"C": 2})
	for _, tok := range me.ManaPool {
		if len(tok.Restrictions) != 1 || tok.Restrictions[0] != game.ManaRestrictNotNonartifactSpell {
			t.Errorf("Karn mana restrictions = %v", tok.Restrictions)
		}
	}
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{"C": 2})
	mkAdvanceTo(t, g, game.StepEnd)
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{})
}

func TestTheLastAgniKaiFightsAddsExcessRedAndKeepsRedUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Brute", TypeLine: "Creature — Ogre",
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Squire", TypeLine: "Creature — Human",
		Power: 1, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	mkAdd(t, g, me, "{G}", game.AddManaOptions{})
	castCatalogSpell(t, g, "The Last Agni Kai", "Instant", mkAgniKaiOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	})
	passPriorityAroundTable(t, g)
	// 5 damage to a 2-toughness creature: 3 excess, so {R}{R}{R}.
	mkWantPool(t, me, map[string]int{"R": 3, "G": 1})
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{"R": 3})
	mkAdvanceTo(t, g, game.StepEnd)
	mkWantPool(t, me, map[string]int{"R": 3})
	mkAdvance(t, g)
	mkWantPool(t, me, map[string]int{})
}

func TestKeptManaRoundTripsThroughASnapshot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	mkAdd(t, g, me, "{G}{G}", game.AddManaOptions{Riders: []game.ManaSpendRider{KeepManaUntilEndOfTurn()}})
	g.WithWriteLock(func() {
		g.GrantKeepManaForEffect(me.ID, []string{"R"}, "test grant", uuid.Nil, g.UntilEndOfTurnDuration())
	})
	mkAdd(t, g, me, "{R}", game.AddManaOptions{})

	restored := restoreThroughJSON(t, g)
	rme := restored.Seats[g.Turn.ActiveSeat]
	mkWantPool(t, rme, map[string]int{"G": 2, "R": 1})
	marked := 0
	for _, tok := range rme.ManaPool {
		if tok.Color == "G" && len(tok.Riders) == 1 && tok.Riders[0].Kind == game.ManaRiderKeepUntilEndOfTurn {
			marked++
		}
	}
	if marked != 2 {
		t.Fatalf("marked tokens after restore = %d, want 2", marked)
	}
	mkAdvance(t, restored)
	mkWantPool(t, rme, map[string]int{"G": 2, "R": 1})
	mkAdvanceTo(t, restored, game.StepEnd)
	mkAdvance(t, restored)
	mkWantPool(t, rme, map[string]int{})

	// And the undo path: a clone keeps the mark too.
	clone := g.Snapshot()
	if got := mkColors(clone.Seats[g.Turn.ActiveSeat]); got["G"] != 2 {
		t.Errorf("clone pool = %v", got)
	}
}

// The bot and the client's timing lookup read the same pool, so mana
// kept across a step is mana they can spend.
func TestEnumeratorOffersACastPaidWithKeptMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	spell := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: spell, Name: "Test Bolt", TypeLine: "Instant", ManaCost: "{R}",
		Owner: me.ID, Controller: me.ID,
	})
	mkAdd(t, g, me, "{R}", game.AddManaOptions{Riders: []game.ManaSpendRider{KeepManaUntilEndOfTurn()}})
	mkAdvance(t, g)
	offered := func() bool {
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type == "cast_spell" && m.Source == spell {
				return true
			}
		}
		return false
	}
	if !offered() {
		t.Error("the kept {R} should make the {R} instant castable")
	}
	mkAdvanceTo(t, g, game.StepEnd)
	mkAdvance(t, g)
	if len(me.ManaPool) != 0 {
		t.Fatalf("pool = %v after cleanup", me.ManaPool)
	}
}
