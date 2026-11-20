package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b17_cards_test.go — The Brothers' War, batch 17 (S82, #5482).
// Each test plays the real catalog entry through the engine.

const (
	bro17SerpentOracle   = "e469bb22-57f1-4555-b3c0-9ea108e6cc12"
	bro17ClayOracle      = "80920ac0-2bbe-4439-a937-282f63e23141"
	bro17ExemplarOracle  = "ab3b72e3-9a49-4c95-8016-22fa02ad7663"
	bro17LiberatorOracle = "798a4a6f-7443-4e8e-afcb-0a694fdfcbb3"
	bro17SuChiOracle     = "28264920-9d14-4661-83b8-56f950fbff39"
	bro17CoffinOracle    = "a46bfaf5-e631-4603-850a-f0074395b2f5"
	bro17BrainOracle     = "b28dcb41-5e47-4719-892f-da009a5d50b4"
	bro17SylexOracle     = "fb4d99c2-daec-430a-9ca2-5a178a96618b"
	bro17PartitionOracle = "2cfd5c18-47d0-4549-aacc-2428db21df2f"
	bro17HandoffOracle   = "1ecde43f-e418-437b-8b8f-efe120f96cb3"
)

func TestBro17EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		bro17SerpentOracle:   "Bladecoil Serpent",
		bro17ClayOracle:      "Clay Champion",
		bro17ExemplarOracle:  "Steel Exemplar",
		bro17LiberatorOracle: "Liberator, Urza's Battlethopter",
		bro17SuChiOracle:     "Su-Chi Cave Guard",
		bro17CoffinOracle:    "The Stasis Coffin",
		bro17BrainOracle:     "The Stone Brain",
		bro17SylexOracle:     "Urza's Sylex",
		bro17PartitionOracle: "Soul Partition",
		bro17HandoffOracle:   "Fateful Handoff",
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

// bro17Cast casts a permanent spell out of a floated pool under the
// strict gate and resolves what it can.
func bro17Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle, pool string, x int) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	floatForTest(g, me, pool)
	id := castFromHandForTest(t, g, me, name, typeLine, cost, oracle, game.CastSpellParams{Strict: true, XValue: x})
	for i := 0; i < 12 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	return id
}

func bro17HandSize(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = p.Hand.Size() })
	return n
}

// --- Bladecoil Serpent ----------------------------------------------

func TestBro17BladecoilSerpentCountsPairsOfEachColour(t *testing.T) {
	for _, tc := range []struct {
		name, pool      string
		draws, discards int
		power           int
		trample, haste  bool
	}{
		{"one of each pair", "UUBBRR", 1, 1, 1, true, true},
		{"four blue is two cards", "UUUUCC", 2, 0, 0, false, false},
		{"split colours make no pair", "UBRCCC", 0, 0, 0, false, false},
		{"four red", "RRRRCC", 0, 0, 2, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			g.WithWriteLock(func() {
				for i := 0; i < 3; i++ {
					opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
					me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
				}
			})
			before := bro17HandSize(g, me)
			id := bro17Cast(t, g, me, "Bladecoil Serpent", "Artifact Creature — Serpent", "{X}{6}", bro17SerpentOracle, tc.pool, 0)
			answerTriggerOrderLastQueuedFirst(t, g)
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Fatal("the Serpent did not resolve")
			}
			sawPrompt := false
			for _, p := range g.Seats {
				if p.ID == me.ID {
					continue
				}
				if pr := latestChooseCardsFor(g, p.ID); pr != nil {
					if p.ID == opp.ID {
						sawPrompt = true
					}
					if err := g.ResolveChooseCards(pr.ID, p.ID, pr.ChooseCards[:tc.discards]); err != nil {
						t.Fatalf("ResolveChooseCards: %v", err)
					}
				}
			}
			if sawPrompt != (tc.discards > 0) {
				t.Errorf("discard prompt for the opponent = %v, want %v", sawPrompt, tc.discards > 0)
			}
			passPriorityAroundTable(t, g)
			if got := bro17HandSize(g, me); got != before+tc.draws {
				t.Errorf("hand %d -> %d, want %d cards drawn", before, got, tc.draws)
			}
			if p, _ := ptOf(t, g, id); p != tc.power {
				t.Errorf("power %d, want %d (a base of 0 plus one per {R}{R})", p, tc.power)
			}
			if got := hasEffectiveKeyword(t, g, id, "trample"); got != tc.trample {
				t.Errorf("trample = %v, want %v", got, tc.trample)
			}
			if got := hasEffectiveKeyword(t, g, id, "haste"); got != tc.haste {
				t.Errorf("haste = %v, want %v", got, tc.haste)
			}
		})
	}
}

// --- Clay Champion --------------------------------------------------

func TestBro17ClayChampionEntersWithThreeCountersPerGreenPair(t *testing.T) {
	for _, tc := range []struct {
		name, pool string
		x, power   int // power is the +1/+1 counters: the helper card has a base of 0
	}{
		{"one pair", "GGCC", 0, 3},
		{"two pairs with X", "GGGGCC", 2, 6},
		{"split colours", "GWCC", 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := bro17Cast(t, g, me, "Clay Champion", "Artifact Creature — Construct", "{X}{4}", bro17ClayOracle, tc.pool, tc.x)
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Fatal("the Champion did not resolve")
			}
			if p, _ := ptOf(t, g, id); p != tc.power {
				t.Errorf("power %d, want %d", p, tc.power)
			}
		})
	}
}

func TestBro17ClayChampionPutsCountersOnUpToTwoOthersPerWhitePair(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, me.ID, "Ally A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Ally B", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "Ally C", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	id := bro17Cast(t, g, me, "Clay Champion", "Artifact Creature — Construct", "{X}{4}", bro17ClayOracle, "WWWWCC", 2)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if hasID(pick.PickTargetCards, id) || hasID(pick.PickTargetCards, theirs) {
		t.Fatalf("offered %v: neither the Champion itself nor an opponent's creature", pick.PickTargetCards)
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, tc := range []struct {
		id   uuid.UUID
		name string
		want int
	}{{a, "A", 2}, {b, "B", 2}, {c, "C", 0}, {theirs, "theirs", 0}, {id, "the Champion", 0}} {
		if got := countersOn(g, tc.id, game.CounterPlusOne); got != tc.want {
			t.Errorf("%s has %d +1/+1 counters, want %d", tc.name, got, tc.want)
		}
	}
}

// --- Steel Exemplar -------------------------------------------------

func TestBro17SteelExemplarCountersUnlessTwoColorsSpent(t *testing.T) {
	for _, tc := range []struct {
		name, pool string
		power      int
	}{
		{"colourless only", "CCCCC", 2},
		{"one colour", "RRRRR", 2},
		{"two colours", "RRRRG", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := bro17Cast(t, g, me, "Steel Exemplar", "Artifact Creature — Wizard", "{5}", bro17ExemplarOracle, tc.pool, 0)
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Fatal("the Exemplar did not resolve")
			}
			if p, _ := ptOf(t, g, id); p != tc.power {
				t.Errorf("power %d, want %d", p, tc.power)
			}
		})
	}
}

// Not cast, no mana spent: it enters with the counters.
func TestBro17SteelExemplarPutOntoTheBattlefieldGetsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Steel Exemplar", OracleID: bro17ExemplarOracle,
			TypeLine: "Artifact Creature — Wizard", ManaCost: "{5}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	if p, _ := ptOf(t, g, id); p != 6 {
		t.Errorf("power %d, want 6 (a printed 4 and two counters)", p)
	}
}

// --- Su-Chi Cave Guard ----------------------------------------------

func TestBro17SuChiCaveGuardAddsEightColorlessWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := pushPermanentForTest(g, me.ID, "Su-Chi Cave Guard", bro17SuChiOracle, "Artifact Creature — Construct")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 8, 8
			}
		}
	})
	if !hasEffectiveKeyword(t, g, id, "vigilance") {
		t.Error("Su-Chi Cave Guard has vigilance")
	}
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{id}) })
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 8 {
		t.Errorf("pool holds %d mana, want 8", got)
	}
	// Moving to the next step keeps it (CR 500.4 is waived by the rider).
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if got := me.ManaPool.Total(); got != 8 {
		t.Errorf("after the step ends the pool holds %d mana, want 8", got)
	}
}

// --- The Stasis Coffin ----------------------------------------------

func TestBro17TheStasisCoffinGivesProtectionFromEverything(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	coffin := pushPermanentForTest(g, me.ID, "The Stasis Coffin", bro17CoffinOracle, "Legendary Artifact")

	if err := g.ActivateCatalogAbility(me.ID, coffin, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana to pay {2}")
	}
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, coffin, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(coffin) {
		t.Error("the Coffin is exiled as a cost")
	}
	passPriorityAroundTable(t, g)
	if !hasPlayerAbility(playerAbilities(g, me), ProtectionFromEverything) {
		t.Error("the activator did not gain protection from everything")
	}
	for _, p := range g.Seats {
		if p.ID != me.ID && hasPlayerAbility(playerAbilities(g, p), ProtectionFromEverything) {
			t.Errorf("%s gained protection too", p.Name)
		}
	}
}

// --- Liberator, Urza's Battlethopter --------------------------------

func TestBro17LiberatorGrowsOnlyWhenTheSpellCostsMoreThanItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	lib := pushPermanentForTest(g, me.ID, "Liberator, Urza's Battlethopter", bro17LiberatorOracle, "Legendary Artifact Creature — Thopter")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == lib {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 1, 2
			}
		}
	})
	cast := func(cost, pool string) {
		t.Helper()
		floatForTest(g, me, pool)
		castFromHandForTest(t, g, me, "Spell", "Sorcery", cost, "", game.CastSpellParams{Strict: true})
		passPriorityAroundTable(t, g)
	}
	cast("{1}", "C") // 1 is not greater than power 1
	if got := countersOn(g, lib, game.CounterPlusOne); got != 0 {
		t.Fatalf("after a one-mana spell: %d counters, want 0", got)
	}
	cast("{2}", "CC") // 2 > 1
	if got := countersOn(g, lib, game.CounterPlusOne); got != 1 {
		t.Fatalf("after a two-mana spell: %d counters, want 1", got)
	}
	cast("{2}", "CC") // power 2 now; 2 is not greater
	if got := countersOn(g, lib, game.CounterPlusOne); got != 1 {
		t.Fatalf("after a second two-mana spell: %d counters, want 1", got)
	}
	cast("{3}", "CCC")
	if got := countersOn(g, lib, game.CounterPlusOne); got != 2 {
		t.Fatalf("after a three-mana spell: %d counters, want 2", got)
	}
}

func TestBro17LiberatorLetsYouCastArtifactAndColorlessSpellsAtInstantSpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushPermanentForTest(g, me.ID, "Liberator, Urza's Battlethopter", bro17LiberatorOracle, "Legendary Artifact Creature — Thopter")
	advanceTo(t, g, game.StepEnd)
	hand := func(name, typeLine, cost string) uuid.UUID {
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
		})
		return id
	}
	if err := g.CastSpell(me.ID, hand("Green Bear", "Creature — Bear", "{1}{G}"), game.CastSpellParams{}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("a green creature spell in the end step = %v, want ErrSorcerySpeedRequired", err)
	}
	floatForTest(g, me, "CC")
	if err := g.CastSpell(me.ID, hand("Gizmo", "Artifact", "{2}"), game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("an artifact spell in the end step: %v", err)
	}
	floatForTest(g, me, "CC")
	if err := g.CastSpell(me.ID, hand("Drone", "Creature — Eldrazi Drone", "{2}"), game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a colorless creature spell in the end step: %v", err)
	}
}

// --- The Stone Brain ------------------------------------------------

func TestBro17TheStoneBrainExilesNamedCardsAndDrawsForTheHandOnes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	brain := pushPermanentForTest(g, me.ID, "The Stone Brain", bro17BrainOracle, "Legendary Artifact")
	var handA, handB, gy, lib, other uuid.UUID
	mk := func(name string) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID}
	}
	g.WithWriteLock(func() {
		a, b, c, d, e := mk("Sol Ring"), mk("Sol Ring"), mk("Sol Ring"), mk("Sol Ring"), mk("Mox Pearl")
		handA, handB, gy, lib, other = a.InstanceID, b.InstanceID, c.InstanceID, d.InstanceID, e.InstanceID
		opp.Hand.PushTop(a)
		opp.Hand.PushTop(b)
		opp.Hand.PushTop(e)
		opp.Graveyard.PushTop(c)
		opp.Library.PushTop(d)
		for i := 0; i < 4; i++ {
			opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
		}
	})
	handBefore, libBefore, myHand := bro17HandSize(g, opp), opp.Library.Size(), bro17HandSize(g, me)
	floatForTest(g, me, "CC")
	if err := g.ActivateCatalogAbility(me.ID, brain, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	var name *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceCardName {
			name = c
		}
	}
	if name == nil {
		t.Fatal("no card-name prompt")
	}
	if err := g.ResolveCardNameChoice(name.ID, me.ID, "sol ring"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	search := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if search == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{handA, handB, gy}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{handA, handB, gy} {
		if !inExile(g, id) {
			t.Errorf("card %v was not exiled", id)
		}
	}
	if inExile(g, lib) || inExile(g, other) {
		t.Error("a card that was not taken left its zone")
	}
	// Two hand cards were exiled, two were drawn: the hand is where it
	// was and the library is two cards lighter.
	if got := bro17HandSize(g, opp); got != handBefore {
		t.Errorf("the opponent's hand is %d, want %d (two exiled, two drawn)", got, handBefore)
	}
	if got := opp.Library.Size(); got != libBefore-2 {
		t.Errorf("the opponent's library is %d, want %d (two drawn; the library copy was not taken)", got, libBefore-2)
	}
	if got := bro17HandSize(g, me); got != myHand {
		t.Errorf("the activator's hand changed: %d -> %d", myHand, got)
	}
}

// --- Urza's Sylex ---------------------------------------------------

func TestBro17UrzasSylexDestroysAllButSixLandsEach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	sylex := pushPermanentForTest(g, me.ID, "Urza's Sylex", bro17SylexOracle, "Legendary Artifact")
	myLand := pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	myBear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirBear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	var lands []uuid.UUID
	for i := 0; i < 8; i++ {
		lands = append(lands, pushPermanentForTest(g, opp.ID, "Island", "", "Basic Land — Island"))
	}
	floatForTest(g, me, "WWCC")
	if err := g.ActivateCatalogAbility(me.ID, sylex, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	// #5793: the exile cost's trigger resolves first, and its "you may pay
	// {2}" holds the table until it is answered. Decline it.
	if hasPayUnlessFor(g, me.ID) {
		answerPayUnless(t, g, me.ID, false)
		passPriorityAroundTable(t, g)
	}
	pr := latestChooseCardsFor(g, opp.ID)
	if pr == nil {
		t.Fatal("the opponent with eight lands was not asked for six")
	}
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Error("the activator has one land and must not be asked")
	}
	if err := g.ResolveChooseCards(pr.ID, opp.ID, lands[:6]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	onBF := func(id uuid.UUID) bool { return g.Battlefield.Contains(id) }
	for i, id := range lands {
		if onBF(id) != (i < 6) {
			t.Errorf("opponent land %d on the battlefield = %v, want %v", i, onBF(id), i < 6)
		}
	}
	if !onBF(myLand) {
		t.Error("a player with six lands or fewer keeps them all")
	}
	if onBF(myBear) || onBF(theirBear) {
		t.Error("creatures survived the destroy-all")
	}
	if onBF(sylex) {
		t.Error("the Sylex is exiled as a cost")
	}
}

func TestBro17UrzasSylexIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	sylex := pushPermanentForTest(g, me.ID, "Urza's Sylex", bro17SylexOracle, "Legendary Artifact")
	advanceTo(t, g, game.StepEnd)
	floatForTest(g, me, "WWCC")
	if err := g.ActivateCatalogAbility(me.ID, sylex, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated in the end step")
	}
}

func TestBro17UrzasSylexExiledFromTheBattlefieldOffersAPlaneswalkerSearch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	sylex := pushPermanentForTest(g, me.ID, "Urza's Sylex", bro17SylexOracle, "Legendary Artifact")
	walker := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: walker, Name: "Jace", TypeLine: "Legendary Planeswalker — Jace", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	})
	floatForTest(g, me, "WWCCCC")
	if err := g.ActivateCatalogAbility(me.ID, sylex, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	pay := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("the exile trigger did not offer {2}")
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	search := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if search == nil {
		t.Fatal("no search prompt after paying")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{walker}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(walker) {
		t.Error("the planeswalker did not reach the hand")
	}
}

// --- Soul Partition -------------------------------------------------

func TestBro17SoulPartitionTaxesOnlyAnOpponentsCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	mine := uuid.New()
	theirs := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: mine, Name: "My Gizmo", TypeLine: "Artifact", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
		g.Battlefield.PushTop(game.Card{InstanceID: theirs, Name: "Their Gizmo", TypeLine: "Artifact", ManaCost: "{1}", Owner: opp.ID, Controller: opp.ID})
	})
	for _, target := range []uuid.UUID{mine, theirs} {
		floatForTest(g, me, "WC")
		castFromHandForTest(t, g, me, "Soul Partition", "Instant", "{1}{W}", bro17PartitionOracle,
			game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
		passPriorityAroundTable(t, g)
		if !inExile(g, target) {
			t.Fatalf("target %v was not exiled", target)
		}
	}
	// My own card: playable at its printed cost.
	advanceToMain(t, g)
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Errorf("casting my own exiled card for {1}: %v", err)
	}
	// Their card: {2} more, on their own turn.
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	floatForTest(g, opp, "C")
	if err := g.CastSpell(opp.ID, theirs, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("their card cast for {1}: the {2} tax is missing")
	}
	floatForTest(g, opp, "CC")
	if err := g.CastSpell(opp.ID, theirs, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Errorf("their card cast for {3}: %v", err)
	}
}

// --- Fateful Handoff ------------------------------------------------

func TestBro17FatefulHandoffDrawsForManaValueAndGivesItAway(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	gizmo := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: gizmo, Name: "Gizmo", TypeLine: "Artifact", ManaCost: "{3}", Owner: me.ID, Controller: me.ID})
		for i := 0; i < 5; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		}
	})
	before := bro17HandSize(g, me)
	floatForTest(g, me, "BCCC")
	castFromHandForTest(t, g, me, "Fateful Handoff", "Sorcery", "{3}{B}", bro17HandoffOracle,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: gizmo}}})
	passPriorityAroundTable(t, g)
	if got := bro17HandSize(g, me); got != before+3 {
		t.Errorf("hand %d -> %d, want 3 cards drawn", before, got)
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if pick == nil {
		t.Fatal("no opponent choice")
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if o.Player == opp.ID {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("the opponent is not among the options %+v", pick.PickOptions)
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	passPriorityAroundTable(t, g)
	var controller uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == gizmo {
				controller = c.Controller
			}
		}
	})
	if controller != opp.ID {
		t.Errorf("the Gizmo is controlled by %v, want the chosen opponent", controller)
	}
}

func TestBro17FatefulHandoffCannotTargetAnOpponentsPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	theirs := pushPermanentForTest(g, opp.ID, "Their Bear", "", "Creature — Bear")
	floatForTest(g, me, "BCCC")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fateful Handoff", TypeLine: "Sorcery", ManaCost: "{3}{B}", OracleID: bro17HandoffOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err == nil {
		t.Error("targeted a permanent the caster does not control")
	}
}

func TestBro17SoulPartitionExilesATokenForGood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	tok := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: tok, Name: "Soldier", TypeLine: "Token Creature — Soldier",
			Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	})
	floatForTest(g, me, "WC")
	castFromHandForTest(t, g, me, "Soul Partition", "Instant", "{1}{W}", bro17PartitionOracle,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tok}}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tok) {
		t.Error("the token is still on the battlefield")
	}
	if inExile(g, tok) {
		t.Error("a token cannot stay in exile to be played")
	}
}
