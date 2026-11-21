package testscm

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Merfolk batch (#5368, cmdr2 batch 5).

const (
	cmdr2B05SvyelunOracle     = "872bebff-720d-459b-8f86-fb9129a8d1fb"
	cmdr2B05KopalaOracle      = "0237cccc-3161-4d91-8e69-e3d315dcb246"
	cmdr2B05BanneretOracle    = "0d6202c7-b19c-429a-85c4-332024381dd7"
	cmdr2B05SurgespannerOracl = "3f15349a-6ac6-4091-95a3-ae9f593e4650"
	cmdr2B05SingerOracle      = "7df43670-c948-464d-9bb9-db672e755f0c"
	cmdr2B05MihailOracle      = "6c901ff4-15c3-468f-9a1b-4c69a7bbc31a"
	cmdr2B05DeeprootOracle    = "764e8e6e-0ace-4fc8-8ca5-468a10ae3b73"
	cmdr2B05KumenaOracle      = "e7125493-febc-4e8e-abdc-5fc4f94a281c"
	cmdr2B05MistbinderOracle  = "be56dfea-1ece-4f06-b1cc-d7425b5018b4"
	cmdr2B05ShorestalkerOracl = "71a60247-17e5-44bd-9fed-f7d550b09ef3"
)

func TestCmdr2B05Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdr2B05SvyelunOracle:     "Svyelun of Sea and Sky",
		cmdr2B05KopalaOracle:      "Kopala, Warden of Waves",
		cmdr2B05BanneretOracle:    "Stonybrook Banneret",
		cmdr2B05SurgespannerOracl: "Surgespanner",
		cmdr2B05SingerOracle:      "Singer of Swift Rivers",
		cmdr2B05MihailOracle:      "Emperor Mihail II",
		cmdr2B05DeeprootOracle:    "Deeproot Elite",
		cmdr2B05KumenaOracle:      "Kumena, Tyrant of Orazca",
		cmdr2B05MistbinderOracle:  "Merfolk Mistbinder",
		cmdr2B05ShorestalkerOracl: "Triton Shorestalker",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 || spec.Name != name {
			t.Errorf("%s: %+v %v", name, spec, ok)
		}
	}
}

func TestCmdr2B05ShorestalkerCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.Cmdr2B05Card(g, me.ID, "Triton Shorestalker", "Creature — Merfolk Rogue", cmdr2B05ShorestalkerOracl, 1, 1)
	effectstest.AssertRestrictions(t, g, id, game.CantBeBlocked)
}

func TestCmdr2B05MistbinderPumpsOnlyYourOtherMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mb := effectstest.Cmdr2B05Card(g, me.ID, "Merfolk Mistbinder", "Creature — Merfolk Shaman", cmdr2B05MistbinderOracle, 2, 2)
	mine := effectstest.B39Creature(g, me.ID, "Mine", "Creature — Merfolk", 1, 1)
	theirs := effectstest.B39Creature(g, opp.ID, "Theirs", "Creature — Merfolk", 1, 1)
	bear := effectstest.SeedBear(g, me.ID)
	if got := effectstest.EffectivePower(t, g, mine); got != 2 {
		t.Errorf("your Merfolk power %d, want 2", got)
	}
	for name, id := range map[string]uuid.UUID{"self": mb, "opponent's": theirs} {
		want := 1
		if name == "self" {
			want = 2
		}
		if got := effectstest.EffectivePower(t, g, id); got != want {
			t.Errorf("%s power %d, want %d", name, got, want)
		}
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("bear power %d, want 2", got)
	}
}

func TestCmdr2B05SvyelunIndestructibleNeedsTwoOtherMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sv := effectstest.Cmdr2B05Card(g, me.ID, "Svyelun of Sea and Sky", "Legendary Creature — Merfolk God", cmdr2B05SvyelunOracle, 3, 4)
	if effectstest.HasKeywordOf(g, sv, "indestructible") {
		t.Fatal("alone, Svyelun is not indestructible")
	}
	effectstest.B39Creature(g, me.ID, "One", "Creature — Merfolk", 1, 1)
	if effectstest.HasKeywordOf(g, sv, "indestructible") {
		t.Fatal("one other Merfolk is not enough")
	}
	effectstest.B39Creature(g, me.ID, "Two", "Creature — Merfolk", 1, 1)
	if !effectstest.HasKeywordOf(g, sv, "indestructible") {
		t.Fatal("two other Merfolk make Svyelun indestructible")
	}
}

func TestCmdr2B05SvyelunOpponentsMerfolkDoNotCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sv := effectstest.Cmdr2B05Card(g, me.ID, "Svyelun of Sea and Sky", "Legendary Creature — Merfolk God", cmdr2B05SvyelunOracle, 3, 4)
	effectstest.B39Creature(g, opp.ID, "A", "Creature — Merfolk", 1, 1)
	effectstest.B39Creature(g, opp.ID, "B", "Creature — Merfolk", 1, 1)
	if effectstest.HasKeywordOf(g, sv, "indestructible") {
		t.Fatal("only Merfolk you control count")
	}
}

func TestCmdr2B05SvyelunAttackDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sv := effectstest.Cmdr2B05Card(g, me.ID, "Svyelun of Sea and Sky", "Legendary Creature — Merfolk God", cmdr2B05SvyelunOracle, 3, 4)
	effectstest.AdvanceToMain(t, g)
	before := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, sv)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func TestCmdr2B05SvyelunGivesOtherMerfolkWard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Cmdr2B05Card(g, me.ID, "Svyelun of Sea and Sky", "Legendary Creature — Merfolk God", cmdr2B05SvyelunOracle, 3, 4)
	merfolk := effectstest.B12Creature(g, me.ID, "Merfolk", "Creature — Merfolk", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastAtWardedCreature(t, g, opp, bear)
	for i := 0; i < 6; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Error("a non-Merfolk has ward")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastAtWardedCreature(t, g, opp, merfolk)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward prompt for a spell aimed at another Merfolk")
	}
}

func TestCmdr2B05KopalaTaxesOpponentsSpellsAtYourMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Cmdr2B05Card(g, me.ID, "Kopala, Warden of Waves", "Legendary Creature — Merfolk Wizard", cmdr2B05KopalaOracle, 2, 2)
	merfolk := effectstest.B39Creature(g, me.ID, "Merfolk", "Creature — Merfolk", 1, 1)
	bear := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.B39Creature(g, opp.ID, "Their Merfolk", "Creature — Merfolk", 1, 1)
	at := func(id uuid.UUID) []game.TargetRef { return []game.TargetRef{{Kind: game.TargetCard, ID: id}} }

	if got := effectstest.PriceOpponentsBoltAgainst(t, g, opp, at(merfolk)); got != 3 {
		t.Errorf("opponent's spell at my Merfolk costs %d, want 3", got)
	}
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, opp, at(bear)); got != 1 {
		t.Errorf("opponent's spell at my non-Merfolk costs %d, want 1", got)
	}
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, opp, at(theirs)); got != 1 {
		t.Errorf("opponent's spell at their own Merfolk costs %d, want 1", got)
	}
	if got := effectstest.PriceOpponentsBoltAgainst(t, g, me, at(merfolk)); got != 1 {
		t.Errorf("my own spell at my Merfolk costs %d, want 1", got)
	}
}

func TestCmdr2B05KopalaTaxesOpponentsAbilitiesAtYourMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Cmdr2B05Card(g, me.ID, "Kopala, Warden of Waves", "Legendary Creature — Merfolk Wizard", cmdr2B05KopalaOracle, 2, 2)
	merfolk := effectstest.B39Creature(g, me.ID, "Merfolk", "Creature — Merfolk", 1, 1)
	bear := effectstest.SeedBear(g, me.ID)
	probe := effectstest.PushCatalogPermanent(g, opp.ID, "B05 Probe", "Artifact", effectstest.Cmdr2B05ProbeOracle, false)

	if err := effectstest.StrictProbe(t, g, opp.ID, probe, "{C}", merfolk); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("an opponent's ability at my Merfolk with {1}: err = %v, want insufficient mana", err)
	}
	effectstest.DrainPool(t, g, opp.ID)
	if err := effectstest.StrictProbe(t, g, opp.ID, probe, "{C}{C}{C}", merfolk); err != nil {
		t.Fatalf("with {3} the ability is payable: %v", err)
	}
	effectstest.DrainPool(t, g, opp.ID)
	probe2 := effectstest.PushCatalogPermanent(g, opp.ID, "B05 Probe", "Artifact", effectstest.Cmdr2B05ProbeOracle, false)
	if err := effectstest.StrictProbe(t, g, opp.ID, probe2, "{C}", bear); err != nil {
		t.Fatalf("an ability at a non-Merfolk is untaxed: %v", err)
	}
}

// --- Stonybrook Banneret -------------------------------------------

func TestCmdr2B05BanneretDiscountsMerfolkAndWizardSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Cmdr2B05Card(g, me.ID, "Stonybrook Banneret", "Creature — Merfolk Wizard", cmdr2B05BanneretOracle, 1, 1)
	for _, c := range []struct {
		name, typeLine, cost string
		seat                 *game.Player
		want                 int
	}{
		{"Merfolk", "Creature — Merfolk", "{1}{U}", me, 1},
		{"Wizard", "Creature — Human Wizard", "{2}{U}", me, 2},
		{"Both", "Creature — Merfolk Wizard", "{3}", me, 2},
		{"Bear", "Creature — Bear", "{1}{G}", me, 2},
		{"Opponent's Merfolk", "Creature — Merfolk", "{1}{U}", opp, 2},
	} {
		if got := effectstest.PriceInHand(t, g, c.seat, c.name, c.typeLine, c.cost); got != c.want {
			t.Errorf("%s costs %d, want %d", c.name, got, c.want)
		}
	}
}

func TestCmdr2B05BanneretHasIslandwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	b := effectstest.Cmdr2B05Card(g, me.ID, "Stonybrook Banneret", "Creature — Merfolk Wizard", cmdr2B05BanneretOracle, 1, 1)
	bear := effectstest.PushTribalCreature(g, them.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.SeedLand(g, them.ID, "Island", "Basic Land — Island", "")
	effectstest.AttackIntoBlocks(t, g, them.ID, b)
	if r := effectstest.BlockRefusal(t, g, bear, b); r != game.BlockReasonLandwalk {
		t.Fatalf("blocking an islandwalker against an Island: reason %q, want landwalk", r)
	}
}

// --- Singer of Swift Rivers ----------------------------------------

func TestCmdr2B05SingerGrantsFlashToMerfolkSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Cmdr2B05Card(g, me.ID, "Singer of Swift Rivers", "Creature — Merfolk Shaman", cmdr2B05SingerOracle, 3, 2)
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	cast := func(name, typeLine string) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{1}{U}", Colors: []string{"U"},
			Owner: me.ID, Controller: me.ID,
		})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := cast("Bear", "Creature — Bear"); err == nil {
		t.Fatal("a non-Merfolk creature was cast in the upkeep")
	}
	if err := cast("Fish", "Creature — Merfolk"); err != nil {
		t.Fatalf("a Merfolk spell in the upkeep: %v", err)
	}
}

func TestCmdr2B05SingerEntersAndShieldsAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "Runner", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Singer of Swift Rivers", TypeLine: "Creature — Merfolk Shaman",
			OracleID: cmdr2B05SingerOracle, ManaCost: "{1}{G}{U}", Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FloatForTest(g, me, "GUU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mine, game.CounterShield); got != 1 {
		t.Errorf("the other creature has %d shield counters, want 1", got)
	}
	if got := effectstest.CounterOn(g, id, game.CounterShield); got != 0 {
		t.Errorf("the Singer shielded itself: %d", got)
	}
}

// --- Surgespanner --------------------------------------------------

func cmdr2B05SurgespannerAttack(t *testing.T) (g *game.Game, me *game.Player, victim uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	sp := effectstest.Cmdr2B05Card(g, me.ID, "Surgespanner", "Creature — Merfolk Wizard", cmdr2B05SurgespannerOracl, 2, 2)
	victim = effectstest.B12Creature(g, g.Seats[1].ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sp, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.B40PickCard(t, g, me.ID, victim)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, victim
}

func TestCmdr2B05SurgespannerPaysToBounceWhenItBecomesTapped(t *testing.T) {
	g, me, victim := cmdr2B05SurgespannerAttack(t)
	effectstest.B06AddMana(me, "U", "C")
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil {
		t.Fatal("tapping Surgespanner does not ask about {1}{U}")
	}
	if ask.PayCost != "{1}{U}" {
		t.Errorf("the optional cost is %s", ask.PayCost)
	}
	if err := g.ResolvePayUnless(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("paying returns the target permanent")
	}
}

func TestCmdr2B05SurgespannerDecliningBouncesNothing(t *testing.T) {
	g, me, victim := cmdr2B05SurgespannerAttack(t)
	effectstest.B06AddMana(me, "U", "C")
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil {
		t.Fatal("no payment prompt")
	}
	if err := g.ResolvePayUnless(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if !g.Battlefield.Contains(victim) {
		t.Error("declining bounced the permanent")
	}
}

// --- Deeproot Elite ------------------------------------------------

func TestCmdr2B05DeeprootEliteCountersATargetMerfolkWhenAnotherEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	elite := effectstest.Cmdr2B05Card(g, me.ID, "Deeproot Elite", "Creature — Merfolk Warrior", cmdr2B05DeeprootOracle, 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Other Merfolk", "Creature — Merfolk", 1, 1)

	entered := effectstest.CastCatalogSpell(t, g, "Newcomer", "Creature — Merfolk", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B40PickCard(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, other, game.CounterPlusOne); got != 1 {
		t.Errorf("the chosen Merfolk has %d counters, want 1", got)
	}
	for name, id := range map[string]uuid.UUID{"elite": elite, "newcomer": entered} {
		if got := effectstest.CounterOn(g, id, game.CounterPlusOne); got != 0 {
			t.Errorf("%s got %d counters", name, got)
		}
	}
	// A non-Merfolk entering does not trigger it.
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Error("a Bear triggered the Elite")
	}
}

// --- Kumena --------------------------------------------------------

func cmdr2B05KumenaTable(t *testing.T, others int) (g *game.Game, me *game.Player, kumena uuid.UUID, fish []uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	kumena = effectstest.Cmdr2B05Card(g, me.ID, "Kumena, Tyrant of Orazca", "Legendary Creature — Merfolk Shaman", cmdr2B05KumenaOracle, 2, 4)
	for i := 0; i < others; i++ {
		fish = append(fish, effectstest.B12Creature(g, me.ID, "Fish", "Creature — Merfolk", 1, 1))
	}
	effectstest.AdvanceToMain(t, g)
	return
}

func TestCmdr2B05KumenaTapAnotherMerfolkMakesItUnblockable(t *testing.T) {
	g, me, kumena, fish := cmdr2B05KumenaTable(t, 1)
	bear := effectstest.SeedBear(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, kumena, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{kumena}, Strict: true}); err == nil {
		t.Fatal("Kumena paid for 'another' Merfolk by tapping itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, kumena, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{bear}, Strict: true}); err == nil {
		t.Fatal("a non-Merfolk paid the cost")
	}
	effectstest.B16Activate(t, g, me.ID, kumena, 0, game.ActivateAbilityParams{TapIDs: fish})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, kumena, game.CantBeBlocked)
	if c, _ := g.LookupCardForEffect(fish[0]); !c.Tapped {
		t.Error("the Merfolk was not tapped to pay")
	}
}

func TestCmdr2B05KumenaThreeMerfolkDrawACard(t *testing.T) {
	g, me, kumena, fish := cmdr2B05KumenaTable(t, 2)
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, kumena, 1, game.ActivateAbilityParams{TapIDs: fish, Strict: true}); err == nil {
		t.Fatal("two Merfolk paid a cost of three")
	}
	effectstest.B16Activate(t, g, me.ID, kumena, 1, game.ActivateAbilityParams{TapIDs: append(append([]uuid.UUID{}, fish...), kumena)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestCmdr2B05KumenaFiveMerfolkCounterEachMerfolk(t *testing.T) {
	g, me, kumena, fish := cmdr2B05KumenaTable(t, 4)
	bear := effectstest.SeedBear(g, me.ID)
	opp := effectstest.B12Creature(g, g.Seats[1].ID, "Their Fish", "Creature — Merfolk", 1, 1)
	effectstest.B16Activate(t, g, me.ID, kumena, 2, game.ActivateAbilityParams{TapIDs: append(append([]uuid.UUID{}, fish...), kumena)})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range append(append([]uuid.UUID{}, fish...), kumena) {
		if got := effectstest.CounterOn(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("a Merfolk has %d counters, want 1", got)
		}
	}
	if effectstest.CounterOn(g, bear, game.CounterPlusOne) != 0 || effectstest.CounterOn(g, opp, game.CounterPlusOne) != 0 {
		t.Error("a non-Merfolk or an opponent's Merfolk got a counter")
	}
}

// --- Emperor Mihail II ---------------------------------------------

func TestCmdr2B05MihailCastsMerfolkFromTheTopAndPaysForAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	effectstest.PushTopSource(t, g, me.ID, "Emperor Mihail II", "Legendary Creature — Merfolk Noble", cmdr2B05MihailOracle)
	var vis game.LibraryTopVisibility
	g.ReadSnapshot(func() { vis = g.LibraryTopVisibilityLocked(me.ID) })
	if vis != game.LibraryTopOwner {
		t.Errorf("top-card visibility = %v, want owner-only", vis)
	}
	bear := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{0}"})
	if effectstest.GrantedPermissionOn(g, me.ID, bear, game.ZoneLibrary).Granted() {
		t.Error("a non-Merfolk was opened")
	}
	fish := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Fish", TypeLine: "Creature — Merfolk", ManaCost: "{0}"})
	if err := g.CastSpell(me.ID, fish, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast a Merfolk off the top: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "C")
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil {
		t.Fatal("casting a Merfolk spell does not ask about {1}")
	}
	if err := g.ResolvePayUnless(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if got := len(effectstest.BattlefieldIDsNamed(g, "Merfolk")); got != 1 {
		t.Errorf("%d Merfolk tokens, want 1", got)
	}
}

func TestCmdr2B05MihailDecliningMakesNoTokenAndBearsDoNotTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	effectstest.PushTopSource(t, g, me.ID, "Emperor Mihail II", "Legendary Creature — Merfolk Noble", cmdr2B05MihailOracle)
	effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless) != nil {
		t.Fatal("a Bear spell triggered Mihail")
	}
	effectstest.CastCatalogSpell(t, g, "Fish", "Creature — Merfolk", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "C")
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
	if ask == nil {
		t.Fatal("a Merfolk spell did not ask")
	}
	if err := g.ResolvePayUnless(ask.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	if got := len(effectstest.BattlefieldIDsNamed(g, "Merfolk")); got != 0 {
		t.Errorf("declining made %d tokens", got)
	}
}
