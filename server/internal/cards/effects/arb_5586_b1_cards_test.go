package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// arb_5586_b1_cards_test.go — Alara Reborn, batch 1 (#5586). Every test
// plays the real catalog entry through the engine.

const (
	arbB1LeotauOracle       = "59edd0e7-37c2-45f9-9f9d-8002be459662"
	arbB1RhoxBruteOracle    = "c497c4f2-62c5-44fc-b485-f8d40868d8dc"
	arbB1YearlingOracle     = "6db09a0a-6c8e-4829-9b55-00745f0587b5"
	arbB1TwinclawsOracle    = "1fa3839a-127e-4740-bd2e-97b18fdddedc"
	arbB1SewnEyeOracle      = "994cf8bb-c900-448f-b055-ecc2ee6cf9e2"
	arbB1TalonOracle        = "518d0773-40fa-4b02-a89e-5711c19617cd"
	arbB1CoatlOracle        = "11369d8b-6443-43c3-a60f-dd920a4e8c08"
	arbB1SphinxOracle       = "1c96772b-463c-4655-836b-cf6e9dc6319f"
	arbB1EthercasteOracle   = "b34a7a9e-d740-4ddc-9b3c-8c49555ee2aa"
	arbB1BehemothOracle     = "277ff122-2d6e-441e-9ec1-980cdb393982"
	arbB1AbominationOracle  = "e1135564-ca93-4c5b-8686-cafdbaf0205e"
	arbB1JhessianOracle     = "40cc8ac6-114f-439d-816d-f65cd52bbbd2"
	arbB1RecluseOracle      = "488ca652-653b-4bb3-891c-874efdd418ff"
	arbB1PlowbeastOracle    = "ff3fae6e-d8a1-46b3-a6cc-fb4e8800ee68"
	arbB1RannetOracle       = "d0a3b261-3daf-4082-8bc5-e66326a6c382"
	arbB1PouncerOracle      = "ee6f1ddb-6749-4394-b2b0-4d293376f1af"
	arbB1SurebladeOracle    = "610adb2b-9d52-4b70-92d1-0c7adeb93552"
	arbB1StormbladeOracle   = "186b2fbd-8a59-4447-bd54-45a7c3ff5af2"
	arbB1GrimbladeOracle    = "04b1280a-b084-4f65-a3f4-2ba41d6c26fa"
	arbB1HackbladeOracle    = "e0e2184e-d4e3-4620-89c4-8378ce3a59c3"
	arbB1HushbladeOracle    = "d1f986dc-e0f0-4cc6-9dc4-fe97608da332"
	arbB1FalconsOracle      = "01fda82a-227d-4e3d-8481-a13c3812392a"
	arbB1RenegadesOracle    = "78fa248d-7f45-4d43-b3e3-c726db938b0f"
	arbB1ArmorguardOracle   = "9738f9b7-f281-44f6-a746-9c83380b35b7"
	arbB1GnawersOracle      = "e82308d5-29ee-48eb-9d1c-f24897ef191c"
	arbB1FiligreeOracle     = "4a261c42-825d-4965-8ea7-80d50c43474e"
	arbB1CyclopsOracle      = "f45a2a66-04cc-4622-9c70-f2dfc1391bb1"
	arbB1MagefireOracle     = "ea8c4a7d-f7fb-4620-b3a4-ae8280dd60d9"
	arbB1SangriteOracle     = "03220765-8a9d-4d4a-88d5-98df975ce13c"
	arbB1BreathOracle       = "81e6ade9-a38f-4655-b09c-9ca41a11face"
	arbB1OfferingOracle     = "d061972e-6f4f-41a7-966a-88839f105f19"
	arbB1ZealousOracle      = "ccf8d919-1399-4bef-94ac-04c67f122828"
	arbB1FlurryOracle       = "259b3d11-02e2-43c6-9eed-4ffabb6f9fcf"
	arbB1RebornHopeOracle   = "683d8deb-2ab4-4c65-9a23-f4e733f06ba2"
	arbB1ColossalOracle     = "1d8e39c9-e04d-4328-af66-db6e1241ac06"
	arbB1SunlightOracle     = "937cb494-7dcb-4a40-b1e0-6e6b31f5c8a6"
	arbB1DreadOracle        = "b090f1e5-1f7e-4492-9928-c5911a8ef95b"
	arbB1DenyRealityOracle  = "f9e31993-bfe1-46fb-8a32-418e97f9cfe1"
	arbB1ArdentPleaOracle   = "49a8ca54-30d3-42d3-b122-334d24805e1b"
	arbB1RemnantOracle      = "bbf07946-33d0-48fe-85af-022e9b0b834c"
	arbB1LeechOracle        = "16b723bc-aab9-40f2-8e1b-b9c2bd2af3bc"
	arbB1PridemageOracle    = "10821e37-8a83-4d8e-a12f-249a10933a22"
	arbB1TineOracle         = "cdc0a1a5-d9e4-4923-b472-2fa78212167f"
	arbB1SniperOracle       = "f5bdf226-7a83-469b-bc49-87af388d9020"
	arbB1LorescaleOracle    = "c87440aa-4616-49ff-a211-ebdc37767f62"
	arbB1GloryscaleOracle   = "d94a6e12-0197-4125-9a8c-431e42edb1da"
	arbB1GhoulOracle        = "708b13ef-f297-4176-9d09-e633c4d600da"
	arbB1HereticOracle      = "001c6369-df13-427d-89df-718d5c09f382"
	arbB1CarabidOracle      = "2e29025e-6954-4fbd-933f-81c2309c704e"
	arbB1MinotaurOracle     = "3992369a-2158-4cf8-b7e5-be517b91c1f5"
	arbB1BearType           = "Creature — Bear"
	arbB1ColossalMightName  = "Colossal Might"
	arbB1SangriteBacklashNm = "Sangrite Backlash"
)

func TestArb5586B1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		arbB1LeotauOracle:      "Grizzled Leotau",
		arbB1RhoxBruteOracle:   "Rhox Brute",
		arbB1YearlingOracle:    "Cerodon Yearling",
		arbB1TwinclawsOracle:   "Marisi's Twinclaws",
		arbB1SewnEyeOracle:     "Sewn-Eye Drake",
		arbB1TalonOracle:       "Talon Trooper",
		arbB1CoatlOracle:       "Winged Coatl",
		arbB1SphinxOracle:      "Sphinx of the Steel Wind",
		arbB1EthercasteOracle:  "Ethercaste Knight",
		arbB1BehemothOracle:    "Sigiled Behemoth",
		arbB1AbominationOracle: "Etherium Abomination",
		arbB1JhessianOracle:    "Jhessian Zombies",
		arbB1RecluseOracle:     "Pale Recluse",
		arbB1PlowbeastOracle:   "Sanctum Plowbeast",
		arbB1RannetOracle:      "Valley Rannet",
		arbB1PouncerOracle:     "Igneous Pouncer",
		arbB1SurebladeOracle:   "Bant Sureblade",
		arbB1StormbladeOracle:  "Esper Stormblade",
		arbB1GrimbladeOracle:   "Grixis Grimblade",
		arbB1HackbladeOracle:   "Jund Hackblade",
		arbB1HushbladeOracle:   "Naya Hushblade",
		arbB1FalconsOracle:     "Messenger Falcons",
		arbB1RenegadesOracle:   "Vithian Renegades",
		arbB1ArmorguardOracle:  "Leonin Armorguard",
		arbB1GnawersOracle:     "Sanity Gnawers",
		arbB1FiligreeOracle:    "Filigree Angel",
		arbB1CyclopsOracle:     "Madrush Cyclops",
		arbB1MagefireOracle:    "Magefire Wings",
		arbB1SangriteOracle:    arbB1SangriteBacklashNm,
		arbB1BreathOracle:      "Breath of Malfegor",
		arbB1OfferingOracle:    "Offering to Asha",
		arbB1ZealousOracle:     "Zealous Persecution",
		arbB1FlurryOracle:      "Flurry of Wings",
		arbB1RebornHopeOracle:  "Reborn Hope",
		arbB1ColossalOracle:    arbB1ColossalMightName,
		arbB1SunlightOracle:    "Captured Sunlight",
		arbB1DreadOracle:       "Demonic Dread",
		arbB1DenyRealityOracle: "Deny Reality",
		arbB1ArdentPleaOracle:  "Ardent Plea",
		arbB1RemnantOracle:     "Kathari Remnant",
		arbB1LeechOracle:       "Putrid Leech",
		arbB1PridemageOracle:   "Qasali Pridemage",
		arbB1TineOracle:        "Unbender Tine",
		arbB1SniperOracle:      "Stun Sniper",
		arbB1LorescaleOracle:   "Lorescale Coatl",
		arbB1GloryscaleOracle:  "Gloryscale Viashino",
		arbB1GhoulOracle:       "Vedalken Ghoul",
		arbB1HereticOracle:     "Vedalken Heretic",
		arbB1CarabidOracle:     "Monstrous Carabid",
		arbB1MinotaurOracle:    "Deadshot Minotaur",
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

// arbB1Push puts a catalog card with the given colours onto the
// battlefield under owner's control.
func arbB1Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Colors: colors, Owner: owner, Controller: owner,
	})
}

// arbB1Graveyard puts a card of the given colours into p's graveyard.
func arbB1Graveyard(p *game.Player, name, typeLine string, colors ...string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Colors: colors, Owner: p.ID, Controller: p.ID})
	return id
}

// arbB1Flyer puts a creature with printed flying onto the battlefield.
func arbB1Flyer(g *game.Game, owner uuid.UUID, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird",
		Power: power, Toughness: toughness, Keywords: []string{"flying"}, Owner: owner, Controller: owner,
	})
}

// --- printed keywords ------------------------------------------------------------

func TestArb5586B1CardsHaveTheirPrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Cerodon Yearling", "Creature — Beast", arbB1YearlingOracle, []string{"vigilance", "haste"}},
		{"Marisi's Twinclaws", "Creature — Cat Warrior", arbB1TwinclawsOracle, []string{"double strike"}},
		{"Sewn-Eye Drake", "Creature — Zombie Drake", arbB1SewnEyeOracle, []string{"flying", "haste"}},
		{"Talon Trooper", "Creature — Bird Scout", arbB1TalonOracle, []string{"flying"}},
		{"Winged Coatl", "Creature — Snake", arbB1CoatlOracle, []string{"flash", "flying", "deathtouch"}},
		{"Sphinx of the Steel Wind", "Artifact Creature — Sphinx", arbB1SphinxOracle,
			[]string{"flying", "first strike", "vigilance", "lifelink", "protection from red", "protection from green"}},
		{"Ethercaste Knight", "Artifact Creature — Human Knight", arbB1EthercasteOracle, []string{game.KeywordExalted}},
		{"Sigiled Behemoth", "Creature — Beast", arbB1BehemothOracle, []string{game.KeywordExalted}},
		{"Qasali Pridemage", "Creature — Cat Wizard", arbB1PridemageOracle, []string{game.KeywordExalted}},
		{"Ardent Plea", "Enchantment", arbB1ArdentPleaOracle, []string{game.KeywordExalted}},
		{"Jhessian Zombies", "Creature — Zombie", arbB1JhessianOracle, []string{"fear"}},
		{"Pale Recluse", "Creature — Spider", arbB1RecluseOracle, []string{"reach"}},
		{"Sanctum Plowbeast", "Artifact Creature — Beast", arbB1PlowbeastOracle, []string{"defender"}},
		{"Igneous Pouncer", "Creature — Elemental", arbB1PouncerOracle, []string{"haste"}},
		{"Messenger Falcons", "Creature — Bird", arbB1FalconsOracle, []string{"flying"}},
		{"Filigree Angel", "Artifact Creature — Angel", arbB1FiligreeOracle, []string{"flying"}},
		{"Kathari Remnant", "Creature — Bird Skeleton", arbB1RemnantOracle, []string{"flying"}},
	} {
		g := newCatalogGame(t)
		id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		for _, kw := range tc.keywords {
			if !hasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

// Flash: Winged Coatl is cast at the beginning of combat.
func TestArb5586B1WingedCoatlHasFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b13HandCard(me, "Winged Coatl", "Creature — Snake", arbB1CoatlOracle)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("Winged Coatl at the beginning of combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("Winged Coatl did not resolve")
	}
}

// Protection from red: a red-green spell can't target the Sphinx.
func TestArb5586B1SphinxOfTheSteelWindHasProtectionFromRed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sphinx := arbB1Push(g, opp.ID, "Sphinx of the Steel Wind", "Artifact Creature — Sphinx", arbB1SphinxOracle, 6, 6, "W", "U", "B")
	might := handCardFull(me, arbB1ColossalMightName, "Instant", "{R}{G}", arbB1ColossalOracle, []string{"R", "G"})
	if err := g.CastSpell(me.ID, might, game.CastSpellParams{Targets: cardTarget(sphinx)}); err == nil {
		t.Error("a red spell targeted a creature with protection from red")
	}
}

// --- exalted -------------------------------------------------------------------------

func TestArb5586B1ExaltedPumpsALoneAttacker(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Ethercaste Knight", "Artifact Creature — Human Knight", arbB1EthercasteOracle},
		{"Sigiled Behemoth", "Creature — Beast", arbB1BehemothOracle},
		{"Qasali Pridemage", "Creature — Cat Wizard", arbB1PridemageOracle},
		{"Ardent Plea", "Enchantment", arbB1ArdentPleaOracle},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 4)
		bear := seedBear(g, me.ID)
		alaB1Attack(t, g, opp.ID, bear)
		if got := effectivePower(t, g, bear); got != 3 {
			t.Errorf("%s: the lone attacker has power %d, want 3", tc.name, got)
		}
	}
}

// --- unearth ---------------------------------------------------------------------------

func TestArb5586B1EtheriumAbominationUnearths(t *testing.T) {
	tc := broB04Card{"Etherium Abomination", "Artifact Creature — Horror", arbB1AbominationOracle, 4, 3, "{1}{U}{B}"}
	var found *game.ActivatedAbilityShape
	for _, ab := range game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle}) {
		if game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
			ab := ab
			found = &ab
		}
	}
	if found == nil || found.Cost.Mana != tc.unearth || !found.SorcerySpeed {
		t.Fatalf("unearth row %+v, want {1}{U}{B} at sorcery speed", found)
	}
	g := newCatalogGame(t)
	id, _ := broB04Unearth(t, g, tc)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Etherium Abomination did not return")
	}
	if !hasEffectiveKeyword(t, g, id, "haste") {
		t.Error("the unearthed Abomination lacks haste")
	}
}

// --- landcycling ---------------------------------------------------------------------------

func TestArb5586B1DualLandcyclersFetchEachType(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		idx                    int
		land, other            string
	}{
		{"Jhessian Zombies", "Creature — Zombie", arbB1JhessianOracle, 0, "Island", "Mountain"},
		{"Jhessian Zombies", "Creature — Zombie", arbB1JhessianOracle, 1, "Swamp", "Forest"},
		{"Pale Recluse", "Creature — Spider", arbB1RecluseOracle, 0, "Forest", "Island"},
		{"Pale Recluse", "Creature — Spider", arbB1RecluseOracle, 1, "Plains", "Swamp"},
		{"Sanctum Plowbeast", "Artifact Creature — Beast", arbB1PlowbeastOracle, 0, "Plains", "Mountain"},
		{"Sanctum Plowbeast", "Artifact Creature — Beast", arbB1PlowbeastOracle, 1, "Island", "Forest"},
		{"Valley Rannet", "Creature — Beast", arbB1RannetOracle, 0, "Mountain", "Plains"},
		{"Valley Rannet", "Creature — Beast", arbB1RannetOracle, 1, "Forest", "Island"},
		{"Igneous Pouncer", "Creature — Elemental", arbB1PouncerOracle, 0, "Swamp", "Plains"},
		{"Igneous Pouncer", "Creature — Elemental", arbB1PouncerOracle, 1, "Mountain", "Island"},
	} {
		t.Run(tc.name+"/"+tc.land, func(t *testing.T) {
			scgB1Typecycle(t, tc.name, tc.typeLine, tc.oracle, tc.idx,
				tc.land, "Basic Land — "+tc.land, tc.other, "Basic Land — "+tc.other)
		})
	}
}

// --- the blades -------------------------------------------------------------------------------

func TestArb5586B1BladesNeedAnotherMulticoloredPermanent(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, keyword string
	}{
		{"Bant Sureblade", "Creature — Human Soldier", arbB1SurebladeOracle, "first strike"},
		{"Esper Stormblade", "Artifact Creature — Vedalken Wizard", arbB1StormbladeOracle, "flying"},
		{"Grixis Grimblade", "Creature — Zombie Warrior", arbB1GrimbladeOracle, "deathtouch"},
		{"Jund Hackblade", "Creature — Goblin Berserker", arbB1HackbladeOracle, "haste"},
		{"Naya Hushblade", "Creature — Elf Rogue", arbB1HushbladeOracle, "shroud"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			// The blade is multicolored itself: it is never "another".
			blade := arbB1Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 1, "W", "U")
			arbB1Push(g, me.ID, "Mono Relic", "Artifact", "", 0, 0, "R")
			arbB1Push(g, opp.ID, "Their Gold Relic", "Artifact", "", 0, 0, "B", "G")
			if p, tough := effectivePTOf(t, g, blade); p != 2 || tough != 1 {
				t.Errorf("without another multicolored permanent it is %d/%d, want 2/1", p, tough)
			}
			if hasEffectiveKeyword(t, g, blade, tc.keyword) {
				t.Errorf("without another multicolored permanent it has %s", tc.keyword)
			}
			arbB1Push(g, me.ID, "Gold Relic", "Artifact", "", 0, 0, "R", "G")
			if p, tough := effectivePTOf(t, g, blade); p != 3 || tough != 2 {
				t.Errorf("with another multicolored permanent it is %d/%d, want 3/2", p, tough)
			}
			if !hasEffectiveKeyword(t, g, blade, tc.keyword) {
				t.Errorf("with another multicolored permanent it lacks %s", tc.keyword)
			}
		})
	}
}

// --- enters triggers -------------------------------------------------------------------------

func TestArb5586B1MessengerFalconsDrawsOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B")
	advanceToMain(t, g)
	hand := me.Hand.Size()
	fdnCCast(t, g, "Messenger Falcons", "Creature — Bird", arbB1FalconsOracle, 2, 2)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want the Falcons cast and one card drawn", hand, got)
	}
}

func TestArb5586B1VithianRenegadesDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	art := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", arbB1BearType, 2, 2)
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Vithian Renegades", TypeLine: "Creature — Human Shaman",
		OracleID: arbB1RenegadesOracle, Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	s12Drive(t, g, me.ID, art, bear)
	if g.Battlefield.Contains(art) {
		t.Error("the artifact survived the Renegades' enter trigger")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was destroyed")
	}
}

func TestArb5586B1DeadshotMinotaurShootsAFlyer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flyer := arbB1Flyer(g, opp.ID, 2, 3)
	walker := b12Creature(g, opp.ID, "Their Bear", arbB1BearType, 2, 2)
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Deadshot Minotaur", TypeLine: "Creature — Minotaur",
		OracleID: arbB1MinotaurOracle, Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	s12Drive(t, g, me.ID, flyer, walker)
	if g.Battlefield.Contains(flyer) {
		t.Error("3 damage did not kill the 2/3 flyer")
	}
	if !g.Battlefield.Contains(walker) {
		t.Error("the creature without flying died")
	}
}

func TestArb5586B1CyclingCreaturesCycle(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle, mana string }{
		{"Deadshot Minotaur", "Creature — Minotaur", arbB1MinotaurOracle, "{R}"},
		{"Monstrous Carabid", "Creature — Insect", arbB1CarabidOracle, "{B}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			seedLibrary(g.Seats[g.Turn.ActiveSeat], "A", "B")
			ons1Cycle(t, g, tc.name, tc.typeLine, tc.oracle, tc.mana)
		})
	}
}

func TestArb5586B1LeoninArmorguardPumpsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", arbB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", arbB1BearType, 2, 2)
	guard := fdnCCast(t, g, "Leonin Armorguard", "Creature — Cat Soldier", arbB1ArmorguardOracle, 3, 3)
	if p, tough := effectivePTOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, guard); p != 4 || tough != 4 {
		t.Errorf("the Armorguard is %d/%d, want 4/4", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestArb5586B1SanityGnawersDiscardsAtRandom(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Hand.Cards = nil
	card := pushToHand(g, opp.ID, "Their Card", "Sorcery")
	advanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Sanity Gnawers", TypeLine: "Creature — Rat",
		OracleID: arbB1GnawersOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickPlayer(t, g, me.ID, opp.ID)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(card) || opp.Hand.Contains(card) {
		t.Error("the target player's only card was not discarded")
	}
}

func TestArb5586B1FiligreeAngelGainsThreePerArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Permanent(g, me.ID, "Rock", "Artifact")
	b12Permanent(g, me.ID, "Other Rock", "Artifact")
	b12Permanent(g, g.Seats[1].ID, "Their Rock", "Artifact")
	advanceToMain(t, g)
	life := me.Life
	fdnCCast(t, g, "Filigree Angel", "Artifact Creature — Angel", arbB1FiligreeOracle, 4, 4)
	if me.Life != life+9 {
		t.Errorf("life %d -> %d, want +9 for three artifacts (the Angel counts)", life, me.Life)
	}
}

// --- statics -------------------------------------------------------------------------------

func TestArb5586B1MadrushCyclopsGivesYourCreaturesHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cyclops := b12Push(g, me.ID, "Madrush Cyclops", "Creature — Cyclops Warrior", arbB1CyclopsOracle, 3, 4)
	mine := b12Creature(g, me.ID, "My Bear", arbB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", arbB1BearType, 2, 2)
	for _, id := range []uuid.UUID{cyclops, mine} {
		if !hasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%s lacks haste", id)
		}
	}
	if hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature has haste")
	}
}

func TestArb5586B1MagefireWingsPumpsAndGivesFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", arbB1BearType, 2, 2)
	castCatalogSpell(t, g, "Magefire Wings", auraTypeLine, arbB1MagefireOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the enchanted creature is %d/%d, want 4/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the enchanted creature lacks flying")
	}
}

func TestArb5586B1SangriteBacklashGivesPlusThreeMinusThree(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 2, 5)
	castCatalogSpell(t, g, arbB1SangriteBacklashNm, auraTypeLine, arbB1SangriteOracle, cardRefs(big))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, big); p != 5 || tough != 2 {
		t.Errorf("the enchanted creature is %d/%d, want 5/2", p, tough)
	}

	g = newCatalogGame(t)
	opp = g.Seats[1]
	small := b12Creature(g, opp.ID, "Their Bear", arbB1BearType, 2, 3)
	castCatalogSpell(t, g, arbB1SangriteBacklashNm, auraTypeLine, arbB1SangriteOracle, cardRefs(small))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("a creature with toughness 3 survived -3")
	}
}

// --- spells ---------------------------------------------------------------------------------

func TestArb5586B1BreathOfMalfegorHitsEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	castCatalogSpell(t, g, "Breath of Malfegor", "Instant", arbB1BreathOracle, nil)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := before[p.ID] - 5
		if p.ID == me.ID {
			want = before[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %s life %d -> %d, want %d", p.ID, before[p.ID], p.Life, want)
		}
	}
}

// arbB1OfferingBoard has seat 0 cast a sorcery and seat 1 answer it with
// Offering to Asha, stopping at the tax prompt.
func arbB1OfferingBoard(t *testing.T) (g *game.Game, victim uuid.UUID, caster, them *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	caster, them = g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	victim = castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	offering := handCardFull(them, "Offering to Asha", "Instant", "", arbB1OfferingOracle, []string{"W", "U"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	if err := g.CastSpell(them.ID, offering, game.CastSpellParams{Targets: cardRefs(victim)}); err != nil {
		t.Fatalf("casting Offering to Asha: %v", err)
	}
	passUntilTaxed(t, g, caster.ID)
	return g, victim, caster, them
}

func TestArb5586B1OfferingToAshaCountersAndGainsFour(t *testing.T) {
	g, victim, caster, them := arbB1OfferingBoard(t)
	life := them.Life
	prompt := answerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{4}" {
		t.Errorf("the tax is %q, want {4}", prompt.PayCost)
	}
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Error("a declined tax must counter the spell")
	}
	if them.Life != life+4 {
		t.Errorf("life %d -> %d, want +4", life, them.Life)
	}
}

func TestArb5586B1OfferingToAshaGainsFourWhenTheTaxIsPaid(t *testing.T) {
	g, victim, caster, them := arbB1OfferingBoard(t)
	life := them.Life
	for i := 0; i < 4; i++ {
		caster.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	answerPayUnless(t, g, caster.ID, true)
	if !g.Stack.Contains(victim) {
		t.Error("a paid tax countered the spell anyway")
	}
	passPriorityAroundTable(t, g)
	if them.Life != life+4 {
		t.Errorf("life %d -> %d, want +4 whether or not the tax was paid", life, them.Life)
	}
}

func TestArb5586B1ZealousPersecutionPumpsYoursAndShrinksTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", arbB1BearType, 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", arbB1BearType, 2, 2)
	weenie := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1)
	castCatalogSpell(t, g, "Zealous Persecution", "Instant", arbB1ZealousOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, mine); p != 3 || tough != 3 {
		t.Errorf("your creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("an opponent's creature is %d/%d, want 1/1", p, tough)
	}
	if g.Battlefield.Contains(weenie) {
		t.Error("an opponent's 1/1 survived -1/-1")
	}
}

func TestArb5586B1FlurryOfWingsMakesABirdPerAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a, b := seedBear(g, me.ID), seedBear(g, me.ID)
	sfAttackHold(t, g, opp.ID, a, b)
	sfCast(t, g, me, sfInstant("Flurry of Wings", arbB1FlurryOracle), game.CastSpellParams{})
	var birds []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bird Soldier" && c.Controller == me.ID {
			birds = append(birds, c.InstanceID)
		}
	}
	if len(birds) != 2 {
		t.Fatalf("%d Bird Soldiers for two attackers, want 2", len(birds))
	}
	if !hasEffectiveKeyword(t, g, birds[0], "flying") {
		t.Error("the Bird Soldier lacks flying")
	}
	if p, tough := effectivePTOf(t, g, birds[0]); p != 1 || tough != 1 {
		t.Errorf("the Bird Soldier is %d/%d, want 1/1", p, tough)
	}
}

func TestArb5586B1RebornHopeReturnsAMulticoloredCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mono := arbB1Graveyard(me, "Mono Card", "Sorcery", "G")
	gold := arbB1Graveyard(me, "Gold Card", "Sorcery", "G", "W")
	advanceToMain(t, g)
	refused := b13HandCard(me, "Reborn Hope", "Sorcery", arbB1RebornHopeOracle)
	if err := g.CastSpell(me.ID, refused, game.CastSpellParams{Targets: cardTarget(mono)}); err == nil {
		t.Fatal("Reborn Hope targeted a monocolored card")
	}
	castCatalogSpell(t, g, "Reborn Hope", "Sorcery", arbB1RebornHopeOracle, cardTarget(gold))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(gold) {
		t.Error("the multicolored card did not return to hand")
	}
}

func TestArb5586B1ColossalMightPumpsAndGivesTrample(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[0].ID, "My Bear", arbB1BearType, 2, 2)
	castCatalogSpell(t, g, arbB1ColossalMightName, "Instant", arbB1ColossalOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 6 || tough != 4 {
		t.Errorf("the target is %d/%d, want 6/4", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the target lacks trample")
	}
}

// --- cascade --------------------------------------------------------------------------------

func TestArb5586B1CascadeIsWired(t *testing.T) {
	for _, oracle := range []string{
		arbB1SunlightOracle, arbB1DreadOracle, arbB1DenyRealityOracle, arbB1ArdentPleaOracle, arbB1RemnantOracle,
	} {
		spec, _ := Lookup(oracle)
		if len(spec.Triggered) != 1 || spec.Triggered[0].Keyword != KeywordCascade {
			t.Errorf("%s: want exactly one cascade trigger, got %d triggers", spec.Name, len(spec.Triggered))
		}
	}
}

func TestArb5586B1CapturedSunlightCascadesThenGainsFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedCheapLibrary(g, me, 3)
	life := me.Life
	castWithCost(t, g, "Captured Sunlight", "Sorcery", "{2}{G}{W}", arbB1SunlightOracle)
	passPriorityAroundTable(t, g)
	if got := countMayCastPrompts(g); got != 1 {
		t.Fatalf("cascade offers = %d, want 1", got)
	}
	answerAllMayCast(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d -> %d, want +4", life, me.Life)
	}
}

func TestArb5586B1DemonicDreadStopsABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := seedBear(g, me.ID)
	victim := b12Creature(g, opp.ID, "Their Bear", arbB1BearType, 2, 2)
	castCatalogSpell(t, g, "Demonic Dread", "Sorcery", arbB1DreadOracle, cardTarget(victim))
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, mine)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(victim, mine); err == nil {
		t.Error("the target blocked the turn it can't block")
	}
}

func TestArb5586B1DenyRealityBouncesAPermanent(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	castCatalogSpell(t, g, "Deny Reality", "Sorcery", arbB1DenyRealityOracle, cardTarget(relic))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) || !opp.Hand.Contains(relic) {
		t.Error("the permanent did not return to its owner's hand")
	}
}

func TestArb5586B1KathariRemnantRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b12Push(g, me.ID, "Kathari Remnant", "Creature — Bird Skeleton", arbB1RemnantOracle, 0, 1)
	ravB2MustActivate(t, g, me.ID, id, 0, game.ActivateAbilityParams{}, "B")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	if !g.Battlefield.Contains(id) {
		t.Fatal("the regeneration shield did not save it")
	}
	if !isTapped(g, id) {
		t.Error("a regenerated creature is tapped")
	}
}

// --- activated abilities -----------------------------------------------------------------------

func TestArb5586B1PutridLeechPumpsForTwoLifeOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b12Push(g, me.ID, "Putrid Leech", "Creature — Zombie Leech", arbB1LeechOracle, 2, 2)
	life := me.Life
	ravB2MustActivate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, me.Life)
	}
	if p, tough := effectivePTOf(t, g, id); p != 4 || tough != 4 {
		t.Errorf("it is %d/%d, want 4/4", p, tough)
	}
	if err := ravB2Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated a second time in one turn")
	}
}

func TestArb5586B1QasaliPridemageDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := seatsForCombat(g)
	mage := b12Push(g, me.ID, "Qasali Pridemage", "Creature — Cat Wizard", arbB1PridemageOracle, 2, 2)
	bear := b12Creature(g, opp.ID, "Their Bear", arbB1BearType, 2, 2)
	if err := ravB2Activate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)}, "C"); err == nil {
		t.Fatal("the Pridemage targeted a creature")
	}
	shrine := b12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	ravB2MustActivate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{Targets: cardTarget(shrine)}, "C")
	if g.Battlefield.Contains(shrine) {
		t.Error("the enchantment survived")
	}
	if g.Battlefield.Contains(mage) {
		t.Error("the Pridemage was not sacrificed")
	}
}

func TestArb5586B1UnbenderTineUntapsAnotherPermanent(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := seatsForCombat(g)
	tine := b12Push(g, me.ID, "Unbender Tine", "Artifact", arbB1TineOracle, 0, 0)
	if err := ravB2Activate(t, g, me.ID, tine, 0, game.ActivateAbilityParams{Targets: cardTarget(tine)}); err == nil {
		t.Fatal("Unbender Tine targeted itself")
	}
	land := seedLand(g, them.ID, "Forest", "Basic Land — Forest", "")
	findBattlefieldCardForTest(g, land).Tapped = true
	ravB2MustActivate(t, g, me.ID, tine, 0, game.ActivateAbilityParams{Targets: cardTarget(land)})
	if isTapped(g, land) {
		t.Error("the land was not untapped")
	}
	if !isTapped(g, tine) {
		t.Error("the Tine did not tap")
	}
}

func TestArb5586B1StunSniperPingsAndTaps(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := seatsForCombat(g)
	sniper := b12Push(g, me.ID, "Stun Sniper", "Creature — Human Archer", arbB1SniperOracle, 1, 1)
	bear := b12Creature(g, them.ID, "Their Bear", arbB1BearType, 2, 2)
	ravB2MustActivate(t, g, me.ID, sniper, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)}, "C")
	if !g.Battlefield.Contains(bear) {
		t.Fatal("1 damage killed a 2/2")
	}
	if !isTapped(g, bear) {
		t.Error("the target was not tapped")
	}

	g = newCatalogGame(t)
	advanceToMain(t, g)
	me, them = seatsForCombat(g)
	sniper = b12Push(g, me.ID, "Stun Sniper", "Creature — Human Archer", arbB1SniperOracle, 1, 1)
	elf := b12Creature(g, them.ID, "Their Elf", "Creature — Elf", 1, 1)
	ravB2MustActivate(t, g, me.ID, sniper, 0, game.ActivateAbilityParams{Targets: cardTarget(elf)}, "C")
	if g.Battlefield.Contains(elf) {
		t.Error("1 damage did not kill a 1/1")
	}
}

// --- triggers ------------------------------------------------------------------------------------

func TestArb5586B1LorescaleCoatlGrowsOnYourDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B")
	seedLibrary(opp, "C")
	coatl := b12Push(g, me.ID, "Lorescale Coatl", "Creature — Snake", arbB1LorescaleOracle, 2, 2)
	advanceToMain(t, g)
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, coatl); n != 1 {
		t.Fatalf("%d counters after one draw, want 1", n)
	}
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatalf("opponent draw: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, coatl); n != 1 {
		t.Errorf("an opponent's draw grew it to %d counters", n)
	}
}

func TestArb5586B1GloryscaleViashinoGrowsOnAMulticoloredSpell(t *testing.T) {
	for _, tc := range []struct {
		colors []string
		want   int
	}{
		{[]string{"R", "G"}, 6},
		{[]string{"R"}, 3},
	} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		viashino := b12Push(g, me.ID, "Gloryscale Viashino", "Creature — Lizard Soldier", arbB1GloryscaleOracle, 3, 3)
		spell := handCardFull(me, "Some Spell", "Instant", "", "", tc.colors)
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
		if got := effectivePower(t, g, viashino); got != tc.want {
			t.Errorf("after a %v spell the Viashino has power %d, want %d", tc.colors, got, tc.want)
		}
	}
}

func TestArb5586B1VedalkenGhoulDrainsWhenBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghoul := b12Push(g, me.ID, "Vedalken Ghoul", "Creature — Vedalken Zombie", arbB1GhoulOracle, 1, 1)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, ghoul)
	advanceTo(t, g, game.StepDeclareBlockers)
	life := opp.Life
	if err := g.DeclareBlocker(wall, ghoul); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("defending player's life %d -> %d, want -4", life, opp.Life)
	}
}

func TestArb5586B1VedalkenHereticMayDrawOnDamageToAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B")
	heretic := b12Push(g, me.ID, "Vedalken Heretic", "Creature — Vedalken Rogue", arbB1HereticOracle, 1, 1)
	emitNoncombatDamage(g, heretic, me.ID, 1)
	if len(g.PendingChoices) != 0 {
		t.Fatal("the Heretic asked about damage dealt to its own controller")
	}
	before := me.Hand.Size()
	emitNoncombatDamage(g, heretic, opp.ID, 1)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func TestArb5586B1MonstrousCarabidMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	carabid := b12Push(g, me.ID, "Monstrous Carabid", "Creature — Insect", arbB1CarabidOracle, 4, 4)
	advanceToDeclareAttackersOf(t, g, seat)
	if err := g.PassPriority(); err == nil {
		t.Fatal("passed the declaration with the Carabid at home")
	}
	if err := g.DeclareAttacker(carabid, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Carabid attacking: %v", err)
	}
}
