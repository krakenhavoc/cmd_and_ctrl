package testsds

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestEntwineBatchBCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.EwbJourneyOracle: "Journey of Discovery", effectstest.EwbKayasOracle: "Kaya's Guile",
		effectstest.EwbMirageOracle: "Mirage Mockery", effectstest.EwbDozenOracle: "One Dozen Eyes",
		effectstest.EwbPlungeOracle: "Plunge into Darkness", effectstest.EwbPromiseOracle: "Promise of Power",
		effectstest.EwbRainOracle: "Rain of Rust", effectstest.EwbReapOracle: "Reap and Sow",
		effectstest.EwbRoadOracle: "Road of Return", effectstest.EwbRoarOracle: "Roar of the Kha",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
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
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CmcSeats(t, g)
	basics := effectstest.SeedTypedLibrary(me, "Basic Land — Forest", "Forest A", "Forest B", "Forest C")
	if _, err := effectstest.CastEntwineSpell(g, me, "Journey of Discovery", "Sorcery", "{2}{G}", effectstest.EwbJourneyOracle,
		[]string{"G"}, "{C}{C}{C}{C}{G}{G}", []int{0, 1}, true, nil); err != nil {
		t.Fatalf(effectstest.EwbNoTargetsError, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	search := effectstest.SearchPromptFor(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CmcSeats(t, g)
	effectstest.SeedTypedLibrary(me, "Basic Land — Forest", "Forest A")
	if _, err := effectstest.CastEntwineSpell(g, me, "Journey of Discovery", "Sorcery", "{2}{G}", effectstest.EwbJourneyOracle,
		[]string{"G"}, "{C}{C}{G}", []int{1}, false, nil); err != nil {
		t.Fatalf(effectstest.EwbNoTargetsError, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchPromptFor(g, me.ID) != nil {
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
	_, err := effectstest.CastEntwineSpell(g, me, "Kaya's Guile", "Instant", "{1}{W}{B}", effectstest.EwbKayasOracle,
		[]string{"W", "B"}, pool, modes, entwined, nil)
	return err
}

// "Choose two", or all four entwined — never three, and never two once
// the entwine cost is announced.
func TestKayasGuileCountsTwoOrAllFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CmcSeats(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CmcSeats(t, g)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	dead := effectstest.B17GraveyardCard(opp, "Old Card", "Sorcery", "{1}")
	life := me.Life
	if err := ewbCastKaya(g, me, []int{3, 2, 1, 0}, true); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ewbSettle(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the opponent sacrificed their creature")
	}
	if opp.Graveyard.Contains(dead) || opp.Graveyard.Contains(bear) {
		t.Error("the opponent's graveyard — the sacrificed bear included — is exiled")
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Spirit"); n != 1 {
		t.Errorf("%d Spirit tokens, want 1", n)
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

// --- Mirage Mockery -------------------------------------------------------

func TestMirageMockeryEntwinedCopiesOneOfEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CmcSeats(t, g)
	myr := effectstest.TwPut(g, me.ID, "Myr", "Artifact Creature — Myr", nil)
	bear := effectstest.TwPut(g, me.ID, "Bear", "Creature — Bear", []string{"G"})
	targets := []game.TargetRef{
		{Kind: game.TargetCard, ID: myr, Mode: 0},
		{Kind: game.TargetCard, ID: bear, Mode: 1},
	}
	if _, err := effectstest.CastEntwineSpell(g, me, "Mirage Mockery", "Sorcery", "{2}{U}", effectstest.EwbMirageOracle,
		[]string{"U"}, "{C}{C}{C}{C}{U}{U}", []int{0, 1}, true, targets); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Myr"); n != 2 {
		t.Errorf("%d Myr, want the original and a token copy", n)
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Bear"); n != 2 {
		t.Errorf("%d Bears, want the original and a token copy", n)
	}
}

// The artifact bullet does not take a nonartifact creature, and the
// other way round.
func TestMirageMockeryBulletsSplitOnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CmcSeats(t, g)
	bear := effectstest.TwPut(g, me.ID, "Bear", "Creature — Bear", []string{"G"})
	wrong := []game.TargetRef{{Kind: game.TargetCard, ID: bear, Mode: 0}}
	if _, err := effectstest.CastEntwineSpell(g, me, "Mirage Mockery", "Sorcery", "{2}{U}", effectstest.EwbMirageOracle,
		[]string{"U"}, "{C}{C}{U}", []int{0}, false, wrong); err == nil {
		t.Error("the artifact bullet must refuse a nonartifact creature")
	}
	right := []game.TargetRef{{Kind: game.TargetCard, ID: bear, Mode: 0}}
	if _, err := effectstest.CastEntwineSpell(g, me, "Mirage Mockery", "Sorcery", "{2}{U}", effectstest.EwbMirageOracle,
		[]string{"U"}, "{C}{C}{U}", []int{1}, false, right); err != nil {
		t.Errorf("the nonartifact bullet takes the bear: %v", err)
	}
}

// --- One Dozen Eyes -------------------------------------------------------

func TestOneDozenEyesEntwinedMakesTheBeastAndFiveInsects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CmcSeats(t, g)
	if _, err := effectstest.CastEntwineSpell(g, me, "One Dozen Eyes", "Sorcery", "{5}{G}", effectstest.EwbDozenOracle,
		[]string{"G"}, "{C}{C}{C}{C}{C}{G}{G}{G}{G}", []int{0, 1}, true, nil); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Beast"); n != 1 {
		t.Errorf("%d Beasts, want 1", n)
	}
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Insect"); n != 5 {
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
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CmcSeats(t, g)
	a := effectstest.PushVanillaCreature(g, me.ID, "Fodder A", 1, 1)
	b := effectstest.PushVanillaCreature(g, me.ID, "Fodder B", 1, 1)
	top := effectstest.SeedTypedLibrary(me, "Sorcery", "Deep", "Deeper", "Top Three", "Top Two", "Top One")
	life := me.Life
	if _, err := effectstest.CastEntwineSpell(g, me, "Plunge into Darkness", "Instant", "{1}{B}", effectstest.EwbPlungeOracle,
		[]string{"B"}, "{C}{B}{B}", []int{0, 1}, true, nil); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	sac := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if sac == nil {
		t.Fatal("the sacrifice bullet asks which creatures")
	}
	if effectstest.PayAmountPrompt(g, me.ID) != nil {
		t.Error("the life bullet asked before the sacrifice was answered (CR 702.42b)")
	}
	if err := g.ResolveOwnPermanents(sac.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if me.Life != life+6 {
		t.Fatalf("life %d after two sacrifices, want %d", me.Life, life+6)
	}
	pay := effectstest.PayAmountPrompt(g, me.ID)
	if pay == nil {
		t.Fatal("the life bullet asks how much to pay")
	}
	if err := g.ResolvePayAmount(pay.ID, me.ID, 3); err != nil {
		t.Fatalf("pay 3: %v", err)
	}
	// top is seeded bottom-first: the last three are the top three.
	looked := top[2:]
	effectstest.AnswerChooseCards(t, g, me.ID, looked[1])
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
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CmcSeats(t, g)
	keep := effectstest.PushVanillaCreature(g, me.ID, "Keeper", 1, 1)
	life := me.Life
	if _, err := effectstest.CastEntwineSpell(g, me, "Plunge into Darkness", "Instant", "{1}{B}", effectstest.EwbPlungeOracle,
		[]string{"B"}, "{C}{B}", []int{0}, false, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	sac := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
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
	if effectstest.PayAmountPrompt(g, me.ID) != nil {
		t.Error("the unchosen life bullet asked")
	}
}

// --- Promise of Power -----------------------------------------------------

// Entwined, the five cards are drawn before X is counted (CR 702.42b).
func TestPromiseOfPowerEntwinedCountsTheDrawnCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CmcSeats(t, g)
	effectstest.SeedTypedLibrary(me, "Sorcery", "L1", "L2", "L3", "L4", "L5", "L6")
	life := me.Life
	if _, err := effectstest.CastEntwineSpell(g, me, "Promise of Power", "Sorcery", "{2}{B}{B}{B}", effectstest.EwbPromiseOracle,
		[]string{"B"}, "{C}{C}{C}{C}{C}{C}{B}{B}{B}", []int{1, 0}, true, nil); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CmcSeats(t, g)
	rock := effectstest.TwPut(g, opp.ID, "Mind Stone", "Artifact", nil)
	land := effectstest.TwPut(g, opp.ID, "Their Land", "Land", nil)
	targets := []game.TargetRef{
		{Kind: game.TargetCard, ID: rock, Mode: 0},
		{Kind: game.TargetCard, ID: land, Mode: 1},
	}
	if _, err := effectstest.CastEntwineSpell(g, me, "Rain of Rust", "Instant", "{3}{R}{R}", effectstest.EwbRainOracle,
		[]string{"R"}, "{C}{C}{C}{C}{C}{C}{R}{R}{R}", []int{0, 1}, true, targets); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(land) {
		t.Error("both targets are destroyed")
	}
}

func TestRainOfRustUnentwinedDestroysOnlyTheLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CmcSeats(t, g)
	rock := effectstest.TwPut(g, opp.ID, "Mind Stone", "Artifact", nil)
	land := effectstest.TwPut(g, opp.ID, "Their Land", "Land", nil)
	if _, err := effectstest.CastEntwineSpell(g, me, "Rain of Rust", "Instant", "{3}{R}{R}", effectstest.EwbRainOracle,
		[]string{"R"}, "{C}{C}{C}{R}{R}", []int{1}, false,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land, Mode: 0}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) || !g.Battlefield.Contains(rock) {
		t.Error("the land is destroyed and the artifact is not")
	}
}

// --- Reap and Sow ---------------------------------------------------------

func TestReapAndSowEntwinedDestroysThenFetchesUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.CmcSeats(t, g)
	theirs := effectstest.TwPut(g, opp.ID, "Their Land", "Land", nil)
	fetch := effectstest.SeedTypedLibrary(me, "Land — Urza's Tower", "Tower")
	if _, err := effectstest.CastEntwineSpell(g, me, "Reap and Sow", "Sorcery", "{3}{G}", effectstest.EwbReapOracle,
		[]string{"G"}, "{C}{C}{C}{C}{G}{G}", []int{1, 0}, true,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs, Mode: 1}}); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the target land is destroyed")
	}
	if search := effectstest.SearchPromptFor(g, me.ID); search != nil {
		if err := g.ResolveSearchLibrary(search.ID, me.ID, fetch); err != nil {
			t.Fatalf("ResolveSearchLibrary: %v", err)
		}
	}
	g.SettleResolution()
	c, ok := effectstest.BattlefieldCard(g, fetch[0])
	if !ok {
		t.Fatal("the fetched land is on the battlefield")
	}
	if c.Tapped {
		t.Error("the fetched land enters untapped")
	}
}

// --- Road of Return -------------------------------------------------------

func TestRoadOfReturnEntwinedReturnsACardAndTheCommander(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CmcSeats(t, g)
	relic := effectstest.B17GraveyardCard(me, "Old Relic", "Artifact", "{2}")
	cmdr := uuid.New()
	me.Command.PushTop(game.Card{
		InstanceID: cmdr, Name: "Kenrith, the Returned King", TypeLine: "Legendary Creature — Human Noble",
		Owner: me.ID, Controller: me.ID, IsCommander: true,
	})
	if _, err := effectstest.CastEntwineSpell(g, me, "Road of Return", "Sorcery", "{G}{G}", effectstest.EwbRoadOracle,
		[]string{"G"}, "{C}{C}{G}{G}", []int{0, 1}, true,
		[]game.TargetRef{{Kind: game.TargetCard, ID: relic, Mode: 0}}); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CmcSeats(t, g)
	bolt := effectstest.B17GraveyardCard(me, "Old Bolt", "Instant", "{R}")
	if _, err := effectstest.CastEntwineSpell(g, me, "Road of Return", "Sorcery", "{G}{G}", effectstest.EwbRoadOracle,
		[]string{"G"}, "{G}{G}", []int{0}, false,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt, Mode: 0}}); err == nil {
		t.Error("an instant card is not a permanent card")
	}
}

// --- Roar of the Kha ------------------------------------------------------

func TestRoarOfTheKhaEntwinedPumpsAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.CmcSeats(t, g)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Tired Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if _, err := effectstest.CastEntwineSpell(g, me, "Roar of the Kha", "Instant", "{1}{W}", effectstest.EwbRoarOracle,
		[]string{"W"}, "{C}{C}{W}{W}", []int{0, 1}, true, nil); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p, tough := effectstest.EffectivePTOf(t, g, bear)
	if p != 3 || tough != 3 {
		t.Errorf("the bear is %d/%d, want 3/3", p, tough)
	}
	if c, _ := effectstest.BattlefieldCard(g, bear); c.Tapped {
		t.Error("the bear is untapped")
	}
}
