package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ons_batch3_cards_test.go — Onslaught, batch 3 (#5610, S82): fourteen
// A–C cards built from seams the engine already has. Each test plays
// the real catalog entry.

const (
	ons3AetherChargeOracle      = "15e83068-6253-4c65-8679-7295f3dc2075"
	ons3AirdropCondorOracle     = "25fc274f-d5af-4174-b8e8-c2e65bb4c13f"
	ons3AkromasBlessingOracle   = "a3fc6888-0a4b-4e1d-b709-042b8b6b1a04"
	ons3AphettoVultureOracle    = "aeba5c63-aa80-4f31-90fb-f404c9a3057d"
	ons3AstralSlideOracle       = "70f14b10-8129-4249-a267-8b4be5d9befa"
	ons3AuraExtractionOracle    = "84100106-1f97-4859-aadd-a35796ee000f"
	ons3BattlefieldMedicOracle  = "a600af0b-ea04-4813-b700-f51820868c86"
	ons3BreakOpenOracle         = "1fcec3b2-47ba-4c56-9ce1-a4807e8ad262"
	ons3BrightstoneRitualOracle = "08e90e85-4103-4acb-a8a7-e1329b460aa7"
	ons3CabalSlaverOracle       = "22a189a9-e755-4cbb-b419-5ba2e5acbe1b"
	ons3ChainOfAcidOracle       = "0b1a27bd-bb98-44f8-8357-666fabfeabf0"
	ons3ChainOfPlasmaOracle     = "029d08cb-da4d-4f4e-b667-62712af63260"
	ons3ChokingTethersOracle    = "a60348fa-b182-4667-9e38-15bdf3511f82"
	ons3ComplicateOracle        = "fc09cbc0-2d28-44c6-b9a9-b95e9e58bb1b"
)

func TestOns3CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		ons3AetherChargeOracle:      "Aether Charge",
		ons3AirdropCondorOracle:     "Airdrop Condor",
		ons3AkromasBlessingOracle:   "Akroma's Blessing",
		ons3AphettoVultureOracle:    "Aphetto Vulture",
		ons3AstralSlideOracle:       "Astral Slide",
		ons3AuraExtractionOracle:    "Aura Extraction",
		ons3BattlefieldMedicOracle:  "Battlefield Medic",
		ons3BreakOpenOracle:         "Break Open",
		ons3BrightstoneRitualOracle: "Brightstone Ritual",
		ons3CabalSlaverOracle:       "Cabal Slaver",
		ons3ChainOfAcidOracle:       "Chain of Acid",
		ons3ChainOfPlasmaOracle:     "Chain of Plasma",
		ons3ChokingTethersOracle:    "Choking Tethers",
		ons3ComplicateOracle:        "Complicate",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull {
			t.Errorf("%s: name %q, completeness %v", name, def.Name, def.Completeness)
		}
	}
}

// The two cycling spells with no cycle trigger each cycle for their
// printed cost and draw a card.
func TestOns3CyclingDrawsACard(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle, mana string }{
		{"Akroma's Blessing", "Instant", ons3AkromasBlessingOracle, "{W}"},
		{"Aura Extraction", "Instant", ons3AuraExtractionOracle, "{C}{C}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.SeedLibrary(g.Seats[g.Turn.ActiveSeat], "A", "B")
			effectstest.Ons1Cycle(t, g, tc.name, tc.typeLine, tc.oracle, tc.mana)
		})
	}
}

// --- Aether Charge -----------------------------------------------------------

func TestOns3AetherChargeLetsAnEnteringBeastDealFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Aether Charge", "Enchantment", ons3AetherChargeOracle, false)
	life := opp.Life
	effectstest.TrekB31Enter(t, g, me, game.Card{Name: "My Beast", TypeLine: "Creature — Beast", Power: 3, Toughness: 3})
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want -4", life, opp.Life)
	}
}

func TestOns3AetherChargeIgnoresOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Aether Charge", "Enchantment", ons3AetherChargeOracle, false)
	effectstest.TrekB31Enter(t, g, me, game.Card{Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.TrekB31Enter(t, g, opp, game.Card{Name: "Their Beast", TypeLine: "Creature — Beast", Power: 3, Toughness: 3})
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestPickTarget(g, opp.ID) != nil || g.Stack.Size() != 0 {
		t.Error("a non-Beast or an opponent's Beast triggered Aether Charge")
	}
}

// --- Airdrop Condor ----------------------------------------------------------

func TestOns3AirdropCondorFlingsAGoblinsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	condor := effectstest.B12Push(g, me.ID, "Airdrop Condor", "Creature — Bird", ons3AirdropCondorOracle, 2, 2)
	if !effectstest.EffectiveAbilitiesContain(t, g, condor, "flying") {
		t.Error("the Condor has no flying")
	}
	goblin := effectstest.B12Creature(g, me.ID, "My Goblin", "Creature — Goblin", 3, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, condor, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{bear}, Targets: effectstest.PlayerRefs(opp.ID),
	}); err == nil {
		t.Fatal("a non-Goblin was sacrificed to the Condor")
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, condor, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{goblin}, Targets: effectstest.PlayerRefs(opp.ID), Strict: true,
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(goblin) {
		t.Error("the Goblin was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3 (the Goblin's power)", life, opp.Life)
	}
}

// --- Akroma's Blessing -------------------------------------------------------

func TestOns3AkromasBlessingProtectsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Akroma's Blessing", "Instant", ons3AkromasBlessingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "B")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, mine, "protection from black") {
		t.Error("your creature has no protection from black")
	}
	if effectstest.P1g2HasKeyword(g, theirs, "protection from black") {
		t.Error("an opponent's creature gained protection")
	}
	effectstest.DftMPPassTurn(t, g)
	if effectstest.P1g2HasKeyword(g, mine, "protection from black") {
		t.Error("the protection outlasted the turn")
	}
}

// --- Aphetto Vulture ---------------------------------------------------------

func TestOns3AphettoVultureMayPutItselfOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vulture := effectstest.B12Push(g, me.ID, "Aphetto Vulture", "Creature — Zombie Bird", ons3AphettoVultureOracle, 3, 2)
	bear := effectstest.PushGraveyardCardTyped(me, "Grizzly Bears", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(vulture) })
	g.RunStateChecksForTest()
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the dies trigger asked for no target")
	}
	if effectstest.HasID(pick.PickTargetCards, bear) {
		t.Error("a non-Zombie card was offered")
	}
	effectstest.PickCard(t, g, me.ID, vulture)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != vulture {
		t.Error("the Vulture is not on top of the library")
	}
}

// --- Astral Slide ------------------------------------------------------------

func TestOns3AstralSlideBlinksOnAnyPlayersCycle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	effectstest.PushCatalogPermanent(g, opp.ID, "Astral Slide", "Enchantment", ons3AstralSlideOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.CycleFromHand(t, g, "Aura Extraction", "Instant", ons3AuraExtractionOracle, "{C}{C}")
	effectstest.PickCard(t, g, opp.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || effectstest.B12ZoneOf(g, bear) != game.ZoneExile {
		t.Fatal("the creature was not exiled")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByName(g, "Bear") {
		t.Error("the creature did not return at the end step")
	}
}

// --- Aura Extraction ---------------------------------------------------------

func TestOns3AuraExtractionTucksAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Ons2Opponent(g)
	ench := effectstest.B12Push(g, opp.ID, "Their Enchantment", "Enchantment", "", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Aura Extraction", "Instant", ons3AuraExtractionOracle, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Aura Extraction targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Aura Extraction", "Instant", ons3AuraExtractionOracle, effectstest.CardRefs(ench))
	effectstest.PassPriorityAroundTable(t, g)
	if top := opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID; top != ench {
		t.Error("the enchantment is not on top of its owner's library")
	}
}

// --- Battlefield Medic -------------------------------------------------------

func TestOns3BattlefieldMedicPreventsOnePerClericAtTheTable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	medic := effectstest.B12Push(g, me.ID, "Battlefield Medic", "Creature — Human Cleric", ons3BattlefieldMedicOracle, 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Cleric", "Creature — Human Cleric", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 5)
	if err := g.ActivateCatalogAbility(me.ID, medic, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	if got := effectstest.DamageMarkedOn(g, bear); got != 1 {
		t.Errorf("3 damage into a two-Cleric shield: marked %d, want 1", got)
	}
}

// --- Break Open --------------------------------------------------------------

func TestOns3BreakOpenTurnsAnOpponentsMorphUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	theirs := ons2FaceDown(g, opp.ID, "Cabal Executioner", "Creature — Human Cleric", "b7817579-df4d-41db-bb90-2aabac58f299", 2, 2)
	mine := ons2FaceDown(g, me.ID, "Boneknitter", "Creature — Zombie Cleric", "34b2f465-0b15-428a-a1c2-bcd41af07d01", 1, 1)
	if err := effectstest.CastCatalogSpellErr(t, g, "Break Open", "Instant", ons3BreakOpenOracle, effectstest.CardRefs(mine)); err == nil {
		t.Fatal("Break Open targeted your own face-down creature")
	}
	effectstest.CastCatalogSpell(t, g, "Break Open", "Instant", ons3BreakOpenOracle, effectstest.CardRefs(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Card(t, g, theirs).FaceDownIsPermanent() {
		t.Error("the opponent's creature is still face down")
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("turning it face up made it a new object")
	}
}

// --- Brightstone Ritual ------------------------------------------------------

func TestOns3BrightstoneRitualAddsRedPerGoblin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	effectstest.B12Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Goblin", "Creature — Goblin Warrior", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Brightstone Ritual", "Instant", ons3BrightstoneRitualOracle, nil)
	before := me.ManaPool.Total()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total() - before; got != 2 {
		t.Errorf("added %d mana, want {R}{R} for two Goblins", got)
	}
}

// --- Cabal Slaver ------------------------------------------------------------

func TestOns3CabalSlaverMakesAnyGoblinsVictimDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	effectstest.PushCatalogPermanent(g, opp.ID, "Cabal Slaver", "Creature — Human Cleric", ons3CabalSlaverOracle, false)
	goblin := effectstest.B12Creature(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		t.Fatal("a non-Goblin's damage made its victim discard")
	}
	effectstest.DealCombatDamageToPlayer(g, goblin, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
		t.Fatal("a Goblin's combat damage asked its victim for no discard")
	}
	pick := opp.Hand.Cards[0].InstanceID
	effectstest.AnswerDiscard(t, g, opp.ID, pick)
	if !opp.Graveyard.Contains(pick) {
		t.Error("the chosen card was not discarded")
	}
}

// --- Chain of Acid -----------------------------------------------------------

func TestOns3ChainOfAcidDestroysAndOffersItsVictimACopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	theirs := effectstest.B12Push(g, opp.ID, "Their Artifact", "Artifact", "", 0, 0)
	mine := effectstest.B12Push(g, me.ID, "My Artifact", "Artifact", "", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Chain of Acid", "Sorcery", ons3ChainOfAcidOracle, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Chain of Acid targeted a creature")
	}
	effectstest.CastCatalogSpell(t, g, "Chain of Acid", "Sorcery", ons3ChainOfAcidOracle, effectstest.CardRefs(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the target was not destroyed")
	}
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	effectstest.PickCard(t, g, opp.ID, mine)
	copies := effectstest.CopiesOnStack(g)
	if len(copies) != 1 || copies[0].Controller != opp.ID {
		t.Fatalf("want one copy controlled by the destroyed permanent's controller, got %+v", copies)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the copy did not destroy its new target")
	}
}

// The offer is about a player, not about the permanent dying: an
// indestructible target survives and its controller may still copy.
func TestOns3ChainOfAcidOffersTheCopyWhenNothingDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Ons2Opponent(g)
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Relic", TypeLine: "Artifact", Keywords: []string{"indestructible"},
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Chain of Acid", "Sorcery", ons3ChainOfAcidOracle, effectstest.CardRefs(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Fatal("an indestructible permanent was destroyed")
	}
	effectstest.AnswerMayChoice(t, g, opp.ID, false)
	if len(effectstest.CopiesOnStack(g)) != 0 {
		t.Error("a copy was made after the offer was declined")
	}
}

// --- Chain of Plasma ---------------------------------------------------------

func TestOns3ChainOfPlasmaCopyCostsTheDamagedPlayerACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	life := opp.Life
	hand := len(opp.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Chain of Plasma", "Instant", ons3ChainOfPlasmaOracle, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Fatalf("opponent life %d -> %d, want -3", life, opp.Life)
	}
	c := effectstest.DiscardChoiceFor(g, opp.ID)
	if c == nil || c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Fatalf("the damaged player may discard up to one card: %+v", c)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID)
	if len(opp.Hand.Cards) != hand-1 {
		t.Fatalf("hand %d -> %d, want one discarded", hand, len(opp.Hand.Cards))
	}
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	effectstest.PickPlayer(t, g, opp.ID, me.ID)
	copies := effectstest.CopiesOnStack(g)
	if len(copies) != 1 || copies[0].Controller != opp.ID {
		t.Fatalf("want one copy controlled by the discarding player, got %+v", copies)
	}
	mine := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine-3 {
		t.Errorf("the copy dealt %d to its new target, want 3", mine-me.Life)
	}
}

func TestOns3ChainOfPlasmaNoDiscardNoCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 4)
	effectstest.CastCatalogSpell(t, g, "Chain of Plasma", "Instant", ons3ChainOfPlasmaOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DamageOn(t, g, bear); n != 3 {
		t.Fatalf("the creature has %d damage, want 3", n)
	}
	if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
		t.Fatal("the creature's controller was not offered the discard")
	}
	effectstest.AnswerDiscard(t, g, opp.ID)
	if len(effectstest.CopiesOnStack(g)) != 0 || len(g.PendingChoices) != 0 {
		t.Errorf("declining the discard still offered or made a copy: %+v", g.PendingChoices)
	}
	_ = me
}

// --- Choking Tethers ---------------------------------------------------------

func TestOns3ChokingTethersTapsUpToFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Ons2Opponent(g)
	var ids []uuid.UUID
	for i := 0; i < 4; i++ {
		ids = append(ids, effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	}
	effectstest.CastCatalogSpell(t, g, "Choking Tethers", "Instant", ons3ChokingTethersOracle, effectstest.CardRefs(ids...))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids {
		if !effectstest.TmtB1Card(t, g, id).Tapped {
			t.Error("a target is still untapped")
		}
	}
}

func TestOns3ChokingTethersCycleMayTapACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.SeedLibrary(me, "A", "B")
	hand := len(me.Hand.Cards)
	effectstest.CycleFromHand(t, g, "Choking Tethers", "Instant", ons3ChokingTethersOracle, "{C}{U}")
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TmtB1Card(t, g, bear).Tapped {
		t.Error("the cycle trigger did not tap the creature")
	}
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand %d -> %d, want the Tethers out and a card in", hand, got)
	}
}

// --- Complicate --------------------------------------------------------------

func TestOns3ComplicateCountersUnlessThreeIsPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	effectstest.CastCatalogSpell(t, g, "Complicate", "Instant", ons3ComplicateOracle, effectstest.CardTarget(spell))
	effectstest.PassUntilTaxed(t, g, me.ID)
	if c := effectstest.AnswerPayUnless(t, g, me.ID, false); c == nil {
		t.Fatal("no pay-unless prompt")
	}
	if g.Stack.Contains(spell) {
		t.Error("the spell was not countered")
	}
}

func TestOns3ComplicateCycleTaxesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B")
	spell := effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	id := effectstest.PushCatalogHandCard(me, "Complicate", "Instant", ons3ComplicateOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PickTriggerTarget(t, g, me.ID, spell)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassUntilTaxed(t, g, me.ID)
	effectstest.FillPool(me, 1)
	c := effectstest.AnswerPayUnless(t, g, me.ID, true)
	if c == nil {
		t.Fatal("no pay-unless prompt")
	}
	if !g.Stack.Contains(spell) {
		t.Error("the spell was countered though its {1} was paid")
	}
}
