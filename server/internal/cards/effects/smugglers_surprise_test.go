package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const smugglersSurpriseOracle = "605f719d-01f8-406c-9d4c-3f4992c6a69f"

// ssCard pushes a card with the given P/T onto the top of a zone.
func ssCard(z *game.Zone, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	id := uuid.New()
	z.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
	return id
}

// All three bullets, in printed order: a creature card milled and put
// into hand is one the second bullet may put onto the battlefield, and
// the creatures that arrive with power 4 or greater are the ones the
// third bullet protects.
func TestSmugglersSurpriseAllThreeBulletsSeeEachOther(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b12Creature(g, me.ID, "Small Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 6, 6)

	// The top four, bottom of the four first: Elf, Bolt, Forest, Beast.
	elf := ssCard(me.Library, me.ID, "Milled Elf", "Creature — Elf", 1, 1)
	bolt := ssCard(me.Library, me.ID, "Milled Bolt", "Instant", 0, 0)
	forest := ssCard(me.Library, me.ID, "Milled Forest", "Basic Land — Forest", 0, 0)
	beast := ssCard(me.Library, me.ID, "Milled Beast", "Creature — Beast", 5, 5)
	me.Hand.Cards = nil
	giant := ssCard(me.Hand, me.ID, "Hand Giant", "Creature — Giant", 4, 4)
	ssCard(me.Hand, me.ID, "Hand Sorcery", "Sorcery", 0, 0)

	castModal(t, g, "Smuggler's Surprise", "Instant", smugglersSurpriseOracle, []int{0, 1, 2}, nil)
	passPriorityAroundTable(t, g)

	// Bullet one: the milled creature and land cards, not the instant.
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the first bullet offers the milled creature and land cards")
	}
	for _, id := range []uuid.UUID{elf, forest, beast} {
		if !hasID(pick.ChooseCards, id) {
			t.Errorf("milled creature/land %v is not offered", id)
		}
	}
	if hasID(pick.ChooseCards, bolt) {
		t.Error("the milled instant is offered")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 2 {
		t.Errorf("up to two: bounds [%d,%d]", pick.ChooseMin, pick.ChooseMax)
	}
	answerChooseCards(t, g, me.ID, beast, forest)
	if !me.Hand.Contains(beast) || !me.Hand.Contains(forest) {
		t.Fatal("the two chosen cards go to hand")
	}
	if !me.Graveyard.Contains(elf) || !me.Graveyard.Contains(bolt) {
		t.Error("the rest stay in the graveyard")
	}

	// Bullet two: the creature cards in hand, the one just taken
	// included; the land and the sorcery are not creature cards.
	pick = chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the second bullet offers the creature cards in hand")
	}
	if !hasID(pick.ChooseCards, giant) || !hasID(pick.ChooseCards, beast) {
		t.Fatalf("both creature cards in hand are offered: %v", pick.ChooseCards)
	}
	if hasID(pick.ChooseCards, forest) {
		t.Error("a land card is offered to the creature bullet")
	}
	answerChooseCards(t, g, me.ID, giant, beast)
	if !g.Battlefield.Contains(giant) || !g.Battlefield.Contains(beast) {
		t.Fatal("both creatures are put onto the battlefield")
	}

	// Bullet three: power 4 or greater, yours.
	for _, id := range []uuid.UUID{giant, beast} {
		ab := effectiveAbilities(t, g, id)
		if !hasAbility(ab, "hexproof") || !hasAbility(ab, "indestructible") {
			t.Errorf("a power-4+ creature you control gains hexproof and indestructible: %v", ab)
		}
	}
	if hasAbility(effectiveAbilities(t, g, small), "hexproof") {
		t.Error("a 2-power creature is not protected")
	}
	if hasAbility(effectiveAbilities(t, g, theirs), "hexproof") {
		t.Error("an opponent's creature is not protected")
	}
}

// The third bullet alone: no prompts, only the grant.
func TestSmugglersSurpriseProtectionAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	big := b12Creature(g, me.ID, "Big Beast", "Creature — Beast", 4, 4)
	small := b12Creature(g, me.ID, "Small Bear", "Creature — Bear", 3, 3)

	castModal(t, g, "Smuggler's Surprise", "Instant", smugglersSurpriseOracle, []int{2}, nil)
	passPriorityAroundTable(t, g)

	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("the protection bullet asks nothing")
	}
	ab := effectiveAbilities(t, g, big)
	if !hasAbility(ab, "hexproof") || !hasAbility(ab, "indestructible") {
		t.Errorf("the 4-power creature is protected: %v", ab)
	}
	if hasAbility(effectiveAbilities(t, g, small), "indestructible") {
		t.Error("a 3-power creature is not")
	}
}

// Declining both "you may"s is a legal answer to each, and the spell
// still finishes.
func TestSmugglersSurpriseDecliningBothPicks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beast := ssCard(me.Library, me.ID, "Milled Beast", "Creature — Beast", 5, 5)
	me.Hand.Cards = nil
	giant := ssCard(me.Hand, me.ID, "Hand Giant", "Creature — Giant", 4, 4)

	castModal(t, g, "Smuggler's Surprise", "Instant", smugglersSurpriseOracle, []int{0, 1}, nil)
	passPriorityAroundTable(t, g)

	answerChooseCards(t, g, me.ID)
	if !me.Graveyard.Contains(beast) {
		t.Error("declining leaves the milled creature in the graveyard")
	}
	answerChooseCards(t, g, me.ID)
	if !me.Hand.Contains(giant) {
		t.Error("declining leaves the creature in hand")
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("no prompt is left open")
	}
}
