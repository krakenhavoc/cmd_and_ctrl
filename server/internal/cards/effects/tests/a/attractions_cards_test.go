package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attractions_cards_test.go pins the Attraction cards (ADR 0207,
// #2121): one test per card, each driving the printed event through the
// engine's open, roll-to-visit and junkyard.

const (
	downForRepairsOracle    = "455dc001-3f6d-4d44-afa7-0f83a2520baf"
	deadbeatAttendantOracle = "ae4b00ab-25ad-4177-8b86-af8b84d67ef3"
	pettingZookeeperOracle  = "ad6dd363-e1b7-4df4-b9b6-4f5c1fae3bc6"
	radRascalOracle         = "66d4a218-5404-4389-b495-eaeda9a34401"
	rideGuideOracle         = "8772e5cc-d3f3-4d68-86e7-6c105f39930a"
	seasonedButtoneerOracle = "6048a682-f218-4039-8d5e-3e5156817fb2"
	stepRightUpOracle       = "aaf6b3ed-2396-46d7-a32d-7d24cc2eb449"
	lineCutterOracle        = "a43d9793-b2ee-427c-95ca-285611aa1a4d"
	comingAttractionOracle  = "5d9c7ca4-f735-4bee-b03f-4c0c76ee48ed"
	draconianGateBotOracle  = "f73ba6ae-87c1-4e5d-ae52-59d11d7d6ba6"
	informationBoothOracle  = "2155f3d8-c876-42bc-b546-a476425a4133"
	fortuneTellerOracle     = "a837fe34-ae85-4bb6-af5a-4dec30b18f52"
	concessionStandOracle   = "2559fe2d-92b9-43c8-b538-d424df25eae8"
	rollerCoasterOracle     = "6efcf15a-2204-486f-9cff-7311054dcee0"
	kiddieCoasterOracle     = "c99ce798-e332-4672-8fba-6da74ab14582"
	spinnyRideOracle        = "cf2cc44e-8134-4b37-bc10-694abb8808f0"
)

// giveAttractionDeck gives `p` an Attraction deck holding `cards`, the
// last one on top.
func giveAttractionDeck(g *game.Game, p *game.Player, cards ...game.Card) {
	g.WithWriteLock(func() {
		p.AttractionDeck = &game.Zone{Kind: game.ZoneCommand, Owner: p.ID}
		for _, c := range cards {
			p.AttractionDeck.PushTop(c)
		}
	})
}

// fillerAttractions is n plain Attraction cards no catalog row runs.
func fillerAttractions(owner uuid.UUID, n int) []game.Card {
	out := make([]game.Card, n)
	for i := range out {
		out[i] = effectstest.AttractionCard(owner, "Filler Attraction", "", 1)
	}
	return out
}

// attractionsOpened counts EventAttractionOpened for `player` since
// event `from`.
func attractionsOpened(g *game.Game, player uuid.UUID, from int) []uuid.UUID {
	var out []uuid.UUID
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventAttractionOpened && ev.Actor == player {
			out = append(out, ev.CardID)
		}
	}
	return out
}

// --- the Employees -----------------------------------------------------

// "When this creature enters, open an Attraction." The top card of the
// Attraction deck enters under the creature's controller's control.
func TestEmployeesOpenAnAttractionOnEntry(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Deadbeat Attendant", "Creature — Vampire Employee", deadbeatAttendantOracle},
		{"Petting Zookeeper", "Creature — Elf Employee", pettingZookeeperOracle},
		{"Rad Rascal", "Creature — Devil Employee", radRascalOracle},
		{"Ride Guide", "Creature — Human Employee", rideGuideOracle},
		{"Seasoned Buttoneer", "Creature — Vedalken Employee", seasonedButtoneerOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			deck := fillerAttractions(me.ID, 3)
			top := deck[2].InstanceID
			giveAttractionDeck(g, me, deck...)
			before := len(g.Events)
			effectstest.CastAndResolveCreature(t, g, tc.name, tc.typeLine, tc.oracle)
			effectstest.PassPriorityAroundTable(t, g)
			opened := attractionsOpened(g, me.ID, before)
			if len(opened) != 1 || opened[0] != top {
				t.Fatalf("opened %v, want the top card %v", opened, top)
			}
			c := effectstest.FindBattlefieldCardForTest(g, top)
			if c == nil || c.Controller != me.ID {
				t.Fatalf("the opened Attraction is not on the battlefield under its owner: %+v", c)
			}
			if n := len(me.AttractionDeckCards()); n != 2 {
				t.Errorf("Attraction deck holds %d, want 2", n)
			}
		})
	}
}

// CR 701.51a: a player with no Attraction deck opens nothing, and the
// creature still resolves.
func TestDeadbeatAttendantWithNoAttractionDeckOpensNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := len(g.Events)
	id := effectstest.CastAndResolveCreature(t, g, "Deadbeat Attendant", "Creature — Vampire Employee", deadbeatAttendantOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, id) == nil {
		t.Fatal("Deadbeat Attendant did not resolve")
	}
	if opened := attractionsOpened(g, me.ID, before); len(opened) != 0 {
		t.Errorf("opened %v with no Attraction deck", opened)
	}
}

// --- the spells --------------------------------------------------------

// Step Right Up opens two, one after the other, each its own open.
func TestStepRightUpOpensTwoAttractions(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	deck := fillerAttractions(me.ID, 3)
	giveAttractionDeck(g, me, deck...)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	before := len(g.Events)
	effectstest.CastCatalogSpell(t, g, "Step Right Up", "Sorcery", stepRightUpOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	opened := attractionsOpened(g, me.ID, before)
	if len(opened) != 2 || opened[0] != deck[2].InstanceID || opened[1] != deck[1].InstanceID {
		t.Fatalf("opened %v, want the top two cards in order", opened)
	}
	if n := len(me.AttractionDeckCards()); n != 1 {
		t.Errorf("Attraction deck holds %d, want 1", n)
	}
}

// With one card left, the second open finds nothing (CR 701.51a).
func TestStepRightUpWithOneCardLeftOpensOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	giveAttractionDeck(g, me, fillerAttractions(me.ID, 1)...)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	before := len(g.Events)
	effectstest.CastCatalogSpell(t, g, "Step Right Up", "Sorcery", stepRightUpOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opened := attractionsOpened(g, me.ID, before); len(opened) != 1 {
		t.Fatalf("opened %d, want 1", len(opened))
	}
	if me.AttractionDeck.Size() != 0 {
		t.Errorf("Attraction deck holds %d, want 0", me.AttractionDeck.Size())
	}
}

// Coming Attraction searches, then opens: the open waits for the search
// to be answered (CR 608.2c).
func TestComingAttractionSearchesThenOpens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	giveAttractionDeck(g, me, fillerAttractions(me.ID, 2)...)
	forest := effectstest.FrPushTop(me, "Forest", "Basic Land — Forest", "")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	before := len(g.Events)
	effectstest.CastCatalogSpell(t, g, "Coming Attraction", "Sorcery", comingAttractionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if search == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if opened := attractionsOpened(g, me.ID, before); len(opened) != 0 {
		t.Fatalf("opened %v before the search was answered", opened)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("search: %v", err)
	}
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	land := effectstest.FindBattlefieldCardForTest(g, forest)
	if land == nil || !land.Tapped {
		t.Fatalf("the Forest is not on the battlefield tapped: %+v", land)
	}
	if opened := attractionsOpened(g, me.ID, before); len(opened) != 1 {
		t.Fatalf("opened %d after the search, want 1", len(opened))
	}
}

// Line Cutter's entry rolls to visit, and a visited Attraction's visit
// ability triggers.
func TestLineCutterRollsToVisitOnEntry(t *testing.T) {
	g := effectstest.InMainPhase(t)
	me := g.Seats[0]
	effectstest.PushAttraction(g, me.ID, "Information Booth", informationBoothOracle, effectstest.EveryNumber...)
	hand := me.Hand.Size()
	before := len(g.Events)
	effectstest.CastAndResolveCreature(t, g, "Line Cutter", "Creature — Human Rogue Guest", lineCutterOracle)
	effectstest.FrSettle(t, g)
	rolled := false
	for _, ev := range g.Events[before:] {
		if ev.Kind == game.EventRollToVisit && ev.Actor == me.ID {
			rolled = true
		}
	}
	if !rolled {
		t.Fatal("Line Cutter's entry did not roll to visit")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d: Information Booth was not visited", me.Hand.Size(), hand+1)
	}
}

// Draconian Gate-Bot's first bullet opens an Attraction.
func TestDraconianGateBotOpens(t *testing.T) {
	g := effectstest.InMainPhase(t)
	me := g.Seats[0]
	giveAttractionDeck(g, me, fillerAttractions(me.ID, 1)...)
	before := len(g.Events)
	effectstest.CastAndResolveCreature(t, g, "Draconian Gate-Bot", "Artifact Creature — Robot", draconianGateBotOracle)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode pick: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("mode: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opened := attractionsOpened(g, me.ID, before); len(opened) != 1 {
		t.Fatalf("opened %d, want 1", len(opened))
	}
}

// Its second bullet destroys an Attraction, which goes to its OWNER's
// junkyard (CR 717.6), not a graveyard.
func TestDraconianGateBotDestroysAnAttractionIntoItsOwnersJunkyard(t *testing.T) {
	g := effectstest.InMainPhase(t)
	me, them := g.Seats[0], g.Seats[1]
	ride := effectstest.PushAttraction(g, them.ID, "Spinny Ride", spinnyRideOracle, 2, 3, 6)
	effectstest.CastAndResolveCreature(t, g, "Draconian Gate-Bot", "Artifact Creature — Robot", draconianGateBotOracle)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode pick: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("mode: %v", err)
	}
	effectstest.PickCard(t, g, me.ID, ride)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ride) || them.Graveyard.Contains(ride) {
		t.Fatal("the Attraction was not put into a junkyard")
	}
	if them.Junkyard == nil || !them.Junkyard.Contains(ride) {
		t.Fatalf("the Attraction is not in its owner's junkyard: %v", them.JunkyardCards())
	}
}

// Down for Repairs: the revealed-hand pick, then up to one target
// Attraction that player controls is destroyed into their junkyard, and
// only after the pick is answered (CR 608.2c).
func TestDownForRepairsDiscardsThenDestroysTheirAttraction(t *testing.T) {
	g := effectstest.InMainPhase(t)
	me, them := g.Seats[0], g.Seats[1]
	ids := effectstest.RevealHand(them, effectstest.RhForest(), effectstest.RhBolt())
	booth := effectstest.PushAttraction(g, them.ID, "Information Booth", informationBoothOracle, 2, 6)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.CastCatalogSpell(t, g, "Down for Repairs", "Sorcery", downForRepairsOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: them.ID},
		{Kind: game.TargetCard, ID: booth, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenPick(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, ids[1:]) {
		t.Fatalf("options = %v, want only the Bolt", pick.DiscardOptions)
	}
	if !g.Battlefield.Contains(booth) {
		t.Fatal("the Attraction was destroyed before the pick was answered")
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if !them.Graveyard.Contains(ids[1]) {
		t.Error("the Bolt was not discarded")
	}
	if them.Junkyard == nil || !them.Junkyard.Contains(booth) {
		t.Fatalf("the Attraction is not in their junkyard (battlefield %v, graveyard %v)",
			g.Battlefield.Contains(booth), them.Graveyard.Contains(booth))
	}
}

// "Up to one": cast with no Attraction at all, it is Thoughtseize
// without the life loss.
func TestDownForRepairsWithNoAttraction(t *testing.T) {
	g := effectstest.InMainPhase(t)
	me, them := g.Seats[0], g.Seats[1]
	ids := effectstest.RevealHand(them, effectstest.RhBolt())
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.CastCatalogSpell(t, g, "Down for Repairs", "Sorcery", downForRepairsOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: them.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenPick(t, g)
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	g.SettleResolution()
	if !them.Graveyard.Contains(ids[0]) {
		t.Error("the Bolt was not discarded")
	}
}

// The Attraction clause is related to the player clause: an Attraction
// a different player controls is refused at announcement (CR 601.2c).
func TestDownForRepairsRefusesAnotherPlayersAttraction(t *testing.T) {
	g := effectstest.InMainPhase(t)
	them, other := g.Seats[1], g.Seats[2]
	effectstest.RevealHand(them, effectstest.RhBolt())
	elsewhere := effectstest.PushAttraction(g, other.ID, "Information Booth", informationBoothOracle, 2, 6)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	err := effectstest.CastCatalogSpellErr(t, g, "Down for Repairs", "Sorcery", downForRepairsOracle, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: them.ID},
		{Kind: game.TargetCard, ID: elsewhere, Slot: 1},
	})
	if err == nil {
		t.Fatal("an Attraction the targeted opponent does not control was accepted")
	}
}

// --- the Attractions ---------------------------------------------------

// Each Attraction's visit does what it prints, and only when its own
// lights match: one with no lit number never triggers.
func TestAttractionVisits(t *testing.T) {
	t.Run("Information Booth draws", func(t *testing.T) {
		g := effectstest.InMainPhase(t)
		me := g.Seats[0]
		effectstest.PushAttraction(g, me.ID, "Information Booth", informationBoothOracle, effectstest.EveryNumber...)
		hand := me.Hand.Size()
		effectstest.RollToVisit(t, g, me.ID)
		effectstest.FrSettle(t, g)
		if me.Hand.Size() != hand+1 {
			t.Errorf("hand = %d, want %d", me.Hand.Size(), hand+1)
		}
	})
	t.Run("Fortune Teller scries 1", func(t *testing.T) {
		g := effectstest.InMainPhase(t)
		me := g.Seats[0]
		effectstest.PushAttraction(g, me.ID, "Fortune Teller", fortuneTellerOracle, effectstest.EveryNumber...)
		effectstest.RollToVisit(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		sc := effectstest.ScryChoiceFor(g, me.ID)
		if sc == nil || len(sc.ScryCards) != 1 {
			t.Fatalf("no scry 1: %+v", sc)
		}
	})
	t.Run("Concession Stand makes a Food", func(t *testing.T) {
		g := effectstest.InMainPhase(t)
		me := g.Seats[0]
		effectstest.PushAttraction(g, me.ID, "Concession Stand", concessionStandOracle, effectstest.EveryNumber...)
		effectstest.RollToVisit(t, g, me.ID)
		effectstest.FrSettle(t, g)
		if n := effectstest.FrTokens(g, me.ID, "Food"); n != 1 {
			t.Errorf("Food tokens = %d, want 1", n)
		}
	})
	for _, tc := range []struct {
		name, oracle string
		power        int
	}{
		{"Roller Coaster", rollerCoasterOracle, 2},
		{"Kiddie Coaster", kiddieCoasterOracle, 1},
	} {
		t.Run(tc.name+" pumps your team", func(t *testing.T) {
			g := effectstest.InMainPhase(t)
			me, them := g.Seats[0], g.Seats[1]
			mine := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
			theirs := effectstest.PushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
			effectstest.PushAttraction(g, me.ID, tc.name, tc.oracle, effectstest.EveryNumber...)
			effectstest.RollToVisit(t, g, me.ID)
			effectstest.FrSettle(t, g)
			if p := effectstest.FindBattlefieldCardForTest(g, mine).CurrentPower(); p != 2+tc.power {
				t.Errorf("your creature's power = %d, want %d", p, 2+tc.power)
			}
			if p := effectstest.FindBattlefieldCardForTest(g, theirs).CurrentPower(); p != 2 {
				t.Errorf("an opponent's creature was pumped to %d", p)
			}
		})
	}
	t.Run("Spinny Ride taps an opponent's creature", func(t *testing.T) {
		g := effectstest.InMainPhase(t)
		me, them := g.Seats[0], g.Seats[1]
		theirs := effectstest.PushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
		effectstest.PushAttraction(g, me.ID, "Spinny Ride", spinnyRideOracle, effectstest.EveryNumber...)
		effectstest.RollToVisit(t, g, me.ID)
		effectstest.FrSettle(t, g, effectstest.FrCard(theirs))
		if c := effectstest.FindBattlefieldCardForTest(g, theirs); c == nil || !c.Tapped {
			t.Fatalf("the opponent's creature is not tapped: %+v", c)
		}
	})
	t.Run("an unlit number visits nothing", func(t *testing.T) {
		g := effectstest.InMainPhase(t)
		me := g.Seats[0]
		effectstest.PushAttraction(g, me.ID, "Information Booth", informationBoothOracle)
		hand := me.Hand.Size()
		effectstest.RollToVisit(t, g, me.ID)
		effectstest.FrSettle(t, g)
		if me.Hand.Size() != hand {
			t.Errorf("an Attraction with nothing lit was visited")
		}
	})
	t.Run("an opponent's roll does not visit yours", func(t *testing.T) {
		g := effectstest.InMainPhase(t)
		me, them := g.Seats[0], g.Seats[1]
		effectstest.PushAttraction(g, me.ID, "Information Booth", informationBoothOracle, effectstest.EveryNumber...)
		hand := me.Hand.Size()
		effectstest.RollToVisit(t, g, them.ID)
		effectstest.FrSettle(t, g)
		if me.Hand.Size() != hand {
			t.Errorf("an opponent's roll visited your Attraction")
		}
	})
}

// CR 717.4 / 703.4g: as your precombat main phase begins you roll to
// visit, and the visit ability triggers with no card telling you to.
func TestAttractionIsVisitedAsThePrecombatMainPhaseBegins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushAttraction(g, me.ID, "Information Booth", informationBoothOracle, effectstest.EveryNumber...)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.FrSettle(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d: no visit at the start of the main phase", me.Hand.Size(), hand+1)
	}
}
