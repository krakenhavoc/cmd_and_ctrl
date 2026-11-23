package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	other := effectstest.ReqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	a := effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := effectstest.ReqCreature(g, opp.ID, "Wall B", "", 0, 4)
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	for _, id := range []uuid.UUID{bear, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	castTarrasque(t, g, me, 1, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 7 || tough != 7 {
		t.Errorf("target is %d/%d, want 7/7", p, tough)
	}
	if !effectstest.HasEffectiveAbility(t, g, bear, "indestructible") {
		t.Error("target should be indestructible")
	}
	effectstest.OnToBlockers(t, g)
	effectstest.ReqRefusal(t, effectstest.ReqBlock(t, g, other, a))
	effectstest.ReqRefusal(t, g.PassPriority())
	if err := effectstest.ReqBlock(t, g, bear, a, b); err != nil {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	effectstest.ReqCreature(g, opp.ID, "Wall B", "", 0, 4)
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	// The attacker passes; the defender holds priority and answers.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("attacker passes priority: %v", err)
	}
	castTarrasque(t, g, opp, 1, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 7 {
		t.Fatalf("target power = %d, want 7 (the spell resolved)", p)
	}
	effectstest.OnToBlockers(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("no blocks should be legal when the lure spares the caster's creatures: %v", err)
	}
}

// TestTarrasqueRunAndHidePreventsCombatDamageToYouAndYourCreatures -
// mode 0: the caster's creature takes no combat damage from its blocker,
// the blocker (an opponent's creature) still takes the attacker's, and
// non-combat damage to the caster's creature is not prevented.
func TestTarrasqueRunAndHidePreventsCombatDamageToYouAndYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 4, 4)
	blocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocker", 3, 5)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, blocker)
	castTarrasque(t, g, me, 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.Pr7aShields(g); n != 1 {
		t.Fatalf("%d shields, want one", n)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, attacker) != 0 || effectstest.DamageMarkedOn(g, blocker) != 4 {
		t.Fatalf("attacker %d, blocker %d damage: want 0 and 4", effectstest.DamageMarkedOn(g, attacker), effectstest.DamageMarkedOn(g, blocker))
	}
	effectstest.Pr6Damage(t, g, blocker, attacker, 2)
	if effectstest.DamageMarkedOn(g, attacker) != 2 {
		t.Errorf("non-combat damage to the attacker was prevented: %d marked", effectstest.DamageMarkedOn(g, attacker))
	}
}
