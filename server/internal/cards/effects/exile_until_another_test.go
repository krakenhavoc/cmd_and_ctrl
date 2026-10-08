package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_until_another_test.go — #2539, ADR 0066 amendment 2026-10-08:
// Unstable Amulet, Furious Rise and Superior Foes of Spider-Man, whose
// exiled card stays playable "until you exile another card with this".
// The engine half is game/exile_until_another_test.go.

const (
	euaUnstableAmulet = "f3d4f5ab-0ff4-48e1-8cf0-163b290804a5"
	euaFuriousRise    = "52b3a979-3b93-4441-9237-0c8a754e5523"
	euaSuperiorFoes   = "a6a7af21-e52c-4f7f-a57f-bd365d075966"
)

// euaTop puts a fresh card on top of p's library and returns it.
func euaTop(g *game.Game, p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	})
	return id
}

// euaPlayable asks the one predicate the cast path, the view and the
// bot all ask: may p play card `id` from exile right now?
func euaPlayable(g *game.Game, p *game.Player, id uuid.UUID) bool {
	var ok bool
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.InstanceID == id {
				ok = g.CastPermissionForLocked(p.ID, c, game.ZoneExile) != nil
			}
		}
	})
	return ok
}

// euaAmulet is a catalog table with an untapped Unstable Amulet for the
// active seat, and enough energy for `activations` activations.
func euaAmulet(t *testing.T, activations int) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := p7Table(t)
	amulet := ebPush(g, me, "Unstable Amulet", euaUnstableAmulet, "Artifact", 0, 0)
	setEnergy(t, g, me, 2*activations)
	return g, me, opp, amulet
}

// euaActivate untaps the Amulet, activates it and resolves the ability.
func euaActivate(t *testing.T, g *game.Game, me *game.Player, amulet uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, amulet).Tapped = false })
	p7Activate(t, g, me, amulet, 0, game.ActivateAbilityParams{})
}

func TestUnstableAmuletEntersWithTwoEnergy(t *testing.T) {
	g, me, _ := p7Table(t)
	setEnergy(t, g, me, 0)
	castCatalogSpell(t, g, "Unstable Amulet", "Artifact", euaUnstableAmulet, nil)
	passPriorityAroundTable(t, g)
	if got := energyOf(me); got != 2 {
		t.Errorf("energy after the Amulet entered = %d, want 2", got)
	}
}

func TestUnstableAmuletWindowLastsUntilItExilesAnother(t *testing.T) {
	g, me, _, amulet := euaAmulet(t, 2)
	first := euaTop(g, me, "First", "Instant", "{R}")
	euaActivate(t, g, me, amulet)
	if !g.Exile.Contains(first) || !euaPlayable(g, me, first) {
		t.Fatal("the first activation did not exile a playable card")
	}
	if got := energyOf(me); got != 2 {
		t.Errorf("energy after one activation = %d, want 2", got)
	}
	// A turn boundary does not close it.
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	if !euaPlayable(g, me, first) {
		t.Fatal("the window closed at a turn boundary")
	}

	second := euaTop(g, me, "Second", "Instant", "{R}")
	euaActivate(t, g, me, amulet)
	if euaPlayable(g, me, first) {
		t.Error("the first card is still playable after the Amulet exiled another")
	}
	if !euaPlayable(g, me, second) {
		t.Error("the newest card is not playable")
	}
}

// Owner answer 2, first half: the Amulet leaving the battlefield leaves
// the newest card playable for as long as it stays exiled (ruling
// 2024-06-07), and a new Amulet's exile does not close it (CR 400.7).
func TestUnstableAmuletLeavingKeepsTheNewestCardPlayable(t *testing.T) {
	g, me, _, amulet := euaAmulet(t, 2)
	first := euaTop(g, me, "First", "Instant", "{R}")
	euaActivate(t, g, me, amulet)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(amulet); err != nil {
			t.Fatalf("destroy the Amulet: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !euaPlayable(g, me, first) {
		t.Fatal("the Amulet leaving closed the window")
	}
	advanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	if !euaPlayable(g, me, first) {
		t.Fatal("the window closed a turn after the Amulet left")
	}

	another := ebPush(g, me, "Unstable Amulet", euaUnstableAmulet, "Artifact", 0, 0)
	second := euaTop(g, me, "Second", "Instant", "{R}")
	euaActivate(t, g, me, another)
	if !euaPlayable(g, me, first) {
		t.Error("another Amulet's exile closed the first Amulet's window")
	}
	if !euaPlayable(g, me, second) {
		t.Error("the new Amulet's card is not playable")
	}
}

// Owner answer 2, second half: an activation already on the stack when
// the Amulet left still closes the old window as it resolves (CR 607.2a:
// the source as it last existed).
func TestUnstableAmuletAbilityOnTheStackClosesTheWindowAfterItLeft(t *testing.T) {
	g, me, _, amulet := euaAmulet(t, 2)
	first := euaTop(g, me, "First", "Instant", "{R}")
	euaActivate(t, g, me, amulet)

	g.WithWriteLock(func() { findBattlefieldCardForTest(g, amulet).Tapped = false })
	second := euaTop(g, me, "Second", "Instant", "{R}")
	if err := g.ActivateCatalogAbility(me.ID, amulet, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(amulet); err != nil {
			t.Fatalf("destroy the Amulet in response: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(second) {
		t.Fatal("the ability did not exile a card after its source left")
	}
	if euaPlayable(g, me, first) {
		t.Error("the ability that resolved after the Amulet left did not close the first window")
	}
	if !euaPlayable(g, me, second) {
		t.Error("the card it exiled is not playable")
	}
}

// Playing the card from exile is a cast from anywhere other than the
// hand, so the Amulet's second ability pings each opponent; the play
// uses the card up (ruling: "You can't play it multiple times").
func TestUnstableAmuletCastFromExilePingsEachOpponent(t *testing.T) {
	g, me, _, amulet := euaAmulet(t, 1)
	bolt := euaTop(g, me, "Test Spark", "Instant", "{R}")
	euaActivate(t, g, me, amulet)

	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID] = p.Life
	}
	floatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := life[p.ID] - 1
		if p.ID == me.ID {
			want = life[p.ID]
		}
		if p.Life != want {
			t.Errorf("%s life = %d, want %d", p.Name, p.Life, want)
		}
	}
	if g.Exile.Contains(bolt) {
		t.Error("the card is still in exile after it was cast")
	}
}

// A land exiled this way is played under the usual rules: it takes the
// turn's land play (ruling 2024-06-07).
func TestUnstableAmuletLandIsPlayedWithALandDrop(t *testing.T) {
	g, me, _, amulet := euaAmulet(t, 1)
	land := euaTop(g, me, "Mountain", "Basic Land — Mountain", "")
	euaActivate(t, g, me, amulet)
	before := g.LandDropsRemainingFor(me.ID)
	if before < 1 {
		t.Fatalf("setup: %d land drops left", before)
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the land from exile: %v", err)
	}
	if findBattlefieldCardForTest(g, land) == nil {
		t.Fatal("the land is not on the battlefield")
	}
	if got := g.LandDropsRemainingFor(me.ID); got != before-1 {
		t.Errorf("land drops = %d, want %d", got, before-1)
	}
}

func TestFuriousRiseExilesOnlyWithAFourPowerCreature(t *testing.T) {
	g, me, _ := p7Table(t)
	seat := g.Turn.ActiveSeat
	ebPush(g, me, "Furious Rise", euaFuriousRise, "Enchantment", 0, 0)
	// The top card is seeded at the end step itself, so the draw step
	// of a later turn cannot take it first.
	advanceToEndStepOf(t, g, seat)
	small := euaTop(g, me, "Small", "Instant", "{R}")
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(small) {
		t.Fatal("Furious Rise exiled with no creature of power 4 or greater")
	}

	apaPush(g, me.ID, me.ID, apaCreature("Big Creature", "", 4, 4))
	advanceToEndStepOf(t, g, (seat+1)%len(g.Seats))
	advanceToEndStepOf(t, g, seat)
	first := euaTop(g, me, "First", "Instant", "{R}")
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(first) || !euaPlayable(g, me, first) {
		t.Fatal("Furious Rise did not exile a playable card")
	}

	// The next one closes the first window.
	advanceToEndStepOf(t, g, (seat+1)%len(g.Seats))
	if !euaPlayable(g, me, first) {
		t.Fatal("the window closed before Furious Rise exiled another card")
	}
	advanceToEndStepOf(t, g, seat)
	next := euaTop(g, me, "Next", "Instant", "{R}")
	passPriorityAroundTable(t, g)
	if euaPlayable(g, me, first) || !euaPlayable(g, me, next) {
		t.Error("the second end-step exile did not move the window to the newest card")
	}
}

// CR 603.4: the condition is checked again on resolution.
func TestFuriousRiseRechecksPowerOnResolution(t *testing.T) {
	g, me, _ := p7Table(t)
	seat := g.Turn.ActiveSeat
	ebPush(g, me, "Furious Rise", euaFuriousRise, "Enchantment", 0, 0)
	big := apaPush(g, me.ID, me.ID, apaCreature("Big Creature", "", 4, 4))
	top := euaTop(g, me, "Top", "Instant", "{R}")
	advanceToEndStepOf(t, g, seat)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(big); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(top) {
		t.Error("Furious Rise exiled although the big creature died before it resolved")
	}
}

// euaCastBig casts a mana value 4 sorcery from p's hand.
func euaCastBig(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Big Thing", TypeLine: "Sorcery", ManaCost: "{2}{R}{R}", Owner: p.ID, Controller: p.ID})
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Big Thing: %v", err)
	}
}

func TestSuperiorFoesExilesOnlyWhenYouSayYes(t *testing.T) {
	g, me, _ := p7Table(t)
	ebPush(g, me, "Superior Foes of Spider-Man", euaSuperiorFoes, "Creature — Human Rogue Villain", 3, 3)

	first := euaTop(g, me, "First", "Instant", "{R}")
	euaCastBig(t, g, me)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(first) || !euaPlayable(g, me, first) {
		t.Fatal("a yes did not exile a playable card")
	}

	// A no exiles nothing, and so leaves the window open.
	keep := euaTop(g, me, "Kept", "Instant", "{R}")
	euaCastBig(t, g, me)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(keep) {
		t.Fatal("a no exiled the top card")
	}
	if !euaPlayable(g, me, first) {
		t.Fatal("declining closed the window")
	}

	// A second yes moves the window.
	euaCastBig(t, g, me)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if euaPlayable(g, me, first) || !euaPlayable(g, me, keep) {
		t.Error("the second yes did not move the window to the newest card")
	}
}
