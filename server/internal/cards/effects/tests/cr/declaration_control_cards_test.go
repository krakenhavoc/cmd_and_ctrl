package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// declaration_control_cards_test.go — #4154: Master Warcraft and Odric,
// Master Tactician. The machinery is pinned in
// game/declaration_control_test.go and actions/declaration_control_test.go;
// these tests pin the cards' use of it.

const (
	masterWarcraftOracle       = "f9bca7af-29b6-46ba-b847-56715d09d1da"
	odricMasterTacticianOracle = "ade898df-14a5-460b-94f6-1f3f74d3ff95"
	brutalHordechiefOracle     = "e70cac5e-674e-4df7-94c7-77f8801e3941"
	meleeOracle                = "93fc1ff8-6ef0-42a2-b090-ba95b6f3c98e"
)

// castMasterWarcraft puts Master Warcraft in seat's hand and casts it.
func castMasterWarcraft(t *testing.T, g *game.Game, seat *game.Player) (uuid.UUID, error) {
	t.Helper()
	id := uuid.New()
	seat.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Master Warcraft", TypeLine: "Instant",
		ManaCost: "{2}{R/W}{R/W}", OracleID: masterWarcraftOracle, Owner: seat.ID, Controller: seat.ID,
	})
	return id, g.CastSpell(seat.ID, id, game.CastSpellParams{})
}

// Cast on an opponent's turn before combat, Master Warcraft has its
// caster choose that turn's attackers and every other defender's
// blocks, as the card says.
func TestMasterWarcraftChoosesTheAttackersAndTheBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active, caster, defender := g.Seats[0], g.Seats[1], g.Seats[2]
	bear := effectstest.B12Creature(g, active.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, active.ID, "Other Bear", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, defender.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	// Seat 0 passes; seat 1 casts it with priority.
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if _, err := castMasterWarcraft(t, g, caster); err != nil {
		t.Fatalf("casting Master Warcraft before combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	recs := effectstest.DeclarationRecords(g)
	if len(recs) != 1 || recs[0].Declarer != caster.ID || !recs[0].Attackers || !recs[0].Blockers || recs[0].ThisCombat || recs[0].SourceName != "Master Warcraft" {
		t.Fatalf("records = %+v, want the caster choosing attackers and blockers this turn", recs)
	}

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if got := g.ControllerOfPlayer(active.ID); got != caster.ID {
		t.Fatalf("the active player's attack declaration is made by %v, want the caster", got)
	}
	if err := g.DeclareAttacker(bear, defender.ID); err != nil {
		t.Fatalf("the caster choosing an attacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("the caster's finish: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("the active player accepting the target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if got := g.ControllerOfPlayer(defender.ID); got != caster.ID {
		t.Fatalf("the defender's block declaration is made by %v, want the caster", got)
	}
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("the caster's block: %v", err)
	}
	if err := g.FinishBlocks(defender.ID); err != nil {
		t.Fatal(err)
	}
	if c := effectstest.MustBattlefieldCard(t, g, wall); c.BlockingTarget != bear {
		t.Fatalf("the wall blocks %v, want the bear", c.BlockingTarget)
	}
	// The effect is this turn's only.
	effectstest.AdvanceToMainOf(t, g, 1)
	if n := len(effectstest.DeclarationRecords(g)); n != 0 {
		t.Fatalf("records on the next turn: %d", n)
	}
}

// "Cast this spell only before attackers are declared": once the
// turn's first declare attackers step has begun it can't be cast, and
// an additional combat does not reopen it.
func TestMasterWarcraftCastOnlyBeforeAttackersAreDeclared(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if _, err := castMasterWarcraft(t, g, me); err != nil {
		t.Fatalf("casting at the beginning of combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if _, err := castMasterWarcraft(t, g, me); err == nil {
		t.Fatal("cast in the declare attackers step")
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if _, err := castMasterWarcraft(t, g, me); err == nil {
		t.Fatal("cast after combat")
	}
}

// A turn that skipped its combat closed the window with the precombat
// main phase (CR 506.8e): Master Warcraft reads the same shared fact as
// every other "before attackers are declared" card (#4180).
func TestMasterWarcraftCantBeCastAfterASkippedCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.BaSkipCombat(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if g.DeclareAttackersHasBegunThisTurnForEffect() {
		t.Fatal("setup: the combat phase was not skipped")
	}
	if _, err := castMasterWarcraft(t, g, me); err == nil {
		t.Fatal("cast after a skipped combat")
	}
}

// Brutal Hordechief: each attacker drains its defending player, and the
// ability makes every opponent's creature block this turn if able (one
// arriving later too), with its controller choosing how they block.
func TestBrutalHordechiefDrainsAndChoosesHowTheyBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chief := effectstest.B12Push(g, me.ID, "Brutal Hordechief", "Creature — Orc Warrior", brutalHordechiefOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"},
		game.ManaToken{Color: "R"}, game.ManaToken{Color: "W"})
	if err := g.ActivateCatalogAbility(me.ID, chief, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Brutal Hordechief: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	recs := effectstest.DeclarationRecords(g)
	if len(recs) != 1 || recs[0].Declarer != me.ID || recs[0].Attackers || !recs[0].Blockers || recs[0].ThisCombat {
		t.Fatalf("records = %+v, want this turn's blockers chosen by the activator", recs)
	}
	late := effectstest.B12Creature(g, opp.ID, "Late Wall", "Creature — Wall", 0, 4)

	meLife, oppLife := me.Life, opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != meLife+1 || opp.Life != oppLife-1 {
		t.Fatalf("life after one attacker: me %d (was %d), opponent %d (was %d), want a 1-point drain", me.Life, meLife, opp.Life, oppLife)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if got := g.ControllerOfPlayer(opp.ID); got != me.ID {
		t.Fatalf("the opponent's blocks are made by %v, want the activator", got)
	}
	// Both walls must block: the requirement covers the one that came
	// after the activation.
	if err := g.FinishBlocks(opp.ID); err == nil {
		t.Fatal("no blocks was accepted though both walls must block")
	}
	if err := g.DeclareBlockers([]game.BlockDeclaration{{Blocker: wall, Attacker: bear}, {Blocker: late, Attacker: bear}}); err != nil {
		t.Fatalf("both walls blocking: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
}

// castMelee puts Melee in seat's hand and casts it.
func castMelee(t *testing.T, g *game.Game, seat *game.Player) (uuid.UUID, error) {
	t.Helper()
	id := uuid.New()
	seat.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Melee", TypeLine: "Instant",
		ManaCost: "{4}{R}", OracleID: meleeOracle, Owner: seat.ID, Controller: seat.ID,
	})
	return id, g.CastSpell(seat.ID, id, game.CastSpellParams{})
}

// Melee: cast during its caster's combat before blockers, it makes this
// combat's blocks theirs, and each attacker nothing blocks is untapped
// and removed from combat.
func TestMeleeChoosesTheBlocksAndPullsTheUnblockedOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elk := effectstest.B12Creature(g, me.ID, "Elk", "Creature — Elk", 3, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if _, err := castMelee(t, g, me); err == nil {
		t.Fatal("Melee was cast before combat")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{bear, elk} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := castMelee(t, g, me); err != nil {
		t.Fatalf("casting Melee in the declare attackers step: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	recs := effectstest.DeclarationRecords(g)
	if len(recs) != 1 || recs[0].Declarer != me.ID || !recs[0].Blockers || recs[0].Attackers || !recs[0].ThisCombat {
		t.Fatalf("records = %+v, want this combat's blockers chosen by the caster", recs)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if got := g.ControllerOfPlayer(opp.ID); got != me.ID {
		t.Fatalf("the defender's blocks are made by %v, want the caster", got)
	}
	if err := g.DeclareBlocker(wall, elk); err != nil {
		t.Fatalf("the caster's block: %v", err)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.MustBattlefieldCard(t, g, bear); c.Tapped || c.AttackingTarget != uuid.Nil {
		t.Fatalf("the unblocked bear is tapped=%v attacking=%v, want untapped and out of combat", c.Tapped, c.AttackingTarget)
	}
	if c := effectstest.MustBattlefieldCard(t, g, elk); c.AttackingTarget == uuid.Nil {
		t.Fatal("the blocked elk was removed from combat")
	}
	before := opp.Life
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if opp.Life != before {
		t.Fatalf("the defender lost %d life; the only unblocked attacker was removed", before-opp.Life)
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if n := len(effectstest.DeclarationRecords(g)); n != 0 {
		t.Fatalf("records after the combat: %d", n)
	}
	var left int
	g.ReadSnapshot(func() { left = len(g.DelayedTriggers) })
	if left != 0 {
		t.Fatalf("%d delayed triggers outlived the combat", left)
	}
}

// Odric's trigger needs Odric and three others; it makes the blocks
// this combat's, and only this combat's, its controller's.
func TestOdricMasterTacticianChoosesThisCombatsBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	odric := effectstest.B12Push(g, me.ID, "Odric, Master Tactician", "Legendary Creature — Human Soldier", odricMasterTacticianOracle, 3, 4)
	a := effectstest.B12Creature(g, me.ID, "Alpha", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bravo", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Charlie", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)

	// Odric and two others: no trigger.
	effectstest.DeclareAttack(t, g, opp.ID, odric, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.DeclarationRecords(g)); n != 0 {
		t.Fatalf("Odric and two others made %d records", n)
	}

	effectstest.NextTurnOfMine(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, odric, a, b, c)
	effectstest.PassPriorityAroundTable(t, g)
	recs := effectstest.DeclarationRecords(g)
	if len(recs) != 1 || recs[0].Declarer != me.ID || recs[0].Attackers || !recs[0].Blockers || !recs[0].ThisCombat {
		t.Fatalf("records = %+v, want this combat's blockers chosen by Odric's controller", recs)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if got := g.ControllerOfPlayer(opp.ID); got != me.ID {
		t.Fatalf("the defender's blocks are made by %v, want Odric's controller", got)
	}
	if err := g.FinishBlocks(opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if n := len(effectstest.DeclarationRecords(g)); n != 0 {
		t.Fatalf("records after the combat: %d", n)
	}
}
