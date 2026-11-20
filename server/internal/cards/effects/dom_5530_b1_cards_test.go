package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dom_5530_b1_cards_test.go — Dominaria, batch 1 (#5530). Every test plays
// the real catalog entry through the engine.

const (
	domAvenSentryOracle          = "36a600fe-91b4-46a0-97ee-0cea2b63ec8e"
	domColdWaterSnapperOracle    = "457e4a49-c8fe-4e1d-a4b4-7a880a47f58d"
	domFeralAbominationOracle    = "1bf06342-0d76-433b-9912-8af03b7e3bc3"
	domMammothSpiderOracle       = "df77cb98-e3c6-4927-99d5-c3380641f356"
	domPardicWandererOracle      = "82ff4500-b118-4ce7-9f7c-92661328b242"
	domPyromanticPilgrimOracle   = "24d52cf1-d4f1-4b4c-a50b-87d3ba0aeebb"
	domCabalEvangelOracle        = "a4fed66f-ef3c-4547-8612-46e65d63ef1f"
	domKnightOfNewBenaliaOracle  = "75f28e54-17ce-49f9-85b2-9f9a04e1d86f"
	domPrimordialWurmOracle      = "ab9778da-f5e3-4c17-a3da-ffef6ed7a7a1"
	domTolarianScholarOracle     = "8a268636-75d5-4631-8272-8b001f569db5"
	domBlessedLightOracle        = "918b3860-e96c-45b5-b0e6-e82cad9f304d"
	domChargeOracle              = "36b5022d-fa62-46ea-97ab-23c0f60f62a5"
	domCloudreaderSphinxOracle   = "b5a4bf91-4d31-4557-9ec4-a96ff4d9be6b"
	domDeathbloomThallidOracle   = "ce873790-de4d-4ba3-8023-c44306ba3ff8"
	domDreadShadeOracle          = "9b90c7d1-38c6-4210-bc76-68c5fcb91202"
	domEviscerateOracle          = "c3a89c67-1931-4d42-98c8-94a40cfd046e"
	domFrenziedRageOracle        = "1a04c31c-f5e4-43ff-a5b5-207426b8dca6"
	domLlanowarScoutOracle       = "3bdc9211-8eff-4b30-8d6f-d26f4087f77f"
	domOrcishVandalOracle        = "933a1811-7e1a-4173-b1c9-34b2c858ba68"
	domPegasusCourserOracle      = "24db5905-d513-42c6-9e9d-114ac5cbd21d"
	domSealAwayOracle            = "c8909015-ea49-47cb-8d37-653904f965bb"
	domShortSwordOracle          = "e842a4e5-ea0b-43ae-9ed8-93e30f4e7b8a"
	domSoulSalvageOracle         = "60b6bd93-3eb3-4b81-bc4d-1e3351ee2e64"
	domWarlordsFuryOracle        = "7340199d-60c7-4c27-9a0f-047790273db0"
	domAcademyDrakeOracle        = "65f3ebeb-1ca5-4923-b681-032579991c8a"
	domUntamedKavuOracle         = "ce585b28-f22f-41ba-8858-aebc0164922d"
	domArcaneFlightOracle        = "40eee955-bec6-455a-addf-68cdb84c76b9"
	domBefuddleOracle            = "664c1219-acfa-41da-b811-503a63fc8d84"
	domBenalishMarshalOracle     = "2cc439e8-d112-44e6-bc5a-6e99333c519a"
	domCastDownOracle            = "cdaab6b0-1a2d-4809-8e6b-56013acd8f78"
	domFerventStrikeOracle       = "35efb1e6-6be2-4a38-969a-199524d48e02"
	domFieryInterventionOracle   = "a8cd3a52-1056-4664-ae55-eb0ba3a43a15"
	domHomaridExplorerOracle     = "ac30745c-addf-433a-be73-d3d9f4cc4328"
	domLlanowarEnvoyOracle       = "6f096e98-d351-4050-90ed-4c6f12e129d6"
	domMemorialToWarOracle       = "f98db69c-b330-4560-ac53-10857674466b"
	domShivanFireOracle          = "854fd120-9a51-4d37-9922-7e4b0464e0f5"
	domSparringConstructOracle   = "bc948e19-816b-4c13-a36b-73b3e0472dca"
	domViciousOfferingOracle     = "c6adfaa7-d9cf-4501-8e7d-e035369390d2"
	domWindgraceAcolyteOracle    = "334a98c0-9a16-4145-8f5b-bdf017afe73f"
	domWizardsLightningOracle    = "2bd81743-84c8-43ec-86f5-68230036827b"
	domYavimayaSapherdOracle     = "42a1954f-774e-4d80-98b3-a736f23e2060"
	domCabalPaladinOracle        = "1104a746-169c-450a-80ad-bea0493c0576"
	domSerraDiscipleOracle       = "0f052d9f-5a40-4eb9-9d8a-e01dec1db191"
	domArtificersAssistantOracle = "6191a7eb-04b5-47e6-a31f-43f9aa91d4d7"
	domExcavationElephantOracle  = "45fa04ee-3b5c-45b3-8c09-bda08cca3630"
	domGhituChroniclerOracle     = "ab8a00c2-dbb0-48c3-9d13-ad6f1e2f3a18"
	domKrosanDruidOracle         = "1fbed9e4-86f6-4115-8468-8593e340701f"
	domSergeantAtArmsOracle      = "2ea95e03-b95d-4ccf-afc6-1a26ac6cf354"
	domCaligoSkinWitchOracle     = "e62c943e-03ac-4a8a-ad3d-faa76b79824c"
	domCallTheCavalryOracle      = "855b168b-e514-4a43-8bbe-7f1df40031ae"

	domBearType = "Creature — Bear"
	domAuraType = "Enchantment — Aura"
)

// domCastHistoric casts an uncatalogued artifact spell, a historic spell
// (CR 700.6), from the active player's hand and lets it resolve.
func domCastHistoric(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Test Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
}

// --- every card is in the catalog -------------------------------------------

func TestDom5530B1EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		domAvenSentryOracle: "Aven Sentry", domColdWaterSnapperOracle: "Cold-Water Snapper",
		domFeralAbominationOracle: "Feral Abomination", domMammothSpiderOracle: "Mammoth Spider",
		domPardicWandererOracle: "Pardic Wanderer", domPyromanticPilgrimOracle: "Pyromantic Pilgrim",
		domCabalEvangelOracle: "Cabal Evangel", domKnightOfNewBenaliaOracle: "Knight of New Benalia",
		domPrimordialWurmOracle: "Primordial Wurm", domTolarianScholarOracle: "Tolarian Scholar",
		domBlessedLightOracle: "Blessed Light", domChargeOracle: "Charge",
		domCloudreaderSphinxOracle: "Cloudreader Sphinx", domDeathbloomThallidOracle: "Deathbloom Thallid",
		domDreadShadeOracle: "Dread Shade", domEviscerateOracle: "Eviscerate",
		domFrenziedRageOracle: "Frenzied Rage", domLlanowarScoutOracle: "Llanowar Scout",
		domOrcishVandalOracle: "Orcish Vandal", domPegasusCourserOracle: "Pegasus Courser",
		domSealAwayOracle: "Seal Away", domShortSwordOracle: "Short Sword",
		domSoulSalvageOracle: "Soul Salvage", domWarlordsFuryOracle: "Warlord's Fury",
		domAcademyDrakeOracle: "Academy Drake", domUntamedKavuOracle: "Untamed Kavu",
		domArcaneFlightOracle: "Arcane Flight", domBefuddleOracle: "Befuddle",
		domBenalishMarshalOracle: "Benalish Marshal", domCastDownOracle: "Cast Down",
		domFerventStrikeOracle: "Fervent Strike", domFieryInterventionOracle: "Fiery Intervention",
		domHomaridExplorerOracle: "Homarid Explorer", domLlanowarEnvoyOracle: "Llanowar Envoy",
		domMemorialToWarOracle: "Memorial to War", domShivanFireOracle: "Shivan Fire",
		domSparringConstructOracle: "Sparring Construct", domViciousOfferingOracle: "Vicious Offering",
		domWindgraceAcolyteOracle: "Windgrace Acolyte", domWizardsLightningOracle: "Wizard's Lightning",
		domYavimayaSapherdOracle: "Yavimaya Sapherd", domCabalPaladinOracle: "Cabal Paladin",
		domSerraDiscipleOracle: "Serra Disciple", domArtificersAssistantOracle: "Artificer's Assistant",
		domExcavationElephantOracle: "Excavation Elephant", domGhituChroniclerOracle: "Ghitu Chronicler",
		domKrosanDruidOracle: "Krosan Druid", domSergeantAtArmsOracle: "Sergeant-at-Arms",
		domCaligoSkinWitchOracle: "Caligo Skin-Witch", domCallTheCavalryOracle: "Call the Cavalry",
	}
	if len(want) != 50 {
		t.Fatalf("the batch registers 50 cards, the table lists %d", len(want))
	}
	for oracle, name := range want {
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

// --- printed keywords and vanillas -------------------------------------------

func TestDom5530B1PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Aven Sentry", "Creature — Bird Soldier", domAvenSentryOracle, []string{"flying"}},
		{"Cold-Water Snapper", "Creature — Turtle", domColdWaterSnapperOracle, []string{"hexproof"}},
		{"Feral Abomination", "Creature — Thrull", domFeralAbominationOracle, []string{"deathtouch"}},
		{"Mammoth Spider", "Creature — Spider", domMammothSpiderOracle, []string{"reach"}},
		{"Pardic Wanderer", "Artifact Creature — Golem", domPardicWandererOracle, []string{"trample"}},
		{"Pyromantic Pilgrim", "Creature — Human Wizard", domPyromanticPilgrimOracle, []string{"haste"}},
		{"Cloudreader Sphinx", "Creature — Sphinx", domCloudreaderSphinxOracle, []string{"flying"}},
		{"Pegasus Courser", "Creature — Pegasus", domPegasusCourserOracle, []string{"flying"}},
		{"Academy Drake", "Creature — Drake", domAcademyDrakeOracle, []string{"flying"}},
		{"Untamed Kavu", "Creature — Kavu", domUntamedKavuOracle, []string{"vigilance", "trample"}},
		{"Windgrace Acolyte", "Creature — Cat Warrior", domWindgraceAcolyteOracle, []string{"flying"}},
		{"Serra Disciple", "Creature — Bird Cleric", domSerraDiscipleOracle, []string{"flying", "first strike"}},
		{"Artificer's Assistant", "Creature — Bird", domArtificersAssistantOracle, []string{"flying"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			for _, kw := range tc.keywords {
				if !hasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

func TestDom5530B1VanillasResolve(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Cabal Evangel", "Creature — Human Cleric", domCabalEvangelOracle},
		{"Knight of New Benalia", "Creature — Human Knight", domKnightOfNewBenaliaOracle},
		{"Primordial Wurm", "Creature — Wurm", domPrimordialWurmOracle},
		{"Tolarian Scholar", "Creature — Human Wizard", domTolarianScholarOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := castHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			if !g.Battlefield.Contains(id) {
				t.Errorf("%s did not resolve", tc.name)
			}
		})
	}
}

// Cold-Water Snapper's hexproof: an opponent's spell can't target it.
func TestDom5530B1ColdWaterSnapperCantBeTargetedByOpponents(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	snapper := b12Push(g, opp.ID, "Cold-Water Snapper", "Creature — Turtle", domColdWaterSnapperOracle, 4, 5)
	if err := castCatalogSpellErr(t, g, "Eviscerate", "Sorcery", domEviscerateOracle, cardTarget(snapper)); err == nil {
		t.Error("an opponent's spell targeted a creature with hexproof")
	}
}

// --- removal and burn -----------------------------------------------------------

func TestDom5530B1BlessedLightExilesACreatureOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	rune := b12Permanent(g, opp.ID, "Rune", "Enchantment")
	if err := castCatalogSpellErr(t, g, "Blessed Light", "Instant", domBlessedLightOracle, cardTarget(rock)); err == nil {
		t.Fatal("Blessed Light targeted an artifact")
	}
	castCatalogSpell(t, g, "Blessed Light", "Instant", domBlessedLightOracle, cardTarget(rune))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rune) || !g.Exile.Contains(rune) {
		t.Error("the enchantment was not exiled")
	}
}

func TestDom5530B1EviscerateDestroysACreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	bear := b12Creature(g, opp.ID, "Bear", domBearType, 2, 2)
	castCatalogSpell(t, g, "Eviscerate", "Sorcery", domEviscerateOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature was not destroyed")
	}
}

func TestDom5530B1CastDownSparesLegends(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	legend := b12Creature(g, opp.ID, "Legend", "Legendary Creature — Human", 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", domBearType, 2, 2)
	if err := castCatalogSpellErr(t, g, "Cast Down", "Instant", domCastDownOracle, cardTarget(legend)); err == nil {
		t.Fatal("Cast Down targeted a legendary creature")
	}
	castCatalogSpell(t, g, "Cast Down", "Instant", domCastDownOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the nonlegendary creature survived")
	}
}

func TestDom5530B1FieryInterventionEachMode(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 6, 6)
	castModal(t, g, "Fiery Intervention", "Sorcery", domFieryInterventionOracle, []int{0}, cardTarget(ogre))
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, ogre).DamageMarked; got != 5 {
		t.Errorf("the creature has %d damage, want 5", got)
	}
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	castModal(t, g, "Fiery Intervention", "Sorcery", domFieryInterventionOracle, []int{1}, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
}

func TestDom5530B1ShivanFireDealsTwoOrFourKicked(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		want     int
	}{
		{"unkicked", nil, 2},
		{"kicked", []int{0}, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 6, 6)
			if _, err := castWithOptionalCosts(t, g, "Shivan Fire", "Instant", domShivanFireOracle, cardTarget(ogre), tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			if got := findCardAnywhere(t, g, ogre).DamageMarked; got != tc.want {
				t.Errorf("the creature has %d damage, want %d", got, tc.want)
			}
		})
	}
}

func TestDom5530B1WizardsLightningCostsLessWithAWizard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, domWizardsLightningOracle, "Instant", "{2}{R}"); got != 3 {
		t.Errorf("without a Wizard: %d, want 3", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Test Wizard", TypeLine: "Creature — Human Wizard", Power: 1, Toughness: 1})
	if got := selfPricedMV(t, g, me, domWizardsLightningOracle, "Instant", "{2}{R}"); got != 1 {
		t.Errorf("with a Wizard: %d, want 1", got)
	}
}

func TestDom5530B1WizardsLightningDealsThree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	life := opp.Life
	castCatalogSpell(t, g, "Wizard's Lightning", "Instant", domWizardsLightningOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
}

func TestDom5530B1ViciousOfferingShrinksMoreWhenKicked(t *testing.T) {
	for _, tc := range []struct {
		name    string
		kicked  bool
		survive bool
	}{
		{"unkicked", false, true},
		{"kicked", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
			var optional []int
			var sacrifices []uuid.UUID
			victim := b12Creature(g, me.ID, "Victim", domBearType, 1, 1)
			if tc.kicked {
				optional, sacrifices = []int{0}, []uuid.UUID{victim}
			}
			if _, err := castWithOptionalCosts(t, g, "Vicious Offering", "Instant", domViciousOfferingOracle, cardTarget(ogre), optional, sacrifices); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			g.RunStateChecksForTest()
			if got := g.Battlefield.Contains(ogre); got != tc.survive {
				t.Fatalf("the 5/5 on the battlefield = %v, want %v", got, tc.survive)
			}
			if tc.survive {
				if p, tough := effectivePTOf(t, g, ogre); p != 3 || tough != 3 {
					t.Errorf("the creature is %d/%d, want 3/3", p, tough)
				}
			}
			if got := g.Battlefield.Contains(victim); got == tc.kicked {
				t.Errorf("the would-be sacrifice on the battlefield = %v", got)
			}
		})
	}
}

// --- pumps and team effects ---------------------------------------------------

func TestDom5530B1ChargePumpsOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine := b12Creature(g, me.ID, "Mine", domBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", domBearType, 2, 2)
	castCatalogSpell(t, g, "Charge", "Instant", domChargeOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestDom5530B1WarlordsFuryGrantsFirstStrikeAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine := b12Creature(g, me.ID, "Mine", domBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", domBearType, 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Warlord's Fury", "Sorcery", domWarlordsFuryOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, mine, "first strike") {
		t.Error("your creature has no first strike")
	}
	if hasEffectiveKeyword(t, g, theirs, "first strike") {
		t.Error("an opponent's creature has first strike")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestDom5530B1BefuddleShrinksAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Befuddle", "Instant", domBefuddleOracle, cardTarget(ogre))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, ogre); p != 1 || tough != 5 {
		t.Errorf("the creature is %d/%d, want 1/5", p, tough)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand is %d, want %d", got, hand+1)
	}
}

func TestDom5530B1FerventStrikePumpsWithFirstStrikeAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bear := b12Creature(g, me.ID, "Bear", domBearType, 2, 2)
	castCatalogSpell(t, g, "Fervent Strike", "Instant", domFerventStrikeOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 3/2", p, tough)
	}
	for _, kw := range []string{"first strike", "haste"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the Bear has no %s", kw)
		}
	}
}

func TestDom5530B1BenalishMarshalPumpsOtherCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marshal := b12Push(g, me.ID, "Benalish Marshal", "Creature — Human Knight", domBenalishMarshalOracle, 3, 3)
	mine := b12Creature(g, me.ID, "Mine", domBearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", domBearType, 2, 2)
	if p, tough := effectivePTOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your other creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, marshal); p != 3 || tough != 3 {
		t.Errorf("the Marshal is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestDom5530B1DreadShadePumpsItself(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	shade := b12Push(g, me.ID, "Dread Shade", "Creature — Shade", domDreadShadeOracle, 3, 3)
	apaMana(me, "B")
	apaActivate(t, g, me, shade, 0, game.ActivateAbilityParams{})
	if p, tough := effectivePTOf(t, g, shade); p != 4 || tough != 4 {
		t.Errorf("the Shade is %d/%d, want 4/4", p, tough)
	}
}

// --- Auras and Equipment ------------------------------------------------------

func TestDom5530B1Auras(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, keyword string
		p, tough              int
	}{
		{"Frenzied Rage", domFrenziedRageOracle, "menace", 4, 3},
		{"Arcane Flight", domArcaneFlightOracle, "flying", 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			bear := b12Creature(g, me.ID, "Bear", domBearType, 2, 2)
			castCatalogSpell(t, g, tc.name, domAuraType, tc.oracle, cardTarget(bear))
			passPriorityAroundTable(t, g)
			if p, tough := effectivePTOf(t, g, bear); p != tc.p || tough != tc.tough {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
			}
			if !hasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear has no %s", tc.keyword)
			}
		})
	}
}

func TestDom5530B1ShortSwordEquipsForOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := seedBear(g, me.ID)
	sword := pushCatalogPermanent(g, me.ID, "Short Sword", "Artifact — Equipment", domShortSwordOracle, false)
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	equipTo(t, g, me.ID, sword, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the equipped Bear is %d/%d, want 3/3", p, tough)
	}
}

// --- enters, dies and attack triggers -----------------------------------------

func TestDom5530B1CloudreaderSphinxScriesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Cloudreader Sphinx", "Creature — Sphinx", domCloudreaderSphinxOracle, 3, 4)
	passPriorityAroundTable(t, g)
	scry := latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID)
	if scry == nil {
		t.Fatal("no scry prompt")
	}
	if len(scry.ScryCards) != 2 {
		t.Errorf("scrying %d cards, want 2", len(scry.ScryCards))
	}
}

func TestDom5530B1SaprolingMakers(t *testing.T) {
	t.Run("Yavimaya Sapherd enters", func(t *testing.T) {
		g := newCatalogGame(t)
		castHoldCreature(t, g, "Yavimaya Sapherd", "Creature — Fungus", domYavimayaSapherdOracle, 2, 2)
		passPriorityAroundTable(t, g)
		if n := len(tokensNamed(g, "Saproling")); n != 1 {
			t.Errorf("%d Saprolings, want 1", n)
		}
	})
	t.Run("Deathbloom Thallid dies", func(t *testing.T) {
		g := newCatalogGame(t)
		thallid := b12Push(g, g.Seats[0].ID, "Deathbloom Thallid", "Creature — Fungus", domDeathbloomThallidOracle, 3, 2)
		b27Kill(g, thallid)
		g.RunStateChecksForTest()
		passPriorityAroundTable(t, g)
		if n := len(tokensNamed(g, "Saproling")); n != 1 {
			t.Errorf("%d Saprolings, want 1", n)
		}
	})
}

func TestDom5530B1HomaridExplorerMillsTargetPlayerFour(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	library := opp.Library.Size()
	castHoldCreature(t, g, "Homarid Explorer", "Creature — Homarid Scout", domHomaridExplorerOracle, 3, 3)
	b04WaitForPick(t, g, g.Seats[g.Turn.ActiveSeat].ID)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := library - opp.Library.Size(); got != 4 {
		t.Errorf("milled %d, want 4", got)
	}
}

func TestDom5530B1WindgraceAcolyteMillsThreeAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	library, life := me.Library.Size(), me.Life
	castHoldCreature(t, g, "Windgrace Acolyte", "Creature — Cat Warrior", domWindgraceAcolyteOracle, 3, 2)
	passPriorityAroundTable(t, g)
	if got := library - me.Library.Size(); got != 3 {
		t.Errorf("milled %d, want 3", got)
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestDom5530B1SparringConstructCountersACreatureWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	construct := b12Push(g, me.ID, "Sparring Construct", "Artifact Creature — Construct", domSparringConstructOracle, 1, 1)
	bear := b12Creature(g, me.ID, "Bear", domBearType, 2, 2)
	b27Kill(g, construct)
	g.RunStateChecksForTest()
	b04WaitForPick(t, g, me.ID)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if n := plusOneCounters(g, bear); n != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", n)
	}
}

func TestDom5530B1PegasusCourserGivesAnotherAttackerFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	courser := pushCatalogPermanent(g, me.ID, "Pegasus Courser", "Creature — Pegasus", domPegasusCourserOracle, false)
	bear := seedBear(g, me.ID)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	declareAttack(t, g, opp.ID, courser, bear)
	answerTriggerTargets(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the other attacker did not gain flying")
	}
}

func TestDom5530B1SealAwayExilesATappedOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	tapped := b12Creature(g, opp.ID, "Tapped Bear", domBearType, 2, 2)
	untapped := b12Creature(g, opp.ID, "Untapped Bear", domBearType, 2, 2)
	tapForTest(t, g, tapped)
	castCatalogSpell(t, g, "Seal Away", "Enchantment", domSealAwayOracle, nil)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, tapped) || hasID(p.PickTargetCards, untapped) {
		t.Fatalf("targets must be tapped creatures an opponent controls: %v", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, tapped)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) || !g.Exile.Contains(tapped) {
		t.Error("the tapped creature was not exiled")
	}
}

// --- historic cast triggers (CR 700.6) ------------------------------------------

func TestDom5530B1CabalPaladinDamagesEachOpponentOnAHistoricSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Cabal Paladin", "Creature — Human Knight", domCabalPaladinOracle, 4, 2)
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	domCastHistoric(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID] - 2
		if s.ID == me.ID {
			want = lives[s.ID]
		}
		if s.Life != want {
			t.Errorf("%s is at %d, want %d", s.Name, s.Life, want)
		}
	}
}

func TestDom5530B1SerraDiscipleGrowsOnAHistoricSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	disciple := b12Push(g, me.ID, "Serra Disciple", "Creature — Bird Cleric", domSerraDiscipleOracle, 1, 1)
	domCastHistoric(t, g)
	if p, tough := effectivePTOf(t, g, disciple); p != 2 || tough != 2 {
		t.Errorf("the Disciple is %d/%d, want 2/2", p, tough)
	}
}

func TestDom5530B1ArtificersAssistantScriesOnAHistoricSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Artificer's Assistant", "Creature — Bird", domArtificersAssistantOracle, 1, 1)
	domCastHistoric(t, g)
	scry := latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID)
	if scry == nil {
		t.Fatal("no scry prompt")
	}
	if len(scry.ScryCards) != 1 {
		t.Errorf("scrying %d cards, want 1", len(scry.ScryCards))
	}
}

// --- kicker ---------------------------------------------------------------------

func TestDom5530B1KickedCountersOnEntry(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		want                   int
	}{
		{"Academy Drake", "Creature — Drake", domAcademyDrakeOracle, 2},
		{"Untamed Kavu", "Creature — Kavu", domUntamedKavuOracle, 3},
	} {
		for _, kicked := range []bool{false, true} {
			t.Run(tc.name, func(t *testing.T) {
				g := newCatalogGame(t)
				var optional []int
				want := 0
				if kicked {
					optional, want = []int{0}, tc.want
				}
				id, err := castWithOptionalCosts(t, g, tc.name, tc.typeLine, tc.oracle, nil, optional, nil)
				if err != nil {
					t.Fatalf("CastSpell: %v", err)
				}
				passPriorityAroundTable(t, g)
				if n := plusOneCounters(g, id); n != want {
					t.Errorf("kicked=%v: %d +1/+1 counters, want %d", kicked, n, want)
				}
			})
		}
	}
}

func TestDom5530B1KrosanDruidGainsTenOnlyWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		life := me.Life
		var optional []int
		want := life
		if kicked {
			optional, want = []int{0}, life+10
		}
		if _, err := castWithOptionalCosts(t, g, "Krosan Druid", "Creature — Centaur Druid", domKrosanDruidOracle, nil, optional, nil); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		passPriorityAroundTable(t, g)
		if me.Life != want {
			t.Errorf("kicked=%v: life %d, want %d", kicked, me.Life, want)
		}
	}
}

func TestDom5530B1SergeantAtArmsMakesSoldiersOnlyWhenKicked(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		var optional []int
		want := 0
		if kicked {
			optional, want = []int{0}, 2
		}
		if _, err := castWithOptionalCosts(t, g, "Sergeant-at-Arms", "Creature — Human Soldier", domSergeantAtArmsOracle, nil, optional, nil); err != nil {
			t.Fatalf("CastSpell: %v", err)
		}
		passPriorityAroundTable(t, g)
		if n := len(tokensNamed(g, "Soldier")); n != want {
			t.Errorf("kicked=%v: %d Soldiers, want %d", kicked, n, want)
		}
	}
}

func TestDom5530B1CaligoSkinWitchMakesOpponentsDiscardWhenKicked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if _, err := castWithOptionalCosts(t, g, "Caligo Skin-Witch", "Creature — Human Wizard", domCaligoSkinWitchOracle, nil, []int{0}, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, opp.ID); n != 2 {
		t.Errorf("the opponent owes %d discards, want 2", n)
	}
	if n := discardOwed(g, me.ID); n != 0 {
		t.Errorf("the caster owes %d discards", n)
	}
}

func TestDom5530B1CaligoSkinWitchUnkickedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if _, err := castWithOptionalCosts(t, g, "Caligo Skin-Witch", "Creature — Human Wizard", domCaligoSkinWitchOracle, nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, opp.ID); n != 0 {
		t.Errorf("the opponent owes %d discards, want 0", n)
	}
}

func TestDom5530B1KickedGraveyardReturns(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		good, bad              string
	}{
		{"Excavation Elephant", "Creature — Elephant", domExcavationElephantOracle, "Artifact", domBearType},
		{"Ghitu Chronicler", "Creature — Human Wizard", domGhituChroniclerOracle, "Instant", "Artifact"},
	} {
		t.Run(tc.name+" kicked", func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			good := wthGraveyardCard(me, "Good", tc.good)
			bad := wthGraveyardCard(me, "Bad", tc.bad)
			if _, err := castWithOptionalCosts(t, g, tc.name, tc.typeLine, tc.oracle, nil, []int{0}, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			b04WaitForPick(t, g, me.ID)
			p := latestPickTarget(g, me.ID)
			if !hasID(p.PickTargetCards, good) || hasID(p.PickTargetCards, bad) {
				t.Fatalf("legal set %v, want only the %s card", p.PickTargetCards, tc.good)
			}
			pickCard(t, g, me.ID, good)
			passPriorityAroundTable(t, g)
			if !me.Hand.Contains(good) {
				t.Error("the card did not return to hand")
			}
		})
		t.Run(tc.name+" unkicked", func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			good := wthGraveyardCard(me, "Good", tc.good)
			if _, err := castWithOptionalCosts(t, g, tc.name, tc.typeLine, tc.oracle, nil, nil, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			if latestPickTarget(g, me.ID) != nil || !me.Graveyard.Contains(good) {
				t.Error("an unkicked creature returned a card")
			}
		})
	}
}

// --- activated and mana abilities -----------------------------------------------

func TestDom5530B1LlanowarScoutPutsALandFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	scout := b12Push(g, me.ID, "Llanowar Scout", "Creature — Elf Scout", domLlanowarScoutOracle, 1, 3)
	land := handCardForTest(me, "Forest", "Basic Land — Forest", "")
	rock := handCardForTest(me, "Sol Ring", "Artifact", "")
	advanceToMain(t, g)
	b16Activate(t, g, me.ID, scout, 0, game.ActivateAbilityParams{Strict: true})
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no land-drop prompt")
	}
	if hasID(pick.ChooseCards, rock) || !hasID(pick.ChooseCards, land) {
		t.Errorf("candidates %v: want only the land", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Error("the land did not reach the battlefield")
	}
}

func TestDom5530B1OrcishVandalSacrificesAnArtifactForTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	vandal := b12Push(g, me.ID, "Orcish Vandal", "Creature — Orc Warrior", domOrcishVandalOracle, 1, 1)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	life := opp.Life
	apaActivate(t, g, me, vandal, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{playerTarget(opp.ID)},
		SacrificeIDs: []uuid.UUID{rock},
	})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
}

func TestDom5530B1LlanowarEnvoyFiltersIntoAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	envoy := b12Push(g, me.ID, "Llanowar Envoy", "Creature — Elf Scout", domLlanowarEnvoyOracle, 3, 2)
	apaMana(me, "C", "G")
	if err := g.ActivateManaAbility(me.ID, envoy, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if n := me.ManaPool.CountColor("U"); n != 1 || me.ManaPool.Total() != 1 {
		t.Errorf("pool holds %d blue of %d, want 1 of 1", n, me.ManaPool.Total())
	}
}

func TestDom5530B1MemorialToWarEntersTappedAndDestroysALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	land := playLandFromHand(t, g, "Memorial to War", domMemorialToWarOracle)
	if !b27Tapped(t, g, land) {
		t.Error("it did not enter tapped")
	}
	b22Untap(g, land)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want R", got)
	}
	me.ManaPool = nil
	b22Untap(g, land)
	forest := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	b06AddMana(me, "C", "C", "C", "C", "R")
	b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{Targets: cardTarget(forest)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) {
		t.Error("the target land was not destroyed")
	}
	if g.Battlefield.Contains(land) {
		t.Error("the Memorial was not sacrificed")
	}
}

// --- the other spells -------------------------------------------------------------

func TestDom5530B1SoulSalvageReturnsTwoCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	one := pushGraveyardCardForTest(me, "One")
	two := pushGraveyardCardForTest(me, "Two")
	rock := wthGraveyardCard(me, "Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Soul Salvage", "Sorcery", domSoulSalvageOracle, cardTarget(rock)); err == nil {
		t.Fatal("Soul Salvage targeted an artifact card")
	}
	castCatalogSpell(t, g, "Soul Salvage", "Sorcery", domSoulSalvageOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: one}, {Kind: game.TargetCard, ID: two},
	})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(one) || !me.Hand.Contains(two) {
		t.Error("the creature cards did not return to hand")
	}
}

func TestDom5530B1CallTheCavalryMakesTwoVigilantKnights(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Call the Cavalry", "Sorcery", domCallTheCavalryOracle, nil)
	passPriorityAroundTable(t, g)
	knights := tokensNamed(g, "Knight")
	if len(knights) != 2 {
		t.Fatalf("%d Knights, want 2", len(knights))
	}
	for _, k := range knights {
		if !hasEffectiveKeyword(t, g, k.InstanceID, "vigilance") {
			t.Error("a Knight has no vigilance")
		}
		if p, tough := effectivePTOf(t, g, k.InstanceID); p != 2 || tough != 2 {
			t.Errorf("a Knight is %d/%d, want 2/2", p, tough)
		}
	}
}
