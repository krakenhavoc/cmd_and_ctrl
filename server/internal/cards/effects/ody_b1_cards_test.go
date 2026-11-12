package effects

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ody_b1_cards_test.go — Odyssey, batch 1 (#5613). Each test plays the
// real catalog entry and asserts what its text says.

const (
	odyB1FireboltOracle      = "7aa31280-12c3-479d-a6dd-f1c669043083"
	odyB1CallOfTheHerdOracle = "ee243f81-f51c-4d9a-a396-f7cef84b46c1"
	odyB1ChatterOracle       = "b0aaa4b1-5188-43a6-997d-7a9b2ad452bc"
	odyB1BeastAttackOracle   = "3f5acd90-199c-4ed3-9c8a-eeac4670d876"
	odyB1MorgueTheftOracle   = "0b8a508b-20c0-4a87-a73a-08af72c39a0a"
	odyB1SkullFractureOracle = "3384606b-7724-4584-89fa-fbe006d5de1f"
	odyB1KrosanBeastOracle   = "042a665d-2f63-4f7e-ac89-0d560da68085"
	odyB1MetamorphicOracle   = "512477f0-1083-423b-9675-a8bc852f7d88"
	odyB1SpringingOracle     = "2f8c4805-694d-4408-8eb5-9fb7ce278ded"
	odyB1MongooseOracle      = "1059f089-2693-4c89-92c5-39fe49beebb3"
	odyB1WerebearOracle      = "96cf3320-c730-4618-8768-28a179f4d9c0"
	odyB1ZealotOracle        = "5a47b99d-19dd-4830-89ee-35f4a74aa656"
	odyB1PenitentOracle      = "d3d59704-569d-42a7-8b31-8a31b9c0395f"
	odyB1VisionaryOracle     = "bde74bcc-41b8-4014-b28e-2f6326e20799"
)

func TestOdyB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		odyB1FireboltOracle:      "Firebolt",
		odyB1CallOfTheHerdOracle: "Call of the Herd",
		odyB1ChatterOracle:       "Chatter of the Squirrel",
		odyB1BeastAttackOracle:   "Beast Attack",
		odyB1MorgueTheftOracle:   "Morgue Theft",
		odyB1SkullFractureOracle: "Skull Fracture",
		odyB1KrosanBeastOracle:   "Krosan Beast",
		odyB1MetamorphicOracle:   "Metamorphic Wurm",
		odyB1SpringingOracle:     "Springing Tiger",
		odyB1MongooseOracle:      "Nimble Mongoose",
		odyB1WerebearOracle:      "Werebear",
		odyB1ZealotOracle:        "Mystic Zealot",
		odyB1PenitentOracle:      "Mystic Penitent",
		odyB1VisionaryOracle:     "Mystic Visionary",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- flashback ---------------------------------------------------------------

// odyB1FlashbackCards is the batch's flashback roster with its printed
// flashback price.
var odyB1FlashbackCards = []struct{ name, oracle, cost string }{
	{"Firebolt", odyB1FireboltOracle, "{4}{R}"},
	{"Call of the Herd", odyB1CallOfTheHerdOracle, "{3}{G}"},
	{"Chatter of the Squirrel", odyB1ChatterOracle, "{1}{G}"},
	{"Beast Attack", odyB1BeastAttackOracle, "{2}{G}{G}{G}"},
	{"Morgue Theft", odyB1MorgueTheftOracle, "{4}{B}"},
	{"Skull Fracture", odyB1SkullFractureOracle, "{3}{B}"},
}

func TestOdyB1FlashbackCardsOfferTheirPrintedPriceFromTheGraveyard(t *testing.T) {
	for _, c := range odyB1FlashbackCards {
		t.Run(c.name, func(t *testing.T) {
			if !game.CardCastableFromZone(c.oracle, game.ZoneGraveyard) {
				t.Fatalf("%s does not open the graveyard", c.name)
			}
			offers := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneGraveyard)
			if len(offers) != 1 || offers[0].Key != "flashback" {
				t.Fatalf("%s graveyard offers = %+v, want one flashback", c.name, offers)
			}
			if offers[0].ManaCost != c.cost || !offers[0].ExileOnLeavingStack {
				t.Errorf("%s flashback pays %q (exile %v), want %q and exile", c.name, offers[0].ManaCost, offers[0].ExileOnLeavingStack, c.cost)
			}
		})
	}
}

// odyB1Flashback casts the card out of the active seat's graveyard for
// its flashback price, resolves it, and checks it was exiled.
func odyB1Flashback(t *testing.T, g *game.Game, id uuid.UUID, cost string, targets ...game.TargetRef) {
	t.Helper()
	odyB1CastFlashback(t, g, id, cost, targets...)
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back card was not exiled")
	}
}

// odyB1CastFlashback casts the card out of the active seat's graveyard
// for its flashback price, paid from a pool holding exactly that price,
// and passes priority around the table once.
func odyB1CastFlashback(t *testing.T, g *game.Game, id uuid.UUID, cost string, targets ...game.TargetRef) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	dcMana(t, g, me.ID, odyB1Pool(cost))
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Strict: true, Targets: targets,
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left after paying %s", n, cost)
	}
	passPriorityAroundTable(t, g)
}

// odyB1Pool is the mana that pays exactly `cost`, its generic part as
// {C}: "{4}{R}" is "{C}{C}{C}{C}{R}".
func odyB1Pool(cost string) string {
	var out strings.Builder
	for _, sym := range strings.Split(strings.Trim(cost, "{}"), "}{") {
		if n, err := strconv.Atoi(sym); err == nil {
			out.WriteString(strings.Repeat("{C}", n))
			continue
		}
		out.WriteString("{" + sym + "}")
	}
	return out.String()
}

func TestOdyB1FireboltDealsTwoAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	life := opp.Life
	castCatalogSpell(t, g, "Firebolt", "Sorcery", odyB1FireboltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Fatalf("Firebolt dealt %d, want 2", got)
	}

	bolt := seedGraveyardCard(t, g, "Firebolt", "Sorcery", odyB1FireboltOracle)
	bear := blbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	odyB1Flashback(t, g, bolt, "{4}{R}", cardTarget(bear)...)
	if g.Battlefield.Contains(bear) {
		t.Error("the flashed-back Firebolt did not kill the 2/2")
	}
}

func TestOdyB1FireboltFlashbackIsNotPaidByTheHandCost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bolt := seedGraveyardCard(t, g, "Firebolt", "Sorcery", odyB1FireboltOracle)
	dcMana(t, g, me.ID, "{R}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Strict: true,
		Targets: []game.TargetRef{playerTarget(opp.ID)},
	}); err == nil {
		t.Fatal("flashback was cast for {R}, the hand cost")
	}
	if !me.Graveyard.Contains(bolt) {
		t.Error("a refused flashback moved Firebolt out of the graveyard")
	}
}

func TestOdyB1TokenSpellsMakeTheirTokenFromHandAndGraveyard(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, cost, token string
		power, toughness                    int
	}{
		{"Call of the Herd", "Sorcery", odyB1CallOfTheHerdOracle, "{3}{G}", "Elephant", 3, 3},
		{"Chatter of the Squirrel", "Sorcery", odyB1ChatterOracle, "{1}{G}", "Squirrel", 1, 1},
		{"Beast Attack", "Instant", odyB1BeastAttackOracle, "{2}{G}{G}{G}", "Beast", 4, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			passPriorityAroundTable(t, g)
			if got := b43TokensNamed(g, me.ID, tc.token); got != 1 {
				t.Fatalf("%d %s tokens after the hand cast, want 1", got, tc.token)
			}
			id := seedGraveyardCard(t, g, tc.name, tc.typeLine, tc.oracle)
			odyB1Flashback(t, g, id, tc.cost)
			if got := b43TokensNamed(g, me.ID, tc.token); got != 2 {
				t.Fatalf("%d %s tokens after the flashback, want 2", got, tc.token)
			}
			var tokens []game.Card
			for _, c := range g.Battlefield.Cards {
				if c.IsToken() && c.Name == tc.token {
					tokens = append(tokens, c)
				}
			}
			for _, c := range tokens {
				if p, tough := ptOf(t, g, c.InstanceID); p != tc.power || tough != tc.toughness || !c.HasColor("G") {
					t.Errorf("the %s token is %d/%d green=%v, want a green %d/%d", tc.token, p, tough, c.HasColor("G"), tc.power, tc.toughness)
				}
			}
		})
	}
}

func TestOdyB1MorgueTheftReturnsACreatureCardOfYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	rite := pushGraveyardCardTyped(me, "Rite", "Sorcery")
	theirs := pushGraveyardCardTyped(opp, "Their Bear", "Creature — Bear")
	for _, bad := range []uuid.UUID{rite, theirs} {
		if err := castCatalogSpellErr(t, g, "Morgue Theft", "Sorcery", odyB1MorgueTheftOracle, cardTarget(bad)); err == nil {
			t.Fatal("a noncreature card, or a card in another graveyard, was accepted as the target")
		}
	}
	castCatalogSpell(t, g, "Morgue Theft", "Sorcery", odyB1MorgueTheftOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Fatal("the creature card did not come back to hand")
	}

	ogre := pushGraveyardCardTyped(me, "Ogre", "Creature — Ogre")
	theft := seedGraveyardCard(t, g, "Morgue Theft", "Sorcery", odyB1MorgueTheftOracle)
	odyB1Flashback(t, g, theft, "{4}{B}", cardTarget(ogre)...)
	if !me.Hand.Contains(ogre) {
		t.Error("the flashed-back Morgue Theft did not return the creature card")
	}
}

func TestOdyB1SkullFractureMakesTheTargetDiscardAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	victim := pushCatalogHandCard(opp, "Victim", "Instant", "")
	castCatalogSpell(t, g, "Skull Fracture", "Sorcery", odyB1SkullFractureOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := discardOwed(g, opp.ID); got != 1 {
		t.Fatalf("the target owes %d discards, want 1", got)
	}
	answerDiscard(t, g, opp.ID, victim)
	if !opp.Graveyard.Contains(victim) {
		t.Fatal("the chosen card was not discarded")
	}

	second := pushCatalogHandCard(opp, "Second", "Instant", "")
	fracture := seedGraveyardCard(t, g, "Skull Fracture", "Sorcery", odyB1SkullFractureOracle)
	odyB1CastFlashback(t, g, fracture, "{3}{B}", playerTarget(opp.ID))
	answerDiscard(t, g, opp.ID, second)
	g.SettleResolution()
	if !opp.Graveyard.Contains(second) {
		t.Error("the flashed-back Skull Fracture did not make the target discard")
	}
	if !g.Exile.Contains(fracture) {
		t.Error("the flashed-back Skull Fracture was not exiled")
	}
}

// --- threshold ---------------------------------------------------------------

func TestOdyB1ThresholdCreaturesGrowAtSevenCards(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		bigP, bigT             int
		flying                 bool
	}{
		{"Krosan Beast", "Creature — Squirrel Beast", odyB1KrosanBeastOracle, 1, 1, 8, 8, false},
		{"Metamorphic Wurm", "Creature — Elephant Wurm", odyB1MetamorphicOracle, 3, 3, 7, 7, false},
		{"Springing Tiger", "Creature — Cat", odyB1SpringingOracle, 3, 3, 5, 5, false},
		{"Nimble Mongoose", "Creature — Mongoose", odyB1MongooseOracle, 1, 1, 3, 3, false},
		{"Werebear", "Creature — Human Bear Druid", odyB1WerebearOracle, 1, 1, 4, 4, false},
		{"Mystic Zealot", "Creature — Human Nomad Mystic", odyB1ZealotOracle, 2, 4, 3, 5, true},
		{"Mystic Penitent", "Creature — Human Nomad Mystic", odyB1PenitentOracle, 1, 1, 2, 2, true},
		{"Mystic Visionary", "Creature — Human Nomad Mystic", odyB1VisionaryOracle, 2, 1, 2, 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp, _ := blbB1Seats(g)
			id := blbB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			// Seven cards in an OPPONENT's graveyard are not threshold.
			blbMPFillGraveyard(g, opp, 7)
			blbMPFillGraveyard(g, me, 6)
			if p, tough := ptOf(t, g, id); p != tc.power || tough != tc.toughness {
				t.Fatalf("six cards in your graveyard: %d/%d, want %d/%d", p, tough, tc.power, tc.toughness)
			}
			if hasKeywordOnBattlefield(t, g, id, "flying") {
				t.Fatal("flying before threshold")
			}
			blbMPFillGraveyard(g, me, 1)
			if p, tough := ptOf(t, g, id); p != tc.bigP || tough != tc.bigT {
				t.Errorf("seven cards in your graveyard: %d/%d, want %d/%d", p, tough, tc.bigP, tc.bigT)
			}
			if got := hasKeywordOnBattlefield(t, g, id, "flying"); got != tc.flying {
				t.Errorf("flying at threshold = %v, want %v", got, tc.flying)
			}
		})
	}
}

func TestOdyB1PrintedKeywords(t *testing.T) {
	for _, tc := range []struct{ name, oracle, keyword string }{
		{"Nimble Mongoose", odyB1MongooseOracle, "shroud"},
		{"Mystic Penitent", odyB1PenitentOracle, "vigilance"},
	} {
		g := newCatalogGame(t)
		id := blbB1Push(g, g.Seats[0].ID, tc.name, "Creature", tc.oracle, 1, 1)
		if !hasKeywordOnBattlefield(t, g, id, tc.keyword) {
			t.Errorf("%s does not have %s", tc.name, tc.keyword)
		}
	}
}

// Shroud bars its own controller's spells too (CR 702.18a).
func TestOdyB1NimbleMongooseCantBeTargetedEvenByItsController(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := blbB1Seats(g)
	mongoose := blbB1Push(g, me.ID, "Nimble Mongoose", "Creature — Mongoose", odyB1MongooseOracle, 1, 1)
	if err := castCatalogSpellErr(t, g, "Firebolt", "Sorcery", odyB1FireboltOracle, cardTarget(mongoose)); err == nil {
		t.Error("Firebolt targeted a creature with shroud")
	}
}

func TestOdyB1WerebearTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	bear := blbB1Push(g, g.Seats[0].ID, "Werebear", "Creature — Human Bear Druid", odyB1WerebearOracle, 1, 1)
	if got := ltProduced(t, g, bear); !slices.Equal(got, []string{"{G}"}) {
		t.Errorf("Werebear adds %v, want {G}", got)
	}
}
