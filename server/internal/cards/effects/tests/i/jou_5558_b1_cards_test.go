package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jou_5558_b1_cards_test.go — Journey into Nyx, batch 1 (#5558, S82).
// Every test plays the real catalog entry through the engine.

const (
	jouB1OreskosSwiftclawOracle      = "014d00fc-434a-4b06-84b2-afd930677d61"
	jouB1PensiveMinotaurOracle       = "cafad9c8-9974-4596-92d4-b7fb615a0353"
	jouB1RottedHulkOracle            = "44fd00d5-84e4-48ec-bd6a-d53afbc2dd62"
	jouB1BassaraTowerArcherOracle    = "39daa265-8d02-4449-9c45-ae8ec2b13e5b"
	jouB1CloakedSirenOracle          = "d38a33ae-59d1-4d08-8e14-5fa5a8c3762d"
	jouB1EagleOfTheWatchOracle       = "954836f8-c225-450a-bb51-f2adb6923ef5"
	jouB1GoldForgedSentinelOracle    = "c338fb29-2dd2-4f76-ba3d-fa6770adcce8"
	jouB1SkyspearCavalryOracle       = "b5868383-a0fb-40fc-b718-f1b5441c5ce3"
	jouB1GoldenHindOracle            = "82679258-345c-4062-ac62-e94a1e5ace78"
	jouB1AkroanMastiffOracle         = "6039b9b7-e49a-4ea3-887c-a17dd502d63f"
	jouB1CyclopsOfEternalFuryOracle  = "f7875f87-d774-4811-a4a7-349ec01bf95b"
	jouB1EidolonOfRhetoricOracle     = "33017203-233e-498e-806e-94b1dccf5c35"
	jouB1DictateOfHeliodOracle       = "b989279c-665d-4f15-afc1-adf3872a4851"
	jouB1FelhidePetrifierOracle      = "1ad604ed-c645-4f94-8c5f-8e0ce66a85ff"
	jouB1PheresBandWarchiefOracle    = "79bc9b71-5b04-46b8-a9de-9e9f17936f29"
	jouB1EidolonOfTheGreatRevel      = "2fd7018e-1c36-4d7f-8f72-22f2f7f60c60"
	jouB1WarWingSirenOracle          = "0e49ec18-f979-4168-b7df-7728f3fe442e"
	jouB1LagonnaBandOracle           = "09cc0faa-71d7-428b-9fc4-5ba09e244e89"
	jouB1DawnbringerOracle           = "29263d48-cb28-478b-a99e-5a18dbc924f0"
	jouB1ThunderhoofOracle           = "651a1f3d-d6f0-4a38-a776-bceba698127b"
	jouB1AkroanLineBreakerOracle     = "01443d75-3f86-45c0-81fa-fbf1b3175d4f"
	jouB1LeoninIconoclastOracle      = "14d14ded-8831-47ad-ad90-f6f242cccf03"
	jouB1TritonCavalryOracle         = "0fd11de4-2dfd-4345-b3cb-d7d65e11c6d9"
	jouB1AgentOfErebosOracle         = "3fa5e109-4a67-4523-8183-90e48a906400"
	jouB1ForgebornOreadsOracle       = "8289190c-b180-48e3-b92f-7dccbca61b39"
	jouB1ThassasDevourerOracle       = "b9794b16-3ef4-424a-b1ee-e2077f53873d"
	jouB1ThoughtrenderLamiaOracle    = "d97c4bca-6433-4cb0-b63a-fdf8360fc29c"
	jouB1OakheartDryadsOracle        = "60bac6e8-0c72-4665-9b01-d07c99dd1716"
	jouB1DreadbringerLampadsOracle   = "e65fa865-3f60-41f2-8b14-b0333c897b93"
	jouB1WhitewaterNaiadsOracle      = "6c73e533-9ed1-4df0-af12-c0174c0b52b9"
	jouB1HumblerOfMortalsOracle      = "91933f1a-4aad-4f67-8151-a47ab6d0e42c"
	jouB1GoldenhideOxOracle          = "2934233a-b90c-45df-b604-ce462480bbb7"
	jouB1SatyrGrovedancerOracle      = "49f3bd5a-9c2a-45b8-aa10-9c6471028a0f"
	jouB1SupplyLineCranesOracle      = "bf3811b3-f3d2-4ffb-8019-e403a2cc1d81"
	jouB1SigiledSkinkOracle          = "4c17f6c3-48d3-4bd0-b0dd-76229d9c5b09"
	jouB1ReturnedRevelerOracle       = "f90df69a-d53f-44b1-833a-5223d2e6c296"
	jouB1FontOfFortunesOracle        = "4aeaf1cd-7346-4492-847d-1c4dad14fafd"
	jouB1FontOfIreOracle             = "88d4c611-3211-41b2-8453-6210f8394a0c"
	jouB1FontOfVigorOracle           = "422876dd-2801-4d0e-93ad-d9d48178f0d0"
	jouB1FontOfFertilityOracle       = "7fb6f960-38f6-45fa-bfec-d43fc174daa7"
	jouB1RevivingMelodyOracle        = "4834e6f2-1544-4e58-9092-ecc136cc0dee"
	jouB1CountermandOracle           = "43402a6c-9fba-4a80-90b3-14bac05c574b"
	jouB1DesecrationPlagueOracle     = "131d55d1-7429-4940-b7d2-d759aabff47b"
	jouB1FeastOfDreamsOracle         = "e86870b1-4028-4111-9717-7005f86afe85"
	jouB1PhalanxFormationOracle      = "799e7bb4-b03c-4568-b321-7810868bdcd7"
	jouB1RouseTheMobOracle           = "bc8b2b89-7118-4662-beda-9108cd08491b"
	jouB1CruelFeedingOracle          = "ef60bb5b-7cf5-4b4d-a725-4531fc0c852a"
	jouB1DesperateStandOracle        = "f8252564-3e61-4ed5-bef1-6ec22882854b"
	jouB1AerialFormationOracle       = "b0168b7e-a8b8-4ce3-af12-a43f3818b6a1"
	jouB1AjanisPresenceOracle        = "925a0b9b-a1fa-4aad-89d9-e2698ac3582f"
	jouB1BearType                    = "Creature — Bear"
	jouB1EnchantmentCreatureTypeLine = "Enchantment Creature — Nymph"
)

func TestJouB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		jouB1OreskosSwiftclawOracle:     "Oreskos Swiftclaw",
		jouB1PensiveMinotaurOracle:      "Pensive Minotaur",
		jouB1RottedHulkOracle:           "Rotted Hulk",
		jouB1BassaraTowerArcherOracle:   "Bassara Tower Archer",
		jouB1CloakedSirenOracle:         "Cloaked Siren",
		jouB1EagleOfTheWatchOracle:      "Eagle of the Watch",
		jouB1GoldForgedSentinelOracle:   "Gold-Forged Sentinel",
		jouB1SkyspearCavalryOracle:      "Skyspear Cavalry",
		jouB1GoldenHindOracle:           "Golden Hind",
		jouB1AkroanMastiffOracle:        "Akroan Mastiff",
		jouB1CyclopsOfEternalFuryOracle: "Cyclops of Eternal Fury",
		jouB1EidolonOfRhetoricOracle:    "Eidolon of Rhetoric",
		jouB1DictateOfHeliodOracle:      "Dictate of Heliod",
		jouB1FelhidePetrifierOracle:     "Felhide Petrifier",
		jouB1PheresBandWarchiefOracle:   "Pheres-Band Warchief",
		jouB1EidolonOfTheGreatRevel:     "Eidolon of the Great Revel",
		jouB1WarWingSirenOracle:         "War-Wing Siren",
		jouB1LagonnaBandOracle:          "Lagonna-Band Trailblazer",
		jouB1DawnbringerOracle:          "Dawnbringer Charioteers",
		jouB1ThunderhoofOracle:          "Pheres-Band Thunderhoof",
		jouB1AkroanLineBreakerOracle:    "Akroan Line Breaker",
		jouB1LeoninIconoclastOracle:     "Leonin Iconoclast",
		jouB1TritonCavalryOracle:        "Triton Cavalry",
		jouB1AgentOfErebosOracle:        "Agent of Erebos",
		jouB1ForgebornOreadsOracle:      "Forgeborn Oreads",
		jouB1ThassasDevourerOracle:      "Thassa's Devourer",
		jouB1ThoughtrenderLamiaOracle:   "Thoughtrender Lamia",
		jouB1OakheartDryadsOracle:       "Oakheart Dryads",
		jouB1DreadbringerLampadsOracle:  "Dreadbringer Lampads",
		jouB1WhitewaterNaiadsOracle:     "Whitewater Naiads",
		jouB1HumblerOfMortalsOracle:     "Humbler of Mortals",
		jouB1GoldenhideOxOracle:         "Goldenhide Ox",
		jouB1SatyrGrovedancerOracle:     "Satyr Grovedancer",
		jouB1SupplyLineCranesOracle:     "Supply-Line Cranes",
		jouB1SigiledSkinkOracle:         "Sigiled Skink",
		jouB1ReturnedRevelerOracle:      "Returned Reveler",
		jouB1FontOfFortunesOracle:       "Font of Fortunes",
		jouB1FontOfIreOracle:            "Font of Ire",
		jouB1FontOfVigorOracle:          "Font of Vigor",
		jouB1FontOfFertilityOracle:      "Font of Fertility",
		jouB1RevivingMelodyOracle:       "Reviving Melody",
		jouB1CountermandOracle:          "Countermand",
		jouB1DesecrationPlagueOracle:    "Desecration Plague",
		jouB1FeastOfDreamsOracle:        "Feast of Dreams",
		jouB1PhalanxFormationOracle:     "Phalanx Formation",
		jouB1RouseTheMobOracle:          "Rouse the Mob",
		jouB1CruelFeedingOracle:         "Cruel Feeding",
		jouB1DesperateStandOracle:       "Desperate Stand",
		jouB1AerialFormationOracle:      "Aerial Formation",
		jouB1AjanisPresenceOracle:       "Ajani's Presence",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// jouB1Heroic casts Hearts on Fire (+2/+1) at the creature, the spell that
// triggers heroic.
func jouB1Heroic(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Hearts on Fire", "Instant", effectstest.J25b1HeartsOnFireOracle, effectstest.CardRefs(target))
}

// jouB1Enters puts a catalog permanent onto the battlefield and announces
// its entry, which is what a constellation trigger watches.
func jouB1Enters(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	id := effectstest.PushCatalogPermanent(g, owner, name, typeLine, oracle, false)
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: owner}) })
	return id
}

// --- printed keywords ------------------------------------------------------

func TestJouB1KeywordCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Bassara Tower Archer", "Creature — Human Archer", jouB1BassaraTowerArcherOracle, []string{"reach", "hexproof"}},
		{"Cloaked Siren", "Creature — Siren", jouB1CloakedSirenOracle, []string{"flash", "flying"}},
		{"Eagle of the Watch", "Creature — Bird", jouB1EagleOfTheWatchOracle, []string{"flying", "vigilance"}},
		{"Gold-Forged Sentinel", "Artifact Creature — Chimera", jouB1GoldForgedSentinelOracle, []string{"flying"}},
		{"Skyspear Cavalry", "Creature — Human Soldier", jouB1SkyspearCavalryOracle, []string{"flying", "double strike"}},
		{"War-Wing Siren", "Creature — Siren Soldier", jouB1WarWingSirenOracle, []string{"flying"}},
		{"Dawnbringer Charioteers", "Creature — Human Soldier", jouB1DawnbringerOracle, []string{"flying", "lifelink"}},
		{"Supply-Line Cranes", "Creature — Bird", jouB1SupplyLineCranesOracle, []string{"flying"}},
		{"Felhide Petrifier", "Creature — Minotaur Warrior", jouB1FelhidePetrifierOracle, []string{"deathtouch"}},
		{"Pheres-Band Warchief", "Creature — Centaur Warrior", jouB1PheresBandWarchiefOracle, []string{"vigilance", "trample"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			abilities := effectstest.EffectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !effectstest.HasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

func TestJouB1CloakedSirenHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B13HandCard(me, "Cloaked Siren", "Creature — Siren", jouB1CloakedSirenOracle)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("Cloaked Siren at the beginning of combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Cloaked Siren did not resolve")
	}
}

// --- activated abilities ---------------------------------------------------

func TestJouB1GoldenHindTapsForGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hind := effectstest.B12Push(g, me.ID, "Golden Hind", "Creature — Elk", jouB1GoldenHindOracle, 2, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, hind, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
	if !effectstest.HmlTapped(t, g, hind) {
		t.Error("the Hind did not tap")
	}
}

func TestJouB1AkroanMastiffTapsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mastiff := effectstest.B12Push(g, me.ID, "Akroan Mastiff", "Creature — Dog", jouB1AkroanMastiffOracle, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", jouB1BearType, 2, 2)
	effectstest.ApaMana(me, "W")
	effectstest.ApaActivate(t, g, me, mastiff, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	if !effectstest.HmlTapped(t, g, bear) {
		t.Error("the target is untapped")
	}
	if !effectstest.HmlTapped(t, g, mastiff) {
		t.Error("the Mastiff did not tap")
	}
}

// --- statics -----------------------------------------------------------------

func TestJouB1CyclopsGivesYourCreaturesHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cyclops := effectstest.B12Push(g, me.ID, "Cyclops of Eternal Fury", "Enchantment Creature — Cyclops", jouB1CyclopsOfEternalFuryOracle, 5, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", jouB1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", jouB1BearType, 2, 2)
	for _, id := range []uuid.UUID{cyclops, mine} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Error("a creature you control has no haste")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature has haste")
	}
}

func TestJouB1DictateOfHeliodPumpsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Dictate of Heliod", "Enchantment", jouB1DictateOfHeliodOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", jouB1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", jouB1BearType, 2, 2)
	if p, tough := effectstest.PtOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("your creature is %d/%d, want 4/4", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestJouB1DictateOfHeliodHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.B13HandCard(me, "Dictate of Heliod", "Enchantment", jouB1DictateOfHeliodOracle)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("Dictate of Heliod at the beginning of combat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Dictate of Heliod did not resolve")
	}
}

func TestJouB1FelhidePetrifierGivesOtherMinotaursDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Felhide Petrifier", "Creature — Minotaur Warrior", jouB1FelhidePetrifierOracle, 2, 3)
	minotaur := effectstest.B12Creature(g, me.ID, "Minotaur", "Creature — Minotaur", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", jouB1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Minotaur", "Creature — Minotaur", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, minotaur, "deathtouch") {
		t.Error("your other Minotaur has no deathtouch")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("a non-Minotaur has deathtouch")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "deathtouch") {
		t.Error("an opponent's Minotaur has deathtouch")
	}
}

func TestJouB1PheresBandWarchiefLeadsOtherCentaurs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chief := effectstest.B12Push(g, me.ID, "Pheres-Band Warchief", "Creature — Centaur Warrior", jouB1PheresBandWarchiefOracle, 3, 3)
	centaur := effectstest.B12Creature(g, me.ID, "Centaur", "Creature — Centaur", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", jouB1BearType, 2, 2)
	if p, tough := effectstest.PtOf(t, g, centaur); p != 3 || tough != 3 {
		t.Errorf("the other Centaur is %d/%d, want 3/3", p, tough)
	}
	for _, kw := range []string{"vigilance", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, centaur, kw) {
			t.Errorf("the other Centaur has no %s", kw)
		}
	}
	if p, tough := effectstest.PtOf(t, g, chief); p != 3 || tough != 3 {
		t.Errorf("the Warchief is %d/%d, want its own 3/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("a non-Centaur is %d/%d, want 2/2", p, tough)
	}
}

func TestJouB1EidolonOfRhetoricRefusesTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Eidolon of Rhetoric", "Enchantment Creature — Spirit", jouB1EidolonOfRhetoricOracle, 1, 4)
	effectstest.CastCatalogSpell(t, g, "First", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Second", "Sorcery", "", nil); err == nil {
		t.Error("a second spell this turn was cast")
	}
}

func TestJouB1EidolonOfTheGreatRevelBurnsCheapSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Eidolon of the Great Revel", "Enchantment Creature — Spirit", jouB1EidolonOfTheGreatRevel, 2, 2)
	life := me.Life
	effectstest.CastFromHand(t, g, game.Card{Name: "Cheap", TypeLine: "Sorcery", ManaCost: "{2}{R}"}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("a mana value 3 spell: life %d -> %d, want -2", life, me.Life)
	}
}

// --- heroic --------------------------------------------------------------------

func TestJouB1HeroicCounters(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		want                   int
	}{
		{"War-Wing Siren", "Creature — Siren Soldier", jouB1WarWingSirenOracle, 1},
		{"Lagonna-Band Trailblazer", "Creature — Centaur Scout", jouB1LagonnaBandOracle, 1},
		{"Dawnbringer Charioteers", "Creature — Human Soldier", jouB1DawnbringerOracle, 1},
		{"Pheres-Band Thunderhoof", "Creature — Centaur Warrior", jouB1ThunderhoofOracle, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			c := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 3)
			jouB1Heroic(t, g, c)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.B18Counters(t, g, c, game.CounterPlusOne); got != tc.want {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.want)
			}
		})
	}
}

func TestJouB1AkroanLineBreakerGainsPowerAndIntimidate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	c := effectstest.B12Push(g, me.ID, "Akroan Line Breaker", "Creature — Human Warrior", jouB1AkroanLineBreakerOracle, 2, 1)
	jouB1Heroic(t, g, c)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, c); p != 6 {
		t.Errorf("power %d, want 6 (2 printed, +2 Hearts on Fire, +2 heroic)", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, c, "intimidate") {
		t.Error("no intimidate")
	}
}

func TestJouB1LeoninIconoclastDestroysAnEnchantmentCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	c := effectstest.B12Push(g, me.ID, "Leonin Iconoclast", "Creature — Cat Monk", jouB1LeoninIconoclastOracle, 3, 2)
	nymph := effectstest.B12Creature(g, opp.ID, "Nymph", jouB1EnchantmentCreatureTypeLine, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", jouB1BearType, 2, 2)
	jouB1Heroic(t, g, c)
	effectstest.AnswerPickTargetIfOpen(t, g, nymph)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(nymph) {
		t.Error("the enchantment creature survived")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature that isn't an enchantment was destroyed")
	}
}

func TestJouB1TritonCavalryMayBounceAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	c := effectstest.B12Push(g, me.ID, "Triton Cavalry", "Creature — Merfolk Soldier", jouB1TritonCavalryOracle, 2, 4)
	ench := effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	jouB1Heroic(t, g, c)
	effectstest.AnswerPickTargetIfOpen(t, g, ench)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) || !opp.Hand.Contains(ench) {
		t.Error("the enchantment did not return to its owner's hand")
	}
}

// --- constellation -------------------------------------------------------------

func TestJouB1AgentOfErebosExilesAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.PushGraveyardCardForTest(opp, "One")
	effectstest.PushGraveyardCardForTest(opp, "Two")
	jouB1Enters(g, me.ID, "Agent of Erebos", "Enchantment Creature — Zombie", jouB1AgentOfErebosOracle)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := opp.Graveyard.Size(); n != 0 {
		t.Errorf("the graveyard holds %d cards, want 0", n)
	}
}

func TestJouB1ForgebornOreadsPings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	jouB1Enters(g, me.ID, "Forgeborn Oreads", jouB1EnchantmentCreatureTypeLine, jouB1ForgebornOreadsOracle)
	effectstest.AnswerPickTargetIfOpen(t, g, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ogre).DamageMarked; got != 1 {
		t.Errorf("the target has %d damage, want 1", got)
	}
}

func TestJouB1ThassasDevourerMillsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	library := opp.Library.Size()
	jouB1Enters(g, me.ID, "Thassa's Devourer", "Enchantment Creature — Elemental", jouB1ThassasDevourerOracle)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := library - opp.Library.Size(); got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

func TestJouB1ThoughtrenderLamiaMakesEachOpponentDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	for _, s := range g.Seats {
		effectstest.D2dSetHand(s, 3)
	}
	jouB1Enters(g, me.ID, "Thoughtrender Lamia", "Enchantment Creature — Lamia", jouB1ThoughtrenderLamiaOracle)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := 1
		if s.ID == me.ID {
			want = 0
		}
		if got := effectstest.DiscardOwed(g, s.ID); got != want {
			t.Errorf("%s owes %d discards, want %d", s.Name, got, want)
		}
	}
}

func TestJouB1ConstellationPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
		keyword                string
	}{
		{"Oakheart Dryads", "Enchantment Creature — Nymph Dryad", jouB1OakheartDryadsOracle, 3, 3, ""},
		{"Dreadbringer Lampads", jouB1EnchantmentCreatureTypeLine, jouB1DreadbringerLampadsOracle, 2, 2, "intimidate"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			bear := effectstest.B12Creature(g, me.ID, "Bear", jouB1BearType, 2, 2)
			jouB1Enters(g, me.ID, tc.name, tc.typeLine, tc.oracle)
			effectstest.AnswerPickTargetIfOpen(t, g, bear)
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, bear); p != tc.p || tough != tc.tough {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
			}
			if tc.keyword != "" && !effectstest.HasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear has no %s", tc.keyword)
			}
		})
	}
}

func TestJouB1HumblerOfMortalsGivesYourCreaturesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", jouB1BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", jouB1BearType, 2, 2)
	jouB1Enters(g, me.ID, "Humbler of Mortals", "Enchantment Creature — Elemental", jouB1HumblerOfMortalsOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("your creature has no trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature has trample")
	}
}

func TestJouB1WhitewaterNaiadsMakesATargetUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	// A second, blockable attacker gives the defender a legal block, so the
	// step parks for the declaration (an attack of only unblockable
	// creatures has none to make).
	goat := effectstest.ReqCreature(g, me.ID, "Goat", "", 1, 1, "haste")
	wall := effectstest.ReqCreature(g, opp.ID, "Wall", "", 0, 4)
	jouB1Enters(g, me.ID, "Whitewater Naiads", jouB1EnchantmentCreatureTypeLine, jouB1WhitewaterNaiadsOracle)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ReqToBlockers(t, g, bear, goat)
	if err := effectstest.ReqBlock(t, g, bear, wall); err == nil {
		t.Error("a creature that can't be blocked was blocked")
	}
	if err := effectstest.ReqBlock(t, g, goat, wall); err != nil {
		t.Errorf("the untargeted attacker could not be blocked: %v", err)
	}
}

func TestJouB1GoldenhideOxMakesATargetBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	wall := effectstest.ReqCreature(g, opp.ID, "Wall", "", 0, 4)
	jouB1Enters(g, me.ID, "Goldenhide Ox", "Enchantment Creature — Ox", jouB1GoldenhideOxOracle)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ReqToBlockers(t, g, bear)
	effectstest.ReqRefusal(t, g.PassPriority())
	if err := effectstest.ReqBlock(t, g, bear, wall); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Bear blocked: %v", err)
	}
}

// --- enters, attacks and dies ------------------------------------------------

func TestJouB1EntersCounters(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Satyr Grovedancer", "Creature — Satyr Shaman", jouB1SatyrGrovedancerOracle},
		{"Supply-Line Cranes", "Creature — Bird", jouB1SupplyLineCranesOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			bear := effectstest.B12Creature(g, me.ID, "Bear", jouB1BearType, 2, 2)
			effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 1, 1)
			effectstest.AnswerPickTargetIfOpen(t, g, bear)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.B18Counters(t, g, bear, game.CounterPlusOne); got != 1 {
				t.Errorf("%d +1/+1 counters on the target, want 1", got)
			}
		})
	}
}

func TestJouB1SigiledSkinkScriesWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	skink := effectstest.PushCatalogPermanent(g, me.ID, "Sigiled Skink", "Creature — Lizard", jouB1SigiledSkinkOracle, false)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(skink, them.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no scry prompt from the attack trigger: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 1 {
		t.Errorf("scry 1 offered %d cards", len(c.ScryCards))
	}
}

func TestJouB1ReturnedRevelerMillsEachPlayerThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	reveler := effectstest.B12Push(g, g.Seats[0].ID, "Returned Reveler", "Creature — Zombie Satyr", jouB1ReturnedRevelerOracle, 1, 3)
	libraries := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		libraries[s.ID] = s.Library.Size()
	}
	effectstest.B27Kill(g, reveler)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if got := libraries[s.ID] - s.Library.Size(); got != 3 {
			t.Errorf("%s milled %d, want 3", s.Name, got)
		}
	}
}

// --- the Fonts -----------------------------------------------------------------

func TestJouB1FontOfFortunesDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	font := effectstest.PushCatalogPermanent(g, me.ID, "Font of Fortunes", "Enchantment", jouB1FontOfFortunesOracle, false)
	hand := me.Hand.Size()
	effectstest.ApaMana(me, "C", "U")
	effectstest.ApaActivate(t, g, me, font, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(font) {
		t.Error("the Font was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

func TestJouB1FontOfIreDealsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	font := effectstest.PushCatalogPermanent(g, me.ID, "Font of Ire", "Enchantment", jouB1FontOfIreOracle, false)
	life := opp.Life
	effectstest.ApaMana(me, "C", "C", "C", "R")
	effectstest.ApaActivate(t, g, me, font, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	if g.Battlefield.Contains(font) {
		t.Error("the Font was not sacrificed")
	}
	if opp.Life != life-5 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-5)
	}
}

func TestJouB1FontOfVigorGainsSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	font := effectstest.PushCatalogPermanent(g, me.ID, "Font of Vigor", "Enchantment", jouB1FontOfVigorOracle, false)
	life := me.Life
	effectstest.ApaMana(me, "C", "C", "W")
	effectstest.ApaActivate(t, g, me, font, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(font) {
		t.Error("the Font was not sacrificed")
	}
	if me.Life != life+7 {
		t.Errorf("life %d, want %d", me.Life, life+7)
	}
}

func TestJouB1FontOfFertilitySearchesForABasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	font := effectstest.PushCatalogPermanent(g, me.ID, "Font of Fertility", "Enchantment", jouB1FontOfFertilityOracle, false)
	effectstest.ApaMana(me, "C", "G")
	effectstest.ApaActivate(t, g, me, font, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(font) {
		t.Error("the Font was not sacrificed")
	}
	if _, ok := effectstest.VisBattlefieldNamed(g, me.ID, "Forest"); !ok && effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Error("no basic land search")
	}
}

// --- the spells ----------------------------------------------------------------

func TestJouB1RevivingMelodyReturnsBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	creature := effectstest.WthGraveyardCard(me, "Dead Bear", jouB1BearType)
	ench := effectstest.WthGraveyardCard(me, "Old Aura", "Enchantment — Aura")
	effectstest.CastModal(t, g, "Reviving Melody", "Sorcery", jouB1RevivingMelodyOracle, []int{0, 1}, effectstest.CardRefs(creature, ench))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{creature, ench} {
		if !me.Hand.Contains(id) || me.Graveyard.Contains(id) {
			t.Error("a card did not return to hand")
		}
	}
}

func TestJouB1CountermandCountersAndMillsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.B13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	library := opp.Library.Size()
	id := effectstest.B13HandCard(me, "Countermand", "Instant", jouB1CountermandOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.B16TargetCard(bolt)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the spell was not countered")
	}
	if got := library - opp.Library.Size(); got != 4 {
		t.Errorf("its controller milled %d, want 4", got)
	}
}

func TestJouB1DesecrationPlagueDestroysAnEnchantmentOrLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", jouB1BearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Desecration Plague", "Sorcery", jouB1DesecrationPlagueOracle, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("Desecration Plague targeted a creature")
	}
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Desecration Plague", "Sorcery", jouB1DesecrationPlagueOracle, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land survived")
	}
}

func TestJouB1FeastOfDreamsDestroysAnEnchantmentCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := effectstest.Top100Opponent(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", jouB1BearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Feast of Dreams", "Instant", jouB1FeastOfDreamsOracle, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("Feast of Dreams targeted an unenchanted, non-enchantment creature")
	}
	nymph := effectstest.B12Creature(g, opp.ID, "Nymph", jouB1EnchantmentCreatureTypeLine, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Feast of Dreams", "Instant", jouB1FeastOfDreamsOracle, effectstest.CardTarget(nymph))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(nymph) {
		t.Error("the enchantment creature survived")
	}
}

// --- strive ------------------------------------------------------------------------

var jouB1Strive = []struct {
	name, typeLine, oracle, cost string
	p, tough                     int
	keywords                     []string
	twoTargetsMV                 int
}{
	{"Phalanx Formation", "Instant", jouB1PhalanxFormationOracle, "{2}{W}", 2, 2, []string{"double strike"}, 5},
	{"Rouse the Mob", "Instant", jouB1RouseTheMobOracle, "{R}", 4, 2, []string{"trample"}, 4},
	{"Cruel Feeding", "Instant", jouB1CruelFeedingOracle, "{B}", 3, 2, []string{"lifelink"}, 4},
	{"Desperate Stand", "Sorcery", jouB1DesperateStandOracle, "{R}{W}", 4, 2, []string{"first strike", "vigilance"}, 4},
	{"Aerial Formation", "Instant", jouB1AerialFormationOracle, "{U}", 3, 3, []string{"flying"}, 4},
	{"Ajani's Presence", "Instant", jouB1AjanisPresenceOracle, "{W}", 3, 3, []string{"indestructible"}, 4},
}

func TestJouB1StriveCostsMorePerExtraTarget(t *testing.T) {
	for _, tc := range jouB1Strive {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			a := effectstest.B12Creature(g, me.ID, "A", jouB1BearType, 2, 2)
			b := effectstest.B12Creature(g, me.ID, "B", jouB1BearType, 2, 2)
			if got := effectstest.SelfPriced(t, g, me, tc.oracle, tc.typeLine, tc.cost, game.ZoneHand, effectstest.CardRefs(a)).ManaValue(); got != jouB1ManaValueOf(t, tc.cost) {
				t.Errorf("one target costs mana value %d, want the printed cost", got)
			}
			if got := effectstest.SelfPriced(t, g, me, tc.oracle, tc.typeLine, tc.cost, game.ZoneHand, effectstest.CardRefs(a, b)).ManaValue(); got != tc.twoTargetsMV {
				t.Errorf("two targets cost mana value %d, want %d", got, tc.twoTargetsMV)
			}
		})
	}
}

func TestJouB1StrivePumpsEachTarget(t *testing.T) {
	for _, tc := range jouB1Strive {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := effectstest.Top100ActiveSeat(g)
			a := effectstest.B12Creature(g, me.ID, "A", jouB1BearType, 2, 2)
			b := effectstest.B12Creature(g, me.ID, "B", jouB1BearType, 2, 2)
			idle := effectstest.B12Creature(g, me.ID, "Idle", jouB1BearType, 2, 2)
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, effectstest.CardRefs(a, b))
			effectstest.PassPriorityAroundTable(t, g)
			for _, id := range []uuid.UUID{a, b} {
				if p, tough := effectstest.PtOf(t, g, id); p != tc.p || tough != tc.tough {
					t.Errorf("a target is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
				}
				for _, kw := range tc.keywords {
					if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
						t.Errorf("a target has no %s", kw)
					}
				}
			}
			if p, tough := effectstest.PtOf(t, g, idle); p != 2 || tough != 2 {
				t.Errorf("an untargeted creature is %d/%d, want 2/2", p, tough)
			}
		})
	}
}

// jouB1ManaValueOf is the mana value of a printed cost.
func jouB1ManaValueOf(t *testing.T, cost string) int {
	t.Helper()
	parsed, err := game.ParseCost(cost)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", cost, err)
	}
	return parsed.ManaValue()
}
