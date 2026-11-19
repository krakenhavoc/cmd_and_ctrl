package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fog_patch_test.go — Fog Patch (#5839): "Cast this spell only during the
// declare blockers step. Attacking creatures become blocked."

const (
	fogPatchOracle = "eb8ff6c1-f9b4-4dae-8f46-0f1ad4c05355"
	fogPatchNeheb  = "b0d3fd57-1420-4dd1-8410-9d659655287f" // Neheb, the Eternal: afflict 3
	fogPatchTroll  = "6f18f41e-df8d-4b59-87cc-ea167823fbda" // Pygmy Troll: becomes blocked by a creature
)

func TestFogPatchIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(fogPatchOracle)
	if !ok || spec.Name != "Fog Patch" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Fog Patch: ok %v, %+v", ok, spec)
	}
}

// fogPatchTryCast puts Fog Patch in the caster's hand and casts it where
// the turn stands.
func fogPatchTryCast(g *game.Game, caster *game.Player) error {
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Fog Patch", TypeLine: "Instant",
		OracleID: fogPatchOracle, Owner: caster.ID, Controller: caster.ID,
	})
	return g.CastSpell(caster.ID, id, game.CastSpellParams{})
}

func TestFogPatchIsCastableOnlyDuringTheDeclareBlockersStep(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	advanceToMain(t, g)
	var cant *game.CantCastError
	if err := fogPatchTryCast(g, me); !errors.As(err, &cant) {
		t.Fatalf("main phase: cast = %v, want the timing refusal", err)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := fogPatchTryCast(g, me); !errors.As(err, &cant) {
		t.Fatalf("declare attackers step: cast = %v, want the timing refusal", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	finishBlockDeclarations(t, g)
	// On the attacker's own turn too: the clause names only the step.
	if err := fogPatchTryCast(g, me); err != nil {
		t.Fatalf("declare blockers step: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	if err := fogPatchTryCast(g, me); !errors.As(err, &cant) {
		t.Fatalf("combat damage step: cast = %v, want the timing refusal", err)
	}
}

// Every attacking creature becomes blocked: an unblocked one "becomes
// blocked" (afflict triggers, CR 509.3c) and assigns no combat damage
// (CR 510.1c), a trampler assigns all of its damage to the player
// (CR 702.19d), "becomes blocked by a creature" does not trigger
// (CR 509.3d), and one already blocked by a creature is not blocked again.
func TestFogPatchMakesEveryAttackerBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	neheb := b12Push(g, me.ID, "Neheb, the Eternal", "Legendary Creature — Zombie Minotaur Warrior", fogPatchNeheb, 4, 6)
	troll := b12Push(g, me.ID, "Pygmy Troll", "Creature — Troll", fogPatchTroll, 1, 1)
	trampler := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Trampler", TypeLine: "Creature — Beast",
		Power: 3, Toughness: 3, Keywords: []string{"trample"}, Owner: me.ID, Controller: me.ID,
	})
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	wall := b12Creature(g, them.ID, "Wall", "Creature — Wall", 0, 4)

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{neheb, troll, trampler, bear} {
		if err := g.DeclareAttacker(id, them.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	finishBlockDeclarations(t, g)
	becameBlocked := func(id uuid.UUID) int {
		n := 0
		for _, ev := range g.Events {
			if ev.Kind == game.EventBecomesBlocked && ev.CardID == id {
				n++
			}
		}
		return n
	}
	if becameBlocked(bear) != 1 || becameBlocked(neheb) != 0 {
		t.Fatalf("setup: the bear became blocked %d times, Neheb %d", becameBlocked(bear), becameBlocked(neheb))
	}

	life := them.Life
	castInPlace(t, g, them.ID, "Fog Patch", fogPatchOracle)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g) // the afflict trigger

	for _, id := range []uuid.UUID{neheb, troll, trampler} {
		if becameBlocked(id) != 1 {
			t.Errorf("an unblocked attacker became blocked %d times, want 1 (CR 509.3c)", becameBlocked(id))
		}
	}
	if becameBlocked(bear) != 1 {
		t.Errorf("the bear a creature blocks became blocked %d times, want 1", becameBlocked(bear))
	}
	if them.Life != life-3 {
		t.Errorf("defender has %d life after the afflict trigger, want %d", them.Life, life-3)
	}
	if p, tough := effectivePTOf(t, g, troll); p != 1 || tough != 1 {
		t.Errorf("Pygmy Troll is %d/%d: no creature blocks it, so it gets nothing (CR 509.3d)", p, tough)
	}

	life = them.Life
	advanceTo(t, g, game.StepEndCombat)
	if them.Life != life-3 {
		t.Errorf("defender went from %d to %d life in combat, want only the trampler's 3 (CR 510.1c, 702.19d)", life, them.Life)
	}
	for _, id := range []uuid.UUID{neheb, troll, trampler, bear} {
		if !attackingNow(g, id) {
			t.Error("an attacker left combat")
		}
	}
	if c, ok := g.LookupCardForEffect(wall); !ok || c.DamageMarked != 2 {
		t.Errorf("the wall was not dealt the bear's 2: %+v", c)
	}
}
