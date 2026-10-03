package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// block_requirement_cards_1684_test.go — #1684's cards on the CR 509.1c
// machinery: the "target creature blocks IT" kind (Goblin Grappler's
// Provoke, Grappling Hook, Turntimber Basilisk), the filtered Lure
// (Marble Priest, Talruum Piper) and the cards one constructor away
// (Alluring Scent, Bloodscent, Elvish Bard, Taunting Elf, Canopy
// Stalker, Watchdog, Nacatl War-Pride). Each: its main behaviour
// through a real attack and the defender's pass, plus one refusal.

const (
	alluringScentOracle      = "ad218276-a44b-4a61-8e42-26a27929bbbb"
	bloodscentOracle         = "b2aabfaf-5996-4b41-987a-6bd941f6ae88"
	elvishBardOracle         = "094b778f-95a0-436f-a9d0-20a46a674486"
	tauntingElfOracle        = "aef07e62-ac7f-4a5b-9732-136052d88a06"
	canopyStalkerOracle      = "b1062763-1fcf-4605-9253-daa1687e8173"
	watchdogOracle           = "6a951d2d-fc5a-4f41-bd77-3e301b76cf47"
	marblePriestOracle       = "96bcabee-e84e-409f-8439-80f5308d73e9"
	talruumPiperOracle       = "d18e60ed-ab6e-4b73-88f6-caa3f8e9c78e"
	goblinGrapplerOracle     = "48c1c84c-f690-42ce-9c5a-dd09b1e4197c"
	grapplingHookOracle      = "24c193cc-3f83-4414-8cac-8b9f48d5bb6d"
	turntimberBasiliskOracle = "661a99ef-8d5b-4aac-98f6-52b17a8ecf52"
	nacatlWarPrideOracle     = "692fdca7-0eff-4d2d-b744-213b6f6e18fb"
)

// seats1684 is the active player and the next seat, the defender every
// test here attacks.
func seats1684(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// typedCreature is reqCreature with a type line of its own.
func typedCreature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, t int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: t, Owner: owner, Controller: owner, Keywords: keywords,
	})
}

// attackAndTrigger declares `attackers` against the next seat and locks
// the declaration in, so the attack triggers are harvested. Returns the
// defender.
func attackAndTrigger(t *testing.T, g *game.Game, attackers ...uuid.UUID) *game.Player {
	t.Helper()
	seat := g.Turn.ActiveSeat
	opp := g.Seats[(seat+1)%len(g.Seats)]
	advanceToDeclareAttackersOf(t, g, seat)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	return opp
}

// onToBlockers walks from a settled declare-attackers step into
// declare blockers and hands the defender (the next seat, as
// attackAndTrigger attacks it) priority, so the next PassPriority is
// theirs — defenderHoldsPriority, since #1501 parks priority while a
// defender declares.
func onToBlockers(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepDeclareBlockers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to declare blockers: %v", err)
		}
	}
	defenderHoldsPriority(t, g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID)
}

func blocksAttackerRecords(g *game.Game) []game.Mod {
	var out []game.Mod
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModAddBlockRequirement && m.Text == string(game.BlockRequirementBlocksAttacker) {
				out = append(out, m)
			}
		}
	}
	return out
}

// --- Provoke: Goblin Grappler ------------------------------------------

// TestGoblinGrapplerProvokes — the provoked creature is untapped and
// must block the Grappler and no other attacker.
func TestGoblinGrapplerProvokes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	grappler := reqCreature(g, me.ID, "Goblin Grappler", goblinGrapplerOracle, 1, 1, "haste")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	victim := reqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(victim) })

	attackAndTrigger(t, g, grappler, other)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, victim); c.Tapped {
		t.Fatal("Provoke did not untap the target")
	}
	recs := blocksAttackerRecords(g)
	if len(recs) != 1 || recs[0].Objects[0].ID != grappler {
		t.Fatalf("records = %+v, want one naming the Grappler", recs)
	}

	onToBlockers(t, g)
	br := reqRefusal(t, reqBlock(t, g, other, victim))
	if got := br.Sentence(opp.ID); got != "Their Bear must block Goblin Grappler if able." {
		t.Errorf("sentence = %q", got)
	}
	reqRefusal(t, g.PassPriority())
	if err := reqBlock(t, g, grappler, victim); err != nil {
		t.Fatalf("the provoked block: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the provoked block made: %v", err)
	}
}

// TestProvokeAgainstACreatureThatCantBlockItAsksNothing — a flying
// provoker and a ground creature: the requirement can't be obeyed, so
// nothing is owed and the creature may block the other attacker.
func TestProvokeAgainstACreatureThatCantBlockItAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	grappler := reqCreature(g, me.ID, "Goblin Grappler", goblinGrapplerOracle, 1, 1, "haste", "flying")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	victim := reqCreature(g, opp.ID, "Their Bear", "", 2, 2)

	attackAndTrigger(t, g, grappler, other)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	onToBlockers(t, g)
	if owed := mustBlockNow(g); owed != nil {
		t.Errorf("must_block = %v, want nothing owed", owed)
	}
	if err := reqBlock(t, g, other, victim); err != nil {
		t.Fatalf("blocking the other attacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
}

func mustBlockNow(g *game.Game) map[uuid.UUID]uuid.UUID {
	var out map[uuid.UUID]uuid.UUID
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		out = g.MustBlockForEffect()
	})
	return out
}

// --- Grappling Hook ------------------------------------------------------

// TestGrapplingHookMakesTargetBlockTheEquippedCreature — double strike,
// and the target must block the equipped creature, not the Hook's
// other attacker.
func TestGrapplingHookMakesTargetBlockTheEquippedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	advanceToMain(t, g)
	hook := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Grappling Hook", TypeLine: equipTypeLine, OracleID: grapplingHookOracle, Owner: me.ID, Controller: me.ID})
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	victim := reqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	equipTo(t, g, me.ID, hook, bear)
	if !hasEffectiveAbility(t, g, bear, "double strike") {
		t.Fatal("the equipped creature lacks double strike")
	}

	attackAndTrigger(t, g, bear, other)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	recs := blocksAttackerRecords(g)
	if len(recs) != 1 || recs[0].Objects[0].ID != bear {
		t.Fatalf("records = %+v, want one naming the equipped Bear", recs)
	}
	onToBlockers(t, g)
	br := reqRefusal(t, reqBlock(t, g, other, victim))
	if got := br.Sentence(opp.ID); got != "Their Bear must block Bear if able (Grappling Hook)." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, bear, victim); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
}

func hasEffectiveAbility(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	for _, a := range effectiveAbilities(t, g, id) {
		if a == kw {
			return true
		}
	}
	return false
}

// --- Turntimber Basilisk -------------------------------------------------

// TestTurntimberBasiliskLandfallMakesTargetBlockIt — a land enters, the
// target is made to block the Basilisk, and later that combat it must.
func TestTurntimberBasiliskLandfallMakesTargetBlockIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	basilisk := reqCreature(g, me.ID, "Turntimber Basilisk", turntimberBasiliskOracle, 2, 1, "haste")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	victim := reqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	playLandFromHand(t, g, "Forest", "")
	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	recs := blocksAttackerRecords(g)
	if len(recs) != 1 || recs[0].Objects[0].ID != basilisk {
		t.Fatalf("records = %+v, want one naming the Basilisk", recs)
	}
	if !hasEffectiveAbility(t, g, basilisk, "deathtouch") {
		t.Error("the Basilisk lacks deathtouch")
	}

	attackAndTrigger(t, g, basilisk, other)
	onToBlockers(t, g)
	reqRefusal(t, reqBlock(t, g, other, victim))
	reqRefusal(t, g.PassPriority())
	if err := reqBlock(t, g, basilisk, victim); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
}

// --- Filtered Lures: Marble Priest, Talruum Piper -----------------------

// TestMarblePriestLuresOnlyWallsAndIgnoresTheirDamage — the Wall must
// block the Priest, the Bear is free, and the Wall's combat damage to
// the Priest is prevented.
func TestMarblePriestLuresOnlyWallsAndIgnoresTheirDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	priest := typedCreature(g, me.ID, "Marble Priest", "Artifact Creature — Cleric", marblePriestOracle, 3, 3, "haste")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	wall := typedCreature(g, opp.ID, "Wall of Spears", "Artifact Creature — Wall", "", 2, 4)
	bear := typedCreature(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	reqToBlockers(t, g, priest, other)
	if owed := mustBlockNow(g); len(owed) != 1 || owed[wall] != priest {
		t.Fatalf("must_block = %v, want only the Wall on the Priest", owed)
	}
	br := reqRefusal(t, reqBlock(t, g, other, wall))
	if got := br.Sentence(opp.ID); got != "Wall of Spears must block Marble Priest if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, other, bear); err != nil {
		t.Fatalf("the non-Wall blocking elsewhere: %v", err)
	}
	if err := reqBlock(t, g, priest, wall); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
	advanceToStepInTurn(t, g, game.StepCombatDamage)
	if c, ok := battlefieldCard(g, priest); !ok || c.DamageMarked != 0 {
		t.Errorf("Marble Priest took %d damage from a Wall, want it prevented", c.DamageMarked)
	}
	if c, ok := battlefieldCard(g, wall); !ok || c.DamageMarked != 3 {
		t.Errorf("the Wall took %d, want 3 — only damage TO the Priest is prevented", c.DamageMarked)
	}
}

// TestTalruumPiperLuresOnlyFlyers — the flyer must block the Piper, the
// ground creature is free.
func TestTalruumPiperLuresOnlyFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	piper := reqCreature(g, me.ID, "Talruum Piper", talruumPiperOracle, 3, 3, "haste")
	bird := reqCreature(g, opp.ID, "Their Bird", "", 1, 1, "flying")
	reqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	reqToBlockers(t, g, piper)
	if owed := mustBlockNow(g); len(owed) != 1 || owed[bird] != piper {
		t.Fatalf("must_block = %v, want only the flyer", owed)
	}
	br := reqRefusal(t, g.PassPriority())
	if br.Blocker != bird {
		t.Errorf("the refusal names %s, want the flyer", br.Blocker)
	}
	if err := reqBlock(t, g, piper, bird); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Bear home: %v", err)
	}
}

// --- Lures: Alluring Scent, Bloodscent, Elvish Bard, Taunting Elf ------

// lureShapeTest: both potential blockers must block the Lure'd attacker;
// one sent at the other attacker is refused.
func lureShapeTest(t *testing.T, g *game.Game, lured, other, a, b uuid.UUID) {
	t.Helper()
	reqToBlockers(t, g, lured, other)
	reqRefusal(t, reqBlock(t, g, other, a))
	reqRefusal(t, g.PassPriority())
	if err := reqBlock(t, g, lured, a, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both blocking: %v", err)
	}
}

func TestAlluringScentLuresTheTargetForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	a := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := reqCreature(g, opp.ID, "Wall B", "", 0, 4)
	castCatalogSpell(t, g, "Alluring Scent", "Sorcery", alluringScentOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	lureShapeTest(t, g, bear, other, a, b)
}

// TestBloodscentLuresAfterAttackersAreDeclared — cast at instant speed
// in the declare-attackers step.
func TestBloodscentLuresAfterAttackersAreDeclared(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	a := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := reqCreature(g, opp.ID, "Wall B", "", 0, 4)
	// Attack, then cast Bloodscent in the declare-attackers step, after
	// the attackers are declared and before any blocker is.
	advanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	for _, id := range []uuid.UUID{bear, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Bloodscent", TypeLine: "Instant",
		OracleID: bloodscentOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: cardTarget(bear)}); err != nil {
		t.Fatalf("cast Bloodscent in the declare-attackers step: %v", err)
	}
	passPriorityAroundTable(t, g)
	onToBlockers(t, g)
	reqRefusal(t, reqBlock(t, g, other, a))
	if err := reqBlock(t, g, bear, a, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
}

func TestElvishBardLuresEveryAbleBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	bard := reqCreature(g, me.ID, "Elvish Bard", elvishBardOracle, 2, 4, "haste")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	lureShapeTest(t, g, bard, other, reqCreature(g, opp.ID, "Wall A", "", 0, 4), reqCreature(g, opp.ID, "Wall B", "", 0, 4))
}

func TestTauntingElfLuresEveryAbleBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	elf := reqCreature(g, me.ID, "Taunting Elf", tauntingElfOracle, 0, 1, "haste")
	other := reqCreature(g, me.ID, "Other Bear", "", 2, 2, "haste")
	lureShapeTest(t, g, elf, other, reqCreature(g, opp.ID, "Wall A", "", 0, 4), reqCreature(g, opp.ID, "Wall B", "", 0, 4))
}

// --- Canopy Stalker --------------------------------------------------------

// TestCanopyStalkerMustBeBlockedAndGainsLifeWhenItDies — nothing home
// is refused; when it dies its controller gains 1 life per creature
// that died this turn, itself included.
func TestCanopyStalkerMustBeBlockedAndGainsLifeWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	stalker := reqCreature(g, me.ID, "Canopy Stalker", canopyStalkerOracle, 4, 2, "haste")
	wall := reqCreature(g, opp.ID, "Spiked Wall", "", 3, 5)
	reqToBlockers(t, g, stalker)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Canopy Stalker must be blocked if able." {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, stalker, wall); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	life := me.Life
	advanceToStepInTurn(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if _, onField := battlefieldCard(g, stalker); onField {
		t.Fatal("the Stalker survived 3 damage")
	}
	if got := me.Life - life; got != 1 {
		t.Errorf("gained %d life, want 1 (the Stalker itself died this turn)", got)
	}
}

// --- Watchdog ---------------------------------------------------------------

// TestWatchdogBlocksAndShrinksAttackersWhileUntapped — it must block,
// creatures attacking its controller get -1/-0 while it is untapped,
// and the shrink is gone once it is tapped.
func TestWatchdogBlocksAndShrinksAttackersWhileUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	bear := reqCreature(g, me.ID, "Bear", "", 3, 3, "haste")
	dog := typedCreature(g, opp.ID, "Watchdog", "Artifact Creature — Dog", watchdogOracle, 1, 2)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Fatalf("power %d before attacking, want 3 — only attackers shrink", got)
	}
	reqToBlockers(t, g, bear)
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("power %d attacking the Watchdog's controller, want 2", got)
	}
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Watchdog must block this combat if able." {
		t.Errorf("sentence = %q", got)
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(dog) })
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d with the Watchdog tapped, want 3", got)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("a tapped Watchdog owes no block: %v", err)
	}
}

// TestWatchdogDoesNotShrinkCreaturesAttackingSomeoneElse — "attacking
// you": a creature attacking another player is untouched.
func TestWatchdogDoesNotShrinkCreaturesAttackingSomeoneElse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	bear := reqCreature(g, me.ID, "Bear", "", 3, 3, "haste")
	typedCreature(g, opp.ID, "Watchdog", "Artifact Creature — Dog", watchdogOracle, 1, 2)
	advanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.DeclareAttacker(bear, third.ID); err != nil {
		t.Fatal(err)
	}
	lockInAttacks(t, g)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("power %d attacking a third player, want 3", got)
	}
}

// --- Nacatl War-Pride ----------------------------------------------------------

// TestNacatlWarPrideCopiesItselfAndMustBeBlockedByExactlyOne — two
// creatures on the defending side make two tapped, attacking token
// copies, each carrying the "exactly one" requirement; a second blocker
// on one is refused; the tokens are exiled at the end step.
func TestNacatlWarPrideCopiesItselfAndMustBeBlockedByExactlyOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	pride := typedCreature(g, me.ID, "Nacatl War-Pride", "Creature — Cat Warrior", nacatlWarPrideOracle, 3, 3, "haste")
	a := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := reqCreature(g, opp.ID, "Wall B", "", 0, 4)

	attackAndTrigger(t, g, pride)
	passPriorityAroundTable(t, g)
	var tokens []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Nacatl War-Pride" && c.InstanceID != pride {
			tokens = append(tokens, c.InstanceID)
			if !c.Tapped || c.AttackingTarget != opp.ID || !IsToken(c) {
				t.Errorf("token %+v: want a tapped token attacking the defender", c.InstanceID)
			}
		}
	}
	if len(tokens) != 2 {
		t.Fatalf("%d token copies, want 2 (the defender controls two creatures)", len(tokens))
	}

	onToBlockers(t, g)
	if err := reqBlock(t, g, pride, a); err != nil {
		t.Fatal(err)
	}
	br := reqRefusal(t, reqBlock(t, g, pride, b))
	if br.Requirement.Kind != game.BlockRequirementExactlyOne {
		t.Errorf("a second blocker on the War-Pride gives up %q, want exactlyOne", br.Requirement.Kind)
	}
	if err := reqBlock(t, g, tokens[0], b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with one blocker on each of two War-Prides: %v", err)
	}
	advanceToStepInTurn(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, id := range tokens {
		if _, ok := battlefieldCard(g, id); ok {
			t.Errorf("token %s survived the end step", id)
		}
	}
	if _, ok := battlefieldCard(g, pride); !ok {
		t.Error("the original War-Pride was exiled with its tokens")
	}
}
