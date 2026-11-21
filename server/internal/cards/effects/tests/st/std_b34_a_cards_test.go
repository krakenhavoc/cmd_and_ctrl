package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b34_a_cards_test.go — Steer Clear, Full Steam Ahead, Quick Draw and
// Bristlepack Sentry (Outlaws of Thunder Junction, Standard leftovers).

const (
	steerClearOracle        = "2c22af49-c6ef-4de6-bb84-6b3e99879148"
	fullSteamAheadOracle    = "65e64c50-4757-4e8e-8043-bc61a5f21750"
	quickDrawOracle         = "ae0d17f6-859c-4b41-8b66-b68abb837bbb"
	bristlepackSentryOracle = "05852d99-8f30-4318-913c-d0b1eaf8c222"
)

func TestStdB34SteerClearDealsTwoWithoutAMount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Ogre", 5, 5)
	effectstest.HelicarrierAttacker(g, victim, me.ID)
	effectstest.CastCatalogSpell(t, g, "Steer Clear", "Instant", steerClearOracle, effectstest.TwTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TwCard(g, victim).DamageMarked; got != 2 {
		t.Errorf("damage %d, want 2", got)
	}
}

func TestStdB34SteerClearDealsFourWithAMount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.B12Creature(g, me.ID, "Horse", "Creature — Mount", 1, 1)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Ogre", 5, 5)
	effectstest.HelicarrierAttacker(g, victim, me.ID)
	effectstest.CastCatalogSpell(t, g, "Steer Clear", "Instant", steerClearOracle, effectstest.TwTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TwCard(g, victim).DamageMarked; got != 4 {
		t.Errorf("damage %d, want 4", got)
	}
}

// The Mount is read as the spell is cast (CR 601.2i): one that leaves in
// response does not shrink the damage.
func TestStdB34SteerClearReadsTheMountAsItIsCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mount := effectstest.B12Creature(g, me.ID, "Horse", "Creature — Mount", 1, 1)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Ogre", 5, 5)
	effectstest.HelicarrierAttacker(g, victim, me.ID)
	effectstest.CastCatalogSpell(t, g, "Steer Clear", "Instant", steerClearOracle, effectstest.TwTarget(victim))
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, mount); err != nil {
		t.Fatalf("move: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TwCard(g, victim).DamageMarked; got != 4 {
		t.Errorf("damage %d, want 4 (the Mount was there as the spell was cast)", got)
	}
}

func TestStdB34SteerClearRefusesACreatureNotInCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	idle := effectstest.PushVanillaCreature(g, opp.ID, "Ogre", 5, 5)
	if err := effectstest.CastCatalogSpellErr(t, g, "Steer Clear", "Instant", steerClearOracle, effectstest.TwTarget(idle)); err == nil {
		t.Fatal("Steer Clear targeted a creature that is neither attacking nor blocking")
	}
}

func TestStdB34FullSteamAheadPumpsOnlyTheCreaturesPresent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Full Steam Ahead", "Sorcery", fullSteamAheadOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.B12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if got := effectstest.EffectivePower(t, g, a); got != 4 {
		t.Errorf("your creature has power %d, want 4", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, a, "trample") {
		t.Error("your creature should have trample")
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("their creature has power %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, late); got != 2 {
		t.Errorf("a creature that arrived afterwards has power %d, want 2 (CR 611.2c)", got)
	}
}

func TestStdB34FullSteamAheadRefusesASecondBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wurm := effectstest.B12Creature(g, me.ID, "Craw Wurm", "Creature — Wurm", 6, 4)
	x := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	y := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Full Steam Ahead", "Sorcery", fullSteamAheadOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BrAttack(t, g, wurm)
	refused := effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: x, Attacker: wurm}, {Blocker: y, Attacker: wurm},
	}), game.BlockReasonTooManyBlockers)
	if refused.N != 1 {
		t.Errorf("the refusal carries max %d, want 1", refused.N)
	}
	if err := g.DeclareBlocker(x, wurm); err != nil {
		t.Fatalf("one blocker is legal: %v", err)
	}
}

func TestStdB34QuickDrawPumpsOursAndStripsTheirs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	fs := effectstest.BrCreature(g, opp.ID, "First Striker", "Creature — Knight", 2, 2, "first strike")
	ds := effectstest.BrCreature(g, opp.ID, "Double Striker", "Creature — Knight", 2, 2, "double strike")
	other := effectstest.BrCreature(g, g.Seats[2].ID, "Third Party", "Creature — Knight", 2, 2, "first strike")
	effectstest.CastCatalogSpell(t, g, "Quick Draw", "Instant", quickDrawOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, mine); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "first strike") {
		t.Error("your creature should have first strike")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, fs, "first strike") {
		t.Error("the opponent's first striker should have lost it")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, ds, "double strike") {
		t.Error("the opponent's double striker should have lost it")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, other, "first strike") {
		t.Error("another opponent's creature is untouched")
	}
}

func TestStdB34QuickDrawRefusesAnOpponentsCreatureAsTheFirstTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Quick Draw", "Instant", quickDrawOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Fatal("Quick Draw pumped a creature you don't control")
	}
}

func TestStdB34BristlepackSentryAttacksOnlyWithAPowerFourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentry := effectstest.B12Push(g, me.ID, "Bristlepack Sentry", "Creature — Raccoon Soldier", bristlepackSentryOracle, 2, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(sentry, opp.ID); err == nil {
		t.Fatal("a defender attacked with no power-4 creature")
	}
	effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
	if err := g.DeclareAttacker(sentry, opp.ID); err != nil {
		t.Fatalf("Bristlepack Sentry could not attack with a power-4 creature: %v", err)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, sentry, "defender") {
		t.Error("the waiver removed defender")
	}
}

var _ = uuid.Nil
