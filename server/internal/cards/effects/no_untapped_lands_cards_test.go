package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// no_untapped_lands_cards_test.go — "if you control no untapped lands"
// (no_untapped_lands.go, #5794) and the seven Prophecy cards that read it.
// Every card test flips the condition by tapping and untapping a real
// land: a Forest tapped through its own mana ability (CR 605) and
// untapped through the sandbox verb, both of which emit the events the
// layer cache listens to. Each reading goes through ReadSnapshot, which
// recomputes only when the layer version moved, so a tap the cache
// missed would show up as a stale size.

const (
	nulScoriaCatOracle       = "88ef3413-21bc-428a-9803-d275e437c4bd"
	nulSpurGrapplerOracle    = "42784364-719a-49ca-b515-64146a8f82e8"
	nulFenStalkerOracle      = "cfd00c0e-d034-4b39-a3d3-f8b8842045dd"
	nulVintaraSnapperOracle  = "57c5cbd6-432d-4d48-9fc6-5b8a6eb223f1"
	nulKeldonBerserkerOracle = "7e0b9e67-3103-4299-934b-46d8fe96f553"
	nulWellOfLifeOracle      = "4e05d1a2-ad53-4f2e-aaea-7ec464b07c04"
	nulWellOfDiscoveryOracle = "d63185ac-c7cd-4501-9248-54d6caf7d091"
)

func TestNoUntappedLandsCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		nulScoriaCatOracle:       "Scoria Cat",
		nulSpurGrapplerOracle:    "Spur Grappler",
		nulFenStalkerOracle:      "Fen Stalker",
		nulVintaraSnapperOracle:  "Vintara Snapper",
		nulKeldonBerserkerOracle: "Keldon Berserker",
		nulWellOfLifeOracle:      "Well of Life",
		nulWellOfDiscoveryOracle: "Well of Discovery",
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

// --- fixtures ------------------------------------------------------------------

// nulPush seeds a catalog permanent with its printed size, not summoning
// sick, with an entry event so it has a CR 613 timestamp.
func nulPush(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// nulForest seeds an untapped basic Forest under owner's control.
func nulForest(g *game.Game, owner uuid.UUID) uuid.UUID {
	return nulPush(g, owner, "Forest", "Basic Land — Forest", "", 0, 0)
}

// nulTapForMana taps the land through its mana ability, the way a player
// pays with it, and empties the pool again.
func nulTapForMana(t *testing.T, g *game.Game, p *game.Player, land uuid.UUID) {
	t.Helper()
	if err := g.ActivateManaAbility(p.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the land for mana: %v", err)
	}
	g.WithWriteLock(func() { p.ManaPool = nil })
}

// nulUntap untaps the land through the sandbox verb (EventUntapCard).
func nulUntap(t *testing.T, g *game.Game, land uuid.UUID) {
	t.Helper()
	if err := g.TapCard(land, false); err != nil {
		t.Fatalf("untap the land: %v", err)
	}
}

// nulPT is the permanent's current power and toughness, read the way a
// broadcast reads it.
func nulPT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	var p, tough int
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				p, tough, found = c.CurrentPower(), c.CurrentToughness(), true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on the battlefield", id)
	}
	return p, tough
}

// nulHas reports whether the permanent currently has the keyword.
func nulHas(t *testing.T, g *game.Game, id uuid.UUID, keyword string) bool {
	t.Helper()
	var has, found bool
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if c := &g.Battlefield.Cards[i]; c.InstanceID == id {
				has, found = game.HasKeyword(c, keyword), true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on the battlefield", id)
	}
	return has
}

func nulHolds(g *game.Game, you uuid.UUID) bool {
	var ok bool
	g.WithWriteLock(func() { ok = NoUntappedLands(g, you) })
	return ok
}

// --- the condition helper ---------------------------------------------------

// No lands at all is no untapped lands; an untapped land of yours breaks
// it and tapping it restores it; an opponent's untapped land and an
// untapped nonland permanent of yours do not count; a land creature is a
// land.
func TestNoUntappedLandsReadsYourUntappedLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if !nulHolds(g, me.ID) {
		t.Fatal("a player with no lands controls an untapped land")
	}
	nulForest(g, opp.ID)
	b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	if !nulHolds(g, me.ID) {
		t.Error("an opponent's Forest or my untapped Bears count as my untapped land")
	}
	forest := nulForest(g, me.ID)
	if nulHolds(g, me.ID) {
		t.Error("holds with my untapped Forest")
	}
	nulTapForMana(t, g, me, forest)
	if !nulHolds(g, me.ID) {
		t.Error("fails with my only Forest tapped")
	}
	nulUntap(t, g, forest)
	if nulHolds(g, me.ID) {
		t.Error("holds after the Forest untapped")
	}
	nulTapForMana(t, g, me, forest)
	arbor := nulPush(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad", "", 1, 1)
	if nulHolds(g, me.ID) {
		t.Error("holds with an untapped land creature")
	}
	g.WithWriteLock(func() {
		if err := g.TapTargetForEffect(arbor); err != nil {
			t.Fatalf("tap the land creature: %v", err)
		}
	})
	if !nulHolds(g, me.ID) {
		t.Error("fails with every land tapped")
	}
}

// --- Scoria Cat and Spur Grappler ---------------------------------------------

func TestNoUntappedLandsSelfPumpsFollowTheLand(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		pumpP, pumpT           int
	}{
		{"Scoria Cat", "Creature — Cat", nulScoriaCatOracle, 3, 3, 3, 3},
		{"Spur Grappler", "Creature — Beast", nulSpurGrapplerOracle, 2, 1, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			forest := nulForest(g, me.ID)
			id := nulPush(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			check := func(when string, pumped bool) {
				t.Helper()
				wantP, wantT := tc.power, tc.toughness
				if pumped {
					wantP, wantT = tc.power+tc.pumpP, tc.toughness+tc.pumpT
				}
				if p, tough := nulPT(t, g, id); p != wantP || tough != wantT {
					t.Errorf("%s: %d/%d, want %d/%d", when, p, tough, wantP, wantT)
				}
			}
			check("with an untapped Forest", false)
			nulTapForMana(t, g, me, forest)
			check("with the Forest tapped", true)
			nulUntap(t, g, forest)
			check("after the Forest untapped", false)
			nulTapForMana(t, g, me, forest)
			check("with the Forest tapped again", true)
			nulForest(g, me.ID)
			check("after an untapped Forest entered", false)
		})
	}
}

// An animated land tapped by attacking (CR 508.1f) emits no EventTapCard;
// the static still sees it once the attack is declared.
func TestScoriaCatSeesALandTappedByAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arbor := nulPush(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad", "", 1, 1)
	cat := nulPush(g, me.ID, "Scoria Cat", "Creature — Cat", nulScoriaCatOracle, 3, 3)
	if p, tough := nulPT(t, g, cat); p != 3 || tough != 3 {
		t.Fatalf("with an untapped land creature: %d/%d, want 3/3", p, tough)
	}
	declareAttack(t, g, opp.ID, arbor)
	if c, _ := cardOnBattlefield(g, arbor); !c.Tapped {
		t.Fatal("the attacking land creature is not tapped")
	}
	if p, tough := nulPT(t, g, cat); p != 6 || tough != 6 {
		t.Errorf("after the land creature attacked: %d/%d, want 6/6", p, tough)
	}
}

// --- Fen Stalker and Vintara Snapper -----------------------------------------

func TestNoUntappedLandsSelfKeywordsFollowTheLand(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
		power, toughness                int
	}{
		{"Fen Stalker", "Creature — Nightstalker", nulFenStalkerOracle, "fear", 3, 2},
		{"Vintara Snapper", "Creature — Turtle", nulVintaraSnapperOracle, "shroud", 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			forest := nulForest(g, me.ID)
			id := nulPush(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			check := func(when string, want bool) {
				t.Helper()
				if got := nulHas(t, g, id, tc.keyword); got != want {
					t.Errorf("%s: has %s = %v, want %v", when, tc.keyword, got, want)
				}
			}
			check("with an untapped Forest", false)
			nulTapForMana(t, g, me, forest)
			check("with the Forest tapped", true)
			nulUntap(t, g, forest)
			check("after the Forest untapped", false)
		})
	}
}

// Fear is the engine's block rule: with the Forest tapped a green
// creature can't block the Stalker; with it untapped it can.
func TestFenStalkerFearRefusesANonblackBlocker(t *testing.T) {
	for _, tapped := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		forest := nulForest(g, me.ID)
		stalker := nulPush(g, me.ID, "Fen Stalker", "Creature — Nightstalker", nulFenStalkerOracle, 3, 2)
		bear := pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Colors: []string{"G"},
			Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
		})
		if tapped {
			nulTapForMana(t, g, me, forest)
		}
		declareAttack(t, g, opp.ID, stalker)
		advanceTo(t, g, game.StepDeclareBlockers)
		err := g.DeclareBlocker(bear, stalker)
		if tapped && err == nil {
			t.Error("Forest tapped: a green creature blocked the Stalker through fear")
		}
		if !tapped && err != nil {
			t.Errorf("Forest untapped: the block was refused: %v", err)
		}
	}
}

// --- Keldon Berserker ----------------------------------------------------------

func TestKeldonBerserkerChecksTheLandsAsItAttacksAndAsItResolves(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		forest := nulForest(g, me.ID)
		berserker := nulPush(g, me.ID, "Keldon Berserker", "Creature — Human Soldier Berserker", nulKeldonBerserkerOracle, 2, 3)
		return g, me, opp.ID, forest, berserker
	}

	t.Run("an untapped land: no trigger", func(t *testing.T) {
		g, _, opp, _, berserker := setup(t)
		declareAttack(t, g, opp, berserker)
		if triggerOnStack(g, berserker) != nil {
			t.Error("the trigger fired with an untapped Forest")
		}
		passPriorityAroundTable(t, g)
		if p, _ := nulPT(t, g, berserker); p != 2 {
			t.Errorf("power %d, want 2", p)
		}
	})

	t.Run("every land tapped: +3/+0 until end of turn", func(t *testing.T) {
		g, me, opp, forest, berserker := setup(t)
		nulTapForMana(t, g, me, forest)
		declareAttack(t, g, opp, berserker)
		if triggerOnStack(g, berserker) == nil {
			t.Fatal("no trigger with every land tapped")
		}
		passPriorityAroundTable(t, g)
		if p, tough := nulPT(t, g, berserker); p != 5 || tough != 3 {
			t.Errorf("after the trigger: %d/%d, want 5/3", p, tough)
		}
		// Untapping the land afterwards does not take the bonus away: the
		// condition was the trigger's, not a static's.
		nulUntap(t, g, forest)
		if p, _ := nulPT(t, g, berserker); p != 5 {
			t.Errorf("after the Forest untapped: power %d, want 5", p)
		}
		advanceToNextSeatsTurn(t, g)
		if p, _ := nulPT(t, g, berserker); p != 2 {
			t.Errorf("on the next turn: power %d, want 2", p)
		}
	})

	t.Run("a land untapped in response: nothing", func(t *testing.T) {
		g, me, opp, forest, berserker := setup(t)
		nulTapForMana(t, g, me, forest)
		declareAttack(t, g, opp, berserker)
		if triggerOnStack(g, berserker) == nil {
			t.Fatal("no trigger with every land tapped")
		}
		nulUntap(t, g, forest)
		passPriorityAroundTable(t, g)
		if p, _ := nulPT(t, g, berserker); p != 2 {
			t.Errorf("power %d after the condition failed on resolution, want 2 (CR 603.4)", p)
		}
	})
}

// --- Well of Life and Well of Discovery --------------------------------------

func TestWellsCheckTheLandsAtTheEndStepAndAsTheyResolve(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		// gained reads what the Well gives its controller.
		gained func(p *game.Player) int
		want   int
	}{
		{"Well of Life", nulWellOfLifeOracle, func(p *game.Player) int { return p.Life }, 2},
		{"Well of Discovery", nulWellOfDiscoveryOracle, func(p *game.Player) int { return p.Hand.Size() }, 1},
	} {
		setup := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			forest := nulForest(g, me.ID)
			well := nulPush(g, me.ID, tc.name, "Artifact", tc.oracle, 0, 0)
			return g, me, forest, well
		}

		t.Run(tc.name+": an untapped land", func(t *testing.T) {
			g, me, _, well := setup(t)
			before := tc.gained(me)
			advanceTo(t, g, game.StepEnd)
			if triggerOnStack(g, well) != nil {
				t.Error("the trigger fired with an untapped Forest")
			}
			passPriorityAroundTable(t, g)
			if got := tc.gained(me) - before; got != 0 {
				t.Errorf("gained %d, want 0", got)
			}
		})

		t.Run(tc.name+": every land tapped", func(t *testing.T) {
			g, me, forest, well := setup(t)
			nulTapForMana(t, g, me, forest)
			before := tc.gained(me)
			advanceTo(t, g, game.StepEnd)
			if triggerOnStack(g, well) == nil {
				t.Fatal("no trigger with every land tapped")
			}
			passPriorityAroundTable(t, g)
			if got := tc.gained(me) - before; got != tc.want {
				t.Errorf("gained %d, want %d", got, tc.want)
			}
		})

		t.Run(tc.name+": a land untapped in response", func(t *testing.T) {
			g, me, forest, well := setup(t)
			nulTapForMana(t, g, me, forest)
			before := tc.gained(me)
			advanceTo(t, g, game.StepEnd)
			if triggerOnStack(g, well) == nil {
				t.Fatal("no trigger with every land tapped")
			}
			nulUntap(t, g, forest)
			passPriorityAroundTable(t, g)
			if got := tc.gained(me) - before; got != 0 {
				t.Errorf("gained %d after the condition failed on resolution, want 0 (CR 603.4)", got)
			}
		})
	}

	// It is "your" end step: an opponent's does nothing.
	t.Run("an opponent's end step", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		well := nulPush(g, me.ID, "Well of Life", "Artifact", nulWellOfLifeOracle, 0, 0)
		advanceToNextSeatsTurn(t, g)
		advanceTo(t, g, game.StepEnd)
		if triggerOnStack(g, well) != nil {
			t.Error("the Well triggered in an opponent's end step")
		}
	})
}
