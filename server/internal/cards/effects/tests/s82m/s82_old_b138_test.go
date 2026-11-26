package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b138_test.go — Theros Beyond Death, S82 set OLD batch 138 (#5514).
// Every test plays the real catalog entry through the engine.

const (
	oldB138TempleThief    = "493ecd11-e980-42da-a79e-cddd9a94cc7a"
	oldB138ArenaTrickster = "69c27223-d676-4916-899b-297dc4ed2d59"
)

func TestOldB138EveryCardIsRegisteredWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB138TempleThief: "Temple Thief", oldB138ArenaTrickster: "Arena Trickster",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB138TempleThiefCantBeBlockedByEnchantedOrEnchantmentCreatures(t *testing.T) {
	for _, tc := range []struct {
		name    string
		blocker func(g *game.Game, opp uuid.UUID) uuid.UUID
		legal   bool
	}{
		{"an ordinary creature", func(g *game.Game, opp uuid.UUID) uuid.UUID {
			return effectstest.B12Creature(g, opp, "Wall", "Creature — Wall", 0, 4)
		}, true},
		{"an enchantment creature", func(g *game.Game, opp uuid.UUID) uuid.UUID {
			return effectstest.B12Creature(g, opp, "Nyx Wall", "Enchantment Creature — Wall", 0, 4)
		}, false},
		{"an enchanted creature", func(g *game.Game, opp uuid.UUID) uuid.UUID {
			wall := effectstest.B12Creature(g, opp, "Wall", "Creature — Wall", 0, 4)
			aura := effectstest.PushTestPermanent(g, opp, game.Card{Name: "Pacifism-ish", TypeLine: "Enchantment — Aura"})
			effectstest.FindBattlefieldCardForTest(g, aura).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: wall}
			return wall
		}, false},
		{"a creature wearing an Equipment", func(g *game.Game, opp uuid.UUID) uuid.UUID {
			wall := effectstest.B12Creature(g, opp, "Wall", "Creature — Wall", 0, 4)
			eq := effectstest.SeedEquipment(g, opp, "Some Equipment", "")
			effectstest.FindBattlefieldCardForTest(g, eq).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: wall}
			return wall
		}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			thief := effectstest.B12Push(g, me.ID, "Temple Thief", "Creature — Human Rogue", oldB138TempleThief, 2, 2)
			blocker := tc.blocker(g, opp.ID)
			effectstest.DeclareAttack(t, g, opp.ID, thief)
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			err := g.DeclareBlocker(blocker, thief)
			if tc.legal != (err == nil) {
				t.Errorf("block by %s: err %v, want legal=%v", tc.name, err, tc.legal)
			}
		})
	}
}

func TestOldB138ArenaTricksterGrowsOnTheFirstSpellOfAnOpponentsTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	trickster := effectstest.B12Push(g, me.ID, "Arena Trickster", "Creature — Human Shaman", oldB138ArenaTrickster, 3, 3)
	counters := func() int { return effectstest.B12Counter(t, g, trickster, game.CounterPlusOne) }

	// Your own turn: nothing.
	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if counters() != 0 {
		t.Fatalf("a spell on your own turn put %d counters", counters())
	}
	// An opponent's turn: the first spell counts, the second does not.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.B22CastInstantAs(t, g, me, "Shock One")
	effectstest.PassPriorityAroundTable(t, g)
	if counters() != 1 {
		t.Fatalf("the first spell on an opponent's turn put %d counters, want 1", counters())
	}
	effectstest.B22CastInstantAs(t, g, me, "Shock Two")
	effectstest.PassPriorityAroundTable(t, g)
	if counters() != 1 {
		t.Errorf("the second spell on the same opponent's turn put the total to %d, want 1", counters())
	}
	if p, th := effectstest.PtOf(t, g, trickster); p != 4 || th != 4 {
		t.Errorf("Arena Trickster is %d/%d, want 4/4", p, th)
	}
}
