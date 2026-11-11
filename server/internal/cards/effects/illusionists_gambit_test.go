package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// illusionists_gambit_test.go — Illusionist's Gambit (#5685) and the seam
// it waited on (ADR 0188 Decision 14): the requirement and the
// restriction follow the creatures the spell removed, in the combat it
// adds and in no other.

const illusionistsGambitOracle = "333745d9-e930-439b-94d6-3aeea2877f69"

func TestIllusionistsGambitIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(illusionistsGambitOracle)
	if !ok || spec.Name != "Illusionist's Gambit" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Illusionist's Gambit: ok %v, %+v", ok, spec)
	}
}

// gambitScene is the active player's three creatures, two of them
// attacking, and two opponents: the caster, who controls a planeswalker,
// and a third player.
type gambitScene struct {
	g                    *game.Game
	me, caster, third    *game.Player
	atCaster, atThird    uuid.UUID
	home, planeswalker   uuid.UUID
	firstCombat          int
}

func newGambitScene(t *testing.T) *gambitScene {
	t.Helper()
	g := newCatalogGame(t)
	s := &gambitScene{g: g, me: g.Seats[0], caster: g.Seats[1], third: g.Seats[2]}
	s.atCaster = b12Creature(g, s.me.ID, "Bear at the Caster", "Creature — Bear", 2, 2)
	s.atThird = b12Creature(g, s.me.ID, "Bear at the Third", "Creature — Bear", 2, 2)
	s.home = b12Creature(g, s.me.ID, "Bear at Home", "Creature — Bear", 2, 2)
	s.planeswalker = b12Permanent(g, s.caster.ID, "Their Planeswalker", "Legendary Planeswalker — Test")
	return s
}

// attack declares the two attackers and stops in the declare blockers
// step with every block declaration finished.
func (s *gambitScene) attack(t *testing.T) {
	t.Helper()
	advanceTo(t, s.g, game.StepDeclareAttackers)
	s.firstCombat = s.g.Turn.PhaseID
	if err := s.g.DeclareAttacker(s.atCaster, s.caster.ID); err != nil {
		t.Fatalf("DeclareAttacker at the caster: %v", err)
	}
	if err := s.g.DeclareAttacker(s.atThird, s.third.ID); err != nil {
		t.Fatalf("DeclareAttacker at the third player: %v", err)
	}
	advanceTo(t, s.g, game.StepDeclareBlockers)
	finishBlockDeclarations(t, s.g)
}

// castGambit attacks, has the caster cast the spell in the declare
// blockers step, and resolves it.
func (s *gambitScene) castGambit(t *testing.T) {
	t.Helper()
	s.attack(t)
	castInPlace(t, s.g, s.caster.ID, "Illusionist's Gambit", illusionistsGambitOracle)
	passPriorityAroundTable(t, s.g)
	if len(s.g.PhaseEffects) != 1 {
		t.Fatalf("the spell left %d waiting rules, want 1", len(s.g.PhaseEffects))
	}
}

// toNextDeclareAttackers leaves the combat in progress and stops in the
// next declare attackers step.
func (s *gambitScene) toNextDeclareAttackers(t *testing.T) {
	t.Helper()
	from := s.g.Turn.PhaseID
	advanceTo(t, s.g, game.StepEndCombat)
	advanceTo(t, s.g, game.StepDeclareAttackers)
	if s.g.Turn.PhaseID == from {
		t.Fatal("no further combat phase began")
	}
}

func attackingNow(g *game.Game, id uuid.UUID) bool {
	c, ok := g.LookupCardForEffect(id)
	return ok && c.AttackingTarget != uuid.Nil
}

func TestIllusionistsGambitRemovesUntapsAndTheAddedCombatBindsThoseCreatures(t *testing.T) {
	s := newGambitScene(t)
	s.attack(t)
	if !tappedNow(s.g, s.atCaster) || !attackingNow(s.g, s.atThird) {
		t.Fatal("setup: the attackers are not tapped and attacking")
	}
	life := s.caster.Life
	castInPlace(t, s.g, s.caster.ID, "Illusionist's Gambit", illusionistsGambitOracle)
	passPriorityAroundTable(t, s.g)
	for _, id := range []uuid.UUID{s.atCaster, s.atThird} {
		if attackingNow(s.g, id) || tappedNow(s.g, id) {
			t.Errorf("an attacker is still attacking (%v) or tapped (%v)", attackingNow(s.g, id), tappedNow(s.g, id))
		}
	}

	s.toNextDeclareAttackers(t)
	if s.caster.Life != life {
		t.Errorf("the caster went from %d to %d life in the combat the spell emptied", life, s.caster.Life)
	}
	// Each of those creatures attacks that combat if able.
	var re *game.AttackRequirementError
	if err := s.g.PassPriority(); !errors.As(err, &re) {
		t.Fatalf("pass with both creatures home = %v, want an attack requirement refusal", err)
	}
	// They can't attack the caster or the caster's planeswalkers.
	if err := s.g.DeclareAttacker(s.atCaster, s.caster.ID); err == nil {
		t.Error("a removed creature attacked the caster in the added combat")
	}
	if err := s.g.DeclareAttacker(s.atCaster, s.planeswalker); err == nil {
		t.Error("a removed creature attacked the caster's planeswalker in the added combat")
	}
	if err := s.g.DeclareAttacker(s.atThird, s.caster.ID); err == nil {
		t.Error("the creature that had attacked the third player attacked the caster in the added combat")
	}
	// Another opponent is a legal attack, and it obeys the requirement.
	if err := s.g.DeclareAttacker(s.atCaster, s.third.ID); err != nil {
		t.Fatalf("a removed creature could not attack the other opponent: %v", err)
	}
	if err := s.g.PassPriority(); !errors.As(err, &re) || re.Attacker != s.atThird {
		t.Fatalf("pass with one of the two attacking = %v, want a refusal naming the other", err)
	}
	if err := s.g.DeclareAttacker(s.atThird, s.third.ID); err != nil {
		t.Fatalf("the second removed creature could not attack the other opponent: %v", err)
	}
	// The creature that was not attacking is not one of "those creatures".
	if err := s.g.PassPriority(); err != nil {
		t.Fatalf("pass with both removed creatures attacking and the third home: %v", err)
	}
}

func TestIllusionistsGambitLeavesACreatureThatWasNotAttackingFree(t *testing.T) {
	s := newGambitScene(t)
	s.castGambit(t)
	s.toNextDeclareAttackers(t)
	if err := s.g.DeclareAttacker(s.home, s.caster.ID); err != nil {
		t.Errorf("a creature that was not attacking could not attack the caster: %v", err)
	}
}

// CR 500.8: a combat another effect adds afterwards runs first, and it is
// not the combat the spell named. The spell's own combat follows and is.
func TestIllusionistsGambitANewerAddedCombatRunsFirstAndIsNotBound(t *testing.T) {
	s := newGambitScene(t)
	s.castGambit(t)
	s.g.WithWriteLock(func() {
		s.g.AddPhasesForEffect(uuid.Nil, game.PhaseAnchor{Kind: game.AnchorThisPhase}, game.PhaseKindCombat)
	})
	s.toNextDeclareAttackers(t)
	if len(s.g.PhaseEffects) != 1 {
		t.Fatalf("in the newer combat %d rules wait, want the spell's one", len(s.g.PhaseEffects))
	}
	// One removed creature attacks the caster and the other stays home:
	// neither is restricted or required here.
	if err := s.g.DeclareAttacker(s.atCaster, s.caster.ID); err != nil {
		t.Fatalf("in the newer combat a removed creature could not attack the caster: %v", err)
	}
	if err := s.g.PassPriority(); err != nil {
		t.Fatalf("pass in the newer combat with a removed creature home: %v", err)
	}

	// The spell's own combat comes next and binds the creature that
	// stayed home, which is still untapped.
	s.toNextDeclareAttackers(t)
	if len(s.g.PhaseEffects) != 0 {
		t.Fatalf("in the spell's combat %d rules still wait", len(s.g.PhaseEffects))
	}
	if err := s.g.DeclareAttacker(s.atThird, s.caster.ID); err == nil {
		t.Error("in the spell's combat a removed creature attacked the caster")
	}
	var re *game.AttackRequirementError
	if err := s.g.PassPriority(); !errors.As(err, &re) || re.Attacker != s.atThird {
		t.Errorf("pass in the spell's combat with the untapped removed creature home = %v, want a refusal naming it", err)
	}
}

// blinkedAndStolenScene casts the spell over two hasty attackers, then
// exiles and returns one and has the caster steal the other until this
// combat ends, so its first controller has it back for the added combat.
func blinkedAndStolenScene(t *testing.T) *gambitScene {
	t.Helper()
	g := newCatalogGame(t)
	s := &gambitScene{g: g, me: g.Seats[0], caster: g.Seats[1], third: g.Seats[2]}
	haste := func(name string) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: s.me.ID, Controller: s.me.ID, Keywords: []string{"haste"},
		})
	}
	s.atCaster = haste("Blinked Bear")
	s.atThird = haste("Stolen Bear")
	s.castGambit(t)
	bf, exile := game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneExile}
	if err := g.MoveCardByID(bf, exile, s.atCaster); err != nil {
		t.Fatalf("exile: %v", err)
	}
	if err := g.MoveCardByID(exile, bf, s.atCaster); err != nil {
		t.Fatalf("return: %v", err)
	}
	g.WithWriteLock(func() {
		d, ok := g.UntilEndOfCombatDuration()
		if !ok || !g.GainControlForEffect(uuid.New(), s.atThird, s.caster.ID, d, "test — steal") {
			t.Fatal("setup: the steal was refused")
		}
	})
	g.ReadSnapshot(func() {})
	s.toNextDeclareAttackers(t)
	for _, id := range []uuid.UUID{s.atCaster, s.atThird} {
		if c, ok := g.LookupCardForEffect(id); !ok || c.Controller != s.me.ID || c.Tapped {
			t.Fatalf("setup: %s is not an untapped creature of the active player", c.Name)
		}
	}
	return s
}

// A creature that left the battlefield and came back is a new object
// (CR 400.7), and one whose controller changed is dropped: neither must
// attack in the added combat.
func TestIllusionistsGambitACreatureThatLeftOrChangedControllerNeedNotAttack(t *testing.T) {
	s := blinkedAndStolenScene(t)
	if err := s.g.PassPriority(); err != nil {
		t.Fatalf("pass with both creatures home: %v, want no requirement on either", err)
	}
}

// And either may attack the caster.
func TestIllusionistsGambitACreatureThatLeftOrChangedControllerMayAttackTheCaster(t *testing.T) {
	s := blinkedAndStolenScene(t)
	for _, id := range []uuid.UUID{s.atCaster, s.atThird} {
		if err := s.g.DeclareAttacker(id, s.caster.ID); err != nil {
			t.Errorf("a creature that left or changed controller could not attack the caster: %v", err)
		}
	}
}

// A restore point taken while the rule waits, and one taken in the added
// combat, both keep the requirement and the restriction.
func TestIllusionistsGambitSurvivesARestoreInTheMiddle(t *testing.T) {
	s := newGambitScene(t)
	s.castGambit(t)
	waiting, err := s.g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("restore while the rule waits: %v", err)
	}
	if len(waiting.PhaseEffects) != 1 || len(waiting.PhaseEffects[0].Effect.Affected) != 2 {
		t.Fatalf("the restored game's waiting rules: %+v", waiting.PhaseEffects)
	}
	advanceTo(t, waiting, game.StepEndCombat)
	advanceTo(t, waiting, game.StepDeclareAttackers)
	if err := waiting.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Errorf("restored while waiting: pass with both home = %v, want ErrAttackRequirement", err)
	}
	if err := waiting.DeclareAttacker(s.atCaster, s.caster.ID); err == nil {
		t.Error("restored while waiting: a removed creature attacked the caster")
	}

	s.toNextDeclareAttackers(t)
	inCombat, err := s.g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("restore in the added combat: %v", err)
	}
	if err := inCombat.PassPriority(); !errors.Is(err, game.ErrAttackRequirement) {
		t.Errorf("restored in the added combat: pass with both home = %v, want ErrAttackRequirement", err)
	}
	if err := inCombat.DeclareAttacker(s.atCaster, s.planeswalker); err == nil {
		t.Error("restored in the added combat: a removed creature attacked the caster's planeswalker")
	}
	if err := inCombat.DeclareAttacker(s.atCaster, s.third.ID); err != nil {
		t.Errorf("restored in the added combat: a removed creature could not attack the other opponent: %v", err)
	}
}

// "Cast this spell only during the declare blockers step on an
// opponent's turn."
func TestIllusionistsGambitCastTiming(t *testing.T) {
	strictCast := func(g *game.Game, p *game.Player) error {
		id := uuid.New()
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Illusionist's Gambit", TypeLine: "Instant",
			OracleID: illusionistsGambitOracle, Owner: p.ID, Controller: p.ID,
		})
		p.ManaPool.AddMana(game.ManaToken{Color: "U"}, game.ManaToken{Color: "U"}, game.ManaToken{Color: "U"}, game.ManaToken{Color: "U"})
		return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true})
	}
	// Its caster's own declare blockers step.
	s := newGambitScene(t)
	s.attack(t)
	if err := strictCast(s.g, s.me); err == nil {
		t.Error("the active player cast it in their own declare blockers step")
	}
	// An opponent's turn, outside the declare blockers step.
	s = newGambitScene(t)
	advanceTo(t, s.g, game.StepDeclareAttackers)
	if err := s.g.DeclareAttacker(s.atCaster, s.caster.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the caster: %v", err)
	}
	if s.g.Turn.Step != game.StepDeclareAttackers || s.g.Turn.PriorityHolder != 1 {
		t.Fatalf("setup: at %s with priority at seat %d, want the caster in declare attackers", s.g.Turn.Step, s.g.Turn.PriorityHolder)
	}
	if err := strictCast(s.g, s.caster); err == nil {
		t.Error("an opponent cast it in the declare attackers step")
	}
	// An opponent's declare blockers step.
	s = newGambitScene(t)
	s.attack(t)
	for s.g.Turn.PriorityHolder != 1 {
		if err := s.g.PassPriority(); err != nil {
			t.Fatalf("PassPriority toward the caster: %v", err)
		}
	}
	if err := strictCast(s.g, s.caster); err != nil {
		t.Errorf("an opponent could not cast it in the declare blockers step: %v", err)
	}
}
