package testsst

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers_cards2_test.go: one test per sticker or ticket card beyond the
// two sheets and Ticketomaton (ADR 0208, #2138).

const (
	aerialephantOracle       = "20046568-b067-49fb-93b4-2ee86421f14b"
	minotaurDeForceOracle    = "67151297-3d25-4882-b562-56025ee81f48"
	stiltstriderOracle       = "9c0d7d7d-5513-4e2b-b834-1ff999ab24f5"
	carnivalCarnivoreOracle  = "ba5bc14b-db2c-4561-a6a4-a02443791260"
	glitterflitterOracle     = "3786047e-f136-49f8-81d6-08beb3f9bcfd"
	grabbyTabbyOracle        = "67ee367c-e346-48b6-8694-01c8d5f020cb"
	scaredStiffOracle        = "4a5d8acd-a460-43b2-816a-70f61d7831af"
	sanguineSipperOracle     = "ea787e29-5e24-4136-bdf6-bee338899e34"
	croakidOracle            = "8b1796d2-5b19-48ab-8699-01f45d427498"
	bigWinnerOracle          = "b181f5f7-d857-4ed0-996c-bb65b37bfb1b"
	ticketTurbotubesOracle   = "85abdc70-6fab-4d44-b286-b044eb839c5e"
	blorbianBuddyOracle      = "ebd06311-4aa6-4db9-a90a-666cc4e744d1"
	prizeWallOracle          = "41b05346-8531-4086-96a5-e973aff150cb"
	bioluminaryOracle        = "fb8b8c58-7bdc-4daa-b48b-728a26e174dd"
	costumeShopOracle        = "24c52448-03f2-43cd-9172-487384baeb18"
	weeChampionOracle        = "ad8b90c0-4c66-4fc6-a749-cd10e6deabdb"
	goblinAirbrusherOracle   = "906881da-6bb3-4a0e-9bd1-d1a5bc484144"
	tuskAndWhiskersOracle    = "de45de7e-5bba-4612-8bb7-755710d93ecf"
	roboPinataOracle         = "9ce17fde-71ec-482a-b9f9-78681bc894d8"
	commandPerformanceOracle = "f6c079cd-69c2-4da4-ba41-91844a4d2f06"
	finishingMoveOracle      = "d41dc3a5-58e4-444a-b388-5a32ccbed0eb"
	baaallerinaOracle        = "3bf3e503-0628-4550-84a9-7c1dd0a97457"
	seraphOracle             = "01efde9a-a20d-4972-b17f-c2d9eb024b0b"
	pyrodancerOracle         = "ddd3f530-1ea4-4b1e-9f9b-5d34fa626529"
	lineprancersOracle       = "71433aef-5a06-449e-93b1-9565cb0afd93"
	parkBleaterOracle        = "02f31e00-646d-4971-8704-1583bbfcddd6"
)

// stkAskedOptions are the labels of the open sticker question, after the
// object question is answered with `obj`.
func stkAskedOptions(t *testing.T, g *game.Game, me *game.Player, obj uuid.UUID) []string {
	t.Helper()
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no question asks which permanent gets the sticker")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{obj}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	which := effectstest.LatestOptionPickFor(g, me.ID)
	if which == nil {
		t.Fatal("no question asks which sticker")
	}
	out := make([]string, 0, len(which.PickOptions))
	for _, o := range which.PickOptions {
		out = append(out, o.Label)
	}
	return out
}

func stkHasKeyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	c := effectstest.StkCard(t, g, id)
	return game.HasKeyword(&c, kw)
}

// The five "you get {TK}, then you may put a sticker" creatures.
func TestStkStickerEntryCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		tickets                int
	}{
		{"Aerialephant", "Creature — Elephant Performer", aerialephantOracle, 1},
		{"Minotaur de Force", "Creature — Minotaur Performer", minotaurDeForceOracle, 1},
		{"Stiltstrider", "Creature — Rhino Performer", stiltstriderOracle, 2},
		{"Carnival Carnivore", "Creature — Alien Crab Horror", carnivalCarnivoreOracle, 1},
		{"Glitterflitter", "Creature — Faerie Performer", glitterflitterOracle, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me := stkTable(t, 0)
			id := effectstest.CastAndResolveCreature(t, g, tc.name, tc.typeLine, tc.oracle)
			effectstest.PassPriorityAroundTable(t, g)
			if got := stkTickets(g, me); got != tc.tickets {
				t.Fatalf("tickets = %d, want %d", got, tc.tickets)
			}
			effectstest.StkPutOn(t, g, me, id, "“Night”", 0)
			if got := effectstest.StkCard(t, g, id).Effective().Name; !strings.HasPrefix(got, "Night ") {
				t.Errorf("name = %q, want the Night sticker first", got)
			}
		})
	}
}

// "This creature has <keyword> as long as you control a stickered
// permanent."
func TestStkStickeredPermanentStatics(t *testing.T) {
	for _, tc := range []struct{ name, oracle, kw string }{
		{"Grabby Tabby", grabbyTabbyOracle, "vigilance"},
		{"Scared Stiff", scaredStiffOracle, "menace"},
		{"Sanguine Sipper", sanguineSipperOracle, "lifelink"},
		{"Croakid Amphibonaut", croakidOracle, "flying"},
		{"Big Winner", bigWinnerOracle, "trample"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me := stkTable(t, 0)
			guest := effectstest.P2Push(g, me.ID, tc.name, "Creature — Guest", tc.oracle, 3, 3)
			bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
			if stkHasKeyword(t, g, guest, tc.kw) {
				t.Fatalf("%s has %s with nothing stickered", tc.name, tc.kw)
			}
			effectstest.StkGive(t, g, me, bear, "Art sticker 1")
			if !stkHasKeyword(t, g, guest, tc.kw) {
				t.Fatalf("%s lacks %s with a stickered permanent", tc.name, tc.kw)
			}
			g.WithWriteLock(func() { _ = g.BounceToHandForEffect(bear) })
			if stkHasKeyword(t, g, guest, tc.kw) {
				t.Errorf("%s kept %s after the stickered permanent left", tc.name, tc.kw)
			}
		})
	}
}

// Ticket Turbotubes, Blorbian Buddy and Prize Wall's "{…}, {T}: You get
// {TK}."
func TestStkTicketActivations(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana string
		index                        int
	}{
		{"Ticket Turbotubes", "Artifact", ticketTurbotubesOracle, "CCC", 0},
		{"Blorbian Buddy", "Creature — Alien Guest", blorbianBuddyOracle, "G", 0},
		{"Prize Wall", "Creature — Wall", prizeWallOracle, "U", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me := stkTable(t, 0)
			id := effectstest.P2Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			if err := effectstest.P2Activate(t, g, me, id, tc.index, tc.mana); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := stkTickets(g, me); got != 1 {
				t.Errorf("tickets = %d, want 1", got)
			}
		})
	}
}

// Prize Wall's second ability asks for a sticker.
func TestStkPrizeWallPutsASticker(t *testing.T) {
	g, me := stkTable(t, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	wall := effectstest.P2Push(g, me.ID, "Prize Wall", "Creature — Wall", prizeWallOracle, 0, 4)
	if err := effectstest.P2Activate(t, g, me, wall, 1, "CCCCU"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.StkPutOn(t, g, me, wall, "“Midway”", 0)
	if got := effectstest.StkCard(t, g, wall).Effective().Name; got != "Midway Prize Wall" {
		t.Errorf("name = %q, want %q", got, "Midway Prize Wall")
	}
}

// Bioluminary: combat damage to a player gives {TK}{TK}, then a sticker.
func TestStkBioluminaryCombatDamage(t *testing.T) {
	g, me := stkTable(t, 0)
	bio := effectstest.P2Push(g, me.ID, "Bioluminary", "Creature — Merfolk Performer", bioluminaryOracle, 1, 3)
	effectstest.AttackWith(t, g, g.Seats[1].ID, bio)
	effectstest.PassPriorityAroundTable(t, g)
	if got := stkTickets(g, me); got != 2 {
		t.Fatalf("tickets = %d, want 2", got)
	}
	effectstest.StkPutOn(t, g, me, bio, "sticker 4/2", 0)
	if p, tough := effectstest.P2PT(g, bio); p != 4 || tough != 2 {
		t.Errorf("P/T = %d/%d, want 4/2", p, tough)
	}
}

// Costume Shop's visit asks for a sticker.
func TestStkCostumeShopVisit(t *testing.T) {
	g := effectstest.InMainPhase(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		me.StickerSheets = []game.StickerSheet{stkSheet("Night Brushwagg Ringmaster", effectstest.NightBrushwaggOracle, effectstest.NightBrushwaggText)}
	})
	shop := effectstest.PushAttraction(g, me.ID, "Costume Shop", costumeShopOracle, effectstest.EveryNumber...)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.RollToVisit(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.StkPutOn(t, g, me, bear, "“Brushwagg”", 0)
	if got := effectstest.StkCard(t, g, bear).Effective().Name; got != "Brushwagg Bear" {
		t.Errorf("name = %q, want %q", got, "Brushwagg Bear")
	}
	_ = shop
}

// Wee Champion: +1/+1 until end of turn, or a counter for an art sticker.
func TestStkWeeChampion(t *testing.T) {
	g, me := stkTable(t, 0)
	wee := effectstest.P2Push(g, me.ID, "Wee Champion", "Creature — Minotaur Child Guest", weeChampionOracle, 0, 1)
	effectstest.StkGive(t, g, me, wee, "“Night”")
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.P2PT(g, wee); p != 1 || tough != 2 {
		t.Fatalf("after a name sticker P/T = %d/%d, want 1/2", p, tough)
	}
	if n := effectstest.CountersOn(g, wee, game.CounterPlusOne); n != 0 {
		t.Fatalf("a name sticker put %d counters on", n)
	}
	effectstest.StkGive(t, g, me, wee, "Art sticker 1")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, wee, game.CounterPlusOne); n != 1 {
		t.Errorf("an art sticker put %d +1/+1 counters on, want 1", n)
	}
}

// Goblin Airbrusher: a Treasure, or two for an art sticker.
func TestStkGoblinAirbrusher(t *testing.T) {
	g, me := stkTable(t, 0)
	gob := effectstest.P2Push(g, me.ID, "Goblin Airbrusher", "Creature — Goblin Employee", goblinAirbrusherOracle, 2, 1)
	effectstest.StkGive(t, g, me, gob, "“Night”")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Treasure"); n != 1 {
		t.Fatalf("Treasures after a name sticker = %d, want 1", n)
	}
	effectstest.StkGive(t, g, me, gob, "Art sticker 2")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Treasure"); n != 3 {
		t.Errorf("Treasures after an art sticker = %d, want 3", n)
	}
}

// Tusk and Whiskers: an ability sticker on a creature gets a +1/+1
// counter; a name sticker does not.
func TestStkTuskAndWhiskers(t *testing.T) {
	g, me := stkTable(t, 2)
	tusk := effectstest.P2Push(g, me.ID, "Tusk and Whiskers", "Legendary Creature — Elephant Mouse Performer", tuskAndWhiskersOracle, 4, 4)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.StkGive(t, g, me, bear, "“Night”")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 0 {
		t.Fatalf("a name sticker put %d counters on", n)
	}
	effectstest.StkGive(t, g, me, bear, "Menace")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Fatalf("an ability sticker put %d counters on, want 1", n)
	}
	if err := effectstest.P2Activate(t, g, me, tusk, 0, "CCGW"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := stkTickets(g, me); got != 1 {
		t.Errorf("tickets after the activation = %d, want 1", got)
	}
	if effectstest.LatestChooseCardsFor(g, me.ID) == nil {
		t.Error("the activation did not ask for a sticker")
	}
}

// Robo-Piñata: the dies trigger's two modes.
func TestStkRoboPinata(t *testing.T) {
	for _, mode := range []int{0, 1} {
		g, me := stkTable(t, 0)
		pin := effectstest.P2Push(g, me.ID, "Robo-Piñata", "Artifact Creature — Clown Robot", roboPinataOracle, 2, 1)
		bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
		effectstest.B25Destroy(g, pin)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.ModePickChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("the dies trigger asks which mode")
		}
		if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
			t.Fatalf("mode: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if mode == 0 {
			if got := stkTickets(g, me); got != 2 {
				t.Errorf("tickets = %d, want 2", got)
			}
			continue
		}
		effectstest.StkPutOn(t, g, me, bear, "“Night”", 0)
		if got := effectstest.StkCard(t, g, bear).Effective().Name; got != "Night Bear" {
			t.Errorf("name = %q, want Night Bear", got)
		}
	}
}

// Command Performance: "get {TK}{TK}" and "put a sticker", chosen
// together, pay for the sticker with those tickets.
func TestStkCommandPerformanceTicketsPayForTheSticker(t *testing.T) {
	g, me := stkTable(t, 0)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpellWithModes(t, g, "Command Performance", "Sorcery", commandPerformanceOracle, []int{2, 3})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.StkPutOn(t, g, me, bear, "sticker 2/3", 0)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.P2PT(g, bear); p != 2 || tough != 3 {
		t.Errorf("P/T = %d/%d, want 2/3", p, tough)
	}
	if got := stkTickets(g, me); got != 0 {
		t.Errorf("tickets = %d, want 0", got)
	}
}

// Finishing Move: the damage waits for the sticker, so a power and
// toughness sticker on the biter changes what it deals (CR 608.2c).
func TestStkFinishingMoveDamageWaitsForTheSticker(t *testing.T) {
	g, me := stkTable(t, 0)
	biter := effectstest.P2Push(g, me.ID, "Biter", "Creature — Bear", "", 1, 1)
	victim := effectstest.P2Push(g, g.Seats[1].ID, "Victim", "Creature — Bear", "", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Finishing Move", "Sorcery", finishingMoveOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: biter}, {Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.StkPutOn(t, g, me, biter, "sticker 4/2", 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, victim) {
		t.Error("the 4/4 survived: the bite did not wait for the 4/2 sticker")
	}
}

// Baaallerina: only name stickers are offered, and its ability reaches
// only a creature with one.
func TestStkBaaallerina(t *testing.T) {
	g, me := stkTable(t, 2)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	bal := effectstest.CastAndResolveCreature(t, g, "Baaallerina", "Creature — Sheep Performer", baaallerinaOracle)
	effectstest.PassPriorityAroundTable(t, g)
	for _, label := range stkAskedOptions(t, g, me, bear) {
		if label != game.DeclineStickerLabel && !strings.HasPrefix(label, "Name sticker") {
			t.Errorf("offered %q to a name-sticker instruction", label)
		}
	}
	which := effectstest.LatestOptionPickFor(g, me.ID)
	if err := g.ResolveOptionPick(which.ID, me.ID, effectstest.StkOption(t, which, "“Night”")); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if where := effectstest.LatestOptionPickFor(g, me.ID); where != nil {
		_ = g.ResolveOptionPick(where.ID, me.ID, 0)
	}
	if err := effectstest.P2Activate(t, g, me, bal, 0, "CCU", effectstest.P2Card(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !stkHasKeyword(t, g, bear, "flying") {
		t.Error("the named creature did not gain flying")
	}
	plain := effectstest.P2Push(g, me.ID, "Plain Bear", "Creature — Bear", "", 2, 2)
	if err := effectstest.P2Activate(t, g, me, bal, 0, "CCU", effectstest.P2Card(plain)); err == nil {
		t.Error("a creature with no name sticker was a legal target")
	}
}

// Sword-Swallowing Seraph: a +1/+1 counter on a creature with a name
// sticker.
func TestStkSwordSwallowingSeraph(t *testing.T) {
	g, me := stkTable(t, 0)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	seraph := effectstest.P2Push(g, me.ID, "Sword-Swallowing Seraph", "Creature — Angel Performer", seraphOracle, 4, 4)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.StkGive(t, g, me, bear, "“Night”")
	if err := effectstest.P2Activate(t, g, me, seraph, 0, "CW", effectstest.P2Card(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("counters = %d, want 1", n)
	}
}

// Proficient Pyrodancer: only art stickers offered; +2/+0 and menace for a
// creature with one.
func TestStkProficientPyrodancer(t *testing.T) {
	g, me := stkTable(t, 0)
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	pyro := effectstest.CastAndResolveCreature(t, g, "Proficient Pyrodancer", "Creature — Human Performer", pyrodancerOracle)
	effectstest.PassPriorityAroundTable(t, g)
	for _, label := range stkAskedOptions(t, g, me, bear) {
		if label != game.DeclineStickerLabel && !strings.HasPrefix(label, "Art sticker") {
			t.Errorf("offered %q to an art-sticker instruction", label)
		}
	}
	which := effectstest.LatestOptionPickFor(g, me.ID)
	if err := g.ResolveOptionPick(which.ID, me.ID, effectstest.StkOption(t, which, "Art sticker 1")); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if err := effectstest.P2Activate(t, g, me, pyro, 0, "CCR", effectstest.P2Card(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.P2PT(g, bear); p != 4 {
		t.Errorf("power = %d, want 4", p)
	}
	if !stkHasKeyword(t, g, bear, "menace") {
		t.Error("no menace")
	}
}

// Lineprancers: {TK}{TK}, then only power and toughness stickers on
// creatures; its block requirement reaches only a creature with one.
func TestStkLineprancers(t *testing.T) {
	g, me := stkTable(t, 0)
	lp := effectstest.CastAndResolveCreature(t, g, "Lineprancers", "Creature — Centaur Performer", lineprancersOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if got := stkTickets(g, me); got != 2 {
		t.Fatalf("tickets = %d, want 2", got)
	}
	for _, label := range stkAskedOptions(t, g, me, lp) {
		if label != game.DeclineStickerLabel && !strings.HasPrefix(label, "Power and toughness sticker") {
			t.Errorf("offered %q to a power-and-toughness instruction", label)
		}
	}
	which := effectstest.LatestOptionPickFor(g, me.ID)
	if err := g.ResolveOptionPick(which.ID, me.ID, effectstest.StkOption(t, which, "sticker 2/3")); err != nil {
		t.Fatalf("pick: %v", err)
	}
	bear := effectstest.P2Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.P2Push(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if err := effectstest.P2Activate(t, g, me, lp, 0, "CCCG", effectstest.P2Card(theirs), effectstest.P2Card(bear)); err == nil {
		t.Error("a creature with no power and toughness sticker was a legal second target")
	}
}

// Park Bleater: another creature you own entering gives {TK}; its ability
// stickers a creature that entered this turn.
func TestStkParkBleater(t *testing.T) {
	g, me := stkTable(t, 0)
	bleater := effectstest.P2Push(g, me.ID, "Park Bleater", "Creature — Goat Employee", parkBleaterOracle, 2, 2)
	bear := effectstest.CastAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := stkTickets(g, me); got != 1 {
		t.Fatalf("tickets = %d, want 1 (Park Bleater's)", got)
	}
	if err := effectstest.P2Activate(t, g, me, bleater, 0, "W"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the activation did not ask")
	}
	for _, id := range pick.ChooseCards {
		if id == bleater {
			t.Error("Park Bleater, which did not enter this turn, was offered")
		}
	}
	found := false
	for _, id := range pick.ChooseCards {
		found = found || id == bear
	}
	if !found {
		t.Error("the creature that entered this turn was not offered")
	}
}
