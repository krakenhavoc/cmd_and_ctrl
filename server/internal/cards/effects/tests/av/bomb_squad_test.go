package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func bsPushSquad(g *game.Game, controller uuid.UUID) uuid.UUID {
	return effectstest.ApaPush(g, controller, controller, effectstest.StCard("Bomb Squad", effectstest.BsBombSquad, "Creature — Dwarf", 1, 1))
}

func bsPushBear(g *game.Game, controller uuid.UUID, fuses int) uuid.UUID {
	id := effectstest.ApaPush(g, controller, controller, effectstest.StCard("Bear", "", "Creature — Bear", 2, 2))
	if fuses > 0 {
		g.WithWriteLock(func() {
			effectstest.FindBattlefieldCardForTest(g, id).Counters = map[string]int{"fuse": fuses}
		})
	}
	return id
}

func bsFuses(g *game.Game, id uuid.UUID) int {
	c := effectstest.FindBattlefieldCardForTest(g, id)
	if c == nil {
		return -1
	}
	return c.Counters["fuse"]
}

func bsLightFuse(t *testing.T, g *game.Game, p *game.Player, squad, target uuid.UUID) {
	t.Helper()
	effectstest.ApaActivate(t, g, p, squad, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
}

// "{T}: Put a fuse counter on target creature."
func TestBombSquadTapAbilityLightsAFuse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	bear := bsPushBear(g, bob.ID, 0)
	squad := bsPushSquad(g, me.ID)
	bsLightFuse(t, g, me, squad, bear)
	if n := bsFuses(g, bear); n != 1 {
		t.Fatalf("bear has %d fuse counters, want 1", n)
	}
}

// "At the beginning of your upkeep, put a fuse counter on each creature
// with a fuse counter on it" — anyone's, and only those.
func TestBombSquadUpkeepGrowsEveryLitFuse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	mine := bsPushBear(g, me.ID, 1)
	theirs := bsPushBear(g, bob.ID, 2)
	unlit := bsPushBear(g, bob.ID, 0)
	bsPushSquad(g, me.ID)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if a, b, c := bsFuses(g, mine), bsFuses(g, theirs), bsFuses(g, unlit); a != 2 || b != 3 || c != 0 {
		t.Fatalf("fuses after the upkeep: %d, %d, %d; want 2, 3, 0", a, b, c)
	}
}

// The fourth fuse counter: the counters come off, the creature is
// destroyed and it deals 4 damage to its controller.
func TestBombSquadDetonatesACreatureAtFourFuses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	bear := bsPushBear(g, bob.ID, 3)
	squad := bsPushSquad(g, me.ID)
	if n := effectstest.StStateItems(g, squad, "Bomb Squad — remove"); n != 0 {
		t.Fatalf("three fuse counters triggered Bomb Squad (%d items)", n)
	}
	life, mine := bob.Life, me.Life
	bsLightFuse(t, g, me, squad, bear)
	if effects.OnBattlefield(g, bear) {
		t.Fatal("the bear survived its fourth fuse counter")
	}
	if bob.Life != life-4 {
		t.Fatalf("the bear's controller is at %d, want %d", bob.Life, life-4)
	}
	if me.Life != mine {
		t.Fatalf("Bomb Squad's controller lost life (%d -> %d)", mine, me.Life)
	}
}

// The state is per creature: two creatures that reach four fuse counters
// in one upkeep each trigger the ability, and each deals 4 damage to its
// own controller.
func TestBombSquadTriggersForEachCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	mine := bsPushBear(g, me.ID, 3)
	theirs := bsPushBear(g, bob.ID, 3)
	squad := bsPushSquad(g, me.ID)
	myLife, bobLife := me.Life, bob.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.StStateItems(g, squad, "Bomb Squad"); n != 0 {
		t.Fatalf("%d Bomb Squad items still waiting", n)
	}
	if effects.OnBattlefield(g, mine) || effects.OnBattlefield(g, theirs) {
		t.Fatalf("a creature with four fuse counters survived (mine %v, theirs %v)", effects.OnBattlefield(g, mine), effects.OnBattlefield(g, theirs))
	}
	if me.Life != myLife-4 || bob.Life != bobLife-4 {
		t.Fatalf("life: me %d (want %d), bob %d (want %d)", me.Life, myLife-4, bob.Life, bobLife-4)
	}
}

// The card's ruling: with two Bomb Squads both abilities trigger, and
// both deal 4 damage even though the first destroys the creature.
func TestTwoBombSquadsBothDetonate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	bear := bsPushBear(g, bob.ID, 3)
	squad := bsPushSquad(g, me.ID)
	bsPushSquad(g, me.ID)
	life := bob.Life
	bsLightFuse(t, g, me, squad, bear)
	// Two abilities from two sources trigger at once: their controller
	// orders them (CR 603.3b).
	if !effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID) {
		t.Fatal("no ordering prompt: only one Bomb Squad triggered")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, bear) {
		t.Fatal("the bear survived")
	}
	if bob.Life != life-8 {
		t.Fatalf("bob is at %d, want %d: both Bomb Squads deal their 4", bob.Life, life-8)
	}
}

// The card's ruling: a creature that regenerates still loses its fuse
// counters and still deals the 4 damage.
func TestBombSquadRegeneratedCreatureStillDealsTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	bear := bsPushBear(g, bob.ID, 3)
	squad := bsPushSquad(g, me.ID)
	g.WithWriteLock(func() {
		if err := g.RegenerateForEffect(bear); err != nil {
			t.Fatalf("regenerate: %v", err)
		}
	})
	life := bob.Life
	bsLightFuse(t, g, me, squad, bear)
	if !effects.OnBattlefield(g, bear) {
		t.Fatal("the regenerated bear left the battlefield")
	}
	if n := bsFuses(g, bear); n != 0 {
		t.Fatalf("the regenerated bear kept %d fuse counters, want 0", n)
	}
	if bob.Life != life-4 {
		t.Fatalf("bob is at %d, want %d", bob.Life, life-4)
	}
	if n := effectstest.StStateItems(g, squad, "Bomb Squad"); n != 0 {
		t.Fatalf("Bomb Squad triggered again for a creature with no fuse counters (%d items)", n)
	}
}

// A creature that leaves before the ability resolves is not destroyed
// again, and still deals the 4 damage, as it last existed (CR 608.2h).
func TestBombSquadCreatureGoneInResponseStillDealsTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bob := g.Seats[0], g.Seats[1]
	bear := bsPushBear(g, bob.ID, 4)
	squad := bsPushSquad(g, me.ID) // its entry is an event, so the board is asked
	if n := effectstest.StStateItems(g, squad, "Bomb Squad — remove"); n != 1 {
		t.Fatalf("%d Bomb Squad items for a creature with four fuse counters, want 1", n)
	}
	effectstest.StDestroy(t, g, bear)
	life := bob.Life
	effectstest.PassPriorityAroundTable(t, g)
	if bob.Life != life-4 {
		t.Fatalf("bob is at %d, want %d: the departed bear still deals its 4", bob.Life, life-4)
	}
}
