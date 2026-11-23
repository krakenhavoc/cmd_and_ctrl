package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsap "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ap"
	cardsb "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/b"
	cardse "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/e"
	cardskr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/kr"
	cardso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/o"
	cardssy "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/sy"
	cardsth "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/th"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b17_cards_test.go — Kamigawa: Neon Dynasty's simple transforming
// Sagas (#5493, S82), each driven through the engine's lore-counter road.

func neoB17Row(oracleID, front, cost, back, backType, p, t, colour string) cards.Card {
	row := effectstest.TransformRow(oracleID, front, "Enchantment — Saga", cost, back, backType, p, t, []string{colour})
	row.Keywords = []string{"Transform"}
	return row
}

// neoB17Chapter moves to the next precombat main, answers one target prompt
// when the chapter asks for it, and lets the chapter resolve.
func neoB17Chapter(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	neoB17ToMain(t, g, me)
	if target != uuid.Nil && effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, target)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// neoB17ToMain advances to the next precombat main of `me`'s turn.
func neoB17ToMain(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for i := range g.Seats {
		if g.Seats[i].ID == me.ID {
			effectstest.AdvanceToPrecombatMainOf(t, g, i)
			return
		}
	}
	t.Fatal("seat not found")
}

// neoB17Back is the transformed permanent named `name` after chapter III.
func neoB17Back(t *testing.T, g *game.Game, saga uuid.UUID, me *game.Player, name string) *game.Card {
	t.Helper()
	if effectstest.BattlefieldCardFor(g, saga) != nil {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := effectstest.ImportAndCast(t, g, neoB17Row(cardse.EraOfEnlightenmentOracleID, "Era of Enlightenment", "{1}{W}", "Hand of Enlightenment", "Enchantment Creature — Human Monk", "2", "2", "W"), me)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("chapter I raised no scry prompt")
	} else if n := len(effectstest.ScryChoiceFor(g, me.ID).ScryCards); n != 2 {
		t.Errorf("scry looked at %d cards, want 2", n)
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	life := me.Life
	neoB17Chapter(t, g, me, uuid.Nil)
	if me.Life != life+2 {
		t.Errorf("chapter II: life %d, want %d", me.Life, life+2)
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	neoB17Check(t, neoB17Back(t, g, saga, me, "Hand of Enlightenment"), 2, 2, "first strike")
	effectstest.AssertFaceInvariant(t, g)
}

func TestTheModernAge(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.HandCardForTest(me, "Spare Forest", "Basic Land — Forest", "")
	saga := effectstest.ImportAndCast(t, g, neoB17Row(cardsth.TheModernAgeOracleID, "The Modern Age", "{1}{U}", "Vector Glider", "Enchantment Creature — Spirit", "2", "3", "U"), me)
	hand := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("chapter I raised no discard prompt after the draw")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if len(me.Hand.Cards) != hand {
		t.Errorf("hand after chapter I = %d, want %d (drew one, discarded one)", len(me.Hand.Cards), hand)
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("chapter II raised no discard prompt")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	neoB17Chapter(t, g, me, uuid.Nil)
	neoB17Check(t, neoB17Back(t, g, saga, me, "Vector Glider"), 2, 3, "flying")
}

func TestBefriendingTheMoths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	saga := effectstest.ImportAndCast(t, g, neoB17Row(cardsb.BefriendingTheMothsOracleID, "Befriending the Moths", "{3}{W}", "Imperial Moth", "Enchantment Creature — Insect", "2", "4", "W"), me)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
	}
	c := effectstest.BattlefieldCardFor(g, bear)
	neoB17Check(t, c, 3, 3, "flying")
	neoB17Chapter(t, g, me, bear)
	// Until end of turn only: the first chapter's boost ended with its turn,
	// and the second adds its own.
	c = effectstest.BattlefieldCardFor(g, bear)
	neoB17Check(t, c, 3, 3, "flying")
	neoB17Chapter(t, g, me, uuid.Nil)
	neoB17Check(t, neoB17Back(t, g, saga, me, "Imperial Moth"), 2, 4, "flying")
}

func TestTalesOfMasterSeshiro(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	van := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Parked Van", TypeLine: "Artifact — Vehicle", Power: 3, Toughness: 3})
	saga := effectstest.ImportAndCast(t, g, neoB17Row(cardssy.TalesOfMasterSeshiroOracleID, "Tales of Master Seshiro", "{4}{G}", "Seshiro's Living Legacy", "Enchantment Creature — Snake Warrior", "5", "5", "G"), me)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("chapter I: %d counters on the creature, want 1", got)
	}
	if !game.HasKeyword(effectstest.BattlefieldCardFor(g, bear), "vigilance") {
		t.Error("the creature did not gain vigilance")
	}
	neoB17Chapter(t, g, me, van)
	if got := effectstest.CounterCount(g, van, game.CounterPlusOne); got != 1 {
		t.Errorf("chapter II: %d counters on the Vehicle, want 1", got)
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	neoB17Check(t, neoB17Back(t, g, saga, me, "Seshiro's Living Legacy"), 5, 5, "vigilance", "haste")
}

func TestTheShatteredStatesEra(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	mine := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3, Tapped: true})
	saga := effectstest.ImportAndCast(t, g, neoB17Row(cardsth.TheShatteredStatesEraOracleID, "The Shattered States Era", "{4}{R}", "Nameless Conqueror", "Enchantment Creature — Human Samurai", "3", "3", "R"), me)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, theirs)
		effectstest.PassPriorityAroundTable(t, g)
	}
	stolen := effectstest.BattlefieldCardFor(g, theirs)
	if stolen.Controller != me.ID || stolen.Tapped || !game.HasKeyword(stolen, "haste") {
		t.Errorf("chapter I: controller mine=%v tapped=%v haste=%v", stolen.Controller == me.ID, stolen.Tapped, game.HasKeyword(stolen, "haste"))
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if effectstest.BattlefieldCardFor(g, theirs).Controller != opp.ID {
		t.Error("control did not revert at end of turn")
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	if p := effectstest.BattlefieldCardFor(g, mine).CurrentPower(); p != 3 {
		t.Errorf("chapter II: my creature has power %d, want 3", p)
	}
	if p := effectstest.BattlefieldCardFor(g, theirs).CurrentPower(); p != 3 {
		t.Errorf("chapter II touched their creature: power %d, want 3", p)
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	neoB17Check(t, neoB17Back(t, g, saga, me, "Nameless Conqueror"), 3, 3, "trample", "haste")
}

func TestOkibaReckonerRaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	van := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Parked Van", TypeLine: "Artifact — Vehicle", Power: 3, Toughness: 3})
	theirVan := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Van", TypeLine: "Artifact — Vehicle", Power: 3, Toughness: 3})
	saga := effectstest.ImportAndCast(t, g, neoB17Row(cardso.OkibaReckonerRaidOracleID, "Okiba Reckoner Raid", "{B}", "Nezumi Road Captain", "Enchantment Creature — Rat Rogue", "2", "2", "B"), me)
	myLife, theirLife := me.Life, opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+1 || opp.Life != theirLife-1 {
		t.Errorf("chapter I: life %d/%d, want %d/%d", me.Life, opp.Life, myLife+1, theirLife-1)
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	if me.Life != myLife+2 || opp.Life != theirLife-2 {
		t.Errorf("chapter II: life %d/%d, want %d/%d", me.Life, opp.Life, myLife+2, theirLife-2)
	}
	neoB17Chapter(t, g, me, uuid.Nil)
	neoB17Check(t, neoB17Back(t, g, saga, me, "Nezumi Road Captain"), 2, 2, "menace")
	if !game.HasKeyword(effectstest.BattlefieldCardFor(g, van), "menace") {
		t.Error("my Vehicle lacks menace")
	}
	if game.HasKeyword(effectstest.BattlefieldCardFor(g, theirVan), "menace") {
		t.Error("their Vehicle has menace")
	}
}

func TestLifeOfToshiroUmezawa(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	saga := effectstest.ImportAndCast(t, g, neoB17Row(cardskr.LifeOfToshiroUmezawaOracleID, "Life of Toshiro Umezawa", "{1}{B}", "Memory of Toshiro", "Enchantment Creature — Human Samurai", "2", "3", "B"), me)
	effectstest.PassPriorityAroundTable(t, g)
	mode := effectstest.ModePickChoiceFor(g, me.ID)
	if mode == nil {
		t.Fatal("chapter I raised no mode prompt")
	}
	effectstest.AnswerMode(t, g, mode, me.ID, 0)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickTriggerTarget(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	neoB17Check(t, effectstest.BattlefieldCardFor(g, bear), 4, 4)

	life := me.Life
	neoB17ToMain(t, g, me)
	mode = effectstest.ModePickChoiceFor(g, me.ID)
	if mode == nil {
		t.Fatal("chapter II raised no mode prompt")
	}
	effectstest.AnswerMode(t, g, mode, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := effectstest.ImportAndCast(t, g, neoB17Row(cardsap.AzusasManyJourneysOracleID, "Azusa's Many Journeys", "{1}{G}", "Likeness of the Seeker", "Enchantment Creature — Human Monk", "3", "3", "G"), me)
	before := g.ExtraLandDropsThisTurn[me.ID]
	effectstest.PassPriorityAroundTable(t, g)
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
