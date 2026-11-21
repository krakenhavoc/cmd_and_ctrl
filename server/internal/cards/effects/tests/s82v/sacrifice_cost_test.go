package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestVillageRitesEatsACreatureAndDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := effectstest.SeedCreature(g, "Doomed Traveler", me.ID)
	for i := 0; i < 4; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{
			InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery",
			Owner: me.ID, Controller: me.ID,
		})
	}

	handBefore := me.Hand.Size()
	effectstest.CastWithSacrifice(t, g, "Village Rites", "Instant", effectstest.VillageRitesOracle, victim)

	// The creature is gone the moment the spell is CAST — before it
	// resolves, before priority is even passed.
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("victim still on the battlefield after the cast; the sacrifice is a cost, not an effect")
	}
	if !me.Graveyard.Contains(victim) {
		t.Error("victim not in the graveyard")
	}

	effectstest.PassPriorityAroundTable(t, g)

	// handBefore is sampled before the spell is seeded into hand, so
	// the spell's arrival and its departure to the graveyard cancel:
	// the delta is exactly the cards drawn.
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

func TestAltarsReapEatsACreatureAndDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := effectstest.SeedCreature(g, "Doomed Traveler", me.ID)
	for i := 0; i < 4; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{
			InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery",
			Owner: me.ID, Controller: me.ID,
		})
	}

	handBefore := me.Hand.Size()
	effectstest.CastWithSacrifice(t, g, "Altar's Reap", "Instant", effectstest.AltarsReapOracle, victim)
	effectstest.PassPriorityAroundTable(t, g)

	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Error("victim still on the battlefield")
	}
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

// TestDeadlyDisputeAcceptsAnArtifact is the point of the wider clause:
// "an artifact or creature" must admit a non-creature artifact, which
// is how the card is actually cast (eat a Treasure, get one back).
func TestDeadlyDisputeAcceptsAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.SeedPermanentFor(g, me.ID, "Ornithopter Husk", "Artifact")
	for i := 0; i < 4; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{
			InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery",
			Owner: me.ID, Controller: me.ID,
		})
	}

	handBefore := me.Hand.Size()
	effectstest.CastWithSacrifice(t, g, "Deadly Dispute", "Instant", effectstest.DeadlyDisputeOracle, rock)
	effectstest.PassPriorityAroundTable(t, g)

	if _, ok := effectstest.BattlefieldCard(g, rock); ok {
		t.Error("artifact still on the battlefield")
	}
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); got != 1 {
		t.Errorf("%d Treasure tokens, want 1", got)
	}
}

// TestVillageRitesDrainTriggersDuringTheCast is the ordering claim the
// card comments make, and the reason additional costs must be costs.
//
// The direct evidence: while Village Rites is still sitting unresolved
// on the stack, Blood Artist's trigger already exists and is asking for
// a target. The creature therefore died during the cast, not on
// resolution, so the drain is above the spell and settles first. An
// implementation that sacrificed in OnResolve could not produce that
// prompt at this point in the sequence.
func TestVillageRitesDrainTriggersDuringTheCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", effectstest.BloodArtistOracle, false)
	victim := effectstest.PushCatalogPermanent(g, me.ID, "Doomed Traveler", "Creature — Human Soldier", "", false)
	for i := 0; i < 4; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{
			InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery",
			Owner: me.ID, Controller: me.ID,
		})
	}
	handBefore, meBefore, oppBefore := me.Hand.Size(), me.Life, opp.Life

	spell := effectstest.CastWithSacrifice(t, g, "Village Rites", "Instant", effectstest.VillageRitesOracle, victim)

	if !g.Stack.Contains(spell) {
		t.Fatal("Village Rites is not on the stack after being cast")
	}
	if _, ok := effectstest.BattlefieldCard(g, victim); ok {
		t.Fatal("victim alive while the spell is on the stack; the cost was not paid at announce")
	}
	// handBefore was sampled before the spell was seeded, so its
	// arrival in hand and its departure to the stack cancel: a delta
	// of 0 here means nothing has been drawn yet.
	if got := me.Hand.Size() - handBefore; got != 0 {
		t.Fatalf("hand delta %d before resolution, want 0; cards were drawn too early", got)
	}

	// The trigger wants a target before it can go on the stack — and
	// it wants one NOW, with the spell still unresolved.
	for i := 0; i < 6 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("no Blood Artist target prompt; the death did not happen during the cast")
	}
	if !g.Stack.Contains(spell) {
		t.Fatal("Village Rites resolved before the drain trigger was even targeted")
	}

	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != oppBefore-1 {
		t.Errorf("target lost %d life, want 1", oppBefore-opp.Life)
	}
	if me.Life != meBefore+1 {
		t.Errorf("controller gained %d life, want 1", me.Life-meBefore)
	}
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

// TestVillageRitesUncastableWithNoCreature — an unpayable additional
// cost makes the announcement illegal (CR 601.2f) and it is rewound.
// The spell must stay in hand rather than resolving for free value.
func TestVillageRitesUncastableWithNoCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	id, err := effectstest.TryCastWithSacrifice(g, "Village Rites", "Instant", effectstest.VillageRitesOracle, nil)
	if err == nil {
		t.Fatal("Village Rites cast with no creature to sacrifice")
	}
	if !me.Hand.Contains(id) {
		t.Error("rejected spell left the caster's hand")
	}
	if g.Stack.Contains(id) {
		t.Error("rejected spell reached the stack")
	}
}

func TestAdditionalCostRejectsIllegalChoices(t *testing.T) {
	t.Run("opponent's creature", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		var them *game.Player
		for _, s := range g.Seats {
			if s.ID != me.ID {
				them = s
				break
			}
		}
		theirs := effectstest.SeedCreature(g, "Their Bear", them.ID)
		effectstest.SeedCreature(g, "My Bear", me.ID)

		if _, err := effectstest.TryCastWithSacrifice(g, "Village Rites", "Instant",
			effectstest.VillageRitesOracle, []uuid.UUID{theirs}); err == nil {
			t.Fatal("sacrificed an opponent's creature to my own cost")
		}
		if _, ok := effectstest.BattlefieldCard(g, theirs); !ok {
			t.Error("opponent's creature died to a rejected cast")
		}
	})

	t.Run("not a creature", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		rock := effectstest.SeedPermanentFor(g, me.ID, "Sol Ring", "Artifact")

		if _, err := effectstest.TryCastWithSacrifice(g, "Village Rites", "Instant",
			effectstest.VillageRitesOracle, []uuid.UUID{rock}); err == nil {
			t.Fatal("Village Rites ate an artifact; its clause is creatures only")
		}
		if _, ok := effectstest.BattlefieldCard(g, rock); !ok {
			t.Error("artifact died to a rejected cast")
		}
	})

	t.Run("two creatures for a one-permanent clause", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		a := effectstest.SeedCreature(g, "Bear One", me.ID)
		b := effectstest.SeedCreature(g, "Bear Two", me.ID)

		if _, err := effectstest.TryCastWithSacrifice(g, "Village Rites", "Instant",
			effectstest.VillageRitesOracle, []uuid.UUID{a, b}); err == nil {
			t.Fatal("paid a one-permanent clause with two permanents")
		}
		if _, ok := effectstest.BattlefieldCard(g, a); !ok {
			t.Error("first creature died to a rejected cast")
		}
		if _, ok := effectstest.BattlefieldCard(g, b); !ok {
			t.Error("second creature died to a rejected cast")
		}
	})
}

// TestSacrificeIDsRejectedOnCardWithoutTheClause keeps a client bug
// loud: a sacrifice offered for a spell that charges none is an error,
// not something to silently drop. Divination draws two either way, so
// dropping it would look like it worked.
func TestSacrificeIDsRejectedOnCardWithoutTheClause(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.SeedCreature(g, "Bear", me.ID)

	if _, err := effectstest.TryCastWithSacrifice(g, "Divination", "Sorcery",
		"273b339c-964b-4a18-8eb5-ceb8abcdfd9e", []uuid.UUID{bear}); err == nil {
		t.Fatal("Divination accepted a sacrifice it does not charge")
	}
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("creature died to a rejected cast")
	}
}

// TestAdditionalCostViewOffersOnlyYourPermanents is the client's half:
// the picker's candidate list must be stamped on the caster's own hand
// card, filtered to permanents they control, and absent from an
// opponent's view of the same card.
func TestAdditionalCostViewOffersOnlyYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var them *game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID {
			them = s
			break
		}
	}
	mine := effectstest.SeedCreature(g, "My Bear", me.ID)
	effectstest.SeedCreature(g, "Their Bear", them.ID)

	spell := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: spell, Name: "Village Rites", TypeLine: "Instant",
		OracleID: effectstest.VillageRitesOracle, Owner: me.ID, Controller: me.ID,
		// Drawn cards are known to their owner; a seeded one has to
		// say so, or the view redacts the clause under test (#95).
		KnownBy: map[uuid.UUID]bool{me.ID: true},
	})

	view := protocolViewFor(t, g, me.ID)
	ac := additionalCostOfHandCard(t, view, me.ID, spell)
	if ac.Label == "" {
		t.Error("no label; the picker has no banner copy")
	}
	if ac.SacrificeOptions == nil {
		t.Fatal("no sacrifice options; the picker would have nothing to show")
	}
	got := ac.SacrificeOptions.Cards
	if len(got) != 1 || got[0] != mine.String() {
		t.Errorf("sacrifice options %v, want just %v (mine)", got, mine)
	}
	if len(ac.SacrificeOptions.Players) != 0 {
		t.Error("options list players; a sacrifice cost takes a permanent")
	}
}

// protocolViewFor renders the game as one seat sees it.
func protocolViewFor(t *testing.T, g *game.Game, viewer uuid.UUID) protocol.GameView {
	t.Helper()
	return protocol.ViewOfGameFor(g, viewer.String())
}

// additionalCostOfHandCard pulls the additional-cost clause off one
// card in one seat's hand, failing rather than returning a zero value
// so a missing stamp can't pass as an empty one.
func additionalCostOfHandCard(t *testing.T, v protocol.GameView, owner, card uuid.UUID) protocol.AdditionalCostView {
	t.Helper()
	for _, seat := range v.Seats {
		if seat.ID != owner.String() {
			continue
		}
		for _, c := range seat.Hand.Cards {
			if c.InstanceID != card.String() {
				continue
			}
			if c.AdditionalCost == nil {
				t.Fatalf("card %s carries no additional_cost", card)
			}
			return *c.AdditionalCost
		}
		t.Fatalf("card %s not in seat %s's hand view", card, owner)
	}
	t.Fatalf("seat %s not in the view", owner)
	return protocol.AdditionalCostView{}
}
