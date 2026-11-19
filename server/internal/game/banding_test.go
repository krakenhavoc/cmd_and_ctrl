package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// banding_test.go pins banding (CR 702.22, ADR 0224, #5723): forming a
// band as attackers are declared and the refusals of an illegal one, a
// blocker of one member blocking the band, the two damage divisions that
// change hands, granted banding, a member leaving combat, and a band
// surviving a snapshot. The cards (Benalish Hero, Mesa Pegasus, Timber
// Wolves, Helm of Chatzuk) are pinned in cards/effects.

// declareBand walks to declare_attackers and declares `ids` at seat 1 as
// one band (band number 1), through the bulk verb that announces bands.
func declareBand(t *testing.T, g *Game, ids ...uuid.UUID) error {
	t.Helper()
	advanceIntoStep(t, g, StepDeclareAttackers)
	decls := make([]AttackDeclaration, 0, len(ids))
	for _, id := range ids {
		decls = append(decls, AttackDeclaration{Attacker: id, Target: g.Seats[1].ID, Band: 1})
	}
	_, err := g.DeclareAttackersWith(decls, DeclareAttackersParams{})
	return err
}

// damageChoice is the one damage-assignment prompt queued, or nil.
func damageChoice(g *Game) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceDamageAssignment {
			return c
		}
	}
	return nil
}

func bandRefusal(t *testing.T, err error, want BandRefusal) {
	t.Helper()
	var be *BandRefusedError
	if !errors.As(err, &be) || be.Reason != want {
		t.Fatalf("err = %v, want a %s band refusal", err, want)
	}
	if !errors.Is(err, ErrIllegalBand) {
		t.Errorf("a band refusal does not wrap ErrIllegalBand")
	}
	if be.Sentence() == "" {
		t.Error("a band refusal with no sentence")
	}
}

func isAttacking(g *Game, id uuid.UUID) bool {
	c := findBattlefieldCard(g, id)
	return c != nil && c.AttackingTarget != uuid.Nil
}

// TestBandingFormsABandOfBandersAndOneOther — CR 702.22c: any number of
// creatures with banding and up to one without are a band, and the view
// numbers them together.
func TestBandingFormsABandOfBandersAndOneOther(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hero := pushCombatant(t, g, me, "Benalish Hero", 1, 1, KeywordBanding)
	wolves := pushCombatant(t, g, me, "Timber Wolves", 1, 1, KeywordBanding)
	bear := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	loner := pushCombatant(t, g, me, "Hill Giant", 3, 3)
	advanceIntoStep(t, g, StepDeclareAttackers)
	_, err := g.DeclareAttackersWith([]AttackDeclaration{
		{Attacker: hero, Target: g.Seats[1].ID, Band: 1},
		{Attacker: wolves, Target: g.Seats[1].ID, Band: 1},
		{Attacker: bear, Target: g.Seats[1].ID, Band: 1},
		{Attacker: loner, Target: g.Seats[1].ID},
	}, DeclareAttackersParams{})
	if err != nil {
		t.Fatalf("a legal band was refused: %v", err)
	}
	var mates []uuid.UUID
	var numbers map[uuid.UUID]int
	g.ReadSnapshot(func() {
		mates = g.bandMatesLocked(bear)
		numbers = g.AttackBandNumbersLocked()
	})
	if !sameIDs(mates, []uuid.UUID{hero, wolves}) && !sameIDs(mates, []uuid.UUID{wolves, hero}) {
		t.Errorf("the bear's band-mates = %v, want the Hero and the Wolves", mates)
	}
	if numbers[hero] == 0 || numbers[hero] != numbers[wolves] || numbers[hero] != numbers[bear] {
		t.Errorf("band numbers = %v, want the three together", numbers)
	}
	if _, in := numbers[loner]; in {
		t.Error("the attacker declared without a band is numbered")
	}
}

// TestBandingRefusesTwoWithoutBanding — CR 702.22c allows up to ONE
// creature without banding. Two, with or without a bander beside them,
// refuse the whole declaration, and nothing is declared or tapped.
func TestBandingRefusesTwoWithoutBanding(t *testing.T) {
	for _, withHero := range []bool{false, true} {
		g := newActiveGame(t)
		me := g.Seats[0]
		var ids []uuid.UUID
		if withHero {
			ids = append(ids, pushCombatant(t, g, me, "Benalish Hero", 1, 1, KeywordBanding))
		}
		a := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
		b := pushCombatant(t, g, me, "Hill Giant", 3, 3)
		ids = append(ids, a, b)
		bandRefusal(t, declareBand(t, g, ids...), BandRefusalTwoWithout)
		for _, id := range ids {
			if isAttacking(g, id) {
				t.Errorf("with hero %v: a refused band declared %s", withHero, findBattlefieldCard(g, id).Name)
			}
			if findBattlefieldCard(g, id).Tapped {
				t.Errorf("with hero %v: a refused band tapped a creature", withHero)
			}
		}
	}
}

// TestBandingRefusesABandSplitAcrossDefenders — CR 702.22d: a band
// attacks one player, planeswalker or battle.
func TestBandingRefusesABandSplitAcrossDefenders(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[0]
	hero := pushCombatant(t, g, me, "Benalish Hero", 1, 1, KeywordBanding)
	bear := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	_, err := g.DeclareAttackersWith([]AttackDeclaration{
		{Attacker: hero, Target: g.Seats[1].ID, Band: 1},
		{Attacker: bear, Target: g.Seats[2].ID, Band: 1},
	}, DeclareAttackersParams{})
	bandRefusal(t, err, BandRefusalSplitTarget)
	if isAttacking(g, hero) || isAttacking(g, bear) {
		t.Error("a refused band declared an attacker")
	}
}

// TestBandingBlockingOneMemberBlocksTheBand — CR 702.22h: a creature
// that blocks one member of a band blocks every member. Each member
// becomes blocked by it, every member deals it damage, and its own
// damage is divided among them by the ACTIVE player, since it blocks a
// creature with banding (CR 702.22k).
func TestBandingBlockingOneMemberBlocksTheBand(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hero := pushCombatant(t, g, me, "Benalish Hero", 1, 1, KeywordBanding)
	bear := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	wall := pushCombatant(t, g, opp, "Wall of Test", 3, 6)
	if err := declareBand(t, g, hero, bear); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := block(t, g, wall, bear); err != nil {
		t.Fatalf("block the bear: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if got := blockedSet(g, wall); !sameIDs(got, []uuid.UUID{bear, hero}) {
		t.Fatalf("the wall blocks %v, want the bear then the Hero", got)
	}
	var heroBlocked bool
	g.WithWriteLock(func() { heroBlocked = g.BlockedAttackerForEffect(hero) })
	if !heroBlocked {
		t.Error("the Hero is not blocked")
	}
	if n := len(blockDeclEvents(g, EventBecomesBlocked, hero)); n != 1 {
		t.Errorf("the Hero became blocked %d times, want 1", n)
	}
	// EventBlock for the band-mate's pair names the blocker, at a place
	// past the first, so "whenever this creature blocks" stays one.
	var pairs []Event
	for _, ev := range blockDeclEvents(g, EventBlock, wall) {
		pairs = append(pairs, ev)
	}
	if len(pairs) != 2 || pairs[0].Target != bear || pairs[0].Amount != 1 || pairs[1].Target != hero || pairs[1].Amount != 2 {
		t.Errorf("EventBlock pairs = %+v, want bear at 1 and the Hero at 2", pairs)
	}

	advanceIntoStep(t, g, StepCombatDamage)
	choice := damageChoice(g)
	if choice == nil {
		t.Fatal("no prompt for the wall's damage")
	}
	f := choice.DamageAssignment
	if choice.Chooser != me.ID || f.AttackerID != wall || !f.BlockerDivides || !f.Banding || f.AllowTrample {
		t.Fatalf("prompt = chooser %s, frame %+v; want the active player dividing the wall's damage", choice.Chooser, *f)
	}
	if err := g.ResolveDamageAssignment(choice.ID, opp.ID,
		[]DamageAssignmentEntry{{BlockerID: bear, Amount: 3}, {BlockerID: hero, Amount: 0}}, 0); !errors.Is(err, ErrNotTheChooser) {
		t.Fatalf("the blocker's controller answered the active player's division: %v", err)
	}
	// The active player puts all 3 on the Hero; the bear lives.
	if err := g.ResolveDamageAssignment(choice.ID, me.ID,
		[]DamageAssignmentEntry{{BlockerID: bear, Amount: 0}, {BlockerID: hero, Amount: 3}}, 0); err != nil {
		t.Fatalf("the active player's division: %v", err)
	}
	if findBattlefieldCard(g, hero) != nil {
		t.Error("the Hero took 3 and is still on the battlefield")
	}
	if got := damageOn(g, bear); got != 0 {
		t.Errorf("the bear has %d damage, want 0", got)
	}
	if got := damageOn(g, wall); got != 3 {
		t.Errorf("the wall has %d damage, want 3 from both members", got)
	}
	if opp.Life != StartingLife {
		t.Errorf("the defender lost %d life to a blocked band", StartingLife-opp.Life)
	}
}

// TestBandingDefendingPlayerDividesAttackerDamage — CR 702.22j: an
// attacker blocked by a creature with banding has its damage divided by
// the DEFENDING player among the creatures blocking it, and only among
// them: its trample puts nothing on the player.
func TestBandingDefendingPlayerDividesAttackerDamage(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := pushCombatant(t, g, me, "Trampling Beast", 4, 4, "trample")
	hero := pushCombatant(t, g, opp, "Benalish Hero", 1, 1, KeywordBanding)
	bear := pushCombatant(t, g, opp, "Grizzly Bears", 2, 2)
	declareAttacks(t, g, beast)
	if err := block(t, g, hero, beast, bear, beast); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	choice := damageChoice(g)
	if choice == nil {
		t.Fatal("no prompt for the beast's damage")
	}
	f := choice.DamageAssignment
	if choice.Chooser != opp.ID || f.AttackerID != beast || f.BlockerDivides || !f.Banding || f.AllowTrample || f.AsThoughUnblocked {
		t.Fatalf("prompt = chooser %s, frame %+v; want the defending player dividing, no trample", choice.Chooser, *f)
	}
	if err := g.ResolveDamageAssignment(choice.ID, opp.ID,
		[]DamageAssignmentEntry{{BlockerID: hero, Amount: 1}, {BlockerID: bear, Amount: 2}}, 1); err == nil {
		t.Fatal("the defending player's division put trample damage on the player")
	}
	// All 4 on the Hero: the bear lives, though the attacker's controller
	// would have killed both and trampled one over.
	if err := g.ResolveDamageAssignment(choice.ID, opp.ID,
		[]DamageAssignmentEntry{{BlockerID: hero, Amount: 4}, {BlockerID: bear, Amount: 0}}, 0); err != nil {
		t.Fatalf("the defending player's division: %v", err)
	}
	if findBattlefieldCard(g, hero) != nil {
		t.Error("the Hero took 4 and is still on the battlefield")
	}
	if got := damageOn(g, bear); got != 0 {
		t.Errorf("the bear has %d damage, want 0", got)
	}
	if opp.Life != StartingLife {
		t.Errorf("the defender lost %d life, want none: no trample through a banding block", StartingLife-opp.Life)
	}
}

// TestBandingLoneBlockerTakesATramplersDamage — CR 702.22j with one
// blocker: there is nothing to divide, so no prompt, and the whole
// power lands on the banding blocker rather than trampling over.
func TestBandingLoneBlockerTakesATramplersDamage(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := pushCombatant(t, g, me, "Trampling Beast", 5, 5, "trample")
	hero := pushCombatant(t, g, opp, "Benalish Hero", 1, 1, KeywordBanding)
	declareAttacks(t, g, beast)
	if err := block(t, g, hero, beast); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	if c := damageChoice(g); c != nil {
		t.Fatalf("a prompt with one blocker: %+v", *c)
	}
	if opp.Life != StartingLife {
		t.Errorf("the defender lost %d life to a trampler a banding creature blocked", StartingLife-opp.Life)
	}
	if findBattlefieldCard(g, hero) != nil {
		t.Error("the Hero is still on the battlefield")
	}
}

// TestBandingWithoutBandingIsTheOrdinaryDivision — the two divisions
// change hands only for banding: an attacker blocked by two creatures
// without it is divided by its own controller, trample and all.
func TestBandingWithoutBandingIsTheOrdinaryDivision(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := pushCombatant(t, g, me, "Trampling Beast", 4, 4, "trample")
	a := pushCombatant(t, g, opp, "Bear A", 1, 1)
	b := pushCombatant(t, g, opp, "Bear B", 2, 2)
	declareAttacks(t, g, beast)
	if err := block(t, g, a, beast, b, beast); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	choice := damageChoice(g)
	if choice == nil || choice.Chooser != me.ID || choice.DamageAssignment.Banding || !choice.DamageAssignment.AllowTrample {
		t.Fatalf("prompt = %+v, want the attacker's controller with trample", choice)
	}
}

// bandingGrantOracle is a test creature granted banding by a layer-6
// static, the road "gains banding" travels.
const bandingGrantOracle = "test-banding-granted"

// TestBandingGrantedBandingBands — a creature that has banding only
// through an effect bands, and blocking it hands the blocker's division
// to the active player, exactly as printed banding does.
func TestBandingGrantedBandingBands(t *testing.T) {
	self := func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID }
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != bandingGrantOracle {
			return nil
		}
		return []StaticAbility{{Layer: Layer6Ability, AppliesTo: self,
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				c.Abilities = append(c.Abilities, KeywordBanding)
			}}}
	})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	granted := NewCard("Granted Bander", me.ID)
	granted.TypeLine = "Creature — Test"
	granted.OracleID = bandingGrantOracle
	granted.Power, granted.Toughness = 2, 2
	// It entered this turn (pushTypedTestCard fires the zone move).
	granted.Keywords = []string{"haste"}
	gid := pushTypedTestCard(g, granted)
	bear := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	other := pushCombatant(t, g, me, "Hill Giant", 3, 3)
	wall := pushCombatant(t, g, opp, "Wall of Test", 3, 6)

	// Two without banding is refused; with the granted bander it is fine.
	advanceIntoStep(t, g, StepDeclareAttackers)
	_, err := g.DeclareAttackersWith([]AttackDeclaration{
		{Attacker: bear, Target: opp.ID, Band: 1},
		{Attacker: other, Target: opp.ID, Band: 1},
	}, DeclareAttackersParams{})
	bandRefusal(t, err, BandRefusalTwoWithout)
	if err := declareBand(t, g, gid, bear); err != nil {
		t.Fatalf("a band with granted banding: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := block(t, g, wall, bear); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	choice := damageChoice(g)
	if choice == nil || choice.Chooser != me.ID || !choice.DamageAssignment.Banding {
		t.Fatalf("prompt = %+v, want the active player dividing", choice)
	}
}

// TestBandingMemberRemovedFromCombatLeavesTheBand — CR 702.22f: a member
// removed from combat is out of the band, so blocking the rest does not
// block it, and the view stops numbering a band of one.
func TestBandingMemberRemovedFromCombatLeavesTheBand(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hero := pushCombatant(t, g, me, "Benalish Hero", 1, 1, KeywordBanding)
	bear := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	wolves := pushCombatant(t, g, me, "Timber Wolves", 1, 1, KeywordBanding)
	wall := pushCombatant(t, g, opp, "Wall of Test", 0, 6)
	if err := declareBand(t, g, hero, bear, wolves); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.RemoveFromCombat(wolves); err != nil {
		t.Fatal(err)
	}
	if err := block(t, g, wall, bear); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if got := blockedSet(g, wall); !sameIDs(got, []uuid.UUID{bear, hero}) {
		t.Fatalf("the wall blocks %v, want the bear and the Hero only", got)
	}
	// The Hero leaves too: the bear is a band of one, which the view does
	// not number.
	if err := g.RemoveFromCombat(hero); err != nil {
		t.Fatal(err)
	}
	var numbers map[uuid.UUID]int
	g.ReadSnapshot(func() { numbers = g.AttackBandNumbersLocked() })
	if len(numbers) != 0 {
		t.Errorf("band numbers = %v, want none for a band of one", numbers)
	}
}

// TestBandingLastsWhenBandingIsLost — CR 702.22e: the band lasts for the
// rest of combat even if a member loses banding, so blocking one still
// blocks both; but the division reads banding as damage is assigned
// (CR 702.22k), so with no bander left it is the blocker's controller's.
func TestBandingLastsWhenBandingIsLost(t *testing.T) {
	self := func(target *Card, _ *Game, source *Card) bool { return target.InstanceID == source.InstanceID }
	grant := true
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != bandingGrantOracle {
			return nil
		}
		return []StaticAbility{{Layer: Layer6Ability, AppliesTo: self,
			Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
				if grant {
					c.Abilities = append(c.Abilities, KeywordBanding)
				}
			}}}
	})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	granted := NewCard("Granted Bander", me.ID)
	granted.TypeLine = "Creature — Test"
	granted.OracleID = bandingGrantOracle
	granted.Power, granted.Toughness = 2, 2
	// It entered this turn (pushTypedTestCard fires the zone move).
	granted.Keywords = []string{"haste"}
	gid := pushTypedTestCard(g, granted)
	bear := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	wall := pushCombatant(t, g, opp, "Wall of Test", 3, 6)
	if err := declareBand(t, g, gid, bear); err != nil {
		t.Fatal(err)
	}
	grant = false
	g.WithWriteLock(func() { g.layerVersion.Add(1) })
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := block(t, g, wall, bear); err != nil {
		t.Fatal(err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if got := blockedSet(g, wall); !sameIDs(got, []uuid.UUID{bear, gid}) {
		t.Fatalf("the wall blocks %v, want both members of the band", got)
	}
	advanceIntoStep(t, g, StepCombatDamage)
	choice := damageChoice(g)
	if choice == nil || choice.Chooser != opp.ID || choice.DamageAssignment.Banding {
		t.Fatalf("prompt = %+v, want the wall's controller dividing (no banding left)", choice)
	}
}

// TestBandingSurvivesASnapshotMidCombat — a band declared, captured,
// restored: blocking one member on the restored game still blocks the
// band (CR 702.22h), and the division still changes hands.
func TestBandingSurvivesASnapshotMidCombat(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hero := pushCombatant(t, g, me, "Benalish Hero", 1, 1, KeywordBanding)
	bear := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	wall := pushCombatant(t, g, opp, "Wall of Test", 3, 6)
	if err := declareBand(t, g, hero, bear); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)

	snap, restored := roundTrip(t, g)
	if len(snap.AttackBands) != 2 {
		t.Fatalf("snapshot AttackBands = %v, want both members", snap.AttackBands)
	}
	if err := block(t, restored, wall, hero); err != nil {
		t.Fatal(err)
	}
	if err := restored.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	if got := blockedSet(restored, wall); !sameIDs(got, []uuid.UUID{hero, bear}) {
		t.Fatalf("after the restore the wall blocks %v, want the band", got)
	}
	advanceIntoStep(t, restored, StepCombatDamage)
	choice := damageChoice(restored)
	if choice == nil || choice.Chooser != me.ID || !choice.DamageAssignment.Banding {
		t.Fatalf("restored prompt = %+v, want the active player dividing", choice)
	}

	// And a prompt captured mid-answer keeps its chooser and its flag.
	_, again := roundTrip(t, restored)
	c2 := damageChoice(again)
	if c2 == nil || c2.Chooser != me.ID || !c2.DamageAssignment.Banding {
		t.Fatalf("prompt after a second restore = %+v", c2)
	}
	if err := again.ResolveDamageAssignment(c2.ID, me.ID,
		[]DamageAssignmentEntry{{BlockerID: hero, Amount: 3}, {BlockerID: bear, Amount: 0}}, 0); err != nil {
		t.Fatalf("answering the restored prompt: %v", err)
	}
	if findBattlefieldCard(again, hero) != nil {
		t.Error("the Hero is still on the battlefield")
	}
}

// TestBandingEffectBlockingOneBlocksTheBand — CR 702.22i: an effect that
// makes one member of a band blocked blocks the entire band.
func TestBandingEffectBlockingOneBlocksTheBand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	hero := pushCombatant(t, g, me, "Benalish Hero", 1, 1, KeywordBanding)
	bear := pushCombatant(t, g, me, "Grizzly Bears", 2, 2)
	if err := declareBand(t, g, hero, bear); err != nil {
		t.Fatal(err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	finishPendingBlockDeclarations(t, g)
	if n := becomeBlocked(g, bear); n != 2 {
		t.Fatalf("marked %d blocked, want the whole band of 2", n)
	}
	passUntilStep(t, g, StepCombatDamage)
	if got := StartingLife - g.Seats[1].Life; got != 0 {
		t.Errorf("defender lost %d life to a band an effect blocked", got)
	}
}
