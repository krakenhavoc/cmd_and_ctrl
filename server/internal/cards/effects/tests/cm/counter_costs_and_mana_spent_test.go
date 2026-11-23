package testscm

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- #789: counter costs on mana abilities --------------------------

// Vivid Creek enters tapped with two charge counters, taps for its
// printed {U}, and spends a charge counter for any colour.
func TestVividCreekEntersWithChargeCountersAndFixesColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.B12PlayFromHand(t, g, "Vivid Creek", "Land", effectstest.VividCreekOracle, game.CastSpellParams{})

	card := effectstest.B12Card(t, g, land)
	if !card.Tapped {
		t.Error("Vivid Creek did not enter tapped")
	}
	if got := effectstest.CounterCount(g, land, game.CounterCharge); got != 2 {
		t.Fatalf("charge counters %d, want 2", got)
	}

	// Untap it by hand and spend a charge counter on the fixer.
	effectstest.UntapForTest(g, land)
	effectstest.ActivateManaFor(t, g, me.ID, land, 1, game.ManaAbilityParams{})
	if got := effectstest.CounterCount(g, land, game.CounterCharge); got != 1 {
		t.Errorf("charge counters %d after the fixer, want 1", got)
	}
}

// Ramos banks a counter per colour of each spell cast, and cashes
// five of them for ten mana — once a turn.
func TestRamosBanksSpellColoursAndCashesFiveCountersForTenMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ramos := effectstest.B12Push(g, me.ID, "Ramos, Dragon Engine", "Legendary Artifact Creature — Dragon", effectstest.RamosOracle, 4, 4)
	effectstest.AdvanceToMain(t, g)

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
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, ramos, game.CounterPlusOne); got != 2 {
		t.Fatalf("+1/+1 counters %d after a two-colour spell, want 2", got)
	}

	// Stack it to five by hand and cash them.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ramos, game.CounterPlusOne, 3) })
	effectstest.ActivateManaFor(t, g, me.ID, ramos, 0, game.ManaAbilityParams{})
	if got := effectstest.CounterCount(g, ramos, game.CounterPlusOne); got != 0 {
		t.Errorf("+1/+1 counters %d after the payout, want 0", got)
	}
	if me.ManaPool.Total() != 10 {
		t.Errorf("pool has %d tokens, want 10", me.ManaPool.Total())
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.B12Push(g, me.ID, "Mage-Ring Network", "Land", effectstest.MageRingOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(land, game.CounterStorage, 4) })

	effectstest.ActivateManaFor(t, g, me.ID, land, 1, game.ManaAbilityParams{CounterCounts: []int{3}})
	if got := effectstest.CounterCount(g, land, game.CounterStorage); got != 1 {
		t.Errorf("storage counters %d, want 1", got)
	}
	if me.ManaPool.Total() != 3 {
		t.Fatalf("pool has %d tokens, want 3 — one per counter removed", me.ManaPool.Total())
	}
	for _, tok := range me.ManaPool.Units() {
		if tok.Color != "C" {
			t.Errorf("minted %q, want colourless", tok.Color)
		}
	}
}

// --- #789: split and added counters ---------------------------------

// Iron Spider draws a card for two +1/+1 counters taken from among
// the artifacts its own tap ability grew.
func TestIronSpiderDrawsForCountersFromAmongArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spider := effectstest.B12Push(g, me.ID, "Iron Spider, Stark Upgrade", "Legendary Artifact Creature — Spider Hero", effectstest.B30IronSpiderOracle, 2, 3)
	other := effectstest.B12Creature(g, me.ID, "Other Construct", "Artifact Creature — Construct", 1, 1)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(spider, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(other, game.CounterPlusOne, 1)
	})

	hand := me.Hand.Size()
	effectstest.FloatForTest(g, me, "CC")
	effectstest.B16Activate(t, g, me.ID, spider, 1, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{spider, other},
		CounterCounts:    []int{1, 1},
		Strict:           true,
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, spider, game.CounterPlusOne) != 0 || effectstest.CounterCount(g, other, game.CounterPlusOne) != 0 {
		t.Error("the split payment did not take one counter from each artifact")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
}

// Hopeful Initiate destroys an artifact for two counters taken from
// among your creatures — both from one creature here.
func TestHopefulInitiateDestroysForCountersFromAmongCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	initiate := effectstest.B12Push(g, me.ID, "Hopeful Initiate", "Creature — Human Warlock", effectstest.HopefulInitiateOrcle, 1, 1)
	victim := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(initiate, game.CounterPlusOne, 2) })

	effectstest.FloatForTest(g, me, "WCC")
	effectstest.B16Activate(t, g, me.ID, initiate, 0, game.ActivateAbilityParams{
		CounterSourceIDs: []uuid.UUID{initiate},
		CounterCounts:    []int{2},
		Targets:          []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		Strict:           true,
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterCount(g, initiate, game.CounterPlusOne) != 0 {
		t.Error("the counters were not spent")
	}
	if z := effectstest.B12ZoneOf(g, victim); z != game.ZoneGraveyard {
		t.Errorf("the artifact is in %s, want the graveyard", z)
	}
}

// Devoted Druid untaps itself by putting a -1/-1 counter on, and the
// cost is not doubled by anything — it is a cost, not an effect.
func TestDevotedDruidUntapsItselfForAMinusOneCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	druid := effectstest.B12Push(g, me.ID, "Devoted Druid", "Creature — Elf Druid", effectstest.DevotedDruidOracle, 0, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Tap(g, druid)

	effectstest.B16Activate(t, g, me.ID, druid, 0, game.ActivateAbilityParams{})
	if got := effectstest.CounterCount(g, druid, game.CounterMinusOne); got != 1 {
		t.Fatalf("-1/-1 counters %d at announce, want 1", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12Card(t, g, druid).Tapped {
		t.Error("the Druid did not untap")
	}
}

// --- #761: the mana spent -------------------------------------------

// Painful Truths converges for the number of colours that paid.
func TestPainfulTruthsConvergesForTheColoursSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "BUG")
	hand, life := me.Hand.Size(), me.Life

	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Painful Truths", TypeLine: "Sorcery", ManaCost: "{2}{B}",
			OracleID: effectstest.PainfulTruthsOracle, Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards, want 3 — one per colour spent", got)
	}
	if got := life - me.Life; got != 3 {
		t.Errorf("lost %d life, want 3", got)
	}
}

// #3855: a permissive cast — the human default — with nothing in the
// pool asks what was paid outside it, because Painful Truths reads the
// colours. The prompt names the shortfall slot by slot, coloured first;
// a declaration that is not a legal payment of exactly that is refused
// with nothing paid; a legal one converges for every colour declared.
func TestPainfulTruthsAsksWhatAWaivedPaymentPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()

	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Painful Truths", TypeLine: "Sorcery", ManaCost: "{2}{B}",
			OracleID: effectstest.PainfulTruthsOracle, Owner: me.ID, Controller: me.ID,
		})
	})
	var ask *game.PaperPaymentRequiredError
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); !errors.As(err, &ask) {
		t.Fatalf("cast = %v, want a paper payment prompt", err)
	}
	var symbols []string
	for _, s := range ask.Slots {
		symbols = append(symbols, s.Symbol)
	}
	if !reflect.DeepEqual(symbols, []string{"{B}", "{1}", "{1}"}) || !reflect.DeepEqual(ask.Slots[0].Options, []string{"B"}) ||
		len(ask.Slots[1].Options) != 6 || ask.ArtifactSources {
		t.Fatalf("prompt = %+v, want {B} {1} {1}", ask)
	}
	if !me.Hand.Contains(id) {
		t.Fatal("the refused cast moved the card")
	}
	for name, bad := range map[string][]game.PaperMana{
		"no black":     {{Color: "R"}, {Color: "W"}, {Color: "U"}},
		"too few":      {{Color: "B"}, {Color: "W"}},
		"too many":     {{Color: "B"}, {Color: "W"}, {Color: "U"}, {Color: "G"}},
		"not a symbol": {{Color: "B"}, {Color: "W"}, {Color: "X"}},
	} {
		err := g.CastSpell(me.ID, id, game.CastSpellParams{PaperPayment: &game.PaperPayment{Mana: bad}})
		if !errors.Is(err, game.ErrInvalidPaperPayment) {
			t.Errorf("%s: cast = %v, want ErrInvalidPaperPayment", name, err)
		}
	}
	if !me.Hand.Contains(id) {
		t.Fatal("a refused declaration moved the card")
	}
	decl := &game.PaperPayment{Mana: []game.PaperMana{{Color: "W"}, {Color: "B"}, {Color: "U"}}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{PaperPayment: decl}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if paid := g.StackMeta[id].Paid; paid.OnPaper || !paid.Spent().Known() || paid.Spent().Total() != 3 {
		t.Fatalf("a declared payment is exact: %+v", paid)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards, want 3 — one per colour declared", got)
	}
}

// #3844 + #3855: a waived payment spends what the pool can pay, and the
// caster declares only the rest. Here {U}{G} floats for a {2}{B}
// Painful Truths: both pay the generic half, the {B} is the one slot
// asked about, and the spell converges for three.
func TestPainfulTruthsConvergesForThePoolShareAndTheDeclaredRest(t *testing.T) {
	for _, tc := range []struct {
		name   string
		params game.CastSpellParams
	}{
		{"insufficient permissive pool", game.CastSpellParams{}},
		{"cast anyway", game.CastSpellParams{Strict: true, ForceCast: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			effectstest.FloatForTest(g, me, "UG")
			hand := me.Hand.Size()

			id := uuid.New()
			g.WithWriteLock(func() {
				me.Hand.PushTop(game.Card{
					InstanceID: id, Name: "Painful Truths", TypeLine: "Sorcery", ManaCost: "{2}{B}",
					OracleID: effectstest.PainfulTruthsOracle, Owner: me.ID, Controller: me.ID,
				})
			})
			var ask *game.PaperPaymentRequiredError
			if err := g.CastSpell(me.ID, id, tc.params); !errors.As(err, &ask) || len(ask.Slots) != 1 || ask.Slots[0].Symbol != "{B}" {
				t.Fatalf("cast = %v (%+v), want a prompt for the {B} alone", err, ask)
			}
			if me.ManaPool.Total() != 2 {
				t.Fatal("the refused cast spent the pool")
			}
			params := tc.params
			params.PaperPayment = &game.PaperPayment{Mana: []game.PaperMana{{Color: "B"}}}
			if err := g.CastSpell(me.ID, id, params); err != nil {
				t.Fatalf("cast: %v", err)
			}
			if paid := g.StackMeta[id].Paid; paid.OnPaper || paid.Spent().Total() != 3 {
				t.Fatalf("want {U}{G} spent and {B} declared, got %+v", paid)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != 3 {
				t.Errorf("drew %d cards, want 3", got)
			}
		})
	}
}

// Etched Oracle enters with a +1/+1 counter per colour of mana spent,
// and cashes four of them for three cards.
func TestEtchedOracleEntersWithSunburstCountersAndDrawsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "WUBR")

	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Etched Oracle", TypeLine: "Artifact Creature — Wizard", ManaCost: "{4}",
			OracleID: effectstest.EtchedOracleOracle, Power: 0, Toughness: 0, Owner: me.ID, Controller: me.ID,
		})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 4 {
		t.Fatalf("+1/+1 counters %d, want 4 — one per colour spent", got)
	}

	hand := me.Hand.Size()
	effectstest.FloatForTest(g, me, "C")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
		Strict:  true,
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards, want 3", got)
	}
}

// Vexing Bauble counters a spell nothing was spent on, and leaves a
// paid one alone.
func TestVexingBaubleCountersOnlyAKnownFreeCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Vexing Bauble", "Artifact", effectstest.VexingBaubleOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)

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
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, free); z != game.ZoneGraveyard {
		t.Errorf("the free spell is in %s, want the graveyard (countered)", z)
	}

	// A spell actually paid for is left alone.
	effectstest.FloatForTest(g, me, "R")
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
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			effectstest.AdvanceToMain(t, g)
			effectstest.FloatForTest(g, me, tc.pool)
			before := opp.Life

			id := uuid.New()
			g.WithWriteLock(func() {
				me.Hand.PushTop(game.Card{
					InstanceID: id, Name: "Slaying Fire", TypeLine: "Instant", ManaCost: "{2}{R}",
					OracleID: effectstest.SlayingFireOracle, Owner: me.ID, Controller: me.ID,
				})
			})
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{
				Strict:  true,
				Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
			}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
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
		effectstest.VividCreekOracle:     "Vivid Creek",
		effectstest.VividGroveOracle:     "Vivid Grove",
		effectstest.RamosOracle:          "Ramos, Dragon Engine",
		effectstest.MageRingOracle:       "Mage-Ring Network",
		effectstest.DevotedDruidOracle:   "Devoted Druid",
		effectstest.HopefulInitiateOrcle: "Hopeful Initiate",
		effectstest.PainfulTruthsOracle:  "Painful Truths",
		effectstest.EtchedOracleOracle:   "Etched Oracle",
		effectstest.VexingBaubleOracle:   "Vexing Bauble",
		effectstest.SlayingFireOracle:    "Slaying Fire",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("oracle %s registered as %q, want %q", oracle, spec.Name, name)
		}
	}
}
