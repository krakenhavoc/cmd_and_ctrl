package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	berserkersFrenzyOracle    = "5cd51d39-aebc-4367-9157-da70b8d9b1a3"
	campaignOfVengeanceOracle = "627e9668-5e1c-4ab6-83f4-612a42d0afce"
)

// bfCast puts Berserker's Frenzy (instance id `id`) in seat's hand and
// casts it.
func bfCast(t *testing.T, g *game.Game, seat *game.Player, id uuid.UUID) error {
	t.Helper()
	seat.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Berserker's Frenzy", TypeLine: "Instant",
		ManaCost: "{2}{R}", OracleID: berserkersFrenzyOracle, Owner: seat.ID, Controller: seat.ID,
	})
	return g.CastSpell(seat.ID, id, game.CastSpellParams{})
}

// bfKeyWhere finds an RNG key under which Berserker's Frenzy (instance
// id `id`, cast by seat 0) keeps a roll that `want` accepts, by peeking
// at the roll on a throwaway table. The kept roll is the higher of the
// two draws.
func bfKeyWhere(t *testing.T, id uuid.UUID, want func(kept int) bool) byte {
	t.Helper()
	for n := 0; n < 256; n++ {
		g := effectstest.NewCatalogGame(t)
		g.SetRNGKeyForTest(effectstest.RandomModesKey(byte(n)))
		var kept []int
		g.WithWriteLock(func() {
			kept, _ = g.RollDiceIgnoringLowestForEffect(game.RandomDraw{Player: g.Seats[0].ID, Source: id}, 20, 2, 1)
		})
		if len(kept) == 1 && want(kept[0]) {
			return byte(n)
		}
	}
	t.Fatal("no key in 0..255 gives the wanted kept roll")
	return 0
}

// bfKeyFor is a key whose kept roll is 15 or more (high) or 14 or less.
func bfKeyFor(t *testing.T, id uuid.UUID, high bool) byte {
	t.Helper()
	return bfKeyWhere(t, id, func(kept int) bool { return (kept >= 15) == high })
}

// 15-20: the caster chooses the blocks for the turn (a blockers-only
// DeclarationControl), and only the kept roll was announced: one d20
// event, not two (CR 706.6).
func TestBerserkersFrenzyHighRollTakesTheBlocksAndAnnouncesOneRoll(t *testing.T) {
	id := uuid.New()
	key := bfKeyFor(t, id, true)
	g := effectstest.NewCatalogGame(t)
	g.SetRNGKeyForTest(effectstest.RandomModesKey(key))
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := bfCast(t, g, me, id); err != nil {
		t.Fatalf("cast at the beginning of combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	recs := effectstest.DeclarationRecords(g)
	if len(recs) != 1 || recs[0].Declarer != me.ID || recs[0].Attackers || !recs[0].Blockers || recs[0].ThisCombat {
		t.Fatalf("records = %+v, want this turn's blockers chosen by the caster", recs)
	}
	if effectstest.PendingOfKind(g, game.PendingChoiceOwnPermanents) != nil || effectstest.PendingOfKind(g, game.PendingChoiceTheirPermanents) != nil {
		t.Error("the 15-20 row asked for creatures")
	}
	rolls := effectstest.RollEventsOf(g)
	if len(rolls) != 1 || rolls[0].Sides != 20 || rolls[0].Amount < 15 {
		t.Fatalf("roll events = %+v, want one d20 of 15 or more (the ignored roll never happened)", rolls)
	}
}

// 15-20 end to end: the caster declares the opponent's blocks.
func TestBerserkersFrenzyHighRollTheCasterDeclaresTheBlocks(t *testing.T) {
	id := uuid.New()
	key := bfKeyFor(t, id, true)
	g := effectstest.NewCatalogGame(t)
	g.SetRNGKeyForTest(effectstest.RandomModesKey(key))
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := bfCast(t, g, me, id); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if got := g.ControllerOfPlayer(opp.ID); got != me.ID {
		t.Fatalf("the opponent's blocks are made by %v, want the caster", got)
	}
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: wall, Attacker: bear}}); err != nil {
		t.Fatalf("the caster blocking with the opponent's wall: %v", err)
	}
}

// 1-14: the caster chooses any number of creatures on any board and
// those, and only those, must block this turn if able.
func TestBerserkersFrenzyLowRollMakesTheChosenCreaturesBlock(t *testing.T) {
	id := uuid.New()
	key := bfKeyFor(t, id, false)
	g := effectstest.NewCatalogGame(t)
	g.SetRNGKeyForTest(effectstest.RandomModesKey(key))
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	chosen := effectstest.B12Creature(g, opp.ID, "Chosen Wall", "Creature — Wall", 0, 4)
	effectstest.B12Creature(g, opp.ID, "Spared Wall", "Creature — Wall", 0, 4)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := bfCast(t, g, me, id); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	legs := effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{opp.ID: {chosen}})
	if legs != 2 {
		t.Fatalf("%d creature legs, want one per board with a creature (2)", legs)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if recs := effectstest.DeclarationRecords(g); len(recs) != 0 {
		t.Fatalf("the 1-14 row took control of the blocks: %+v", recs)
	}
	if rolls := effectstest.RollEventsOf(g); len(rolls) != 1 || rolls[0].Amount > 14 {
		t.Fatalf("roll events = %+v, want one d20 of 14 or less", rolls)
	}

	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.FinishBlocks(opp.ID); err == nil {
		t.Fatal("no blocks was accepted though the chosen wall must block")
	}
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: chosen, Attacker: bear}}); err != nil {
		t.Fatalf("the chosen wall blocking alone (the other one was not chosen): %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("finishing with the chosen wall blocking: %v", err)
	}
}

// Choosing nobody is allowed ("any number"), and then nothing must block.
func TestBerserkersFrenzyLowRollMayChooseNoCreature(t *testing.T) {
	id := uuid.New()
	key := bfKeyFor(t, id, false)
	g := effectstest.NewCatalogGame(t)
	g.SetRNGKeyForTest(effectstest.RandomModesKey(key))
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if err := bfCast(t, g, me, id); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPermanentPicks(t, g, me.ID, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatalf("no one chosen, but blocking was required: %v", err)
	}
}

// The cast window: before combat, and during combat before blockers; not
// from the declare blockers step on, and not after combat. A turn whose
// combat is skipped keeps it to the end of the precombat main phase.
func TestBerserkersFrenzyCastWindow(t *testing.T) {
	cast := func(t *testing.T, g *game.Game) error {
		t.Helper()
		return bfCast(t, g, g.Seats[0], uuid.New())
	}
	g := effectstest.NewCatalogGame(t)
	for _, step := range []game.Step{game.StepPrecombatMain, game.StepBeginCombat, game.StepDeclareAttackers} {
		effectstest.AdvanceTo(t, g, step)
		if err := cast(t, g); err != nil {
			t.Fatalf("cast in %s: %v", step, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPermanentPicks(t, g, g.Seats[0].ID, nil)
		effectstest.PassPriorityAroundTable(t, g)
	}
	// With an attacker the declare blockers step exists and shuts it.
	g2 := effectstest.NewCatalogGame(t)
	me, opp := g2.Seats[0], g2.Seats[1]
	bear := effectstest.B12Creature(g2, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g2, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g2)
	for _, step := range []game.Step{game.StepDeclareBlockers, game.StepEndCombat, game.StepPostcombatMain, game.StepEnd} {
		effectstest.AdvanceTo(t, g2, step)
		if err := cast(t, g2); err == nil {
			t.Errorf("cast in %s", step)
		}
	}
	// Combat skipped: open in the precombat main phase, shut after it.
	g3 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g3, game.StepPrecombatMain)
	effectstest.BaSkipCombat(t, g3)
	if err := cast(t, g3); err != nil {
		t.Fatalf("cast before a skipped combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g3)
	effectstest.AnswerPermanentPicks(t, g3, g3.Seats[0].ID, nil)
	effectstest.PassPriorityAroundTable(t, g3)
	effectstest.AdvanceTo(t, g3, game.StepPostcombatMain)
	if err := cast(t, g3); err == nil {
		t.Error("cast after a skipped combat")
	}
}

// The row boundary: a kept 15 is the 15-20 row and a kept 14 is the 1-14
// row.
func TestBerserkersFrenzyRowBoundary(t *testing.T) {
	for _, tc := range []struct {
		kept     int
		takeOver bool
	}{{15, true}, {14, false}} {
		id := uuid.New()
		key := bfKeyWhere(t, id, func(kept int) bool { return kept == tc.kept })
		g := effectstest.NewCatalogGame(t)
		g.SetRNGKeyForTest(effectstest.RandomModesKey(key))
		me := g.Seats[0]
		effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
		if err := bfCast(t, g, me, id); err != nil {
			t.Fatalf("kept %d: cast: %v", tc.kept, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if rolls := effectstest.RollEventsOf(g); len(rolls) != 1 || rolls[0].Amount != tc.kept {
			t.Fatalf("kept %d: roll events = %+v", tc.kept, rolls)
		}
		if got := len(effectstest.DeclarationRecords(g)) == 1; got != tc.takeOver {
			t.Errorf("kept %d: the caster took the blocks = %v, want %v", tc.kept, got, tc.takeOver)
		}
	}
}

// Campaign of Vengeance: each attacking creature drains its defending
// player for 1 and gains its controller 1.
func TestCampaignOfVengeanceDrainsPerAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Campaign of Vengeance", "Enchantment", campaignOfVengeanceOracle, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
	meLife, oppLife := me.Life, opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != meLife+2 || opp.Life != oppLife-2 {
		t.Fatalf("life after two attackers: me %d (was %d), opponent %d (was %d), want a 2-point drain", me.Life, meLife, opp.Life, oppLife)
	}
}
