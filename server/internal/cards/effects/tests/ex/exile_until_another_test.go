package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestUnstableAmuletEntersWithTwoEnergy(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	effectstest.SetEnergy(t, g, me, 0)
	effectstest.CastCatalogSpell(t, g, "Unstable Amulet", "Artifact", effectstest.EuaUnstableAmulet, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EnergyOf(me); got != 2 {
		t.Errorf("energy after the Amulet entered = %d, want 2", got)
	}
}

func TestUnstableAmuletWindowLastsUntilItExilesAnother(t *testing.T) {
	g, me, _, amulet := effectstest.EuaAmulet(t, 2)
	first := effectstest.EuaTop(g, me, "First", "Instant", "{R}")
	effectstest.EuaActivate(t, g, me, amulet)
	if !g.Exile.Contains(first) || !effectstest.EuaPlayable(g, me, first) {
		t.Fatal("the first activation did not exile a playable card")
	}
	if got := effectstest.EnergyOf(me); got != 2 {
		t.Errorf("energy after one activation = %d, want 2", got)
	}
	// A turn boundary does not close it.
	effectstest.AdvanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	if !effectstest.EuaPlayable(g, me, first) {
		t.Fatal("the window closed at a turn boundary")
	}

	second := effectstest.EuaTop(g, me, "Second", "Instant", "{R}")
	effectstest.EuaActivate(t, g, me, amulet)
	if effectstest.EuaPlayable(g, me, first) {
		t.Error("the first card is still playable after the Amulet exiled another")
	}
	if !effectstest.EuaPlayable(g, me, second) {
		t.Error("the newest card is not playable")
	}
}

// Owner answer 2, first half: the Amulet leaving the battlefield leaves
// the newest card playable for as long as it stays exiled (ruling
// 2024-06-07), and a new Amulet's exile does not close it (CR 400.7).
func TestUnstableAmuletLeavingKeepsTheNewestCardPlayable(t *testing.T) {
	g, me, _, amulet := effectstest.EuaAmulet(t, 2)
	first := effectstest.EuaTop(g, me, "First", "Instant", "{R}")
	effectstest.EuaActivate(t, g, me, amulet)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(amulet); err != nil {
			t.Fatalf("destroy the Amulet: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EuaPlayable(g, me, first) {
		t.Fatal("the Amulet leaving closed the window")
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, g.Turn.ActiveSeat)
	if !effectstest.EuaPlayable(g, me, first) {
		t.Fatal("the window closed a turn after the Amulet left")
	}

	another := effectstest.EbPush(g, me, "Unstable Amulet", effectstest.EuaUnstableAmulet, "Artifact", 0, 0)
	second := effectstest.EuaTop(g, me, "Second", "Instant", "{R}")
	effectstest.EuaActivate(t, g, me, another)
	if !effectstest.EuaPlayable(g, me, first) {
		t.Error("another Amulet's exile closed the first Amulet's window")
	}
	if !effectstest.EuaPlayable(g, me, second) {
		t.Error("the new Amulet's card is not playable")
	}
}

// Owner answer 2, second half: an activation already on the stack when
// the Amulet left still closes the old window as it resolves (CR 607.2a:
// the source as it last existed).
func TestUnstableAmuletAbilityOnTheStackClosesTheWindowAfterItLeft(t *testing.T) {
	g, me, _, amulet := effectstest.EuaAmulet(t, 2)
	first := effectstest.EuaTop(g, me, "First", "Instant", "{R}")
	effectstest.EuaActivate(t, g, me, amulet)

	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, amulet).Tapped = false })
	second := effectstest.EuaTop(g, me, "Second", "Instant", "{R}")
	if err := g.ActivateCatalogAbility(me.ID, amulet, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(amulet); err != nil {
			t.Fatalf("destroy the Amulet in response: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(second) {
		t.Fatal("the ability did not exile a card after its source left")
	}
	if effectstest.EuaPlayable(g, me, first) {
		t.Error("the ability that resolved after the Amulet left did not close the first window")
	}
	if !effectstest.EuaPlayable(g, me, second) {
		t.Error("the card it exiled is not playable")
	}
}

// Playing the card from exile is a cast from anywhere other than the
// hand, so the Amulet's second ability pings each opponent; the play
// uses the card up (ruling: "You can't play it multiple times").
func TestUnstableAmuletCastFromExilePingsEachOpponent(t *testing.T) {
	g, me, _, amulet := effectstest.EuaAmulet(t, 1)
	bolt := effectstest.EuaTop(g, me, "Test Spark", "Instant", "{R}")
	effectstest.EuaActivate(t, g, me, amulet)

	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID] = p.Life
	}
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g, me, _, amulet := effectstest.EuaAmulet(t, 1)
	land := effectstest.EuaTop(g, me, "Mountain", "Basic Land — Mountain", "")
	effectstest.EuaActivate(t, g, me, amulet)
	before := g.LandDropsRemainingFor(me.ID)
	if before < 1 {
		t.Fatalf("setup: %d land drops left", before)
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the land from exile: %v", err)
	}
	if effectstest.FindBattlefieldCardForTest(g, land) == nil {
		t.Fatal("the land is not on the battlefield")
	}
	if got := g.LandDropsRemainingFor(me.ID); got != before-1 {
		t.Errorf("land drops = %d, want %d", got, before-1)
	}
}

func TestFuriousRiseExilesOnlyWithAFourPowerCreature(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	seat := g.Turn.ActiveSeat
	effectstest.EbPush(g, me, "Furious Rise", effectstest.EuaFuriousRise, "Enchantment", 0, 0)
	// The top card is seeded at the end step itself, so the draw step
	// of a later turn cannot take it first.
	effectstest.AdvanceToEndStepOf(t, g, seat)
	small := effectstest.EuaTop(g, me, "Small", "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(small) {
		t.Fatal("Furious Rise exiled with no creature of power 4 or greater")
	}

	effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Big Creature", "", 4, 4))
	effectstest.AdvanceToEndStepOf(t, g, (seat+1)%len(g.Seats))
	effectstest.AdvanceToEndStepOf(t, g, seat)
	first := effectstest.EuaTop(g, me, "First", "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(first) || !effectstest.EuaPlayable(g, me, first) {
		t.Fatal("Furious Rise did not exile a playable card")
	}

	// The next one closes the first window.
	effectstest.AdvanceToEndStepOf(t, g, (seat+1)%len(g.Seats))
	if !effectstest.EuaPlayable(g, me, first) {
		t.Fatal("the window closed before Furious Rise exiled another card")
	}
	effectstest.AdvanceToEndStepOf(t, g, seat)
	next := effectstest.EuaTop(g, me, "Next", "Instant", "{R}")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EuaPlayable(g, me, first) || !effectstest.EuaPlayable(g, me, next) {
		t.Error("the second end-step exile did not move the window to the newest card")
	}
}

// CR 603.4: the condition is checked again on resolution.
func TestFuriousRiseRechecksPowerOnResolution(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	seat := g.Turn.ActiveSeat
	effectstest.EbPush(g, me, "Furious Rise", effectstest.EuaFuriousRise, "Enchantment", 0, 0)
	big := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Big Creature", "", 4, 4))
	top := effectstest.EuaTop(g, me, "Top", "Instant", "{R}")
	effectstest.AdvanceToEndStepOf(t, g, seat)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(big); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
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
	g, me, _ := effectstest.P7Table(t)
	effectstest.EbPush(g, me, "Superior Foes of Spider-Man", effectstest.EuaSuperiorFoes, "Creature — Human Rogue Villain", 3, 3)

	first := effectstest.EuaTop(g, me, "First", "Instant", "{R}")
	euaCastBig(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(first) || !effectstest.EuaPlayable(g, me, first) {
		t.Fatal("a yes did not exile a playable card")
	}

	// A no exiles nothing, and so leaves the window open.
	keep := effectstest.EuaTop(g, me, "Kept", "Instant", "{R}")
	euaCastBig(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(keep) {
		t.Fatal("a no exiled the top card")
	}
	if !effectstest.EuaPlayable(g, me, first) {
		t.Fatal("declining closed the window")
	}

	// A second yes moves the window.
	euaCastBig(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EuaPlayable(g, me, first) || !effectstest.EuaPlayable(g, me, keep) {
		t.Error("the second yes did not move the window to the newest card")
	}
}
