package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b35_cards_test.go — Secrets of Strixhaven, batch 35 (#4456). Each test
// plays the real catalog entry through the engine and asserts what the
// printed text says.

const (
	socB35SnarlSongOracle    = "3b6fb13b-f126-4697-90a1-64b14ffcecd6"
	socB35ShadewriterOracle  = "21a094d6-de3d-48d4-a35d-b99617da55db"
	socB35SnoopingOracle     = "f07c07d7-00f1-44a5-b51f-0a22cc51fd2b"
	socB35SnubOracle         = "b284913e-3530-4a7e-8c75-efb18788dad0"
	socB35SummitOracle       = "33a4e73d-d93a-4b6f-88ff-cd53f20d178c"
	socB35SkywhaleOracle     = "6a6b19db-cdeb-4922-8642-5d872f90e7a1"
	socB35SeekerOracle       = "df5e4100-8c83-42d6-af96-1a1ec84a9210"
	socB35EnthusiastOracle   = "c77ddcac-ce54-4348-bde2-5d9caf3d5b04"
	socB35MascotOracle       = "34c572b2-dccf-43d9-a9c4-846653ad22da"
	socB35TidalmageOracle    = "1c45e872-3af8-47a4-9dc7-a99c293338f8"
	socB35InterjectionOracle = "47bbadcf-11c6-4a4d-9575-1eb8284a174c"
)

func TestSocB35EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB35SnarlSongOracle:         "Snarl Song",
		socB35ShadewriterOracle:       "Sneering Shadewriter",
		socB35SnoopingOracle:          "Snooping Page",
		socB35SnubOracle:              "Social Snub",
		socB35SummitOracle:            "Spectacle Summit",
		socB35SkywhaleOracle:          "Spectacular Skywhale",
		socB35SeekerOracle:            "Spellbook Seeker",
		socB35SeekerOracle + "#1":     "Careful Study",
		socB35EnthusiastOracle:        "Spiritcall Enthusiast",
		socB35EnthusiastOracle + "#1": "Scrollboost",
		socB35MascotOracle:            "Spirit Mascot",
		socB35TidalmageOracle:         "Stadium Tidalmage",
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

// --- Snarl Song -------------------------------------------------------------

// socB35CastSnarlSong casts the real spell from hand, paid with exactly `pool`.
func socB35CastSnarlSong(t *testing.T, g *game.Game, pool string) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.SocB17Hand(me, "Snarl Song", "Sorcery", "{5}{G}", socB35SnarlSongOracle)
	effectstest.FloatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Snarl Song: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB35SnarlSongMakesTwoFractalsWithOneCounterPerColorSpentAndGainsThatMuchLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	socB35CastSnarlSong(t, g, "GWUBRC")
	fractals := effectstest.SocB17Names(g, me.ID, "Fractal")
	if len(fractals) != 2 {
		t.Fatalf("%d Fractals, want 2", len(fractals))
	}
	for _, f := range fractals {
		if got := effectstest.CounterOn(g, f.InstanceID, game.CounterPlusOne); got != 5 {
			t.Errorf("a Fractal has %d +1/+1 counters, want 5 (five colors spent)", got)
		}
		if p := effectstest.S58p6View(t, g, f.InstanceID).CurrentPower(); p != 5 {
			t.Errorf("a Fractal's power is %d, want 5", p)
		}
	}
	if me.Life != life+5 {
		t.Errorf("life %d, want %d", me.Life, life+5)
	}
}

func TestSocB35SnarlSongCountsColorsNotMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	socB35CastSnarlSong(t, g, "GGGGGG")
	fractals := effectstest.SocB17Names(g, me.ID, "Fractal")
	if len(fractals) != 2 {
		t.Fatalf("%d Fractals, want 2", len(fractals))
	}
	for _, f := range fractals {
		if got := effectstest.CounterOn(g, f.InstanceID, game.CounterPlusOne); got != 1 {
			t.Errorf("a Fractal has %d +1/+1 counters, want 1 (green only)", got)
		}
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

func TestSocB35SnarlSongCannotBeCastWithoutItsGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.SocB17Hand(me, "Snarl Song", "Sorcery", "{5}{G}", socB35SnarlSongOracle)
	effectstest.FloatForTest(g, me, "WWWWWW")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("Snarl Song was cast without {G}")
	}
}

// --- Sneering Shadewriter ---------------------------------------------------

func TestSocB35ShadewriterFliesAndDrainsTwoFromEachOpponentWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	id := effectstest.Msc17Play(t, g, me, game.Card{
		Name: "Sneering Shadewriter", TypeLine: "Creature — Vampire Warlock", OracleID: socB35ShadewriterOracle,
		Power: 3, Toughness: 3,
	})
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Sneering Shadewriter has no flying")
	}
	for i, p := range g.Seats {
		want := lives[i] - 2
		if i == 0 {
			want = lives[i] + 2
		}
		if p.Life != want {
			t.Errorf("seat %d has %d life, want %d", i, p.Life, want)
		}
	}
}

// --- Snooping Page ----------------------------------------------------------

func TestSocB35SnoopingPageCantBeBlockedThisTurnAfterARepartee(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	page := effectstest.B12Push(g, me.ID, "Snooping Page", "Creature — Human Cleric", socB35SnoopingOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB27BearType, 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 8)
	effectstest.AdvanceToMain(t, g)
	if effectstest.RestrictionsOf(t, g, page)&game.CantBeBlocked != 0 {
		t.Fatal("the Page starts unblockable")
	}
	effectstest.CastCatalogSpell(t, g, "Interjection", "Instant", socB35InterjectionOracle, effectstest.CardRefs(bear))
	if n := effectstest.TriggersOnStackFrom(g, page); n != 1 {
		t.Fatalf("Repartee triggers %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, page)&game.CantBeBlocked == 0 {
		t.Fatal("the Page can still be blocked after a Repartee")
	}
	effectstest.BrAttack(t, g, page)
	effectstest.BrRefusal(t, g.DeclareBlocker(wall, page), game.BlockReasonCantBeBlocked)
}

func TestSocB35SnoopingPageLapsesAtEndOfTurnAndIgnoresASpellWithNoCreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	page := effectstest.B12Push(g, me.ID, "Snooping Page", "Creature — Human Cleric", socB35SnoopingOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Stream of Life", "Sorcery", "4cbbdebc-b807-4d7b-92dd-312637d2a5de#1", []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if n := effectstest.TriggersOnStackFrom(g, page); n != 0 {
		t.Errorf("Repartee triggered %d times for a spell that targets only a player", n)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB27BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Interjection", "Instant", socB35InterjectionOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, page)&game.CantBeBlocked == 0 {
		t.Fatal("the Page can be blocked after a Repartee")
	}
	effectstest.EndTurn(t, g)
	if effectstest.RestrictionsOf(t, g, page)&game.CantBeBlocked != 0 {
		t.Error("the Page was still unblockable on the next turn")
	}
}

func TestSocB35SnoopingPageDrawsAndLosesOneLifeOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	page := effectstest.B12Push(g, me.ID, "Snooping Page", "Creature — Human Cleric", socB35SnoopingOracle, 2, 3)
	effectstest.SeedLibrary(me, "One", "Two")
	hand, life := me.Hand.Size(), me.Life
	effectstest.DealCombatDamageToPlayer(g, page, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

// --- Social Snub ------------------------------------------------------------

// socB35SnubTable gives every seat two creatures and returns them by seat.
func socB35SnubTable(g *game.Game) [][]uuid.UUID {
	out := make([][]uuid.UUID, len(g.Seats))
	for i, p := range g.Seats {
		out[i] = []uuid.UUID{
			effectstest.B12Creature(g, p.ID, "Bear A", socB27BearType, 2, 2),
			effectstest.B12Creature(g, p.ID, "Bear B", socB27BearType, 2, 2),
		}
	}
	return out
}

// socB35SnubSacrificeRound answers every seat's sacrifice prompt with the
// first of its creatures still on the battlefield, in turn order.
func socB35SnubSacrificeRound(t *testing.T, g *game.Game, creatures [][]uuid.UUID) {
	t.Helper()
	active := g.Turn.ActiveSeat
	for offset := range g.Seats {
		seat := (active + offset) % len(g.Seats)
		p := g.Seats[seat]
		if effectstest.SacrificeChoiceFor(g, p.ID) == nil {
			t.Fatalf("seat %s was not asked to sacrifice a creature", p.Name)
		}
		for _, id := range creatures[seat] {
			if g.Battlefield.Contains(id) {
				effectstest.AnswerSacrifice(t, g, p.ID, id)
				break
			}
		}
	}
}

func TestSocB35SocialSnubCopiedWhileYouControlACreatureDoesEverythingTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creatures := socB35SnubTable(g)
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}

	effectstest.CastCatalogSpell(t, g, "Social Snub", "Sorcery", socB35SnubOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatal("no \"copy this spell?\" prompt for the caster")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The copy resolves first: every player sacrifices a creature of their choice.
	socB35SnubSacrificeRound(t, g, creatures)
	for i := range g.Seats {
		if g.Battlefield.Contains(creatures[i][0]) || !g.Battlefield.Contains(creatures[i][1]) {
			t.Errorf("seat %d: after the copy, the chosen creature survived or the other one died", i)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Then the original: each player has one creature left, so there is
	// nothing to choose and it is sacrificed without a question.
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("a player with a single creature was asked which to sacrifice")
	}
	for i, p := range g.Seats {
		if g.Battlefield.Contains(creatures[i][0]) || g.Battlefield.Contains(creatures[i][1]) {
			t.Errorf("seat %d kept a creature after two edicts", i)
		}
		want := lives[i] - 2
		if p.ID == me.ID {
			want = lives[i] + 2
		}
		if p.Life != want {
			t.Errorf("seat %d has %d life, want %d (two drains of 1)", i, p.Life, want)
		}
	}
}

func TestSocB35SocialSnubMayDeclineTheCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creatures := socB35SnubTable(g)
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	effectstest.CastCatalogSpell(t, g, "Social Snub", "Sorcery", socB35SnubOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatal("no \"copy this spell?\" prompt for the caster")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	socB35SnubSacrificeRound(t, g, creatures)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if g.Battlefield.Contains(creatures[i][0]) || !g.Battlefield.Contains(creatures[i][1]) {
			t.Errorf("seat %d: wrong creature sacrificed", i)
		}
		want := lives[i] - 1
		if p.ID == me.ID {
			want = lives[i] + 1
		}
		if p.Life != want {
			t.Errorf("seat %d has %d life, want %d", i, p.Life, want)
		}
	}
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("a second edict ran after the copy was declined")
	}
}

func TestSocB35SocialSnubOffersNoCopyWithoutACreatureOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var theirs []uuid.UUID
	for _, p := range g.Seats {
		if p.ID != me.ID {
			theirs = append(theirs, effectstest.B12Creature(g, p.ID, "Bear", socB27BearType, 2, 2))
		}
	}
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Social Snub", "Sorcery", socB35SnubOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm); ask != nil {
		t.Fatal("the copy was offered to a player with no creature")
	}
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the caster, with no creature, was asked to sacrifice")
	}
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		effectstest.AnswerSacrifice(t, g, p.ID, theirs[0])
		theirs = theirs[1:]
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	for _, id := range theirs {
		if !g.Battlefield.Contains(id) {
			t.Error("a creature was sacrificed twice")
		}
	}
}

// --- Spectacle Summit -------------------------------------------------------

func TestSocB35SpectacleSummitEntersTappedMakesBlueOrRedAndSurveilsForFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "Top", "Next")
	id := effectstest.PlayLandFromHand(t, g, "Spectacle Summit", socB35SummitOracle)
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatalf("on battlefield %v, tapped %v; want it to enter tapped", ok, card.Tapped)
	}
	effectstest.UntapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Fatal("the land made {B}")
	}
	for _, colour := range []string{"U", "R"} {
		effectstest.UntapForTest(g, id)
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{colour}}); err != nil {
			t.Fatalf("{T}: Add {%s}: %v", colour, err)
		}
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	}

	effectstest.UntapForTest(g, id)
	effectstest.FloatForTest(g, me, "CCUU")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("surveil activated without {R}")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "CCUR")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{2}{U}{R}, {T}: surveil 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
	top := c.ScryCards[0]
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not go to the graveyard")
	}
	if card, _ := effectstest.BattlefieldCard(g, id); !card.Tapped {
		t.Error("the {T} cost did not tap the land")
	}
}

// --- Spectacular Skywhale ---------------------------------------------------

func socB35Skywhale(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.SocB17Perm(g, owner, "Spectacular Skywhale", "Creature — Elemental Whale", socB35SkywhaleOracle, "{2}{U}{R}", 1, 4)
}

func TestSocB35SkywhaleGetsPlusThreeOnAnInstantOrSorceryUnderFiveMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	whale := socB35Skywhale(g, me.ID)
	if !effectstest.HasEffectiveKeyword(t, g, whale, "flying") {
		t.Error("Spectacular Skywhale has no flying")
	}
	effectstest.SocB17CastSpellFor(t, g, "Instant", "{3}{R}", "RRRR")
	c := effectstest.S58p6View(t, g, whale)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Fatalf("Skywhale is %d/%d after a four-mana spell, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	if got := effectstest.CounterOn(g, whale, game.CounterPlusOne); got != 0 {
		t.Errorf("%d +1/+1 counters after a four-mana spell, want 0", got)
	}
	effectstest.SocB17CastSpellFor(t, g, "Sorcery", "{R}", "R")
	c = effectstest.S58p6View(t, g, whale)
	if c.CurrentPower() != 7 || c.CurrentToughness() != 4 {
		t.Errorf("Skywhale is %d/%d after a second spell, want 7/4", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.EndTurn(t, g)
	c = effectstest.S58p6View(t, g, whale)
	if c.CurrentPower() != 1 {
		t.Errorf("the +3/+0 outlasted the turn: power %d", c.CurrentPower())
	}
}

func TestSocB35SkywhalePutsThreeCountersInsteadOnFiveOrMoreMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	whale := socB35Skywhale(g, me.ID)
	effectstest.SocB17CastSpellFor(t, g, "Sorcery", "{4}{R}", "RRRRR")
	if got := effectstest.CounterOn(g, whale, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters after a five-mana spell, want 3", got)
	}
	c := effectstest.S58p6View(t, g, whale)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 7 {
		t.Errorf("Skywhale is %d/%d, want 4/7 (counters instead of +3/+0)", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.EndTurn(t, g)
	c = effectstest.S58p6View(t, g, whale)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 7 {
		t.Errorf("Skywhale is %d/%d next turn, want the counters to stay (4/7)", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSocB35SkywhaleIgnoresCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	whale := socB35Skywhale(g, me.ID)
	effectstest.SocB17CastSpellFor(t, g, "Creature — Bear", "{4}{R}", "RRRRR")
	c := effectstest.S58p6View(t, g, whale)
	if c.CurrentPower() != 1 || effectstest.CounterOn(g, whale, game.CounterPlusOne) != 0 {
		t.Errorf("Skywhale is %d/%d after a creature spell, want it unchanged", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Spellbook Seeker // Careful Study --------------------------------------

func TestSocB35SpellbookSeekerFliesEntersPreparedAndCarefulStudyLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	c := sos1Prep(me.ID, socB35SeekerOracle, []string{"flying"}, "Spellbook Seeker", "Creature — Bird Wizard", "{3}{U}", 3, 3,
		"Careful Study", "Sorcery", "{U}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Spellbook Seeker did not enter prepared")
	}
	if !effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") {
		t.Error("Spellbook Seeker has no flying")
	}
	effectstest.Msc18Library(me, 6)
	hand := me.Hand.Size()
	effectstest.FraBCastCopy(t, g, me, "Careful Study")
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand %d after drawing two, want %d", me.Hand.Size(), hand+2)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after discarding two, want %d", me.Hand.Size(), hand)
	}
}

// --- Spiritcall Enthusiast // Scrollboost -----------------------------------

func socB35Enthusiast(t *testing.T, g *game.Game, me *game.Player) game.Card {
	t.Helper()
	c := sos1Prep(me.ID, socB35EnthusiastOracle, nil, "Spiritcall Enthusiast", "Creature — Cat Cleric", "{2}{W}", 3, 3,
		"Scrollboost", "Sorcery", "{1}{W}")
	effectstest.CastFromHandAndResolve(t, g, me, c)
	return c
}

func TestSocB35EnthusiastBecomesPreparedWhenATokenYouControlEntersAndScrollboostPumpsTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	c := socB35Enthusiast(t, g, me)
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("the Enthusiast entered prepared")
	}
	socB35Token(t, g, me.ID, "1/1 colorless Spirit")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("a token entering did not prepare the Enthusiast")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB27BearType, 2, 2)
	other := effectstest.B12Creature(g, g.Seats[1].ID, "Other", socB27BearType, 3, 3)
	effectstest.FraBCastCopy(t, g, me, "Scrollboost", effectstest.CardRefs(bear, other)...)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("first target is %d/%d, want 4/4", p, tough)
	}
	if p, tough := effectstest.EffectivePower(t, g, other), effectstest.EffectiveToughness(t, g, other); p != 5 || tough != 5 {
		t.Errorf("second target is %d/%d, want 5/5", p, tough)
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
	effectstest.EndTurn(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("the pump outlasted the turn: power %d", p)
	}
}

func TestSocB35EnthusiastIgnoresATokenAnOpponentControlsAndANonTokenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	c := socB35Enthusiast(t, g, me)
	socB35Token(t, g, g.Seats[1].ID, "1/1 colorless Spirit")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("the Enthusiast prepared without a token of its controller entering")
	}
}

func TestSocB35ScrollboostTakesOneTargetAndRefusesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	c := socB35Enthusiast(t, g, me)
	socB35Token(t, g, me.ID, "1/1 colorless Spirit")
	effectstest.PassPriorityAroundTable(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", socB27BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", socB27BearType, 2, 2)
	d := effectstest.B12Creature(g, me.ID, "Bear C", socB27BearType, 2, 2)
	cp, ok := effectstest.PrepareCopyOf(g, "Scrollboost")
	if !ok {
		t.Fatal("no Scrollboost copy in exile")
	}
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), Targets: effectstest.CardRefs(bear, b, d), Strict: true}); err == nil {
		t.Fatal("Scrollboost took three targets")
	}
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("a refused cast unprepared the Enthusiast")
	}
	effectstest.FraBCastCopy(t, g, me, "Scrollboost", effectstest.CardRefs(bear)...)
	if p := effectstest.EffectivePower(t, g, bear); p != 4 {
		t.Errorf("the lone target's power is %d, want 4", p)
	}
}

// --- Spirit Mascot ----------------------------------------------------------

func TestSocB35SpiritMascotGrowsWhenACardLeavesYourGraveyardOncePerBatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mascot := effectstest.PushCatalogPermanent(g, me.ID, "Spirit Mascot", "Creature — Spirit Ox", socB35MascotOracle, false)
	dead := effectstest.SocB11Graveyard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	socB11Move(t, g, me.ID, dead, game.ZoneGraveyard, game.ZoneExile)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mascot, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after a card left the graveyard, want 1", got)
	}
	theirDead := effectstest.SocB11Graveyard(opp, "Their Dead", "Creature — Bear", "{1}{G}")
	socB11Move(t, g, opp.ID, theirDead, game.ZoneGraveyard, game.ZoneExile)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mascot, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after an opposing graveyard card left, want 1", got)
	}
	// A card going to the hand leaves the graveyard too.
	second := effectstest.SocB11Graveyard(me, "Second Bear", "Creature — Bear", "{1}{G}")
	socB11Move(t, g, me.ID, second, game.ZoneGraveyard, game.ZoneHand)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mascot, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters after a second card left, want 2", got)
	}
}

func TestSocB35SpiritMascotCountsCastingACardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	mascot := effectstest.PushCatalogPermanent(g, me.ID, "Spirit Mascot", "Creature — Spirit Ox", socB35MascotOracle, false)
	// A card entering the graveyard is not one leaving it.
	looting := effectstest.SeedGraveyardCard(t, g, "Faithless Looting", "Sorcery", "3d6fa57a-aa53-4b5c-b8af-a7612c823117")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mascot, game.CounterPlusOne); got != 0 {
		t.Fatalf("%d counters after a card entered the graveyard, want 0", got)
	}
	// Casting it with flashback takes it out of the graveyard (CR 601.2a).
	if err := g.CastSpell(me.ID, looting, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	for i := 0; i < 8 && effectstest.CounterOn(g, mascot, game.CounterPlusOne) < 1; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if got := effectstest.CounterOn(g, mascot, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after casting a card from the graveyard, want 1", got)
	}
}

// --- Stadium Tidalmage ------------------------------------------------------

func TestSocB35TidalmageMayLootWhenItEntersAndWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(me, 8)
	socB35CastTidalmage(t, g, me)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatal("no \"draw a card?\" prompt on entering")
	}
	hand := me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after drawing, want %d", me.Hand.Size(), hand+1)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// The attack trigger: decline, and nothing happens.
	tidal := socB35TidalmageOnBattlefield(g, me.ID)
	effectstest.DeclareAttack(t, g, opp.ID, tidal)
	effectstest.PassPriorityAroundTable(t, g)
	ask = effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatal("no \"draw a card?\" prompt on attacking")
	}
	hand = me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm(decline): %v", err)
	}
	if me.Hand.Size() != hand {
		t.Errorf("declining the draw changed the hand: %d, want %d", me.Hand.Size(), hand)
	}
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("declining the draw still asked for a discard")
	}
}

func socB35TidalmageOnBattlefield(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Stadium Tidalmage", "Creature — Djinn Sorcerer", socB35TidalmageOracle, 4, 4)
}

// socB35CastTidalmage casts a free Stadium Tidalmage and resolves the spell, leaving
// the enters trigger's prompt open.
func socB35CastTidalmage(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id, err := effectstest.Msc17Cast(t, g, me, game.Card{
		Name: "Stadium Tidalmage", TypeLine: "Creature — Djinn Sorcerer", OracleID: socB35TidalmageOracle,
		Power: 4, Toughness: 4,
	}, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast Stadium Tidalmage: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}
