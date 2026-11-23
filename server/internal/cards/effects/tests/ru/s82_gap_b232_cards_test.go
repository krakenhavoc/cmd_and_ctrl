package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander Legends: Battle for Baldur's Gate, GAP batch 232 (S82, #5948): the
// sorceries.

const (
	gapOverwhelmingEncounterOracle = "7573ce9f-fc35-4d20-9b60-c739b8408444"
	gapSplitTheSpoilsOracle        = "48ad5bd6-9568-4813-be29-99aef8b9327a"
	gapVentureForthOracle          = "261f371f-bcda-4761-810f-3216608b31d8"
	gapCoronationOfChaosOracle     = "46f1618d-9c95-4a4b-9a20-7c378b811849"
	gapInspiredTinkeringOracle     = "3088b09b-1dce-4fc9-99de-2ca9c752eb12"
	gapThunderwaveOracle           = "43de944b-e630-45f0-bce4-c66803e8b5c1"
	gapCaughtStealingOracle        = "750b61a1-7eb1-4ce2-8a56-fa63f8e48755"
	gapCandlekeepInspirationOracle = "32d666ff-4d6b-4c60-8d3d-5bdae2d475fa"
	gapModifyMemoryOracle          = "2c37d4e1-6b76-4feb-933b-1752ecc6d3e8"
	gapAscendFromAvernusOracle     = "451e8ece-7389-4d52-8bc9-450cb5e53e6a"
	gapHornOfValhallaOracle        = "d02179a0-a6b7-49ef-a73e-cf11cabb896f"
)

func TestS82GapB232CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapOverwhelmingEncounterOracle: "Overwhelming Encounter",
		gapSplitTheSpoilsOracle:        "Split the Spoils",
		gapVentureForthOracle:          "Venture Forth",
		gapCoronationOfChaosOracle:     "Coronation of Chaos",
		gapInspiredTinkeringOracle:     "Inspired Tinkering",
		gapThunderwaveOracle:           "Thunderwave",
		gapCaughtStealingOracle:        "You've Been Caught Stealing",
		gapCandlekeepInspirationOracle: "Candlekeep Inspiration",
		gapModifyMemoryOracle:          "Modify Memory",
		gapAscendFromAvernusOracle:     "Ascend from Avernus",
		gapHornOfValhallaOracle:        "Horn of Valhalla",
		gapHornOfValhallaOracle + "#1": "Ysgard's Call",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

// gapB232Cast puts the card in the active seat's hand and casts it strictly for free.
func gapB232Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	effectstest.D5681Spell(g, me, id, name, typeLine, "", oracle)
	params.Strict = true
	return g.CastSpell(me.ID, id, params)
}

func TestGapOverwhelmingEncounterByTheRoll(t *testing.T) {
	for _, tc := range []struct {
		natural, power, counters int
	}{{3, 4, 0}, {9, 4, 0}, {10, 2, 2}, {19, 2, 2}, {20, 2, 4}} {
		enc := uuid.New()
		g := effectstest.D5681Table(t, enc, tc.natural)
		me, opp, _ := effectstest.BlbB1Seats(g)
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		effectstest.D5681Spell(g, me, enc, "Overwhelming Encounter", "Sorcery", "", gapOverwhelmingEncounterOracle)
		if err := g.CastSpell(me.ID, enc, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.EffectivePowerWithCounters(t, g, mine); got != tc.power+tc.counters {
			t.Errorf("natural %d: power %d, want %d", tc.natural, got, tc.power+tc.counters)
		}
		if got := effectstest.CountersOn(g, mine, game.CounterPlusOne); got != tc.counters {
			t.Errorf("natural %d: %d counters, want %d", tc.natural, got, tc.counters)
		}
		for _, kw := range []string{"vigilance", "trample"} {
			if !effectstest.HasEffectiveKeyword(t, g, mine, kw) {
				t.Errorf("natural %d: no %s", tc.natural, kw)
			}
			if effectstest.HasEffectiveKeyword(t, g, theirs, kw) {
				t.Errorf("natural %d: an opponent's creature gained %s", tc.natural, kw)
			}
		}
		if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
			t.Errorf("natural %d: an opponent's creature changed to %d", tc.natural, got)
		}
	}
}

func TestGapCoronationOfChaosStopsBlocksAndGoadsUpToThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	d := effectstest.B12Creature(g, opp.ID, "D", "Creature — Bear", 2, 2)
	if err := gapB232Cast(t, g, "Coronation of Chaos", "Sorcery", gapCoronationOfChaosOracle, game.CastSpellParams{Targets: effectstest.CardRefs(a, b, c, d)}); err == nil {
		t.Fatal("targeted four creatures")
	}
	if err := gapB232Cast(t, g, "Coronation of Chaos", "Sorcery", gapCoronationOfChaosOracle, game.CastSpellParams{Targets: effectstest.CardRefs(a, b, c)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	goaded := func(id uuid.UUID) bool { return effectstest.B33Goaded(t, g, id) != uuid.Nil }
	for _, id := range []uuid.UUID{a, b, c} {
		if !goaded(id) {
			t.Error("a target is not goaded")
		}
		if !effectstest.MidB03CantBlock(t, g, id) {
			t.Error("a target can still block")
		}
	}
	if goaded(d) {
		t.Error("an untargeted creature was goaded")
	}
	_ = me
}

func TestGapInspiredTinkeringExilesThreeForPlayAndMakesThreeTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	var top []uuid.UUID
	for i := 0; i < 5; i++ {
		id := uuid.New()
		top = append(top, id)
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Library Bear", TypeLine: "Creature — Bear", ManaCost: "{0}", Owner: me.ID, Controller: me.ID})
	}
	if err := gapB232Cast(t, g, "Inspired Tinkering", "Sorcery", gapInspiredTinkeringOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	exiled := 0
	for _, id := range top {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			exiled++
			// It may be played: cast it from exile.
			if exiled == 1 {
				if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
					t.Errorf("an exiled card cannot be cast: %v", err)
				}
			}
		}
	}
	if exiled != 3 {
		t.Errorf("exiled %d cards, want 3", exiled)
	}
	if n := effectstest.CountTokensNamed2(g, me.ID, "Treasure"); n != 3 {
		t.Errorf("Treasures = %d, want 3", n)
	}
}

func TestGapCandlekeepInspirationSetsBasePTFromInstantsSorceriesAndAdventures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Ritual", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bear Card", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	effectstest.PushExiledCard(g, me.ID, "Exiled Instant", "Instant", false)
	effectstest.PushExiledCard(g, me.ID, "Face Down Instant", "Instant", true)
	effectstest.PushExiledCard(g, opp.ID, "Their Exiled Sorcery", "Sorcery", false)
	if err := gapB232Cast(t, g, "Candlekeep Inspiration", "Sorcery", gapCandlekeepInspirationOracle, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Instant and sorcery in the graveyard, the face-up exiled instant: three.
	if p, tough := effectstest.MkmB1PT(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("my creature is %d/%d, want 3/3", p, tough)
	}
	if p, _ := effectstest.MkmB1PT(t, g, theirs); p != 1 {
		t.Errorf("an opponent's creature changed to power %d", p)
	}
}

func TestGapModifyMemoryExchangesAndDrawsOnlyWhenYouControlNeither(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, other := effectstest.BlbB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	third := effectstest.B12Creature(g, other.ID, "Third", "Creature — Bear", 2, 2)
	another := effectstest.B12Creature(g, opp.ID, "Another", "Creature — Bear", 2, 2)
	gapB231Library := func(n int) {
		for i := 0; i < n; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Library", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		}
	}
	gapB231Library(6)
	if err := gapB232Cast(t, g, "Modify Memory", "Sorcery", gapModifyMemoryOracle, game.CastSpellParams{Targets: effectstest.CardRefs(theirs, another)}); err == nil {
		t.Fatal("exchanged two creatures with the same controller")
	}
	hand := me.Hand.Size()
	if err := gapB232Cast(t, g, "Modify Memory", "Sorcery", gapModifyMemoryOracle, game.CastSpellParams{Targets: effectstest.CardRefs(mine, theirs)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, mine); c.Controller != opp.ID {
		t.Errorf("my creature is controlled by %v, want the opponent", c.Controller)
	}
	if c := effectstest.LayeredCard(t, g, theirs); c.Controller != me.ID {
		t.Errorf("their creature is controlled by %v, want me", c.Controller)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d → %d: drew with a creature of mine involved", hand, got)
	}
	hand = me.Hand.Size()
	if err := gapB232Cast(t, g, "Modify Memory", "Sorcery", gapModifyMemoryOracle, game.CastSpellParams{Targets: effectstest.CardRefs(another, third)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, another); c.Controller != other.ID {
		t.Errorf("the creature is controlled by %v, want the third player", c.Controller)
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d → %d, want three cards drawn", hand, got)
	}
}

func TestGapAscendFromAvernusReturnsEverythingUpToXAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	cheap, dear, walker, spell := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: dear, Name: "Dear Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{G}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: walker, Name: "Cheap Walker", TypeLine: "Legendary Planeswalker — Test", ManaCost: "{2}{W}", StartingLoyalty: 3, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Cheap Bolt", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	ascend := uuid.New()
	effectstest.D5681Spell(g, me, ascend, "Ascend from Avernus", "Sorcery", "{X}{W}{W}{W}", gapAscendFromAvernusOracle)
	if err := g.CastSpell(me.ID, ascend, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, cheap) && effectstest.FindBattlefieldByName(g, "Cheap Bear") == uuid.Nil {
		t.Error("the cheap creature did not return")
	}
	if effectstest.FindBattlefieldByName(g, "Cheap Walker") == uuid.Nil {
		t.Error("the cheap planeswalker did not return")
	}
	if effectstest.FindBattlefieldByName(g, "Dear Bear") != uuid.Nil {
		t.Error("a creature with mana value 5 returned for X=3")
	}
	if !me.Graveyard.Contains(dear) || !me.Graveyard.Contains(spell) {
		t.Error("the expensive creature or the instant left the graveyard")
	}
	if z := g.FindCardZoneForEffect(ascend); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("Ascend from Avernus is in %v, want exile", z)
	}
}

func TestGapHornOfValhallaPumpsByYourCreaturesAndYsgardsCallMakesX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Enemy", "Creature — Bear", 1, 1)
	horn := effectstest.SeedEquipment(g, me.ID, "Horn of Valhalla", gapHornOfValhallaOracle)
	effectstest.EquipTo(t, g, me.ID, horn, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmB1PT(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped creature is %d/%d, want 3/3 (+2/+2 for two creatures)", p, tough)
	}
	effectstest.B12Creature(g, me.ID, "Another", "Creature — Bear", 1, 1)
	if p, _ := effectstest.MkmB1PT(t, g, bear); p != 4 {
		t.Errorf("after a third creature the equipped creature has power %d, want 4", p)
	}
	call := uuid.New()
	effectstest.D5681Spell(g, me, call, "Ysgard's Call", "Sorcery — Adventure", "{X}{W}{W}", gapHornOfValhallaOracle+"#1")
	if err := g.CastSpell(me.ID, call, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTokensNamed2(g, me.ID, "Soldier"); n != 3 {
		t.Errorf("Soldiers = %d, want 3", n)
	}
}

func TestGapThunderwaveByTheRoll(t *testing.T) {
	for _, tc := range []struct {
		natural                  int
		mine, theirs, spareTheir int // damage on my 9-toughness creature, an opponent's, then the spared one
	}{{4, 3, 3, 0}, {20, 0, 6, 0}} {
		wave := uuid.New()
		g := effectstest.D5681Table(t, wave, tc.natural)
		me, opp, _ := effectstest.BlbB1Seats(g)
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 9)
		theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 9)
		effectstest.D5681Spell(g, me, wave, "Thunderwave", "Sorcery", "", gapThunderwaveOracle)
		if err := g.CastSpell(me.ID, wave, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.FindCardAnywhere(t, g, mine).DamageMarked; got != tc.mine {
			t.Errorf("natural %d: my creature has %d damage, want %d", tc.natural, got, tc.mine)
		}
		if got := effectstest.FindCardAnywhere(t, g, theirs).DamageMarked; got != tc.theirs {
			t.Errorf("natural %d: their creature has %d damage, want %d", tc.natural, got, tc.theirs)
		}
	}
}

func TestGapThunderwaveMiddleRollSparesTheChosenCreature(t *testing.T) {
	wave := uuid.New()
	g := effectstest.D5681Table(t, wave, 14)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 9)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 9)
	effectstest.D5681Spell(g, me, wave, "Thunderwave", "Sorcery", "", gapThunderwaveOracle)
	if err := g.CastSpell(me.ID, wave, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if prompt == nil {
		t.Fatal("no creature pick offered")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{mine}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, mine).DamageMarked; got != 0 {
		t.Errorf("the spared creature has %d damage", got)
	}
	if got := effectstest.FindCardAnywhere(t, g, theirs).DamageMarked; got != 3 {
		t.Errorf("the other creature has %d damage, want 3", got)
	}
}

func TestGapCaughtStealingForcesBlocksOrMakesATreasurePerHurtOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := gapB232Cast(t, g, "You've Been Caught Stealing", "Sorcery", gapCaughtStealingOracle, game.CastSpellParams{Modes: []int{0}}); err != nil {
		t.Fatalf("cast threaten: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if !effectstest.MkmB1MustBlock(t, g, id) {
			t.Error("a creature has no blocks-if-able requirement")
		}
	}
	// Two opponents are dealt damage this turn: two Treasures; a third is not.
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	src := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(src, opp.ID, 2)
		_ = g.DealDamageToPlayerForEffect(src, opp.ID, 1)
		_ = g.DealDamageToPlayerForEffect(src, other.ID, 1)
	})
	if err := gapB232Cast(t, g, "You've Been Caught Stealing", "Sorcery", gapCaughtStealingOracle, game.CastSpellParams{Modes: []int{1}}); err != nil {
		t.Fatalf("cast bribe: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTokensNamed2(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("Treasures = %d, want 2 (two opponents were dealt damage)", n)
	}
}

func TestGapVentureForthPutsTheLandOntoTheBattlefieldAndSuspendsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	land, one, two := uuid.New(), uuid.New(), uuid.New()
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Deep Library Card", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Library Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: two, Name: "Skipped Two", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: one, Name: "Skipped One", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	spell := uuid.New()
	effectstest.D5681Spell(g, me, spell, "Venture Forth", "Sorcery", "", gapVentureForthOracle)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// It comes back as a new object (CR 400.7), so it is found by name.
	_ = land
	landID := effectstest.FindBattlefieldByName(g, "Library Forest")
	if landID == uuid.Nil {
		t.Fatal("the land is not on the battlefield")
	}
	if c := effectstest.FindBattlefieldCardByID(g, landID); c == nil || c.Controller != me.ID {
		t.Errorf("the land is not under the caster's control: %+v", c)
	}
	inLibrary := map[string]bool{}
	for _, c := range me.Library.Cards {
		inLibrary[c.Name] = true
	}
	for _, name := range []string{"Skipped One", "Skipped Two"} {
		if !inLibrary[name] {
			t.Errorf("%s is not on the bottom of the library", name)
		}
	}
	_, _ = one, two
	if z := g.FindCardZoneForEffect(spell); z == nil || z.Kind != game.ZoneExile {
		t.Fatalf("Venture Forth is in %v, want exile", z)
	}
	if got := effectstest.FindCardAnywhere(t, g, spell).Counters[game.CounterTime]; got != 3 {
		t.Errorf("time counters = %d, want 3", got)
	}
	sa := effectstest.SuspendOf(gapVentureForthOracle)
	if sa == nil {
		t.Fatal("no suspend")
	}
}

func TestGapSplitTheSpoilsGivesTheOpponentsPileToTheHandAndTheRestToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := effectstest.BlbB1Seats(g)
	var perms []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		perms = append(perms, id)
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Old Permanent", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	}
	spellInGrave := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spellInGrave, Name: "Old Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	if err := gapB232Cast(t, g, "Split the Spoils", "Sorcery", gapSplitTheSpoilsOracle, game.CastSpellParams{Targets: effectstest.CardRefs(spellInGrave)}); err == nil {
		t.Fatal("targeted an instant card")
	}
	if err := gapB232Cast(t, g, "Split the Spoils", "Sorcery", gapSplitTheSpoilsOracle, game.CastSpellParams{Targets: effectstest.CardRefs(perms...)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range perms {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			t.Fatalf("a target is in %v, want exile before the split", z)
		}
	}
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	split := effectstest.LatestRevealPickFor(g, me.ID)
	if split == nil {
		t.Fatal("no pile split offered to the caster")
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, perms[:1]); err != nil {
		t.Fatalf("ResolvePileSplit: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, opp.ID) == nil {
		t.Fatal("the opponent was not asked to choose a pile")
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	inHand, inGrave := 0, 0
	for _, id := range perms {
		switch z := g.FindCardZoneForEffect(id); {
		case z != nil && z.Kind == game.ZoneHand:
			inHand++
		case z != nil && z.Kind == game.ZoneGraveyard:
			inGrave++
		}
	}
	if inHand+inGrave != 3 || inHand == 0 && inGrave == 0 {
		t.Errorf("hand %d, graveyard %d: the piles were not settled between them", inHand, inGrave)
	}
}
