package testsu

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// viashino_bey_test.go — #5840: "If this creature attacks, all
// creatures you control attack if able." A CR 508.1d requirement that
// exists only in a declaration with the Bey in it: an attack set that
// includes the Bey must include every creature that can attack for
// free; a set without it, the empty one included, is free.

const viashinoBeyOracle = "d9c9b933-74f4-4898-bbb8-2da76b86ccda"

// beyTable seats the Bey and one plain creature on the active seat's
// side and returns the game, the active player, an opponent, the Bey
// and the creature.
func beyTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	bey := effectstest.B12Push(g, me.ID, "Viashino Bey", "Creature — Lizard", viashinoBeyOracle, 4, 3)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	return g, me, opp, bey, bear
}

func beyHasPass(g *game.Game, seat uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Kind == legal.KindPass {
			return true
		}
	}
	return false
}

// TestViashinoBeyNoAttackersIsLegal: with the Bey home the requirement
// does not exist, so declaring nothing obeys the maximum.
func TestViashinoBeyNoAttackersIsLegal(t *testing.T) {
	g, me, _, _, _ := beyTable(t)
	if !beyHasPass(g, me.ID) {
		t.Fatal("the enumerator withholds the pass with nothing declared")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with no attackers: %v", err)
	}
}

// TestViashinoBeyAloneIsRefusedNamingIt: the Bey attacking alone while
// another creature could attack is refused at the pass and at the
// skip-ahead, with a sentence that names the Bey; the enumerator
// withholds the pass and offers the owed attack as always legal; and
// once the other creature attacks too, the pass is accepted.
func TestViashinoBeyAloneIsRefusedNamingIt(t *testing.T) {
	g, me, opp, bey, bear := beyTable(t)
	if err := g.DeclareAttacker(bey, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(Bey): %v", err)
	}

	var re *game.AttackRequirementError
	err := g.PassPriority()
	if !errors.As(err, &re) || !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("pass with the Bey alone = %v, want an attack-requirement refusal", err)
	}
	if re.Attacker != bear || re.Requirement.Source != bey || !re.Requirement.IfSourceAttacks {
		t.Errorf("refusal = %+v, want the Bears owing the Bey's requirement", re)
	}
	const want = "Grizzly Bears must attack if able, because Viashino Bey is attacking."
	if got := re.Sentence(me.ID); got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	if _, err := g.AdvanceStep(); !errors.As(err, &re) {
		t.Fatalf("skip-ahead with the Bey alone = %v, want the same refusal", err)
	}

	if beyHasPass(g, me.ID) {
		t.Error("the enumerator offers the pass the engine refuses")
	}
	owed := 0
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind == legal.KindAttack && m.Source == bear && m.AlwaysLegal {
			owed++
		}
	}
	if owed == 0 {
		t.Error("no always-legal attack for the Bears, so a declining seat has nothing to fall back on")
	}
	effectstest.ClAgreeOnAttacks(t, g, me.ID)

	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(Bears): %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both attacking: %v", err)
	}
}

// TestViashinoBeyIsNotRequiredItself: the other creatures attacking
// without the Bey owe nothing, and nothing asks the Bey to join them.
func TestViashinoBeyIsNotRequiredItself(t *testing.T) {
	g, _, opp, _, bear := beyTable(t)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(Bears): %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Bears attacking and the Bey home: %v", err)
	}
}

// TestViashinoBeyAsksNothingOfACreatureThatCantAttack: a tapped
// creature and one with defender are not asked to attack (CR 508.1d,
// "without disobeying any restrictions"), so the Bey may attack alone.
func TestViashinoBeyAsksNothingOfACreatureThatCantAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	bey := effectstest.B12Push(g, me.ID, "Viashino Bey", "Creature — Lizard", viashinoBeyOracle, 4, 3)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall of Stone", TypeLine: "Creature — Wall",
		Power: 0, Toughness: 8, Owner: me.ID, Controller: me.ID, Keywords: []string{"defender"},
	})
	if err := g.DeclareAttacker(bey, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker(Bey): %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Bey alone and nothing else able to attack: %v", err)
	}
}

// TestViashinoBeyUnderSilentArbiter: with one attack slot in the
// combat, the Bey may take it (nothing else can attack beside it), but
// not from Zurgo, whose own requirement needs it.
func TestViashinoBeyUnderSilentArbiter(t *testing.T) {
	t.Run("the Bey takes the one slot", func(t *testing.T) {
		g, _, opp, bey, _ := beyTable(t)
		effectstest.B12Push(g, opp.ID, "Silent Arbiter", "Artifact Creature — Construct", effectstest.SilentArbiterOracle, 1, 5)
		if err := g.DeclareAttacker(bey, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker(Bey): %v", err)
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass with the Bey in the only slot: %v", err)
		}
	})
	t.Run("not from Zurgo", func(t *testing.T) {
		g, me, opp, bey, _ := beyTable(t)
		effectstest.B12Push(g, opp.ID, "Silent Arbiter", "Artifact Creature — Construct", effectstest.SilentArbiterOracle, 1, 5)
		zurgo := effectstest.B12Push(g, me.ID, "Zurgo Helmsmasher", "Legendary Creature — Orc Warrior", effectstest.ZurgoOracle, 7, 2)
		var re *game.AttackRequirementError
		if err := g.DeclareAttacker(bey, opp.ID); !errors.As(err, &re) || re.Attacker != zurgo {
			t.Fatalf("DeclareAttacker(Bey) into Zurgo's slot = %v, want a refusal naming Zurgo", err)
		}
		if err := g.DeclareAttacker(zurgo, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker(Zurgo): %v", err)
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass with Zurgo in the slot: %v", err)
		}
	})
}
