package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// memory_vessel_test.go — #2559: every player plays the cards they
// exiled, and nobody plays from their hand, until the activator's next
// turn (Memory Vessel); and the same permission without the ban, until
// your next end step (Rocco, Street Chef).

const (
	memoryVesselOracle = "0179bc62-823e-46b9-b536-342904fedafc"
	roccoOracle        = "071f5b4b-6a1e-4dc3-9c08-b79713191811"
)

// vesselLibraryTop puts a card on top of p's library and returns it.
func vesselLibraryTop(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// vesselTable is seat 0's precombat main with every library topped by a
// Forest under a {0} artifact (the artifact is exiled first, both among
// the seven), seat 0's hand holding a Plains and a {0} artifact, and
// Memory Vessel activated and resolved.
type vesselTable struct {
	g                    *game.Game
	me, opp              *game.Player
	myLand, myArtifact   uuid.UUID
	oppLand              uuid.UUID
	handLand, handSpell  uuid.UUID
	exiledBefore, vessel int
	librariesBefore      map[uuid.UUID]int
}

func newVesselTable(t *testing.T) vesselTable {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	vt := vesselTable{g: g, me: me, opp: opp}
	for _, p := range g.Seats {
		land := vesselLibraryTop(p, "Forest", "Basic Land — Forest", "")
		art := vesselLibraryTop(p, "Bauble", "Artifact", "{0}")
		switch p {
		case me:
			vt.myLand, vt.myArtifact = land, art
		case opp:
			vt.oppLand = land
		}
	}
	vt.handLand = uuid.New()
	vt.handSpell = uuid.New()
	known := map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(game.Card{InstanceID: vt.handLand, Name: "Plains", TypeLine: "Basic Land — Plains", Owner: me.ID, Controller: me.ID, KnownBy: known})
	me.Hand.PushTop(game.Card{InstanceID: vt.handSpell, Name: "Trinket", TypeLine: "Artifact", ManaCost: "{0}", Owner: me.ID, Controller: me.ID, KnownBy: known})
	vessel := pushCatalogPermanent(g, me.ID, "Memory Vessel", "Artifact", memoryVesselOracle, false)
	vt.exiledBefore = g.Exile.Size()
	vt.librariesBefore = map[uuid.UUID]int{}
	for _, p := range g.Seats {
		vt.librariesBefore[p.ID] = p.Library.Size()
	}
	if err := g.ActivateCatalogAbility(me.ID, vessel, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Memory Vessel: %v", err)
	}
	if !g.Exile.Contains(vessel) {
		t.Fatal("Memory Vessel is not exiled by its own cost")
	}
	vt.vessel = 1
	passPriorityAroundTable(t, g)
	return vt
}

func permissionFor(g *game.Game, player, card uuid.UUID) *game.CastPermission {
	var perm *game.CastPermission
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		perm = g.CastPermissionForLocked(player, c, game.ZoneExile)
	})
	return perm
}

func TestMemoryVesselEachPlayerExilesSevenAndMayPlayOnlyTheirOwn(t *testing.T) {
	vt := newVesselTable(t)
	g, me, opp := vt.g, vt.me, vt.opp

	if got, want := g.Exile.Size()-vt.exiledBefore, 7*len(g.Seats)+vt.vessel; got != want {
		t.Fatalf("exile grew by %d, want %d (seven per player and the Vessel)", got, want)
	}
	for _, p := range g.Seats {
		if got, want := p.Library.Size(), vt.librariesBefore[p.ID]-7; got != want {
			t.Errorf("seat %s library = %d, want %d", p.Name, got, want)
		}
	}
	if permissionFor(g, me.ID, vt.myArtifact) == nil {
		t.Error("the activator may not play their own exiled card")
	}
	if permissionFor(g, opp.ID, vt.oppLand) == nil {
		t.Error("an opponent may not play the card they exiled")
	}
	if permissionFor(g, me.ID, vt.oppLand) != nil {
		t.Error("the activator may play a card an opponent exiled")
	}
	if permissionFor(g, opp.ID, vt.myArtifact) != nil {
		t.Error("an opponent may play a card the activator exiled")
	}
}

func TestMemoryVesselBansTheHandButNotTheExile(t *testing.T) {
	vt := newVesselTable(t)
	g, me := vt.g, vt.me

	err := g.CastSpell(me.ID, vt.handSpell, game.CastSpellParams{})
	var cant *game.CantCastError
	if !errors.As(err, &cant) || cant.Reason != "You can't play cards from your hand — Memory Vessel" {
		t.Fatalf("a spell cast from hand: err = %v, want the hand ban", err)
	}
	if err := g.CastSpell(me.ID, vt.handLand, game.CastSpellParams{}); !errors.Is(err, game.ErrCantPlayLand) {
		t.Fatalf("a land played from hand: err = %v, want ErrCantPlayLand", err)
	}
	if !me.Hand.Contains(vt.handLand) || !me.Hand.Contains(vt.handSpell) {
		t.Fatal("a refused play moved a card out of the hand")
	}

	if err := g.CastSpell(me.ID, vt.myArtifact, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("the exiled artifact: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(vt.myArtifact) {
		t.Error("the artifact cast from exile did not resolve onto the battlefield")
	}
	if err := g.CastSpell(me.ID, vt.myLand, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("the exiled land: %v", err)
	}
	if got := g.LandsPlayedThisTurnFor(me.ID); got != 1 {
		t.Errorf("lands played = %d, want 1: a land from exile is the turn's land play", got)
	}
}

func TestMemoryVesselLandFromExileNeedsALandDropLeft(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	first := vesselLibraryTop(me, "Forest", "Basic Land — Forest", "")
	second := vesselLibraryTop(me, "Island", "Basic Land — Island", "")
	vessel := pushCatalogPermanent(g, me.ID, "Memory Vessel", "Artifact", memoryVesselOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, vessel, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("first land from exile: %v", err)
	}
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{FromZone: "exile"}); !errors.Is(err, game.ErrLandDropUnavailable) {
		t.Fatalf("second land from exile: err = %v, want ErrLandDropUnavailable", err)
	}
}

func TestMemoryVesselEndsAsTheActivatorsNextTurnBegins(t *testing.T) {
	vt := newVesselTable(t)
	g, me, opp := vt.g, vt.me, vt.opp

	// On the next player's turn: their own land from exile, and still no
	// hand.
	passTurnsTo(t, g, 1)
	advanceToMain(t, g)
	oppHand := handCardForTest(opp, "Swamp", "Basic Land — Swamp", "")
	if err := g.CastSpell(opp.ID, oppHand, game.CastSpellParams{}); !errors.Is(err, game.ErrCantPlayLand) {
		t.Fatalf("an opponent's land from hand on their turn: err = %v, want the ban", err)
	}
	if err := g.CastSpell(opp.ID, vt.oppLand, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("an opponent's own land from exile on their turn: %v", err)
	}

	// The activator's next turn: both halves are over, the cards stay.
	passTurnsTo(t, g, 0)
	advanceToMain(t, g)
	if permissionFor(g, me.ID, vt.myLand) != nil {
		t.Error("the permission outlived the activator's next turn beginning")
	}
	if !g.Exile.Contains(vt.myLand) {
		t.Error("a card still exiled when the window closed left exile")
	}
	if err := g.CastSpell(me.ID, vt.handLand, game.CastSpellParams{}); err != nil {
		t.Fatalf("the hand ban outlived the activator's next turn beginning: %v", err)
	}
	if err := g.CastSpell(me.ID, vt.myArtifact, game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Error("an exiled card was cast after the window closed")
	}
}

// Activating abilities of cards in hand is not playing them, and
// neither is a card the activator's commander casts from the command
// zone: the ban is the hand, as printed.
func TestMemoryVesselBanIsOnlyTheHand(t *testing.T) {
	vt := newVesselTable(t)
	g, me := vt.g, vt.me
	var cast, land error
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(vt.handSpell)
		cast = g.CastGateLocked(me.ID, c, game.ZoneGraveyard, game.CastSpellParams{})
		land = g.LandPlayGateLocked(me.ID, game.Card{TypeLine: "Land"}, game.ZoneExile)
	})
	if cast != nil || land != nil {
		t.Errorf("the ban reached a zone other than the hand: cast %v, land %v", cast, land)
	}
}

func TestMemoryVesselBotIsOfferedTheExileAndNotTheHand(t *testing.T) {
	vt := newVesselTable(t)
	g, me := vt.g, vt.me
	if enumeratedFor(g, me.ID, vt.handLand) || enumeratedFor(g, me.ID, vt.handSpell) {
		t.Error("the enumerator offers a play from the banned hand")
	}
	if !enumeratedFor(g, me.ID, vt.myLand) || !enumeratedFor(g, me.ID, vt.myArtifact) {
		t.Error("the enumerator does not offer the exiled land and artifact")
	}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == legal.TypeCastSpell && m.Source == vt.oppLand {
			t.Error("the enumerator offers a card an opponent exiled")
		}
	}
}

func TestMemoryVesselViewShowsTheBanAndEveryonesGrant(t *testing.T) {
	vt := newVesselTable(t)
	g, me, opp := vt.g, vt.me, vt.opp
	v := protocol.ViewOfGameFor(g, opp.ID.String())
	for _, s := range v.Seats {
		if s.CantPlayFromHand != "You can't play cards from your hand — Memory Vessel" {
			t.Errorf("seat %s cant_play_from_hand = %q", s.Name, s.CantPlayFromHand)
		}
	}
	found := 0
	for _, c := range v.Exile.Cards {
		if c.InstanceID != vt.myArtifact.String() && c.InstanceID != vt.oppLand.String() {
			continue
		}
		found++
		ep := c.ExilePlay
		if ep == nil {
			t.Fatalf("%s carries no exile_play", c.Name)
		}
		if ep.Player != c.Owner {
			t.Errorf("%s exile_play.player = %s, want its owner %s", c.Name, ep.Player, c.Owner)
		}
		if ep.Until != game.PermissionUntilNextTurn || ep.UntilPlayer != me.ID.String() {
			t.Errorf("%s window = %q/%q, want next_turn of the activator", c.Name, ep.Until, ep.UntilPlayer)
		}
	}
	if found != 2 {
		t.Fatalf("found %d of the two exiled cards in the view", found)
	}
	mine := protocol.ViewOfGameFor(g, me.ID.String())
	for _, s := range mine.Seats {
		if s.ID != me.ID.String() {
			continue
		}
		for _, c := range s.Hand.Cards {
			if c.InstanceID != vt.handLand.String() && c.InstanceID != vt.handSpell.String() {
				continue
			}
			if c.CantCast == "" {
				t.Errorf("%s in the banned hand carries no cant_cast", c.Name)
			}
		}
	}
}

func TestMemoryVesselSurvivesARestore(t *testing.T) {
	vt := newVesselTable(t)
	g := restoreRoundTrip(t, vt.g, true)
	me := g.Seats[0]
	if err := g.CastSpell(me.ID, vt.handLand, game.CastSpellParams{}); !errors.Is(err, game.ErrCantPlayLand) {
		t.Errorf("after a restore the hand ban is gone: %v", err)
	}
	if permissionFor(g, me.ID, vt.myLand) == nil {
		t.Error("after a restore the exile permission is gone")
	}
}

// --- Rocco, Street Chef ---------------------------------------------

func TestRoccoEachPlayerMayPlayTheirCardUntilYourNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rocco := pushCatalogPermanent(g, me.ID, "Rocco, Street Chef", "Legendary Creature — Elf Druid", roccoOracle, false)
	oppCard := vesselLibraryTop(opp, "Bauble", "Artifact", "{0}")
	myCard := vesselLibraryTop(me, "Forest", "Basic Land — Forest", "")
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(oppCard) || !g.Exile.Contains(myCard) {
		t.Fatal("each player's top card is not in exile")
	}
	perm := permissionFor(g, opp.ID, oppCard)
	if perm == nil || permissionFor(g, me.ID, oppCard) != nil {
		t.Fatal("the opponent does not hold the only permission over their own card")
	}
	var until string
	g.ReadSnapshot(func() { until, _ = g.PermissionWindowLocked(perm) })
	if until != game.PermissionUntilNextEndStep {
		t.Errorf("window = %q, want next_end_step", until)
	}

	// On the opponent's turn they cast it, and Rocco's controller gets a
	// counter on a creature and a Food.
	passTurnsTo(t, g, 1)
	advanceToMain(t, g)
	foodBefore := countNamed(g, "Food")
	if err := g.CastSpell(opp.ID, oppCard, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("the opponent's exiled card: %v", err)
	}
	pickTriggerTarget(t, g, me.ID, rocco)
	passPriorityAroundTable(t, g)
	if got := countNamed(g, "Food"); got != foodBefore+1 {
		t.Errorf("there are %d Food, want %d", got, foodBefore+1)
	}
	var counters int
	var foodOwner uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == rocco {
				counters = c.Counters[game.CounterPlusOne]
			}
			if c.Name == "Food" {
				foodOwner = c.Controller
			}
		}
	})
	if counters != 1 || foodOwner != me.ID {
		t.Errorf("Rocco has %d +1/+1 counters and the Food is %s's, want 1 and Rocco's controller's", counters, foodOwner)
	}

	// Still open on Rocco's controller's next turn, before its end step.
	passTurnsTo(t, g, 0)
	advanceToMain(t, g)
	if permissionFor(g, me.ID, myCard) == nil {
		t.Fatal("the permission closed before Rocco's controller's next end step")
	}
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if permissionFor(g, me.ID, myCard) != nil {
		t.Error("the permission outlived Rocco's controller's next end step")
	}
}
