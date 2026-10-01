package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// disturb_batch_c_test.go — the harder disturb cards (ADR 0107 §4,
// #1855): Dorothea, Katilda, Faithbound Judge, Mirrorhall Mimic, Brine
// Comber and Covert Cutpurse. The shared cast and exile behaviour is in
// disturb_test.go; this file proves each card's own abilities, on rows
// driven down the import road.

func dorotheaRow() cards.Card {
	return disturbRow(dorotheaOracleID, []string{"W", "U"},
		disturbFace{"Dorothea, Vengeful Victim", "Legendary Creature — Spirit", "{W}{U}", "4", "4"},
		disturbFace{"Dorothea's Retribution", "Enchantment — Aura", "", "", ""})
}

func katildaRow() cards.Card {
	return disturbRow(katildaOracleID, []string{"W"},
		disturbFace{"Katilda, Dawnhart Martyr", "Legendary Creature — Spirit Warlock", "{1}{W}{W}", "*", "*"},
		disturbFace{"Katilda's Rising Dawn", "Legendary Enchantment — Aura", "", "", ""})
}

// faithboundJudgeRow carries Scryfall's keyword list, as the dump does,
// so the importer stamps a printed defender onto the card's baseline —
// the road the Judge's "as though it didn't have defender" has to beat.
func faithboundJudgeRow() cards.Card {
	row := disturbRow(faithboundJudgeOracleID, []string{"W"},
		disturbFace{"Faithbound Judge", "Creature — Spirit Soldier", "{1}{W}{W}", "4", "4"},
		disturbFace{"Sinner's Judgment", "Enchantment — Aura Curse", "", "", ""})
	row.Keywords = []string{"Defender", "Flying", "Vigilance", "Disturb"}
	return row
}

func mirrorhallMimicRow() cards.Card {
	return disturbRow(mirrorhallMimicOracleID, []string{"U"},
		disturbFace{"Mirrorhall Mimic", "Creature — Spirit", "{3}{U}", "0", "0"},
		disturbFace{"Ghastly Mimicry", "Enchantment — Aura", "", "", ""})
}

func brineComberRow() cards.Card {
	return disturbRow(brineComberOracleID, []string{"W", "U"},
		disturbFace{"Brine Comber", "Creature — Spirit", "{1}{W}{U}", "1", "1"},
		disturbFace{"Brinebound Gift", "Enchantment — Aura", "", "", ""})
}

func covertCutpurseRow() cards.Card {
	return disturbRow(covertCutpurseOracleID, []string{"B"},
		disturbFace{"Covert Cutpurse", "Creature — Human Rogue", "{2}{B}", "2", "1"},
		disturbFace{"Covetous Geist", "Creature — Spirit Rogue", "", "2", "2"})
}

// dbcSettled runs a row down the import road onto the battlefield as a
// permanent that has been there since the turn began (no summoning
// sickness), front face up.
func dbcSettled(g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	c := deck.ToGameCard(row, false)
	c.Owner, c.Controller = p.ID, p.ID
	return pushBattlefieldCardWithTimestamp(g, c)
}

// dbcCastFromHand imports a row into the active seat's hand and casts
// its front face at a main phase.
func dbcCastFromHand(t *testing.T, g *game.Game, row cards.Card, p *game.Player, targets ...game.TargetRef) uuid.UUID {
	t.Helper()
	id := importToHand(row, p)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", row.Name, err)
	}
	return id
}

func spiritTokens(g *game.Game) int { return len(battlefieldIDsNamed(g, "Spirit")) }

// --- Dorothea, Vengeful Victim -----------------------------------

// "When Dorothea attacks or blocks, sacrifice it at end of combat": she
// deals her damage, and is sacrificed as the end of combat step begins.
func TestDorotheaIsSacrificedAtEndOfCombatAfterAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := dbcSettled(g, dorotheaRow(), me)
	before := opp.Life

	declareAttack(t, g, opp.ID, id)
	if n := triggersOnStackFrom(g, id); n != 1 {
		t.Fatalf("Dorothea's attack trigger: %d on the stack, want 1", n)
	}
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("defender at %d, want %d", opp.Life, before-4)
	}
	if zone, _ := cardWhere(g, me, id); zone != "battlefield" {
		t.Fatalf("Dorothea left before end of combat (now in %s)", zone)
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if zone, _ := cardWhere(g, me, id); zone != "graveyard" {
		t.Fatalf("Dorothea is in %s after end of combat, want the graveyard", zone)
	}
}

// The block half of the same trigger.
func TestDorotheaIsSacrificedAtEndOfCombatAfterBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := auraBear(g, me.ID)
	dorothea := dbcSettled(g, dorotheaRow(), opp)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(dorothea, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if zone, _ := cardWhere(g, opp, dorothea); zone != "graveyard" {
		t.Fatalf("a blocking Dorothea is in %s after end of combat, want the graveyard", zone)
	}
}

// Dorothea's Retribution, disturbed onto a creature, gives it the
// attack trigger: a 4/4 flying Spirit enters tapped and attacking, and
// is sacrificed at end of combat. The enchanted creature is not.
func TestDorotheasRetributionMakesAnAttackingSpiritForOneCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := auraBear(g, me.ID)
	aura := importToGraveyard(t, g, dorotheaRow(), me)
	disturb(t, g, me, aura, game.TargetRef{Kind: game.TargetCard, ID: bear})
	passPriorityAroundTable(t, g)
	if zone, c := cardWhere(g, me, aura); zone != "battlefield" || c.Name != "Dorothea's Retribution" {
		t.Fatalf("the disturbed Aura is in %s as %q", zone, c.Name)
	}

	before := opp.Life
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	spirits := battlefieldIDsNamed(g, "Spirit")
	if len(spirits) != 1 {
		t.Fatalf("%d Spirit tokens, want 1", len(spirits))
	}
	tok := cardByID(g, spirits[0])
	if !tok.Tapped || tok.AttackingTarget != opp.ID || effectivePower(t, g, spirits[0]) != 4 ||
		!hasEffectiveKeyword(t, g, spirits[0], "flying") {
		t.Errorf("the Spirit is not a tapped, attacking 4/4 flyer: %+v", tok)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != before-6 {
		t.Errorf("defender at %d, want %d (the bear and the Spirit)", opp.Life, before-6)
	}
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if n := spiritTokens(g); n != 0 {
		t.Errorf("%d Spirits left after end of combat, want 0", n)
	}
	if zone, _ := cardWhere(g, me, bear); zone != "battlefield" {
		t.Errorf("the enchanted bear is in %s, want the battlefield", zone)
	}
}

// --- Katilda, Dawnhart Martyr ------------------------------------

// Katilda's power and toughness count the permanents you control that
// are Spirits and/or enchantments — herself among them — and each such
// permanent once.
func TestKatildaCountsSpiritsAndEnchantments(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := dbcSettled(g, katildaRow(), me)
	if p, tough := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 1 || tough != 1 {
		t.Fatalf("Katilda alone is %d/%d, want 1/1", p, tough)
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spirit Enchantment",
		TypeLine: "Enchantment Creature — Spirit", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Plain Enchantment",
		TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID})
	auraBear(g, me.ID)
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Spirit",
		TypeLine: "Creature — Spirit", Power: 1, Toughness: 1, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	if p := effectivePower(t, g, id); p != 3 {
		t.Errorf("Katilda with a Spirit enchantment and an enchantment is %d power, want 3", p)
	}
	for _, kw := range []string{"flying", "lifelink", "protection from Vampires"} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Katilda lacks %s", kw)
		}
	}
}

// Katilda's Rising Dawn gives its creature the three keywords and +X/+X,
// counting itself as an enchantment you control.
func TestKatildasRisingDawnGrantsAndPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	aura := importToGraveyard(t, g, katildaRow(), me)
	disturb(t, g, me, aura, game.TargetRef{Kind: game.TargetCard, ID: bear})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the enchanted bear is %d/%d, want 3/3 (the Aura counts itself)", p, tough)
	}
	for _, kw := range []string{"flying", "lifelink", "protection from Vampires"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the enchanted bear lacks %s", kw)
		}
	}
}

// --- Faithbound Judge // Sinner's Judgment -----------------------

// dbcNextUpkeepOf walks to seat's next upkeep (stepping off the current
// one first) and settles its triggers.
func dbcNextUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	advanceToUpkeepOf(t, g, seat)
	passPriorityAroundTable(t, g)
}

// The Judge gains a judgment counter each upkeep while it has two or
// fewer, stops at three, and with three can attack despite defender.
func TestFaithboundJudgeCountsToThreeThenAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := dbcSettled(g, faithboundJudgeRow(), me)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); !errors.Is(err, game.ErrDefender) {
		t.Fatalf("the Judge with no counters attacked: err = %v, want ErrDefender", err)
	}
	for want := 1; want <= 3; want++ {
		dbcNextUpkeepOf(t, g, 0)
		if got := cardByID(g, id).Counters[judgmentCounter]; got != want {
			t.Fatalf("after upkeep %d the Judge has %d judgment counters", want, got)
		}
	}
	dbcNextUpkeepOf(t, g, 0)
	if got := cardByID(g, id).Counters[judgmentCounter]; got != 3 {
		t.Errorf("a fourth upkeep left %d judgment counters, want 3 (the intervening if)", got)
	}
	if hasEffectiveKeyword(t, g, id, "defender") {
		t.Error("with three counters the Judge still can't attack")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Errorf("the Judge with three counters could not attack: %v", err)
	}
}

// Sinner's Judgment, disturbed onto an opponent, makes them lose the
// game at its controller's third upkeep.
func TestSinnersJudgmentMakesTheEnchantedPlayerLoseOnTheThirdCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	curse := importToGraveyard(t, g, faithboundJudgeRow(), me)
	disturb(t, g, me, curse, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	passPriorityAroundTable(t, g)
	zone, c := cardWhere(g, me, curse)
	if zone != "battlefield" || c.Name != "Sinner's Judgment" || c.AttachedTo.ID != opp.ID {
		t.Fatalf("the Curse is in %s as %q attached to %v", zone, c.Name, c.AttachedTo)
	}
	for upkeep := 1; upkeep <= 2; upkeep++ {
		dbcNextUpkeepOf(t, g, 0)
		if opp.Eliminated {
			t.Fatalf("the enchanted player lost at upkeep %d", upkeep)
		}
	}
	dbcNextUpkeepOf(t, g, 0)
	if !opp.Eliminated {
		t.Errorf("the enchanted player is still in the game with %d judgment counters", cardByID(g, curse).Counters[judgmentCounter])
	}
}

// --- Mirrorhall Mimic // Ghastly Mimicry -------------------------

// The Mimic enters as a copy of the chosen creature, and is a Spirit
// besides.
func TestMirrorhallMimicCopiesAndIsASpirit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bears := seedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	id := dbcCastFromHand(t, g, mirrorhallMimicRow(), me)
	resolveWithCopyChoice(t, g, bears)
	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Grizzly Bears" || effectivePower(t, g, id) != 2 {
		t.Errorf("the Mimic is %q at %d power, want a 2-power Grizzly Bears", got.Name, effectivePower(t, g, id))
	}
	if !got.HasSubtype("Bear") || !got.HasSubtype("Spirit") {
		t.Errorf("the Mimic's type line is %q, want a Bear and a Spirit", got.TypeLine)
	}
}

// Ghastly Mimicry copies its creature each of its controller's upkeeps,
// and the token is a Spirit in addition.
func TestGhastlyMimicryCopiesTheEnchantedCreatureEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	aura := importToGraveyard(t, g, mirrorhallMimicRow(), me)
	disturb(t, g, me, aura, game.TargetRef{Kind: game.TargetCard, ID: bear})
	passPriorityAroundTable(t, g)

	dbcNextUpkeepOf(t, g, 0)
	bears := battlefieldIDsNamed(g, "Bear")
	if len(bears) != 2 {
		t.Fatalf("%d Bears after the upkeep, want the bear and its token copy", len(bears))
	}
	for _, id := range bears {
		if id == bear {
			continue
		}
		c := cardByID(g, id)
		if !c.IsToken() || !c.HasSubtype("Spirit") || effectivePower(t, g, id) != 2 {
			t.Errorf("the copy is %q (%q), want a 2/2 token that is a Spirit too", c.Name, c.TypeLine)
		}
	}
}

// --- Brine Comber // Brinebound Gift -----------------------------

// Brine Comber makes a Spirit as it enters and whenever an Aura spell
// targets it.
func TestBrineComberMakesSpiritsOnEntryAndAuraTargeting(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := dbcCastFromHand(t, g, brineComberRow(), me)
	passPriorityAroundTable(t, g)
	if n := spiritTokens(g); n != 1 {
		t.Fatalf("%d Spirits after Brine Comber entered, want 1", n)
	}

	// A disturbed Aura spell targeting the Comber is an Aura spell.
	armaments := importToGraveyard(t, g, drogskolInfantryRow(), me)
	disturb(t, g, me, armaments, game.TargetRef{Kind: game.TargetCard, ID: id})
	passPriorityAroundTable(t, g)
	if n := spiritTokens(g); n != 2 {
		t.Errorf("%d Spirits after an Aura spell targeted Brine Comber, want 2", n)
	}
	if _, c := cardWhere(g, me, armaments); c.AttachedTo.ID != id {
		t.Errorf("the Aura is not on Brine Comber")
	}
}

// Brinebound Gift makes a Spirit as it enters, and whenever an Aura
// spell targets the creature it enchants.
func TestBrineboundGiftMakesSpiritsOnEntryAndForItsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := auraBear(g, me.ID)
	gift := importToGraveyard(t, g, brineComberRow(), me)
	disturb(t, g, me, gift, game.TargetRef{Kind: game.TargetCard, ID: bear})
	passPriorityAroundTable(t, g)
	if n := spiritTokens(g); n != 1 {
		t.Fatalf("%d Spirits after Brinebound Gift entered, want 1", n)
	}
	armaments := importToGraveyard(t, g, drogskolInfantryRow(), me)
	disturb(t, g, me, armaments, game.TargetRef{Kind: game.TargetCard, ID: bear})
	passPriorityAroundTable(t, g)
	if n := spiritTokens(g); n != 2 {
		t.Errorf("%d Spirits after an Aura spell targeted the enchanted bear, want 2", n)
	}
}

// --- Covert Cutpurse // Covetous Geist ---------------------------

// Covert Cutpurse's entry destroys a creature an opponent controls that
// was dealt damage this turn; an undamaged one is not a legal target.
func TestCovertCutpurseDestroysADamagedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	hurt := auraBear(g, opp.ID)
	fresh := auraBear(g, opp.ID)
	mine := auraBear(g, me.ID)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, mine, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	var legal []bool
	g.ReadSnapshot(func() {
		pred := And(OpponentControls(), DealtDamageThisTurn())
		for _, id := range []uuid.UUID{hurt, fresh, mine} {
			legal = append(legal, pred(g, me.ID, cardByID(g, id)))
		}
	})
	if !legal[0] || legal[1] || legal[2] {
		t.Fatalf("legal targets (damaged, undamaged, mine) = %v, want [true false false]", legal)
	}

	dbcCastFromHand(t, g, covertCutpurseRow(), me)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if !hasID(p.PickTargetCards, hurt) || hasID(p.PickTargetCards, fresh) || hasID(p.PickTargetCards, mine) {
			t.Errorf("the Cutpurse offers %v, want the damaged opposing creature only", p.PickTargetCards)
		}
		pickCard(t, g, me.ID, hurt)
	}
	passPriorityAroundTable(t, g)
	if zone, _ := cardWhere(g, opp, hurt); zone != "graveyard" {
		t.Errorf("the damaged creature is in %s, want the graveyard", zone)
	}
	if zone, _ := cardWhere(g, opp, fresh); zone != "battlefield" {
		t.Errorf("the undamaged creature is in %s, want the battlefield", zone)
	}
}

// Disturbed, the Cutpurse is Covetous Geist, a flying deathtouch 2/2.
func TestCovetousGeistIsTheDisturbedBackFace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := importToGraveyard(t, g, covertCutpurseRow(), me)
	disturb(t, g, me, id)
	passPriorityAroundTable(t, g)
	zone, c := cardWhere(g, me, id)
	if zone != "battlefield" || c.Name != "Covetous Geist" {
		t.Fatalf("the disturbed Cutpurse is in %s as %q", zone, c.Name)
	}
	for _, kw := range []string{"flying", "deathtouch"} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Covetous Geist lacks %s", kw)
		}
	}
	destroy(t, g, id)
	if zone, _ := cardWhere(g, me, id); zone != "exile" {
		t.Errorf("a destroyed Covetous Geist went to %s, want exile", zone)
	}
}
