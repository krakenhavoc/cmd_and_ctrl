package testsav

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blight_x_test.go — #2174: "As an additional cost to cast this spell,
// blight X" (CR 701.68a, CR 107.3a), end to end through CastSpell.

const soulImmolationOracle = "338747e8-bed8-4e60-8b19-5c2b80799477"

// castSoulImmolation puts the card in the caster's hand and announces
// it with X and the named creature.
func castSoulImmolation(t *testing.T, g *game.Game, x int, blight []uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Soul Immolation", TypeLine: "Sorcery",
		OracleID: soulImmolationOracle, Owner: active.ID, Controller: active.ID})
	effectstest.AdvanceToMain(t, g)
	return id, g.CastSpell(active.ID, id, game.CastSpellParams{XValue: x, BlightIDs: blight})
}

func TestSoulImmolationBlightsXAndDamagesXToEachOpponentSide(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	big := effectstest.PushVanillaCreature(g, me.ID, "Giant", 5, 5)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Bystander", 1, 4)
	wall := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 10)
	small := effectstest.PushVanillaCreature(g, opp.ID, "Imp", 1, 2)
	oppLife, myLife := opp.Life, me.Life

	if _, err := castSoulImmolation(t, g, 2, []uuid.UUID{big}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// CR 601.2h: the counters are already on the creature, with the
	// spell still on the stack.
	if got := effectstest.TwCard(g, big).Counters[game.CounterMinusOne]; got != 2 {
		t.Errorf("-1/-1 counters at announce: %d, want 2", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Life; got != oppLife-2 {
		t.Errorf("opponent life: %d, want %d", got, oppLife-2)
	}
	if got := effectstest.TwCard(g, wall).DamageMarked; got != 2 {
		t.Errorf("wall damage: %d, want 2", got)
	}
	if effectstest.TwCard(g, small) != nil {
		t.Errorf("a 1/2 opposing creature survived 2 damage")
	}
	if me.Life != myLife || effectstest.TwCard(g, mine).DamageMarked != 0 {
		t.Errorf("the caster's side took damage")
	}
}

func TestSoulImmolationAtXZeroPaysAndDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	wall := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 10)
	oppLife := opp.Life
	// No creature named: X = 0 puts nothing anywhere.
	if _, err := castSoulImmolation(t, g, 0, nil); err != nil {
		t.Fatalf("CastSpell at X=0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TwCard(g, mine).Counters[game.CounterMinusOne] != 0 || opp.Life != oppLife || effectstest.TwCard(g, wall).DamageMarked != 0 {
		t.Errorf("X=0 changed the game")
	}
}

func TestSoulImmolationAtXZeroAcceptsAnOptionalCreatureName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if _, err := castSoulImmolation(t, g, 0, []uuid.UUID{mine}); err != nil {
		t.Fatalf("CastSpell at X=0 naming a creature: %v", err)
	}
	if effectstest.TwCard(g, mine).Counters[game.CounterMinusOne] != 0 {
		t.Errorf("blight 0 put counters on the creature")
	}
}

func TestSoulImmolationRefusesAnUnpayableX(t *testing.T) {
	for _, tc := range []struct {
		name  string
		x     int
		pick  func(big, small uuid.UUID) []uuid.UUID
		setup func(g *game.Game, me uuid.UUID)
	}{
		{"X above the greatest toughness", 6, func(b, _ uuid.UUID) []uuid.UUID { return []uuid.UUID{b} }, nil},
		{"X>0 naming no creature", 2, func(_, _ uuid.UUID) []uuid.UUID { return nil }, nil},
		{"X>0 naming two creatures", 2, func(b, s uuid.UUID) []uuid.UUID { return []uuid.UUID{b, s} }, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			big := effectstest.PushVanillaCreature(g, me.ID, "Giant", 5, 5)
			small := effectstest.PushVanillaCreature(g, me.ID, "Imp", 1, 1)
			_, err := castSoulImmolation(t, g, tc.x, tc.pick(big, small))
			if !errors.Is(err, game.ErrInvalidParam) {
				t.Fatalf("err = %v, want ErrInvalidParam", err)
			}
			if effectstest.TwCard(g, big).Counters[game.CounterMinusOne] != 0 {
				t.Errorf("a refused cast still blighted the creature")
			}
		})
	}
}

func TestSoulImmolationRefusesXWithNoCreatureOrAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 10)
	// X=1 with no creature of your own: the ceiling is 0.
	if _, err := castSoulImmolation(t, g, 1, []uuid.UUID{theirs}); err == nil {
		t.Fatalf("X=1 with no creature of your own was accepted")
	}
}

func TestSoulImmolationCeilingReadsEffectiveToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if got := g.BlightXCeilingForEffect(me.ID); got != 2 {
		t.Fatalf("ceiling = %d, want 2", got)
	}
	if err := g.AddCounterForEffect(bear, game.CounterMinusOne, 1); err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
	if got := g.BlightXCeilingForEffect(me.ID); got != 1 {
		t.Errorf("ceiling after a -1/-1 counter = %d, want 1", got)
	}
}

// The creature dies of its own blight before the spell resolves; the
// cost stays paid and the spell still deals X.
func TestSoulImmolationBlightCanKillTheCreatureBeforeResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	goblin := effectstest.PushVanillaCreature(g, me.ID, "Goblin", 1, 2)
	oppLife := opp.Life
	if _, err := castSoulImmolation(t, g, 2, []uuid.UUID{goblin}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if effectstest.TwCard(g, goblin) != nil {
		t.Errorf("a 1/2 blighted for 2 is still on the battlefield before the spell resolves")
	}
	if len(g.Stack.Cards) != 1 {
		t.Errorf("stack depth %d, want the spell alone", len(g.Stack.Cards))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("opponent life %d, want %d — X read back as the announced 2", opp.Life, oppLife-2)
	}
}

func TestRegisterRefusesMisplacedBlightX(t *testing.T) {
	for name, spec := range map[string]effects.Spec{
		"optional": {OracleID: "blightx-optional", Name: "BX Optional",
			OptionalCosts: []game.AdditionalCost{{BlightX: true, Key: "blight", Label: "x"}}},
		"with pay X life": {OracleID: "blightx-life", Name: "BX Life",
			AdditionalCost: &game.AdditionalCost{BlightX: true, PayLifeX: true, Label: "x"}},
		"with optional blight": {OracleID: "blightx-opt-blight", Name: "BX Pair",
			AdditionalCost: effects.BlightXCost(), OptionalCosts: []game.AdditionalCost{effects.OptionalBlight(1)}},
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("Register accepted a misplaced blight X")
				}
			}()
			effects.Register(spec)
		})
	}
}
