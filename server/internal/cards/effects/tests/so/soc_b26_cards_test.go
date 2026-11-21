package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b26_cards_test.go — Secrets of Strixhaven (sos), batch 26 (#4456).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	socB26ExcavatorOracle    = "5d808e64-a648-4f1b-a3a9-fd09568db281"
	socB26ArthropodOracle    = "7f577b46-c69d-4c75-9db5-7e10206b6b59"
	socB26DecayOracle        = "ac4277d7-d78a-4d6e-8d6d-11f9e56f8ae7"
	socB26GlasswrightOracle  = "23c3de07-8aa2-48e9-848c-50b5a40bd133"
	socB26GraduationOracle   = "40e0665d-57b8-4c19-9fb8-b3155d7bbb94"
	socB26GrappleOracle      = "dced5cab-1438-4264-b44c-54dc37ec73d3"
	socB26CurveOracle        = "1efd6082-15a8-4322-892d-02d01e381e2a"
	socB26AcademicOracle     = "921252df-6679-4f70-ae26-d4956c90b7df"
	socB26AnnotationOracle   = "1657be64-b38f-4074-a743-bf558dc57901"
	socB26ArgumentOracle     = "329f6434-8f1b-4a10-8feb-2a9dc5f2712b"
	socB26DaydreamFlashback  = "flashback"
	socB26BearTypeLine       = "Creature — Bear"
	socB26ArtifactTypeLine   = "Artifact"
	socB26InstantTypeLineFor = "Instant"
)

func TestSocB26EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB26ExcavatorOracle:          "Garrison Excavator",
		socB26ArthropodOracle:          "Geometer's Arthropod",
		socB26DecayOracle:              "Glorious Decay",
		socB26GlasswrightOracle:        "Goblin Glasswright",
		socB26GlasswrightOracle + "#1": "Craft with Pride",
		socB26GraduationOracle:         "Graduation Day",
		socB26GrappleOracle:            "Grapple with Death",
		socB26CurveOracle:              "Growth Curve",
		socB26AcademicOracle:           "Hardened Academic",
		socB26AnnotationOracle:         "Harsh Annotation",
		socB26ArgumentOracle:           "Heated Argument",
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

// socB26Tokens is every token named `name` that `owner` controls.
func socB26Tokens(g *game.Game, owner uuid.UUID, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == owner && c.IsToken() {
				out = append(out, c)
			}
		}
	})
	return out
}

// socB26FlashbackDaydream puts Daydream in the graveyard and casts it back
// from there on `bear`, which takes a card out of the graveyard by a cast.
func socB26FlashbackDaydream(t *testing.T, g *game.Game, me *game.Player, bear uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Daydream", TypeLine: "Sorcery", ManaCost: "{W}",
		OracleID: socB22DaydreamOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPoolColored(me, "W", 3)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: string(game.ZoneGraveyard),
		AlternativeCost: socB26DaydreamFlashback, Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("flash Daydream back: %v", err)
	}
	return id
}

// socB26SettleAnsweringPick lets the table settle, answering a target prompt
// for `chooser` with `id` when a trigger asks for one.
func socB26SettleAnsweringPick(t *testing.T, g *game.Game, chooser, id uuid.UUID) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, chooser) != nil {
		effectstest.PickCard(t, g, chooser, id)
		effectstest.PassPriorityAroundTable(t, g)
	}
}

// --- Garrison Excavator ------------------------------------------------------

func TestSocB26ExcavatorHasMenace(t *testing.T) {
	spec, ok := effects.Lookup(socB26ExcavatorOracle)
	if !ok || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "menace" {
		t.Fatalf("printed keywords %v, want [menace]", spec.PrintedKeywords)
	}
}

func TestSocB26ExcavatorMakesASpiritWhenACardLeavesYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Garrison Excavator", "Creature — Orc Sorcerer", socB26ExcavatorOracle, 3, 4)
	dead := effectstest.SocB11Graveyard(me, "Dead Bear", socB26BearTypeLine, "{1}{G}")
	socB11Move(t, g, me.ID, dead, game.ZoneGraveyard, game.ZoneExile)
	effectstest.PassPriorityAroundTable(t, g)
	spirits := socB26Tokens(g, me.ID, "Spirit")
	if len(spirits) != 1 {
		t.Fatalf("%d Spirits, want 1", len(spirits))
	}
	s := spirits[0]
	if s.CurrentPower() != 2 || s.CurrentToughness() != 2 || len(s.Colors) != 2 || !s.HasSubtypeWord("Spirit") {
		t.Errorf("token %s %d/%d colors %v, want a 2/2 red and white Spirit", s.Name, s.CurrentPower(), s.CurrentToughness(), s.Colors)
	}
	// A card leaving an opponent's graveyard does nothing.
	theirs := effectstest.SocB11Graveyard(opp, "Their Bear", socB26BearTypeLine, "{1}{G}")
	socB11Move(t, g, opp.ID, theirs, game.ZoneGraveyard, game.ZoneExile)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB26Tokens(g, me.ID, "Spirit")); n != 1 {
		t.Errorf("an opposing graveyard card made a Spirit: %d", n)
	}
	// A card moving into a graveyard is not leaving one.
	socB11Move(t, g, me.ID, dead2(me), game.ZoneHand, game.ZoneGraveyard)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB26Tokens(g, me.ID, "Spirit")); n != 1 {
		t.Errorf("a card entering the graveyard made a Spirit: %d", n)
	}
}

// dead2 is a plain card in the hand to move into the graveyard.
func dead2(p *game.Player) uuid.UUID {
	return effectstest.PushCatalogHandCard(p, "Discard Fodder", "Sorcery", "")
}

func TestSocB26ExcavatorCountsACastFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Garrison Excavator", "Creature — Orc Sorcerer", socB26ExcavatorOracle, 3, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB26BearTypeLine, 2, 2)
	socB26FlashbackDaydream(t, g, me, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB26Tokens(g, me.ID, "Spirit")); n != 1 {
		t.Errorf("%d Spirits after a flashback cast out of the graveyard, want 1", n)
	}
}

// --- Geometer's Arthropod ----------------------------------------------------

// socB26ArthropodLibrary replaces the library with five named cards, "A" on top.
func socB26ArthropodLibrary(me *game.Player) []uuid.UUID {
	me.Library.Cards = nil
	var cards []game.Card
	for _, n := range []string{"A", "B", "C", "D", "E"} {
		cards = append(cards, game.Card{Name: n, TypeLine: "Sorcery"})
	}
	return effectstest.ZnrStackLibrary(me, cards...)
}

func socB26CastFireball(g *game.Game, me *game.Player, x int) error {
	return effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Fireball", TypeLine: "Sorcery", ManaCost: "{X}{R}"}, "R", x+1,
		game.CastSpellParams{XValue: x})
}

func TestSocB26ArthropodLooksAtTopXPutsOneInHandAndTheRestOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arthropod := effectstest.B12Push(g, me.ID, "Geometer's Arthropod", "Creature — Fractal Crab", socB26ArthropodOracle, 1, 4)
	ids := socB26ArthropodLibrary(me)
	effectstest.AdvanceToMain(t, g)
	if err := socB26CastFireball(g, me, 3); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if n := effectstest.TriggersOnStackFrom(g, arthropod); n != 1 {
		t.Fatalf("triggers %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt")
	}
	if len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards bounds %d..%d, want the top 3, exactly one", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	for _, id := range pick.ChooseCards {
		if id == ids[3] || id == ids[4] {
			t.Errorf("a card below the top three was offered")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err == nil {
		t.Fatal("taking nothing must be refused: the card is not a may")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the chosen card is not in hand")
	}
	n := len(me.Library.Cards)
	// The top of a library is the last element: D on top, E under it, then
	// the two cards not taken at the bottom.
	if n != 4 || me.Library.Cards[3].InstanceID != ids[3] || me.Library.Cards[2].InstanceID != ids[4] {
		t.Fatalf("library of %d does not have D on top of E", n)
	}
	rest := map[uuid.UUID]bool{me.Library.Cards[0].InstanceID: true, me.Library.Cards[1].InstanceID: true}
	if !rest[ids[0]] || !rest[ids[2]] {
		t.Error("the two cards not taken are not on the bottom of the library")
	}
}

func TestSocB26ArthropodIgnoresASpellWithoutXInItsManaCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	arthropod := effectstest.B12Push(g, me.ID, "Geometer's Arthropod", "Creature — Fractal Crab", socB26ArthropodOracle, 1, 4)
	socB26ArthropodLibrary(me)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"}, "R", 1,
		game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if n := effectstest.TriggersOnStackFrom(g, arthropod); n != 0 {
		t.Errorf("a spell with no {X} triggered the Arthropod %d times", n)
	}
}

func TestSocB26ArthropodWithXOfZeroLooksAtNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Geometer's Arthropod", "Creature — Fractal Crab", socB26ArthropodOracle, 1, 4)
	ids := socB26ArthropodLibrary(me)
	effectstest.AdvanceToMain(t, g)
	if err := socB26CastFireball(g, me, 0); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.ChooseCardsChoiceFor(g, me.ID); p != nil {
		t.Fatalf("X=0 asked about %d cards", len(p.ChooseCards))
	}
	if me.Library.Cards[4].InstanceID != ids[0] || len(me.Library.Cards) != 5 {
		t.Error("the library changed with X=0")
	}
}

func TestSocB26ArthropodLooksAtAShortLibraryWhole(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Geometer's Arthropod", "Creature — Fractal Crab", socB26ArthropodOracle, 1, 4)
	ids := socB26ArthropodLibrary(me)
	effectstest.AdvanceToMain(t, g)
	if err := socB26CastFireball(g, me, 9); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || len(pick.ChooseCards) != 5 {
		t.Fatalf("prompt %+v, want all five cards offered", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[4])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[4]) || len(me.Library.Cards) != 4 {
		t.Error("the fifth card did not reach the hand")
	}
}

// --- Glorious Decay ----------------------------------------------------------

func TestSocB26DecayDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	rock := effectstest.B12Permanent(g, g.Seats[1].ID, "Rock", socB26ArtifactTypeLine)
	if err := effectstest.RfCastModal(t, g, "Glorious Decay", socB26DecayOracle, []int{0}, effectstest.RfCardTarget(rock)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestSocB26DecayDealsFourToAFlyerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ground := effectstest.B12Creature(g, opp.ID, "Ground", socB26BearTypeLine, 2, 6)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird",
		Power: 2, Toughness: 6, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	if err := effectstest.RfCastModal(t, g, "Glorious Decay", socB26DecayOracle, []int{1}, effectstest.RfCardTarget(ground)); err == nil {
		t.Error("a creature without flying is not a legal target of the damage mode")
	}
	if err := effectstest.RfCastModal(t, g, "Glorious Decay", socB26DecayOracle, []int{1}, effectstest.RfCardTarget(flyer)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, flyer); got != 4 {
		t.Errorf("%d damage marked, want 4", got)
	}
}

func TestSocB26DecayExilesAGraveyardCardAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := effectstest.SocB11Graveyard(opp, "Their Bear", socB26BearTypeLine, "{1}{G}")
	hand := me.Hand.Size()
	if err := effectstest.RfCastModal(t, g, "Glorious Decay", socB26DecayOracle, []int{2}, effectstest.RfCardTarget(dead)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZnrZoneOf(g, dead); z != game.ZoneExile {
		t.Errorf("the card is in %v, want exile", z)
	}
	// The helper puts the spell in hand and the cast takes it out again, so
	// the hand ends one card bigger by the draw.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (one card drawn)", got, hand+1)
	}
}

// --- Goblin Glasswright // Craft with Pride ----------------------------------

func TestSocB26GlasswrightEntersPreparedAndCraftWithPrideMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB26GlasswrightOracle, nil, "Goblin Glasswright", "Creature — Goblin Sorcerer", "{1}{R}", 2, 2,
		"Craft with Pride", "Sorcery", "{R}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Goblin Glasswright did not enter prepared")
	}
	if n := len(socB26Tokens(g, me.ID, "Treasure")); n != 0 {
		t.Fatalf("%d Treasures before the copy was cast", n)
	}
	effectstest.FraBCastCopy(t, g, me, "Craft with Pride")
	if n := len(socB26Tokens(g, me.ID, "Treasure")); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
}

// --- Graduation Day ----------------------------------------------------------

func TestSocB26GraduationDayPutsACounterOnYourCreatureWhenYouTargetACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Graduation Day", TypeLine: "Enchantment",
		OracleID: socB26GraduationOracle, Owner: me.ID, Controller: me.ID})
	mine := effectstest.B12Creature(g, me.ID, "Mine", socB26BearTypeLine, 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", socB26BearTypeLine, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Ajani's Response", socB22InstantTypeLine, socB19ResponseOracle, effectstest.CardRefs(theirs))
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, mine)
	}
	if n := effectstest.TriggersOnStackFrom(g, id); n != 1 {
		t.Fatalf("Repartee triggers %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, mine, game.CounterPlusOne); n != 1 {
		t.Errorf("your creature has %d counters, want 1", n)
	}
	if n := effectstest.CountersOn(g, theirs, game.CounterPlusOne); n != 0 {
		t.Errorf("their creature has %d counters, want 0", n)
	}
}

func TestSocB26GraduationDayCanPutTheCounterOnAnotherCreatureOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Graduation Day", TypeLine: "Enchantment",
		OracleID: socB26GraduationOracle, Owner: me.ID, Controller: me.ID})
	a := effectstest.B12Creature(g, me.ID, "A", socB26BearTypeLine, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", socB26BearTypeLine, 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", socB26BearTypeLine, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Ajani's Response", socB22InstantTypeLine, socB19ResponseOracle, effectstest.CardRefs(theirs))
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, b, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, a, game.CounterPlusOne) != 0 {
		t.Errorf("counters A=%d B=%d, want 0 and 1", effectstest.CountersOn(g, a, game.CounterPlusOne), effectstest.CountersOn(g, b, game.CounterPlusOne))
	}
}

func TestSocB26GraduationDayIgnoresASpellThatTargetsNoCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Graduation Day", TypeLine: "Enchantment",
		OracleID: socB26GraduationOracle, Owner: me.ID, Controller: me.ID})
	effectstest.B12Creature(g, me.ID, "Mine", socB26BearTypeLine, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Sign in Blood", "Sorcery", "c6207f6a-a624-4754-88f5-dbe700c841ff",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if n := effectstest.TriggersOnStackFrom(g, id); n != 0 {
		t.Errorf("Repartee triggered %d times for a spell that targets only a player", n)
	}
}

func TestSocB26GraduationDayDoesNothingWithoutACreatureOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Graduation Day", TypeLine: "Enchantment",
		OracleID: socB26GraduationOracle, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", socB26BearTypeLine, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Ajani's Response", socB22InstantTypeLine, socB19ResponseOracle, effectstest.CardRefs(theirs))
	if n := effectstest.TriggersOnStackFrom(g, id); n != 0 {
		t.Errorf("the trigger went on the stack with no creature of yours to target (%d)", n)
	}
}

// --- Grapple with Death ------------------------------------------------------

func TestSocB26GrappleDestroysAnArtifactOrCreatureAndGainsOneLife(t *testing.T) {
	for name, build := range map[string]func(g *game.Game, owner uuid.UUID) uuid.UUID{
		"creature": func(g *game.Game, owner uuid.UUID) uuid.UUID {
			return effectstest.B12Creature(g, owner, "Victim", socB26BearTypeLine, 2, 2)
		},
		"artifact": func(g *game.Game, owner uuid.UUID) uuid.UUID {
			return effectstest.B12Permanent(g, owner, "Rock", socB26ArtifactTypeLine)
		},
	} {
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			victim := build(g, g.Seats[1].ID)
			life := me.Life
			effectstest.CastCatalogSpell(t, g, "Grapple with Death", "Sorcery", socB26GrappleOracle, effectstest.CardRefs(victim))
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(victim) {
				t.Error("the target survived")
			}
			if me.Life != life+1 {
				t.Errorf("life %d, want %d", me.Life, life+1)
			}
		})
	}
}

func TestSocB26GrappleGainsTheLifeEvenWhenTheTargetIsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tough := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Tough", TypeLine: "Creature — Golem",
		Power: 2, Toughness: 2, Keywords: []string{"indestructible"}, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Grapple with Death", "Sorcery", socB26GrappleOracle, effectstest.CardRefs(tough))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tough) {
		t.Error("an indestructible creature was destroyed")
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d: the gain is its own sentence", me.Life, life+1)
	}
}

func TestSocB26GrappleRefusesALandAndAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.B12Permanent(g, g.Seats[1].ID, "Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErr(t, g, "Grapple with Death", "Sorcery", socB26GrappleOracle, effectstest.CardRefs(land)); err == nil {
		t.Error("a land is not an artifact or creature")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Grapple with Death", "Sorcery", socB26GrappleOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}); err == nil {
		t.Error("a player is not a legal target")
	}
}

// --- Growth Curve ------------------------------------------------------------

func TestSocB26GrowthCurveOnABareCreatureMakesTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB26BearTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Growth Curve", "Sorcery", socB26CurveOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters, want 2 (one, then doubled)", n)
	}
}

func TestSocB26GrowthCurveWithThreeCountersEndsOnEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 3})
	effectstest.CastCatalogSpell(t, g, "Growth Curve", "Sorcery", socB26CurveOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 8 {
		t.Errorf("%d counters, want 8 ((3+1) doubled)", n)
	}
}

func TestSocB26GrowthCurveRefusesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", socB26BearTypeLine, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Growth Curve", "Sorcery", socB26CurveOracle, effectstest.CardRefs(theirs)); err == nil {
		t.Error("Growth Curve targeted an opponent's creature")
	}
}

// --- Hardened Academic -------------------------------------------------------

func TestSocB26AcademicHasFlyingAndHaste(t *testing.T) {
	spec, ok := effects.Lookup(socB26AcademicOracle)
	if !ok {
		t.Fatal("not registered")
	}
	want := map[string]bool{"flying": false, "haste": false}
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
}

func TestSocB26AcademicDiscardGivesLifelinkUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	academic := effectstest.B12Push(g, me.ID, "Hardened Academic", "Creature — Bird Cleric", socB26AcademicOracle, 2, 1)
	card := effectstest.PushCatalogHandCard(me, "Fodder", "Sorcery", "")
	if socB22HasKeyword(g, academic, "lifelink") {
		t.Fatal("lifelink before the ability")
	}
	effectstest.B16Activate(t, g, me.ID, academic, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{card}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(card) {
		t.Error("the discarded card is not in the graveyard")
	}
	if !socB22HasKeyword(g, academic, "lifelink") {
		t.Error("no lifelink after the discard")
	}
	effectstest.EndTurn(t, g)
	if socB22HasKeyword(g, academic, "lifelink") {
		t.Error("lifelink lasted past the end of the turn")
	}
}

func TestSocB26AcademicCannotActivateWithAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	academic := effectstest.B12Push(g, me.ID, "Hardened Academic", "Creature — Bird Cleric", socB26AcademicOracle, 2, 1)
	me.Hand.Cards = nil
	if err := g.ActivateCatalogAbility(me.ID, academic, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no card to discard")
	}
}

func TestSocB26AcademicPutsACounterOnItselfWhenACardLeavesYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	academic := effectstest.B12Push(g, me.ID, "Hardened Academic", "Creature — Bird Cleric", socB26AcademicOracle, 2, 1)
	dead := effectstest.SocB11Graveyard(me, "Dead Bear", socB26BearTypeLine, "{1}{G}")
	socB11Move(t, g, me.ID, dead, game.ZoneGraveyard, game.ZoneExile)
	socB26SettleAnsweringPick(t, g, me.ID, academic)
	if n := effectstest.CountersOn(g, academic, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d counters, want 1", n)
	}
	theirs := effectstest.SocB11Graveyard(opp, "Their Bear", socB26BearTypeLine, "{1}{G}")
	socB11Move(t, g, opp.ID, theirs, game.ZoneGraveyard, game.ZoneExile)
	socB26SettleAnsweringPick(t, g, me.ID, academic)
	if n := effectstest.CountersOn(g, academic, game.CounterPlusOne); n != 1 {
		t.Errorf("an opposing graveyard card put a counter: %d", n)
	}
}

func TestSocB26AcademicCounterCanGoOnAnotherCreatureAndNotAnOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	academic := effectstest.B12Push(g, me.ID, "Hardened Academic", "Creature — Bird Cleric", socB26AcademicOracle, 2, 1)
	other := effectstest.B12Creature(g, me.ID, "Other", socB26BearTypeLine, 2, 2)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", socB26BearTypeLine, 2, 2)
	dead := effectstest.SocB11Graveyard(me, "Dead Bear", socB26BearTypeLine, "{1}{G}")
	socB11Move(t, g, me.ID, dead, game.ZoneGraveyard, game.ZoneExile)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	for _, id := range p.PickTargetCards {
		if id == theirs {
			t.Error("an opponent's creature was offered as the target")
		}
	}
	effectstest.PickCard(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, other, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, academic, game.CounterPlusOne) != 0 {
		t.Errorf("counters other=%d academic=%d, want 1 and 0", effectstest.CountersOn(g, other, game.CounterPlusOne), effectstest.CountersOn(g, academic, game.CounterPlusOne))
	}
}

func TestSocB26AcademicCountsACastFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	academic := effectstest.B12Push(g, me.ID, "Hardened Academic", "Creature — Bird Cleric", socB26AcademicOracle, 2, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB26BearTypeLine, 2, 2)
	socB26FlashbackDaydream(t, g, me, bear)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, academic)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, academic, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after a flashback cast out of the graveyard, want 1", n)
	}
}

// --- Harsh Annotation --------------------------------------------------------

func TestSocB26AnnotationDestroysTheCreatureAndItsControllerGetsAnInkling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", socB26BearTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Harsh Annotation", socB26InstantTypeLineFor, socB26AnnotationOracle, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature survived")
	}
	inklings := socB26Tokens(g, opp.ID, "Inkling")
	if len(inklings) != 1 {
		t.Fatalf("the victim's controller has %d Inklings, want 1", len(inklings))
	}
	ink := inklings[0]
	if ink.CurrentPower() != 1 || ink.CurrentToughness() != 1 || !game.HasKeyword(&ink, "flying") || len(ink.Colors) != 2 {
		t.Errorf("Inkling %d/%d colors %v, want a 1/1 white and black flyer", ink.CurrentPower(), ink.CurrentToughness(), ink.Colors)
	}
	if n := len(socB26Tokens(g, me.ID, "Inkling")); n != 0 {
		t.Errorf("the caster got %d Inklings", n)
	}
}

func TestSocB26AnnotationGivesTheInklingToTheControllerNotTheOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen", TypeLine: socB26BearTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Harsh Annotation", socB26InstantTypeLineFor, socB26AnnotationOracle, effectstest.CardRefs(stolen))
	effectstest.PassPriorityAroundTable(t, g)
	if len(socB26Tokens(g, opp.ID, "Inkling")) != 1 || len(socB26Tokens(g, me.ID, "Inkling")) != 0 {
		t.Error("the Inkling did not go to the creature's controller")
	}
}

func TestSocB26AnnotationStillMakesTheInklingForAnIndestructibleCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	tough := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Tough", TypeLine: "Creature — Golem",
		Power: 2, Toughness: 2, Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Harsh Annotation", socB26InstantTypeLineFor, socB26AnnotationOracle, effectstest.CardRefs(tough))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tough) {
		t.Error("an indestructible creature was destroyed")
	}
	if n := len(socB26Tokens(g, opp.ID, "Inkling")); n != 1 {
		t.Errorf("%d Inklings, want 1: the token is a separate sentence", n)
	}
}

// --- Heated Argument ---------------------------------------------------------

func socB26CastArgument(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Heated Argument", socB26InstantTypeLineFor, socB26ArgumentOracle, effectstest.CardRefs(target))
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB26ArgumentDealsSixAndExilingACardDealsTwoToTheControllerToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Big", socB26BearTypeLine, 3, 8)
	fodder := effectstest.SocB11Graveyard(me, "Fodder", "Sorcery", "{1}")
	other := effectstest.SocB11Graveyard(me, "Other", "Sorcery", "{1}")
	life, mine := opp.Life, me.Life
	socB26CastArgument(t, g, victim)
	if got := effectstest.DamageMarkedOn(g, victim); got != 6 {
		t.Fatalf("%d damage marked, want 6", got)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no exile offer")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 || len(pick.ChooseCards) != 2 {
		t.Errorf("offer %d..%d over %d cards, want 0..1 over the two in your graveyard", pick.ChooseMin, pick.ChooseMax, len(pick.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZnrZoneOf(g, fodder); z != game.ZoneExile {
		t.Errorf("the chosen card is in %v, want exile", z)
	}
	if z := effectstest.ZnrZoneOf(g, other); z != game.ZoneGraveyard {
		t.Errorf("the other card is in %v, want the graveyard", z)
	}
	if opp.Life != life-2 {
		t.Errorf("the controller is at %d, want %d", opp.Life, life-2)
	}
	if me.Life != mine {
		t.Errorf("the caster's life changed: %d", me.Life)
	}
}

func TestSocB26ArgumentDecliningTheExileDealsNoDamageToTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Big", socB26BearTypeLine, 3, 8)
	fodder := effectstest.SocB11Graveyard(me, "Fodder", "Sorcery", "{1}")
	life := opp.Life
	socB26CastArgument(t, g, victim)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZnrZoneOf(g, fodder); z != game.ZoneGraveyard {
		t.Errorf("the card is in %v, want the graveyard", z)
	}
	if opp.Life != life {
		t.Errorf("the controller took damage with no exile: %d", opp.Life)
	}
}

func TestSocB26ArgumentWithAnEmptyGraveyardAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	me.Graveyard.Cards = nil
	victim := effectstest.B12Creature(g, opp.ID, "Big", socB26BearTypeLine, 3, 8)
	life := opp.Life
	socB26CastArgument(t, g, victim)
	if p := effectstest.ChooseCardsChoiceFor(g, me.ID); p != nil {
		t.Fatalf("asked about %d cards with an empty graveyard", len(p.ChooseCards))
	}
	if opp.Life != life {
		t.Errorf("the controller took damage: %d", opp.Life)
	}
}

func TestSocB26ArgumentHitsTheCreaturesControllerNotItsOwnerAndEvenWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen", TypeLine: socB26BearTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: third.ID})
	fodder := effectstest.SocB11Graveyard(me, "Fodder", "Sorcery", "{1}")
	oppLife, thirdLife := opp.Life, third.Life
	socB26CastArgument(t, g, stolen)
	effectstest.AnswerChooseCards(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(stolen) {
		t.Error("the 2/2 survived 6 damage")
	}
	if third.Life != thirdLife-2 {
		t.Errorf("the controller is at %d, want %d", third.Life, thirdLife-2)
	}
	if opp.Life != oppLife {
		t.Errorf("the owner took damage: %d", opp.Life)
	}
}

func TestSocB26ArgumentCanBeCastOnYourOwnCreatureAndHitsYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "Mine", socB26BearTypeLine, 2, 9)
	fodder := effectstest.SocB11Graveyard(me, "Fodder", "Sorcery", "{1}")
	life := me.Life
	socB26CastArgument(t, g, mine)
	effectstest.AnswerChooseCards(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d, want %d: you control the creature", me.Life, life-2)
	}
}
