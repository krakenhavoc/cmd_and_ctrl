package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kuja_b_talent_test.go — Artist's Talent (#3213, slice kuja-b).

const (
	kbArtistsTalentOracle = "ea734863-5793-49d6-b10e-109432aec0a3"
	kbBoltOracle          = "4457ed35-7c10-48c8-9776-456485fdf070"
)

// Level 1: casting a noncreature spell offers a discard; discarding
// draws.
func TestArtistsTalentRummagesOnANoncreatureCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushClass(g, me.ID, "Artist's Talent", "Enchantment — Class", kbArtistsTalentOracle)

	castCatalogSpell(t, g, "Divination", "Sorcery", "273b339c-964b-4a18-8eb5-ceb8abcdfd9e", nil)
	hand, graves := me.Hand.Size(), me.Graveyard.Size()
	pitch := me.Hand.Cards[0].InstanceID
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) {
		t.Fatal("the pitched card is not in the graveyard")
	}
	// -1 pitched, +1 rummage draw, +2 Divination.
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want +2", hand, got)
	}
	if got := me.Graveyard.Size(); got != graves+2 {
		t.Errorf("graveyard %d -> %d, want the pitched card and Divination", graves, got)
	}
}

// Declining the discard draws nothing ("if you do").
func TestArtistsTalentDeclinedRummageDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushClass(g, me.ID, "Artist's Talent", "Enchantment — Class", kbArtistsTalentOracle)
	castCatalogSpell(t, g, "Overkill", "Instant", kbOverkillOracle,
		kbCardTarget(pushVanillaCreature(g, g.Seats[1].ID, "Bear", 2, 2)))
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want unchanged: no discard, no draw", hand, got)
	}
}

// Level 2's reduction waits for level 2 and covers noncreature spells
// you cast, not creature spells or an opponent's.
func TestArtistsTalentReductionWaitsForLevelTwo(t *testing.T) {
	seat := uuid.New()
	talent := game.Card{InstanceID: uuid.New(), Name: "Artist's Talent", TypeLine: "Enchantment — Class",
		OracleID: kbArtistsTalentOracle, Owner: seat, Controller: seat}
	if got := len(game.CostModifiersForCard(talent)); got != 0 {
		t.Fatalf("at level 1 the reduction must not exist: %d modifiers", got)
	}
	talent.ClassLevel = 2
	mods := game.CostModifiersForCard(talent)
	if len(mods) != 1 {
		t.Fatalf("at level 2 the reduction must exist: %d modifiers", len(mods))
	}
	m := mods[0]
	q := game.CostQuery{Controller: seat, Source: talent,
		Card: game.Card{Name: "Spell", TypeLine: "Instant"}, FromZone: game.ZoneHand}
	if !m.AppliesTo(q) || m.Amount(q) != 1 {
		t.Error("an instant you cast should cost {1} less")
	}
	q.Card.TypeLine = "Creature — Bear"
	if m.AppliesTo(q) {
		t.Error("a creature spell must not be discounted")
	}
	q.Card.TypeLine = "Instant"
	q.Controller = uuid.New()
	if m.AppliesTo(q) {
		t.Error("an opponent's spell must not be discounted")
	}
}

// Level 3: +2 on noncombat damage to an opponent, and only from level
// 3. Damage to yourself is untouched.
func TestArtistsTalentLevelThreeAddsTwoToNoncombatDamageToOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	class := pushClass(g, me.ID, "Artist's Talent", "Enchantment — Class", kbArtistsTalentOracle)
	advanceToMain(t, g)

	bolt := func(target uuid.UUID) {
		castCatalogSpell(t, g, "Lightning Bolt", "Instant", kbBoltOracle,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: target}})
		passPriorityAroundTable(t, g)
		// The level-1 rummage asks; decline it.
		answerDiscard(t, g, me.ID)
		passPriorityAroundTable(t, g)
	}
	before := opp.Life
	bolt(opp.ID)
	if opp.Life != before-3 {
		t.Fatalf("level 1: opponent life %d -> %d, want -3", before, opp.Life)
	}
	if err := levelUpTo(t, g, me.ID, class, 0, "{2}{R}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	before = opp.Life
	bolt(opp.ID)
	if opp.Life != before-3 {
		t.Fatalf("level 2: opponent life %d -> %d, want -3", before, opp.Life)
	}
	if err := levelUpTo(t, g, me.ID, class, 1, "{2}{R}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	before = opp.Life
	bolt(opp.ID)
	if opp.Life != before-5 {
		t.Errorf("level 3: opponent life %d -> %d, want -5 (3 plus 2)", before, opp.Life)
	}
	mine := me.Life
	bolt(me.ID)
	if me.Life != mine-3 {
		t.Errorf("level 3: my own life %d -> %d, want -3 (not increased)", mine, me.Life)
	}
}
