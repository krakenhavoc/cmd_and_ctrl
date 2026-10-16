package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// karn_liberated_test.go — Karn Liberated's −14 (#3856, CR 727, ADR
// 0200) through the real card: the +4 and the −3 link what they exile
// to this Karn object, and the −14 restarts the game with the non-Aura
// permanent cards among them exempt, then puts those onto the
// battlefield under Karn's controller once the mulligans are over. The
// +4 and −3 on their own are superfriends_batch2_test.go's.

// karnKeepEveryHand keeps each opening hand in the order the mulligan
// window asks for them.
func karnKeepEveryHand(t *testing.T, g *game.Game) {
	t.Helper()
	for range len(g.Seats) {
		seat := g.MulliganDecider()
		if seat < 0 {
			return
		}
		if err := g.KeepHand(g.Seats[seat].ID); err != nil {
			t.Fatalf("KeepHand seat %d: %v", seat, err)
		}
	}
}

// karnActivate activates ability `idx` of Karn and lets it resolve.
func karnActivate(t *testing.T, g *game.Game, owner, karn uuid.UUID, idx int, targets ...game.TargetRef) {
	t.Helper()
	if err := g.ActivateCatalogAbility(owner, karn, idx, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("Karn ability %d: %v", idx, err)
	}
	passPriorityAroundTable(t, g)
}

// nextKarnTurn runs the table round to `seat`'s next main phase. Hands
// are emptied first so no cleanup step stops to ask for a discard.
func nextKarnTurn(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for _, p := range g.Seats {
		p.Hand.Cards = nil
	}
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	advanceToMainOf(t, g, seat)
}

// zoneOfNamed finds the one card named `name` anywhere a card can be,
// and the zone it is in.
func zoneOfNamed(g *game.Game, name string) (*game.Card, game.ZoneKind) {
	zones := []*game.Zone{g.Battlefield, g.Exile, g.Stack}
	for _, p := range g.Seats {
		zones = append(zones, p.Library, p.Hand, p.Graveyard, p.Command)
	}
	for _, z := range zones {
		for i := range z.Cards {
			if z.Cards[i].Name == name {
				return &z.Cards[i], z.Kind
			}
		}
	}
	return nil, ""
}

func TestKarnMinusFourteenRestartsTheGameAroundWhatHeExiled(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	them := g.Seats[(seat+1)%len(g.Seats)]
	karn := pushWalkerForTest(g, owner.ID, "Karn Liberated", karnLiberatedOracle, 20)
	bear := pushCreatureToBattlefieldForTest(g, them.ID, "Their Bear")
	aura := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: aura, Name: "Their Aura", TypeLine: "Enchantment — Aura",
		Owner: them.ID, Controller: them.ID,
	})
	// Exiled by something that is not Karn: not "exiled with Karn".
	g.Exile.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Elsewhere Exiled", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Owner: them.ID, Controller: them.ID,
	})
	advanceToMainOf(t, g, seat)

	// Three turns of Karn: −3 the bear, −3 the Aura, +4 an instant.
	karnActivate(t, g, owner.ID, karn, 1, game.TargetRef{Kind: game.TargetCard, ID: bear})
	link := game.PermissionCardRef{ID: karn}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == karn {
			link.Epoch = c.ObjectEpoch
		}
	}
	if got := g.ExiledWithForEffect(link); len(got) != 1 || got[0].InstanceID != bear {
		t.Fatalf("the −3 did not link the bear to Karn: %v", got)
	}
	nextKarnTurn(t, g, seat)
	karnActivate(t, g, owner.ID, karn, 1, game.TargetRef{Kind: game.TargetCard, ID: aura})
	nextKarnTurn(t, g, seat)
	them = g.Seats[(seat+1)%len(g.Seats)]
	them.Hand.Cards = nil
	loss := pushHandCardWithManaCost(them, "Their Loss", "Instant", "{U}")
	karnActivate(t, g, owner.ID, karn, 0, game.TargetRef{Kind: game.TargetPlayer, ID: them.ID})
	pick := chooseCardsChoiceFor(g, them.ID)
	if pick == nil {
		t.Fatal("the +4 asked nothing")
	}
	if err := g.ResolveChooseCards(pick.ID, them.ID, []uuid.UUID{loss}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := g.ExiledWithForEffect(link); len(got) != 3 {
		t.Fatalf("Karn's exile record has %d cards, want the bear, the Aura and the instant", len(got))
	}

	// The −14.
	nextKarnTurn(t, g, seat)
	if l := loyaltyOf(g, karn); l < 14 {
		t.Fatalf("Karn has %d loyalty, the test needs 14", l)
	}
	karnActivate(t, g, owner.ID, karn, 2)

	if !g.MulligansOpen || g.StartingSeat != seat || g.Turn.ActiveSeat != seat || g.Turn.Seq != 1 {
		t.Fatalf("no new game started by Karn's controller: mulligans %v starting %d active %d turn %d",
			g.MulligansOpen, g.StartingSeat, g.Turn.ActiveSeat, g.Turn.Seq)
	}
	if c, z := zoneOfNamed(g, "Their Bear"); c == nil || z != game.ZoneExile {
		t.Fatalf("the bear is in %s, want left in exile through the mulligans", z)
	}
	restarts := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventGameRestarted {
			restarts++
			if ev.Actor != owner.ID || ev.Label != "Karn Liberated" || ev.Amount != 1 {
				t.Errorf("EventGameRestarted = %+v", ev)
			}
		}
	}
	if restarts != 1 {
		t.Errorf("%d restart events, want 1", restarts)
	}

	karnKeepEveryHand(t, g)
	c, z := zoneOfNamed(g, "Their Bear")
	if c == nil || z != game.ZoneBattlefield {
		t.Fatalf("the bear is in %s, want the battlefield", z)
	}
	if c.Controller != owner.ID || c.Owner != them.ID {
		t.Errorf("the bear is controlled by %s and owned by %s, want Karn's controller and its owner", c.Controller, c.Owner)
	}
	// The Aura and the instant were exiled with Karn but are not
	// non-Aura permanent cards; the creature exiled by something else
	// was never his. All three are back in their owner's deck, and so
	// is Karn himself.
	for _, name := range []string{"Their Aura", "Their Loss", "Elsewhere Exiled"} {
		c, z := zoneOfNamed(g, name)
		if c == nil || (z != game.ZoneLibrary && z != game.ZoneHand) || c.Owner != them.ID {
			t.Errorf("%s is in %s, want back in its owner's deck", name, z)
		}
	}
	if c, z := zoneOfNamed(g, "Karn Liberated"); c == nil || (z != game.ZoneLibrary && z != game.ZoneHand) {
		t.Errorf("Karn is in %s, want back in his owner's deck", z)
	}
	if g.Turn.ActiveSeat != seat || g.Turn.Step != game.StepUpkeep {
		t.Errorf("first turn at seat %d %s, want Karn's controller's upkeep", g.Turn.ActiveSeat, g.Turn.Step)
	}
}

// The −14 is offered only with fourteen loyalty to pay.
func TestKarnMinusFourteenNeedsFourteenLoyalty(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	karn := pushWalkerForTest(g, owner.ID, "Karn Liberated", karnLiberatedOracle, 13)
	advanceToMainOf(t, g, seat)
	if err := g.ActivateCatalogAbility(owner.ID, karn, 2, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the −14 was activated with 13 loyalty")
	}
	if g.MulligansOpen {
		t.Error("a refused −14 restarted the game")
	}
}
