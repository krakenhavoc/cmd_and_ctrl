package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers_names2_test.go — the two `_____` Auras (ADR 0208, #4307).

const (
	fightTheBlankFightOracle = "5f589a11-1601-4a17-92d6-58e9dd9b00f2"
	lastVoyageOracle         = "54ddf766-b528-488d-9b98-8a101bdb5de9"
)

func stkPlace(g *game.Game, owner uuid.UUID, zone *game.Zone, name, typeLine string, p, tough int) uuid.UUID {
	id := uuid.New()
	zone.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Power: p, Toughness: tough,
		Owner: owner, Controller: owner,
	})
	return id
}

// Fight the _____ Fight: a name sticker, then "when you do" the enchanted
// creature fights up to one target creature you don't control.
func TestStkFightTheBlankFightFights(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	opp := g.Seats[1]
	mine := auraBear(g, me.ID)
	theirs := p2Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Fight the _____ Fight", "Enchantment — Aura", fightTheBlankFightOracle,
		[]game.TargetRef{p2Card(mine)})
	passPriorityAroundTable(t, g)
	aura := latestChooseCardsFor(g, me.ID)
	if aura == nil {
		t.Fatal("the enters trigger asked for no sticker")
	}
	if err := g.ResolveChooseCards(aura.ID, me.ID, aura.ChooseCards[:1]); err != nil {
		t.Fatalf("choose the Aura: %v", err)
	}
	which := latestOptionPickFor(g, me.ID)
	if err := g.ResolveOptionPick(which.ID, me.ID, stkOption(t, which, "“Brushwagg”")); err != nil {
		t.Fatalf("choose the sticker: %v", err)
	}
	if where := latestOptionPickFor(g, me.ID); where != nil {
		if err := g.ResolveOptionPick(where.ID, me.ID, 0); err != nil {
			t.Fatalf("position: %v", err)
		}
	}
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the reflexive fight raised no target prompt")
	}
	if prompt.PickTargetMin != 0 || prompt.PickTargetMax != 1 {
		t.Errorf("targets %d..%d, want up to one", prompt.PickTargetMin, prompt.PickTargetMax)
	}
	if hasID(prompt.PickTargetCards, mine) {
		t.Error("a creature you control is offered as a target")
	}
	if err := g.ResolvePickTargets(prompt.ID, me.ID, []game.TargetRef{p2Card(theirs)}); err != nil {
		t.Fatalf("pick the target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, theirs) {
		t.Error("their Bear survived the fight (2 damage to a 2/2)")
	}
	if !onBattlefield(g, mine) {
		t.Error("my Bear died: the nine-letter sticker makes it 2/4, and it takes 2")
	}
}

// Declined: no fight.
func TestStkFightTheBlankFightDeclinedDoesNotFight(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	opp := g.Seats[1]
	mine := auraBear(g, me.ID)
	theirs := p2Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	castCatalogSpell(t, g, "Fight the _____ Fight", "Enchantment — Aura", fightTheBlankFightOracle,
		[]game.TargetRef{p2Card(mine)})
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a fight target prompt is open though no sticker went on")
	}
	if !onBattlefield(g, theirs) || !onBattlefield(g, mine) {
		t.Error("a fight happened")
	}
}

// +0/+2 for each name sticker on the Aura with eight or more letters.
func TestStkFightTheBlankFightPumpsForLongNameStickers(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	mine := auraBear(g, me.ID)
	aura := castCatalogSpell(t, g, "Fight the _____ Fight", "Enchantment — Aura", fightTheBlankFightOracle,
		[]game.TargetRef{p2Card(mine)})
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	base := effectiveToughness(t, g, mine)
	// "Night" has five letters: nothing. "Brushwagg" nine and "Ringmaster"
	// ten: +2 each.
	stkGive(t, g, me, aura, "“Night”")
	if got := effectiveToughness(t, g, mine); got != base {
		t.Errorf("toughness after a five-letter sticker = %d, want %d", got, base)
	}
	stkGive(t, g, me, aura, "“Brushwagg”")
	if got := effectiveToughness(t, g, mine); got != base+2 {
		t.Errorf("toughness after Brushwagg = %d, want %d", got, base+2)
	}
	stkGive(t, g, me, aura, "“Ringmaster”")
	if got := effectiveToughness(t, g, mine); got != base+4 {
		t.Errorf("toughness after Ringmaster = %d, want %d", got, base+4)
	}
	if got := effectivePower(t, g, mine); got != 2 {
		t.Errorf("power = %d, want 2 (+0 per sticker)", got)
	}
}

// The reflexive trigger is a restore point.
func TestStkFightTheBlankFightReflexiveTriggerOnTheStackIsARestorePoint(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	opp := g.Seats[1]
	mine := auraBear(g, me.ID)
	theirs := p2Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	aura := castCatalogSpell(t, g, "Fight the _____ Fight", "Enchantment — Aura", fightTheBlankFightOracle,
		[]game.TargetRef{p2Card(mine)})
	passPriorityAroundTable(t, g)
	stkPutOn(t, g, me, aura, "“Brushwagg”", 0)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("the reflexive fight raised no target prompt")
	}
	if err := g.ResolvePickTargets(prompt.ID, me.ID, []game.TargetRef{p2Card(theirs)}); err != nil {
		t.Fatalf("pick the target: %v", err)
	}
	// The reflexive trigger is on the stack with its target chosen.
	g = restoreRoundTrip(t, g, true)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, theirs) {
		t.Error("their Bear survived the fight after a restore")
	}
}

func stkIsAura(g *game.Game, id uuid.UUID) bool {
	is := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c, ok := g.LookupCardForEffect(id); ok {
			is = c.IsAura()
		}
	})
	return is
}

// Last Voyage of the _____: a name sticker, it becomes an Aura, a creature
// card from the graveyard returns and the Aura is attached to it.
func TestStkLastVoyageReturnsACreatureAndEnchantsIt(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	dead := stkPlace(g, me.ID, me.Graveyard, "Dead Bear", "Creature — Bear", 2, 2)
	id := castAndResolveCreature(t, g, "Last Voyage of the _____", "Enchantment", lastVoyageOracle)
	passPriorityAroundTable(t, g)
	stkPutOn(t, g, me, id, "“Night”", 0)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no question asks which creature card returns")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != dead || pick.ChooseMin != 1 {
		t.Fatalf("candidates = %v min=%d, want just the Dead Bear, mandatory", pick.ChooseCards, pick.ChooseMin)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{dead}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, dead) {
		t.Fatal("the creature card did not return")
	}
	if !stkIsAura(g, id) {
		t.Error("the enchantment did not become an Aura")
	}
	if !attachedToNow(t, g, id, dead) {
		t.Error("the Aura is not attached to the returned creature")
	}
	// "Night" has five letters: +2/+0.
	if got := effectivePower(t, g, dead); got != 4 {
		t.Errorf("power = %d, want 4 (2 + 2 for the five-letter sticker)", got)
	}
	// A sticker with more than seven letters does not count.
	stkGive(t, g, me, id, "“Ringmaster”")
	if got := effectivePower(t, g, dead); got != 4 {
		t.Errorf("power after Ringmaster = %d, want 4 (ten letters: no bonus)", got)
	}
}

// "When this Aura leaves the battlefield, sacrifice enchanted creature."
func TestStkLastVoyageSacrificesTheCreatureWhenItLeaves(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	dead := stkPlace(g, me.ID, me.Graveyard, "Dead Bear", "Creature — Bear", 2, 2)
	id := castAndResolveCreature(t, g, "Last Voyage of the _____", "Enchantment", lastVoyageOracle)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline the sticker: %v", err)
	}
	pick = latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("declining the sticker skipped the return")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{dead}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, dead) {
		t.Fatal("the creature card did not return")
	}
	b25Destroy(g, id)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, dead) {
		t.Error("the enchanted creature was not sacrificed")
	}
	if !me.Graveyard.Contains(dead) {
		t.Error("the sacrificed creature is not in the graveyard")
	}
}

// An Aura whose creature dies is illegal (CR 704.5m): it goes, and its
// leaves trigger finds the creature gone and sacrifices nothing.
func TestStkLastVoyageAuraGoesWhenItsCreatureDies(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	dead := stkPlace(g, me.ID, me.Graveyard, "Dead Bear", "Creature — Bear", 2, 2)
	id := castAndResolveCreature(t, g, "Last Voyage of the _____", "Enchantment", lastVoyageOracle)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	_ = g.ResolveChooseCards(pick.ID, me.ID, nil)
	pick = latestChooseCardsFor(g, me.ID)
	_ = g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{dead})
	passPriorityAroundTable(t, g)
	// The creature dies first: the Aura is illegal (CR 704.5m) and goes.
	b25Destroy(g, dead)
	settleStateBasedActions(t, g)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, id) {
		t.Error("the Aura stayed after its creature died")
	}
	if !me.Graveyard.Contains(dead) {
		t.Error("the creature is not in the graveyard")
	}
}

// With no creature card in the graveyard the Aura attaches to nothing and
// is put into the graveyard (CR 704.5m).
func TestStkLastVoyageWithNoCreatureCardGoesToTheGraveyard(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	id := castAndResolveCreature(t, g, "Last Voyage of the _____", "Enchantment", lastVoyageOracle)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if onBattlefield(g, id) {
		t.Error("an Aura attached to nothing stayed on the battlefield")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the Aura is not in the graveyard")
	}
}

// The question of which creature returns is a restore point, and the Aura
// form is data the snapshot carries.
func TestStkLastVoyageReturnQuestionIsARestorePoint(t *testing.T) {
	g, me := stkTableWith(t, 0, "Night Brushwagg Ringmaster")
	dead := stkPlace(g, me.ID, me.Graveyard, "Dead Bear", "Creature — Bear", 2, 2)
	id := castAndResolveCreature(t, g, "Last Voyage of the _____", "Enchantment", lastVoyageOracle)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	if !stkIsAura(g, id) {
		t.Error("the restored enchantment is not an Aura")
	}
	pick = latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the restored table lost the return question")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{dead}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, dead) || !attachedToNow(t, g, id, dead) {
		t.Error("after a restore the creature did not return enchanted")
	}
}
