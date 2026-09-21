package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// aang_swift_savior_test.go — issues #325 and #343.
//
// #325 is the whole of this file: "when I cast my commander Aang
// Swift Savior he entered the battlefield but his enter-the-
// battlefield trigger never triggered to airbend a creature or
// spell." It never triggered because no spec existed for the card,
// and no spec was worth writing while the card imported as a
// costless colourless 0/0 — which is #343's other symptom and what
// the face model fixes.

// aangSwiftSaviorCard is the card as deck.toGameCard now imports it:
// per-face printed data with face 0 materialised. The point of
// building it this way rather than stamping the flat fields is that
// the pre-ADR-0034 import produced ManaCost "", Power 0, Toughness 0
// and TypeLine "Legendary Creature — Human Avatar Ally // Legendary
// Creature — Avatar Spirit Ally", and every assertion below would
// have been meaningless against that.
func aangSwiftSaviorCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   aangSwiftSaviorOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{
				Name:      "Aang, Swift Savior",
				TypeLine:  "Legendary Creature — Human Avatar Ally",
				ManaCost:  "{1}{W}{U}",
				Colors:    []string{"U", "W"},
				Power:     2,
				Toughness: 3,
			},
			{
				Name:      "Aang and La, Ocean's Fury",
				TypeLine:  "Legendary Creature — Avatar Spirit Ally",
				Colors:    []string{"U", "W"},
				Power:     5,
				Toughness: 5,
			},
		},
	}
	c.SetFace(0)
	return c
}

// TestAangSwiftSaviorImportsAsARealCard is #343's "he shows as a
// 0/0". The front face's printed stats, cost and colours all live on
// card_faces[0]; the top level carries nulls.
func TestAangSwiftSaviorImportsAsARealCard(t *testing.T) {
	c := aangSwiftSaviorCard(uuid.New())
	if c.Name != "Aang, Swift Savior" {
		t.Errorf("Name = %q", c.Name)
	}
	if c.ManaCost != "{1}{W}{U}" {
		t.Errorf("ManaCost = %q, want {1}{W}{U} — the top-level cost "+
			"is null on a transform printing, so he was FREE", c.ManaCost)
	}
	if c.Power != 2 || c.Toughness != 3 {
		t.Errorf("P/T = %d/%d, want 2/3 — #343 reported a 0/0",
			c.Power, c.Toughness)
	}
	eff := c.Effective()
	if len(eff.Types) != 1 || eff.Types[0] != "Creature" {
		t.Errorf("types = %v, want [Creature] — the concatenated type "+
			"line used to yield the literal types \"//\" and \"—\"", eff.Types)
	}
	if len(eff.Colors) != 2 {
		t.Errorf("colours = %v, want two", eff.Colors)
	}
	// CR 712.4: his back face is reached by transforming him, never
	// by casting it. The Waterbend {8} verb is a later PR.
	if faces := c.CastableFaces(); len(faces) != 1 || faces[0] != 0 {
		t.Errorf("CastableFaces = %v, want [0]", faces)
	}
	// But the back face IS imported and on the card, waiting for it.
	if c.Faces[1].Power != 5 || c.Faces[1].Toughness != 5 {
		t.Errorf("back face P/T = %d/%d, want 5/5",
			c.Faces[1].Power, c.Faces[1].Toughness)
	}
}

// TestAangSwiftSaviorAirbendsOnEntry is #325. Aang enters, the
// trigger goes on the stack, the controller says yes and picks an
// opponent's creature, and that creature is exiled with its owner's
// {2} rebuy.
func TestAangSwiftSaviorAirbendsOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Their Blocker")

	aang := aangSwiftSaviorCard(me.ID)
	me.Hand.PushTop(aang)
	if err := g.CastSpell(me.ID, aang.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !g.Battlefield.Contains(aang.InstanceID) {
		t.Fatal("Aang did not resolve onto the battlefield")
	}

	answerLatestTriggerPrompt(t, g, me.ID, true)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	assertAirbent(t, g, victim, opp.ID)
}

// "Up to one" means declining is a real answer, not a failed cast.
func TestAangSwiftSaviorDeclinedAirbendsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushCreatureToBattlefieldForTest(g, opp.ID, "Their Blocker")

	aang := aangSwiftSaviorCard(me.ID)
	me.Hand.PushTop(aang)
	if err := g.CastSpell(me.ID, aang.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(victim) {
		t.Error("declining the airbend exiled something anyway")
	}
}

// TestAangSwiftSaviorSpecKeysOnFaceZero pins the catalog key: the
// front face takes the BARE oracle ID, so this spec is found the
// same way every single-faced spec is. Since ADR 0079 the back face
// is reachable too (aang_and_la_oceans_fury.go), through the
// Waterbend {8} ability's TransformThis rather than an entry.
func TestAangSwiftSaviorSpecKeysOnFaceZero(t *testing.T) {
	c := aangSwiftSaviorCard(uuid.New())
	key := game.CatalogKey(c)
	if key != aangSwiftSaviorOracle {
		t.Fatalf("front-face key = %q, want the bare oracle ID", key)
	}
	spec, ok := Lookup(key)
	if !ok {
		t.Fatal("no spec registered for Aang, Swift Savior")
	}
	if len(spec.Triggered) != 1 {
		t.Errorf("%d triggered abilities, want 1 (the ETB airbend)",
			len(spec.Triggered))
	}
	if len(spec.Activated) != 1 {
		t.Errorf("%d activated abilities, want 1 (Waterbend {8}: Transform Aang)",
			len(spec.Activated))
	}
	c.SetFace(1)
	backKey := game.CatalogKey(c)
	if backKey != aangSwiftSaviorOracle+"#1" {
		t.Fatalf("back-face key = %q, want the oracle ID plus #1", backKey)
	}
	backSpec, ok := Lookup(backKey)
	if !ok {
		t.Fatal("no spec registered for Aang and La, Ocean's Fury")
	}
	if len(backSpec.Triggered) != 1 {
		t.Errorf("%d triggered abilities on the back face, want 1 (the attack trigger)",
			len(backSpec.Triggered))
	}
}

// TestAangTransformsThroughWaterbend is the Waterbend {8} ability's
// card-level proof: paying the flat mana cost (the declared
// simplification — no tap-artifacts-and-creatures discount) flips
// Aang onto Aang and La, Ocean's Fury IN PLACE (CR 712.18 — same
// object, ADR 0079's first verb, not the Sagas' exile-and-return).
func TestAangTransformsThroughWaterbend(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aang := aangSwiftSaviorCard(me.ID)
	me.Hand.PushTop(aang)
	if err := g.CastSpell(me.ID, aang.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Decline the ETB airbend; it isn't what this test is about.
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if err := g.ActivateCatalogAbility(me.ID, aang.InstanceID, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility (Waterbend): %v", err)
	}
	passPriorityAroundTable(t, g)

	card := battlefieldCardFor(g, aang.InstanceID)
	if card == nil {
		t.Fatal("Aang left the battlefield — TransformThis is not a zone change")
	}
	if card.ActiveFace != 1 {
		t.Fatalf("ActiveFace = %d, want the back face", card.ActiveFace)
	}
	if card.Name != "Aang and La, Ocean's Fury" {
		t.Errorf("name = %q, want Aang and La, Ocean's Fury", card.Name)
	}
	if !card.IsCreature() || card.Power != 5 || card.Toughness != 5 {
		t.Errorf("P/T = %d/%d, want 5/5", card.Power, card.Toughness)
	}
}

// TestAangAndLaCountersEachTappedCreatureOnAttack is the back face's
// own ability, reachable only because the transform above works:
// "Whenever Aang and La attack, put a +1/+1 counter on each tapped
// creature you control." Aang and La has no vigilance, so by the time
// the trigger RESOLVES it is tapped from attacking and is itself one
// of the "each tapped creature" the effect counts.
func TestAangAndLaCountersEachTappedCreatureOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]

	aang := aangSwiftSaviorCard(me.ID)
	aang.SetFace(1)
	g.Battlefield.PushTop(aang)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == aang.InstanceID {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	tapped := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped Friend", TypeLine: "Creature — Bear",
		Owner: me.ID, Controller: me.ID, Power: 2, Toughness: 2, Tapped: true,
	})
	untapped := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Untapped Friend", TypeLine: "Creature — Bear",
		Owner: me.ID, Controller: me.ID, Power: 2, Toughness: 2,
	})

	declareAttack(t, g, opp.ID, aang.InstanceID)
	passPriorityAroundTable(t, g)

	if got := countersOn(g, tapped, "+1/+1"); got != 1 {
		t.Errorf("the already-tapped creature has %d +1/+1 counters, want 1", got)
	}
	if got := countersOn(g, untapped, "+1/+1"); got != 0 {
		t.Errorf("the untapped creature has %d +1/+1 counters, want 0", got)
	}
	if got := countersOn(g, aang.InstanceID, "+1/+1"); got != 1 {
		t.Errorf("Aang and La (tapped from attacking) has %d +1/+1 counters, want 1", got)
	}
}
