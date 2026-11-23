package testsst

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardse "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/e"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_y_restore_test.go — #4253's last part: the continuations that
// parked behind a CR 614 / 616 prompt on a MOVE (an exile, a sacrifice, a
// tuck, a return, a move to hand) are keyed on the move itself. Each test
// pauses the move on a CR 616 ordering prompt from two catalog probe
// replacements over a card named Needle (or an Obstinate Baloth entering),
// takes a STRICT restore (the boot path), answers on the restored table,
// and checks the rest of the card ran off the disk.

// strictRestoreAtOrderPrompt is restoreRoundTrip(strict) that also checks
// an open CR 616 ordering prompt was written with its resume frame.
func strictRestoreAtOrderPrompt(t *testing.T, g *game.Game) *game.Game {
	t.Helper()
	snap := g.CaptureSnapshot()
	found := false
	for _, c := range snap.PendingChoices {
		if strings.HasPrefix(c.PickThen, "replacement-resume/") {
			found = true
		}
	}
	if !found {
		t.Fatalf("no prompt in the snapshot names a replacement resume frame: %+v", snap.Continuations)
	}
	return effectstest.RestoreRoundTrip(t, g, true)
}

// --- Palace Jailer: the event-conditioned ExileUntil -------------------

// Palace Jailer over an opponent's Needle: the exile asks for the order,
// the restored answer exiles it with its "until an opponent becomes the
// monarch" record, and an opponent taking the crown brings it back.
func TestSweepYPalaceJailerExileIsARestorePointAndStillReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, owner, thief := g.Seats[0], g.Seats[1], g.Seats[2]
	victim := effectstest.PushVanillaCreature(g, owner.ID, "Needle", 2, 2)
	effectstest.SeedExileOrderProbes(g, me.ID)
	effectstest.CastAndResolveCreature(t, g, "Palace Jailer", "Creature — Human Soldier", effectstest.PalaceJailerOracle)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.MonSettle(t, g)

	g = strictRestoreAtOrderPrompt(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Fatal("the Needle left the battlefield before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	effectstest.MonSettle(t, g)
	if !effectstest.InExile(g, victim) {
		t.Fatal("the Needle was not exiled after the restored answer")
	}
	effectstest.MonCrown(t, g, thief.ID)
	effectstest.MonSettle(t, g)
	if !g.Battlefield.Contains(victim) && !effectstest.OnBattlefieldByName(g, "Needle") {
		t.Fatal("an opponent became the monarch and the Needle did not return: the record did not ride the tail")
	}
}

// --- Memory Vessel: the each-player impulse exile chain ----------------

// Memory Vessel with a Needle on the second library in APNAP order: the
// first library is exiled, the Needle's exile asks for the order, and the
// restored answer finishes the chain and bans every hand (the chain's done,
// a registered tail).
func TestSweepYMemoryVesselChainIsARestorePointAndStillBansTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	needle := effectstest.PushLibraryTopForTest(opp, "Needle")
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Trinket", TypeLine: "Artifact", ManaCost: "{0}",
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	effectstest.SeedExileOrderProbes(g, me.ID)
	vessel := effectstest.PushCatalogPermanent(g, me.ID, "Memory Vessel", "Artifact", effectstest.MemoryVesselOracle, false)
	myTop := me.Library.Cards[len(me.Library.Cards)-1].InstanceID
	if err := g.ActivateCatalogAbility(me.ID, vessel, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Memory Vessel: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, myTop) {
		t.Fatal("the active player's library was not exiled from before the Needle's prompt")
	}

	g = strictRestoreAtOrderPrompt(t, g)
	me, opp = g.Seats[0], g.Seats[1]
	if !opp.Library.Contains(needle) {
		t.Fatal("the Needle left the library before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, needle) {
		t.Fatal("the Needle was not exiled after the restored answer")
	}
	if effectstest.PermissionFor(g, opp.ID, needle) == nil {
		t.Error("the Needle's owner may not play it: the grant did not ride the chain")
	}
	var cant *game.CantCastError
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); !errors.As(err, &cant) {
		t.Fatalf("a spell cast from hand after the restored chain: err = %v, want the hand ban", err)
	}
}

// --- Scapeshift: a sacrificing permanent pick --------------------------

// Scapeshift over a land and a Needle: the Needle's sacrifice asks for the
// order, and the restored answer finishes the sacrifice and asks the search
// for that many (the pick's body, run by its key off the disk).
func TestSweepYScapeshiftSacrificeIsARestorePointAndStillSearches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var lands []uuid.UUID
	g.WithWriteLock(func() {
		for _, name := range []string{"Scape Forest", "Needle"} {
			id := uuid.New()
			g.Battlefield.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
			lands = append(lands, id)
		}
	})
	effectstest.SeedTypedLibrary(me, "Basic Land — Mountain", "Fetch One", "Fetch Two")
	effectstest.SeedOrderProbes(g, me.ID, effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")
	effectstest.CastCatalogSpell(t, g, "Scapeshift", "Sorcery", effectstest.ScapeshiftOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	sac := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if sac == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(sac.ID, me.ID, lands); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}

	g = strictRestoreAtOrderPrompt(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	if effectstest.SearchPromptFor(g, me.ID) != nil {
		t.Fatal("the search opened before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	if !me.Graveyard.Contains(lands[0]) || !me.Graveyard.Contains(lands[1]) {
		t.Fatal("the lands were not sacrificed after the restored answer")
	}
	search := effectstest.SearchPromptFor(g, me.ID)
	if search == nil {
		t.Fatal("the search did not open after the restored answer: the pick's body did not run")
	}
	if search.SearchMax != 2 {
		t.Errorf("up to THAT MANY: ceiling %d, want 2", search.SearchMax)
	}
}

// A "may sacrifice" whose run carried two numbers: they ride the
// sacrifice's tail (Nums), so the sacrifice's ordering prompt is a restore
// point and the "if you do" is handed both off the disk.
func TestSweepYMaySacrificeWithTwoNumbersIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	needle := effectstest.PushVanillaCreature(g, me.ID, "Needle", 1, 1)
	source := effectstest.PushVanillaCreature(g, me.ID, "Asker", 1, 1)
	effectstest.SeedOrderProbes(g, me.ID, effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")
	effectstest.TestSacrificedNums = nil
	g.WithWriteLock(func() {
		item := &game.StackItem{Controller: me.ID, SourceCardID: source}
		if err := effects.MaySacrificeOne(g, item, "sacrifice a creature?", effects.Creature(), true, effectstest.TestSacrificedTwoNums, 3, 4); err != nil {
			t.Fatalf("maySacrificeOne: %v", err)
		}
	})
	sac := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if sac == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(sac.ID, me.ID, []uuid.UUID{needle}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}

	g = strictRestoreAtOrderPrompt(t, g)
	effectstest.AnswerExileOrder(t, g)
	if !g.Seats[0].Graveyard.Contains(needle) {
		t.Fatal("the Needle was not sacrificed after the restored answer")
	}
	if !slices.Equal(effectstest.TestSacrificedNums, []int{3, 4}) {
		t.Errorf("the \"if you do\" was handed %v, want [3 4]", effectstest.TestSacrificedNums)
	}
}

// --- Emergent Ultimatum: the chosen card's tuck ------------------------

// The opponent's pick tucked: the Needle's tuck asks for the order, and the
// restored answer shuffles it in, grants the free cast over the card left
// and exiles the spell.
func TestSweepYEmergentUltimatumTuckIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	var needle, left, spell uuid.UUID
	g.WithWriteLock(func() {
		needle, left, spell = uuid.New(), uuid.New(), uuid.New()
		g.Exile.PushTop(game.Card{InstanceID: needle, Name: "Needle", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
		g.Exile.PushTop(game.Card{InstanceID: left, Name: "Left", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: cardse.EmergentUltimatumName, TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})
	effectstest.SeedOrderProbes(g, me.ID, effectstest.LibraryHandOrderProbeOracles, "Library Hand Order Probe")
	g.WithWriteLock(func() {
		if err := cardse.EmergentUltimatumSettle(effects.KeyedContext(g, me.ID, spell), []uuid.UUID{needle}, []uuid.UUID{left}); err != nil {
			t.Fatalf("emergentUltimatumSettle: %v", err)
		}
	})

	g = strictRestoreAtOrderPrompt(t, g)
	me = g.Seats[0]
	if !effectstest.InExile(g, needle) || effectstest.InExile(g, spell) {
		t.Fatal("the tuck or the spell's exile happened before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	if !me.Library.Contains(needle) {
		t.Fatal("the chosen card was not shuffled into the library after the restored answer")
	}
	if effectstest.PermissionFor(g, me.ID, left) == nil {
		t.Error("the card left was not granted its free cast off the tail's carry")
	}
	if !effectstest.InExile(g, spell) {
		t.Error("the spell did not exile itself")
	}
}

// --- Last Voyage of the _____: the return and the attach ---------------

// Last Voyage returns an Obstinate Baloth: the entry asks for the order,
// and the restored answer returns it and attaches the Aura.
func TestSweepYLastVoyageReturnIsARestorePointAndStillAttaches(t *testing.T) {
	g, me := effectstest.StkTableWith(t, 0, "Night Brushwagg Ringmaster")
	dead := stkPlace(g, me.ID, me.Graveyard, "Obstinate Baloth", "Creature — Beast", 4, 4)
	effectstest.SeedOrderProbes(g, me.ID, effectstest.EntryOrderProbeOracles, "Entry Order Probe")
	id := effectstest.CastAndResolveCreature(t, g, "Last Voyage of the _____", "Enchantment", lastVoyageOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.StkPutOn(t, g, me, id, "“Night”", 0)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no question asks which creature card returns")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{dead}); err != nil {
		t.Fatalf("choose: %v", err)
	}

	g = strictRestoreAtOrderPrompt(t, g)
	if effects.OnBattlefield(g, dead) {
		t.Fatal("the Baloth entered before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, dead) {
		t.Fatal("the creature card did not return after the restored answer")
	}
	if !effectstest.AttachedToNow(t, g, id, dead) {
		t.Error("the Aura is not attached to the returned creature: the tail did not run off the disk")
	}
}

// --- Demonic Consultation: the named card's move to hand ---------------

// Demonic Consultation naming Needle: the found Needle's move to hand asks
// for the order, and the restored answer puts it in hand and exiles the
// other revealed card (the tail's carry).
func TestSweepYDemonicConsultationTakeIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	after := effectstest.SweepFTop(me, "After", "Sorcery")
	wanted := effectstest.SweepFTop(me, "Needle", "Sorcery")
	other := effectstest.SweepFTop(me, "Other", "Sorcery")
	for i := 0; i < 6; i++ {
		effectstest.SweepFTop(me, "Junk", "Sorcery")
	}
	effectstest.SeedOrderProbes(g, me.ID, effectstest.LibraryHandOrderProbeOracles, "Library Hand Order Probe")
	effectstest.CastCatalogSpell(t, g, "Demonic Consultation", "Instant", effectstest.SweepFConsultationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "Needle")
	effectstest.PassPriorityAroundTable(t, g)

	g = strictRestoreAtOrderPrompt(t, g)
	me = g.Seats[g.Turn.ActiveSeat]
	if effectstest.InExile(g, other) {
		t.Fatal("the other revealed card was exiled before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, wanted) {
		t.Fatal("the named card is not in hand after the restored answer")
	}
	if !effectstest.InExile(g, other) {
		t.Error("the other revealed card was not exiled: the tail's carry did not ride the disk")
	}
	if !me.Library.Contains(after) {
		t.Error("a card below the named one left the library")
	}
}

// --- a registered take: Silundi Vision ---------------------------------

// Silundi Vision takes a Needle: the move to hand asks for the order, and
// the restored answer puts it in hand and bottoms the rest (the take's own
// Then, run by its key).
func TestSweepYSilundiVisionTakeIsARestorePointAndStillBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := effectstest.ZnrStackLibrary(me,
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Needle", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	seventh := me.Library.Cards[len(me.Library.Cards)-7].InstanceID
	effectstest.SeedOrderProbes(g, me.ID, effectstest.LibraryHandOrderProbeOracles, "Library Hand Order Probe")
	effectstest.CastCatalogSpell(t, g, "Silundi Vision", "Instant", effectstest.SilundiVisionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChooseCards(g, me.ID)
	if prompt == nil {
		t.Fatal("no choose-cards prompt")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	g = strictRestoreAtOrderPrompt(t, g)
	me = g.Seats[0]
	if me.Hand.Contains(top[1]) {
		t.Fatal("the Needle reached the hand before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	if !me.Hand.Contains(top[1]) {
		t.Fatal("the Needle is not in hand after the restored answer")
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != seventh {
		t.Error("the rest were not bottomed: the take's Then did not run off the disk")
	}
}

// --- a fan-out of takes: Explore the Vastlands -------------------------

// Explore the Vastlands with a Needle in the second player's take: the
// first player's take finishes, the Needle's move asks for the order, and
// the restored answer finishes every take and gains each player 3 life
// (the fan-out's rest, run once off the disk).
func TestSweepYExploreTheVastlandsTakesAreARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seats := effectstest.ApnapSeats(g)
	tops := make([][]uuid.UUID, len(seats))
	tops[0] = effectstest.StackTop(seats[0], effectstest.VastFive)
	tops[1] = effectstest.StackTop(seats[1], []effectstest.VastCard{
		{Name: "Forest", TypeLine: "Basic Land — Forest"}, {Name: "Needle", TypeLine: "Instant"}, {Name: "Grizzly Bears", TypeLine: "Creature — Bear"},
		{Name: "Bear B", TypeLine: "Creature — Bear"}, {Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid"},
	})
	life := make([]int, len(seats))
	for i, p := range seats {
		life[i] = p.Life
	}
	effectstest.SeedOrderProbes(g, seats[0].ID, effectstest.LibraryHandOrderProbeOracles, "Library Hand Order Probe")
	effectstest.CastExplore(t, g)
	for i := range seats[:2] {
		c := effectstest.LatestChooseCards(g, seats[i].ID)
		if c == nil {
			t.Fatalf("seat %d was not asked", i)
		}
		if err := g.ResolveChooseCards(c.ID, seats[i].ID, []uuid.UUID{tops[i][1]}); err != nil {
			t.Fatalf("seat %d answer: %v", i, err)
		}
	}
	if !seats[0].Hand.Contains(tops[0][1]) {
		t.Fatal("the first player's take did not finish before the Needle's prompt")
	}

	g = strictRestoreAtOrderPrompt(t, g)
	seats = effectstest.ApnapSeats(g)
	if seats[1].Hand.Contains(tops[1][1]) {
		t.Fatal("the Needle reached the hand before the restored answer")
	}
	effectstest.AnswerExileOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !seats[1].Hand.Contains(tops[1][1]) {
		t.Fatal("the Needle is not in hand after the restored answer")
	}
	rest := effectstest.Without(tops[1], tops[1][1])
	if !effectstest.SameSet(effectstest.LibraryBottom(seats[1], len(rest)), rest) {
		t.Error("the second player's rest is not on the bottom: their own Then did not run off the disk")
	}
	for i, p := range seats {
		if p.Life != life[i]+3 {
			t.Errorf("seat %d life %d, want %d: the fan-out's rest did not run once", i, p.Life, life[i]+3)
		}
	}
}

// --- Underrealm Lich: the replaced draw's take -------------------------

// Underrealm Lich takes a Needle: the move to hand asks for the order, and
// the restored answer puts it in hand, the rest of the look in the
// graveyard, and pays the rest of the draw (the parked draw on the carry).
func TestSweepYUnderrealmLichTakeIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	p := g.Seats[1]
	effectstest.BfCardWithOracle(g, p, "Underrealm Lich", "Creature — Zombie Elf Shaman", effectstest.UnderrealmLichOracle)
	var needle uuid.UUID
	g.WithWriteLock(func() { needle = effectstest.PushLibraryTopForTest(p, "Needle") })
	effectstest.SeedOrderProbes(g, p.ID, effectstest.LibraryHandOrderProbeOracles, "Library Hand Order Probe")
	gyBefore := p.Graveyard.Size()
	effectstest.StartLoot(g, p, 1, 1)
	pick := effectstest.ChooseCardsChoiceFor(g, p.ID)
	if pick == nil {
		t.Fatal("the Lich did not ask")
	}
	effectstest.AnswerChooseCards(t, g, p.ID, needle)

	g = strictRestoreAtOrderPrompt(t, g)
	p = g.Seats[1]
	if handDiscardFor(g, p) != nil {
		t.Fatal("the restored table owed the discard before the take finished")
	}
	effectstest.AnswerExileOrder(t, g)
	if !effectstest.ZoneHas(p.Hand, needle) {
		t.Fatal("the Needle is not in hand after the restored answer")
	}
	if p.Graveyard.Size() != gyBefore+2 {
		t.Errorf("graveyard %d, want %d: the rest of the look did not go there", p.Graveyard.Size(), gyBefore+2)
	}
	if handDiscardFor(g, p) == nil {
		t.Error("the discard the instruction owed did not open: the parked draw was not paid off the carry")
	}
	if parkedDrawsOf(g) != 0 {
		t.Errorf("%d draw tails left parked", parkedDrawsOf(g))
	}
}
