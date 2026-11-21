package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s40_cards_test.go — batch 40 of the wave-2 Commander-deck cards
// (#5368): suspend, foretell, plot, miracle and casting from outside the
// hand. Each test plays the real catalog entry through the engine.

const (
	s40CharnelOracle   = "54c3d8ba-ee7c-4f15-8ee3-e5daf860b428"
	s40SentenceOracle  = "98641070-93bd-4088-90cb-956c2b21979e"
	s40WatcherOracle   = "b2354092-7a4c-4e69-9c0a-0a2e146dc00c"
	s40SageOracle      = "fc74f6a6-9064-4e1b-8dc1-446192bbc448"
	s40RavenformOracle = "d3d61a6a-7870-44df-92e5-513dac2c2c3e"
	s40LockOracle      = "9355db6e-d8eb-43be-9c1d-04f1806dc366"
	s40PyreticOracle   = "01766d9b-b2af-4a27-9e65-08c82e7d7637"
	s40FanaticOracle   = "017aa9b3-a8ea-4588-9c50-e914a7d8e4ee"
	s40RedressOracle   = "e8ff6780-5f22-42fa-a7c1-7406e8203063"
	s40AminatouOracle  = "bf45f5ee-4d5f-4920-b612-b5492c2f035d"
	s40TraderOracle    = "c1aba8c7-d940-4d5d-a176-68cc226906de"
	s40PilfererOracle  = "fe168bf1-f626-460c-890b-1ce758af01fc"
)

func TestS40EveryCardIsRegisteredFull(t *testing.T) {
	for _, id := range []string{s40CharnelOracle, s40SentenceOracle, s40WatcherOracle, s40SageOracle,
		s40RavenformOracle, s40LockOracle, s40PyreticOracle, s40FanaticOracle, s40RedressOracle,
		s40AminatouOracle, s40TraderOracle, s40PilfererOracle} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s is not registered Full", id)
		}
	}
}

// --- Suspended Sentence -------------------------------------------------

func TestS40SuspendedSentenceDestroysDrainsThreeAndSuspendsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.S40SuspendAction(t, s40SentenceOracle, 3, "{1}{B}")
	wolf := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.CastSpell(me.ID, s40Hand(g, me, "Suspended Sentence", "Instant", "{3}{B}", s40SentenceOracle), game.CastSpellParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}}}); err == nil {
		t.Fatal("targeted my own creature")
	}
	id := effectstest.CastFromHandForTest(t, g, me, "Suspended Sentence", "Instant", "{3}{B}", s40SentenceOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wolf}}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wolf) {
		t.Error("the Wolf survived")
	}
	if opp.Life != 37 {
		t.Errorf("opponent life %d, want 37", opp.Life)
	}
	if !effectstest.InExile(g, id) || effectstest.CounterCountInExile(g, id, game.CounterTime) != 3 {
		t.Error("Suspended Sentence is not exiled with three time counters")
	}
}

// s40Hand puts a card in a hand and returns its id.
func s40Hand(g *game.Game, p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: p.ID, Controller: p.ID})
	})
	return id
}

// --- Charnel Serenade ------------------------------------------------------

func TestS40CharnelSerenadeSurveilsReturnsACreatureWithFinalityAndSuspends(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.S40SuspendAction(t, s40CharnelOracle, 3, "{2}{B}")
	a := uuid.New()
	b := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: a, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: b, Name: "Elk", TypeLine: "Creature — Elk", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	})
	id := effectstest.CastFromHandForTest(t, g, me, "Charnel Serenade", "Sorcery", "{4}{B}{B}", s40CharnelOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RfmaAnswerSurveil(t, g, me)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(b) {
		t.Fatal("the chosen creature did not return")
	}
	if n := effectstest.BattlefieldCardPtr(t, g, b).Counters[game.CounterFinality]; n != 1 {
		t.Errorf("%d finality counters, want 1", n)
	}
	if !me.Graveyard.Contains(a) {
		t.Error("the other creature left the graveyard")
	}
	if !effectstest.InExile(g, id) || effectstest.CounterCountInExile(g, id, game.CounterTime) != 3 {
		t.Error("Charnel Serenade is not exiled with three time counters")
	}
}

func TestS40CharnelSerenadeWithNoCreatureStillSuspendsItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.CastFromHandForTest(t, g, me, "Charnel Serenade", "Sorcery", "{4}{B}{B}", s40CharnelOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RfmaAnswerSurveil(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, id) || effectstest.CounterCountInExile(g, id, game.CounterTime) != 3 {
		t.Error("Charnel Serenade is not exiled with three time counters")
	}
}

// --- Watcher of Hours ------------------------------------------------------

func TestS40WatcherOfHoursHasWardSuspendAndSurveilsPerTimeCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.S40SuspendAction(t, s40WatcherOracle, 6, "{1}{U}")
	id := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: id, Name: "Watcher of Hours", TypeLine: "Creature — Sphinx", ManaCost: "{5}{U}",
		OracleID: s40WatcherOracle, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
	effectstest.PutCounters(t, g, id, game.CounterTime, 6)
	effectstest.CrRemove(t, g, id, game.CounterTime, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Fatal("removing a time counter from the exiled card did not surveil")
	}
	effectstest.RfmaAnswerSurveil(t, g, me)
	// Two counters at once are two surveils.
	effectstest.CrRemove(t, g, id, game.CounterTime, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Fatal("no surveil for the second removal")
	}
	spec, _ := effects.Lookup(s40WatcherOracle)
	if len(spec.Triggered) != 2 {
		t.Errorf("%d triggers, want ward and the surveil", len(spec.Triggered))
	}
}

// --- Sage of the Beyond and Savvy Trader --------------------------------

// s40ExileCastPaysWith reports whether a three-mana colourless spell cast
// from exile (as a permission grants) is paid by `pool` mana of {C}, with
// `source` oracle on the battlefield for me.
func s40ExileCastPaysWith(t *testing.T, sourceName, sourceOracle string, pool int) bool {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	if sourceOracle != "" {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: sourceName, TypeLine: "Creature — Test",
			OracleID: sourceOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	}
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.Cards = append(g.Exile.Cards, game.Card{InstanceID: id, Name: "Test Golem", TypeLine: "Artifact Creature — Golem",
			ManaCost: "{3}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		_ = g.ExileCardWithPermissionForEffect(id, game.CastPermission{Player: me.ID, Duration: g.UntilEndOfTurnDuration()})
	})
	effectstest.B09AddColorless(me, pool)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "exile"}) == nil
}

func TestS40SageOfTheBeyondTakesTwoOffSpellsCastFromExileOnly(t *testing.T) {
	if s40ExileCastPaysWith(t, "", "", 1) {
		t.Fatal("control: a {3} spell from exile was paid with 1 mana")
	}
	if !s40ExileCastPaysWith(t, "Sage of the Beyond", s40SageOracle, 1) {
		t.Error("with Sage, a {3} spell from exile was not paid with 1 mana")
	}
	if s40ExileCastPaysWith(t, "Sage of the Beyond", s40SageOracle, 0) {
		t.Error("with Sage, the {3} spell from exile was free")
	}
	// A hand cast is not reduced.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sage of the Beyond", TypeLine: "Creature — Spirit Giant",
		OracleID: s40SageOracle, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	hand := s40Hand(g, me, "Test Golem", "Artifact Creature — Golem", "{3}", "")
	effectstest.B09AddColorless(me, 1)
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a spell cast from the hand was reduced")
	}
	spec, _ := effects.Lookup(s40SageOracle)
	if len(spec.SpecialActions) != 1 || spec.SpecialActions[0].Kind != game.SpecialActionForetell || spec.SpecialActions[0].CastCost != "{4}{U}" {
		t.Errorf("special actions %+v, want foretell {4}{U}", spec.SpecialActions)
	}
}

func TestS40SavvyTraderTakesOneOffAndExilesAPermanentCardToPlay(t *testing.T) {
	if !s40ExileCastPaysWith(t, "Savvy Trader", s40TraderOracle, 2) {
		t.Error("with Savvy Trader, a {3} spell from exile was not paid with 2 mana")
	}
	if s40ExileCastPaysWith(t, "Savvy Trader", s40TraderOracle, 1) {
		t.Error("with Savvy Trader, the reduction was more than {1}")
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	land := uuid.New()
	spell := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Shock", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	})
	effectstest.CastFromHandForTest(t, g, me, "Savvy Trader", "Creature — Human Citizen", "{3}{G}", s40TraderOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, land)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, land) || !me.Graveyard.Contains(spell) {
		t.Fatalf("the land is in exile = %v, the instant stayed in the graveyard = %v", effectstest.InExile(g, land), me.Graveyard.Contains(spell))
	}
	if effectstest.GrantedPermissionOn(g, me.ID, land, game.ZoneExile) == nil {
		t.Error("no permission to play the exiled land")
	}
}

// --- Ravenform -------------------------------------------------------------

func TestS40RavenformExilesAnArtifactOrCreatureAndItsControllerGetsABird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	spec, _ := effects.Lookup(s40RavenformOracle)
	if len(spec.SpecialActions) != 1 || spec.SpecialActions[0].Kind != game.SpecialActionForetell || spec.SpecialActions[0].CastCost != "{U}" {
		t.Errorf("special actions %+v, want foretell {U}", spec.SpecialActions)
	}
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	if err := g.CastSpell(me.ID, s40Hand(g, me, "Ravenform", "Sorcery", "{2}{U}", s40RavenformOracle), game.CastSpellParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err == nil {
		t.Fatal("targeted a land")
	}
	wolf := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	effectstest.CastFromHandForTest(t, g, me, "Ravenform", "Sorcery", "{2}{U}", s40RavenformOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wolf}}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wolf) || !effectstest.InExile(g, wolf) {
		t.Error("the Wolf was not exiled")
	}
	birds := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bird" && c.Controller == opp.ID && c.IsToken() {
			birds++
		}
	}
	if birds != 1 {
		t.Errorf("%d Birds for its controller, want 1", birds)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bird" && c.Controller == me.ID {
			t.Error("the caster got the Bird")
		}
	}
}

// --- Lock and Load ---------------------------------------------------------

func TestS40LockAndLoadDrawsOnePlusOnePerOtherInstantOrSorceryCast(t *testing.T) {
	for _, tc := range []struct {
		others, want int
	}{{0, 1}, {2, 3}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		spec, _ := effects.Lookup(s40LockOracle)
		if len(spec.SpecialActions) != 1 || spec.SpecialActions[0].Kind != game.SpecialActionPlot || spec.SpecialActions[0].Cost != "{3}{U}" {
			t.Errorf("special actions %+v, want plot {3}{U}", spec.SpecialActions)
		}
		for i := 0; i < tc.others; i++ {
			effectstest.CastFromHandForTest(t, g, me, "Opt Test", "Instant", "{U}", "", game.CastSpellParams{})
			effectstest.PassPriorityAroundTable(t, g)
		}
		before := me.Hand.Size()
		effectstest.CastFromHandForTest(t, g, me, "Lock and Load", "Sorcery", "{2}{U}", s40LockOracle, game.CastSpellParams{})
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Size() - before; got != tc.want-0 && got != tc.want {
			t.Errorf("%d other spells: hand grew by %d, want %d", tc.others, got, tc.want)
		}
	}
}

// --- Pyretic Charge --------------------------------------------------------

func TestS40PyreticChargeDiscardsHandDrawsFourAndPumpsPerCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	spec, _ := effects.Lookup(s40PyreticOracle)
	if len(spec.SpecialActions) != 1 || spec.SpecialActions[0].Kind != game.SpecialActionPlot || spec.SpecialActions[0].Cost != "{3}{R}" {
		t.Errorf("special actions %+v, want plot {3}{R}", spec.SpecialActions)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, n := range []string{"A", "B", "C"} {
		s40Hand(g, me, n, "Instant", "{R}", "")
	}
	effectstest.CastFromHandForTest(t, g, me, "Pyretic Charge", "Sorcery", "{4}{R}", s40PyreticOracle, game.CastSpellParams{})
	discarded := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 4 {
		t.Errorf("hand %d, want 4 after discarding and drawing four", got)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2+discarded {
		t.Errorf("Bear power %d, want %d (+1/+0 per card discarded)", p, 2+discarded)
	}
}

// --- Metamorphosis Fanatic -------------------------------------------------

func TestS40MetamorphosisFanaticReturnsACreatureWithALifelinkCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	elk := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: elk, Name: "Elk", TypeLine: "Creature — Elk", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	})
	effectstest.CastFromHandForTest(t, g, me, "Metamorphosis Fanatic", "Creature — Human Cleric", "{4}{B}{B}", s40FanaticOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, elk)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(elk) {
		t.Fatal("the Elk did not return")
	}
	if n := effectstest.BattlefieldCardPtr(t, g, elk).Counters[game.CounterLifelink]; n != 1 {
		t.Errorf("%d lifelink counters, want 1", n)
	}
	effectstest.AssertKeywords(t, g, elk, "lifelink")
}

func TestS40MetamorphosisFanaticMiracleCostsOneBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	fan := effectstest.DrawMiracleInOwnDrawStep(t, g, 1, "Metamorphosis Fanatic", "Creature — Human Cleric", "{4}{B}{B}", s40FanaticOracle)
	effectstest.RevealAndResolve(t, g, me)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, fan, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyMiracle}); err == nil {
		t.Fatal("miracle {1}{B} was paid with {B}")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, fan, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyMiracle}); err != nil {
		t.Fatalf("the miracle cast: %v", err)
	}
}

// --- Redress Fate ----------------------------------------------------------

func TestS40RedressFateReturnsArtifactsAndEnchantmentsOnlyAndHasMiracle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rock, ench, bear, land := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		for id, c := range map[uuid.UUID][2]string{rock: {"Sol Ring", "Artifact"}, ench: {"Glorious Anthem", "Enchantment"},
			bear: {"Bear", "Creature — Bear"}, land: {"Plains", "Basic Land — Plains"}} {
			me.Graveyard.PushTop(game.Card{InstanceID: id, Name: c[0], TypeLine: c[1], Owner: me.ID, Controller: me.ID})
		}
	})
	effectstest.CastFromHandForTest(t, g, me, "Redress Fate", "Sorcery", "{6}{W}{W}", s40RedressOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) || !g.Battlefield.Contains(ench) {
		t.Error("the artifact and enchantment did not return")
	}
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(land) || !me.Graveyard.Contains(bear) || !me.Graveyard.Contains(land) {
		t.Error("a creature or land card moved")
	}
	spec, _ := effects.Lookup(s40RedressOracle)
	if len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].ManaCost != "{3}{W}" {
		t.Errorf("alternative costs %+v, want miracle {3}{W}", spec.AlternativeCosts)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[1]
	fate := effectstest.DrawMiracleInOwnDrawStep(t, g2, 1, "Redress Fate", "Sorcery", "{6}{W}{W}", s40RedressOracle)
	effectstest.RevealAndResolve(t, g2, me2)
	if err := g2.AddManaForEffect(me2.ID, uuid.Nil, "{C}{C}{C}{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g2.CastSpell(me2.ID, fate, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyMiracle}); err != nil {
		t.Fatalf("the miracle cast for {3}{W}: %v", err)
	}
}

// --- Aminatou, Veil Piercer ------------------------------------------------

func TestS40AminatouGivesEnchantmentsMiracleAtTheirCostLessFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Aminatou, Veil Piercer", "Legendary Creature — Human Wizard", s40AminatouOracle, false)
	ench := s40DrawMiracle(t, g, 1, "Test Pillar", "Enchantment", "{4}{W}{W}")
	effectstest.RevealAndResolve(t, g, me)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, ench, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err == nil {
		t.Fatal("{W} paid the miracle cost {W}{W}")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, ench, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedMiracle}); err != nil {
		t.Fatalf("the miracle cast for {W}{W}: %v", err)
	}
}

func TestS40AminatouOffersNothingForACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Aminatou, Veil Piercer", "Legendary Creature — Human Wizard", s40AminatouOracle, false)
	s40DrawMiracle(t, g, 1, "Bear", "Creature — Bear", "{4}{G}")
	if effectstest.MiraclePrompt(g, me.ID) != nil {
		t.Error("a creature card was offered miracle")
	}
}

// --- Ghostly Pilferer ------------------------------------------------------

func TestS40GhostlyPilfererDrawsWhenAnOpponentCastsFromOutsideTheirHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Ghostly Pilferer", "Creature — Spirit Rogue", s40PilfererOracle, false)
	// A cast from the opponent's hand draws nothing.
	before := me.Hand.Size()
	effectstest.B13OpponentCasts(t, g, opp, "Their Instant", "Instant", "", "{U}", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Fatalf("a cast from the hand drew: hand %d → %d", before, me.Hand.Size())
	}
	// A cast from exile draws one.
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.Cards = append(g.Exile.Cards, game.Card{InstanceID: id, Name: "Their Exiled Instant", TypeLine: "Instant",
			ManaCost: "{U}", Owner: opp.ID, Controller: opp.ID})
		_ = g.ExileCardWithPermissionForEffect(id, game.CastPermission{Player: opp.ID, Duration: g.UntilEndOfTurnDuration()})
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("a cast from exile: hand %d → %d, want one drawn", before, me.Hand.Size())
	}
	// My own cast from exile draws nothing ("an opponent").
	mine := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.Cards = append(g.Exile.Cards, game.Card{InstanceID: mine, Name: "My Exiled Instant", TypeLine: "Instant",
			ManaCost: "{U}", Owner: me.ID, Controller: me.ID})
		_ = g.ExileCardWithPermissionForEffect(mine, game.CastPermission{Player: me.ID, Duration: g.UntilEndOfTurnDuration()})
	})
	before = me.Hand.Size()
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("my cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Errorf("my own cast drew: hand %d → %d", before, me.Hand.Size())
	}
}

func TestS40GhostlyPilfererDiscardMakesItUnblockableThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pil := effectstest.PushCatalogPermanent(g, me.ID, "Ghostly Pilferer", "Creature — Spirit Rogue", s40PilfererOracle, false)
	effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	if err := g.ActivateCatalogAbility(me.ID, pil, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without naming a card to discard")
	}
	fodder := s40Hand(g, me, "Fodder", "Instant", "{R}", "")
	if err := g.ActivateCatalogAbility(me.ID, pil, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(fodder) {
		t.Error("the card was not discarded")
	}
	effectstest.BrAttack(t, g, pil)
	if n := len(effectstest.BrOffered(t, g, opp.ID)); n != 0 {
		t.Errorf("the defender is offered %d blocks of an unblockable creature", n)
	}
}

func TestS40GhostlyPilfererUntapMayPayTwoToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	pil := effectstest.PushCatalogPermanent(g, me.ID, "Ghostly Pilferer", "Creature — Spirit Rogue", s40PilfererOracle, false)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, pil).Tapped = true })
	if err := g.UntapAll(me.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B09AddColorless(me, 2)
	before := me.Hand.Size()
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay prompt for the untap trigger")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d → %d, want one drawn after paying {2}", before, me.Hand.Size())
	}
}

// s40DrawMiracle is drawMiracleInOwnDrawStep for a table where Aminatou's
// upkeep surveil asks first: it answers that prompt, then makes the draw
// step's turn-based draw the first card seat draws.
func s40DrawMiracle(t *testing.T, g *game.Game, seat int, name, typeLine, manaCost string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	p := g.Seats[seat]
	effectstest.RfmaAnswerSurveil(t, g, p)
	effectstest.PassPriorityAroundTable(t, g)
	id := effectstest.PushMiracleOnTop(g, p, name, typeLine, manaCost, "")
	for g.Turn.Step != game.StepDraw {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to the draw step: %v", err)
		}
	}
	return id
}
