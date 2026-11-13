package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// uds_b2_cards_test.go — Urza's Destiny, batch 2 (#5621). Each test plays
// the real catalog entry through the engine and asserts what its text
// says.

const (
	udsB2GoblinBerserkerOracle   = "5f71f205-9551-416a-bba6-55bc65d11c85"
	udsB2GoliathBeetleOracle     = "86ab4400-fbdd-4c18-a893-441286a9d7d0"
	udsB2WildColosOracle         = "cb6b8ce3-9f9d-418c-94b0-c4469a254938"
	udsB2TormentedAngelOracle    = "176eeafd-0482-4fa4-9174-10f8b9494d9b"
	udsB2VoiceOfDutyOracle       = "b5d5b9c9-accc-4dbe-aaf4-c806f3e1c61c"
	udsB2VoiceOfReasonOracle     = "372e2d6e-427c-4152-b936-27dcf536ca84"
	udsB2PlatedSpiderOracle      = "15163c99-9388-4225-b707-4e057b2adfbc"
	udsB2SquirmingMassOracle     = "2c526c61-6cdb-440c-b4c8-b6d1503cef73"
	udsB2ElvishLookoutOracle     = "9e9aea99-1eee-4bb6-9043-a1ac2884d071"
	udsB2MetathranSoldierOracle  = "b723fcec-f208-4397-a178-e1e47b729cf1"
	udsB2HulkingOgreOracle       = "a95aa9ba-fa61-4a3d-a493-1ea790bd20ad"
	udsB2AncientSilverbackOracle = "5f12134c-d626-4a5e-8653-700681a491fc"
	udsB2AttritionOracle         = "a85ee0cf-39ce-49f2-874a-f6f54ca1775f"
	udsB2BloodshotCyclopsOracle  = "84d47913-53ac-4d4d-bac3-bc0306d43e12"
	udsB2BraidwoodSextantOracle  = "b44816b4-44ef-446d-a254-b4b7bc015a42"
	udsB2MentalDisciplineOracle  = "b22080d6-a9ed-4bdd-a604-058e0e3e9463"
	udsB2FodderCannonOracle      = "aaf171bd-a4bb-4ce4-836a-da193c94f42e"
	udsB2BlizzardElementalOracle = "dd489f30-d1cf-4236-8390-9c21cd8659e0"
	udsB2MantisEngineOracle      = "57bea272-5b38-444e-acc2-6104af2ece22"
	udsB2TemporalAdeptOracle     = "fd9b5462-6578-4bf6-8026-23eaf5af3eee"
	udsB2MasterHealerOracle      = "7695b61a-8d44-44b1-9783-ea3114111c3e"
	udsB2SerraAdvocateOracle     = "4c67292a-7fcc-4646-8a06-99a7b4b3e005"
	udsB2ElvishPiperOracle       = "872ef617-7cdb-4a7a-85f5-efc9e12bece5"
	udsB2OppositionOracle        = "bf0b252d-1295-44e4-bac3-113b5732a2a9"
	udsB2FieldSurgeonOracle      = "730809d7-cb8c-495c-8972-7e85b967358e"
	udsB2RofellosOracle          = "3015882b-4897-4d2c-8e33-6731c85a0d03"
	udsB2RescueOracle            = "f0ae687a-7222-4521-a514-ba10373fed7e"
	udsB2SoulFeastOracle         = "8186fd80-015f-470c-9e1c-cbf45764a057"
	udsB2SolidarityOracle        = "7b094a2b-2b69-4a9b-9829-c1aabc53f75e"
	udsB2FlameJetOracle          = "8a9d6a30-e2c2-4da5-ad3b-d64fb2b5543c"
	udsB2RecklessAbandonOracle   = "f24d5187-0f80-4528-bf44-03ae146f3551"
	udsB2MultanisDecreeOracle    = "b233473d-d22b-45e4-9fe7-55d1170a788b"
	udsB2DiseaseCarriersOracle   = "ccdd6376-9b74-4d13-9f80-87d0bd3d4a2a"
	udsB2MarkerBeetlesOracle     = "4b1614fd-2810-4d4d-a0d6-934f27adbe65"
	udsB2PlagueDogsOracle        = "a7cc69fd-0c67-44fc-be9d-d8f5b1bd54f5"
	udsB2GoblinMasonsOracle      = "cfec9d17-2d11-4b89-8b30-84bd25a38466"
	udsB2ThievingMagpieOracle    = "2f074636-6a36-46a3-b872-add79e19aa8b"
	udsB2TelepathicSpiesOracle   = "8162730e-2573-4bc0-827f-5b7b3dbf3bff"
	udsB2MaskOfLawAndGraceOracle = "21474195-5d91-466d-8874-4818c900dea7"
	udsB2CapashenStandardOracle  = "75510429-41bb-409e-b6fe-04a8bb174c6b"
	udsB2IlluminatedWingsOracle  = "7a22389a-34e7-4726-a551-f6fbc225cefe"
)

func TestUdsB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		udsB2GoblinBerserkerOracle:   "Goblin Berserker",
		udsB2GoliathBeetleOracle:     "Goliath Beetle",
		udsB2WildColosOracle:         "Wild Colos",
		udsB2TormentedAngelOracle:    "Tormented Angel",
		udsB2VoiceOfDutyOracle:       "Voice of Duty",
		udsB2VoiceOfReasonOracle:     "Voice of Reason",
		udsB2PlatedSpiderOracle:      "Plated Spider",
		udsB2SquirmingMassOracle:     "Squirming Mass",
		udsB2ElvishLookoutOracle:     "Elvish Lookout",
		udsB2MetathranSoldierOracle:  "Metathran Soldier",
		udsB2HulkingOgreOracle:       "Hulking Ogre",
		udsB2AncientSilverbackOracle: "Ancient Silverback",
		udsB2AttritionOracle:         "Attrition",
		udsB2BloodshotCyclopsOracle:  "Bloodshot Cyclops",
		udsB2BraidwoodSextantOracle:  "Braidwood Sextant",
		udsB2MentalDisciplineOracle:  "Mental Discipline",
		udsB2FodderCannonOracle:      "Fodder Cannon",
		udsB2BlizzardElementalOracle: "Blizzard Elemental",
		udsB2MantisEngineOracle:      "Mantis Engine",
		udsB2TemporalAdeptOracle:     "Temporal Adept",
		udsB2MasterHealerOracle:      "Master Healer",
		udsB2SerraAdvocateOracle:     "Serra Advocate",
		udsB2ElvishPiperOracle:       "Elvish Piper",
		udsB2OppositionOracle:        "Opposition",
		udsB2FieldSurgeonOracle:      "Field Surgeon",
		udsB2RofellosOracle:          "Rofellos, Llanowar Emissary",
		udsB2RescueOracle:            "Rescue",
		udsB2SoulFeastOracle:         "Soul Feast",
		udsB2SolidarityOracle:        "Solidarity",
		udsB2FlameJetOracle:          "Flame Jet",
		udsB2RecklessAbandonOracle:   "Reckless Abandon",
		udsB2MultanisDecreeOracle:    "Multani's Decree",
		udsB2DiseaseCarriersOracle:   "Disease Carriers",
		udsB2MarkerBeetlesOracle:     "Marker Beetles",
		udsB2PlagueDogsOracle:        "Plague Dogs",
		udsB2GoblinMasonsOracle:      "Goblin Masons",
		udsB2ThievingMagpieOracle:    "Thieving Magpie",
		udsB2TelepathicSpiesOracle:   "Telepathic Spies",
		udsB2MaskOfLawAndGraceOracle: "Mask of Law and Grace",
		udsB2CapashenStandardOracle:  "Capashen Standard",
		udsB2IlluminatedWingsOracle:  "Illuminated Wings",
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

// udsB2Activate floats `mana` and activates row `row` of `id` under the
// strict mana gate.
func udsB2Activate(t *testing.T, g *game.Game, who *game.Player, id uuid.UUID, row int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	floatForTest(g, who, mana)
	params.Strict = true
	return g.ActivateCatalogAbility(who.ID, id, row, params)
}

// --- keyword bodies ---------------------------------------------------------

func TestUdsB2KeywordCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Goblin Berserker", "Creature — Goblin Berserker", udsB2GoblinBerserkerOracle, []string{"first strike", "haste"}},
		{"Goliath Beetle", "Creature — Insect", udsB2GoliathBeetleOracle, []string{"trample"}},
		{"Wild Colos", "Creature — Goat Beast", udsB2WildColosOracle, []string{"haste"}},
		{"Tormented Angel", "Creature — Angel", udsB2TormentedAngelOracle, []string{"flying"}},
		{"Voice of Duty", "Creature — Angel", udsB2VoiceOfDutyOracle, []string{"flying", "protection from green"}},
		{"Voice of Reason", "Creature — Angel", udsB2VoiceOfReasonOracle, []string{"flying", "protection from blue"}},
		{"Plated Spider", "Creature — Spider", udsB2PlatedSpiderOracle, []string{"reach"}},
		{"Squirming Mass", "Creature — Horror", udsB2SquirmingMassOracle, []string{"fear"}},
		{"Elvish Lookout", "Creature — Elf", udsB2ElvishLookoutOracle, []string{"shroud"}},
		{"Blizzard Elemental", "Creature — Elemental", udsB2BlizzardElementalOracle, []string{"flying"}},
		{"Serra Advocate", "Creature — Angel", udsB2SerraAdvocateOracle, []string{"flying"}},
		{"Thieving Magpie", "Creature — Bird", udsB2ThievingMagpieOracle, []string{"flying"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			for _, kw := range tc.keywords {
				if !hasKeywordOnBattlefield(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Shroud refuses even its controller's own targeted spell (CR 702.18a).
func TestUdsB2ElvishLookoutCantBeTargeted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elf := b12Push(g, me.ID, "Elvish Lookout", "Creature — Elf", udsB2ElvishLookoutOracle, 1, 1)
	if err := castCatalogSpellErr(t, g, "Rescue", "Instant", udsB2RescueOracle, cardTarget(elf)); err == nil {
		t.Error("Rescue targeted a creature with shroud")
	}
}

// --- blocking: can't block, can't be blocked, fear, reach -------------------

func TestUdsB2HulkingOgreCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Raider", "Creature — Human", 2, 2)
	ogre := b12Push(g, opp.ID, "Hulking Ogre", "Creature — Ogre", udsB2HulkingOgreOracle, 3, 3)
	other := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	brAttack(t, g, attacker)
	brRefusal(t, g.DeclareBlocker(ogre, attacker), game.BlockReasonCantBlock)
	if err := g.DeclareBlocker(other, attacker); err != nil {
		t.Fatalf("another creature can't block: %v", err)
	}
}

func TestUdsB2MetathranSoldierCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	soldier := b12Push(g, me.ID, "Metathran Soldier", "Creature — Metathran Soldier", udsB2MetathranSoldierOracle, 1, 1)
	blocker := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	brAttack(t, g, soldier)
	brRefusal(t, g.DeclareBlocker(blocker, soldier), game.BlockReasonCantBeBlocked)
}

func TestUdsB2SquirmingMassHasFear(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mass := b12Push(g, me.ID, "Squirming Mass", "Creature — Horror", udsB2SquirmingMassOracle, 1, 1)
	green := tmpB1Push(g, opp.ID, "Green Bear", "Creature — Bear", "", 2, 2, "G")
	black := tmpB1Push(g, opp.ID, "Black Knight", "Creature — Human Knight", "", 2, 2, "B")
	brAttack(t, g, mass)
	brRefusal(t, g.DeclareBlocker(green, mass), game.BlockReasonFear)
	if err := g.DeclareBlocker(black, mass); err != nil {
		t.Fatalf("a black creature can't block a creature with fear: %v", err)
	}
}

func TestUdsB2PlatedSpiderBlocksAFlier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := sthB2Keyworded(g, me.ID, "Bird", "Creature — Bird", 2, 2, "flying")
	spider := b12Push(g, opp.ID, "Plated Spider", "Creature — Spider", udsB2PlatedSpiderOracle, 4, 4)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	brAttack(t, g, bird)
	brRefusal(t, g.DeclareBlocker(bear, bird), game.BlockReasonFlying)
	if err := g.DeclareBlocker(spider, bird); err != nil {
		t.Fatalf("Plated Spider can't block a flier: %v", err)
	}
}

// --- Ancient Silverback ---------------------------------------------------------

func TestUdsB2AncientSilverbackRegeneratesForG(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ape := b12Push(g, me.ID, "Ancient Silverback", "Creature — Ape", udsB2AncientSilverbackOracle, 6, 5)
	if err := udsB2Activate(t, g, me, ape, 0, "", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("regenerated with no mana")
	}
	if err := udsB2Activate(t, g, me, ape, 0, "G", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := regenShieldsOn(t, g, ape); n != 1 {
		t.Fatalf("shields = %d, want 1", n)
	}
	destroyIt(t, g, ape)
	if c, ok := battlefieldCard(g, ape); !ok || !c.Tapped {
		t.Fatal("the shield should have saved it, tapped")
	}
}

// --- Attrition, Bloodshot Cyclops, Fodder Cannon ------------------------------------

func TestUdsB2AttritionSacrificesToDestroyANonblackCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attrition := pushCatalogPermanent(g, me.ID, "Attrition", "Enchantment", udsB2AttritionOracle, true)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
	other := b12Creature(g, me.ID, "Other Fodder", "Creature — Rat", 1, 1)
	black := tmpB1Push(g, opp.ID, "Black Knight", "Creature — Human Knight", "", 2, 2, "B")
	green := tmpB1Push(g, opp.ID, "Green Bear", "Creature — Bear", "", 2, 2, "G")
	if err := udsB2Activate(t, g, me, attrition, 0, "B", game.ActivateAbilityParams{
		Targets: cardRefs(black), SacrificeIDs: []uuid.UUID{fodder},
	}); err == nil {
		t.Fatal("a black creature was accepted as the target")
	}
	if err := udsB2Activate(t, g, me, attrition, 0, "", game.ActivateAbilityParams{
		Targets: cardRefs(green), SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green) {
		t.Error("the nonblack creature was not destroyed")
	}
	// No {T}: a second activation the same turn.
	green2 := tmpB1Push(g, opp.ID, "Green Bear Two", "Creature — Bear", "", 2, 2, "G")
	if err := udsB2Activate(t, g, me, attrition, 0, "B", game.ActivateAbilityParams{
		Targets: cardRefs(green2), SacrificeIDs: []uuid.UUID{other},
	}); err != nil {
		t.Fatalf("second activation: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(green2) {
		t.Error("the second activation did not destroy its target")
	}
}

func TestUdsB2BloodshotCyclopsFlingsTheSacrificedPower(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sick := pushCatalogPermanent(g, me.ID, "Bloodshot Cyclops", "Creature — Cyclops Giant", udsB2BloodshotCyclopsOracle, true)
	ogre := b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	if err := udsB2Activate(t, g, me, sick, 0, "", game.ActivateAbilityParams{
		Targets: playerRefs(opp.ID), SacrificeIDs: []uuid.UUID{ogre},
	}); err == nil {
		t.Fatal("a summoning-sick Cyclops paid {T} (CR 302.6)")
	}
	cyclops := b12Push(g, me.ID, "Bloodshot Cyclops", "Creature — Cyclops Giant", udsB2BloodshotCyclopsOracle, 4, 4)
	life := opp.Life
	if err := udsB2Activate(t, g, me, cyclops, 0, "", game.ActivateAbilityParams{
		Targets: playerRefs(opp.ID), SacrificeIDs: []uuid.UUID{ogre},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(ogre) {
		t.Error("the Ogre was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3 (the Ogre's power)", life, opp.Life)
	}
}

func TestUdsB2FodderCannonShootsATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cannon := pushCatalogPermanent(g, me.ID, "Fodder Cannon", "Artifact", udsB2FodderCannonOracle, true)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Giant", 4, 4)
	if err := udsB2Activate(t, g, me, cannon, 0, "CCC", game.ActivateAbilityParams{
		Targets: cardRefs(victim), SacrificeIDs: []uuid.UUID{fodder},
	}); err == nil {
		t.Fatal("activated for {3}; it costs {4}")
	}
	if err := udsB2Activate(t, g, me, cannon, 0, "C", game.ActivateAbilityParams{
		Targets: cardRefs(victim), SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) || g.Battlefield.Contains(victim) {
		t.Error("the fodder is sacrificed and the 4/4 takes 4")
	}
	if !g.Battlefield.Contains(cannon) || !b13Tapped(t, g, cannon) {
		t.Error("the Cannon stays, tapped")
	}
}

// --- Braidwood Sextant -------------------------------------------------------------

func TestUdsB2BraidwoodSextantFetchesABasicLandToHand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sextant := pushCatalogPermanent(g, me.ID, "Braidwood Sextant", "Artifact", udsB2BraidwoodSextantOracle, true)
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")
	urborg := stapleLibraryCard(me, "Urborg, Tomb of Yawgmoth", "Legendary Land")
	if err := udsB2Activate(t, g, me, sextant, 0, "CC", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(sextant) {
		t.Error("the Sextant is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if err := func() error {
		beginSearchIfAsked(t, g, c)
		return g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{urborg})
	}(); err == nil {
		t.Error("a nonbasic land was accepted")
	}
	answerSearchByID(t, g, me.ID, forest)
	if !me.Hand.Contains(forest) {
		t.Error("the basic land did not reach the hand")
	}
}

// --- Mental Discipline ----------------------------------------------------------------

func TestUdsB2MentalDisciplineDiscardsToDraw(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	md := pushCatalogPermanent(g, me.ID, "Mental Discipline", "Enchantment", udsB2MentalDisciplineOracle, true)
	junk := handCardFull(me, "Junk", "Sorcery", "{1}", "", nil)
	if err := udsB2Activate(t, g, me, md, 0, "U", game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{junk}}); err == nil {
		t.Fatal("activated for {U}; it costs {1}{U}")
	}
	if err := udsB2Activate(t, g, me, md, 0, "C", game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{junk}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(junk) {
		t.Error("the card was not discarded as the cost")
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- Blizzard Elemental, Mantis Engine ------------------------------------------------

func TestUdsB2BlizzardElementalUntapsItself(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	el := b12Push(g, me.ID, "Blizzard Elemental", "Creature — Elemental", udsB2BlizzardElementalOracle, 5, 5)
	b16Tap(g, el)
	if err := udsB2Activate(t, g, me, el, 0, "CC", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated for {2}; it costs {3}{U}")
	}
	if err := udsB2Activate(t, g, me, el, 0, "CU", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b13Tapped(t, g, el) {
		t.Error("the Elemental is still tapped")
	}
}

func TestUdsB2MantisEngineGainsFlyingAndFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	engine := pushCatalogPermanent(g, me.ID, "Mantis Engine", "Artifact Creature — Insect", udsB2MantisEngineOracle, true)
	for row, kw := range []string{"flying", "first strike"} {
		if hasKeywordOnBattlefield(t, g, engine, kw) {
			t.Fatalf("the Engine has %s before the activation", kw)
		}
		if err := udsB2Activate(t, g, me, engine, row, "CC", game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate row %d: %v", row, err)
		}
		passPriorityAroundTable(t, g)
		if !hasKeywordOnBattlefield(t, g, engine, kw) {
			t.Errorf("the Engine did not gain %s", kw)
		}
	}
}

// --- Temporal Adept, Rescue -----------------------------------------------------------

func TestUdsB2TemporalAdeptBouncesAnyPermanent(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	adept := b12Push(g, me.ID, "Temporal Adept", "Creature — Human Wizard", udsB2TemporalAdeptOracle, 1, 1)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := udsB2Activate(t, g, me, adept, 0, "CUU", game.ActivateAbilityParams{Targets: cardRefs(rock)}); err == nil {
		t.Fatal("activated without {U}{U}{U}")
	}
	if err := udsB2Activate(t, g, me, adept, 0, "U", game.ActivateAbilityParams{Targets: cardRefs(rock)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !opp.Hand.Contains(rock) {
		t.Error("the artifact did not return to its owner's hand")
	}
}

func TestUdsB2RescueReturnsOnlyYourOwnPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Rescue", "Instant", udsB2RescueOracle, cardTarget(theirs)); err == nil {
		t.Fatal("Rescue targeted a permanent you don't control")
	}
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Rescue", "Instant", udsB2RescueOracle, cardTarget(mine))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || !me.Hand.Contains(mine) {
		t.Error("your creature did not return to your hand")
	}
}

// --- Master Healer, Field Surgeon -------------------------------------------------------

func TestUdsB2MasterHealerPreventsFour(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	healer := b12Push(g, me.ID, "Master Healer", "Creature — Human Cleric", udsB2MasterHealerOracle, 1, 4)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 6)
	if err := udsB2Activate(t, g, me, healer, 0, "", game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 5) })
	if got := damageMarkedOn(g, bear); got != 1 {
		t.Errorf("5 damage into a 4-shield: marked %d, want 1", got)
	}
}

func TestUdsB2FieldSurgeonTapsACreatureToPreventOne(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	// Summoning sick: tapping a creature is not the {T} symbol (CR 302.6).
	surgeon := pushCatalogPermanent(g, me.ID, "Field Surgeon", "Creature — Human Cleric", udsB2FieldSurgeonOracle, true)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 4)
	if err := udsB2Activate(t, g, me, surgeon, 0, "", game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{surgeon}, Targets: cardRefs(bear),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !b13Tapped(t, g, surgeon) {
		t.Error("the Surgeon did not tap itself to pay")
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	if got := damageMarkedOn(g, bear); got != 2 {
		t.Errorf("3 damage into a 1-shield: marked %d, want 2", got)
	}
}

// --- Opposition ------------------------------------------------------------------------------

func TestUdsB2OppositionTapsACreatureToTapATarget(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opposition := pushCatalogPermanent(g, me.ID, "Opposition", "Enchantment", udsB2OppositionOracle, true)
	mine := pushCatalogPermanent(g, me.ID, "Fresh Bear", "Creature — Bear", "", true)
	land := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	ench := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	if err := udsB2Activate(t, g, me, opposition, 0, "", game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{mine}, Targets: cardRefs(ench),
	}); err == nil {
		t.Fatal("an enchantment was accepted as the target")
	}
	if err := udsB2Activate(t, g, me, opposition, 0, "", game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{mine}, Targets: cardRefs(land),
	}); err != nil {
		t.Fatalf("activate with a summoning-sick creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !b13Tapped(t, g, land) || !b13Tapped(t, g, mine) {
		t.Error("the land and the paying creature are both tapped")
	}
	if err := udsB2Activate(t, g, me, opposition, 0, "", game.ActivateAbilityParams{
		TapIDs: []uuid.UUID{mine}, Targets: cardRefs(land),
	}); err == nil {
		t.Error("a tapped creature paid the cost again")
	}
}

// --- Serra Advocate ---------------------------------------------------------------------------

func TestUdsB2SerraAdvocatePumpsOnlyACombatant(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advocate := b12Push(g, me.ID, "Serra Advocate", "Creature — Angel", udsB2SerraAdvocateOracle, 2, 2)
	idle := b12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	raider := b12Creature(g, me.ID, "Raider", "Creature — Bear", 2, 2)
	if err := udsB2Activate(t, g, me, advocate, 0, "", game.ActivateAbilityParams{Targets: cardRefs(idle)}); err == nil {
		t.Fatal("a creature out of combat was accepted")
	}
	arnSetCombat(g, raider, opp.ID, uuid.Nil)
	if err := udsB2Activate(t, g, me, advocate, 0, "", game.ActivateAbilityParams{Targets: cardRefs(raider)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, raider); p != 4 || tough != 4 {
		t.Errorf("the attacker is %d/%d, want 4/4", p, tough)
	}
}

// --- Elvish Piper --------------------------------------------------------------------------------

func TestUdsB2ElvishPiperPutsACreatureFromHand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	piper := b12Push(g, me.ID, "Elvish Piper", "Creature — Elf Shaman", udsB2ElvishPiperOracle, 1, 1)
	fatty := handCardFull(me, "Big Fatty", "Creature — Wurm", "{8}", "", nil)
	shock := handCardFull(me, "Shock", "Instant", "{R}", "", nil)
	if err := udsB2Activate(t, g, me, piper, 0, "G", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("the put-from-hand prompt did not open")
	}
	if !hasID(c.ChooseCards, fatty) || hasID(c.ChooseCards, shock) {
		t.Errorf("offered %v: want the creature card and not the instant", c.ChooseCards)
	}
	if c.ChooseMin != 0 {
		t.Error("\"you may put\": declining is a legal answer")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{fatty}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Battlefield.Contains(fatty) || me.Hand.Contains(fatty) {
		t.Error("the creature did not arrive on the battlefield")
	}
}

// --- Rofellos ---------------------------------------------------------------------------------------

func TestUdsB2RofellosAddsGreenPerForest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rofellos := pushCatalogPermanent(g, me.ID, "Rofellos, Llanowar Emissary", "Legendary Creature — Elf Druid", udsB2RofellosOracle, false)
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	if err := g.ActivateManaAbility(me.ID, rofellos, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	got := poolColors(me)
	if len(got) != 2 || got[0] != "G" || got[1] != "G" {
		t.Errorf("pool %v, want G G (two Forests you control)", got)
	}
}

// --- spells --------------------------------------------------------------------------------------------

func TestUdsB2SoulFeastDrainsFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	myLife, oppLife := me.Life, opp.Life
	castCatalogSpell(t, g, "Soul Feast", "Sorcery", udsB2SoulFeastOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-4 || me.Life != myLife+4 {
		t.Errorf("opponent %d -> %d, me %d -> %d: want -4 and +4", oppLife, opp.Life, myLife, me.Life)
	}
}

func TestUdsB2SolidarityGivesYourCreaturesPlusZeroFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Solidarity", "Instant", udsB2SolidarityOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 2 || tough != 7 {
		t.Errorf("your creature is %d/%d, want 2/7", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestUdsB2FlameJetHitsAPlayerAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Flame Jet", "Sorcery", udsB2FlameJetOracle, cardTarget(bear)); err == nil {
		t.Fatal("Flame Jet targeted a creature")
	}
	g = newCatalogGame(t)
	opp = g.Seats[1]
	life := opp.Life
	castCatalogSpell(t, g, "Flame Jet", "Sorcery", udsB2FlameJetOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3", life, opp.Life)
	}
	g = newCatalogGame(t)
	seedLibrary(g.Seats[g.Turn.ActiveSeat], "A", "B")
	ons1Cycle(t, g, "Flame Jet", "Sorcery", udsB2FlameJetOracle, "{C}{C}")
}

func TestUdsB2RecklessAbandonSacrificesACreatureToDealFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	toMainForCost(t, g)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
	life := opp.Life
	cast := func(sac uuid.UUID) error {
		id := handCardFull(me, "Reckless Abandon", "Sorcery", "{R}", udsB2RecklessAbandonOracle, []string{"R"})
		floatForTest(g, me, "R")
		return g.CastSpell(me.ID, id, game.CastSpellParams{
			Targets: []game.TargetRef{playerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{sac}, Strict: true,
		})
	}
	if err := cast(rock); err == nil {
		t.Fatal("an artifact paid the creature sacrifice")
	}
	me.ManaPool.EmptyPool()
	if err := cast(fodder); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("opponent life %d -> %d, want -4", life, opp.Life)
	}
}

func TestUdsB2MultanisDecreeDestroysEnchantmentsForTwoLifeEach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := b12Permanent(g, me.ID, "My Enchantment", "Enchantment")
	theirs := b12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	life := me.Life
	castCatalogSpell(t, g, "Multani's Decree", "Sorcery", udsB2MultanisDecreeOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("an enchantment survived")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("an artifact was destroyed")
	}
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 (two enchantments)", life, me.Life)
	}
}

// --- dies triggers ----------------------------------------------------------------------------------------

func TestUdsB2DiesTriggersChangeATargetCreature(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		wantP, wantT           int
	}{
		{"Disease Carriers", "Creature — Rat", udsB2DiseaseCarriersOracle, 2, 2, 1, 3},
		{"Marker Beetles", "Creature — Insect", udsB2MarkerBeetlesOracle, 2, 3, 4, 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := jmpB7Seats(g)
			dier := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			target := b12Creature(g, opp.ID, "Target", "Creature — Giant", 3, 5)
			advanceToMain(t, g)
			killCreature(t, g, me.ID, dier)
			assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: target}, true)
			if p, tough := ptOf(t, g, target); p != tc.wantP || tough != tc.wantT {
				t.Errorf("the target is %d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestUdsB2DiesTriggersDeclareTheirDirection(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		want         game.TargetEffect
	}{
		{"Disease Carriers", udsB2DiseaseCarriersOracle, game.TargetHarm},
		{"Marker Beetles", udsB2MarkerBeetlesOracle, game.TargetBenefit},
	} {
		spec, _ := Lookup(tc.oracle)
		p := spec.Triggered[0].Purpose
		if p.Targets == nil || len(*p.Targets) != 1 || (*p.Targets)[0].UntilEndOfTurn != tc.want {
			t.Errorf("%s declares %+v, want one %v target", tc.name, p.Targets, tc.want)
		}
	}
}

func TestUdsB2PlagueDogsShrinkEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	dogs := b12Push(g, me.ID, "Plague Dogs", "Creature — Phyrexian Zombie Dog", udsB2PlagueDogsOracle, 3, 3)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	elf := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	advanceToMain(t, g)
	killCreature(t, g, me.ID, dogs)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 1 || tough != 1 {
		t.Errorf("your Bear is %d/%d, want 1/1", p, tough)
	}
	if g.Battlefield.Contains(elf) {
		t.Error("a 1/1 survived at 0/0 (CR 704.5f)")
	}
}

// Sacrificing the creature to draw is a death: the dies trigger fires.
func TestUdsB2SacrificeToDrawCreaturesAlsoTriggerOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B")
	dogs := pushCatalogPermanent(g, me.ID, "Plague Dogs", "Creature — Phyrexian Zombie Dog", udsB2PlagueDogsOracle, true)
	elf := b12Creature(g, me.ID, "My Elf", "Creature — Elf", 1, 1)
	hand := me.Hand.Size()
	if err := udsB2Activate(t, g, me, dogs, 0, "C", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated for {1}; it costs {2}")
	}
	if err := udsB2Activate(t, g, me, dogs, 0, "C", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if g.Battlefield.Contains(elf) {
		t.Error("the dies trigger did not shrink the 1/1")
	}
}

func TestUdsB2GoblinMasonsDestroyOnlyAWall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := jmpB7Seats(g)
	masons := b12Push(g, me.ID, "Goblin Masons", "Creature — Goblin", udsB2GoblinMasonsOracle, 2, 1)
	wall := b12Creature(g, opp.ID, "Wall of Stone", "Creature — Wall", 0, 8)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	killCreature(t, g, me.ID, masons)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the dies trigger")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Error("a non-Wall was accepted as the target")
	}
	assistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: wall}, true)
	if g.Battlefield.Contains(wall) {
		t.Error("the Wall was not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear was destroyed")
	}
}

// --- Thieving Magpie, Telepathic Spies ---------------------------------------------------------------------

func TestUdsB2ThievingMagpieDrawsOnDamageToAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B")
	magpie := b12Push(g, me.ID, "Thieving Magpie", "Creature — Bird", udsB2ThievingMagpieOracle, 1, 3)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(magpie, opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(magpie, bear, 1) })
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("damage to a creature drew a card: %d drawn in all", got)
	}
}

func TestUdsB2TelepathicSpiesLookAtAnOpponentsHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	handCardFull(opp, "Secret", "Instant", "{U}", "", nil)
	castCatalogSpell(t, g, "Telepathic Spies", "Creature — Human Wizard", udsB2TelepathicSpiesOracle, nil)
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	for i := range opp.Hand.Cards {
		if !opp.Hand.Cards[i].IsKnownTo(me.ID) {
			t.Error("the Spies' controller did not see the opponent's hand")
		}
		if opp.Hand.Cards[i].IsKnownTo(other.ID) {
			t.Error("\"look at\" showed the hand to another player")
		}
	}
}

// --- Auras ------------------------------------------------------------------------------------------------------

func TestUdsB2MaskOfLawAndGraceGivesProtection(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	auraCast(t, g, "Mask of Law and Grace", udsB2MaskOfLawAndGraceOracle, bear)
	assertKeywords(t, g, bear, "protection from black", "protection from red")
	if !odyB2BoltRefused(t, g, bear) {
		t.Error("a red Lightning Bolt targeted the enchanted creature")
	}
}

func TestUdsB2AurasSacrificeThemselvesToDraw(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int
		keyword      string
	}{
		{"Capashen Standard", udsB2CapashenStandardOracle, 3, 3, ""},
		{"Illuminated Wings", udsB2IlluminatedWingsOracle, 2, 2, "flying"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			seedLibrary(me, "A", "B")
			bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			aura := auraCast(t, g, tc.name, tc.oracle, bear)
			if p, tough := ptOf(t, g, bear); p != tc.power || tough != tc.tough {
				t.Errorf("the enchanted Bear is %d/%d, want %d/%d", p, tough, tc.power, tc.tough)
			}
			if tc.keyword != "" && !hasKeywordOnBattlefield(t, g, bear, tc.keyword) {
				t.Errorf("the enchanted Bear has no %s", tc.keyword)
			}
			hand := me.Hand.Size()
			if err := udsB2Activate(t, g, me, aura, 0, "CC", game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if g.Battlefield.Contains(aura) {
				t.Error("the Aura was not sacrificed as the cost")
			}
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != 1 {
				t.Errorf("drew %d, want 1", got)
			}
			if p, tough := ptOf(t, g, bear); p != 2 || tough != 2 {
				t.Errorf("the Bear is %d/%d after the Aura left, want 2/2", p, tough)
			}
		})
	}
}
