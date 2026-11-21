package testsex

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// forced_attack_cards_test.go is the "attacks this turn if able, destroy it
// at the beginning of the next end step if it didn't attack" family of
// #4062: Nettling Imp, Norritt, Arcum's Whistle and Siren's Call, all
// printing the before-attackers window.

const (
	faNettlingImp    = "c58dfcbf-49e6-4ef0-bd31-ebd81b0cfa41"
	faNorritt        = "4318c704-888f-4d85-b5f7-6d9abb14e60c"
	faArcumsWhistle  = "de691eed-d90b-48f9-bae4-f982767f2bfd"
	faSirensCall     = "269fc857-a052-4f0a-9759-467ccf42bebb"
	faImpAbilityText = "Nettling Imp"
)

// faTable is the active player (seat 0, "me": the one who attacks), and
// the next seat, who owns the forcing permanent.
func faTable(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	return g, g.Seats[0], g.Seats[1]
}

// faFinishTurnToEndStep walks to the end step and lets the delayed
// trigger resolve.
func faFinishTurnToEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepEnd)
	// Two or more delayed triggers at once ask their controller for an
	// order; any order will do.
	for i := 0; i < 8; i++ {
		c := effectstest.PendingOfKind(g, game.PendingChoiceTriggerOrder)
		if c == nil {
			break
		}
		if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// Nettling Imp, on the attacker's turn, in the beginning of combat step:
// the creature must attack, and attacking spares it.
func TestForcedAttackNettlingImpForcesTheAttackAndSparesAnAttacker(t *testing.T) {
	g, me, opp := faTable(t)
	imp := effectstest.PushCatalogPermanent(g, opp.ID, "Nettling Imp", "Creature — Imp", faNettlingImp, false)
	bear := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Bear", "", 2, 2))
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(opp.ID, imp, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.FaMustAttack(t, g) {
		t.Fatal("the creature was not required to attack")
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	faFinishTurnToEndStep(t, g)
	if !effects.OnBattlefield(g, bear) {
		t.Error("the creature attacked and was destroyed anyway")
	}
}

// A creature that could not attack (it is tapped) is destroyed at the
// beginning of the next end step.
func TestForcedAttackNettlingImpDestroysACreatureThatDidNotAttack(t *testing.T) {
	g, me, opp := faTable(t)
	imp := effectstest.PushCatalogPermanent(g, opp.ID, "Nettling Imp", "Creature — Imp", faNettlingImp, false)
	bear := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Bear", "", 2, 2))
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(opp.ID, imp, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RtRetap(g, bear)
	if effectstest.FaMustAttack(t, g) {
		t.Fatal("a tapped creature was required to attack")
	}
	faFinishTurnToEndStep(t, g)
	if effects.OnBattlefield(g, bear) {
		t.Error("the creature that did not attack survived the end step")
	}
}

// The target: a non-Wall creature the active player has controlled since the
// turn began.
func TestForcedAttackNettlingImpTargetClause(t *testing.T) {
	g, me, opp := faTable(t)
	imp := effectstest.PushCatalogPermanent(g, opp.ID, "Nettling Imp", "Creature — Imp", faNettlingImp, false)
	wall := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"}})
	fresh := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Fresh", "", 2, 2))
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, fresh).SummonedThisTurn = true })
	mine := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Imp's own", "", 2, 2))
	for name, id := range map[string]uuid.UUID{"a Wall": wall, "a creature that arrived this turn": fresh, "the Imp controller's own creature": mine} {
		if err := g.ActivateCatalogAbility(opp.ID, imp, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(id)}); err == nil {
			t.Errorf("targeted %s", name)
		}
	}
	if effectstest.B16Tapped(t, g, imp) {
		t.Error("a refused activation tapped the Imp")
	}
}

// Activate only during an opponent's turn, before attackers are declared.
func TestForcedAttackNettlingImpWindow(t *testing.T) {
	g, me, opp := faTable(t)
	imp := effectstest.PushCatalogPermanent(g, opp.ID, "Nettling Imp", "Creature — Imp", faNettlingImp, false)
	bear := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Bear", "", 2, 2))
	params := game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(bear)}

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	effectstest.AcRefused(t, g, opp.ID, imp, 0, params)

	// Its controller's own turn is not an opponent's turn.
	effectstest.AdvanceToMainOf(t, g, 1)
	other := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Own", "", 2, 2))
	effectstest.AcRefused(t, g, opp.ID, imp, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(other)})
}

// Norritt: the first ability untaps a blue creature at any time; the second is
// Nettling Imp's, on any turn, before attackers.
func TestForcedAttackNorrittUntapsABlueCreatureAndForcesAnAttack(t *testing.T) {
	g, me, _ := faTable(t)
	norritt := effectstest.PushCatalogPermanent(g, me.ID, "Norritt", "Creature — Imp", faNorritt, false)
	blue := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Blue", TypeLine: "Creature — Test", Colors: []string{"U"}, Power: 1, Toughness: 1})
	red := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Red", TypeLine: "Creature — Test", Colors: []string{"R"}, Power: 1, Toughness: 1})
	effectstest.RtRetap(g, blue)
	if err := g.ActivateCatalogAbility(me.ID, norritt, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(red)}); err == nil {
		t.Fatal("untapped a creature that is not blue")
	}
	effectstest.BaActivate(t, g, me, norritt, effectstest.LtCardTarget(blue))
	if effectstest.B16Tapped(t, g, blue) {
		t.Error("the blue creature stayed tapped")
	}

	// The second ability, on my own turn: the target is my own creature
	// because I am the active player.
	effectstest.RtUntap(g, norritt)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, norritt, 1, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(red)}); err != nil {
		t.Fatalf("second ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.FaMustAttack(t, g) {
		t.Error("the chosen creature was not required to attack")
	}
	// With attackers declared the second ability is shut.
	effectstest.RtUntap(g, norritt)
	effectstest.AcRefused(t, g, me.ID, norritt, 1, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(blue)})
}

// Arcum's Whistle asks the ACTIVE player to pay the creature's mana
// value; paying leaves the creature alone, refusing forces the attack.
func TestForcedAttackArcumsWhistleAsksTheActivePlayerToPayItsManaValue(t *testing.T) {
	for _, pay := range []bool{true, false} {
		name := "declined"
		if pay {
			name = "paid"
		}
		t.Run(name, func(t *testing.T) {
			g, me, opp := faTable(t)
			whistle := effectstest.PushCatalogPermanent(g, opp.ID, "Arcum's Whistle", "Artifact", faArcumsWhistle, false)
			ogre := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", ManaCost: "{1}{R}", Power: 3, Toughness: 3})
			effectstest.AdvanceTo(t, g, game.StepBeginCombat)
			effectstest.B06AddMana(opp, "C", "C", "C")
			if err := g.ActivateCatalogAbility(opp.ID, whistle, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(ogre)}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			ask := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
			if ask == nil {
				t.Fatal("nobody was asked to pay")
			}
			if ask.Chooser != me.ID {
				t.Fatalf("asked %s, want the active player %s", ask.Chooser, me.ID)
			}
			if pay {
				effectstest.ApaMana(me, "C", "C")
			}
			if err := g.ResolvePayUnless(ask.ID, me.ID, pay); err != nil {
				t.Fatalf("ResolvePayUnless: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.FaMustAttack(t, g); got == pay {
				t.Errorf("required to attack = %v after paying = %v", got, pay)
			}
			if !pay {
				// Never attacked; make it unable and see it destroyed.
				effectstest.RtRetap(g, ogre)
				faFinishTurnToEndStep(t, g)
				if effects.OnBattlefield(g, ogre) {
					t.Error("the creature survived without attacking")
				}
			}
		})
	}
}

// A creature with a mana value of zero costs {0}, which is always paid.
func TestForcedAttackArcumsWhistleAsksNothingOfAZeroManaValue(t *testing.T) {
	g, me, opp := faTable(t)
	whistle := effectstest.PushCatalogPermanent(g, opp.ID, "Arcum's Whistle", "Artifact", faArcumsWhistle, false)
	token := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Token", "", 1, 1))
	effectstest.B06AddMana(opp, "C", "C", "C")
	if err := g.ActivateCatalogAbility(opp.ID, whistle, 0, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(token)}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless) != nil {
		t.Error("asked to pay {0}")
	}
}

func faCastSirensCall(t *testing.T, g *game.Game, caster *game.Player) error {
	t.Helper()
	effectstest.ApaMana(caster, "U")
	id := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: id, Name: "Siren's Call", TypeLine: "Instant", ManaCost: "{U}",
		OracleID: faSirensCall, Owner: caster.ID, Controller: caster.ID})
	err := g.CastSpell(caster.ID, id, game.CastSpellParams{})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

// Siren's Call: cast on an opponent's turn before attackers; every
// non-Wall creature the active player had since the turn began must attack
// and is destroyed at the end step if it did not. Walls and creatures that
// arrived this turn are ignored.
func TestForcedAttackSirensCallForcesTheActivePlayersCreatures(t *testing.T) {
	g, me, opp := faTable(t)
	a := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Attacker", "", 2, 2))
	b := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Reluctant", "", 2, 2))
	wall := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"}})
	fresh := effectstest.ApaPush(g, me.ID, me.ID, effectstest.ApaCreature("Fresh", "", 2, 2))
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, fresh).SummonedThisTurn = true })
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := faCastSirensCall(t, g, opp); err != nil {
		t.Fatalf("cast: %v", err)
	}
	// "Reluctant" is tapped and cannot attack; "Attacker" can.
	effectstest.RtRetap(g, b)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Fatalf("passing without attacking: err = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatal(err)
	}
	faFinishTurnToEndStep(t, g)
	if effects.OnBattlefield(g, b) {
		t.Error("the creature that did not attack survived")
	}
	if !effects.OnBattlefield(g, a) || !effects.OnBattlefield(g, wall) || !effects.OnBattlefield(g, fresh) {
		t.Errorf("survivors: attacker %v, wall %v, fresh %v, want all", effects.OnBattlefield(g, a), effects.OnBattlefield(g, wall), effects.OnBattlefield(g, fresh))
	}
}

// Siren's Call cast windows: not on your own turn, not once attackers are
// declared.
func TestForcedAttackSirensCallCastWindow(t *testing.T) {
	g, me, opp := faTable(t)
	_ = me
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := faCastSirensCall(t, g, opp); err == nil {
		t.Error("cast in the declare attackers step")
	}

	g2, me2, _ := faTable(t)
	effectstest.AdvanceTo(t, g2, game.StepBeginCombat)
	if err := faCastSirensCall(t, g2, me2); err == nil {
		t.Error("cast on its caster's own turn")
	}
}

// CR 506.8e, 506.8g (#4180): on a turn whose combat phase is skipped the
// "before attackers are declared" window closes when the precombat main
// phase ends, for Siren's Call (a spell) and for Norritt's ability alike.
func TestForcedAttackSkippedCombatClosesTheWindowAtTheEndOfThePrecombatMain(t *testing.T) {
	g, me, opp := faTable(t)
	norritt := effectstest.PushCatalogPermanent(g, me.ID, "Norritt", "Creature — Imp", faNorritt, false)
	target := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Red", TypeLine: "Creature — Test", Colors: []string{"R"}, Power: 1, Toughness: 1})
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.BaSkipCombat(t, g)

	// Still the precombat main phase: Norritt's ability is open.
	if err := g.ActivateCatalogAbility(me.ID, norritt, 1, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(target)}); err != nil {
		t.Fatalf("Norritt in the precombat main phase: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RtUntap(g, norritt)

	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if g.DeclareAttackersHasBegunThisTurnForEffect() {
		t.Fatal("the combat phase was not skipped")
	}
	effectstest.AcRefused(t, g, me.ID, norritt, 1, game.ActivateAbilityParams{Targets: effectstest.LtCardTarget(target)})
	if err := faCastSirensCall(t, g, opp); err == nil {
		t.Error("Siren's Call was cast after the precombat main phase of a turn whose combat was skipped")
	}
}

// Siren's Call is still castable in the precombat main phase of a turn
// whose combat will be skipped.
func TestForcedAttackSirensCallIsCastableBeforeASkippedCombat(t *testing.T) {
	g, _, opp := faTable(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.BaSkipCombat(t, g)
	if err := faCastSirensCall(t, g, opp); err != nil {
		t.Fatalf("cast in the precombat main phase: %v", err)
	}
}
