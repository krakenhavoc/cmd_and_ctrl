package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b17_cards_test.go — Kamigawa: Neon Dynasty's simple transforming
// Sagas (#5493, S82), each driven through the engine's lore-counter road.

func neoB17Row(oracleID, front, cost, back, backType, p, t, colour string) cards.Card {
	row := transformRow(oracleID, front, "Enchantment — Saga", cost, back, backType, p, t, []string{colour})
	row.Keywords = []string{"Transform"}
	return row
}

// neoB17Chapter moves to the next precombat main, answers one target prompt
// when the chapter asks for it, and lets the chapter resolve.
func neoB17Chapter(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	neoB17ToMain(t, g, me)
	if target != uuid.Nil && latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, target)
	}
	passPriorityAroundTable(t, g)
}

// neoB17ToMain advances to the next precombat main of `me`'s turn.
func neoB17ToMain(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for i := range g.Seats {
		if g.Seats[i].ID == me.ID {
			advanceToPrecombatMainOf(t, g, i)
			return
		}
	}
	t.Fatal("seat not found")
}

// neoB17Back is the transformed permanent named `name` after chapter III.
func neoB17Back(t *testing.T, g *game.Game, saga uuid.UUID, me *game.Player, name string) *game.Card {
	t.Helper()
	if battlefieldCardFor(g, saga) != nil {
		t.Fatal("the original Saga object is still on the battlefield")
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == name {
			if c.ActiveFace != 1 || c.Controller != me.ID || c.Counters[game.CounterLore] != 0 {
				t.Errorf("%s: face %d, controller %s, %d lore counters", name, c.ActiveFace, c.Controller, c.Counters[game.CounterLore])
			}
			if !c.IsCreature() {
				t.Errorf("%s is not a creature", name)
			}
			return c
		}
	}
	t.Fatalf("%s never reached the battlefield", name)
	return nil
}

func neoB17Check(t *testing.T, c *game.Card, power, toughness int, kws ...string) {
	t.Helper()
	if c.CurrentPower() != power || c.CurrentToughness() != toughness {
		t.Errorf("%s is %d/%d, want %d/%d", c.Name, c.CurrentPower(), c.CurrentToughness(), power, toughness)
	}
	for _, k := range kws {
		if !game.HasKeyword(c, k) {
			t.Errorf("%s lacks %s", c.Name, k)
		}
	}
}

func TestEraOfEnlightenment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := importAndCast(t, g, neoB17Row(eraOfEnlightenmentOracleID, "Era of Enlightenment", "{1}{W}", "Hand of Enlightenment", "Enchantment Creature — Human Monk", "2", "2", "W"), me)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("chapter I raised no scry prompt")
	} else if n := len(scryChoiceFor(g, me.ID).ScryCards); n != 2 {
		t.Errorf("scry looked at %d cards, want 2", n)
	}
	answerScryKeepAll(t, g, me.ID)
	life := me.Life
	neoB17Chapter(t, g, me, uuid.Nil)
	if me.Life != life+2 {
		t.Errorf("chapter II: life %d, want %d", me.Life, life+2)
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	neoB17Check(t, neoB17Back(t, g, saga, me, "Hand of Enlightenment"), 2, 2, "first strike")
	assertFaceInvariant(t, g)
}

func TestTheModernAge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	handCardForTest(me, "Spare Forest", "Basic Land — Forest", "")
	saga := importAndCast(t, g, neoB17Row(theModernAgeOracleID, "The Modern Age", "{1}{U}", "Vector Glider", "Enchantment Creature — Spirit", "2", "3", "U"), me)
	hand := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("chapter I raised no discard prompt after the draw")
	}
	discardFromHand(t, g, me.ID)
	if len(me.Hand.Cards) != hand {
		t.Errorf("hand after chapter I = %d, want %d (drew one, discarded one)", len(me.Hand.Cards), hand)
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("chapter II raised no discard prompt")
	}
	discardFromHand(t, g, me.ID)
	neoB17Chapter(t, g, me, uuid.Nil)
	neoB17Check(t, neoB17Back(t, g, saga, me, "Vector Glider"), 2, 3, "flying")
}

func TestBefriendingTheMoths(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	saga := importAndCast(t, g, neoB17Row(befriendingTheMothsOracleID, "Befriending the Moths", "{3}{W}", "Imperial Moth", "Enchantment Creature — Insect", "2", "4", "W"), me)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, bear)
		passPriorityAroundTable(t, g)
	}
	c := battlefieldCardFor(g, bear)
	neoB17Check(t, c, 3, 3, "flying")
	neoB17Chapter(t, g, me, bear)
	// Until end of turn only: the first chapter's boost ended with its turn,
	// and the second adds its own.
	c = battlefieldCardFor(g, bear)
	neoB17Check(t, c, 3, 3, "flying")
	neoB17Chapter(t, g, me, uuid.Nil)
	neoB17Check(t, neoB17Back(t, g, saga, me, "Imperial Moth"), 2, 4, "flying")
}

func TestTalesOfMasterSeshiro(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	van := apaPush(g, me.ID, me.ID, game.Card{Name: "Parked Van", TypeLine: "Artifact — Vehicle", Power: 3, Toughness: 3})
	saga := importAndCast(t, g, neoB17Row(talesOfMasterSeshiroOracleID, "Tales of Master Seshiro", "{4}{G}", "Seshiro's Living Legacy", "Enchantment Creature — Snake Warrior", "5", "5", "G"), me)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, bear)
		passPriorityAroundTable(t, g)
	}
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("chapter I: %d counters on the creature, want 1", got)
	}
	if !game.HasKeyword(battlefieldCardFor(g, bear), "vigilance") {
		t.Error("the creature did not gain vigilance")
	}
	neoB17Chapter(t, g, me, van)
	if got := counterCount(g, van, game.CounterPlusOne); got != 1 {
		t.Errorf("chapter II: %d counters on the Vehicle, want 1", got)
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	neoB17Check(t, neoB17Back(t, g, saga, me, "Seshiro's Living Legacy"), 5, 5, "vigilance", "haste")
}

func TestTheShatteredStatesEra(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	mine := apaPush(g, me.ID, me.ID, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3, Tapped: true})
	saga := importAndCast(t, g, neoB17Row(theShatteredStatesEraOracleID, "The Shattered States Era", "{4}{R}", "Nameless Conqueror", "Enchantment Creature — Human Samurai", "3", "3", "R"), me)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, theirs)
		passPriorityAroundTable(t, g)
	}
	stolen := battlefieldCardFor(g, theirs)
	if stolen.Controller != me.ID || stolen.Tapped || !game.HasKeyword(stolen, "haste") {
		t.Errorf("chapter I: controller mine=%v tapped=%v haste=%v", stolen.Controller == me.ID, stolen.Tapped, game.HasKeyword(stolen, "haste"))
	}
	advancePastCleanupForTest(t, g)
	if battlefieldCardFor(g, theirs).Controller != opp.ID {
		t.Error("control did not revert at end of turn")
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	if p := battlefieldCardFor(g, mine).CurrentPower(); p != 3 {
		t.Errorf("chapter II: my creature has power %d, want 3", p)
	}
	if p := battlefieldCardFor(g, theirs).CurrentPower(); p != 3 {
		t.Errorf("chapter II touched their creature: power %d, want 3", p)
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	neoB17Check(t, neoB17Back(t, g, saga, me, "Nameless Conqueror"), 3, 3, "trample", "haste")
}

func TestOkibaReckonerRaid(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	van := apaPush(g, me.ID, me.ID, game.Card{Name: "Parked Van", TypeLine: "Artifact — Vehicle", Power: 3, Toughness: 3})
	theirVan := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Van", TypeLine: "Artifact — Vehicle", Power: 3, Toughness: 3})
	saga := importAndCast(t, g, neoB17Row(okibaReckonerRaidOracleID, "Okiba Reckoner Raid", "{B}", "Nezumi Road Captain", "Enchantment Creature — Rat Rogue", "2", "2", "B"), me)
	myLife, theirLife := me.Life, opp.Life
	passPriorityAroundTable(t, g)
	if me.Life != myLife+1 || opp.Life != theirLife-1 {
		t.Errorf("chapter I: life %d/%d, want %d/%d", me.Life, opp.Life, myLife+1, theirLife-1)
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	if me.Life != myLife+2 || opp.Life != theirLife-2 {
		t.Errorf("chapter II: life %d/%d, want %d/%d", me.Life, opp.Life, myLife+2, theirLife-2)
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	neoB17Check(t, neoB17Back(t, g, saga, me, "Nezumi Road Captain"), 2, 2, "menace")
	if !game.HasKeyword(battlefieldCardFor(g, van), "menace") {
		t.Error("my Vehicle lacks menace")
	}
	if game.HasKeyword(battlefieldCardFor(g, theirVan), "menace") {
		t.Error("their Vehicle has menace")
	}
}

func TestLifeOfToshiroUmezawa(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	saga := importAndCast(t, g, neoB17Row(lifeOfToshiroUmezawaOracleID, "Life of Toshiro Umezawa", "{1}{B}", "Memory of Toshiro", "Enchantment Creature — Human Samurai", "2", "3", "B"), me)
	passPriorityAroundTable(t, g)
	mode := modePickChoiceFor(g, me.ID)
	if mode == nil {
		t.Fatal("chapter I raised no mode prompt")
	}
	answerMode(t, g, mode, me.ID, 0)
	if latestPickTarget(g, me.ID) != nil {
		pickTriggerTarget(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)
	neoB17Check(t, battlefieldCardFor(g, bear), 4, 4)

	life := me.Life
	neoB17ToMain(t, g, me)
	mode = modePickChoiceFor(g, me.ID)
	if mode == nil {
		t.Fatal("chapter II raised no mode prompt")
	}
	answerMode(t, g, mode, me.ID, 2)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("gain-life mode: life %d, want %d", me.Life, life+2)
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	back := neoB17Back(t, g, saga, me, "Memory of Toshiro")
	neoB17Check(t, back, 2, 3)
	if len(game.ManaAbilitiesForCard(*back)) == 0 {
		t.Error("Memory of Toshiro has no mana ability")
	}
}

func TestAzusasManyJourneys(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := importAndCast(t, g, neoB17Row(azusasManyJourneysOracleID, "Azusa's Many Journeys", "{1}{G}", "Likeness of the Seeker", "Enchantment Creature — Human Monk", "3", "3", "G"), me)
	before := g.ExtraLandDropsThisTurn[me.ID]
	passPriorityAroundTable(t, g)
	if got := g.ExtraLandDropsThisTurn[me.ID]; got != before+1 {
		t.Errorf("chapter I: land plays %d, want %d", got, before+1)
	}
	life := me.Life
	neoB17Chapter(t, g, me, uuid.Nil)
	if me.Life != life+3 {
		t.Errorf("chapter II: life %d, want %d", me.Life, life+3)
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	neoB17Check(t, neoB17Back(t, g, saga, me, "Likeness of the Seeker"), 3, 3)
}
