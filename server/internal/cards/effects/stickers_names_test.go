package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers_names_test.go — the `_____` cards and the counts on a name
// sticker (CR 123.6d, 123.6e; ADR 0208, #4307).

const (
	blankGoblinOracle      = "88222fd2-8316-426c-8218-64f6be5ca0f8"
	blankBirdOracle        = "2ce375c8-7c37-40b3-95cf-e0aa958bb317"
	blankOSaurusOracle     = "32921952-8846-418d-8689-c42dbc24a7b8"
	wizardsOfTheBlankOracl = "786b0b50-adc7-4d58-b578-1e80680b89bd"
	wolfInBlankOracle      = "58c9454b-7027-48f4-9e17-310e27b916d1"
	blankBallsOracle       = "82c2a58b-0ee2-454b-9ddd-670c1e788166"
	makeABlankSplashOracle = "fe2601a4-cbd5-4866-92a2-5586f737c9af"
	blankTrespasserOracle  = "75854e14-5f15-443d-b342-b27fc3ab4b6e"
	blankRocketshipOracle  = "1009f979-b1fa-4427-bf9e-2a533a4ca19b"
)

// stkPrinted is a sheet by its printed name, as the deck importer reads it.
func stkPrinted(t *testing.T, name string) game.StickerSheet {
	t.Helper()
	for _, s := range stkPrintedSheets {
		if s.Name == name {
			return deck.StickerSheetOf(cards.Card{
				OracleID: uuid.MustParse(s.Oracle), Name: s.Name, TypeLine: stickerSheetTypeLine, OracleText: s.Text,
			})
		}
	}
	t.Fatalf("no printed sheet %q", name)
	return game.StickerSheet{}
}

// stkTableWith is a catalog game whose seat 0 plays with the named printed
// sheets and has `tickets` ticket counters.
func stkTableWith(t *testing.T, tickets int, sheets ...string) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		for _, s := range sheets {
			me.StickerSheets = append(me.StickerSheets, stkPrinted(t, s))
		}
		if tickets > 0 {
			if err := g.AddPlayerCounterByForEffect(me.ID, me.ID, game.CounterTicket, tickets); err != nil {
				t.Fatalf("tickets: %v", err)
			}
		}
	})
	return g, me
}

// stkEnterWithName casts the creature, resolves its enters trigger up to the
// first sticker question and answers the three with the name sticker whose
// label contains `word`, `after` words in. It returns the creature.
func stkEnterWithName(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle, word string, after int) uuid.UUID {
	t.Helper()
	id := castAndResolveCreature(t, g, name, typeLine, oracle)
	passPriorityAroundTable(t, g)
	stkPutOn(t, g, me, id, word, after)
	return id
}

func stkNames(c game.Card) string { return c.Effective().Name }

// CR 123.6e: the number of DIFFERENT vowels on a name sticker; Y counts,
// repeats do not, case does not matter.
func TestStkUniqueVowelsAndLetters(t *testing.T) {
	for _, tc := range []struct {
		text   string
		vowels int
	}{
		{"Night", 1}, {"Brushwagg", 2}, {"Ringmaster", 3}, {"Mystic", 2}, {"Doom", 1},
		{"Yogurt", 3}, {"Hot Dog", 1}, {"Rhythm", 1}, {"NYMPH", 1}, {"Eerie", 2}, {"", 0}, {"Brr", 0},
		{"AEIOUY", 6}, {"aeiouy", 6},
	} {
		if got := game.StickerUniqueVowels(tc.text); got != tc.vowels {
			t.Errorf("StickerUniqueVowels(%q) = %d, want %d", tc.text, got, tc.vowels)
		}
	}
	// CR 123.6d: case-insensitive letter counts; letters only.
	if got := game.StickerLetterCount("Monkey Moo", 'O'); got != 3 {
		t.Errorf("o's in %q = %d, want 3", "Monkey Moo", got)
	}
	if got := game.StickerLetters("Urza's"); got != 5 {
		t.Errorf("letters in Urza's = %d, want 5 (the apostrophe is not a letter)", got)
	}
	if got := game.StickerLetters("Deep-Fried"); got != 9 {
		t.Errorf("letters in Deep-Fried = %d, want 9", got)
	}
	if !game.StickerBeginsWith("Ringmaster", 'r') || game.StickerBeginsWith("Ringmaster", 'g') {
		t.Error("StickerBeginsWith is wrong")
	}
}

// _____ Goblin: {R} for each unique vowel on the sticker it just put on.
func TestStkBlankGoblinAddsManaForUniqueVowels(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	stkEnterWithName(t, g, me, "_____ Goblin", "Creature — Goblin Guest", blankGoblinOracle, "“Ringmaster”", 0)
	if got := poolCount(me, "R"); got != 3 {
		t.Errorf("{R} in the pool = %d, want 3 (Ringmaster has I, A and E)", got)
	}
}

// "You may": declined, there is no sticker and no vowel.
func TestStkBlankGoblinDeclinedAddsNothing(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	id := castAndResolveCreature(t, g, "_____ Goblin", "Creature — Goblin Guest", blankGoblinOracle)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no question asks whether to put a sticker on it")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if got := poolCount(me, "R"); got != 0 {
		t.Errorf("{R} in the pool = %d, want 0", got)
	}
	if c := stkCard(t, g, id); len(c.Stickers) != 0 {
		t.Errorf("declined, yet stickers = %+v", c.Stickers)
	}
}

// Only a name sticker is offered: ability, power and toughness and art
// stickers are not "a name sticker" (CR 123.1).
func TestStkBlankGoblinOffersOnlyNameStickers(t *testing.T) {
	g, me := stkTableWith(t, 9, "Night Brushwagg Ringmaster")
	castAndResolveCreature(t, g, "_____ Goblin", "Creature — Goblin Guest", blankGoblinOracle)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{pick.ChooseCards[0]}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	which := latestOptionPickFor(g, me.ID)
	for _, o := range which.PickOptions {
		if o.Label == game.DeclineStickerLabel {
			continue
		}
		if len(o.Label) < 12 || o.Label[:12] != "Name sticker" {
			t.Errorf("offered %q, which is not a name sticker", o.Label)
		}
	}
}

// The count survives a restore taken while a sticker question is open: the
// record of what was put on is data (step record), not a Go variable.
func TestStkBlankGoblinSurvivesARestoreBetweenTheQuestions(t *testing.T) {
	g, _ := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	id := castAndResolveCreature(t, g, "_____ Goblin", "Creature — Goblin Guest", blankGoblinOracle)
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me := g.Seats[0]
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the restored table lost the object question")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{id}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	which := latestOptionPickFor(g, me.ID)
	if which == nil {
		t.Fatal("the restored table lost the sticker question")
	}
	if err := g.ResolveOptionPick(which.ID, me.ID, stkOption(t, which, "“Brushwagg”")); err != nil {
		t.Fatalf("pick: %v", err)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	where := latestOptionPickFor(g, me.ID)
	if where == nil {
		t.Fatal("the restored table lost the position question")
	}
	if err := g.ResolveOptionPick(where.ID, me.ID, 0); err != nil {
		t.Fatalf("position: %v", err)
	}
	if got := poolCount(g.Seats[0], "R"); got != 2 {
		t.Errorf("{R} after a restore between the questions = %d, want 2 (Brushwagg has U and A)", got)
	}
	if got := stkNames(stkCard(t, g, id)); got != "Brushwagg _____ Goblin" {
		t.Errorf("name = %q, want %q", got, "Brushwagg _____ Goblin")
	}
}

// _____ Bird Gets the Worm: X life for the unique vowels.
func TestStkBlankBirdGainsLife(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	before := me.Life
	stkEnterWithName(t, g, me, "_____ Bird Gets the Worm", "Creature — Bird Guest", blankBirdOracle, "“Night”", 1)
	if me.Life != before+1 {
		t.Errorf("life = %d, want %d (Night has one vowel)", me.Life, before+1)
	}
}

// _____-o-saurus: a +1/+1 counter per unique vowel.
func TestStkBlankOSaurusGetsCounters(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	id := stkEnterWithName(t, g, me, "_____-o-saurus", "Creature — Alien Dinosaur", blankOSaurusOracle, "“Ringmaster”", 0)
	if got := countersOn(g, id, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
}

// Wizards of the _____: look at the top X cards, one to hand, the rest on
// the bottom.
func TestStkWizardsOfTheBlankLooksAtTheTopX(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	ids := seedLibrary(me, "A", "B", "C", "D", "E")
	handBefore := me.Hand.Size()
	stkEnterWithName(t, g, me, "Wizards of the _____", "Creature — Human Wizard Performer", wizardsOfTheBlankOracl, "“Ringmaster”", 0)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("Wizards of the _____ raised no take prompt")
	}
	if len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("take exactly one of %d (want 3, Ringmaster's vowels): %d..%d", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("take: %v", err)
	}
	if !me.Hand.Contains(ids[1]) || me.Hand.Size() != handBefore+1 {
		t.Error("the picked card did not go to hand")
	}
}

// Declined: X = 0, so nothing is looked at.
func TestStkWizardsOfTheBlankDeclinedLooksAtNothing(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	seedLibrary(me, "A", "B", "C")
	castAndResolveCreature(t, g, "Wizards of the _____", "Creature — Human Wizard Performer", wizardsOfTheBlankOracl)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if p := latestChooseCardsFor(g, me.ID); p != nil {
		t.Errorf("a take prompt is open with no sticker put on: %+v", p.ChooseCards)
	}
}

// Wolf in _____ Clothing: "When you do" is a reflexive trigger with up to X
// targets; each gets -1/-1 until end of turn.
func TestStkWolfInBlankClothingShrinksUpToX(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	opp := g.Seats[1]
	a := p2Push(g, opp.ID, "Bear A", "Creature — Bear", "", 3, 3)
	b := p2Push(g, opp.ID, "Bear B", "Creature — Bear", "", 3, 3)
	stkEnterWithName(t, g, me, "Wolf in _____ Clothing", "Creature — Wolf Guest", wolfInBlankOracle, "“Brushwagg”", 0)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the reflexive trigger raised no target prompt")
	}
	if prompt.PickTargetMax != 2 {
		t.Errorf("up to %d targets, want 2 (Brushwagg has U and A)", prompt.PickTargetMax)
	}
	if err := g.ResolvePickTargets(prompt.ID, me.ID, []game.TargetRef{p2Card(a), p2Card(b)}); err != nil {
		t.Fatalf("pick targets: %v", err)
	}
	// The reflexive trigger is on the stack with its targets chosen: a
	// restore point, and "up to X" is data on the trigger (Params.Amount).
	g = restoreRoundTrip(t, g, true)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := p2PT(g, id); p != 2 || tough != 2 {
			t.Errorf("target P/T = %d/%d, want 2/2", p, tough)
		}
	}
}

// Declined: "when you do" never happens.
func TestStkWolfInBlankClothingDeclinedDoesNothing(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	opp := g.Seats[1]
	a := p2Push(g, opp.ID, "Bear A", "Creature — Bear", "", 3, 3)
	castAndResolveCreature(t, g, "Wolf in _____ Clothing", "Creature — Wolf Guest", wolfInBlankOracle)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a target prompt is open though no sticker went on")
	}
	if p, tough := p2PT(g, a); p != 3 || tough != 3 {
		t.Errorf("bear P/T = %d/%d, want 3/3", p, tough)
	}
}

// _____ Balls of Fire: the sticker its enters trigger puts on triggers the
// second ability; damage is the o's in name stickers on the enchantment, now.
func TestStkBlankBallsOfFireDealsDamageForOs(t *testing.T) {
	g, me := stkTableWith(t, 0, "Mystic Doom Sandwich")
	opp := g.Seats[1]
	before := opp.Life
	id := castAndResolveCreature(t, g, "_____ Balls of Fire", "Enchantment", blankBallsOracle)
	passPriorityAroundTable(t, g)
	stkPutOn(t, g, me, id, "“Doom”", 0)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the sticker put on it raised no target prompt")
	}
	if err := g.ResolvePickTargets(prompt.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life = %d, want %d (Doom has two o's)", opp.Life, before-2)
	}
	// A later sticker, whichever effect puts it on, triggers it again and
	// the count is of every name sticker on it.
	stkGive(t, g, me, id, "“Mystic”")
	prompt = latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("a second sticker raised no target prompt")
	}
	if err := g.ResolvePickTargets(prompt.ID, me.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("pick the opponent: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("life = %d, want %d (still two o's on the enchantment)", opp.Life, before-4)
	}
}

// A sticker put on some other permanent does not trigger it (CR 603.2: "on
// this enchantment").
func TestStkBlankBallsOfFireIgnoresStickersOnOtherPermanents(t *testing.T) {
	g, me := stkTableWith(t, 0, "Mystic Doom Sandwich")
	castAndResolveCreature(t, g, "_____ Balls of Fire", "Enchantment", blankBallsOracle)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	bear := p2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	stkGive(t, g, me, bear, "“Doom”")
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a sticker on a Bear triggered the enchantment")
	}
}

// Make a _____ Splash: up to X targets, X the u's in name stickers on it,
// chosen as the ability goes on the stack.
func TestStkMakeABlankSplashTapsUpToX(t *testing.T) {
	g, me := stkTableWith(t, 0, "Cursed Firebreathing Yogurt")
	opp := g.Seats[1]
	a := p2Push(g, opp.ID, "Bear A", "Creature — Bear", "", 2, 2)
	b := p2Push(g, opp.ID, "Bear B", "Creature — Bear", "", 2, 2)
	id := castAndResolveCreature(t, g, "Make a _____ Splash", "Enchantment", makeABlankSplashOracle)
	passPriorityAroundTable(t, g)
	stkPutOn(t, g, me, id, "“Cursed”", 0)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the sticker put on it raised no target prompt")
	}
	if prompt.PickTargetMax != 1 {
		t.Fatalf("up to %d targets, want 1 (one u in Cursed)", prompt.PickTargetMax)
	}
	if err := g.ResolvePickTargets(prompt.ID, me.ID, []game.TargetRef{p2Card(a)}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !stkTapped(g, a) || stkTapped(g, b) {
		t.Errorf("tapped: a=%v b=%v, want only a", stkTapped(g, a), stkTapped(g, b))
	}
	// The second sticker counts every name sticker on it: two u's now.
	stkGive(t, g, me, id, "“Yogurt”")
	prompt = latestPickTarget(g, me.ID)
	if prompt == nil || prompt.PickTargetMax != 2 {
		t.Fatalf("second prompt = %+v, want up to 2", prompt)
	}
	if err := g.ResolvePickTargets(prompt.ID, me.ID, []game.TargetRef{p2Card(b)}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !stkTapped(g, b) {
		t.Error("the second ability did not tap Bear B")
	}
}

func stkTapped(g *game.Game, id uuid.UUID) bool {
	tapped := false
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			tapped = c.Tapped
		}
	})
	return tapped
}

// With no u on its stickers the ability is "up to 0": it goes on the stack
// with no target and taps nothing.
func TestStkMakeABlankSplashWithNoUTapsNothing(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	opp := g.Seats[1]
	a := p2Push(g, opp.ID, "Bear A", "Creature — Bear", "", 2, 2)
	id := castAndResolveCreature(t, g, "Make a _____ Splash", "Enchantment", makeABlankSplashOracle)
	passPriorityAroundTable(t, g)
	stkPutOn(t, g, me, id, "“Night”", 0)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a target prompt is open with no u on the stickers")
	}
	passPriorityAroundTable(t, g)
	if stkTapped(g, a) {
		t.Error("Bear A was tapped")
	}
}

// _____ _____ _____ Trespasser: +1/+0 for each name sticker on it, and it
// can't be blocked this turn.
func TestStkBlankTrespasserPumpsPerNameSticker(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	id := stkEnterWithName(t, g, me, "_____ _____ _____ Trespasser", "Creature — Human Rogue Guest", blankTrespasserOracle, "“Night”", 0)
	stkGive(t, g, me, id, "“Ringmaster”")
	base, _ := p2PT(g, id)
	if err := p2Activate(t, g, me, id, 0, "UUUU"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := p2PT(g, id); p != base+2 {
		t.Errorf("power = %d, want %d (+1 for each of the two name stickers)", p, base+2)
	}
	assertRestrictions(t, g, id, game.CantBeBlocked)
}

// _____ _____ Rocketship: up to two name stickers on it as it enters.
func TestStkBlankRocketshipTakesUpToTwoNameStickers(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	id := castAndResolveCreature(t, g, "_____ _____ Rocketship", "Artifact — Vehicle", blankRocketshipOracle)
	passPriorityAroundTable(t, g)
	stkPutOn(t, g, me, id, "“Night”", 0)
	stkPutOn(t, g, me, id, "“Ringmaster”", 0)
	c := stkCard(t, g, id)
	if got := game.StickerCountOfKind(&c, game.StickerName); got != 2 {
		t.Errorf("name stickers = %d, want 2", got)
	}
	// A third question is not asked.
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Error("a third sticker question is open")
	}
}

// It may decline the first and still take one, or take none.
func TestStkBlankRocketshipMayTakeNone(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	id := castAndResolveCreature(t, g, "_____ _____ Rocketship", "Artifact — Vehicle", blankRocketshipOracle)
	passPriorityAroundTable(t, g)
	for i := 0; i < 2; i++ {
		pick := latestChooseCardsFor(g, me.ID)
		if pick == nil {
			t.Fatalf("question %d is not open", i+1)
		}
		if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("decline %d: %v", i+1, err)
		}
	}
	if c := stkCard(t, g, id); len(c.Stickers) != 0 {
		t.Errorf("stickers = %+v, want none", c.Stickers)
	}
}

// stkAttackWith declares the creature as an attacker and settles the attack
// triggers.
func stkAttackWith(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
}

// "Whenever this Vehicle attacks, choose a letter": +1/+1 for each name
// sticker that begins with it.
func TestStkBlankRocketshipChoosesALetterWhenItAttacks(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	id := p2Push(g, me.ID, "_____ _____ Rocketship", "Artifact Creature — Vehicle", blankRocketshipOracle, 2, 3)
	stkGive(t, g, me, id, "“Ringmaster”")
	stkGive(t, g, me, id, "“Night”")
	stkAttackWith(t, g, id)
	which := latestOptionPickFor(g, me.ID)
	if which == nil {
		t.Fatal("the attack trigger asked for no letter")
	}
	if len(which.PickOptions) != 26 {
		t.Errorf("%d letters offered, want 26", len(which.PickOptions))
	}
	if err := g.ResolveOptionPick(which.ID, me.ID, stkOption(t, which, "R")); err != nil {
		t.Fatalf("choose R: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := p2PT(g, id); p != 3 || tough != 4 {
		t.Errorf("P/T = %d/%d, want 3/4 (one name sticker begins with R)", p, tough)
	}
}

// A letter no sticker begins with is a legal choice and gives nothing.
func TestStkBlankRocketshipAnUnusedLetterGivesNothing(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	id := p2Push(g, me.ID, "_____ _____ Rocketship", "Artifact Creature — Vehicle", blankRocketshipOracle, 2, 3)
	stkGive(t, g, me, id, "“Ringmaster”")
	stkAttackWith(t, g, id)
	which := latestOptionPickFor(g, me.ID)
	if err := g.ResolveOptionPick(which.ID, me.ID, stkOption(t, which, "Z")); err != nil {
		t.Fatalf("choose Z: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := p2PT(g, id); p != 2 || tough != 3 {
		t.Errorf("P/T = %d/%d, want 2/3", p, tough)
	}
}

// The letter question is a restore point.
func TestStkBlankRocketshipLetterQuestionIsARestorePoint(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	id := p2Push(g, me.ID, "_____ _____ Rocketship", "Artifact Creature — Vehicle", blankRocketshipOracle, 2, 3)
	stkGive(t, g, me, id, "“Ringmaster”")
	stkAttackWith(t, g, id)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	which := latestOptionPickFor(g, me.ID)
	if which == nil {
		t.Fatal("the restored table lost the letter question")
	}
	if err := g.ResolveOptionPick(which.ID, me.ID, stkOption(t, which, "R")); err != nil {
		t.Fatalf("choose R: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := p2PT(g, id); p != 3 {
		t.Errorf("power = %d, want 3", p)
	}
}
