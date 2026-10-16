package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entwine_cards_b_test.go — the second batch of entwine cards (CR
// 702.42, #3866): Journey of Discovery, Kaya's Guile, Mirage Mockery,
// One Dozen Eyes, Plunge into Darkness, Promise of Power, Rain of Rust,
// Reap and Sow, Road of Return and Roar of the Kha. Each entwined cast
// resolves every bullet, in the order written (CR 702.42b).

const (
	ewbJourneyOracle  = "1c586d8a-9d1a-48a7-bb3e-9b2c0c329f8d"
	ewbKayasOracle    = "a6becd5f-d279-4df8-81d9-deeb1b513c95"
	ewbMirageOracle   = "2b0c08fb-7c46-484a-a446-b26a94db7172"
	ewbDozenOracle    = "b1fbf818-6699-4f05-9a91-19aa296526bf"
	ewbPlungeOracle   = "07a39452-f9d5-41ba-acbc-fc5e7f698f64"
	ewbPromiseOracle  = "fde459b2-f45f-40d2-b328-58f65440f494"
	ewbRainOracle     = "074e467e-f574-4b44-b5a3-798b592a5a87"
	ewbReapOracle     = "efce3575-66ee-412f-af5b-961fd242bbde"
	ewbRoadOracle     = "0007c283-5b7a-4c00-9ca1-b455c8dff8c3"
	ewbRoarOracle     = "9dddb69a-00a3-448b-8f8b-8e50cdd70d57"
	ewbNoTargetsError = "cast: %v"
)

func TestEntwineBatchBCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		ewbJourneyOracle: "Journey of Discovery", ewbKayasOracle: "Kaya's Guile",
		ewbMirageOracle: "Mirage Mockery", ewbDozenOracle: "One Dozen Eyes",
		ewbPlungeOracle: "Plunge into Darkness", ewbPromiseOracle: "Promise of Power",
		ewbRainOracle: "Rain of Rust", ewbReapOracle: "Reap and Sow",
		ewbRoadOracle: "Road of Return", ewbRoarOracle: "Roar of the Kha",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s is %v, want Full", name, spec.Completeness)
		}
		if offers := game.OptionalCostsFor(oracle); len(offers) != 1 || offers[0].Key != game.EntwineKey {
			t.Errorf("%s optional costs %+v, want one entwine cost", name, offers)
		}
	}
}

// ewbSettle answers whatever the resolution left open that these tests
// do not care to steer — an opponent's edict (the first creature), a
// commander's CR 903.9b offer (declined) — until nothing is pending.
func ewbSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for range 16 {
		var c *game.PendingChoice
		for _, pc := range g.PendingChoices {
			if pc != nil && (pc.Kind == game.PendingChoiceSacrifice || pc.Kind == game.PendingChoiceOptionalReplacement) {
				c = pc
				break
			}
		}
		if c == nil {
			break
		}
		var err error
		if c.Kind == game.PendingChoiceSacrifice {
			err = g.ResolveSacrificeChoice(c.ID, c.Chooser, c.SacrificeOptions[0])
		} else {
			err = g.ResolveOptionalReplacement(c.ID, c.Chooser, false)
		}
		if err != nil {
			t.Fatalf("settling %v: %v", c.Kind, err)
		}
	}
	g.SettleResolution()
}

// --- Journey of Discovery ------------------------------------------------

func TestJourneyOfDiscoveryEntwinedFetchesTwoBasicsThenGrantsTwoLandPlays(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	basics := seedTypedLibrary(me, "Basic Land — Forest", "Forest A", "Forest B", "Forest C")
	if _, err := castEntwineSpell(g, me, "Journey of Discovery", "Sorcery", "{2}{G}", ewbJourneyOracle,
		[]string{"G"}, "{C}{C}{C}{C}{G}{G}", []int{0, 1}, true, nil); err != nil {
		t.Fatalf(ewbNoTargetsError, err)
	}
	passPriorityAroundTable(t, g)
	search := searchPromptFor(g, me.ID)
	if search == nil {
		t.Fatal("the search bullet asks which basics")
	}
	if g.ExtraLandDropsThisTurn[me.ID] != 0 {
		t.Error("the land-play bullet ran before the search was answered (CR 702.42b)")
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, basics[:2]); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	g.SettleResolution()
	if !me.Hand.Contains(basics[0]) || !me.Hand.Contains(basics[1]) {
		t.Error("the two basics are in hand")
	}
	if got := g.ExtraLandDropsThisTurn[me.ID]; got != 2 {
		t.Errorf("additional land plays %d, want 2", got)
	}
}

func TestJourneyOfDiscoveryUnentwinedGrantsOnlyTheLandPlays(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	seedTypedLibrary(me, "Basic Land — Forest", "Forest A")
	if _, err := castEntwineSpell(g, me, "Journey of Discovery", "Sorcery", "{2}{G}", ewbJourneyOracle,
		[]string{"G"}, "{C}{C}{G}", []int{1}, false, nil); err != nil {
		t.Fatalf(ewbNoTargetsError, err)
	}
	passPriorityAroundTable(t, g)
	if searchPromptFor(g, me.ID) != nil {
		t.Error("the unchosen search bullet asked")
	}
	if got := g.ExtraLandDropsThisTurn[me.ID]; got != 2 {
		t.Errorf("additional land plays %d, want 2", got)
	}
}

// --- Kaya's Guile ---------------------------------------------------------

func ewbCastKaya(g *game.Game, me *game.Player, modes []int, entwined bool) error {
	pool := "{C}{W}{B}"
	if entwined {
		pool = "{C}{C}{C}{C}{W}{B}"
	}
	_, err := castEntwineSpell(g, me, "Kaya's Guile", "Instant", "{1}{W}{B}", ewbKayasOracle,
		[]string{"W", "B"}, pool, modes, entwined, nil)
	return err
}

// "Choose two", or all four entwined — never three, and never two once
// the entwine cost is announced.
func TestKayasGuileCountsTwoOrAllFour(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	if err := ewbCastKaya(g, me, []int{0, 1, 2}, false); err == nil {
		t.Error("unentwined: three bullets must be refused")
	}
	if err := ewbCastKaya(g, me, []int{0, 1, 2}, true); err == nil {
		t.Error("entwined: three bullets must be refused")
	}
	if err := ewbCastKaya(g, me, []int{2, 3}, true); err == nil {
		t.Error("entwined: two bullets must be refused")
	}
	if err := ewbCastKaya(g, me, []int{0, 1, 2, 3}, false); err == nil {
		t.Error("unentwined: four bullets must be refused")
	}
	if err := ewbCastKaya(g, me, []int{2, 3}, false); err != nil {
		t.Errorf("unentwined: two bullets are legal: %v", err)
	}
	if err := ewbCastKaya(g, me, []int{0, 1, 2, 3}, true); err != nil {
		t.Errorf("entwined: all four are legal: %v", err)
	}
}

// Entwined, in printed order: the opponent's creature is sacrificed
// FIRST, so the graveyard exile that follows takes it too.
func TestKayasGuileEntwinedDoesAllFourInOrder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	bear := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	dead := b17GraveyardCard(opp, "Old Card", "Sorcery", "{1}")
	life := me.Life
	if err := ewbCastKaya(g, me, []int{3, 2, 1, 0}, true); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	ewbSettle(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the opponent sacrificed their creature")
	}
	if opp.Graveyard.Contains(dead) || opp.Graveyard.Contains(bear) {
		t.Error("the opponent's graveyard — the sacrificed bear included — is exiled")
	}
	if n := countBattlefieldNamed(g, me.ID, "Spirit"); n != 1 {
		t.Errorf("%d Spirit tokens, want 1", n)
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

// --- Mirage Mockery -------------------------------------------------------

func TestMirageMockeryEntwinedCopiesOneOfEach(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	myr := twPut(g, me.ID, "Myr", "Artifact Creature — Myr", nil)
	bear := twPut(g, me.ID, "Bear", "Creature — Bear", []string{"G"})
	targets := []game.TargetRef{
		{Kind: game.TargetCard, ID: myr, Mode: 0},
		{Kind: game.TargetCard, ID: bear, Mode: 1},
	}
	if _, err := castEntwineSpell(g, me, "Mirage Mockery", "Sorcery", "{2}{U}", ewbMirageOracle,
		[]string{"U"}, "{C}{C}{C}{C}{U}{U}", []int{0, 1}, true, targets); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Myr"); n != 2 {
		t.Errorf("%d Myr, want the original and a token copy", n)
	}
	if n := countBattlefieldNamed(g, me.ID, "Bear"); n != 2 {
		t.Errorf("%d Bears, want the original and a token copy", n)
	}
}

// The artifact bullet does not take a nonartifact creature, and the
// other way round.
func TestMirageMockeryBulletsSplitOnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	bear := twPut(g, me.ID, "Bear", "Creature — Bear", []string{"G"})
	wrong := []game.TargetRef{{Kind: game.TargetCard, ID: bear, Mode: 0}}
	if _, err := castEntwineSpell(g, me, "Mirage Mockery", "Sorcery", "{2}{U}", ewbMirageOracle,
		[]string{"U"}, "{C}{C}{U}", []int{0}, false, wrong); err == nil {
		t.Error("the artifact bullet must refuse a nonartifact creature")
	}
	right := []game.TargetRef{{Kind: game.TargetCard, ID: bear, Mode: 0}}
	if _, err := castEntwineSpell(g, me, "Mirage Mockery", "Sorcery", "{2}{U}", ewbMirageOracle,
		[]string{"U"}, "{C}{C}{U}", []int{1}, false, right); err != nil {
		t.Errorf("the nonartifact bullet takes the bear: %v", err)
	}
}

// --- One Dozen Eyes -------------------------------------------------------

func TestOneDozenEyesEntwinedMakesTheBeastAndFiveInsects(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	if _, err := castEntwineSpell(g, me, "One Dozen Eyes", "Sorcery", "{5}{G}", ewbDozenOracle,
		[]string{"G"}, "{C}{C}{C}{C}{C}{G}{G}{G}{G}", []int{0, 1}, true, nil); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := countBattlefieldNamed(g, me.ID, "Beast"); n != 1 {
		t.Errorf("%d Beasts, want 1", n)
	}
	if n := countBattlefieldNamed(g, me.ID, "Insect"); n != 5 {
		t.Errorf("%d Insects, want 5", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Beast" && c.Controller == me.ID && (c.Power != 5 || c.Toughness != 5) {
			t.Errorf("the Beast is %d/%d, want 5/5", c.Power, c.Toughness)
		}
	}
}

// --- Plunge into Darkness -------------------------------------------------

// Entwined: two creatures sacrificed for 6 life, then 3 life paid to
// look at three cards, one kept and two exiled.
func TestPlungeIntoDarknessEntwinedSacrificesThenLooks(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	a := pushVanillaCreature(g, me.ID, "Fodder A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "Fodder B", 1, 1)
	top := seedTypedLibrary(me, "Sorcery", "Deep", "Deeper", "Top Three", "Top Two", "Top One")
	life := me.Life
	if _, err := castEntwineSpell(g, me, "Plunge into Darkness", "Instant", "{1}{B}", ewbPlungeOracle,
		[]string{"B"}, "{C}{B}{B}", []int{0, 1}, true, nil); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	sac := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if sac == nil {
		t.Fatal("the sacrifice bullet asks which creatures")
	}
	if payAmountPrompt(g, me.ID) != nil {
		t.Error("the life bullet asked before the sacrifice was answered (CR 702.42b)")
	}
	if err := g.ResolveOwnPermanents(sac.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if me.Life != life+6 {
		t.Fatalf("life %d after two sacrifices, want %d", me.Life, life+6)
	}
	pay := payAmountPrompt(g, me.ID)
	if pay == nil {
		t.Fatal("the life bullet asks how much to pay")
	}
	if err := g.ResolvePayAmount(pay.ID, me.ID, 3); err != nil {
		t.Fatalf("pay 3: %v", err)
	}
	// top is seeded bottom-first: the last three are the top three.
	looked := top[2:]
	answerChooseCards(t, g, me.ID, looked[1])
	g.SettleResolution()
	if me.Life != life+3 {
		t.Errorf("life %d, want %d (+6, then -3)", me.Life, life+3)
	}
	if !me.Hand.Contains(looked[1]) {
		t.Error("the chosen card is in hand")
	}
	for _, id := range []uuid.UUID{looked[0], looked[2]} {
		if me.Library.Contains(id) || me.Hand.Contains(id) {
			t.Errorf("%s was not exiled", id)
		}
	}
	if !me.Library.Contains(top[0]) || !me.Library.Contains(top[1]) {
		t.Error("cards below the three stay in the library")
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("both creatures were sacrificed")
	}
}

// Unentwined, the sacrifice alone: zero is an answer, and gains nothing.
func TestPlungeIntoDarknessSacrificingNothingGainsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	keep := pushVanillaCreature(g, me.ID, "Keeper", 1, 1)
	life := me.Life
	if _, err := castEntwineSpell(g, me, "Plunge into Darkness", "Instant", "{1}{B}", ewbPlungeOracle,
		[]string{"B"}, "{C}{B}", []int{0}, false, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	sac := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if sac == nil {
		t.Fatal("the sacrifice bullet asks")
	}
	if err := g.ResolveOwnPermanents(sac.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveOwnPermanents(empty): %v", err)
	}
	g.SettleResolution()
	if me.Life != life || !g.Battlefield.Contains(keep) {
		t.Errorf("life %d (want %d), keeper on battlefield %v", me.Life, life, g.Battlefield.Contains(keep))
	}
	if payAmountPrompt(g, me.ID) != nil {
		t.Error("the unchosen life bullet asked")
	}
}

// --- Promise of Power -----------------------------------------------------

// Entwined, the five cards are drawn before X is counted (CR 702.42b).
func TestPromiseOfPowerEntwinedCountsTheDrawnCards(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	seedTypedLibrary(me, "Sorcery", "L1", "L2", "L3", "L4", "L5", "L6")
	life := me.Life
	if _, err := castEntwineSpell(g, me, "Promise of Power", "Sorcery", "{2}{B}{B}{B}", ewbPromiseOracle,
		[]string{"B"}, "{C}{C}{C}{C}{C}{C}{B}{B}{B}", []int{1, 0}, true, nil); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+5 {
		t.Fatalf("hand %d, want %d", me.Hand.Size(), hand+5)
	}
	if me.Life != life-5 {
		t.Errorf("life %d, want %d", me.Life, life-5)
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Demon" && c.Controller == me.ID {
			found = true
			if c.Power != hand+5 || c.Toughness != hand+5 {
				t.Errorf("the Demon is %d/%d, want %d/%d — X is read after the draw", c.Power, c.Toughness, hand+5, hand+5)
			}
			if !slices.Contains(c.Keywords, "flying") {
				t.Error("the Demon has flying")
			}
		}
	}
	if !found {
		t.Error("no Demon was created")
	}
}

// --- Rain of Rust ---------------------------------------------------------

func TestRainOfRustEntwinedDestroysBoth(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	rock := twPut(g, opp.ID, "Mind Stone", "Artifact", nil)
	land := twPut(g, opp.ID, "Their Land", "Land", nil)
	targets := []game.TargetRef{
		{Kind: game.TargetCard, ID: rock, Mode: 0},
		{Kind: game.TargetCard, ID: land, Mode: 1},
	}
	if _, err := castEntwineSpell(g, me, "Rain of Rust", "Instant", "{3}{R}{R}", ewbRainOracle,
		[]string{"R"}, "{C}{C}{C}{C}{C}{C}{R}{R}{R}", []int{0, 1}, true, targets); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(land) {
		t.Error("both targets are destroyed")
	}
}

func TestRainOfRustUnentwinedDestroysOnlyTheLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	rock := twPut(g, opp.ID, "Mind Stone", "Artifact", nil)
	land := twPut(g, opp.ID, "Their Land", "Land", nil)
	if _, err := castEntwineSpell(g, me, "Rain of Rust", "Instant", "{3}{R}{R}", ewbRainOracle,
		[]string{"R"}, "{C}{C}{C}{R}{R}", []int{1}, false,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land, Mode: 0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !g.Battlefield.Contains(rock) {
		t.Error("the land is destroyed and the artifact is not")
	}
}

// --- Reap and Sow ---------------------------------------------------------

func TestReapAndSowEntwinedDestroysThenFetchesUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	theirs := twPut(g, opp.ID, "Their Land", "Land", nil)
	fetch := seedTypedLibrary(me, "Land — Urza's Tower", "Tower")
	if _, err := castEntwineSpell(g, me, "Reap and Sow", "Sorcery", "{3}{G}", ewbReapOracle,
		[]string{"G"}, "{C}{C}{C}{C}{G}{G}", []int{1, 0}, true,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs, Mode: 1}}); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the target land is destroyed")
	}
	if search := searchPromptFor(g, me.ID); search != nil {
		if err := g.ResolveSearchLibrary(search.ID, me.ID, fetch); err != nil {
			t.Fatalf("ResolveSearchLibrary: %v", err)
		}
	}
	g.SettleResolution()
	c, ok := battlefieldCard(g, fetch[0])
	if !ok {
		t.Fatal("the fetched land is on the battlefield")
	}
	if c.Tapped {
		t.Error("the fetched land enters untapped")
	}
}

// --- Road of Return -------------------------------------------------------

func TestRoadOfReturnEntwinedReturnsACardAndTheCommander(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	relic := b17GraveyardCard(me, "Old Relic", "Artifact", "{2}")
	cmdr := uuid.New()
	me.Command.PushTop(game.Card{
		InstanceID: cmdr, Name: "Kenrith, the Returned King", TypeLine: "Legendary Creature — Human Noble",
		Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	if _, err := castEntwineSpell(g, me, "Road of Return", "Sorcery", "{G}{G}", ewbRoadOracle,
		[]string{"G"}, "{C}{C}{G}{G}", []int{0, 1}, true,
		[]game.TargetRef{{Kind: game.TargetCard, ID: relic, Mode: 0}}); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	ewbSettle(t, g)
	if !me.Hand.Contains(relic) {
		t.Error("the permanent card returned to hand")
	}
	if !me.Hand.Contains(cmdr) || me.Command.Contains(cmdr) {
		t.Error("the commander is in hand")
	}
}

// A nonpermanent card in the graveyard is not a legal target.
func TestRoadOfReturnTargetsOnlyAPermanentCard(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	bolt := b17GraveyardCard(me, "Old Bolt", "Instant", "{R}")
	if _, err := castEntwineSpell(g, me, "Road of Return", "Sorcery", "{G}{G}", ewbRoadOracle,
		[]string{"G"}, "{G}{G}", []int{0}, false,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt, Mode: 0}}); err == nil {
		t.Error("an instant card is not a permanent card")
	}
}

// --- Roar of the Kha ------------------------------------------------------

func TestRoarOfTheKhaEntwinedPumpsAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := cmcSeats(t, g)
	bear := pushVanillaCreature(g, me.ID, "Tired Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if _, err := castEntwineSpell(g, me, "Roar of the Kha", "Instant", "{1}{W}", ewbRoarOracle,
		[]string{"W"}, "{C}{C}{W}{W}", []int{0, 1}, true, nil); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	p, tough := effectivePTOf(t, g, bear)
	if p != 3 || tough != 3 {
		t.Errorf("the bear is %d/%d, want 3/3", p, tough)
	}
	if c, _ := battlefieldCard(g, bear); c.Tapped {
		t.Error("the bear is untapped")
	}
}
