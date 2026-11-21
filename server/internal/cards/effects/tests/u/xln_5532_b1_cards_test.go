package testsu

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// xln_5532_b1_cards_test.go — Ixalan, batch 1 (#5532): seven vanilla
// bodies, seven creatures whose whole text is keywords, eight creatures
// that explore as they enter, and twenty-eight cards whose text an
// existing catalog card already prints. Each test plays the real catalog
// entry through the engine and asserts what its text says.

const (
	xlnB1SeekersSquireOracle   = "a494cbe7-a4a5-4657-9f68-3811fe56a17e"
	xlnB1SailorOfMeansOracle   = "c4efae03-99de-4439-925b-504602623bfd"
	xlnB1WilyGoblinOracle      = "d27ffb4c-3a34-41aa-b464-6b897132c4ea"
	xlnB1ProsperousOracle      = "9dc6281b-08e0-4c66-9aef-abc7184ca36a"
	xlnB1InspiringClericOracle = "65e19aec-e8ea-412f-a239-bd8d1b78249c"
	xlnB1PyromancerOracle      = "866586f4-ad97-44fa-a110-6a7a3cb3591b"
	xlnB1HeartstopperOracle    = "d07f497e-4cf8-4e12-b17e-e3d58ff9df47"
	xlnB1HorncrestOracle       = "bf71c210-8fef-4eda-a943-8c52cbae6159"
	xlnB1CeratopsOracle        = "399ea8a4-55f5-4f14-ad7a-7af306959496"
	xlnB1DeeprootOracle        = "507c2921-7070-482d-8754-dc667e3315bd"
	xlnB1DaggertoothOracle     = "f5ef25fb-545e-41b7-8fed-d3cfff3d8668"
	xlnB1HatchlingOracle       = "91638137-06ba-4b21-b088-28d2b4d1d8da"
	xlnB1CastawayOracle        = "9dc4b172-da33-40c0-ab29-d68807dbe93f"
	xlnB1OneWithTheWindOracle  = "9203bee0-718a-4ede-9556-5521417c4080"
	xlnB1SwashbucklingOracle   = "4992144f-7a36-4f30-9e8a-4355288b50c5"
	xlnB1PiousOracle           = "e6762a65-5c88-4652-ac3d-b9de220634c7"
	xlnB1UnfriendlyFireOracle  = "ebb5b792-3906-4cb5-b8f9-063f94c9d135"
	xlnB1DarkNourishOracle     = "95650a60-6f1d-4594-8b53-7fd406212207"
	xlnB1WalkThePlankOracle    = "9093e6f1-12bc-4dfb-bc78-6ead0752ad61"
	xlnB1BrightReprisalOracle  = "1de2c216-0921-45eb-9613-a68d18fbee77"
	xlnB1CallToTheFeastOracle  = "30536d1f-8b1a-474f-a508-d3426480a532"
	xlnB1QueensCommissionOrcl  = "5e4c02c2-2e54-4e8d-8d94-6b2454027e7e"
	xlnB1PiratesPrizeOracle    = "6de8bdc0-9b2a-4105-bbac-d5a2f73f52dc"
	xlnB1CrashOracle           = "8832b76b-58b7-41b2-b2f6-842563b55b36"
	xlnB1SpreadingRotOracle    = "3c1b95bd-a2fc-47cd-8940-e4af03bd3ce2"
	xlnB1CannonadeOracle       = "7e108b31-2ea3-4129-b324-a2d31c78bbe6"
	xlnB1RitualOracle          = "7c9d8065-9ab0-4468-a11c-383bd036029f"
	xlnB1ContemptOracle        = "38ddbe0b-07ef-4cdf-ad6c-cd8b8ba6d206"
	xlnB1PounceOracle          = "f4126dac-add4-4d5f-bf52-6f74c7c6aa67"
	xlnB1RaptorCompanionOracle = "20e65887-c68c-4dc0-b74b-9ce093f40b06"

	xlnB1BearType = "Creature — Bear"
)

// xlnB1Vanilla are the batch's bodies with no rules text.
var xlnB1Vanilla = map[string]string{
	"745be03c-69dc-4239-a5e7-38587b56eac0": "Ancient Brontodon",
	"655c59d5-8666-476c-b494-411fa2e243a7": "Frenzied Raptor",
	"a382165d-d715-4836-b1ce-9e5d8dc56fb4": "Gilded Sentinel",
	"17aa790d-d16a-4a90-b8ba-cd50e62f02f0": "Headwater Sentries",
	"af5302d3-782b-4d92-bc97-dccbfc76ac95": "Looming Altisaur",
	"b288c863-5377-473d-a628-bb199993a29c": "Queen's Bay Soldier",
	xlnB1RaptorCompanionOracle:             "Raptor Companion",
}

// xlnB1Keyworded are the batch's creatures whose whole text is printed
// keywords.
var xlnB1Keyworded = map[string]struct {
	name, typeLine string
	keywords       []string
}{
	"21dd60d5-76c8-469e-8c4e-112e377d2d59": {"Charging Monstrosaur", "Creature — Dinosaur", []string{"trample", "haste"}},
	"94399b1f-e66a-422e-ae41-6013fe741e8c": {"Grazing Whiptail", "Creature — Dinosaur", []string{"reach"}},
	"8516d8ba-b3f9-4c45-8d70-5dab476e478f": {"Huatli's Snubhorn", "Creature — Dinosaur", []string{"vigilance"}},
	"b909ff27-423f-4e84-887b-7840149a85be": {"Nest Robber", "Creature — Dinosaur", []string{"haste"}},
	"a7520962-46e0-4abd-bb19-569334913a05": {"Shining Aerosaur", "Creature — Dinosaur", []string{"flying"}},
	"5fe2ed67-55a6-4816-910f-0e3c38ea06a6": {"Sky Terror", "Creature — Dinosaur", []string{"flying", "menace"}},
	"678c76bf-fc31-47e4-8c19-712e15bba9e3": {"Skyblade of the Legion", "Creature — Vampire Soldier", []string{"flying"}},
}

// xlnB1Explorers are the batch's creatures that explore as they enter,
// with any keyword they also print.
var xlnB1Explorers = map[string]struct {
	name, typeLine string
	keywords       []string
}{
	xlnB1SeekersSquireOracle:               {"Seekers' Squire", "Creature — Human Scout", nil},
	"86468528-833d-48f9-b299-cdf54dd47d77": {"Tishana's Wayfinder", "Creature — Merfolk Scout", nil},
	"db05f30f-2866-48ed-8a7d-264d94c0257a": {"Sunrise Seeker", "Creature — Human Scout", []string{"vigilance"}},
	"59104038-28c0-4ebd-9201-d7ae60d79c1c": {"Queen's Agent", "Creature — Vampire Scout", []string{"lifelink"}},
	"e2155e61-f389-4b1d-9598-9a79cbccaaa2": {"Emissary of Sunrise", "Creature — Human Cleric", []string{"first strike"}},
	"9f0d82ae-38bf-45d8-8cda-982b6ead1d72": {"Siren Lookout", "Creature — Siren Pirate", []string{"flying"}},
	"c8afa4b5-01be-400a-836f-aa27e1ae1346": {"Brazen Buccaneers", "Creature — Human Pirate", []string{"haste"}},
	"7ae7cfde-8f2a-4e66-abe2-c15477e9fac3": {"Dire Fleet Interloper", "Creature — Human Pirate", []string{"menace"}},
}

func TestXlnB1EveryCardIsRegisteredFull(t *testing.T) {
	all := map[string]string{
		xlnB1SailorOfMeansOracle:   "Sailor of Means",
		xlnB1WilyGoblinOracle:      "Wily Goblin",
		xlnB1ProsperousOracle:      "Prosperous Pirates",
		xlnB1InspiringClericOracle: "Inspiring Cleric",
		xlnB1PyromancerOracle:      "Storm Fleet Pyromancer",
		xlnB1HeartstopperOracle:    "Skittering Heartstopper",
		xlnB1HorncrestOracle:       "Bonded Horncrest",
		xlnB1CeratopsOracle:        "Spike-Tailed Ceratops",
		xlnB1DeeprootOracle:        "Deeproot Champion",
		xlnB1DaggertoothOracle:     "Ravenous Daggertooth",
		xlnB1HatchlingOracle:       "Raptor Hatchling",
		xlnB1CastawayOracle:        "Castaway's Despair",
		xlnB1OneWithTheWindOracle:  "One With the Wind",
		xlnB1SwashbucklingOracle:   "Swashbuckling",
		xlnB1PiousOracle:           "Pious Interdiction",
		xlnB1UnfriendlyFireOracle:  "Unfriendly Fire",
		xlnB1DarkNourishOracle:     "Dark Nourishment",
		xlnB1WalkThePlankOracle:    "Walk the Plank",
		xlnB1BrightReprisalOracle:  "Bright Reprisal",
		xlnB1CallToTheFeastOracle:  "Call to the Feast",
		xlnB1QueensCommissionOrcl:  "Queen's Commission",
		xlnB1PiratesPrizeOracle:    "Pirate's Prize",
		xlnB1CrashOracle:           "Crash the Ramparts",
		xlnB1SpreadingRotOracle:    "Spreading Rot",
		xlnB1CannonadeOracle:       "Fiery Cannonade",
		xlnB1RitualOracle:          "Ritual of Rejuvenation",
		xlnB1ContemptOracle:        "Vraska's Contempt",
		xlnB1PounceOracle:          "Pounce",
	}
	for oracle, name := range xlnB1Vanilla {
		all[oracle] = name
	}
	for oracle, k := range xlnB1Keyworded {
		all[oracle] = k.name
	}
	for oracle, k := range xlnB1Explorers {
		all[oracle] = k.name
	}
	if len(all) != 50 {
		t.Fatalf("the batch lists %d cards, want 50", len(all))
	}
	for oracle, name := range all {
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

// --- bodies and keywords -----------------------------------------------------

func TestXlnB1VanillaBodiesDeclareNoAbilities(t *testing.T) {
	for oracle, name := range xlnB1Vanilla {
		spec, _ := effects.Lookup(oracle)
		if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 ||
			len(spec.Activated) != 0 || len(spec.PrintedKeywords) != 0 {
			t.Errorf("%s is a vanilla body and declares no abilities", name)
		}
	}
}

func TestXlnB1RaptorCompanionIsAPlainCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raptor := effectstest.B12Push(g, me.ID, "Raptor Companion", "Creature — Dinosaur", xlnB1RaptorCompanionOracle, 3, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", xlnB1BearType, 2, 2)
	if err := effectstest.PorBlock(t, g, raptor, opp.ID, bear); err != nil {
		t.Errorf("a vanilla Raptor Companion couldn't be blocked: %v", err)
	}
	if p, tough := effectstest.PtOf(t, g, raptor); p != 3 || tough != 1 {
		t.Errorf("Raptor Companion is %d/%d, want 3/1", p, tough)
	}
}

func TestXlnB1KeywordedCreaturesHaveTheirKeywords(t *testing.T) {
	for oracle, k := range xlnB1Keyworded {
		t.Run(k.name, func(t *testing.T) {
			spec, _ := effects.Lookup(oracle)
			if !slices.Equal(spec.PrintedKeywords, k.keywords) {
				t.Errorf("printed keywords %v, want %v", spec.PrintedKeywords, k.keywords)
			}
			if spec.OnResolve != nil || len(spec.Triggered) != 0 || len(spec.Static) != 0 || len(spec.Activated) != 0 {
				t.Errorf("%s declares more than its keywords", k.name)
			}
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, k.name, k.typeLine, oracle, 2, 2)
			for _, kw := range k.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s on the battlefield", k.name, kw)
				}
			}
		})
	}
}

// Flying: a ground creature can't block Shining Aerosaur; Grazing
// Whiptail (reach) can (CR 702.9b).
func TestXlnB1AerosaurIsBlockedOnlyByFlyingOrReach(t *testing.T) {
	const aerosaur, whiptail = "a7520962-46e0-4abd-bb19-569334913a05", "94399b1f-e66a-422e-ae41-6013fe741e8c"
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flyer := effectstest.B12Push(g, me.ID, "Shining Aerosaur", "Creature — Dinosaur", aerosaur, 3, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", xlnB1BearType, 2, 2)
	if err := effectstest.PorBlock(t, g, flyer, opp.ID, bear); err == nil {
		t.Error("a ground creature blocked Shining Aerosaur")
	}
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	flyer = effectstest.B12Push(g, me.ID, "Shining Aerosaur", "Creature — Dinosaur", aerosaur, 3, 4)
	reach := effectstest.B12Push(g, opp.ID, "Grazing Whiptail", "Creature — Dinosaur", whiptail, 3, 4)
	if err := effectstest.PorBlock(t, g, flyer, opp.ID, reach); err != nil {
		t.Errorf("Grazing Whiptail (reach) couldn't block Shining Aerosaur: %v", err)
	}
}

// Haste: Nest Robber attacks the turn it is cast (CR 702.10b).
func TestXlnB1NestRobberAttacksTheTurnItArrives(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	id := effectstest.CastHoldCreature(t, g, "Nest Robber", "Creature — Dinosaur", "b909ff27-423f-4e84-887b-7840149a85be", 2, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Errorf("Nest Robber couldn't attack the turn it arrived: %v", err)
	}
}

// --- explore -------------------------------------------------------------------

func TestXlnB1ExplorersExploreALandToHand(t *testing.T) {
	for oracle, k := range xlnB1Explorers {
		t.Run(k.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			top := effectstest.PutOnTopForTest(me, "Forest", "Basic Land — Forest")
			id := effectstest.Msc17Play(t, g, me, game.Card{Name: k.name, TypeLine: k.typeLine, OracleID: oracle, Power: 2, Toughness: 2})
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(top) {
				t.Error("the revealed land did not go to the hand")
			}
			if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 0 {
				t.Errorf("%d +1/+1 counters after revealing a land, want 0", got)
			}
			for _, kw := range k.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
					t.Errorf("%s has no %s", k.name, kw)
				}
			}
		})
	}
}

func TestXlnB1SeekersSquireExploresANonlandIntoACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	top := effectstest.PutOnTopForTest(me, "Some Sorcery", "Sorcery")
	squire := effectstest.Msc17Play(t, g, me, game.Card{Name: "Seekers' Squire", TypeLine: "Creature — Human Scout", OracleID: xlnB1SeekersSquireOracle, Power: 1, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, squire, game.CounterPlusOne); got != 1 {
		t.Errorf("the Squire has %d +1/+1 counters, want 1", got)
	}
	effectstest.AnswerExploreQuestion(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(top) {
		t.Error("the revealed card was not put into the graveyard")
	}
}

// --- enters triggers -------------------------------------------------------------

func TestXlnB1TreasureMakersCreateTheirTreasures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		want                   int
	}{
		{"Sailor of Means", "Creature — Human Pirate", xlnB1SailorOfMeansOracle, 1},
		{"Wily Goblin", "Creature — Goblin Pirate", xlnB1WilyGoblinOracle, 1},
		{"Prosperous Pirates", "Creature — Human Pirate", xlnB1ProsperousOracle, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 1, 1)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Fatalf("%s did not resolve", tc.name)
			}
			if got := effectstest.TreasureCount(g); got != tc.want {
				t.Errorf("%d Treasures, want %d", got, tc.want)
			}
		})
	}
}

func TestXlnB1InspiringClericGainsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastHoldCreature(t, g, "Inspiring Cleric", "Creature — Vampire Cleric", xlnB1InspiringClericOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

func TestXlnB1StormFleetPyromancerDealsTwoOnlyAfterAnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", xlnB1BearType, 2, 2)
	effectstest.AvB16CastCreature(t, g, "Storm Fleet Pyromancer", "Creature — Human Pirate Wizard", xlnB1PyromancerOracle, 3, 2)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the raid trigger fired without an attack this turn")
	}
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", xlnB1BearType, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.AvB16CastCreature(t, g, "Storm Fleet Pyromancer", "Creature — Human Pirate Wizard", xlnB1PyromancerOracle, 3, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 2/2 survived 2 damage")
	}
}

// --- other creatures -------------------------------------------------------------

func TestXlnB1SkitteringHeartstopperGainsDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	insect := effectstest.B12Push(g, me.ID, "Skittering Heartstopper", "Creature — Insect", xlnB1HeartstopperOracle, 1, 2)
	if effectstest.HasEffectiveKeyword(t, g, insect, "deathtouch") {
		t.Fatal("deathtouch before the ability")
	}
	effectstest.ApaMana(me, "B")
	effectstest.ApaActivate(t, g, me, insect, 0, game.ActivateAbilityParams{})
	if !effectstest.HasEffectiveKeyword(t, g, insect, "deathtouch") {
		t.Error("no deathtouch after the ability resolved")
	}
}

func TestXlnB1BondedHorncrestCantAttackAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	horn := effectstest.B12Push(g, me.ID, "Bonded Horncrest", "Creature — Dinosaur", xlnB1HorncrestOracle, 5, 5)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(horn, opp.ID); err != nil {
		t.Fatalf("staging the Horncrest: %v", err)
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrAttackLimit) {
		t.Fatalf("passing with the Horncrest alone: %v, want ErrAttackLimit", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("the companion was refused: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, horn); c == nil || c.AttackingTarget == uuid.Nil {
		t.Error("the Horncrest is not attacking beside the Bear")
	}
}

func TestXlnB1SpikeTailedCeratopsBlocksTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ceratops := effectstest.B12Push(g, g.Seats[1].ID, "Spike-Tailed Ceratops", "Creature — Dinosaur", xlnB1CeratopsOracle, 4, 4)
	if got := effectstest.MbCapacity(t, g, ceratops); got != 2 {
		t.Errorf("capacity = %d, want 2", got)
	}
}

func TestXlnB1DeeprootChampionGrowsOnNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	champ := effectstest.B12Push(g, me.ID, "Deeproot Champion", "Creature — Merfolk Shaman", xlnB1DeeprootOracle, 1, 1)
	effectstest.CastHoldCreature(t, g, "Raptor Companion", "Creature — Dinosaur", xlnB1RaptorCompanionOracle, 3, 1)
	if got := effectstest.CountersOn(g, champ, game.CounterPlusOne); got != 0 {
		t.Fatalf("a creature spell gave %d counters", got)
	}
	effectstest.CastCatalogSpell(t, g, "Ritual of Rejuvenation", "Instant", xlnB1RitualOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, champ, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters after a noncreature spell, want 1", got)
	}
}

func TestXlnB1EnrageTriggers(t *testing.T) {
	t.Run("Ravenous Daggertooth", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dino := effectstest.B12Push(g, me.ID, "Ravenous Daggertooth", "Creature — Dinosaur", xlnB1DaggertoothOracle, 3, 2)
		src := effectstest.B12Creature(g, opp.ID, "Pinger", xlnB1BearType, 1, 1)
		life := me.Life
		effectstest.Cmdr2B08Damage(g, src, dino, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life+2 {
			t.Errorf("life %d, want %d", me.Life, life+2)
		}
	})
	t.Run("Raptor Hatchling", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dino := effectstest.B12Push(g, me.ID, "Raptor Hatchling", "Creature — Dinosaur", xlnB1HatchlingOracle, 1, 3)
		src := effectstest.B12Creature(g, opp.ID, "Pinger", xlnB1BearType, 1, 1)
		effectstest.Cmdr2B08Damage(g, src, dino, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if n, _ := effectstest.CountTokensNamed(g, me.ID, "Dinosaur"); n != 1 {
			t.Fatalf("%d Dinosaur tokens, want 1", n)
		}
		for _, c := range effectstest.TokensNamed(g, "Dinosaur") {
			if c.Power != 3 || c.Toughness != 3 || !slices.Contains(c.Keywords, "trample") {
				t.Errorf("the token is %d/%d with %v, want a 3/3 with trample", c.Power, c.Toughness, c.Keywords)
			}
		}
	})
}

// --- auras ------------------------------------------------------------------------

func TestXlnB1CastawaysDespairTapsAndHoldsItsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Host", xlnB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Castaway's Despair", effectstest.AuraTypeLine, xlnB1CastawayOracle, effectstest.CardTarget(host))
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Fatal("the enters trigger did not tap the enchanted creature")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Error("the enchanted creature untapped during its controller's untap step")
	}
}

func TestXlnB1PumpAurasGiveTwoTwoAndAKeyword(t *testing.T) {
	for _, tc := range []struct{ name, oracle, keyword string }{
		{"One With the Wind", xlnB1OneWithTheWindOracle, "flying"},
		{"Swashbuckling", xlnB1SwashbucklingOracle, "haste"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			bear := effectstest.B12Creature(g, g.Seats[0].ID, "Bear", xlnB1BearType, 2, 2)
			effectstest.CastCatalogSpell(t, g, tc.name, effectstest.AuraTypeLine, tc.oracle, effectstest.CardTarget(bear))
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
				t.Errorf("the enchanted creature is %d/%d, want 4/4", p, tough)
			}
			if !effectstest.HasKeywordOnBattlefield(t, g, bear, tc.keyword) {
				t.Errorf("the enchanted creature does not have %s", tc.keyword)
			}
		})
	}
}

func TestXlnB1PiousInterdictionGainsTwoAndPacifies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", xlnB1BearType, 2, 2)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Pious Interdiction", effectstest.AuraTypeLine, xlnB1PiousOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	effectstest.AssertRestrictions(t, g, bear, game.CantAttackOrBlock)
}

// --- spells ---------------------------------------------------------------------

func TestXlnB1UnfriendlyFireDealsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Unfriendly Fire", "Instant", xlnB1UnfriendlyFireOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("the player is at %d, want %d", opp.Life, life-4)
	}
}

func TestXlnB1DarkNourishmentDealsThreeAndGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Dark Nourishment", "Instant", xlnB1DarkNourishOracle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 3/3 survived 3 damage")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestXlnB1WalkThePlankSparesMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	merfolk := effectstest.B12Creature(g, opp.ID, "Merfolk", "Creature — Merfolk Wizard", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", xlnB1BearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Walk the Plank", "Sorcery", xlnB1WalkThePlankOracle, effectstest.CardTarget(merfolk)); err == nil {
		t.Fatal("Walk the Plank targeted a Merfolk")
	}
	effectstest.CastCatalogSpell(t, g, "Walk the Plank", "Sorcery", xlnB1WalkThePlankOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear was not destroyed")
	}
}

func TestXlnB1BrightReprisalDestroysAnAttackerAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacking Ogre", "Creature — Ogre", 3, 3)
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", xlnB1BearType, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	if _, err := effectstest.RprB19Cast(t, g, opp, "Bright Reprisal", "Instant", "{4}{W}", xlnB1BrightReprisalOracle, "{W}{W}{W}{W}{W}", game.CastSpellParams{Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("a creature that is not attacking was accepted as the target")
	}
	g.WithWriteLock(func() { opp.ManaPool = game.ManaPool{} })
	hand := opp.Hand.Size()
	if _, err := effectstest.RprB19Cast(t, g, opp, "Bright Reprisal", "Instant", "{4}{W}", xlnB1BrightReprisalOracle, "{W}{W}{W}{W}{W}", game.CastSpellParams{Targets: effectstest.CardTarget(attacker)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking creature was not destroyed")
	}
	// The helper put the spell into the hand and cast it; the spell drew one.
	if got := opp.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestXlnB1VampireTokenSpells(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		want         int
	}{
		{"Call to the Feast", xlnB1CallToTheFeastOracle, 3},
		{"Queen's Commission", xlnB1QueensCommissionOrcl, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.CastCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if n, _ := effectstest.CountTokensNamed(g, me.ID, "Vampire"); n != tc.want {
				t.Fatalf("%d Vampire tokens, want %d", n, tc.want)
			}
			for _, c := range effectstest.TokensNamed(g, "Vampire") {
				if c.Power != 1 || c.Toughness != 1 || !slices.Contains(c.Keywords, "lifelink") {
					t.Errorf("a token is %d/%d with %v, want a 1/1 with lifelink", c.Power, c.Toughness, c.Keywords)
				}
			}
		})
	}
}

func TestXlnB1PiratesPrizeDrawsTwoAndMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Pirate's Prize", "Sorcery", xlnB1PiratesPrizeOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
	if got := effectstest.TreasureCount(g); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
}

func TestXlnB1CrashTheRampartsPumpsAndGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[0].ID, "Bear", xlnB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Crash the Ramparts", "Instant", xlnB1CrashOracle, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the Bear is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the Bear has no trample")
	}
}

func TestXlnB1SpreadingRotDestroysALandAndItsControllerLosesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", xlnB1BearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Spreading Rot", "Sorcery", xlnB1SpreadingRotOracle, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("Spreading Rot targeted a creature")
	}
	mine, theirs := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Spreading Rot", "Sorcery", xlnB1SpreadingRotOracle, effectstest.CardTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
	if opp.Life != theirs-2 || me.Life != mine {
		t.Errorf("lives %d/%d, want %d/%d", me.Life, opp.Life, mine, theirs-2)
	}
}

func TestXlnB1FieryCannonadeSparesPirates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pirate := effectstest.B12Creature(g, me.ID, "Pirate", "Creature — Human Pirate", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", xlnB1BearType, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Fiery Cannonade", "Instant", xlnB1CannonadeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the non-Pirate survived")
	}
	if !g.Battlefield.Contains(pirate) {
		t.Fatal("the Pirate died")
	}
	if got := effectstest.FindCardAnywhere(t, g, pirate).DamageMarked; got != 0 {
		t.Errorf("the Pirate has %d damage, want 0", got)
	}
}

func TestXlnB1RitualOfRejuvenationGainsFourAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Ritual of Rejuvenation", "Instant", xlnB1RitualOracle, nil)
	life, hand := me.Life, me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestXlnB1VraskasContemptExilesAndGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	victim := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	if err := effectstest.CastCatalogSpellErr(t, g, "Vraska's Contempt", "Instant", xlnB1ContemptOracle, effectstest.CardTarget(land)); err == nil {
		t.Fatal("Vraska's Contempt targeted a land")
	}
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Vraska's Contempt", "Instant", xlnB1ContemptOracle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || opp.Graveyard.Contains(victim) {
		t.Error("the creature was not exiled")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestXlnB1PounceFightsInTheRightSlots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", xlnB1BearType, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Pounce", "Instant", xlnB1PounceOracle, effectstest.CardRefs(theirs, mine)); err == nil {
		t.Fatal("Pounce accepted the targets in the wrong slots")
	}
	effectstest.CastCatalogSpell(t, g, "Pounce", "Instant", xlnB1PounceOracle, effectstest.CardRefs(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the 2/2 survived the fight")
	}
	if got := effectstest.FindCardAnywhere(t, g, mine).DamageMarked; got != 2 {
		t.Errorf("your creature has %d damage, want 2", got)
	}
}
