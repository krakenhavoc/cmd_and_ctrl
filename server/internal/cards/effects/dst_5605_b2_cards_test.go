package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dst_5605_b2_cards_test.go — Darksteel, batch 2 (#5605, S82). Every test
// plays the real catalog entry through the engine.

const (
	dstB2BurdenOfGreedOracle       = "a80c13db-949d-4784-8177-04aaa22b96b0"
	dstB2ChitteringRatsOracle      = "08dfe42e-35c0-4be0-abba-57269792ff3d"
	dstB2CrazedGoblinOracle        = "fafe2d90-e5fc-401a-a971-829192bec6c0"
	dstB2DarksteelBruteOracle      = "4f7d2774-2aec-4992-b35b-e03265e517af"
	dstB2DarksteelGargoyleOracle   = "73010421-374f-458e-aa88-248ef8ae4f8b"
	dstB2DarksteelPendantOracle    = "431838a8-f020-4e4e-a6f4-2d4ca27c56df"
	dstB2DemonsHornOracle          = "ee38377f-8d3a-402e-a5d7-a01cfe3a5322"
	dstB2DragonsClawOracle         = "bd8d707d-1201-42c6-bce6-fbda12ce5ea8"
	dstB2DrossGolemOracle          = "336d51cf-dcb3-4257-bd6c-926820a2f9de"
	dstB2EmissaryOfDespairOracle   = "4efa1654-c822-4477-ad24-f379ed7ecb11"
	dstB2EmissaryOfHopeOracle      = "62a6dbdb-70fc-43ab-8e63-367a609ecdeb"
	dstB2EssenceDrainOracle        = "4bf5402f-7bba-4c0c-aebd-1b231768f2b2"
	dstB2FangrenFirstbornOracle    = "2affbe3f-5292-43d9-a08c-23bf44629cc9"
	dstB2GrimclawBatsOracle        = "161ebbc8-b8f9-4ac4-80f3-e290070a1e30"
	dstB2HoverguardObserverOracle  = "5ca8d756-a55a-4b36-b0a6-562a39484952"
	dstB2HungerOfTheNimOracle      = "1af3c6ff-2884-4d8c-a01f-6f74d8ea10cc"
	dstB2KrakensEyeOracle          = "14a21d8a-2097-4150-8af5-8c155049477a"
	dstB2KrarkClanStokerOracle     = "756b4bc9-6f3f-4eae-b6b0-63c94ed8482f"
	dstB2LoxodonMysticOracle       = "f3e22884-a9ca-4479-b400-df1c1d2fd153"
	dstB2MagneticFluxOracle        = "d50f756d-9ab6-4597-86ae-80beb70879ac"
	dstB2MetalFatigueOracle        = "f817352a-908a-4cb5-8e10-feeb770540fc"
	dstB2MyrMoonvesselOracle       = "3e922661-80df-4e84-a12a-524bc74e6c9d"
	dstB2NourishOracle             = "9d5e82d3-79ac-4dbe-80e8-1db6dd6c8767"
	dstB2OxiddaGolemOracle         = "9c51d843-d155-44ed-ab1f-dfa2ff848f53"
	dstB2OxidizeOracle             = "0ff78353-a26e-4b5a-948d-1c3b41d2dd1f"
	dstB2PurgeOracle               = "d06f0f7c-0270-44be-991a-a5ebbddb0900"
	dstB2RazorGolemOracle          = "c7b2188a-7571-48b8-ac66-ca741bf2e779"
	dstB2RetractOracle             = "31cefc9d-cc13-4b32-a1f8-edbcaf35aa3b"
	dstB2RitualOfRestorationOracle = "7cfd0320-0b1c-4ef4-8e17-600cbcbddd6f"
	dstB2RoaringSlagwurmOracle     = "856d8eb4-3834-49f8-9696-579948590fd0"
	dstB2ScavengingScarabOracle    = "3ce2f3ad-23b3-40fe-bb31-25dbe9d1150d"
	dstB2SoulscourOracle           = "d7644325-8a0f-4f40-bab6-518df7ac1d14"
	dstB2TangleGolemOracle         = "733c6e30-3c0c-41f0-b1e9-4c2eb76f7b82"
	dstB2TangleSpiderOracle        = "597bd341-c5a0-47a3-9864-1fccab8a2a00"
	dstB2TelJiladOutriderOracle    = "d83d784e-e95b-48a1-9b4a-362f4fccc529"
	dstB2UrGolemsEyeOracle         = "fb34fc00-e60d-41fc-9393-ca4248ec0a1c"
	dstB2ViridianAcolyteOracle     = "9db7d8f0-0712-4e32-825c-1339e7a8d768"
	dstB2ViridianZealotOracle      = "9410ae41-b6db-47c6-ae0a-4e33aad1d3ce"
	dstB2VoltaicConstructOracle    = "f5677831-6ca9-4c7b-b02d-f44079c3f663"
	dstB2VulshokMorningstarOracle  = "12a8adc4-927f-4314-b2ef-9c647ace68d5"
	dstB2WurmsToothOracle          = "69476326-fda5-48c8-a506-e71eb7c81c2f"

	dstB2GolemType = "Artifact Creature — Golem"
)

func TestDst5605B2CardsAreRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Burden of Greed": dstB2BurdenOfGreedOracle, "Chittering Rats": dstB2ChitteringRatsOracle,
		"Crazed Goblin": dstB2CrazedGoblinOracle, "Darksteel Brute": dstB2DarksteelBruteOracle,
		"Darksteel Gargoyle": dstB2DarksteelGargoyleOracle, "Darksteel Pendant": dstB2DarksteelPendantOracle,
		"Demon's Horn": dstB2DemonsHornOracle, "Dragon's Claw": dstB2DragonsClawOracle,
		"Dross Golem": dstB2DrossGolemOracle, "Emissary of Despair": dstB2EmissaryOfDespairOracle,
		"Emissary of Hope": dstB2EmissaryOfHopeOracle, "Essence Drain": dstB2EssenceDrainOracle,
		"Fangren Firstborn": dstB2FangrenFirstbornOracle, "Grimclaw Bats": dstB2GrimclawBatsOracle,
		"Hoverguard Observer": dstB2HoverguardObserverOracle, "Hunger of the Nim": dstB2HungerOfTheNimOracle,
		"Kraken's Eye": dstB2KrakensEyeOracle, "Krark-Clan Stoker": dstB2KrarkClanStokerOracle,
		"Loxodon Mystic": dstB2LoxodonMysticOracle, "Magnetic Flux": dstB2MagneticFluxOracle,
		"Metal Fatigue": dstB2MetalFatigueOracle, "Myr Moonvessel": dstB2MyrMoonvesselOracle,
		"Nourish": dstB2NourishOracle, "Oxidda Golem": dstB2OxiddaGolemOracle,
		"Oxidize": dstB2OxidizeOracle, "Purge": dstB2PurgeOracle,
		"Razor Golem": dstB2RazorGolemOracle, "Retract": dstB2RetractOracle,
		"Ritual of Restoration": dstB2RitualOfRestorationOracle, "Roaring Slagwurm": dstB2RoaringSlagwurmOracle,
		"Scavenging Scarab": dstB2ScavengingScarabOracle, "Soulscour": dstB2SoulscourOracle,
		"Tangle Golem": dstB2TangleGolemOracle, "Tangle Spider": dstB2TangleSpiderOracle,
		"Tel-Jilad Outrider": dstB2TelJiladOutriderOracle, "Ur-Golem's Eye": dstB2UrGolemsEyeOracle,
		"Viridian Acolyte": dstB2ViridianAcolyteOracle, "Viridian Zealot": dstB2ViridianZealotOracle,
		"Voltaic Construct": dstB2VoltaicConstructOracle, "Vulshok Morningstar": dstB2VulshokMorningstarOracle,
		"Wurm's Tooth": dstB2WurmsToothOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered as %q, %v, caveats %v; want its name, Full and none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// dstB2Artifact puts a plain noncreature artifact onto the battlefield.
func dstB2Artifact(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return b12Permanent(g, owner, name, "Artifact")
}

// --- printed keywords ----------------------------------------------------------

func TestDst5605B2KeywordCards(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Darksteel Gargoyle", "Artifact Creature — Gargoyle", dstB2DarksteelGargoyleOracle, []string{"flying", "indestructible"}},
		{"Dross Golem", dstB2GolemType, dstB2DrossGolemOracle, []string{"fear"}},
		{"Oxidda Golem", dstB2GolemType, dstB2OxiddaGolemOracle, []string{"haste"}},
		{"Razor Golem", dstB2GolemType, dstB2RazorGolemOracle, []string{"vigilance"}},
		{"Tangle Spider", "Creature — Spider", dstB2TangleSpiderOracle, []string{"flash", "reach"}},
		{"Tel-Jilad Outrider", "Creature — Elf Warrior", dstB2TelJiladOutriderOracle, []string{"protection from artifacts"}},
		{"Hoverguard Observer", "Creature — Drone", dstB2HoverguardObserverOracle, []string{"flying"}},
		{"Grimclaw Bats", "Creature — Bat", dstB2GrimclawBatsOracle, []string{"flying"}},
		{"Emissary of Despair", "Creature — Spirit", dstB2EmissaryOfDespairOracle, []string{"flying"}},
		{"Emissary of Hope", "Creature — Spirit", dstB2EmissaryOfHopeOracle, []string{"flying"}},
		{"Darksteel Pendant", "Artifact", dstB2DarksteelPendantOracle, []string{"indestructible"}},
		{"Darksteel Brute", "Artifact", dstB2DarksteelBruteOracle, []string{"indestructible"}},
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

// --- affinity for a basic land type -------------------------------------------

func TestDst5605B2GolemsHaveAffinityForTheirLandType(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, cost, land, other string
		printed                         int
	}{
		{"Dross Golem", dstB2DrossGolemOracle, "{5}", "Swamp", "Island", 5},
		{"Oxidda Golem", dstB2OxiddaGolemOracle, "{6}", "Mountain", "Forest", 6},
		{"Razor Golem", dstB2RazorGolemOracle, "{6}", "Plains", "Swamp", 6},
		{"Tangle Golem", dstB2TangleGolemOracle, "{7}", "Forest", "Plains", 7},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			price := func() int {
				base, _ := game.ParseCost(tc.cost)
				out, err := g.ApplyCostModifiers(base, game.CostQuery{
					Card: game.Card{InstanceID: uuid.New(), Name: tc.name, TypeLine: dstB2GolemType,
						ManaCost: tc.cost, OracleID: tc.oracle, Owner: me.ID, Controller: me.ID},
					Controller: me.ID, FromZone: game.ZoneHand,
				})
				if err != nil {
					t.Fatal(err)
				}
				return out.ManaValue()
			}
			if got := price(); got != tc.printed {
				t.Fatalf("no lands: %d, want %d", got, tc.printed)
			}
			b12Permanent(g, me.ID, tc.land, "Basic Land — "+tc.land)
			b12Permanent(g, me.ID, tc.land+" 2", "Basic Land — "+tc.land)
			b12Permanent(g, me.ID, tc.other, "Basic Land — "+tc.other)
			b12Permanent(g, opp.ID, "Their "+tc.land, "Basic Land — "+tc.land)
			if got := price(); got != tc.printed-2 {
				t.Errorf("two %s: %d, want %d", tc.land, got, tc.printed-2)
			}
		})
	}
}

// --- combat statics ------------------------------------------------------------

func TestDst5605B2CrazedGoblinMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	goblin := b12Push(g, me.ID, "Crazed Goblin", "Creature — Goblin Warrior", dstB2CrazedGoblinOracle, 1, 1)
	advanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != goblin {
		t.Fatalf("pass with the Goblin home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(goblin, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Goblin attacking: %v", err)
	}
}

func TestDst5605B2HoverguardObserverBlocksOnlyFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := b12Creature(g, me.ID, "Walker", "Creature — Human", 2, 2)
	flyer := apaPush(g, me.ID, me.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	observer := b12Push(g, opp.ID, "Hoverguard Observer", "Creature — Drone", dstB2HoverguardObserverOracle, 3, 3)
	brAttack(t, g, walker, flyer)
	if err := g.DeclareBlocker(observer, walker); err == nil {
		t.Error("Hoverguard Observer blocked a creature without flying")
	}
	if err := g.DeclareBlocker(observer, flyer); err != nil {
		t.Errorf("Hoverguard Observer can't block a flyer: %v", err)
	}
}

func TestDst5605B2ScavengingScarabCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := b12Creature(g, me.ID, "Walker", "Creature — Human", 2, 2)
	scarab := b12Push(g, opp.ID, "Scavenging Scarab", "Creature — Insect", dstB2ScavengingScarabOracle, 3, 3)
	brAttack(t, g, walker)
	if err := g.DeclareBlocker(scarab, walker); err == nil {
		t.Error("Scavenging Scarab blocked")
	}
}

// --- "whenever a player casts a <colour> spell, you may gain 1 life" ------------

func TestDst5605B2ColourArtifactsGainOnAnyPlayersSpellOfTheirColour(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, color, cost, otherColor, otherCost string
	}{
		{"Demon's Horn", dstB2DemonsHornOracle, "B", "{B}", "W", "{W}"},
		{"Dragon's Claw", dstB2DragonsClawOracle, "R", "{R}", "G", "{G}"},
		{"Kraken's Eye", dstB2KrakensEyeOracle, "U", "{U}", "R", "{R}"},
		{"Wurm's Tooth", dstB2WurmsToothOracle, "G", "{G}", "B", "{B}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			mrd2Catalog(g, opp.ID, tc.name, "Artifact", tc.oracle, "{2}", 0, 0)
			life := opp.Life

			off := dstHand(me, "Off Colour", "Sorcery", tc.otherCost, tc.otherColor)
			if err := g.CastSpell(me.ID, off, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast the off-colour spell: %v", err)
			}
			if owedMays(g, opp.ID) != 0 || triggerOnStackFrom(g, tc.name) {
				t.Errorf("a spell of another colour triggered %s", tc.name)
			}
			passPriorityAroundTable(t, g)

			on := dstHand(me, "On Colour", "Sorcery", tc.cost, tc.color)
			if err := g.CastSpell(me.ID, on, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast the on-colour spell: %v", err)
			}
			b06AnswerAllTriggerPrompts(t, g, opp.ID, true)
			passPriorityAroundTable(t, g)
			if opp.Life != life+1 {
				t.Errorf("an opponent's %s spell: %s's controller has %d life, want %d", tc.color, tc.name, opp.Life, life+1)
			}
		})
	}
}

// --- dies, enters and combat damage triggers -----------------------------------

func TestDst5605B2MyrMoonvesselAddsColourlessWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	vessel := b12Push(g, me.ID, "Myr Moonvessel", "Artifact Creature — Myr", dstB2MyrMoonvesselOracle, 1, 1)
	b25Destroy(g, vessel)
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.CountColor("C"); got != 1 {
		t.Errorf("pool holds %d {C} after the Moonvessel died, want 1", got)
	}
}

func TestDst5605B2ChitteringRatsMakesTheTargetOpponentPutACardBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	secret := handCardForTest(opp, "Secret Plan", "Sorcery", "")
	hand := handSize(opp)

	mrdEnter(t, g, me, "Chittering Rats", "Creature — Rat", dstB2ChitteringRatsOracle)
	answerTriggerTargets2(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, opp.ID, secret)
	passPriorityAroundTable(t, g)

	if got := handSize(opp); got != hand-1 {
		t.Errorf("opponent's hand %d, want %d", got, hand-1)
	}
	if top, err := opp.Library.Top(); err != nil || top.InstanceID != secret {
		t.Error("the chosen card is not on top of the opponent's library")
	}
}

func TestDst5605B2EmissaryOfDespairDrainsPerArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	emissary := b12Push(g, me.ID, "Emissary of Despair", "Creature — Spirit", dstB2EmissaryOfDespairOracle, 2, 1)
	dstB2Artifact(g, opp.ID, "Their Rock")
	dstB2Artifact(g, opp.ID, "Their Other Rock")
	dstB2Artifact(g, me.ID, "My Rock")
	life := opp.Life
	dealCombatDamageToPlayer(g, emissary, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the damaged player has %d life, want %d: one for each of their two artifacts", opp.Life, life-2)
	}
}

func TestDst5605B2EmissaryOfHopeGainsPerArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	emissary := b12Push(g, me.ID, "Emissary of Hope", "Creature — Spirit", dstB2EmissaryOfHopeOracle, 2, 1)
	dstB2Artifact(g, opp.ID, "Their Rock")
	dstB2Artifact(g, opp.ID, "Their Other Rock")
	dstB2Artifact(g, opp.ID, "Their Third Rock")
	life := me.Life
	dealCombatDamageToPlayer(g, emissary, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("the Emissary's controller has %d life, want %d: one for each of their three artifacts", me.Life, life+3)
	}
}

// --- attack triggers -----------------------------------------------------------

func TestDst5605B2RoaringSlagwurmTapsAllArtifactsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wurm := b12Push(g, me.ID, "Roaring Slagwurm", "Creature — Wurm", dstB2RoaringSlagwurmOracle, 6, 4)
	mine := dstB2Artifact(g, me.ID, "My Rock")
	theirs := dstB2Artifact(g, opp.ID, "Their Rock")
	bear := seedBear(g, opp.ID)
	declareAttack(t, g, opp.ID, wurm)
	passPriorityAroundTable(t, g)
	if !isTapped(g, mine) || !isTapped(g, theirs) {
		t.Errorf("artifacts tapped: mine %v, theirs %v; want both", isTapped(g, mine), isTapped(g, theirs))
	}
	if isTapped(g, bear) {
		t.Error("a nonartifact creature was tapped")
	}
}

func TestDst5605B2FangrenFirstbornPutsACounterOnEachAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fangren := b12Push(g, me.ID, "Fangren Firstborn", "Creature — Beast", dstB2FangrenFirstbornOracle, 4, 2)
	other := b12Creature(g, me.ID, "Attacker", "Creature — Human", 2, 2)
	home := b12Creature(g, me.ID, "Stay Home", "Creature — Human", 2, 2)
	declareAttack(t, g, opp.ID, fangren, other)
	passPriorityAroundTable(t, g)
	if countersOn(g, fangren, game.CounterPlusOne) != 1 || countersOn(g, other, game.CounterPlusOne) != 1 {
		t.Errorf("counters: Firstborn %d, other attacker %d; want one each",
			countersOn(g, fangren, game.CounterPlusOne), countersOn(g, other, game.CounterPlusOne))
	}
	if countersOn(g, home, game.CounterPlusOne) != 0 {
		t.Error("a creature that didn't attack got a counter")
	}
}

// --- activated abilities -------------------------------------------------------

func TestDst5605B2LoxodonMysticTapsTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mystic := b12Push(g, me.ID, "Loxodon Mystic", "Creature — Elephant Cleric", dstB2LoxodonMysticOracle, 3, 3)
	bear := seedBear(g, opp.ID)
	b06AddMana(me, "W")
	if err := g.ActivateCatalogAbility(me.ID, mystic, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !isTapped(g, bear) || !isTapped(g, mystic) {
		t.Errorf("bear tapped %v, Mystic tapped %v; want both", isTapped(g, bear), isTapped(g, mystic))
	}
}

func TestDst5605B2ViridianZealotSacrificesToDestroyAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	zealot := b12Push(g, me.ID, "Viridian Zealot", "Creature — Elf Warrior", dstB2ViridianZealotOracle, 2, 1)
	rock := dstB2Artifact(g, opp.ID, "Their Rock")
	b06AddMana(me, "C", "G")
	if err := g.ActivateCatalogAbility(me.ID, zealot, 0, game.ActivateAbilityParams{Targets: cardRefs(rock)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if g.Battlefield.Contains(zealot) {
		t.Error("the Zealot should be sacrificed")
	}
}

func TestDst5605B2GrimclawBatsPaysBlackAndLifeForPlusOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bats := b12Push(g, me.ID, "Grimclaw Bats", "Creature — Bat", dstB2GrimclawBatsOracle, 1, 1)
	life := me.Life
	b06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, bats, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bats); p != 2 || tt != 2 {
		t.Errorf("the Bats are %d/%d, want 2/2", p, tt)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d after paying 1", me.Life, life-1)
	}
}

func TestDst5605B2DarksteelPendantScriesOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pendant := b12Push(g, me.ID, "Darksteel Pendant", "Artifact", dstB2DarksteelPendantOracle, 0, 0)
	pushLibraryCardForTest(me, game.Card{Name: "Library Card", TypeLine: "Sorcery"})
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, pendant, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry prompt")
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !isTapped(g, pendant) {
		t.Error("the Pendant did not tap")
	}
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("the scry prompt is still open after it was answered")
	}
}

func TestDst5605B2DarksteelBruteBecomesATwoTwoBeast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	brute := b12Push(g, me.ID, "Darksteel Brute", "Artifact", dstB2DarksteelBruteOracle, 0, 0)
	if layeredCard(t, g, brute).IsCreature() {
		t.Fatal("the Brute is a creature before it is animated")
	}
	b06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, brute, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !layeredCard(t, g, brute).IsCreature() {
		t.Fatal("the Brute did not become a creature")
	}
	if p, tt := ptOf(t, g, brute); p != 2 || tt != 2 {
		t.Errorf("the Brute is %d/%d, want 2/2", p, tt)
	}
}

func TestDst5605B2VoltaicConstructUntapsAnArtifactCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	construct := b12Push(g, me.ID, "Voltaic Construct", "Artifact Creature — Golem Construct", dstB2VoltaicConstructOracle, 2, 2)
	thopter := b12Creature(g, me.ID, "Thopter", "Artifact Creature — Thopter", 1, 1)
	bear := seedBear(g, me.ID)
	tapForTest(t, g, thopter, bear)
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, construct, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Fatal("a nonartifact creature was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, construct, 0, game.ActivateAbilityParams{Targets: cardRefs(thopter)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, thopter) {
		t.Error("the artifact creature is still tapped")
	}
}

func TestDst5605B2VulshokMorningstarGivesPlusTwoPlusTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	star := b12Push(g, me.ID, "Vulshok Morningstar", "Artifact — Equipment", dstB2VulshokMorningstarOracle, 0, 0)
	bear := seedBear(g, me.ID)
	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, star, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 4 {
		t.Errorf("the equipped bear is %d/%d, want 4/4", p, tt)
	}
}

// --- mana abilities ------------------------------------------------------------

func TestDst5605B2UrGolemsEyeTapsForTwoColourless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	eye := b12Push(g, me.ID, "Ur-Golem's Eye", "Artifact", dstB2UrGolemsEyeOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, eye, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.ManaPool.CountColor("C"); got != 2 {
		t.Errorf("pool holds %d {C}, want 2", got)
	}
}

func TestDst5605B2KrarkClanStokerSacrificesAnArtifactForTwoRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	stoker := b12Push(g, me.ID, "Krark-Clan Stoker", "Creature — Goblin Shaman", dstB2KrarkClanStokerOracle, 2, 2)
	rock := dstB2Artifact(g, me.ID, "Rock")
	bear := seedBear(g, me.ID)
	if err := g.ActivateManaAbility(me.ID, stoker, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a creature that is not an artifact is not a legal sacrifice")
	}
	if err := g.ActivateManaAbility(me.ID, stoker, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := me.ManaPool.CountColor("R"); got != 2 {
		t.Errorf("pool holds %d {R}, want 2", got)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact should be sacrificed")
	}
	if !isTapped(g, stoker) {
		t.Error("the Stoker did not tap")
	}
}

func TestDst5605B2ViridianAcolyteFiltersOneIntoAnyColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	acolyte := b12Push(g, me.ID, "Viridian Acolyte", "Creature — Elf Shaman", dstB2ViridianAcolyteOracle, 1, 1)
	b06AddMana(me, "C")
	if err := g.ActivateManaAbility(me.ID, acolyte, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.CountColor("U") != 1 || me.ManaPool.CountColor("C") != 0 {
		t.Errorf("pool holds %d {U} and %d {C}, want the {C} spent for one {U}",
			me.ManaPool.CountColor("U"), me.ManaPool.CountColor("C"))
	}
}

// --- spells --------------------------------------------------------------------

func TestDst5605B2EssenceDrainDealsThreeAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myLife, theirLife := me.Life, opp.Life
	castCatalogSpell(t, g, "Essence Drain", "Sorcery", dstB2EssenceDrainOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != theirLife-3 || me.Life != myLife+3 {
		t.Errorf("life: target %d (want %d), caster %d (want %d)", opp.Life, theirLife-3, me.Life, myLife+3)
	}
}

func TestDst5605B2NourishGainsSix(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	castCatalogSpell(t, g, "Nourish", "Instant", dstB2NourishOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life+6 {
		t.Errorf("life %d, want %d", me.Life, life+6)
	}
}

func TestDst5605B2OxidizeDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := dstB2Artifact(g, opp.ID, "Their Rock")
	castCatalogSpell(t, g, "Oxidize", "Instant", dstB2OxidizeOracle, cardRefs(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestDst5605B2PurgeDestroysAnArtifactOrBlackCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	zombie := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Zombie", TypeLine: "Creature — Zombie", Power: 2, Toughness: 2, Colors: []string{"B"}})
	elf := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Colors: []string{"G"}})
	if err := castCatalogSpellErr(t, g, "Purge", "Instant", dstB2PurgeOracle, cardRefs(elf)); err == nil {
		t.Fatal("a green nonartifact creature was accepted as the target")
	}
	castCatalogSpell(t, g, "Purge", "Instant", dstB2PurgeOracle, cardRefs(zombie))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(zombie) {
		t.Error("the black creature survived")
	}
}

func TestDst5605B2RitualOfRestorationReturnsAnArtifactCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	art := game.NewCard("Dead Rock", me.ID)
	art.TypeLine = "Artifact"
	me.Graveyard.PushTop(art)
	castCatalogSpell(t, g, "Ritual of Restoration", "Sorcery", dstB2RitualOfRestorationOracle, cardRefs(art.InstanceID))
	passPriorityAroundTable(t, g)
	if zoneOf(g, art.InstanceID) != game.ZoneHand {
		t.Error("the artifact card is not in its owner's hand")
	}
}

func TestDst5605B2SoulscourDestroysEveryNonartifactPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := dstB2Artifact(g, me.ID, "My Rock")
	golem := b12Creature(g, opp.ID, "Their Golem", dstB2GolemType, 3, 3)
	bear := seedBear(g, opp.ID)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	aura := b12Permanent(g, me.ID, "Mood", "Enchantment")
	castCatalogSpell(t, g, "Soulscour", "Sorcery", dstB2SoulscourOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) || !g.Battlefield.Contains(golem) {
		t.Error("an artifact was destroyed")
	}
	for name, id := range map[string]uuid.UUID{"creature": bear, "land": land, "enchantment": aura} {
		if g.Battlefield.Contains(id) {
			t.Errorf("the nonartifact %s survived", name)
		}
	}
}

func TestDst5605B2MetalFatigueTapsAllArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := dstB2Artifact(g, me.ID, "My Rock")
	theirs := b12Creature(g, opp.ID, "Their Golem", dstB2GolemType, 3, 3)
	bear := seedBear(g, opp.ID)
	castCatalogSpell(t, g, "Metal Fatigue", "Instant", dstB2MetalFatigueOracle, nil)
	passPriorityAroundTable(t, g)
	if !isTapped(g, mine) || !isTapped(g, theirs) {
		t.Errorf("artifacts tapped: mine %v, theirs %v; want both", isTapped(g, mine), isTapped(g, theirs))
	}
	if isTapped(g, bear) {
		t.Error("a nonartifact creature was tapped")
	}
}

func TestDst5605B2MagneticFluxGivesYourArtifactCreaturesFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := b12Creature(g, me.ID, "My Golem", dstB2GolemType, 3, 3)
	bear := seedBear(g, me.ID)
	theirs := b12Creature(g, opp.ID, "Their Golem", dstB2GolemType, 3, 3)
	castCatalogSpell(t, g, "Magnetic Flux", "Instant", dstB2MagneticFluxOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, golem, "flying") {
		t.Error("your artifact creature has no flying")
	}
	if hasEffectiveKeyword(t, g, bear, "flying") || hasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("a nonartifact creature or an opponent's artifact creature gained flying")
	}
}

func TestDst5605B2HungerOfTheNimPumpsPerArtifactYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := seedBear(g, me.ID)
	dstB2Artifact(g, me.ID, "My Rock")
	dstB2Artifact(g, me.ID, "My Other Rock")
	dstB2Artifact(g, opp.ID, "Their Rock")
	castCatalogSpell(t, g, "Hunger of the Nim", "Sorcery", dstB2HungerOfTheNimOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 4 || tt != 2 {
		t.Errorf("the bear is %d/%d, want 4/2: +1/+0 for each of your two artifacts", p, tt)
	}
}

func TestDst5605B2BurdenOfGreedCountsTappedArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := dstB2Artifact(g, opp.ID, "Their Rock")
	b := dstB2Artifact(g, opp.ID, "Their Other Rock")
	dstB2Artifact(g, opp.ID, "Their Untapped Rock")
	tapForTest(t, g, a, b)
	life := opp.Life
	castCatalogSpell(t, g, "Burden of Greed", "Instant", dstB2BurdenOfGreedOracle, playerRefs(opp.ID))
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the target has %d life, want %d: one for each of two tapped artifacts", opp.Life, life-2)
	}
}

func TestDst5605B2RetractReturnsOnlyYourArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := dstB2Artifact(g, me.ID, "My Rock")
	theirs := dstB2Artifact(g, opp.ID, "Their Rock")
	bear := seedBear(g, me.ID)
	castCatalogSpell(t, g, "Retract", "Instant", dstB2RetractOracle, nil)
	passPriorityAroundTable(t, g)
	if zoneOf(g, mine) != game.ZoneHand {
		t.Error("your artifact is not in your hand")
	}
	if !g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(bear) {
		t.Error("an opponent's artifact or a nonartifact permanent left the battlefield")
	}
}
