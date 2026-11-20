package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// con_5587_b2_cards_test.go — Conflux, batch 2 (#5587). Each test plays
// the real catalog entry through the engine and asserts what its text
// says. Shared helpers come from batch 1 (con_b1_cards_test.go) and the
// older suites; this batch's own are prefixed conB2.

const (
	conB2AerieMysticsOracle       = "12134f7d-433a-416a-b668-c1a21984c94b"
	conB2ArmillarySphereOracle    = "3963140c-da67-43e6-9514-fe9dc0a43c4d"
	conB2AshasFavorOracle         = "97e7ada6-dee1-4bb3-a42d-38a61d946bd4"
	conB2BeaconBehemothOracle     = "25571850-d5c6-460f-b212-9fc9aa705277"
	conB2BloodhallOozeOracle      = "79591c18-0065-49bb-a77a-734d4af87a3c"
	conB2CharnelhoardWurmOracle   = "4a430fa3-e693-424b-9981-d7d8193445e3"
	conB2ChildOfAlaraOracle       = "70dbe8a9-505d-41c2-9b5b-a991d13ab459"
	conB2CumberStoneOracle        = "49199997-65cd-4cb9-bdd6-74c93c1f6b5a"
	conB2DarkTemperOracle         = "3a0ea41e-b4f5-4aa1-9f7a-158485313c64"
	conB2DreadwingOracle          = "446b9626-2e64-4933-afc8-891135b8e3d6"
	conB2ElderMasteryOracle       = "f99f44b4-a88f-403b-bbc8-bf1a45557124"
	conB2FaerieMechanistOracle    = "fcbe14c4-1b0d-486f-bb9c-0fed452668cb"
	conB2GleamOfResistanceOracle  = "c4c0a391-4549-4041-bc6d-09f8b756868e"
	conB2GrixisSlavedriverOracle  = "9087928a-4732-4661-b3ed-d0b1ea06edf6"
	conB2HellkiteHatchlingOracle  = "17b1094f-be67-465c-a91e-4010113bb352"
	conB2JhessianBalmgiverOracle  = "2be283da-6d5a-442b-bd07-d06c16cb1214"
	conB2KaleidostoneOracle       = "498ec48f-cb28-48d4-96f4-9d60bb4bb45c"
	conB2KnotvineMysticOracle     = "5d045695-e4b3-4daa-bfab-d77640fedfd6"
	conB2LapseOfCertaintyOracle   = "1b721ad3-d0f6-4eec-9bf0-f57ea9dfa392"
	conB2MagisterSphinxOracle     = "1d86ae29-4705-4bbc-8922-9b4e031df5b1"
	conB2ManaforceMaceOracle      = "7e92be29-ccd1-4556-bb2a-ac7902bc77c2"
	conB2MarkOfAsylumOracle       = "d4327e06-ee89-459c-a43a-dc064e198fef"
	conB2MirrorSigilOracle        = "c021c808-dc58-4386-8758-68b858badbd5"
	conB2NacatlHuntPrideOracle    = "094bc780-3c37-4863-b6ed-e2d16ed94544"
	conB2ObeliskOfAlaraOracle     = "75fcb9ce-ac01-4048-82b7-9100143e3b07"
	conB2PaleolothOracle          = "627de3d8-ad5e-4f9b-b83c-476cc98fea9b"
	conB2ParasiticStrixOracle     = "975b4e0f-1133-4ffa-a6b1-d319d248dcde"
	conB2RottingRatsOracle        = "1f361ae9-98a2-40ff-a579-fa5b9f510df4"
	conB2RuptureSpireOracle       = "7eadffcb-1e15-44c1-b1db-78c71b8ec1ce"
	conB2SacellumArchersOracle    = "3b9117a1-cd2d-4478-beda-1049e0b0f8eb"
	conB2ScarlandThrinaxOracle    = "1a185f90-69eb-496b-8e38-8dcf241a84a4"
	conB2ScattershotArcherOracle  = "a38ef622-d793-4867-a818-5c27982608b7"
	conB2ScepterOfDominanceOracle = "60bf80f9-78b1-4869-afd9-e769b1b0f5ea"
	conB2ScepterOfFugueOracle     = "f0b1dcdc-4fbb-4bcb-a4b0-1aa5928347f9"
	conB2ScornfulLichOracle       = "f8834831-5571-466d-9227-bb91ca2dce3f"
	conB2SedraxisAlchemistOracle  = "56f3f1b4-c36a-41f3-95a1-45e9d7e74cf9"
	conB2SphinxSummonerOracle     = "77ae7caa-e180-4efb-9348-ab389364b186"
	conB2VagrantPlowbeastsOracle  = "e97870a3-855a-4701-8344-91a90095f70a"
	conB2VectisAgentsOracle       = "fd9bb758-a763-46bf-9e56-4be26e271741"
	conB2ViashinoOracle           = "34bdf2f1-0cc9-4588-9a3e-5cfe5cde815d"
	conB2WretchedBanquetOracle    = "dafd4385-e64f-4a4a-93c4-b0ac6bd30d41"
	conB2YokeOfTheDamnedOracle    = "af08dae2-ba32-4a34-b8c1-b49d45ce0b6b"
)

func TestConB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		conB2AerieMysticsOracle:       "Aerie Mystics",
		conB2ArmillarySphereOracle:    "Armillary Sphere",
		conB2AshasFavorOracle:         "Asha's Favor",
		conB2BeaconBehemothOracle:     "Beacon Behemoth",
		conB2BloodhallOozeOracle:      "Bloodhall Ooze",
		conB2CharnelhoardWurmOracle:   "Charnelhoard Wurm",
		conB2ChildOfAlaraOracle:       "Child of Alara",
		conB2CumberStoneOracle:        "Cumber Stone",
		conB2DarkTemperOracle:         "Dark Temper",
		conB2DreadwingOracle:          "Dreadwing",
		conB2ElderMasteryOracle:       "Elder Mastery",
		conB2FaerieMechanistOracle:    "Faerie Mechanist",
		conB2GleamOfResistanceOracle:  "Gleam of Resistance",
		conB2GrixisSlavedriverOracle:  "Grixis Slavedriver",
		conB2HellkiteHatchlingOracle:  "Hellkite Hatchling",
		conB2JhessianBalmgiverOracle:  "Jhessian Balmgiver",
		conB2KaleidostoneOracle:       "Kaleidostone",
		conB2KnotvineMysticOracle:     "Knotvine Mystic",
		conB2LapseOfCertaintyOracle:   "Lapse of Certainty",
		conB2MagisterSphinxOracle:     "Magister Sphinx",
		conB2ManaforceMaceOracle:      "Manaforce Mace",
		conB2MarkOfAsylumOracle:       "Mark of Asylum",
		conB2MirrorSigilOracle:        "Mirror-Sigil Sergeant",
		conB2NacatlHuntPrideOracle:    "Nacatl Hunt-Pride",
		conB2ObeliskOfAlaraOracle:     "Obelisk of Alara",
		conB2PaleolothOracle:          "Paleoloth",
		conB2ParasiticStrixOracle:     "Parasitic Strix",
		conB2RottingRatsOracle:        "Rotting Rats",
		conB2RuptureSpireOracle:       "Rupture Spire",
		conB2SacellumArchersOracle:    "Sacellum Archers",
		conB2ScarlandThrinaxOracle:    "Scarland Thrinax",
		conB2ScattershotArcherOracle:  "Scattershot Archer",
		conB2ScepterOfDominanceOracle: "Scepter of Dominance",
		conB2ScepterOfFugueOracle:     "Scepter of Fugue",
		conB2ScornfulLichOracle:       "Scornful Aether-Lich",
		conB2SedraxisAlchemistOracle:  "Sedraxis Alchemist",
		conB2SphinxSummonerOracle:     "Sphinx Summoner",
		conB2VagrantPlowbeastsOracle:  "Vagrant Plowbeasts",
		conB2VectisAgentsOracle:       "Vectis Agents",
		conB2ViashinoOracle:           "Viashino Slaughtermaster",
		conB2WretchedBanquetOracle:    "Wretched Banquet",
		conB2YokeOfTheDamnedOracle:    "Yoke of the Damned",
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

// conB2SettleWithPick passes priority until the stack is empty, answering
// `chooser`'s target picks with `ref` and saying `may` to their "you
// may" prompts on the way. Any other prompt stops it, for the test to
// answer.
func conB2SettleWithPick(t *testing.T, g *game.Game, chooser uuid.UUID, ref game.TargetRef, may bool) {
	t.Helper()
	for i := 0; i < 48; i++ {
		if p := latestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, ref); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if c := latestTriggerPrompt(g, chooser); c != nil {
			if err := g.ResolveTriggerPrompt(c.ID, chooser, may); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			continue
		}
		if len(g.PendingTriggers) > 0 {
			g.RunStateChecksForTest()
			if latestPickTarget(g, chooser) != nil || latestTriggerPrompt(g, chooser) != nil {
				continue
			}
		}
		if len(g.PendingChoices) > 0 || stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				continue
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack did not settle")
}

// conB2Main is a fresh table at the active seat's main phase, with that
// seat and the next.
func conB2Main(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := conB1Seats(g)
	return g, me, opp
}

// conB2Graveyard puts a card in a player's graveyard.
func conB2Graveyard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

func TestConB2KeywordCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Aerie Mystics", "Creature — Bird Wizard", conB2AerieMysticsOracle, []string{"flying"}},
		{"Charnelhoard Wurm", "Creature — Wurm", conB2CharnelhoardWurmOracle, []string{"trample"}},
		{"Child of Alara", "Legendary Creature — Avatar", conB2ChildOfAlaraOracle, []string{"trample"}},
		{"Faerie Mechanist", "Artifact Creature — Faerie Artificer", conB2FaerieMechanistOracle, []string{"flying"}},
		{"Magister Sphinx", "Artifact Creature — Sphinx", conB2MagisterSphinxOracle, []string{"flying"}},
		{"Mirror-Sigil Sergeant", "Creature — Rhino Soldier", conB2MirrorSigilOracle, []string{"trample"}},
		{"Nacatl Hunt-Pride", "Creature — Cat Warrior", conB2NacatlHuntPrideOracle, []string{"vigilance"}},
		{"Parasitic Strix", "Artifact Creature — Bird", conB2ParasiticStrixOracle, []string{"flying"}},
		{"Sphinx Summoner", "Artifact Creature — Sphinx", conB2SphinxSummonerOracle, []string{"flying"}},
		{"Viashino Slaughtermaster", "Creature — Lizard Warrior", conB2ViashinoOracle, []string{"double strike"}},
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

// --- statics -------------------------------------------------------------------

func TestConB2CumberStoneShrinksOnlyOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Cumber Stone", "Artifact", conB2CumberStoneOracle, 0, 0)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if p, tough := ptOf(t, g, theirs); p != 1 || tough != 2 {
		t.Errorf("the opponent's Bear is %d/%d, want 1/2", p, tough)
	}
	if p, tough := ptOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("my Bear is %d/%d, want 2/2", p, tough)
	}
}

func TestConB2AshasFavorAndElderMasteryGrant(t *testing.T) {
	g := newCatalogGame(t)
	bear := seedBear(g, g.Seats[0].ID)
	auraCast(t, g, "Asha's Favor", conB2AshasFavorOracle, bear)
	for _, kw := range []string{"flying", "first strike", "vigilance"} {
		if !containsString(effectiveAbilities(t, g, bear), kw) {
			t.Errorf("Asha's Favor: the enchanted creature lacks %s", kw)
		}
	}

	g = newCatalogGame(t)
	bear = seedBear(g, g.Seats[0].ID)
	auraCast(t, g, "Elder Mastery", conB2ElderMasteryOracle, bear)
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("Elder Mastery: the Bear is %d/%d, want 5/5", p, tough)
	}
	if !containsString(effectiveAbilities(t, g, bear), "flying") {
		t.Error("Elder Mastery: the enchanted creature lacks flying")
	}
}

func TestConB2ElderMasteryMakesTheDamagedPlayerDiscardTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedBear(g, me.ID)
	auraCast(t, g, "Elder Mastery", conB2ElderMasteryOracle, bear)
	d2dSetHand(opp, 3)
	emitNoncombatDamage(g, bear, opp.ID, 1)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID)
	if opp.Hand.Size() != 1 {
		t.Errorf("the damaged player has %d cards, want 1", opp.Hand.Size())
	}
}

func TestConB2ManaforceMaceCountsBasicLandTypes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	domainLands(g, me, opp, "Forest", "Island")
	mace := b12Push(g, me.ID, "Manaforce Mace", "Artifact — Equipment", conB2ManaforceMaceOracle, 0, 0)
	bear := seedBear(g, me.ID)
	equipTo(t, g, me.ID, mace, bear)
	if p, tough := ptOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the equipped Bear is %d/%d, want 5/5 (three basic land types)", p, tough)
	}
}

func TestConB2HellkiteHatchlingHasKeywordsOnlyIfItDevoured(t *testing.T) {
	g := newCatalogGame(t)
	id := b12Push(g, g.Seats[0].ID, "Hellkite Hatchling", "Creature — Dragon", conB2HellkiteHatchlingOracle, 2, 2)
	if hasKeywordOnBattlefield(t, g, id, "flying") || hasKeywordOnBattlefield(t, g, id, "trample") {
		t.Error("a Hatchling that devoured nothing has flying or trample")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Devoured = 1
			}
		}
		g.BumpLayerVersionForTest()
	})
	for _, kw := range []string{"flying", "trample"} {
		if !hasKeywordOnBattlefield(t, g, id, kw) {
			t.Errorf("a Hatchling that devoured a creature lacks %s", kw)
		}
	}
}

func TestConB2MarkOfAsylumPreventsNoncombatDamageToYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := conB1Seats(g)
	b12Push(g, me.ID, "Mark of Asylum", "Enchantment", conB2MarkOfAsylumOracle, 0, 0)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetCard(mine))
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if !opp.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt did not resolve")
	}
	if !g.Battlefield.Contains(mine) || damageMarkedOn(g, mine) != 0 {
		t.Errorf("my creature took %d noncombat damage", damageMarkedOn(g, mine))
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the opponent's creature should be untouched")
	}
}

// --- activated abilities ----------------------------------------------------------

func TestConB2AerieMysticsGivesYourCreaturesShroud(t *testing.T) {
	g, me, opp := conB2Main(t)
	mystics := b12Push(g, me.ID, "Aerie Mystics", "Creature — Bird Wizard", conB2AerieMysticsOracle, 3, 3)
	bear := seedBear(g, me.ID)
	theirs := seedBear(g, opp.ID)
	if err := conB1Activate(t, g, me, mystics, 0, "GUC", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "shroud") || !hasEffectiveKeyword(t, g, mystics, "shroud") {
		t.Error("a creature I control lacks shroud")
	}
	if hasEffectiveKeyword(t, g, theirs, "shroud") {
		t.Error("an opponent's creature gained shroud")
	}
}

func TestConB2ArmillarySphereFetchesTwoBasics(t *testing.T) {
	g, me, _ := conB2Main(t)
	sphere := b12Push(g, me.ID, "Armillary Sphere", "Artifact", conB2ArmillarySphereOracle, 0, 0)
	seedSearchLibrary(me, searchTestLand("Forest", "Basic Land — Forest"), searchTestLand("Island", "Basic Land — Island"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	if err := conB1Activate(t, g, me, sphere, 0, "CC", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(sphere) {
		t.Fatal("the Sphere was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerSearchNamed(t, g, me.ID, "Forest", "Island")
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want both basics", hand, me.Hand.Size())
	}
}

func TestConB2BeaconBehemothTargetsOnlyPowerFive(t *testing.T) {
	g, me, _ := conB2Main(t)
	behemoth := b12Push(g, me.ID, "Beacon Behemoth", "Creature — Beast", conB2BeaconBehemothOracle, 5, 3)
	bear := seedBear(g, me.ID)
	if err := conB1Activate(t, g, me, behemoth, 0, "C", game.ActivateAbilityParams{Targets: cardTarget(bear)}); err == nil {
		t.Fatal("a 2/2 was a legal target")
	}
	if err := conB1Activate(t, g, me, behemoth, 0, "C", game.ActivateAbilityParams{Targets: cardTarget(behemoth)}); err != nil {
		t.Fatalf("activate on itself: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, behemoth, "vigilance") {
		t.Error("the Behemoth did not gain vigilance")
	}
}

func TestConB2VagrantPlowbeastsRegeneratesAPowerFiveCreature(t *testing.T) {
	g, me, _ := conB2Main(t)
	beasts := b12Push(g, me.ID, "Vagrant Plowbeasts", "Creature — Beast", conB2VagrantPlowbeastsOracle, 6, 6)
	bear := seedBear(g, me.ID)
	if err := conB1Activate(t, g, me, beasts, 0, "C", game.ActivateAbilityParams{Targets: cardTarget(bear)}); err == nil {
		t.Fatal("a 2/2 was a legal target")
	}
	if err := conB1Activate(t, g, me, beasts, 0, "C", game.ActivateAbilityParams{Targets: cardTarget(beasts)}); err != nil {
		t.Fatalf("activate on itself: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(beasts) })
	if !g.Battlefield.Contains(beasts) {
		t.Error("the regenerated Plowbeasts was destroyed")
	}
}

func TestConB2DreadwingAndScornfulAetherLichPumpThemselves(t *testing.T) {
	g, me, _ := conB2Main(t)
	wing := b12Push(g, me.ID, "Dreadwing", "Creature — Zombie", conB2DreadwingOracle, 1, 1)
	if err := conB1Activate(t, g, me, wing, 0, "CUR", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Dreadwing: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, wing); p != 4 || tough != 1 {
		t.Errorf("Dreadwing is %d/%d, want 4/1", p, tough)
	}
	if !hasEffectiveKeyword(t, g, wing, "flying") {
		t.Error("Dreadwing lacks flying")
	}

	lich := b12Push(g, me.ID, "Scornful Aether-Lich", "Artifact Creature — Zombie Wizard", conB2ScornfulLichOracle, 2, 4)
	if err := conB1Activate(t, g, me, lich, 0, "WB", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the Lich: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"fear", "vigilance"} {
		if !hasEffectiveKeyword(t, g, lich, kw) {
			t.Errorf("the Lich lacks %s", kw)
		}
	}
	if p, tough := ptOf(t, g, lich); p != 2 || tough != 4 {
		t.Errorf("the Lich is %d/%d, want an unchanged 2/4", p, tough)
	}
}

func TestConB2VectisAgentsShrinksAndCantBeBlocked(t *testing.T) {
	g, me, _ := conB2Main(t)
	agents := b12Push(g, me.ID, "Vectis Agents", "Artifact Creature — Human Rogue", conB2VectisAgentsOracle, 4, 3)
	if err := conB1Activate(t, g, me, agents, 0, "UB", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, agents); p != 2 || tough != 3 {
		t.Errorf("the Agents are %d/%d, want 2/3", p, tough)
	}
	assertRestrictions(t, g, agents, game.CantBeBlocked)
}

func TestConB2ViashinoSlaughtermasterPumpsOnceATurn(t *testing.T) {
	g, me, _ := conB2Main(t)
	lizard := b12Push(g, me.ID, "Viashino Slaughtermaster", "Creature — Lizard Warrior", conB2ViashinoOracle, 1, 1)
	if err := conB1Activate(t, g, me, lizard, 0, "BG", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, lizard); p != 2 || tough != 2 {
		t.Errorf("the Slaughtermaster is %d/%d, want 2/2", p, tough)
	}
	if err := conB1Activate(t, g, me, lizard, 0, "BG", game.ActivateAbilityParams{}); err == nil {
		t.Error("activated a second time in one turn")
	}
}

func TestConB2ScarlandThrinaxGrowsBySacrificing(t *testing.T) {
	g, me, _ := conB2Main(t)
	thrinax := b12Push(g, me.ID, "Scarland Thrinax", "Creature — Lizard", conB2ScarlandThrinaxOracle, 2, 2)
	bear := seedBear(g, me.ID)
	if err := conB1Activate(t, g, me, thrinax, 0, "", game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, thrinax); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
}

func TestConB2SacellumArchersShootsOnlyACombatant(t *testing.T) {
	g, me, opp := conB2Main(t)
	archers := b12Push(g, me.ID, "Sacellum Archers", "Creature — Elf Archer", conB2SacellumArchersOracle, 2, 3)
	idle := seedBear(g, opp.ID)
	if err := conB1Activate(t, g, me, archers, 0, "RW", game.ActivateAbilityParams{Targets: cardTarget(idle)}); err == nil {
		t.Error("a creature not in combat was a legal target")
	}
}

func TestConB2ScattershotArcherHitsEachFlyer(t *testing.T) {
	g, me, opp := conB2Main(t)
	archer := b12Push(g, me.ID, "Scattershot Archer", "Creature — Elf Archer", conB2ScattershotArcherOracle, 1, 2)
	flyer := b12Push(g, opp.ID, "Esper Cormorants", "Artifact Creature — Bird", conB1EsperCormorantsOracle, 3, 3)
	walker := seedBear(g, opp.ID)
	if err := conB1Activate(t, g, me, archer, 0, "", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, flyer); got != 1 {
		t.Errorf("the flyer has %d damage, want 1", got)
	}
	if got := damageMarkedOn(g, walker); got != 0 {
		t.Errorf("the ground creature has %d damage, want 0", got)
	}
}

func TestConB2ScepterOfDominanceTapsAPermanent(t *testing.T) {
	g, me, opp := conB2Main(t)
	scepter := b12Push(g, me.ID, "Scepter of Dominance", "Artifact", conB2ScepterOfDominanceOracle, 0, 0)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := conB1Activate(t, g, me, scepter, 0, "W", game.ActivateAbilityParams{Targets: cardTarget(rock)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !isTapped(g, rock) {
		t.Error("the target permanent was not tapped")
	}
}

func TestConB2ScepterOfFugueDiscardsOnlyOnYourTurn(t *testing.T) {
	g, me, opp := conB2Main(t)
	scepter := b12Push(g, me.ID, "Scepter of Fugue", "Artifact", conB2ScepterOfFugueOracle, 0, 0)
	d2dSetHand(opp, 2)
	if err := conB1Activate(t, g, me, scepter, 0, "CB", game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}}); err != nil {
		t.Fatalf("activate on my turn: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID)
	if opp.Hand.Size() != 1 {
		t.Errorf("the target has %d cards, want 1", opp.Hand.Size())
	}

	g = newCatalogGame(t)
	advanceToMain(t, g)
	active, other := conB1Seats(g)
	scepter = b12Push(g, other.ID, "Scepter of Fugue", "Artifact", conB2ScepterOfFugueOracle, 0, 0)
	if err := conB1Activate(t, g, other, scepter, 0, "CB", game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(active.ID)}}); err == nil {
		t.Error("activated on an opponent's turn")
	}
}

func TestConB2JhessianBalmgiverMakesACreatureUnblockable(t *testing.T) {
	g, me, _ := conB2Main(t)
	cleric := b12Push(g, me.ID, "Jhessian Balmgiver", "Creature — Human Cleric", conB2JhessianBalmgiverOracle, 1, 1)
	bear := seedBear(g, me.ID)
	if err := conB1Activate(t, g, me, cleric, 1, "", game.ActivateAbilityParams{Targets: cardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, bear, game.CantBeBlocked)
	if err := conB1Activate(t, g, me, cleric, 0, "", game.ActivateAbilityParams{Targets: cardTarget(bear)}); err == nil {
		t.Error("the tapped Balmgiver activated its other {T} row")
	}
}

func TestConB2JhessianBalmgiverPreventsOneDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := conB1Seats(g)
	cleric := b12Push(g, me.ID, "Jhessian Balmgiver", "Creature — Human Cleric", conB2JhessianBalmgiverOracle, 1, 1)
	bear := b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	if err := conB1Activate(t, g, me, cleric, 0, "", game.ActivateAbilityParams{Targets: cardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if got := damageMarkedOn(g, bear); got != 2 {
		t.Errorf("the Ogre has %d damage, want 2 (one prevented)", got)
	}
}

func TestConB2NacatlHuntPrideRows(t *testing.T) {
	g, me, opp := conB2Main(t)
	cat := b12Push(g, me.ID, "Nacatl Hunt-Pride", "Creature — Cat Warrior", conB2NacatlHuntPrideOracle, 5, 4)
	theirs := seedBear(g, opp.ID)
	if err := conB1Activate(t, g, me, cat, 0, "R", game.ActivateAbilityParams{Targets: cardTarget(theirs)}); err != nil {
		t.Fatalf("activate the can't-block row: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertRestrictions(t, g, theirs, game.CantBlock)

	g, me, opp = conB2Main(t)
	cat = b12Push(g, me.ID, "Nacatl Hunt-Pride", "Creature — Cat Warrior", conB2NacatlHuntPrideOracle, 5, 4)
	theirs = seedBear(g, opp.ID)
	if err := conB1Activate(t, g, me, cat, 1, "G", game.ActivateAbilityParams{Targets: cardTarget(theirs)}); err != nil {
		t.Fatalf("activate the must-block row: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !b13Tapped(t, g, cat) {
		t.Error("the Hunt-Pride did not tap")
	}
}

func TestConB2ObeliskOfAlaraRows(t *testing.T) {
	g, me, _ := conB2Main(t)
	obelisk := b12Push(g, me.ID, "Obelisk of Alara", "Artifact", conB2ObeliskOfAlaraOracle, 0, 0)
	life := me.Life
	if err := conB1Activate(t, g, me, obelisk, 0, "CW", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate {1}{W}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life %d -> %d, want +5", life, me.Life)
	}

	g, me, opp := conB2Main(t)
	obelisk = b12Push(g, me.ID, "Obelisk of Alara", "Artifact", conB2ObeliskOfAlaraOracle, 0, 0)
	theirLife := opp.Life
	if err := conB1Activate(t, g, me, obelisk, 3, "CR", game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}}); err != nil {
		t.Fatalf("activate {1}{R}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != theirLife-3 {
		t.Errorf("opponent life %d -> %d, want -3", theirLife, opp.Life)
	}

	g, me, opp = conB2Main(t)
	obelisk = b12Push(g, me.ID, "Obelisk of Alara", "Artifact", conB2ObeliskOfAlaraOracle, 0, 0)
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	if err := conB1Activate(t, g, me, obelisk, 2, "CB", game.ActivateAbilityParams{Targets: cardTarget(theirs)}); err != nil {
		t.Fatalf("activate {1}{B}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("the Ogre is %d/%d, want 1/1", p, tough)
	}
	if err := conB1Activate(t, g, me, obelisk, 4, "CG", game.ActivateAbilityParams{Targets: cardTarget(theirs)}); err == nil {
		t.Error("the tapped Obelisk activated a second row")
	}

	g, me, _ = conB2Main(t)
	obelisk = b12Push(g, me.ID, "Obelisk of Alara", "Artifact", conB2ObeliskOfAlaraOracle, 0, 0)
	bear := seedBear(g, me.ID)
	if err := conB1Activate(t, g, me, obelisk, 4, "CG", game.ActivateAbilityParams{Targets: cardTarget(bear)}); err != nil {
		t.Fatalf("activate {1}{G}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("the Bear is %d/%d, want 6/6", p, tough)
	}

	g, me, _ = conB2Main(t)
	obelisk = b12Push(g, me.ID, "Obelisk of Alara", "Artifact", conB2ObeliskOfAlaraOracle, 0, 0)
	seedLibrary(me, "A", "B")
	fodder := handCard(me, "Fodder", "Instant")
	if err := conB1Activate(t, g, me, obelisk, 1, "CU", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate {1}{U}: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, me.ID, fodder)
	if !me.Graveyard.Contains(fodder) {
		t.Error("the loot did not discard")
	}
}

// --- mana abilities ------------------------------------------------------------

func conB2PoolCount(p *game.Player) map[string]int {
	got := map[string]int{}
	for _, c := range poolColors(p) {
		got[c]++
	}
	return got
}

func TestConB2KnotvineMysticAddsNayaMana(t *testing.T) {
	g, me, _ := conB2Main(t)
	mystic := b12Push(g, me.ID, "Knotvine Mystic", "Creature — Elf Druid", conB2KnotvineMysticOracle, 2, 2)
	floatForTest(g, me, "C")
	if err := g.ActivateManaAbility(me.ID, mystic, 0, game.ManaAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	got := conB2PoolCount(me)
	if got["R"] != 1 || got["G"] != 1 || got["W"] != 1 || len(poolColors(me)) != 3 {
		t.Errorf("pool %v, want exactly {R}{G}{W}", poolColors(me))
	}
}

func TestConB2KaleidostoneDrawsAndSacrificesForFiveColours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	stone := castCatalogSpell(t, g, "Kaleidostone", "Artifact", conB2KaleidostoneOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want a draw", hand, me.Hand.Size())
	}
	floatForTest(g, me, "CCCCC")
	if err := g.ActivateManaAbility(me.ID, stone, 0, game.ManaAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	got := conB2PoolCount(me)
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		if got[c] != 1 {
			t.Errorf("pool %v, want one %s", poolColors(me), c)
		}
	}
	if g.Battlefield.Contains(stone) || !me.Graveyard.Contains(stone) {
		t.Error("the Kaleidostone was not sacrificed")
	}
}

func TestConB2RuptureSpireIsPaidForOrSacrificed(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		land := playLandFromHand(t, g, "Rupture Spire", conB2RuptureSpireOracle)
		passPriorityAroundTable(t, g)
		if c, ok := battlefieldCard(g, land); !ok || !c.Tapped {
			t.Fatal("the Spire did not enter tapped")
		}
		if pay {
			dcMana(t, g, me.ID, "{C}")
		}
		answerPayUnless(t, g, me.ID, pay)
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(land) != pay {
			t.Errorf("paid=%v: the Spire on the battlefield = %v", pay, g.Battlefield.Contains(land))
		}
	}
}

// --- triggers ------------------------------------------------------------------

func TestConB2ChildOfAlaraDestroysAllNonlandPermanentsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := conB1Seats(g)
	child := b12Push(g, me.ID, "Child of Alara", "Legendary Creature — Avatar", conB2ChildOfAlaraOracle, 6, 6)
	bear := seedBear(g, opp.ID)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(child) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(rock) {
		t.Error("a nonland permanent survived")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land was destroyed")
	}
}

func TestConB2GrixisSlavedriverLeavesAZombie(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := conB1Seats(g)
	giant := b12Push(g, me.ID, "Grixis Slavedriver", "Creature — Zombie Giant", conB2GrixisSlavedriverOracle, 4, 4)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(giant) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n := b27CountNamed(g, me.ID, "Zombie"); n != 1 {
		t.Errorf("%d Zombie tokens, want 1", n)
	}

	g = newCatalogGame(t)
	id, _ := activateFromGraveyard(t, g, "Grixis Slavedriver", "Creature — Zombie Giant",
		conB2GrixisSlavedriverOracle, 4, 4, "{C}{C}{C}{B}", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) || !hasKeywordOnBattlefield(t, g, id, "haste") {
		t.Error("the unearthed Slavedriver is not on the battlefield with haste")
	}
}

func TestConB2RottingRatsMakesEachPlayerDiscard(t *testing.T) {
	g := newCatalogGame(t)
	for _, p := range g.Seats {
		d2dSetHand(p, 2)
	}
	castCatalogSpell(t, g, "Rotting Rats", "Creature — Zombie Rat", conB2RottingRatsOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if discardChoiceFor(g, p.ID) == nil {
			t.Errorf("seat %s was not asked to discard", p.ID)
			continue
		}
		answerDiscard(t, g, p.ID, p.Hand.Cards[0].InstanceID)
	}
}

func TestConB2MagisterSphinxSetsALifeTotalToTen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := conB1Seats(g)
	castCatalogSpell(t, g, "Magister Sphinx", "Artifact Creature — Sphinx", conB2MagisterSphinxOracle, nil)
	conB2SettleWithPick(t, g, me.ID, playerTarget(opp.ID), true)
	if opp.Life != 10 {
		t.Errorf("the target's life is %d, want 10", opp.Life)
	}
}

func TestConB2SedraxisAlchemistBouncesOnlyWithABluePermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := conB1Seats(g)
	tmpB1Push(g, me.ID, "Blue Rock", "Artifact", "", 0, 0, "U")
	theirs := seedBear(g, opp.ID)
	castCatalogSpell(t, g, "Sedraxis Alchemist", "Creature — Zombie Wizard", conB2SedraxisAlchemistOracle, nil)
	conB2SettleWithPick(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}, true)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the target was not returned to its owner's hand")
	}

	g = newCatalogGame(t)
	me, opp = conB1Seats(g)
	theirs = seedBear(g, opp.ID)
	castCatalogSpell(t, g, "Sedraxis Alchemist", "Creature — Zombie Wizard", conB2SedraxisAlchemistOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("the trigger asked for a target with no blue permanent")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("a permanent was bounced with no blue permanent")
	}
}

func TestConB2ParasiticStrixDrainsWithABlackPermanent(t *testing.T) {
	for _, black := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := conB1Seats(g)
		if black {
			tmpB1Push(g, me.ID, "Black Rock", "Artifact", "", 0, 0, "B")
		}
		myLife, theirLife := me.Life, opp.Life
		castCatalogSpell(t, g, "Parasitic Strix", "Artifact Creature — Bird", conB2ParasiticStrixOracle, nil)
		conB2SettleWithPick(t, g, me.ID, playerTarget(opp.ID), true)
		want := 0
		if black {
			want = 2
		}
		if opp.Life != theirLife-want || me.Life != myLife+want {
			t.Errorf("black=%v: opponent %d -> %d, me %d -> %d", black, theirLife, opp.Life, myLife, me.Life)
		}
	}
}

func TestConB2BloodhallOozeGrowsWithABlackPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ooze := b12Push(g, me.ID, "Bloodhall Ooze", "Creature — Ooze", conB2BloodhallOozeOracle, 1, 1)
	tmpB1Push(g, me.ID, "Black Rock", "Artifact", "", 0, 0, "B")
	walkToUpkeepOfSeat0(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, ooze); n < 1 {
		t.Errorf("%d +1/+1 counters, want at least 1", n)
	}

	g = newCatalogGame(t)
	me = g.Seats[0]
	ooze = b12Push(g, me.ID, "Bloodhall Ooze", "Creature — Ooze", conB2BloodhallOozeOracle, 1, 1)
	walkToUpkeepOfSeat0(t, g)
	passPriorityAroundTable(t, g)
	if latestTriggerPrompt(g, me.ID) != nil || plusOnes(g, ooze) != 0 {
		t.Error("the Ooze triggered with neither a black nor a green permanent")
	}
}

func TestConB2MirrorSigilSergeantCopiesItselfWithABluePermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Mirror-Sigil Sergeant", "Creature — Rhino Soldier", conB2MirrorSigilOracle, 4, 4)
	tmpB1Push(g, me.ID, "Blue Rock", "Artifact", "", 0, 0, "U")
	walkToUpkeepOfSeat0(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if n := b27CountNamed(g, me.ID, "Mirror-Sigil Sergeant"); n < 2 {
		t.Errorf("%d Sergeants, want the original and a token copy", n)
	}
}

func TestConB2CharnelhoardWurmReturnsACardWhenItDamagesAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := conB1Seats(g)
	wurm := b12Push(g, me.ID, "Charnelhoard Wurm", "Creature — Wurm", conB2CharnelhoardWurmOracle, 6, 6)
	dead := conB2Graveyard(me, "Dead Bear", "Creature — Bear")
	emitNoncombatDamage(g, wurm, opp.ID, 6)
	conB2SettleWithPick(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: dead}, true)
	if !me.Hand.Contains(dead) {
		t.Error("the graveyard card was not returned to hand")
	}
}

func TestConB2PaleolothReturnsACreatureCardWhenABigCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := conB1Seats(g)
	b12Push(g, me.ID, "Paleoloth", "Creature — Beast", conB2PaleolothOracle, 5, 5)
	dead := conB2Graveyard(me, "Dead Bear", "Creature — Bear")
	alaB1CastCreature(t, g, "Big Beast", 5, 5)
	conB2SettleWithPick(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: dead}, true)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card was not returned to hand")
	}

	g = newCatalogGame(t)
	me, _ = conB1Seats(g)
	b12Push(g, me.ID, "Paleoloth", "Creature — Beast", conB2PaleolothOracle, 5, 5)
	conB2Graveyard(me, "Dead Bear", "Creature — Bear")
	alaB1CastCreature(t, g, "Small Beast", 4, 4)
	if latestPickTarget(g, me.ID) != nil || latestTriggerPrompt(g, me.ID) != nil {
		t.Error("a 4-power creature triggered the Paleoloth")
	}
}

func TestConB2YokeOfTheDamnedDestroysItsCreatureWhenAnotherDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := conB1Seats(g)
	host := seedBear(g, opp.ID)
	auraCast(t, g, "Yoke of the Damned", conB2YokeOfTheDamnedOracle, host)
	victim := seedBear(g, me.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(host) {
		t.Error("the enchanted creature survived another creature's death")
	}
}

func TestConB2FaerieMechanistTakesAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	top := make([]uuid.UUID, 3)
	cards := []struct{ name, typeLine string }{
		{"Bauble", "Artifact"},
		{"Bear", "Creature — Bear"},
		{"Shock", "Instant"},
	}
	for i := len(cards) - 1; i >= 0; i-- {
		top[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: top[i], Name: cards[i].name, TypeLine: cards[i].typeLine, Owner: me.ID, Controller: me.ID})
	}
	castCatalogSpell(t, g, "Faerie Mechanist", "Artifact Creature — Faerie Artificer", conB2FaerieMechanistOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[0]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[0]) {
		t.Fatal("the artifact is not in hand")
	}
	if putInLibraryChoiceFor(g, me.ID) == nil {
		t.Error("no order prompt for the rest on the bottom")
	}
}

func TestConB2SphinxSummonerTutorsAnArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me, game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem"}, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	castCatalogSpell(t, g, "Sphinx Summoner", "Artifact Creature — Sphinx", conB2SphinxSummonerOracle, nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	hand := me.Hand.Size()
	answerSearchNamed(t, g, me.ID, "Golem")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want the Golem", hand, me.Hand.Size())
	}
}

// --- spells --------------------------------------------------------------------

func TestConB2DarkTemperDamagesOrDestroys(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := conB1Seats(g)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	castCatalogSpell(t, g, "Dark Temper", "Instant", conB2DarkTemperOracle, cardTarget(ogre))
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(ogre) || damageMarkedOn(g, ogre) != 2 {
		t.Errorf("without a black permanent the 3/3 should survive with 2 damage, marked %d", damageMarkedOn(g, ogre))
	}

	g = newCatalogGame(t)
	me, opp := conB1Seats(g)
	tmpB1Push(g, me.ID, "Black Rock", "Artifact", "", 0, 0, "B")
	ogre = b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	castCatalogSpell(t, g, "Dark Temper", "Instant", conB2DarkTemperOracle, cardTarget(ogre))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("with a black permanent the creature was not destroyed")
	}
}

func TestConB2WretchedBanquetNeedsTheLeastPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := conB1Seats(g)
	small := b12Creature(g, opp.ID, "Small", "Creature — Rat", 1, 1)
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 3, 3)
	b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Wretched Banquet", "Sorcery", conB2WretchedBanquetOracle, cardTarget(big))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Error("a creature without the least power was destroyed")
	}
	castCatalogSpell(t, g, "Wretched Banquet", "Sorcery", conB2WretchedBanquetOracle, cardTarget(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the creature with the least power survived")
	}
}

func TestConB2GleamOfResistancePumpsAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := conB1Seats(g)
	bear := seedBear(g, me.ID)
	tapForTest(t, g, bear)
	castCatalogSpell(t, g, "Gleam of Resistance", "Instant", conB2GleamOfResistanceOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 3/4", p, tough)
	}
	if isTapped(g, bear) {
		t.Error("the Bear was not untapped")
	}

	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me, searchTestLand("Plains", "Basic Land — Plains"))
	cycleFromHand(t, g, "Gleam of Resistance", "Instant", conB2GleamOfResistanceOracle, "{C}{W}")
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerSearchNamed(t, g, me.ID, "Plains")
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want the Plains", hand, me.Hand.Size())
	}
}

func TestConB2LapseOfCertaintyPutsTheSpellOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(opp, "Their Top")
	shock := castShockAt(t, g, opp, me.ID)
	castCounterAt(t, g, me, "Lapse of Certainty", conB2LapseOfCertaintyOracle, shock)
	passPriorityAroundTable(t, g)
	top, err := opp.Library.Top()
	if err != nil || top.InstanceID != shock {
		t.Error("the countered spell is not on top of its owner's library")
	}
	if opp.Graveyard.Contains(shock) {
		t.Error("the countered spell reached the graveyard")
	}
}
