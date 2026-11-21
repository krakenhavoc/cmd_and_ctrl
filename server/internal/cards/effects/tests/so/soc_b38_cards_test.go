package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b38_cards_test.go — Secrets of Strixhaven, batch 38 (#4456). Each test
// plays the real catalog entry through the engine and asserts what the
// printed text says.

const (
	socB38ThornfistOracle    = "d0d13d5d-677d-4e1a-9e58-2f9840581f6d"
	socB38SoloistOracle      = "43b5794a-4ed6-471d-bc45-22fac486895d"
	socB38GraveOracle        = "595f0eb5-f521-4174-9c48-b89e85ea907c"
	socB38TomeBlastOracle    = "f0163420-2f91-493c-bad4-2cd33e9856bf"
	socB38LecturerOracle     = "5f5fe039-4be4-42ae-9fa4-4392328aa665"
	socB38ArchaicOracle      = "fdd4b3a9-83ce-41bf-82e2-7808657e2c09"
	socB38ShopkeepOracle     = "9da7f9ec-872a-471d-8d60-5f4213f0204f"
	socB38MockeryOracle      = "9d059345-1a2f-4ecb-b419-3873ebadf318"
	socB38ScavengerOracle    = "3a97bc8d-bd81-468e-9277-eba4fac11ee0"
	socB38OutburstOracle     = "a0cbb1ba-73b1-4418-bbe3-5d1a02007c08"
	socB38BearTypeLine       = "Creature — Bear"
	socB38FlashbackAltCostID = "flashback"
)

func TestSocB38EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB38ThornfistOracle:        "Thornfist Striker",
		socB38SoloistOracle:          "Thunderdrum Soloist",
		socB38GraveOracle:            "Titan's Grave",
		socB38TomeBlastOracle:        "Tome Blast",
		socB38LecturerOracle:         "Topiary Lecturer",
		socB38ArchaicOracle:          "Transcendent Archaic",
		socB38ShopkeepOracle:         "Ulna Alley Shopkeep",
		socB38MockeryOracle:          "Unsubtle Mockery",
		socB38ScavengerOracle:        "Vastlands Scavenger",
		socB38ScavengerOracle + "#1": "Bind to Life",
		socB38OutburstOracle:         "Vibrant Outburst",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// socB38GainLife makes `p` gain n life the way an effect does, through the
// engine, so the turn tally and the layer cache hear about it.
func socB38GainLife(t *testing.T, g *game.Game, p *game.Player, source uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := (effects.GainLife{Player: p.ID, Amount: n}).Apply(effects.NewContext(g, &game.StackItem{Controller: p.ID, SourceCardID: source})); err != nil {
			t.Fatalf("GainLife: %v", err)
		}
	})
}

// --- Thornfist Striker ------------------------------------------------------

func TestSocB38ThornfistInfusionPumpsAndTramplesYourCreaturesOnlyAfterYouGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	striker := effectstest.B12Push(g, me.ID, "Thornfist Striker", "Creature — Elf Druid", socB38ThornfistOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB38BearTypeLine, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", socB38BearTypeLine, 2, 2)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Fatalf("before any life gain the bear is power %d, trample %v; want 2, false", p, effectstest.HasEffectiveKeyword(t, g, bear, "trample"))
	}
	// An opponent gaining life is not you gaining life.
	socB38GainLife(t, g, opp, theirs, 2)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("an opponent's life gain pumped the bear to %d", p)
	}
	socB38GainLife(t, g, me, striker, 1)
	if p := effectstest.EffectivePower(t, g, striker); p != 4 {
		t.Errorf("the Striker is power %d after you gained life, want 4", p)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 3 || !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Errorf("the bear is power %d, trample %v; want 3, true", p, effectstest.HasEffectiveKeyword(t, g, bear, "trample"))
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Errorf("their bear is power %d, trample %v; want 2, false", p, effectstest.HasEffectiveKeyword(t, g, theirs, "trample"))
	}
	late := effectstest.B12Creature(g, me.ID, "Late Bear", socB38BearTypeLine, 1, 1)
	if p := effectstest.EffectivePower(t, g, late); p != 2 || !effectstest.HasEffectiveKeyword(t, g, late, "trample") {
		t.Errorf("a creature that enters later is power %d, trample %v; want 2, true", p, effectstest.HasEffectiveKeyword(t, g, late, "trample"))
	}
	effectstest.EndTurn(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Errorf("next turn the bear is power %d, trample %v; want 2, false", p, effectstest.HasEffectiveKeyword(t, g, bear, "trample"))
	}
}

func TestSocB38ThornfistWardOneCountersAnUnpaidSpell(t *testing.T) {
	g, _, opp, striker := effectstest.WardTable(t, "Thornfist Striker", "Creature — Elf Druid", socB38ThornfistOracle)
	bolt := effectstest.BoltWardedCreature(t, g, opp, striker)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the opponent's spell")
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BoltResolved(g, striker) || !opp.Graveyard.Contains(bolt) {
		t.Fatal("a declined ward did not counter the spell")
	}
}

// --- Ulna Alley Shopkeep ----------------------------------------------------

func TestSocB38ShopkeepGetsPlusTwoPowerOnlyWhileYouGainedLifeThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shop := effectstest.B12Push(g, me.ID, "Ulna Alley Shopkeep", "Creature — Goblin Warlock", socB38ShopkeepOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB38BearTypeLine, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, shop, "menace") {
		t.Error("the Shopkeep has no menace")
	}
	if p := effectstest.EffectivePower(t, g, shop); p != 2 {
		t.Fatalf("power %d before any life gain, want 2", p)
	}
	socB38GainLife(t, g, opp, shop, 4)
	if p := effectstest.EffectivePower(t, g, shop); p != 2 {
		t.Errorf("an opponent's life gain gave power %d, want 2", p)
	}
	socB38GainLife(t, g, me, shop, 1)
	if p, tg := effectstest.SocB14PT(t, g, shop); p != 4 || tg != 3 {
		t.Errorf("after you gained life it is %d/%d, want 4/3", p, tg)
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("another creature got power %d, want 2 (this creature only)", p)
	}
	effectstest.EndTurn(t, g)
	if p := effectstest.EffectivePower(t, g, shop); p != 2 {
		t.Errorf("next turn the power is %d, want 2", p)
	}
}

// --- Thunderdrum Soloist ----------------------------------------------------

func socB38OpponentLives(g *game.Game) []int {
	var out []int
	for _, s := range g.Seats[1:] {
		out = append(out, s.Life)
	}
	return out
}

func TestSocB38SoloistOpusDealsOneOrThreeToEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Thunderdrum Soloist", "Creature — Dwarf Bard", socB38SoloistOracle, 1, 3)
	effectstest.AdvanceToMain(t, g)
	before := socB38OpponentLives(g)
	if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Four", TypeLine: "Sorcery", ManaCost: "{3}{R}"}, "R", 4, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, life := range socB38OpponentLives(g) {
		if life != before[i]-1 {
			t.Errorf("opponent %d is at %d after a four-mana sorcery, want %d", i+1, life, before[i]-1)
		}
	}
	mid := socB38OpponentLives(g)
	if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Five", TypeLine: "Instant", ManaCost: "{4}{R}"}, "R", 5, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, life := range socB38OpponentLives(g) {
		if life != mid[i]-3 {
			t.Errorf("opponent %d is at %d after a five-mana instant, want %d (3 instead of 1)", i+1, life, mid[i]-3)
		}
	}
}

func TestSocB38SoloistIgnoresACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Thunderdrum Soloist", "Creature — Dwarf Bard", socB38SoloistOracle, 1, 3)
	effectstest.AdvanceToMain(t, g)
	before := socB38OpponentLives(g)
	if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Bear", TypeLine: socB38BearTypeLine, ManaCost: "{1}{R}", Power: 2, Toughness: 2}, "R", 2, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, life := range socB38OpponentLives(g) {
		if life != before[i] {
			t.Errorf("opponent %d lost life to a creature spell: %d -> %d", i+1, before[i], life)
		}
	}
}

func TestSocB38SoloistHasReach(t *testing.T) {
	spec, _ := effects.Lookup(socB38SoloistOracle)
	if !effectstest.HasKeyword(spec.PrintedKeywords, "reach") {
		t.Errorf("printed keywords %v lack reach", spec.PrintedKeywords)
	}
}

// --- Titan's Grave ----------------------------------------------------------

func TestSocB38GraveEntersTappedMakesBOrGAndSurveilsForFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Top", "Next")
	id := effectstest.PlayLandFromHand(t, g, "Titan's Grave", socB38GraveOracle)
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatalf("on battlefield %v, tapped %v; want it to enter tapped", ok, card.Tapped)
	}
	effectstest.UntapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil {
		t.Fatal("the land made {U}")
	}
	for _, c := range []string{"B", "G"} {
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{c}}); err != nil {
			t.Fatalf("{T}: Add {%s}: %v", c, err)
		}
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		effectstest.UntapForTest(g, id)
	}
	effectstest.FloatForTest(g, me, "CCBB")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("surveil activated without {G} in the pool")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "CCBG")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{2}{B}{G},{T}: surveil 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil prompt %v, want one card", c)
	}
	top := c.ScryCards[0]
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not go to the graveyard")
	}
}

// --- Tome Blast -------------------------------------------------------------

func TestSocB38TomeBlastDealsTwoToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", socB38BearTypeLine, 2, 2)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Tome Blast", "Sorcery", socB38TomeBlastOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-2)
	}
	effectstest.CastCatalogSpell(t, g, "Tome Blast", "Sorcery", socB38TomeBlastOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 survived two damage")
	}
}

func TestSocB38TomeBlastFlashbackCastsFromTheGraveyardAndIsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Tome Blast", TypeLine: "Sorcery", ManaCost: "{1}{R}",
		OracleID: socB38TomeBlastOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPoolColored(me, "R", 4)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneGraveyard),
		AlternativeCost: socB38FlashbackAltCostID, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("flashback {4}{R} was paid with only four mana")
	}
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneGraveyard),
		AlternativeCost: socB38FlashbackAltCostID, Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent at %d after the flashback, want %d", opp.Life, life-2)
	}
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("the flashed-back spell was not exiled")
	}
}

// --- Topiary Lecturer -------------------------------------------------------

func TestSocB38LecturerIncrementsAndTapsForGEqualToItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lecturer := effectstest.PushCatalogPermanent(g, me.ID, "Topiary Lecturer", "Creature — Elf Druid", socB38LecturerOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == lecturer {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 1, 2
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, lecturer, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("power 1: pool %v, want [G]", got)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.B08Untap(g, lecturer)
	effectstest.AdvanceToMain(t, g)
	// A three-mana spell is more than its power 1 and its toughness 2.
	if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Three", TypeLine: "Sorcery", ManaCost: "{2}{G}"}, "G", 3, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, lecturer, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d +1/+1 counters after a three-mana spell, want 1", n)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateManaAbility(me.ID, lecturer, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("power 2: pool %v, want [G G]", got)
	}
}

// --- Transcendent Archaic ---------------------------------------------------

// socB38CastArchaic casts Transcendent Archaic with the given pool and lets
// the spell resolve, leaving its enters trigger's question open (if any).
func socB38CastArchaic(t *testing.T, g *game.Game, pool string) (*game.Player, int) {
	t.Helper()
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, pool)
	c := game.Card{Name: "Transcendent Archaic", TypeLine: "Creature — Avatar", ManaCost: "{7}", Power: 6, Toughness: 6,
		OracleID: socB38ArchaicOracle, Owner: me.ID, Controller: me.ID, InstanceID: uuid.New()}
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	handAfterCast := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	return me, handAfterCast
}

func TestSocB38ArchaicDrawsOnePerColorSpentThenDiscardsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, base := socB38CastArchaic(t, g, "WUBCCCC") // three colors of mana
	ch := effectstest.LatestConfirmFor(g, me.ID)
	if ch == nil {
		t.Fatal("no \"you may draw\" question")
	}
	if err := g.ResolveConfirm(ch.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if got := me.Hand.Size(); got != base+3 {
		t.Fatalf("hand %d after drawing, want %d (three colors spent)", got, base+3)
	}
	if owed := effectstest.DiscardOwed(g, me.ID); owed != 2 {
		t.Fatalf("owes %d discards, want 2", owed)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if got := me.Hand.Size(); got != base+1 {
		t.Errorf("hand %d after the discard, want %d", got, base+1)
	}
}

func TestSocB38ArchaicOneColorDrawsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, base := socB38CastArchaic(t, g, "WWWWWWW")
	ch := effectstest.LatestConfirmFor(g, me.ID)
	if ch == nil {
		t.Fatal("no question")
	}
	if err := g.ResolveConfirm(ch.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if got := me.Hand.Size(); got != base+1 {
		t.Errorf("hand %d after drawing, want %d", got, base+1)
	}
	if owed := effectstest.DiscardOwed(g, me.ID); owed != 2 {
		t.Errorf("owes %d discards, want 2", owed)
	}
}

func TestSocB38ArchaicMayDeclineAndNothingIsDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, base := socB38CastArchaic(t, g, "WUBCCCC")
	ch := effectstest.LatestConfirmFor(g, me.ID)
	if ch == nil {
		t.Fatal("no question")
	}
	if err := g.ResolveConfirm(ch.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if got := me.Hand.Size(); got != base {
		t.Errorf("hand %d after declining, want %d", got, base)
	}
	if owed := effectstest.DiscardOwed(g, me.ID); owed != 0 {
		t.Errorf("owes %d discards after declining, want 0", owed)
	}
}

func TestSocB38ArchaicColorlessPaymentDrawsNothingAndAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := socB38CastArchaic(t, g, "CCCCCCC")
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("asked about drawing zero cards")
	}
	if owed := effectstest.DiscardOwed(g, me.ID); owed != 0 {
		t.Errorf("owes %d discards, want 0", owed)
	}
}

// --- Unsubtle Mockery -------------------------------------------------------

func TestSocB38MockeryKillsAFourToughnessCreatureThenSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "Top", "Next")
	big := effectstest.B12Creature(g, opp.ID, "Big", socB38BearTypeLine, 4, 4)
	effectstest.CastCatalogSpell(t, g, "Unsubtle Mockery", "Instant", socB38MockeryOracle, []game.TargetRef{{Kind: game.TargetCard, ID: big}})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil prompt %v, want one card", c)
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("a 4/4 survived four damage")
	}
}

func TestSocB38MockerySurveilsEvenWhenNothingDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "Top", "Next")
	tough := effectstest.B12Creature(g, opp.ID, "Tough", socB38BearTypeLine, 5, 5)
	effectstest.CastCatalogSpell(t, g, "Unsubtle Mockery", "Instant", socB38MockeryOracle, []game.TargetRef{{Kind: game.TargetCard, ID: tough}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tough) {
		t.Error("a 5/5 died to four damage")
	}
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Fatal("no surveil although the spell resolved")
	}
}

func TestSocB38MockeryDoesNothingWhenItsTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", socB38BearTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Unsubtle Mockery", "Instant", socB38MockeryOracle, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, victim); err != nil {
		t.Fatalf("remove the target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Error("the spell surveilled although its only target was gone")
	}
}

// --- Vastlands Scavenger // Bind to Life ------------------------------------

func socB38ScavengerCard(owner uuid.UUID) game.Card {
	return effectstest.PreparationCard(owner, socB38ScavengerOracle, []string{"deathtouch"},
		game.Face{Name: "Vastlands Scavenger", TypeLine: "Creature — Bear Druid", ManaCost: "{1}{G}{G}", Power: 4, Toughness: 4},
		game.Face{Name: "Bind to Life", TypeLine: "Instant", ManaCost: "{4}{G}"})
}

func TestSocB38ScavengerEntersPreparedWithDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := socB38ScavengerCard(me.ID)
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Vastlands Scavenger did not enter prepared")
	}
	if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "deathtouch") {
		t.Error("no deathtouch")
	}
}

func TestSocB38BindToLifeMillsSevenAndPutsAChosenCreatureOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := socB38ScavengerCard(me.ID)
	effectstest.CastFromHandAndResolve(t, g, me, c)
	me.Library.Cards = nil
	ids := effectstest.ZnrStackLibrary(me,
		game.Card{Name: "Bear One", TypeLine: socB38BearTypeLine, Power: 2, Toughness: 2},
		game.Card{Name: "Land", TypeLine: "Land"},
		game.Card{Name: "Bear Two", TypeLine: socB38BearTypeLine, Power: 3, Toughness: 3},
		game.Card{Name: "Spell", TypeLine: "Sorcery"},
		game.Card{Name: "Land B", TypeLine: "Land"},
		game.Card{Name: "Spell B", TypeLine: "Instant"},
		game.Card{Name: "Land C", TypeLine: "Land"},
		game.Card{Name: "Eighth", TypeLine: socB38BearTypeLine, Power: 1, Toughness: 1},
	)
	effectstest.FraBCastCopy(t, g, me, "Bind to Life")
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no creature pick after the mill")
	}
	if len(prompt.ChooseCards) != 2 || !effectstest.HasID(prompt.ChooseCards, ids[0]) || !effectstest.HasID(prompt.ChooseCards, ids[2]) {
		t.Errorf("offered %v, want exactly the two milled Bears (the eighth card was not milled)", prompt.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[2])
	if !g.Battlefield.Contains(ids[2]) {
		t.Error("the chosen Bear is not on the battlefield")
	}
	if !me.Graveyard.Contains(ids[0]) || me.Graveyard.Size() < 6 {
		t.Errorf("the rest of the milled cards stay in the graveyard (size %d)", me.Graveyard.Size())
	}
	if me.Library.Size() != 1 || !me.Library.Contains(ids[7]) {
		t.Errorf("seven cards were not milled: library %d", me.Library.Size())
	}
}

func TestSocB38BindToLifeWithNoCreatureAmongTheMilledCardsDoesNothingMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := socB38ScavengerCard(me.ID)
	effectstest.CastFromHandAndResolve(t, g, me, c)
	me.Library.Cards = nil
	var cards []game.Card
	for i := 0; i < 7; i++ {
		cards = append(cards, game.Card{Name: "Land", TypeLine: "Land"})
	}
	cards = append(cards, game.Card{Name: "Hidden Bear", TypeLine: socB38BearTypeLine, Power: 2, Toughness: 2})
	ids := effectstest.ZnrStackLibrary(me, cards...)
	before := len(g.Battlefield.Cards)
	effectstest.FraBCastCopy(t, g, me, "Bind to Life")
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a pick was offered with no creature among the milled cards")
	}
	if len(g.Battlefield.Cards) != before || g.Battlefield.Contains(ids[7]) {
		t.Error("a creature entered the battlefield")
	}
}

func TestSocB38BindToLifeALoneMilledCreatureComesBackWithoutAsking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := socB38ScavengerCard(me.ID)
	effectstest.CastFromHandAndResolve(t, g, me, c)
	me.Library.Cards = nil
	var cards []game.Card
	cards = append(cards, game.Card{Name: "Only Bear", TypeLine: socB38BearTypeLine, Power: 2, Toughness: 2})
	for i := 0; i < 7; i++ {
		cards = append(cards, game.Card{Name: "Land", TypeLine: "Land"})
	}
	ids := effectstest.ZnrStackLibrary(me, cards...)
	effectstest.FraBCastCopy(t, g, me, "Bind to Life")
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a lone candidate was asked about")
	}
	if !g.Battlefield.Contains(ids[0]) {
		t.Error("the lone milled creature did not come back")
	}
}

// --- Vibrant Outburst -------------------------------------------------------

func TestSocB38OutburstDealsThreeAndTapsUpToOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", socB38BearTypeLine, 2, 2)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Vibrant Outburst", "Instant", socB38OutburstOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: blocker}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-3)
	}
	if c, ok := g.LookupCardForEffect(blocker); !ok || !c.Tapped {
		t.Error("the creature is not tapped")
	}
}

func TestSocB38OutburstTapsTheSameCreatureItDamagesAndMayChooseNoCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big", socB38BearTypeLine, 4, 4)
	effectstest.CastCatalogSpell(t, g, "Vibrant Outburst", "Instant", socB38OutburstOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}, {Kind: game.TargetCard, ID: big}})
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(big); !ok || !c.Tapped {
		t.Error("the creature it damaged was not tapped")
	}
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Vibrant Outburst", "Instant", socB38OutburstOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("with no creature chosen the player is at %d, want %d", opp.Life, life-3)
	}
}
