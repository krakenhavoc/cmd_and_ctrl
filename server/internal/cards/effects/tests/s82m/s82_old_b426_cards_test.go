package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b426_cards_test.go — Return to Ravnica: enters and dies triggers,
// instants and sorceries, creatures with keywords, other creatures, lands (S82
// older sets, batch 426, #5566). Each test plays the real catalog entry
// through the engine.

const (
	oldB426Revenant   = "7191f765-6e66-4bc4-8aa8-85280ac61d1e"
	oldB426City       = "fe1b0b19-0994-4bc7-ab6c-d1419d274f40"
	oldB426Games      = "c38c7899-b774-489b-9eb7-084522635a3f"
	oldB426Molding    = "e01c8122-9159-4f28-ac6c-338bd889650e"
	oldB426Jailbreak  = "08d27af8-7178-4694-8621-fabac54c8f0a"
	oldB426Innkeeper  = "1fa94925-04e5-4073-91af-1857dbc09991"
	oldB426Grove      = "f746612a-fbed-44ca-b2cc-5928e10cf4bb"
	oldB426BearType   = "Creature — Bear"
	oldB426Enchantmnt = "Enchantment"
)

func TestOldB426EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB426Revenant:  "Dark Revenant",
		oldB426City:      "Search the City",
		oldB426Games:     "Slaughter Games",
		oldB426Molding:   "Slime Molding",
		oldB426Jailbreak: "Ogre Jailbreaker",
		oldB426Innkeeper: "Oak Street Innkeeper",
		oldB426Grove:     "Grove of the Guardian",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Dark Revenant ----------------------------------------------------------

func TestOldB426DarkRevenantFliesAndReturnsToTheTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Mine, but an opponent controls it: it goes to MY library when it dies.
	id := effectstest.B12Push(g, me.ID, "Dark Revenant", "Creature — Spirit", oldB426Revenant, 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("no flying")
	}
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.SeedLibrary(opp, "C", "D")
	oppLib := opp.Library.Size()
	effectstest.B25Destroy(g, id)
	effectstest.PassPriorityAroundTable(t, g)
	top, err := me.Library.Top()
	if err != nil || top.InstanceID != id {
		t.Errorf("the top of my library is %v, want the Revenant", top.Name)
	}
	if opp.Library.Size() != oppLib {
		t.Errorf("the opponent's library went from %d to %d cards, want unchanged", oppLib, opp.Library.Size())
	}
}

// --- Slaughter Games --------------------------------------------------------

func TestOldB426SlaughterGamesExilesTheNamedNonlandCardFromEveryZone(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Lightning Bolt", "Instant")
	effectstest.CastCatalogSpell(t, x.G, "Slaughter Games", "Sorcery", oldB426Games, []game.TargetRef{{Kind: game.TargetPlayer, ID: x.Opp.ID}})
	effectstest.PassPriorityAroundTable(t, x.G)
	effectstest.ActNameFor(t, x.G, x.Me.ID, "lightning bolt")
	effectstest.AnswerSearchByID(t, x.G, x.Me.ID, x.Cards["hand"], x.Cards["lib"], x.Cards["gy"], x.Cards["gy 2"])
	effectstest.PassPriorityAroundTable(t, x.G)
	x.Exiled(t, "gy", "gy 2", "hand", "lib")
	x.NotExiled(t, "gy other", "hand other", "lib other", "gy forest")
	if x.Searches() != 1 {
		t.Errorf("the library was searched and shuffled %d times, want 1", x.Searches())
	}
}

func TestOldB426SlaughterGamesFindsNothingForALandName(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, x.G, "Slaughter Games", "Sorcery", oldB426Games, []game.TargetRef{{Kind: game.TargetPlayer, ID: x.Opp.ID}})
	effectstest.PassPriorityAroundTable(t, x.G)
	effectstest.ActNameFor(t, x.G, x.Me.ID, "Forest")
	if c := effectstest.SearchPromptFor(x.G, x.Me.ID); c != nil && len(c.SearchCards) != 0 {
		t.Errorf("a search offered %d candidates for a land name", len(c.SearchCards))
	}
	x.NotExiled(t, "gy", "gy 2", "hand", "lib", "gy forest")
}

func TestOldB426SlaughterGamesCantBeCountered(t *testing.T) {
	spec, ok := effects.Lookup(oldB426Games)
	if !ok || !spec.CantBeCountered {
		t.Error("Slaughter Games can be countered")
	}
}

// --- Slime Molding ----------------------------------------------------------

func TestOldB426SlimeMoldingMakesAnXByXOoze(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if err := effectstest.CastX(t, g, "Slime Molding", "Sorcery", oldB426Molding, 4, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.MidB03Tokens(g, me.ID, "Ooze")
	if len(got) != 1 {
		t.Fatalf("%d Ooze tokens, want 1", len(got))
	}
	if p, tough := effectstest.PtOf(t, g, got[0].InstanceID); p != 4 || tough != 4 {
		t.Errorf("Ooze is %d/%d, want 4/4", p, tough)
	}
	if cols := got[0].EffectiveColors(); len(cols) != 1 || cols[0] != "G" {
		t.Errorf("Ooze colours %v, want green", cols)
	}
}

// --- Ogre Jailbreaker -------------------------------------------------------

func TestOldB426OgreJailbreakerAttacksOnlyWhileYouControlAGate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Push(g, me.ID, "Ogre Jailbreaker", "Creature — Ogre Rogue", oldB426Jailbreak, 4, 4)
	if !effectstest.HasKeywordOnBattlefield(t, g, ogre, "defender") {
		t.Fatal("no defender")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ogre, opp.ID); err == nil {
		t.Fatal("it attacked with no Gate")
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Gate", TypeLine: "Land — Gate",
		Owner: me.ID, Controller: me.ID})
	if err := g.DeclareAttacker(ogre, opp.ID); err != nil {
		t.Fatalf("it can't attack with a Gate: %v", err)
	}
}

func TestOldB426OgreJailbreakerIgnoresAnOpponentsGate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Push(g, me.ID, "Ogre Jailbreaker", "Creature — Ogre Rogue", oldB426Jailbreak, 4, 4)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Gate", TypeLine: "Land — Gate",
		Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ogre, opp.ID); err == nil {
		t.Error("an opponent's Gate let it attack")
	}
}

// --- Oak Street Innkeeper ---------------------------------------------------

func TestOldB426OakStreetInnkeeperGivesTappedCreaturesHexproofOnOtherTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, other.ID, "Oak Street Innkeeper", "Creature — Elf", oldB426Innkeeper, 1, 2)
	tapped := effectstest.B12Creature(g, other.ID, "Tapped Bear", oldB426BearType, 2, 2)
	untapped := effectstest.B12Creature(g, other.ID, "Untapped Bear", oldB426BearType, 2, 2)
	foreign := effectstest.B12Creature(g, active.ID, "Foreign Bear", oldB426BearType, 2, 2)
	effectstest.FindBattlefieldCardByID(g, tapped).Tapped = true
	effectstest.FindBattlefieldCardByID(g, foreign).Tapped = true
	if !effectstest.HasEffectiveKeyword(t, g, tapped, "hexproof") {
		t.Error("a tapped creature has no hexproof on another player's turn")
	}
	if effectstest.HasEffectiveKeyword(t, g, untapped, "hexproof") {
		t.Error("an untapped creature has hexproof")
	}
	if effectstest.HasEffectiveKeyword(t, g, foreign, "hexproof") {
		t.Error("a creature its controller's opponent controls has hexproof")
	}
}

func TestOldB426OakStreetInnkeeperGivesNothingOnYourOwnTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	keeper := effectstest.B12Push(g, me.ID, "Oak Street Innkeeper", "Creature — Elf", oldB426Innkeeper, 1, 2)
	effectstest.FindBattlefieldCardByID(g, keeper).Tapped = true
	if effectstest.HasEffectiveKeyword(t, g, keeper, "hexproof") {
		t.Error("hexproof on its controller's own turn")
	}
}

// --- Grove of the Guardian --------------------------------------------------

func TestOldB426GroveOfTheGuardianTapsForColourless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.S58p6Push(g, me.ID, "Grove of the Guardian", "Land", oldB426Grove, "", 0, 0)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {C}: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool %v, want one mana", me.ManaPool.Units())
	}
}

func TestOldB426GroveOfTheGuardianMakesAnEightEightElemental(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Grove of the Guardian", "Land", oldB426Grove, false)
	a := effectstest.B12Creature(g, me.ID, "Bear A", oldB426BearType, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", oldB426BearType, 2, 2)
	effectstest.ApaMana(me, "C", "C", "C", "G", "W")
	effectstest.ApaActivate(t, g, me, land, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{a, b}})
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, land) {
		t.Error("the land was not sacrificed")
	}
	for _, id := range []uuid.UUID{a, b} {
		if !effectstest.PermanentTapped(g, id) {
			t.Error("a creature was not tapped for the cost")
		}
	}
	toks := effectstest.MidB03Tokens(g, me.ID, "Elemental")
	if len(toks) != 1 {
		t.Fatalf("%d Elemental tokens, want 1", len(toks))
	}
	if p, tough := effectstest.PtOf(t, g, toks[0].InstanceID); p != 8 || tough != 8 {
		t.Errorf("Elemental is %d/%d, want 8/8", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, toks[0].InstanceID, "vigilance") {
		t.Error("the Elemental has no vigilance")
	}
}

func TestOldB426GroveOfTheGuardianNeedsTwoUntappedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Grove of the Guardian", "Land", oldB426Grove, false)
	a := effectstest.B12Creature(g, me.ID, "Bear A", oldB426BearType, 2, 2)
	effectstest.ApaMana(me, "C", "C", "C", "G", "W")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{a}}); err == nil {
		t.Error("activated with one creature to tap")
	}
}

// --- Search the City --------------------------------------------------------

func oldB426SearchTheCity(t *testing.T, g *game.Game, me *game.Player) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	ids := effectstest.SeedLibrary(me, "Card A", "Card B", "Card C", "Card D", "Card E", "Card F")
	city := effectstest.CastCatalogSpell(t, g, "Search the City", oldB426Enchantmnt, oldB426City, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	return city, ids
}

func TestOldB426SearchTheCityExilesFiveAndReturnsAPlayedName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	city, ids := oldB426SearchTheCity(t, g, me)
	for _, id := range ids[:5] {
		if !g.Exile.Contains(id) {
			t.Fatalf("one of the top five cards was not exiled")
		}
	}
	if g.Exile.Contains(ids[5]) {
		t.Fatal("the sixth card was exiled")
	}
	// Cast a card named like the exiled "Card B".
	effectstest.CastCatalogSpell(t, g, "Card B", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt to take the exiled card")
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != ids[1] || c.ChooseMin != 0 {
		t.Fatalf("offered %v (min %d), want only the exiled Card B and a decline", c.ChooseCards, c.ChooseMin)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the exiled card did not go to its owner's hand")
	}
	if !effects.OnBattlefield(g, city) {
		t.Error("the enchantment was sacrificed with four cards still exiled")
	}
	if len(g.ExtraTurnsQueued()) != 0 {
		t.Error("an extra turn was taken early")
	}
}

func TestOldB426SearchTheCityIgnoresAnUnrelatedPlayAndMayBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_, ids := oldB426SearchTheCity(t, g, me)
	effectstest.CastCatalogSpell(t, g, "Something Else", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("an unrelated spell triggered the enchantment")
	}
	effectstest.CastCatalogSpell(t, g, "Card A", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	if me.Hand.Contains(ids[0]) || !g.Exile.Contains(ids[0]) {
		t.Error("declining the may still moved the card")
	}
}

func TestOldB426SearchTheCitySacrificesItselfAndTakesAnExtraTurnWhenTheLastCardLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	city, ids := oldB426SearchTheCity(t, g, me)
	names := []string{"Card A", "Card B", "Card C", "Card D", "Card E"}
	for i, name := range names {
		effectstest.CastCatalogSpell(t, g, name, "Sorcery", "", nil)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.LatestChooseCardsFor(g, me.ID)
		if c == nil {
			t.Fatalf("no prompt for %s", name)
		}
		if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[i]}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if i < len(names)-1 && !effects.OnBattlefield(g, city) {
			t.Fatalf("the enchantment left after %d of 5 cards", i+1)
		}
	}
	if effects.OnBattlefield(g, city) {
		t.Error("the enchantment was not sacrificed with no cards exiled")
	}
	if got := len(g.ExtraTurnsQueued()); got != 1 {
		t.Errorf("%d extra turns queued, want 1", got)
	}
}

func TestOldB426SearchTheCityTriggersOnAPlayedLandToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	// The top five are all Forests; playing a Forest from hand returns one.
	for i := 0; i < 5; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	effectstest.CastCatalogSpell(t, g, "Search the City", oldB426Enchantmnt, oldB426City, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	hand := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: hand, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	_ = hand
	effectstest.B31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) == nil {
		t.Error("playing a land with an exiled name did not trigger the enchantment")
	}
}
