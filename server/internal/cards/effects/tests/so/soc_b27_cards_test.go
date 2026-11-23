package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b27_cards_test.go — Secrets of Strixhaven, batch 27 (#4456). Each test
// plays the real catalog entry through the engine and asserts what the
// printed text says.

const (
	socB27HomesicknessOracle = "ee57fba8-e68b-4c52-8de7-9b9873d84506"
	socB27PageOracle         = "66ee56b9-aca8-4efa-b9b5-50717e0784a7"
	socB27GraffalonOracle    = "9494c499-6ead-491f-aeb0-6dbfac4d960c"
	socB27ChannelerOracle    = "95630a83-3699-4340-b3e5-17e81db1b82f"
	socB27InkmageOracle      = "fcd31653-4964-46a6-820c-f17e8cbaa1fa"
	socB27HealerOracle       = "4cbbdebc-b807-4d7b-92dd-312637d2a5de"
	socB27MascotOracle       = "0f0959ff-4464-4738-9d79-5b06f368f144"
	socB27DemonstratorOracle = "4993651f-f161-4007-b485-a0c66a02f91b"
	socB27InterjectionOracle = "47bbadcf-11c6-4a4d-9575-1eb8284a174c"
	socB27BearType           = "Creature — Bear"
)

func TestSocB27EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB27HomesicknessOracle:  "Homesickness",
		socB27PageOracle:          "Honorbound Page",
		socB27PageOracle + "#1":   "Forum's Favor",
		socB27GraffalonOracle:     "Hungry Graffalon",
		socB27ChannelerOracle:     "Hydro-Channeler",
		socB27InkmageOracle:       "Imperious Inkmage",
		socB27HealerOracle:        "Infirmary Healer",
		socB27HealerOracle + "#1": "Stream of Life",
		socB27MascotOracle:        "Inkling Mascot",
		socB27DemonstratorOracle:  "Inkshape Demonstrator",
		socB27InterjectionOracle:  "Interjection",
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

// --- Homesickness -------------------------------------------------------------

func socB27HomesicknessTargets(player uuid.UUID, creatures ...uuid.UUID) []game.TargetRef {
	out := []game.TargetRef{{Kind: game.TargetPlayer, ID: player, Slot: 0}}
	for _, id := range creatures {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id, Slot: 1})
	}
	return out
}

func TestSocB27HomesicknessDrawsTwoThenTapsAndStunsTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", socB27BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", socB27BearType, 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", socB27BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Homesickness", "Instant", socB27HomesicknessOracle, socB27HomesicknessTargets(me.ID, mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	for _, id := range []uuid.UUID{mine, theirs} {
		if !effectstest.TappedOnBattlefield(t, g, id) {
			t.Error("a target creature is untapped")
		}
		if n := effectstest.CountersOn(g, id, game.CounterStun); n != 1 {
			t.Errorf("a target creature has %d stun counters, want 1", n)
		}
	}
	if effectstest.TappedOnBattlefield(t, g, other) || effectstest.CountersOn(g, other, game.CounterStun) != 0 {
		t.Error("the untargeted creature was tapped or stunned")
	}
}

func TestSocB27HomesicknessTargetsAnotherPlayerToDrawAndAcceptsNoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", socB27BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Homesickness", "Instant", socB27HomesicknessOracle, socB27HomesicknessTargets(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != mine || opp.Hand.Size()-theirs != 2 {
		t.Errorf("hands %d/%d -> %d/%d, want the target player to draw two", mine, theirs, me.Hand.Size(), opp.Hand.Size())
	}
	if effectstest.TappedOnBattlefield(t, g, bear) || effectstest.CountersOn(g, bear, game.CounterStun) != 0 {
		t.Error("a creature that was not targeted was tapped or stunned")
	}
}

func TestSocB27HomesicknessStunsAnAlreadyTappedCreatureAndRefusesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", socB27BearType, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", socB27BearType, 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", socB27BearType, 2, 2)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, a).Tapped = true })
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Homesickness", "Instant", socB27HomesicknessOracle, socB27HomesicknessTargets(me.ID, a, b, c)); err == nil {
		t.Fatal("three creatures were accepted for \"up to two\"")
	}
	effectstest.CastCatalogSpell(t, g, "Homesickness", "Instant", socB27HomesicknessOracle, socB27HomesicknessTargets(me.ID, a))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, a, game.CounterStun); n != 1 {
		t.Errorf("an already tapped target has %d stun counters, want 1", n)
	}
}

// --- Honorbound Page // Forum's Favor --------------------------------------------

func TestSocB27PageEntersPreparedWithFirstStrikeAndForumsFavorPumpsAndGrantsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	c := sos1Prep(me.ID, socB27PageOracle, []string{"first strike"}, "Honorbound Page", "Creature — Cat Cleric", "{3}{W}", 3, 3,
		"Forum's Favor", "Sorcery", "{W}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Honorbound Page did not enter prepared")
	}
	if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "first strike") {
		t.Error("Honorbound Page has no first strike")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB27BearType, 2, 2)
	effectstest.FraBCastCopy(t, g, me, "Forum's Favor", game.TargetRef{Kind: game.TargetCard, ID: bear})
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("target P/T = %d/%d, want 3/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target has no flying")
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
	effectstest.EndTurn(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Errorf("the effect outlasted the turn: power %d", p)
	}
}

// --- Hungry Graffalon ------------------------------------------------------------

func TestSocB27GraffalonIsReachAndIncrementsOnlyPastItsPowerOrToughness(t *testing.T) {
	spec, ok := effects.Lookup(socB27GraffalonOracle)
	if !ok {
		t.Fatal("Hungry Graffalon is not registered")
	}
	want := map[string]bool{"reach": false, effects.KeywordIncrement: false}
	for _, k := range spec.PrintedKeywords {
		if _, ok := want[k]; ok {
			want[k] = true
		}
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("missing printed keyword %q", k)
		}
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	graffalon := effectstest.B12Push(g, me.ID, "Hungry Graffalon", "Creature — Giraffe", socB27GraffalonOracle, 3, 4)
	effectstest.AdvanceToMain(t, g)
	cast := func(cost string, n int) {
		t.Helper()
		if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: cost}, "G", n, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", cost, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	cast("{2}{G}", 3) // 3 is not greater than 3/4
	if n := effectstest.CountersOn(g, graffalon, game.CounterPlusOne); n != 0 {
		t.Fatalf("counters %d after a three-mana spell, want 0", n)
	}
	cast("{3}{G}", 4) // 4 > power 3
	if n := effectstest.CountersOn(g, graffalon, game.CounterPlusOne); n != 1 {
		t.Fatalf("counters %d after a four-mana spell, want 1", n)
	}
	cast("{3}{G}", 4) // 4 is not greater than 4/5
	if n := effectstest.CountersOn(g, graffalon, game.CounterPlusOne); n != 1 {
		t.Fatalf("counters %d after a second four-mana spell, want still 1", n)
	}
	cast("{4}{G}", 5) // 5 > 4
	if n := effectstest.CountersOn(g, graffalon, game.CounterPlusOne); n != 2 {
		t.Fatalf("counters %d after a five-mana spell, want 2", n)
	}
}

// --- Hydro-Channeler -------------------------------------------------------------

func TestSocB27ChannelerBlueManaPaysOnlyForInstantsAndSorceries(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		ok             bool
	}{
		{"instant", "Instant", true},
		{"sorcery", "Sorcery", true},
		{"creature", "Creature — Bear", false},
		{"artifact", "Artifact", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := effectstest.SpendTable(t)
			channeler := effectstest.PushCatalogPermanent(g, me.ID, "Hydro-Channeler", "Creature — Merfolk Wizard", socB27ChannelerOracle, false)
			if err := g.ActivateManaAbility(me.ID, channeler, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatalf("tap for {U}: %v", err)
			}
			if units := me.ManaPool.Units(); len(units) != 1 || units[0].Color != "U" {
				t.Fatalf("pool = %+v, want one U", me.ManaPool)
			}
			spell := effectstest.HandSpell(me, "Test "+tc.name, tc.typeLine, "{U}")
			err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true})
			if tc.ok && err != nil {
				t.Errorf("the Channeler's mana could not pay for a %s: %v", tc.name, err)
			}
			if !tc.ok {
				effectstest.RefusedForMana(t, err, "the Channeler's mana for a "+tc.name)
			}
		})
	}
}

func TestSocB27ChannelerSecondAbilityPaysOneForAnyColorOfRestrictedMana(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	channeler := effectstest.PushCatalogPermanent(g, me.ID, "Hydro-Channeler", "Creature — Merfolk Wizard", socB27ChannelerOracle, false)
	// Without {1} to pay, the ability is refused and the Channeler stays untapped.
	if err := g.ActivateManaAbility(me.ID, channeler, 1, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Fatal("the {1}, {T} ability was activated with nothing to pay")
	}
	if c, _ := effectstest.BattlefieldCard(g, channeler); c.Tapped {
		t.Fatal("a refused activation tapped the Channeler")
	}
	effectstest.FillPool(me, 1)
	if err := g.ActivateManaAbility(me.ID, channeler, 1, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("{1}, {T}: add one mana of any color: %v", err)
	}
	if units := me.ManaPool.Units(); len(units) != 1 || units[0].Color != "R" {
		t.Fatalf("pool = %+v, want the {1} spent and one R added", me.ManaPool)
	}
	if c, _ := effectstest.BattlefieldCard(g, channeler); !c.Tapped {
		t.Error("the Channeler did not tap")
	}
	creature := effectstest.HandSpell(me, "Test creature", "Creature — Bear", "{R}")
	effectstest.RefusedForMana(t, g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}), "the red mana for a creature")
	instant := effectstest.HandSpell(me, "Test instant", "Instant", "{R}")
	if err := g.CastSpell(me.ID, instant, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the red mana could not pay for an instant: %v", err)
	}
}

func TestSocB27ChannelerIsSubjectToSummoningSickness(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	channeler := effectstest.PushCatalogPermanent(g, me.ID, "Hydro-Channeler", "Creature — Merfolk Wizard", socB27ChannelerOracle, true)
	if err := g.ActivateManaAbility(me.ID, channeler, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a summoning-sick Channeler tapped for mana")
	}
}

// --- Imperious Inkmage -----------------------------------------------------------

func TestSocB27InkmageHasVigilanceAndSurveilsTwoWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 6)
	effectstest.AdvanceToMain(t, g)
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	mage := game.Card{
		InstanceID: uuid.New(), Name: "Imperious Inkmage", TypeLine: "Creature — Orc Warlock", OracleID: socB27InkmageOracle,
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	}
	effectstest.CastFromHandAndResolve(t, g, me, mage)
	if !effectstest.HasEffectiveKeyword(t, g, mage.InstanceID, "vigilance") {
		t.Error("Imperious Inkmage has no vigilance")
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.SurveilChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no surveil prompt")
	}
	if len(prompt.ScryCards) != 2 {
		t.Errorf("surveil looked at %d cards, want 2", len(prompt.ScryCards))
	}
	var kept []uuid.UUID
	for _, id := range prompt.ScryCards {
		if id != top {
			kept = append(kept, id)
		}
	}
	if err := g.ResolveSurveil(prompt.ID, me.ID, []uuid.UUID{top}, kept); err != nil {
		t.Fatalf("surveil away the top card: %v", err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not reach the graveyard")
	}
}

// --- Infirmary Healer // Stream of Life ----------------------------------------------

func TestSocB27HealerEntersPreparedAndStreamOfLifeGainsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	c := sos1Prep(me.ID, socB27HealerOracle, nil, "Infirmary Healer", "Creature — Cat Cleric", "{1}{G}", 2, 3,
		"Stream of Life", "Sorcery", "{X}{G}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Infirmary Healer did not enter prepared")
	}
	life, theirs := opp.Life, me.Life
	cp, ok := effectstest.PrepareCopyOf(g, "Stream of Life")
	if !ok {
		t.Fatal("no Stream of Life copy in exile")
	}
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{
		FromZone: string(game.ZoneExile), XValue: 4, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast Stream of Life: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+4 {
		t.Errorf("the target player's life %d -> %d, want +4", life, opp.Life)
	}
	if me.Life != theirs {
		t.Errorf("the caster's life changed %d -> %d", theirs, me.Life)
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
}

func TestSocB27StreamOfLifeAtXZeroGainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	c := sos1Prep(me.ID, socB27HealerOracle, nil, "Infirmary Healer", "Creature — Cat Cleric", "{1}{G}", 2, 3,
		"Stream of Life", "Sorcery", "{X}{G}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	life := me.Life
	cp, _ := effectstest.PrepareCopyOf(g, "Stream of Life")
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{
		FromZone: string(game.ZoneExile), XValue: 0, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("cast Stream of Life for X = 0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d -> %d, want no change at X = 0", life, me.Life)
	}
}

// --- Inkling Mascot ----------------------------------------------------------------

func TestSocB27MascotGainsFlyingAndSurveilsWhenYouTargetACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 6)
	mascot := effectstest.B12Push(g, me.ID, "Inkling Mascot", "Creature — Inkling Cat", socB27MascotOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB27BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	if effectstest.HasEffectiveKeyword(t, g, mascot, "flying") {
		t.Fatal("the Mascot starts with flying")
	}
	effectstest.CastCatalogSpell(t, g, "Interjection", "Instant", socB27InterjectionOracle, effectstest.CardRefs(bear))
	if n := effectstest.TriggersOnStackFrom(g, mascot); n != 1 {
		t.Fatalf("Repartee triggers %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, mascot, "flying") {
		t.Error("the Mascot did not gain flying")
	}
	prompt := effectstest.SurveilChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no surveil prompt")
	}
	if len(prompt.ScryCards) != 1 {
		t.Errorf("surveil looked at %d cards, want 1", len(prompt.ScryCards))
	}
	if err := g.ResolveSurveil(prompt.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("surveil away the top card: %v", err)
	}
	effectstest.EndTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, mascot, "flying") {
		t.Error("flying lasted past the turn")
	}
}

func TestSocB27MascotIgnoresASpellThatTargetsNoCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mascot := effectstest.B12Push(g, me.ID, "Inkling Mascot", "Creature — Inkling Cat", socB27MascotOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Stream of Life", "Sorcery", socB27HealerOracle+"#1", []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if n := effectstest.TriggersOnStackFrom(g, mascot); n != 0 {
		t.Errorf("Repartee triggered %d times for a spell that targets only a player", n)
	}
	if effectstest.HasEffectiveKeyword(t, g, mascot, "flying") {
		t.Error("the Mascot has flying without a qualifying cast")
	}
}

// --- Inkshape Demonstrator ----------------------------------------------------------

func TestSocB27DemonstratorHasWardAndRepartee(t *testing.T) {
	spec, ok := effects.Lookup(socB27DemonstratorOracle)
	if !ok {
		t.Fatal("Inkshape Demonstrator is not registered")
	}
	if len(spec.Triggered) != 2 {
		t.Fatalf("%d triggered abilities, want ward and Repartee", len(spec.Triggered))
	}
}

func TestSocB27DemonstratorReparteeGivesPlusOneAndLifelinkUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	demo := effectstest.B12Push(g, me.ID, "Inkshape Demonstrator", "Creature — Elephant Cleric", socB27DemonstratorOracle, 3, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB27BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Interjection", "Instant", socB27InterjectionOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, demo); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, demo, "lifelink") {
		t.Error("the Demonstrator did not gain lifelink")
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("Interjection's target power %d, want 4", p)
	}
	effectstest.EndTurn(t, g)
	if p := effectstest.EffectivePower(t, g, demo); p != 3 || effectstest.HasEffectiveKeyword(t, g, demo, "lifelink") {
		t.Errorf("the effect outlasted the turn: power %d", p)
	}
}

func TestSocB27DemonstratorReparteeIgnoresASpellWithNoCreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	demo := effectstest.B12Push(g, me.ID, "Inkshape Demonstrator", "Creature — Elephant Cleric", socB27DemonstratorOracle, 3, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Stream of Life", "Sorcery", socB27HealerOracle+"#1", []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if n := effectstest.TriggersOnStackFrom(g, demo); n != 0 {
		t.Errorf("Repartee triggered %d times for a spell that targets only a player", n)
	}
}

// --- Interjection -------------------------------------------------------------------

func TestSocB27InterjectionGivesPlusTwoPlusTwoAndFirstStrikeUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB27BearType, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Interjection", "Instant", socB27InterjectionOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("P/T = %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the target has no first strike")
	}
	effectstest.EndTurn(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 || effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Errorf("the effect outlasted the turn: power %d", p)
	}
}

func TestSocB27InterjectionRefusesANoncreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Interjection", "Instant", socB27InterjectionOracle, effectstest.CardRefs(rock)); err == nil {
		t.Error("a noncreature permanent was accepted as the target")
	}
}
