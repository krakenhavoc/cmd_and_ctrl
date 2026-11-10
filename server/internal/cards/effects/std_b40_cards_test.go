package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b40_cards_test.go — Standard-set leftovers, batch 40 (Murders at
// Karlov Manor, #4469). Each test plays the real catalog entry and asserts
// what its text says.

const (
	stdB40GadgeteerOracle  = "9d5917d5-4fb3-4f37-9278-e08c1f73aae5"
	stdB40MazeOracle       = "3a024389-91a1-460c-b4e8-3f212d2b8400"
	stdB40ScholarOracle    = "0fb9a928-9adb-4dc1-9761-857df7e6c643"
	stdB40SetbackOracle    = "565ba85c-bad8-4cbd-a2ff-b4384b886308"
	stdB40MortipedeOracle  = "1f4d3a96-bc30-4362-9c1f-40bf79f78e78"
	stdB40EliminatorOracle = "93cb9c7e-d6d3-4b1e-8bcf-2a4a1d95a549"
	stdB40AgentOracle      = "3e7879d5-62ea-4c9a-9fc0-659d70f3a8e1"
	stdB40MasksOracle      = "7c1eb874-97f3-463f-bf4e-4d82dcc16e69"
	stdB40WeaponOracle     = "104565b1-2db8-4c01-ae3d-14087768be45"
	stdB40ExposeOracle     = "4f7e4aff-ef68-44fe-a0dc-b9ab0ecd7146"
)

func TestStdB40EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB40GadgeteerOracle:  "Forensic Gadgeteer",
		stdB40MazeOracle:       "Lost in the Maze",
		stdB40ScholarOracle:    "Steamcore Scholar",
		stdB40SetbackOracle:    "Sudden Setback",
		stdB40MortipedeOracle:  "Rot Farm Mortipede",
		stdB40EliminatorOracle: "Undercity Eliminator",
		stdB40AgentOracle:      "Unscrupulous Agent",
		stdB40MasksOracle:      "Case of the Burning Masks",
		stdB40WeaponOracle:     "Concealed Weapon",
		stdB40ExposeOracle:     "Expose the Culprit",
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

// --- Steamcore Scholar ---------------------------------------------------------

// stdB40ScholarDiscard casts the Scholar with a hand holding `held`, and
// returns the open discard prompt.
func stdB40ScholarDiscard(t *testing.T, g *game.Game, me *game.Player, held ...game.Card) (*game.PendingChoice, []uuid.UUID) {
	t.Helper()
	auraSeedLibrary(me, 4)
	var ids []uuid.UUID
	for _, c := range held {
		ids = append(ids, avB1Hand(me, c))
	}
	avB1ResolveCreature(t, g, me, "Steamcore Scholar", "Creature — Weird Detective", stdB40ScholarOracle)
	passPriorityAroundTable(t, g)
	c := discardChoiceFor(g, me.ID)
	if c == nil || c.ChooseMin != 1 || c.ChooseMax != 2 {
		t.Fatalf("a floor of one and a ceiling of two: %+v", c)
	}
	return c, ids
}

func TestStdB40SteamcoreScholarDrawsTwoThenDiscardsTwoOrAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	c, ids := stdB40ScholarDiscard(t, g, me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Shock", TypeLine: "Instant"})
	if got := me.Hand.Size(); got != hand+2+2 {
		t.Fatalf("hand %d → %d: two held cards, the Scholar cast, two drawn", hand, got)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[0]}); err == nil {
		t.Fatal("one creature card without flying was accepted")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("one instant card is a complete discard: %v", err)
	}
	if me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[0]) {
		t.Error("only the instant was discarded")
	}
}

func TestStdB40SteamcoreScholarAcceptsAFlyingCreatureOrAnySecondCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c, ids := stdB40ScholarDiscard(t, g, me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"}})
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("one creature card with flying is a complete discard: %v", err)
	}

	g = newCatalogGame(t)
	me = g.Seats[0]
	c, ids = stdB40ScholarDiscard(t, g, me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Rock", TypeLine: "Artifact"})
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{ids[0], ids[1]}); err != nil {
		t.Fatalf("any two cards are a complete discard: %v", err)
	}
}

// --- Unscrupulous Agent --------------------------------------------------------

func TestStdB40UnscrupulousAgentMakesTheTargetOpponentExileACardOfTheirChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keep := handCardFull(opp, "Keep", "Sorcery", "{1}", "", nil)
	lose := handCardFull(opp, "Lose", "Sorcery", "{1}", "", nil)
	castHoldCreature(t, g, "Unscrupulous Agent", "Creature — Elf Detective", stdB40AgentOracle, 1, 1)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	c := pendingChooseCards(t, g, opp.ID)
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{lose}); err != nil {
		t.Fatalf("the opponent picks the card: %v", err)
	}
	if opp.Hand.Contains(lose) || !opp.Hand.Contains(keep) {
		t.Error("the opponent's chosen card was not the one exiled")
	}
	exiled := false
	g.ReadSnapshot(func() { exiled = g.Exile.Contains(lose) })
	if !exiled {
		t.Error("the card is not in exile")
	}
}

// --- Rot Farm Mortipede --------------------------------------------------------

func TestStdB40RotFarmMortipedeGrowsWhenACreatureCardLeavesYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	m := pushCatalogPermanent(g, me.ID, "Rot Farm Mortipede", "Creature — Insect", stdB40MortipedeOracle, false)
	g.WithWriteLock(func() {
		c := battlefieldCardPtr(t, g, m)
		c.Power, c.Toughness = 3, 4
	})
	rock := mkmMP2Grave(me, "Rock", "Artifact")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, rock); err != nil {
		t.Fatalf("move: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := mkmMPPT(t, g, m); p != 3 {
		t.Fatalf("an artifact card leaving the graveyard made the power %d, want 3", p)
	}
	bear := mkmMP2Grave(me, "Bear", "Creature — Bear")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, bear); err != nil {
		t.Fatalf("move: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := mkmMPPT(t, g, m); p != 4 || tough != 4 {
		t.Errorf("power/toughness %d/%d, want 4/4", p, tough)
	}
	for _, kw := range []string{"menace", "lifelink"} {
		if !hasEffectiveKeyword(t, g, m, kw) {
			t.Errorf("no %s until end of turn", kw)
		}
	}
	// An opponent's graveyard does not count.
	opp := g.Seats[1]
	theirs := mkmMP2Grave(opp, "Their Bear", "Creature — Bear")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, theirs); err != nil {
		t.Fatalf("move: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := mkmMPPT(t, g, m); p != 4 {
		t.Errorf("a card leaving an opponent's graveyard changed the power to %d", p)
	}
}

// --- Undercity Eliminator ------------------------------------------------------

func TestStdB40UndercityEliminatorSacrificesThenExilesATarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	castHoldCreature(t, g, "Undercity Eliminator", "Creature — Gorgon Assassin", stdB40EliminatorOracle, 3, 3)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("no target is chosen when the enters trigger goes on the stack")
	}
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil || !hasID(c.ChooseCards, rock) {
		t.Fatalf("an artifact is a legal sacrifice: %+v", c)
	}
	answerOwnPermanents(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the reflexive trigger asked for no target")
	}
	if !hasID(p.PickTargetCards, victim) {
		t.Fatalf("a creature an opponent controls: %v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not exiled")
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
}

func TestStdB40UndercityEliminatorDecliningDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	castHoldCreature(t, g, "Undercity Eliminator", "Creature — Gorgon Assassin", stdB40EliminatorOracle, 3, 3)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, g.Seats[0].ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Error("the creature was exiled without a sacrifice")
	}
}
