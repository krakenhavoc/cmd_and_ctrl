package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// altar_cutthroat_audit_test.go — the boundary cases behind the
// Full declarations on Phyrexian Altar and Zulaport Cutthroat (#2954).
// The ordinary paths are covered in altars_test.go,
// any_color_mana_test.go, aristocrats_test.go and s23_boardwipes_test.go.

// Every one of the five colours is on offer and each one lands in the
// pool as exactly that colour.
func TestPhyrexianAltarProducesEachOfTheFiveColours(t *testing.T) {
	for _, color := range []string{"W", "U", "B", "R", "G"} {
		t.Run(color, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			altar := seedPermanentWithOracle(g, me.ID, "Phyrexian Altar", "Artifact", phyrexianAltarOracle)
			fodder := seedCreature(g, "Fodder", me.ID)
			if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
				t.Fatalf("ActivateManaAbility: %v", err)
			}
			pick := riderLatestManaPick(g, me.ID)
			if pick == nil {
				t.Fatal("no colour pick queued")
			}
			offered := map[string]bool{}
			for _, o := range pick.ColorOptions {
				offered[o] = true
			}
			if len(offered) != 5 || !offered[color] {
				t.Fatalf("offered %v, want all five colours including %s", pick.ColorOptions, color)
			}
			if err := g.ResolveManaChoice(pick.ID, me.ID, color); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
			if len(me.ManaPool) != 1 || me.ManaPool[0].Color != color {
				t.Fatalf("pool = %+v, want one %s", me.ManaPool, color)
			}
		})
	}
}

// The Altar has no tap symbol: it goes again as often as there are
// creatures, tapped or not, and each activation is one mana.
func TestPhyrexianAltarSacrificesRepeatedlyAndTappedCreaturesToo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	altar := seedPermanentWithOracle(g, me.ID, "Phyrexian Altar", "Artifact", phyrexianAltarOracle)
	victims := []uuid.UUID{seedCreature(g, "A", me.ID), seedCreature(g, "B", me.ID), seedCreature(g, "C", me.ID)}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victims[0] {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	for i, v := range victims {
		if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{v}}); err != nil {
			t.Fatalf("activation %d: %v", i+1, err)
		}
		if got := riderAnswerManaPicks(t, g, me.ID, "B"); got != 1 {
			t.Fatalf("activation %d queued %d picks, want 1", i+1, got)
		}
	}
	if len(me.ManaPool) != 3 {
		t.Errorf("pool = %d tokens, want 3", len(me.ManaPool))
	}
	for _, v := range victims {
		if g.Battlefield.Contains(v) {
			t.Errorf("creature %s survived its sacrifice", v)
		}
	}
	if card, ok := battlefieldCard(g, altar); !ok || card.Tapped {
		t.Errorf("Altar on battlefield=%v tapped=%v, want present and untapped", ok, card.Tapped)
	}
}

// "Sacrifice a creature": a noncreature (even another artifact), a
// creature somebody else controls and an empty selection are all
// refused, and a refused activation costs nothing.
func TestPhyrexianAltarRefusesWhatIsNotYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	altar := seedPermanentWithOracle(g, me.ID, "Phyrexian Altar", "Artifact", phyrexianAltarOracle)
	rock := seedPermanentWithOracle(g, me.ID, "Sol Ring", "Artifact", "")
	theirs := seedCreature(g, "Theirs", opp.ID)
	mine := seedCreature(g, "Mine", me.ID)

	for name, ids := range map[string][]uuid.UUID{
		"noncreature artifact": {rock},
		"opponent's creature":  {theirs},
		"nothing":              nil,
	} {
		if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: ids}); err == nil {
			t.Errorf("%s: activation accepted", name)
		}
	}
	if len(me.ManaPool) != 0 || riderLatestManaPick(g, me.ID) != nil {
		t.Error("a refused activation produced mana")
	}
	for _, id := range []uuid.UUID{rock, theirs, mine, altar} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("permanent %s left the battlefield on a refused activation", id)
		}
	}
}

// An Altar that is itself a creature (animated) may be sacrificed to
// its own cost: the printed cost is "a creature", not "another".
func TestPhyrexianAltarMaySacrificeItselfWhileACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	altar := seedPermanentWithOracle(g, me.ID, "Phyrexian Altar", "Artifact Creature — Construct", phyrexianAltarOracle)
	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{altar}}); err != nil {
		t.Fatalf("sacrificing an animated Altar to itself: %v", err)
	}
	if g.Battlefield.Contains(altar) {
		t.Error("the Altar is still on the battlefield after paying itself")
	}
	if riderAnswerManaPicks(t, g, me.ID, "R") != 1 || len(me.ManaPool) != 1 {
		t.Errorf("pool = %+v, want the one red mana the ability still makes", me.ManaPool)
	}
}

// Zulaport's own death triggers it.
func TestZulaportTriggersOnItsOwnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	z := pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportCutthroatOracle, false)
	meBefore := me.Life
	oppBefore := []int{g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(z) })
	passPriorityAroundTable(t, g)
	for i, b := range oppBefore {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("seat %d: %d -> %d, want -1", i+1, b, got)
		}
	}
	if me.Life != meBefore+1 {
		t.Errorf("controller life %d -> %d, want +1", meBefore, me.Life)
	}
}

// "Whenever ... dies" follows who controlled the creature as it died
// (CR 603.10a), not who owns it. A stolen Cutthroat works for the
// thief; a creature the Cutthroat's controller stole counts for it;
// and a creature stolen away from the Cutthroat's controller does not.
func TestZulaportFollowsControlNotOwnership(t *testing.T) {
	t.Run("stolen Cutthroat works for the thief only", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		z := pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportCutthroatOracle, false)
		mine := ctrlPushCreature(g, me.ID, "Mine")
		theirs := ctrlPushCreature(g, opp.ID, "Theirs")
		g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, z, opp.ID, game.IndefiniteDuration(), "theft") })

		meBefore, oppBefore := me.Life, opp.Life
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
		passPriorityAroundTable(t, g)
		if me.Life != meBefore || opp.Life != oppBefore {
			t.Errorf("the old controller's creature died: life %d/%d -> %d/%d, want no change", meBefore, oppBefore, me.Life, opp.Life)
		}

		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
		passPriorityAroundTable(t, g)
		if opp.Life != oppBefore+1 || me.Life != meBefore-1 {
			t.Errorf("the thief's creature died: opp %d -> %d (want +1), me %d -> %d (want -1)", oppBefore, opp.Life, meBefore, me.Life)
		}
	})
	t.Run("a stolen creature counts for its new controller", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportCutthroatOracle, false)
		loot := ctrlPushCreature(g, opp.ID, "Loot")
		g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, loot, me.ID, game.IndefiniteDuration(), "theft") })

		meBefore, oppBefore := me.Life, opp.Life
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(loot) })
		passPriorityAroundTable(t, g)
		if me.Life != meBefore+1 || opp.Life != oppBefore-1 {
			t.Errorf("stolen creature died: me %d -> %d (want +1), opp %d -> %d (want -1)", meBefore, me.Life, oppBefore, opp.Life)
		}
	})
	t.Run("a creature stolen from you does not", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportCutthroatOracle, false)
		mine := ctrlPushCreature(g, me.ID, "Mine")
		g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, mine, opp.ID, game.IndefiniteDuration(), "theft") })

		meBefore, oppBefore := me.Life, opp.Life
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
		passPriorityAroundTable(t, g)
		if me.Life != meBefore || opp.Life != oppBefore {
			t.Errorf("a creature I own but do not control died: me %d -> %d, opp %d -> %d, want no change", meBefore, me.Life, oppBefore, opp.Life)
		}
	})
}

// A creature exiled instead of dying never died (CR 700.4), so
// nothing triggers, the Cutthroat's own replaced death included.
func TestZulaportDoesNotSeeCreaturesExiledInsteadOfDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castRestInPeace(t, g)
	z := pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportCutthroatOracle, false)
	bear := ctrlPushCreature(g, me.ID, "Bear")

	before := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		before[s.ID] = s.Life
	}
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(bear)
		_ = g.DestroyPermanentForEffect(z)
	})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear) || !g.Exile.Contains(z) {
		t.Fatal("setup: Rest in Peace did not exile both creatures")
	}
	for _, s := range g.Seats {
		if s.Life != before[s.ID] {
			t.Errorf("seat %v life %d -> %d, want no change", s.ID, before[s.ID], s.Life)
		}
	}
}

// The gain is its own instruction: an opponent who cannot lose life
// does not stop the controller gaining exactly one, and the opponents
// who can still lose theirs.
func TestZulaportGainsOneEvenWhenAnOpponentCannotLoseLife(t *testing.T) {
	g := newCatalogGame(t)
	me, locked, free := g.Seats[0], g.Seats[1], g.Seats[2]
	pushPermanentForTest(g, locked.ID, "Platinum Emperion", platinumEmperionOracle, "Artifact Creature — Golem")
	z := pushCatalogPermanent(g, me.ID, "Zulaport Cutthroat", "Creature — Human Rogue Ally", zulaportCutthroatOracle, false)
	meBefore, lockedBefore, freeBefore := me.Life, locked.Life, free.Life

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(z) })
	passPriorityAroundTable(t, g)
	if locked.Life != lockedBefore {
		t.Errorf("locked opponent %d -> %d, want unchanged", lockedBefore, locked.Life)
	}
	if free.Life != freeBefore-1 {
		t.Errorf("other opponent %d -> %d, want -1", freeBefore, free.Life)
	}
	if me.Life != meBefore+1 {
		t.Errorf("controller %d -> %d, want exactly +1", meBefore, me.Life)
	}
}
