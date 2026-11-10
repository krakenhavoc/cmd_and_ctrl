package game

import (
	"testing"

	"github.com/google/uuid"
)

// off_battlefield_self_test.go — #5427: the readers of a card's own "as
// long as this card isn't on the battlefield, it's a ..." (CR 113.6c),
// without a catalog: the index is fed a test key.

const offBattlefieldSelfTestKey = "test-off-battlefield-self"

func offBattlefieldSelfTestCard() Card {
	IndexOffBattlefieldSelf(offBattlefieldSelfTestKey, &OffBattlefieldSelf{
		Types: []string{"Creature"}, Subtypes: []string{"Insect"}, Power: 1, Toughness: 1,
	})
	return Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Walker",
		OracleID: offBattlefieldSelfTestKey,
	}
}

func TestOffBattlefieldSelfIsReadByEveryColdReader(t *testing.T) {
	c := offBattlefieldSelfTestCard()
	if !c.IsCreature() || !c.IsPlaneswalker() || !c.HasSubtype("insect") || !c.HasSubtype("Walker") {
		t.Errorf("creature %v planeswalker %v Insect %v Walker %v", c.IsCreature(), c.IsPlaneswalker(), c.HasSubtype("insect"), c.HasSubtype("Walker"))
	}
	if c.HasCardType("artifact") || c.HasSubtype("Elf") {
		t.Error("only the declared words are added")
	}
	if c.PrintedIsCreature() {
		t.Error("the printed type line is not changed")
	}
	eff := c.Effective()
	if len(eff.Types) != 2 || eff.Types[0] != "Planeswalker" || eff.Types[1] != "Creature" ||
		len(eff.Subtypes) != 2 || eff.Subtypes[1] != "Insect" {
		t.Errorf("effective types %v — %v", eff.Types, eff.Subtypes)
	}
	if eff.Power != 1 || eff.Toughness != 1 || c.BasePower() != 1 || c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("power and toughness %d/%d, base %d", eff.Power, eff.Toughness, c.BasePower())
	}
}

// The printed cache's entry is shared and read-only: applying the
// sentence must not write it (printed_cache.go). The test binary's
// cache check fails the second read if it did.
func TestOffBattlefieldSelfDoesNotWriteThePrintedCache(t *testing.T) {
	c := offBattlefieldSelfTestCard()
	c.stampPrinted()
	first := c.Effective()
	second := c.Effective()
	if len(first.Types) != 2 || len(second.Types) != 2 {
		t.Fatalf("types %v then %v", first.Types, second.Types)
	}
	if _, types, subtypes := printedTypeParts(&c); len(types) != 1 || len(subtypes) != 1 {
		t.Errorf("the cached printed parts grew: %v — %v", types, subtypes)
	}
}

// A permanent is not reached, with a warm layer cache or a cold one
// (a transform or a phasing drops the cache and keeps the entry
// timestamp), and neither is a last-known copy of it.
func TestOffBattlefieldSelfDoesNotReachAPermanent(t *testing.T) {
	cold := offBattlefieldSelfTestCard()
	cold.EnteredBattlefieldAt = 1
	if cold.IsCreature() || cold.HasSubtype("Insect") || cold.Effective().Power != 0 || len(cold.Effective().Types) != 1 {
		t.Error("a permanent with a cold layer cache reads as a creature")
	}
	warm := offBattlefieldSelfTestCard()
	ch := printedFresh(&warm)
	warm.effective = &ch
	if warm.IsCreature() || warm.HasSubtype("Insect") || warm.BasePower() != 0 {
		t.Error("a permanent with a layer cache reads as a creature")
	}
}

// A face-down object has no text (CR 708.2a), and a card without the
// declaration is untouched.
func TestOffBattlefieldSelfNeedsTheCardsOwnText(t *testing.T) {
	down := offBattlefieldSelfTestCard()
	down.FaceDown = true
	down.FaceDownKind = FaceDownManifested
	if down.HasSubtype("Insect") {
		t.Error("a face-down permanent has the sentence")
	}
	other := Card{InstanceID: uuid.New(), Name: "Other", TypeLine: "Legendary Planeswalker — Walker", OracleID: "no-such-key"}
	if other.IsCreature() || other.Effective().Power != 0 {
		t.Error("a card without the declaration is a creature")
	}
}
