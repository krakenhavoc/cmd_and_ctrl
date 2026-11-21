package testsms

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Every ninjutsu card declares the same hand ability with the same
// cost shape — the boot-time invariant, asserted here so a card file
// that hand-rolls the keyword instead of calling Ninjutsu() fails in
// this package rather than at server start.
func TestNinjutsuCardsDeclareAHandAbilityWithTheReturnCost(t *testing.T) {
	for _, tc := range []struct{ name, oracle, mana string }{
		{"Ninja of the Deep Hours", effectstest.NinjaOfTheDeepHoursOracle, "{1}{U}"},
		{"Ingenious Infiltrator", effectstest.IngeniousInfiltratorOracl, "{U}{B}"},
		{"Moonblade Shinobi", effectstest.MoonbladeShinobiOracle, "{2}{U}"},
		{"Prosperous Thief", effectstest.ProsperousThiefOracle, "{1}{U}"},
	} {
		abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle})
		if len(abilities) != 1 {
			t.Errorf("%s: %d activated abilities, want the one ninjutsu ability", tc.name, len(abilities))
			continue
		}
		ab := abilities[0]
		if ab.Cost.Mana != tc.mana {
			t.Errorf("%s: cost %q, want %q", tc.name, ab.Cost.Mana, tc.mana)
		}
		if ab.Cost.ReturnToHand.Empty() {
			t.Errorf("%s: no return-to-hand cost component", tc.name)
		} else if n := ab.Cost.ReturnToHand.Count; n != 1 {
			t.Errorf("%s: returns %d permanents, want 1", tc.name, n)
		}
		if !game.AbilityFunctionsFromZone(ab, game.ZoneHand) {
			t.Errorf("%s: the ability does not function from the hand (CR 702.49a)", tc.name)
		}
		if game.AbilityFunctionsFromZone(ab, game.ZoneBattlefield) {
			t.Errorf("%s: the ability is offered on the battlefield too", tc.name)
		}
		if ab.SorcerySpeed {
			t.Errorf("%s: sorcery-speed — ninjutsu is activated in the declare-blockers step", tc.name)
		}
		if why := game.AbilityNeedsPermanentSource(ab.Cost); why != "" {
			t.Errorf("%s: declares %s, which needs a permanent on the battlefield", tc.name, why)
		}
	}
}

// The headline. The unblocked attacker goes back to hand as the cost,
// and the ninja arrives from hand tapped and attacking the player the
// returned creature was attacking (CR 702.49a).
func TestNinjutsuReturnsTheAttackerAndTheNinjaArrivesAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ninja, attacker, me, opp := effectstest.NinjutsuBoard(t, g, "Ninja of the Deep Hours", effectstest.NinjaOfTheDeepHoursOracle, "{U}{U}")

	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	if !me.Hand.Contains(attacker) {
		t.Error("the unblocked attacker was not returned to hand — the cost is paid at announce")
	}
	if !me.Hand.Contains(ninja) {
		t.Error("the ninja left the hand at announce; it enters on RESOLUTION")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(ninja) {
		t.Fatal("the ninja is not on the battlefield")
	}
	c, ok := g.LookupCardForEffect(ninja)
	if !ok {
		t.Fatal("the ninja cannot be looked up")
	}
	if !c.Tapped {
		t.Error("the ninja entered untapped — ninjutsu puts it onto the battlefield TAPPED")
	}
	if c.AttackingTarget != opp.ID {
		t.Errorf("the ninja is attacking %s, want %s — the same player the returned creature was attacking (CR 702.49a)",
			c.AttackingTarget, opp.ID)
	}
}

// CR 506.3c: the ninja was never DECLARED as an attacker, so nothing
// that watches attack declarations may see it — no EventAttack, and
// the declaration's lock-in must not pick its AttackingTarget up as a
// staged declaration next time it runs.
func TestNinjutsuNinjaIsNeverDeclaredAsAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ninja, attacker, me, _ := effectstest.NinjutsuBoard(t, g, "Ninja of the Deep Hours", effectstest.NinjaOfTheDeepHoursOracle, "{U}{U}")

	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	// The resolution that put the ninja onto the battlefield ended in
	// a state-check pass, which is where commitAttackDeclarationLocked
	// runs — so the lock-in has already had its chance to announce it.
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(ninja) {
		t.Fatal("the ninja is not on the battlefield")
	}
	if n := effectstest.AttackEventsFor(g, ninja); n != 0 {
		t.Errorf("%d EventAttack for the ninja, want 0 (CR 506.3c)", n)
	}
	if n := effectstest.AttackEventsFor(g, attacker); n != 1 {
		t.Errorf("%d EventAttack for the creature that WAS declared, want 1", n)
	}
}

// The timing restriction, and it is the COST rather than a Condition:
// in the declare-attackers step no creature is an unblocked attacker
// yet (CR 509.1h), so nothing can pay.
func TestNinjutsuIsRefusedBeforeBlockersAreDeclared(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	ninja := effectstest.PushNinjaToHand(me, "Ninja of the Deep Hours", "Creature — Human Ninja", effectstest.NinjaOfTheDeepHoursOracle, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("activate in declare_attackers: %v, want ErrIllegalTarget — nothing is unblocked yet", err)
	}
	if !g.Battlefield.Contains(attacker) {
		t.Error("the refused activation returned the attacker anyway")
	}
	if !me.Hand.Contains(ninja) {
		t.Error("the refused activation moved the ninja")
	}
}

// And a BLOCKED attacker cannot pay either, which is the clause's
// whole point: ninjutsu rewards getting through, not swinging.
func TestNinjutsuRefusesABlockedAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	ninja := effectstest.PushNinjaToHand(me, "Ninja of the Deep Hours", "Creature — Human Ninja", effectstest.NinjaOfTheDeepHoursOracle, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("activate naming a blocked attacker: %v, want ErrIllegalTarget", err)
	}
}

// #1279: while the defending player is still DECLARING — they have a
// creature that could block and have not finished — the attacker is
// neither blocked nor unblocked, so nothing can pay. Before #1279 this
// was the window in which ninjutsu was available a beat early. Once the
// defender finishes with no block, it pays.
func TestNinjutsuWaitsForTheDefenderToFinishDeclaring(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	ninja := effectstest.PushNinjaToHand(me, "Ninja of the Deep Hours", "Creature — Human Ninja", effectstest.NinjaOfTheDeepHoursOracle, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if got := g.BlockDeclarationStatusOf(opp.ID); got != game.BlockDeclarationPending {
		t.Fatalf("setup: the defender should still be declaring, status %q", got)
	}

	err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("activate while the defender is declaring: %v, want ErrIllegalTarget", err)
	}
	if !g.Battlefield.Contains(attacker) {
		t.Fatal("the refused activation returned the attacker anyway")
	}

	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("FinishBlocks: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate once the defender declared none: %v", err)
	}
	if !me.Hand.Contains(attacker) {
		t.Error("the unblocked attacker was not returned as the cost")
	}
}

// #1501, CR 509.1 / 509.1h: the ninja arrives after its defending
// player has declared blockers, so it cannot be blocked — the
// declaration is over. The defender has an untapped Wall that could
// block it on any other board; the engine refuses the late block
// (blocks_declared), where before #1501 it took it as a sandbox
// allowance and "a ninja can't be blocked" was a table convention.
func TestANinjaThatArrivesAfterTheDeclarationCannotBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	ninja := effectstest.PushNinjaToHand(me, "Ninja of the Deep Hours", "Creature — Human Ninja", effectstest.NinjaOfTheDeepHoursOracle, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil { // declared, none
		t.Fatalf("FinishBlocks: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(ninja); !ok || c.AttackingTarget != opp.ID {
		t.Fatal("setup: the ninja is not on the battlefield attacking the defender")
	}

	var br *game.BlockRefusedError
	if err := g.DeclareBlocker(wall, ninja); !errors.As(err, &br) || br.Reason != game.BlockReasonBlocksDeclared {
		t.Fatalf("blocking the ninja after the declaration: %v, want blocks_declared", err)
	}
	if g.SeatOwesBlockDecision(opp.ID) {
		t.Error("the declared defender is told they owe a block decision on the ninja")
	}
	life := opp.Life
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if got := life - opp.Life; got != 2 {
		t.Errorf("defender lost %d life, want 2 — the ninja connects unblocked", got)
	}
}

// The ninja is unblocked when it arrives, so it connects in the combat
// damage step — and Ninja of the Deep Hours' own trigger draws.
func TestNinjaOfTheDeepHoursConnectsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ninja, attacker, me, opp := effectstest.NinjutsuBoard(t, g, "Ninja of the Deep Hours", effectstest.NinjaOfTheDeepHoursOracle, "{U}{U}")

	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	life, hand := opp.Life, me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	if got := life - opp.Life; got != 2 {
		t.Errorf("defender lost %d life, want 2 — the ninja arrived unblocked", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1 (the ninja connected and drew)", got)
	}
}

// Ingenious Infiltrator's "whenever a Ninja you control deals combat
// damage to a player" fires off its OWN arrival — it is a Ninja.
func TestIngeniousInfiltratorDrawsOffItsOwnNinjutsu(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	ninja := effectstest.PushNinjaToHand(me, "Ingenious Infiltrator", "Creature — Vedalken Ninja", effectstest.IngeniousInfiltratorOracl, 2, 3)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{B}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1 — a Ninja you control connected", got)
	}
}

// Moonblade Shinobi's payoff is a body: the Illusion arrives in the
// combat damage step, after blockers, so it is a flier for next turn.
func TestMoonbladeShinobiCreatesAnIllusionWhenItConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ninja, attacker, me, _ := effectstest.NinjutsuBoard(t, g, "Moonblade Shinobi", effectstest.MoonbladeShinobiOracle, "{U}{U}{U}")

	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{attacker},
	}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountTokensControlled(g, me.ID, "Illusion"); n != 1 {
		t.Errorf("%d Illusion tokens, want 1", n)
	}
}

// Prosperous Thief's trigger is a BATCH — "one or more Ninja or Rogue
// creatures you control deal combat damage to a player" is one trigger
// per player, however many of them got there.
func TestProsperousThiefMakesOneTreasurePerPlayerConnected(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rogue := effectstest.B12Creature(g, me.ID, "Sneaky Rogue", "Creature — Human Rogue", 1, 1)
	decoy := effectstest.B12Creature(g, me.ID, "Decoy", "Creature — Rat", 1, 1)
	ninja := effectstest.PushNinjaToHand(me, "Prosperous Thief", "Creature — Human Ninja", effectstest.ProsperousThiefOracle, 3, 2)
	effectstest.DeclareAttack(t, g, opp.ID, rogue, decoy)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}

	// The decoy goes home to pay, so the Rogue and the Thief both
	// connect with the same player — one Treasure, not two.
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{decoy},
	}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.CountTokensControlled(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("%d Treasure tokens, want 1 — the clause is one trigger per player connected with", n)
	}
}
