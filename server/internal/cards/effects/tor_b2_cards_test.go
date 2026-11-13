package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tor_b2_cards_test.go — Torment, batch 2 (#5612). Each test plays the
// real catalog entry and asserts what its text says.

const (
	torB2BalshanOracle      = "7cdbc115-cb15-4ada-8229-d40f0b22dd27"
	torB2PardicOracle       = "9b06cfae-1655-4b0e-a160-9b32f27c7c9a"
	torB2AngelOracle        = "248dcf69-b99e-4a71-b9b3-7fd865f873ab"
	torB2CalmerOracle       = "d4409b15-b9f8-4bdb-9857-5cd70a0d3193"
	torB2WasteAwayOracle    = "d7f25f7f-3fa2-47ba-be33-ba0b08f26c94"
	torB2SpiritFlareOracle  = "d19a2c28-8d09-40bf-aeb1-e699672b009f"
	torB2MorningtideOracle  = "a8432372-672b-4bf7-9000-17e43356f149"
	torB2InvigoratingOracle = "c5c98759-e009-4340-9326-9f5ff70a86bc"
	torB2ChurningEddyOracle = "774e53b2-5aec-43e1-b4fe-8812b8430401"
	torB2ConstrictorOracle  = "c9747b5d-0f6d-4a9a-8921-1d5c0be3bd60"
	torB2CircularOracle     = "b071bbaf-79fa-413b-807b-2cc86c8a8125"
	torB2IsolationOracle    = "87aad9fb-e234-4171-a4ae-c686f613d37e"
	torB2LunacyOracle       = "02a95ea7-cfb9-497c-a004-3e783856d79c"
	torB2ButcherOracle      = "5237a105-60c4-4597-b24e-84c12ce7b01f"
	torB2PetravarkOracle    = "aedf022c-6065-42cb-85bb-225ae1650a5c"
	torB2StalkerOracle      = "0c31edd0-a471-4994-8209-b0a7370fc27b"
)

func TestTorB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		torB2BalshanOracle:      "Balshan Collaborator",
		torB2PardicOracle:       "Pardic Collaborator",
		torB2AngelOracle:        "Angel of Retribution",
		torB2CalmerOracle:       "Nantuko Calmer",
		torB2WasteAwayOracle:    "Waste Away",
		torB2SpiritFlareOracle:  "Spirit Flare",
		torB2MorningtideOracle:  "Morningtide",
		torB2InvigoratingOracle: "Invigorating Falls",
		torB2ChurningEddyOracle: "Churning Eddy",
		torB2ConstrictorOracle:  "Krosan Constrictor",
		torB2CircularOracle:     "Circular Logic",
		torB2IsolationOracle:    "Strength of Isolation",
		torB2LunacyOracle:       "Strength of Lunacy",
		torB2ButcherOracle:      "Faceless Butcher",
		torB2PetravarkOracle:    "Petravark",
		torB2StalkerOracle:      "Slithery Stalker",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// torB2SetColor gives a pushed permanent its colours.
func torB2SetColor(g *game.Game, id uuid.UUID, colors ...string) {
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Colors = colors })
}

// --- keywords and pumps ------------------------------------------------------

func TestTorB2PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Balshan Collaborator", "Creature — Bird Soldier", torB2BalshanOracle, []string{"flying"}},
		{"Pardic Collaborator", "Creature — Human Barbarian", torB2PardicOracle, []string{"first strike"}},
		{"Angel of Retribution", "Creature — Angel", torB2AngelOracle, []string{"flying", "first strike"}},
		{"Krosan Constrictor", "Creature — Snake", torB2ConstrictorOracle, []string{"swampwalk"}},
		{"Slithery Stalker", "Creature — Nightmare Horror", torB2StalkerOracle, []string{"swampwalk"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := blbB1Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			for _, kw := range tc.keywords {
				if !hasKeywordOnBattlefield(t, g, id, kw) {
					t.Errorf("%s does not have %s", tc.name, kw)
				}
			}
		})
	}
}

func TestTorB2CollaboratorsPumpForBlackEachActivation(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Balshan Collaborator", "Creature — Bird Soldier", torB2BalshanOracle},
		{"Pardic Collaborator", "Creature — Human Barbarian", torB2PardicOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := blbB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			advanceToMain(t, g)
			for i := 0; i < 2; i++ {
				floatMana(t, g, me, "{B}")
				if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
					t.Fatalf("activation %d: %v", i+1, err)
				}
				passPriorityAroundTable(t, g)
			}
			if p, tough := ptOf(t, g, id); p != 4 || tough != 4 {
				t.Errorf("after two pumps %d/%d, want 4/4", p, tough)
			}
		})
	}
}

// --- Nantuko Calmer ----------------------------------------------------------

func TestTorB2NantukoCalmerSacrificesToDestroyAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	calmer := blbB1Push(g, me.ID, "Nantuko Calmer", "Creature — Insect Druid", torB2CalmerOracle, 2, 3)
	ench := blbB1Push(g, opp.ID, "Pact", "Enchantment", "", 0, 0)
	bear := blbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	floatMana(t, g, me, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, calmer, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)}); err == nil {
		t.Fatal("a creature was accepted as the target enchantment")
	}
	if err := g.ActivateCatalogAbility(me.ID, calmer, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(ench)}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	if g.Battlefield.Contains(calmer) {
		t.Error("the Calmer was not sacrificed as the ability was activated")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment was not destroyed")
	}
}

func TestTorB2NantukoCalmerGrowsAtThreshold(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	calmer := blbB1Push(g, me.ID, "Nantuko Calmer", "Creature — Insect Druid", torB2CalmerOracle, 2, 3)
	blbMPFillGraveyard(g, opp, 7)
	blbMPFillGraveyard(g, me, 6)
	if p, tough := ptOf(t, g, calmer); p != 2 || tough != 3 {
		t.Fatalf("six cards in your graveyard: %d/%d, want 2/3", p, tough)
	}
	blbMPFillGraveyard(g, me, 1)
	if p, tough := ptOf(t, g, calmer); p != 3 || tough != 4 {
		t.Errorf("seven cards in your graveyard: %d/%d, want 3/4", p, tough)
	}
}

// --- Waste Away --------------------------------------------------------------

func TestTorB2WasteAwayDiscardsToShrinkByFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	advanceToMain(t, g)
	ogre := blbB1Push(g, opp.ID, "Big Ogre", "Creature — Ogre", "", 6, 6)
	waste := pushCatalogHandCard(me, "Waste Away", "Instant", torB2WasteAwayOracle)
	if err := g.CastSpell(me.ID, waste, game.CastSpellParams{Targets: cardTarget(ogre)}); err == nil {
		t.Fatal("Waste Away was cast without its discard")
	}
	fodder := handCard(me, "Fodder", "Sorcery")
	if err := g.CastSpell(me.ID, waste, game.CastSpellParams{Targets: cardTarget(ogre), DiscardIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("cast Waste Away: %v", err)
	}
	if !me.Graveyard.Contains(fodder) {
		t.Fatal("the discard is an additional cost, paid as the spell is cast")
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("the 6/6 is %d/%d after Waste Away, want 1/1", p, tough)
	}
}

// --- Spirit Flare ------------------------------------------------------------

func TestTorB2SpiritFlareTapsYourCreatureToHitAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "My Giant", "Creature — Giant", "", 4, 4)
	attacker := blbB1Push(g, opp.ID, "Their Ogre", "Creature — Ogre", "", 4, 4)
	idle := blbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	targets := func(victim uuid.UUID) []game.TargetRef {
		return []game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: victim, Slot: 1}}
	}
	if err := castCatalogSpellErr(t, g, "Spirit Flare", "Instant", torB2SpiritFlareOracle, targets(idle)); err == nil {
		t.Fatal("a creature that is neither attacking nor blocking was accepted")
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, attacker).AttackingTarget = me.ID })
	castCatalogSpell(t, g, "Spirit Flare", "Instant", torB2SpiritFlareOracle, targets(attacker))
	passPriorityAroundTable(t, g)
	if !findBattlefieldCardForTest(g, mine).Tapped {
		t.Error("your creature was not tapped")
	}
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacker survived 4 damage")
	}

	// Your creature is tapped now, so it is no longer a legal first target.
	other := blbB1Push(g, opp.ID, "Their Elf", "Creature — Elf", "", 1, 1)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, other).AttackingTarget = me.ID })
	if err := castCatalogSpellErr(t, g, "Spirit Flare", "Instant", torB2SpiritFlareOracle, targets(other)); err == nil {
		t.Error("a tapped creature was accepted as the untapped target")
	}
}

// Tapped in response, your creature is an illegal target: nothing is tapped
// and nothing is dealt.
func TestTorB2SpiritFlareDoesNothingWhenYourCreatureWasTappedInResponse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "My Giant", "Creature — Giant", "", 4, 4)
	attacker := blbB1Push(g, opp.ID, "Their Ogre", "Creature — Ogre", "", 4, 4)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, attacker).AttackingTarget = me.ID })
	castCatalogSpell(t, g, "Spirit Flare", "Instant", torB2SpiritFlareOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: attacker, Slot: 1},
	})
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, mine).Tapped = true })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(attacker) {
		t.Error("the attacker was dealt damage though your creature was not tapped by the spell")
	}
}

// Removed from combat in response, the attacker is an illegal target: your
// creature is still tapped, and the attacker is dealt nothing.
func TestTorB2SpiritFlareStillTapsWhenTheAttackerLeftCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "My Giant", "Creature — Giant", "", 4, 4)
	attacker := blbB1Push(g, opp.ID, "Their Ogre", "Creature — Ogre", "", 4, 4)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, attacker).AttackingTarget = me.ID })
	castCatalogSpell(t, g, "Spirit Flare", "Instant", torB2SpiritFlareOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: attacker, Slot: 1},
	})
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, attacker).AttackingTarget = uuid.Nil })
	passPriorityAroundTable(t, g)
	if !findBattlefieldCardForTest(g, mine).Tapped {
		t.Error("your creature was not tapped")
	}
	if !g.Battlefield.Contains(attacker) {
		t.Error("a creature no longer in combat was dealt damage")
	}
}

func TestTorB2SpiritFlareFlashesBackForManaAndLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	mine := blbB1Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	blocker := blbB1Push(g, opp.ID, "Their Elf", "Creature — Elf", "", 2, 2)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, blocker).BlockingTarget = mine })
	life := me.Life
	flare := seedGraveyardCard(t, g, "Spirit Flare", "Instant", torB2SpiritFlareOracle)
	odyB1Flashback(t, g, flare, "{1}{W}",
		game.TargetRef{Kind: game.TargetCard, ID: mine, Slot: 0}, game.TargetRef{Kind: game.TargetCard, ID: blocker, Slot: 1})
	if got := life - me.Life; got != 3 {
		t.Errorf("the flashback cost %d life, want 3", got)
	}
	if g.Battlefield.Contains(blocker) {
		t.Error("the flashed-back Spirit Flare did not kill the blocker")
	}
	offers := game.AlternativeCostsOfferedFromZone(torB2SpiritFlareOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].Label != "Flashback—{1}{W}, Pay 3 life" || offers[0].Life != 3 {
		t.Errorf("graveyard offers = %+v, want one flashback for {1}{W} and 3 life", offers)
	}
}

// --- graveyards --------------------------------------------------------------

func TestTorB2MorningtideExilesEveryGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	cards := []uuid.UUID{
		pushGraveyardCardTyped(me, "Mine", "Sorcery"),
		pushGraveyardCardTyped(opp, "Theirs", "Creature — Bear"),
		pushGraveyardCardTyped(other, "Another's", "Instant"),
	}
	tide := castCatalogSpell(t, g, "Morningtide", "Sorcery", torB2MorningtideOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range cards {
		if !g.Exile.Contains(id) {
			t.Errorf("a graveyard card %s was not exiled", id)
		}
	}
	if !me.Graveyard.Contains(tide) {
		t.Error("Morningtide did not go to its owner's graveyard after resolving")
	}
}

func TestTorB2InvigoratingFallsCountsCreatureCardsInAllGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := blbB1Seats(g)
	pushGraveyardCardTyped(me, "My Bear", "Creature — Bear")
	pushGraveyardCardTyped(me, "My Sorcery", "Sorcery")
	pushGraveyardCardTyped(opp, "Their Elf", "Creature — Elf")
	pushGraveyardCardTyped(other, "Another's Ogre", "Creature — Ogre")
	pushGraveyardCardTyped(other, "Another's Land", "Land")
	life := me.Life
	castCatalogSpell(t, g, "Invigorating Falls", "Sorcery", torB2InvigoratingOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 3 {
		t.Errorf("gained %d life, want 3 (one creature card in each graveyard)", got)
	}
}

// --- Churning Eddy -----------------------------------------------------------

func TestTorB2ChurningEddyReturnsACreatureAndALand(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, other := blbB1Seats(g)
	bear := blbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	land := blbB1Push(g, other.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	refs := []game.TargetRef{{Kind: game.TargetCard, ID: bear, Slot: 0}, {Kind: game.TargetCard, ID: land, Slot: 1}}
	if err := castCatalogSpellErr(t, g, "Churning Eddy", "Sorcery", torB2ChurningEddyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land, Slot: 0}, {Kind: game.TargetCard, ID: bear, Slot: 1}}); err == nil {
		t.Fatal("the targets were accepted in the wrong clauses")
	}
	castCatalogSpell(t, g, "Churning Eddy", "Sorcery", torB2ChurningEddyOracle, refs)
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(bear) || !other.Hand.Contains(land) {
		t.Error("the creature and the land were not returned to their owners' hands")
	}
}

// A target that left in response is skipped; the other still returns.
func TestTorB2ChurningEddyStillReturnsTheSurvivor(t *testing.T) {
	g := newCatalogGame(t)
	_, opp, _ := blbB1Seats(g)
	bear := blbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	land := blbB1Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	castCatalogSpell(t, g, "Churning Eddy", "Sorcery", torB2ChurningEddyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear, Slot: 0}, {Kind: game.TargetCard, ID: land, Slot: 1}})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if !opp.Hand.Contains(land) {
		t.Error("the land was not returned after the creature left")
	}
}

// --- Krosan Constrictor ------------------------------------------------------

func TestTorB2KrosanConstrictorShrinksABlackCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	snake := blbB1Push(g, me.ID, "Krosan Constrictor", "Creature — Snake", torB2ConstrictorOracle, 2, 2)
	black := blbB1Push(g, opp.ID, "Zombie", "Creature — Zombie", "", 3, 3)
	torB2SetColor(g, black, "B")
	green := blbB1Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	torB2SetColor(g, green, "G")
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, snake, 0, game.ActivateAbilityParams{Targets: cardTarget(green)}); err == nil {
		t.Fatal("a green creature was accepted as the target black creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, snake, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(black)}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, black); p != 1 || tough != 3 {
		t.Errorf("the 3/3 is %d/%d, want 1/3", p, tough)
	}
	if !findBattlefieldCardForTest(g, snake).Tapped {
		t.Error("the Constrictor did not tap")
	}
}

// --- Circular Logic ----------------------------------------------------------

func TestTorB2CircularLogicTaxesByTheCardsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	for i := 0; i < 3; i++ {
		pushGraveyardCardTyped(me, "Old Card", "Sorcery")
	}
	for i := 0; i < 5; i++ {
		pushGraveyardCardTyped(opp, "Their Card", "Sorcery")
	}
	victim := castCatalogSpell(t, g, "Divination", "Sorcery", divinationDelveOracle, nil)
	castCatalogSpell(t, g, "Circular Logic", "Instant", torB2CircularOracle, cardTarget(victim))
	passPriorityAroundTable(t, g)
	c := pendingOfKind(g, game.PendingChoicePayUnless)
	if c == nil {
		t.Fatalf("no pay-unless prompt: %+v", g.PendingChoices)
	}
	if c.PayCost != "{3}" {
		t.Errorf("the prompt asks for %q, want {3}: one for each card in the caster's graveyard", c.PayCost)
	}
}

// --- madness Auras -----------------------------------------------------------

func TestTorB2StrengthAurasPumpAndGrantProtection(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, protection string
		power, toughness         int
	}{
		{"Strength of Isolation", torB2IsolationOracle, "protection from black", 3, 4},
		{"Strength of Lunacy", torB2LunacyOracle, "protection from white", 4, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, _, _ := blbB1Seats(g)
			bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
			castCatalogSpell(t, g, tc.name, "Enchantment — Aura", tc.oracle, cardTarget(bear))
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, bear); p != tc.power || tough != tc.toughness {
				t.Errorf("enchanted bear is %d/%d, want %d/%d", p, tough, tc.power, tc.toughness)
			}
			if !hasKeywordOnBattlefield(t, g, bear, tc.protection) {
				t.Errorf("enchanted bear lacks %s", tc.protection)
			}
		})
	}
}

func TestTorB2StrengthAurasHaveMadness(t *testing.T) {
	for _, tc := range []struct{ name, oracle, cost string }{
		{"Strength of Isolation", torB2IsolationOracle, "{W}"},
		{"Strength of Lunacy", torB2LunacyOracle, "{B}"},
		{"Circular Logic", torB2CircularOracle, "{U}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			spec, _ := Lookup(tc.oracle)
			if spec.Madness != tc.cost {
				t.Errorf("madness cost %q, want %q", spec.Madness, tc.cost)
			}
			reps, trigs := madnessAbilitiesFor(tc.oracle)
			if reps != 1 || trigs != 1 {
				t.Errorf("madness halves: %d replacements, %d triggers; want 1 and 1 (CR 702.35a)", reps, trigs)
			}
		})
	}
}

// A discarded Aura is cast from exile for its madness cost, targeting the
// creature it will enchant.
func TestTorB2StrengthOfLunacyIsCastForBAfterADiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := blbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	aura := torB1LootIntoExile(t, g, "Strength of Lunacy", "Enchantment — Aura", torB2LunacyOracle)
	dcMana(t, g, me.ID, "{B}")
	if err := g.CastSpell(me.ID, aura, game.CastSpellParams{
		Strict: true, FromZone: "exile", AlternativeCost: game.AltCostKeyMadness, Targets: cardTarget(bear),
	}); err != nil {
		t.Fatalf("the madness cast: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left, want {B} to be the whole price", n)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the madness-cast Aura is not on the battlefield")
	}
	if p, tough := ptOf(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("enchanted bear is %d/%d, want 4/3", p, tough)
	}
}

// --- the Nightmares ----------------------------------------------------------

func TestTorB2FacelessButcherExilesAndReturnsUnderItsOwnersControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, bear).Controller = me.ID })
	butcher := castHoldCreature(t, g, "Faceless Butcher", "Creature — Nightmare Horror", torB2ButcherOracle, 2, 3)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no pick_target prompt")
	}
	if hasID(p.PickTargetCards, butcher) {
		t.Error("the Butcher offered itself as a target")
	}
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: bear}, true)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Fatal("the creature is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(butcher) })
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		t.Errorf("after the Butcher left, %d Bears under its owner's control, want 1", n)
	}
}

// If the Butcher leaves before its entry trigger resolves, the leave
// trigger finds nothing and the creature is exiled for good.
func TestTorB2FacelessButcherLeavingFirstExilesTheCreatureForGood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	butcher := castHoldCreature(t, g, "Faceless Butcher", "Creature — Nightmare Horror", torB2ButcherOracle, 2, 3)
	pickCard(t, g, me.ID, bear)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(butcher) })
	g.RunStateChecksForTest()
	assistCardSettle(t, g, me.ID, nil, true)
	if g.Battlefield.Contains(bear) || !g.Exile.Contains(bear) {
		t.Error("the creature should stay exiled: its return trigger resolved before the exile did")
	}
}

func TestTorB2PetravarkExilesALandUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	land := blbB1Push(g, opp.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	vark := castHoldCreature(t, g, "Petravark", "Creature — Nightmare Beast", torB2PetravarkOracle, 2, 2)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no pick_target prompt")
	}
	if hasID(p.PickTargetCards, bear) || hasID(p.PickTargetCards, vark) || !hasID(p.PickTargetCards, land) {
		t.Errorf("only the land is a legal target, got %v", p.PickTargetCards)
	}
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: land}, true)
	if !g.Exile.Contains(land) {
		t.Fatal("the land is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(vark) })
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Forest", opp.ID); n != 1 {
		t.Errorf("after Petravark left, %d Forests under its owner's control, want 1", n)
	}
}

func TestTorB2SlitheryStalkerTakesOnlyAGreenOrWhiteOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, _ := blbB1Seats(g)
	white := b12Creature(g, opp.ID, "Their Knight", "Creature — Knight", 2, 2)
	torB2SetColor(g, white, "W")
	blue := b12Creature(g, opp.ID, "Their Drake", "Creature — Drake", 2, 2)
	torB2SetColor(g, blue, "U")
	mine := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	torB2SetColor(g, mine, "G")
	stalker := castHoldCreature(t, g, "Slithery Stalker", "Creature — Nightmare Horror", torB2StalkerOracle, 1, 1)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no pick_target prompt")
	}
	if !hasID(p.PickTargetCards, white) || hasID(p.PickTargetCards, blue) || hasID(p.PickTargetCards, mine) {
		t.Errorf("only the opponent's white creature is legal, got %v", p.PickTargetCards)
	}
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: white}, true)
	if !g.Exile.Contains(white) {
		t.Fatal("the white creature is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(stalker) })
	passPriorityAroundTable(t, g)
	if n := namedOnBattlefield(g, "Their Knight", opp.ID); n != 1 {
		t.Errorf("after the Stalker left, %d Knights under its owner's control, want 1", n)
	}
}
