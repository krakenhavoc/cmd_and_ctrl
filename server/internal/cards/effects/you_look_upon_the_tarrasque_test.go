package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const youLookUponTheTarrasqueOracle = "231a6f8a-624a-4bff-8648-0b4ef910a6aa"

// castTarrasque casts the spell for `caster` (any seat holding priority)
// with the given mode and optional target.
func castTarrasque(t *testing.T, g *game.Game, caster *game.Player, mode int, targets []game.TargetRef) {
	t.Helper()
	id := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: id, Name: "You Look Upon the Tarrasque", TypeLine: "Instant",
		OracleID: youLookUponTheTarrasqueOracle, Owner: caster.ID, Controller: caster.ID})
	if err := g.CastSpell(caster.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: targets}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
}

// TestTarrasqueGatherYourCourageLuresOnlyOpponentsCreatures — the
// caster attacks with a creature it targets: the defender (an opponent)
// must block it with every creature able, and the target is +5/+5 and
// indestructible.
func TestTarrasqueGatherYourCourageLuresOnlyOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	a := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := reqCreature(g, opp.ID, "Wall B", "", 0, 4)
	advanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	for _, id := range []uuid.UUID{bear, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	castTarrasque(t, g, me, 1, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("target is %d/%d, want 7/7", p, tough)
	}
	if !hasEffectiveAbility(t, g, bear, "indestructible") {
		t.Error("target should be indestructible")
	}
	onToBlockers(t, g)
	reqRefusal(t, reqBlock(t, g, other, a))
	reqRefusal(t, g.PassPriority())
	if err := reqBlock(t, g, bear, a, b); err != nil {
		t.Fatalf("both walls block the target: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both blocking: %v", err)
	}
}

// TestTarrasqueSparesTheCastersOwnCreatures — the defender casts it on
// the opponent's attacker: the defender's creatures are the caster's,
// not "creatures your opponents control", so none is required to
// block. Stronger-than-printed (#259) is the direction that would
// refuse this pass.
func TestTarrasqueSparesTheCastersOwnCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	reqCreature(g, opp.ID, "Wall B", "", 0, 4)
	advanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	// The attacker passes; the defender holds priority and answers.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("attacker passes priority: %v", err)
	}
	castTarrasque(t, g, opp, 1, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 7 {
		t.Fatalf("target power = %d, want 7 (the spell resolved)", p)
	}
	onToBlockers(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("no blocks should be legal when the lure spares the caster's creatures: %v", err)
	}
}

// TestTarrasqueRunAndHidePreventsCombatDamageToYouAndYourCreatures -
// mode 0: the caster's creature takes no combat damage from its blocker,
// the blocker (an opponent's creature) still takes the attacker's, and
// non-combat damage to the caster's creature is not prevented.
func TestTarrasqueRunAndHidePreventsCombatDamageToYouAndYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := pushVanillaCreature(g, me.ID, "Attacker", 4, 4)
	blocker := pushVanillaCreature(g, opp.ID, "Blocker", 3, 5)
	pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	castTarrasque(t, g, me, 0, nil)
	passPriorityAroundTable(t, g)
	if n := pr7aShields(g); n != 1 {
		t.Fatalf("%d shields, want one", n)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if damageMarkedOn(g, attacker) != 0 || damageMarkedOn(g, blocker) != 4 {
		t.Fatalf("attacker %d, blocker %d damage: want 0 and 4", damageMarkedOn(g, attacker), damageMarkedOn(g, blocker))
	}
	pr6Damage(t, g, blocker, attacker, 2)
	if damageMarkedOn(g, attacker) != 2 {
		t.Errorf("non-combat damage to the attacker was prevented: %d marked", damageMarkedOn(g, attacker))
	}
}
