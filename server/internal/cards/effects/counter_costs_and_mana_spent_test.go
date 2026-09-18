package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_costs_and_mana_spent_test.go — the catalog half of #789 and
// #761. One assertion per converted card, against the real registered
// spec: the engine-side contract is pinned in game/counter_cost_rest_test.go
// and game/mana_spent_test.go.

const (
	vividCreekOracle     = "2da7c49f-cc1e-45d9-9cbf-067e92b0daef"
	vividGroveOracle     = "b7a68899-c0d3-49e0-854b-19268ae9b89d"
	ramosOracle          = "3ed41d2d-211b-4013-8562-8c64d54cc43a"
	mageRingOracle       = "136596a0-b179-40be-b42d-c0b992621c95"
	devotedDruidOracle   = "cb814e16-acf7-41d5-a357-1323dcc369f3"
	hopefulInitiateOrcle = "317169b8-9014-48e6-862b-ca21b706846e"
	painfulTruthsOracle  = "58a2d15f-1b4b-4303-b524-8cc0d1b3e7d2"
	etchedOracleOracle   = "7ecfa47e-1165-46a6-884c-290b1c14d020"
	vexingBaubleOracle   = "4514777d-0589-4631-978b-ff244167c176"
	slayingFireOracle    = "1e94e647-9150-4b22-aa9f-d195f64fb20a"
)

// floatForTest drops one token per rune into a seat's pool.
func floatForTest(g *game.Game, p *game.Player, spec string) {
	g.WithWriteLock(func() {
		for _, r := range spec {
			p.ManaPool.AddMana(game.ManaToken{Color: string(r), Source: uuid.New()})
		}
	})
}

// untapForTest untaps a battlefield permanent directly — b16Tap's
// inverse, for a fixture that needs an untapped land after an
// enters-tapped replacement.
func untapForTest(g *game.Game, id uuid.UUID) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = false
		}
	}
}

// activateManaFor fires a mana ability and fails the test on refusal.
func activateManaFor(t *testing.T, g *game.Game, controller, card uuid.UUID, idx int, params game.ManaAbilityParams) {
	t.Helper()
	if err := g.ActivateManaAbility(controller, card, idx, params); err != nil {
		t.Fatalf("activate mana ability %d: %v", idx, err)
	}
}

// --- #789: counter costs on mana abilities --------------------------

// Vivid Creek enters tapped with two charge counters, taps for its
// printed {U}, and spends a charge counter for any colour.
func TestVividCreekEntersWithChargeCountersAndFixesColours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := b12PlayFromHand(t, g, "Vivid Creek", "Land", vividCreekOracle, game.CastSpellParams{})

	card := b12Card(t, g, land)
	if !card.Tapped {
		t.Error("Vivid Creek did not enter tapped")
	}
	if got := counterCount(g, land, game.CounterCharge); got != 2 {
		t.Fatalf("charge counters %d, want 2", got)
	}

	// Untap it by hand and spend a charge counter on the fixer.
	untapForTest(g, land)
	activateManaFor(t, g, me.ID, land, 1, game.ManaAbilityParams{})
	if got := counterCount(g, land, game.CounterCharge); got != 1 {
		t.Errorf("charge counters %d after the fixer, want 1", got)
	}
}

// Ramos banks a counter per colour of each spell cast, and cashes
// five of them for ten mana — once a turn.
func TestRamosBanksSpellColoursAndCashesFiveCountersForTenMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ramos := b12Push(g, me.ID, "Ramos, Dragon Engine", "Legendary Artifact Creature — Dragon", ramosOracle, 4, 4)
	advanceToMain(t, g)

	// A two-colour spell banks two counters.
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Gold Spell", TypeLine: "Instant", ManaCost: "{W}{U}",
			Colors: []string{"W", "U"}, Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, ramos, game.CounterPlusOne); got != 2 {
		t.Fatalf("+1/+1 counters %d after a two-colour spell, want 2", got)
	}

	// Stack it to five by hand and cash them.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ramos, game.CounterPlusOne, 3) })
	activateManaFor(t, g, me.ID, ramos, 0, game.ManaAbilityParams{})
	if got := counterCount(g, ramos, game.CounterPlusOne); got != 0 {
		t.Errorf("+1/+1 counters %d after the payout, want 0", got)
	}
	if len(me.ManaPool) != 10 {
		t.Errorf("pool has %d tokens, want 10", len(me.ManaPool))
	}

	// "Activate only once each turn."
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ramos, game.CounterPlusOne, 5) })
	if err := g.ActivateManaAbility(me.ID, ramos, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the second activation this turn was allowed")
	}
}

// Mage-Ring Network banks storage counters and cashes any number of
// them for one {C} each — the variable count reaching the produced
// mana through the paid-cost record.
func TestMageRingNetworkCashesStorageCountersForColourless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := b12Push(g, me.ID, "Mage-Ring Network", "Land", mageRingOracle, 0, 0)
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(land, game.CounterStorage, 4) })

	activateManaFor(t, g, me.ID, land, 1, game.ManaAbilityParams{CounterCounts: []int{3}})
	if got := counterCount(g, land, game.CounterStorage); got != 1 {
		t.Errorf("storage counters %d, want 1", got)
	}
	if len(me.ManaPool) != 3 {
		t.Fatalf("pool has %d tokens, want 3 — one per counter removed", len(me.ManaPool))
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "C" {
			t.Errorf("minted %q, want colourless", tok.Color)
		}
	}
}

// --- #789: split and added counters ---------------------------------

// Iron Spider draws a card for two +1/+1 counters taken from among
// the artifacts its own tap ability grew.
func TestIronSpiderDrawsForCountersFromAmongArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spider := b12Push(g, me.ID, "Iron Spider, Stark Upgrade", "Legendary Artifact Creature — Spider Hero", b30IronSpiderOracle, 2, 3)
	other := b12Creature(g, me.ID, "Other Construct", "Artifact Creature — Construct", 1, 1)
	advanceToMain(t, g)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(spider, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(other, game.CounterPlusOne, 1)
	})

	hand := me.Hand.Size()
	floatForTest(g, me, "CC")
	b16Activate(t, g, me.ID, spider, 1, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{spider, other},
		CounterCounts:    []int{1, 1},
		Strict:           true,
	})
	passPriorityAroundTable(t, g)
	if counterCount(g, spider, game.CounterPlusOne) != 0 || counterCount(g, other, game.CounterPlusOne) != 0 {
		t.Error("the split payment did not take one counter from each artifact")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

// Hopeful Initiate destroys an artifact for two counters taken from
// among your creatures — both from one creature here.
func TestHopefulInitiateDestroysForCountersFromAmongCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	initiate := b12Push(g, me.ID, "Hopeful Initiate", "Creature — Human Warlock", hopefulInitiateOrcle, 1, 1)
	victim := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	advanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(initiate, game.CounterPlusOne, 2) })

	floatForTest(g, me, "WCC")
	b16Activate(t, g, me.ID, initiate, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{initiate},
		CounterCounts:    []int{2},
		Targets:          []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		Strict:           true,
	})
	passPriorityAroundTable(t, g)
	if counterCount(g, initiate, game.CounterPlusOne) != 0 {
		t.Error("the counters were not spent")
	}
	if z := b12ZoneOf(g, victim); z != game.ZoneGraveyard {
		t.Errorf("the artifact is in %s, want the graveyard", z)
	}
}

// Devoted Druid untaps itself by putting a -1/-1 counter on, and the
// cost is not doubled by anything — it is a cost, not an effect.
func TestDevotedDruidUntapsItselfForAMinusOneCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	druid := b12Push(g, me.ID, "Devoted Druid", "Creature — Elf Druid", devotedDruidOracle, 0, 2)
	advanceToMain(t, g)
	b16Tap(g, druid)

	b16Activate(t, g, me.ID, druid, 0, game.ActivateAbilityParams{})
	if got := counterCount(g, druid, game.CounterMinusOne); got != 1 {
		t.Fatalf("-1/-1 counters %d at announce, want 1", got)
	}
	passPriorityAroundTable(t, g)
	if b12Card(t, g, druid).Tapped {
		t.Error("the Druid did not untap")
	}
}

// --- #761: the mana spent -------------------------------------------

// Painful Truths converges for the number of colours that paid.
func TestPainfulTruthsConvergesForTheColoursSpent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	floatForTest(g, me, "BUG")
	hand, life := me.Hand.Size(), me.Life

	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Painful Truths", TypeLine: "Sorcery", ManaCost: "{2}{B}",
			OracleID: painfulTruthsOracle, Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards, want 3 — one per colour spent", got)
	}
	if got := life - me.Life; got != 3 {
		t.Errorf("lost %d life, want 3", got)
	}
}

// And a permissive cast — the human default — converges for nothing,
// because the engine never saw the payment. Weaker than printed, and
// declared on the card.
func TestPainfulTruthsDrawsNothingWhenTheEngineDidNotCharge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	hand := me.Hand.Size()

	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Painful Truths", TypeLine: "Sorcery", ManaCost: "{2}{B}",
			OracleID: painfulTruthsOracle, Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d cards, want 0 on an unrecorded payment", got)
	}
}

// Etched Oracle enters with a +1/+1 counter per colour of mana spent,
// and cashes four of them for three cards.
func TestEtchedOracleEntersWithSunburstCountersAndDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	floatForTest(g, me, "WUBR")

	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Etched Oracle", TypeLine: "Artifact Creature — Wizard", ManaCost: "{4}",
			OracleID: etchedOracleOracle, Power: 0, Toughness: 0, Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, id, game.CounterPlusOne); got != 4 {
		t.Fatalf("+1/+1 counters %d, want 4 — one per colour spent", got)
	}

	hand := me.Hand.Size()
	floatForTest(g, me, "C")
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
		Strict:  true,
	})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards, want 3", got)
	}
}

// Vexing Bauble counters a spell nothing was spent on, and leaves a
// paid one alone.
func TestVexingBaubleCountersOnlyAKnownFreeCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Vexing Bauble", "Artifact", vexingBaubleOracle, 0, 0)
	advanceToMain(t, g)

	// A {0} spell, paid strictly: a KNOWN nothing, so it is countered.
	free := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: free, Name: "Free Spell", TypeLine: "Instant", ManaCost: "{0}",
			Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, free, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the free spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if z := b12ZoneOf(g, free); z != game.ZoneGraveyard {
		t.Errorf("the free spell is in %s, want the graveyard (countered)", z)
	}

	// A spell actually paid for is left alone.
	floatForTest(g, me, "R")
	paid := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: paid, Name: "Paid Spell", TypeLine: "Instant", ManaCost: "{R}",
			Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, paid, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast the paid spell: %v", err)
	}
	g.WithWriteLock(func() {
		if item := g.StackItemForEffect(paid); item == nil || item.Paid.NoManaSpent() {
			t.Error("the paid spell recorded no mana")
		}
	})
}

// Slaying Fire deals 4 when three red mana paid for it, 3 otherwise.
func TestSlayingFireIsAdamantOnThreeRedMana(t *testing.T) {
	for _, tc := range []struct {
		name   string
		pool   string
		damage int
	}{
		{"three red", "RRR", 4},
		{"two red and a colourless", "RRC", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			advanceToMain(t, g)
			floatForTest(g, me, tc.pool)
			before := opp.Life

			id := uuid.New()
			g.WithWriteLock(func() {
				me.Hand.PushTop(game.Card{
					InstanceID: id, Name: "Slaying Fire", TypeLine: "Instant", ManaCost: "{2}{R}",
					OracleID: slayingFireOracle, Owner: me.ID, Controller: me.ID,
				})
			})
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{
				Strict:  true,
				Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
			}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			passPriorityAroundTable(t, g)
			if got := before - opp.Life; got != tc.damage {
				t.Errorf("dealt %d damage, want %d", got, tc.damage)
			}
		})
	}
}

// --- the registry ---------------------------------------------------

// Every card this PR converted is registered under the oracle ID the
// Scryfall dump gives it.
func TestCounterCostAndManaSpentCardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		vividCreekOracle:     "Vivid Creek",
		vividGroveOracle:     "Vivid Grove",
		ramosOracle:          "Ramos, Dragon Engine",
		mageRingOracle:       "Mage-Ring Network",
		devotedDruidOracle:   "Devoted Druid",
		hopefulInitiateOrcle: "Hopeful Initiate",
		painfulTruthsOracle:  "Painful Truths",
		etchedOracleOracle:   "Etched Oracle",
		vexingBaubleOracle:   "Vexing Bauble",
		slayingFireOracle:    "Slaying Fire",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}
