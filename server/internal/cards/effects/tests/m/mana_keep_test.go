package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// Control: with no keep rule the pool empties at the next step, which
// is what every test below departs from.
func TestManaStillEmptiesAtTheNextStepByDefault(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.MkAdd(t, g, me, "{R}{G}", game.AddManaOptions{})
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{})
}

func TestUpwellingKeepsEveryPlayersManaAndLeavingEndsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	wall := effectstest.PushCatalogPermanent(g, opp.ID, "Upwelling", "Enchantment", effectstest.MkUpwellingOracle, false)
	effectstest.MkAdd(t, g, me, "{R}{C}", game.AddManaOptions{})
	effectstest.MkAdd(t, g, opp, "{G}", game.AddManaOptions{})
	effectstest.MkAdvance(t, g)
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 1, "C": 1})
	effectstest.MkWantPool(t, opp, map[string]int{"G": 1})

	effectstest.RemoveFromBattlefield(g, wall)
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{})
	effectstest.MkWantPool(t, opp, map[string]int{})
}

func TestLeylineTyrantKeepsOnlyRedForItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	tyrant := effectstest.PushCatalogPermanent(g, me.ID, "Leyline Tyrant", "Creature — Dragon", effectstest.MkTyrantOracle, false)
	effectstest.MkAdd(t, g, me, "{R}{R}{G}{C}", game.AddManaOptions{})
	effectstest.MkAdd(t, g, opp, "{R}", game.AddManaOptions{})
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 2})
	effectstest.MkWantPool(t, opp, map[string]int{})

	effectstest.RemoveFromBattlefield(g, tyrant)
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{})
}

func TestKruphixConvertsLostManaToColorlessAndKeepsRestrictions(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	god := effectstest.PushCatalogPermanent(g, me.ID, "Kruphix, God of Horizons", "Legendary Enchantment Creature — God", effectstest.MkKruphixOracle, false)
	effectstest.MkAdd(t, g, me, "{R}", game.AddManaOptions{Restrictions: []string{game.ManaRestrictNotNonartifactSpell}})
	effectstest.MkAdd(t, g, me, "{G}{U}", game.AddManaOptions{})
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"C": 3})
	restricted := 0
	for _, tok := range me.ManaPool.Units() {
		if len(tok.Restrictions) == 1 && tok.Restrictions[0] == game.ManaRestrictNotNonartifactSpell {
			restricted++
		}
	}
	if restricted != 1 {
		t.Errorf("restricted tokens = %d, want the one Karn-style mana to keep its restriction", restricted)
	}
	// Colourless mana is "lost" again at the next boundary and becomes
	// colourless again: it is kept for as long as Kruphix is out.
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"C": 3})

	effectstest.RemoveFromBattlefield(g, god)
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{})
}

func TestMarkedManaSurvivesStepsAndExpiresAtCleanup(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.MkAdd(t, g, me, "{R}{R}{G}", game.AddManaOptions{Riders: []game.ManaSpendRider{effects.KeepManaUntilEndOfTurn()}})
	effectstest.MkAdd(t, g, me, "{B}", game.AddManaOptions{})
	effectstest.MkAdvance(t, g)
	// The unmarked {B} went at the first boundary; the marked mana stays.
	effectstest.MkWantPool(t, me, map[string]int{"R": 2, "G": 1})
	mkAdvanceToEnd := func() { effectstest.MkAdvanceTo(t, g, game.StepEnd) }
	mkAdvanceToEnd()
	effectstest.MkWantPool(t, me, map[string]int{"R": 2, "G": 1})
	// The cleanup step ends "until end of turn": the mark goes and the
	// mana is lost like any other.
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{})
}

func TestMarkedManaIsConvertedByKruphixWhenTheMarkExpires(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Kruphix, God of Horizons", "Legendary Enchantment Creature — God", effectstest.MkKruphixOracle, false)
	effectstest.MkAdd(t, g, me, "{R}", game.AddManaOptions{Riders: []game.ManaSpendRider{effects.KeepManaUntilEndOfTurn()}})
	effectstest.MkAdvanceTo(t, g, game.StepEnd)
	effectstest.MkWantPool(t, me, map[string]int{"R": 1})
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"C": 1})
	for _, tok := range me.ManaPool.Units() {
		if tok.Riders != nil {
			t.Errorf("expired mark still on the token: %+v", tok.Riders)
		}
	}
}

func TestSavageVentmawsManaIsMarkedAndKept(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Savage Ventmaw", "Creature — Dragon", effectstest.MkVentmawOracle, false)
	spec, ok := effects.Lookup(effectstest.MkVentmawOracle)
	if !ok || len(spec.Triggered) != 1 {
		t.Fatal("Savage Ventmaw is not registered with its attack trigger")
	}
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: me.ID, SourceCardID: uuid.Nil}
		if err := spec.Triggered[0].Effect(g, item); err != nil {
			t.Fatalf("trigger effect: %v", err)
		}
	})
	effectstest.MkWantPool(t, me, map[string]int{"R": 3, "G": 3})
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 3, "G": 3})
	effectstest.MkAdvanceTo(t, g, game.StepEnd)
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{})
}

func TestKarnSizeAndRestrictedKeptUpkeepMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	karn := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Karn, Legacy Reforged", OracleID: effectstest.MkKarnOracle,
		TypeLine: "Legendary Artifact Creature — Golem", ManaCost: "{5}",
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{7}",
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.FindBattlefieldCardForTest(g, karn)
	if c.CurrentPower() != 7 || c.CurrentToughness() != 7 {
		t.Errorf("Karn is %d/%d, want 7/7 (greatest mana value among artifacts)", c.CurrentPower(), c.CurrentToughness())
	}

	// Drive the upkeep trigger's effect: two artifacts, so {C}{C}.
	spec, _ := effects.Lookup(effectstest.MkKarnOracle)
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: me.ID, SourceCardID: karn}
		if err := spec.Triggered[0].Effect(g, item); err != nil {
			t.Fatalf("trigger effect: %v", err)
		}
	})
	effectstest.MkWantPool(t, me, map[string]int{"C": 2})
	for _, tok := range me.ManaPool.Units() {
		if len(tok.Restrictions) != 1 || tok.Restrictions[0] != game.ManaRestrictNotNonartifactSpell {
			t.Errorf("Karn mana restrictions = %v", tok.Restrictions)
		}
	}
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"C": 2})
	effectstest.MkAdvanceTo(t, g, game.StepEnd)
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{})
}

func TestTheLastAgniKaiFightsAddsExcessRedAndKeepsRedUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Brute", TypeLine: "Creature — Ogre",
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Squire", TypeLine: "Creature — Human",
		Power: 1, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.MkAdd(t, g, me, "{G}", game.AddManaOptions{})
	effectstest.CastCatalogSpell(t, g, "The Last Agni Kai", "Instant", effectstest.MkAgniKaiOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	})
	effectstest.PassPriorityAroundTable(t, g)
	// 5 damage to a 2-toughness creature: 3 excess, so {R}{R}{R}.
	effectstest.MkWantPool(t, me, map[string]int{"R": 3, "G": 1})
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 3})
	effectstest.MkAdvanceTo(t, g, game.StepEnd)
	effectstest.MkWantPool(t, me, map[string]int{"R": 3})
	effectstest.MkAdvance(t, g)
	effectstest.MkWantPool(t, me, map[string]int{})
}

func TestKeptManaRoundTripsThroughASnapshot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.MkAdd(t, g, me, "{G}{G}", game.AddManaOptions{Riders: []game.ManaSpendRider{effects.KeepManaUntilEndOfTurn()}})
	g.WithWriteLock(func() {
		g.GrantKeepManaForEffect(me.ID, []string{"R"}, "test grant", uuid.Nil, g.UntilEndOfTurnDuration())
	})
	effectstest.MkAdd(t, g, me, "{R}", game.AddManaOptions{})

	restored := effectstest.RestoreThroughJSON(t, g)
	rme := restored.Seats[g.Turn.ActiveSeat]
	effectstest.MkWantPool(t, rme, map[string]int{"G": 2, "R": 1})
	marked := 0
	for _, tok := range rme.ManaPool.Units() {
		if tok.Color == "G" && len(tok.Riders) == 1 && tok.Riders[0].Kind == game.ManaRiderKeepUntilEndOfTurn {
			marked++
		}
	}
	if marked != 2 {
		t.Fatalf("marked tokens after restore = %d, want 2", marked)
	}
	effectstest.MkAdvance(t, restored)
	effectstest.MkWantPool(t, rme, map[string]int{"G": 2, "R": 1})
	effectstest.MkAdvanceTo(t, restored, game.StepEnd)
	effectstest.MkAdvance(t, restored)
	effectstest.MkWantPool(t, rme, map[string]int{})

	// And the undo path: a clone keeps the mark too.
	clone := g.Snapshot()
	if got := effectstest.MkColors(clone.Seats[g.Turn.ActiveSeat]); got["G"] != 2 {
		t.Errorf("clone pool = %v", got)
	}
}

// The bot and the client's timing lookup read the same pool, so mana
// kept across a step is mana they can spend.
func TestEnumeratorOffersACastPaidWithKeptMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	spell := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: spell, Name: "Test Bolt", TypeLine: "Instant", ManaCost: "{R}",
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.MkAdd(t, g, me, "{R}", game.AddManaOptions{Riders: []game.ManaSpendRider{effects.KeepManaUntilEndOfTurn()}})
	effectstest.MkAdvance(t, g)
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
	effectstest.MkAdvanceTo(t, g, game.StepEnd)
	effectstest.MkAdvance(t, g)
	if me.ManaPool.Total() != 0 {
		t.Fatalf("pool = %v after cleanup", me.ManaPool)
	}
}
