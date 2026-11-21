package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b34_cards_test.go — Avatar: The Last Airbender Eternal batches 34
// to 37 (#5075).

const (
	avB34SalvagersOracle  = "9e76d56a-67b6-4bd9-8f2d-3328d18a2880"
	avB34NylaOracle       = "3492cee8-b854-449a-ada4-a4e8a4333d87"
	avB34RokuOracle       = "9eb9e4c1-fabf-45ff-a584-fbe4b94fb41c"
	avB34ChongOracle      = "faffc9fb-32b3-4015-8c3f-e6b45002c050"
	avB34FangOracle       = "6665bfe2-942c-4b3d-a976-776fe00c99f4"
	avB34IrohDragonOracle = "98f9b97b-4c15-4d24-895a-a9bebb68e79b"
	avB34VictoryOracle    = "20d94561-7526-4305-8d2d-7daf62729ff4"
	avB34BlazeOracle      = "75137acb-dc8e-439b-8d84-c5cf682ff6bc"
	avB34SmellerbeeOracle = "127bc05b-2945-434c-bc6a-283a267cc72a"
	avB34KyoshiOracle     = "fd751ed4-4002-4cae-9d1d-1fe0f1fd4965"
	avB34HeiBaiOracle     = "11e70692-2057-49c8-90e6-953362f4817f"
	avB34GuidanceOracle   = "66de5b97-0de0-422c-a351-bfe55ad4503a"
	avB34TaleOracle       = "7c376ea9-d50a-4cc3-ae40-7bac3a926f20"
	avB34TeamworkOracle   = "ceddea59-de47-4ddd-b424-6e1dcb13ab79"
	avB34AppaOracle       = "215215b8-6ddc-4e94-8640-6d324ff65f9c"
	avB34KataraOracle     = "703032eb-98b3-4ecf-8b92-4f4f1551c5d7"
	avB34GuardOracle      = "18fe834f-4417-4df8-a49c-98e9bf2f87d2"
	avB34MomoOracle       = "8854222e-590f-4fc0-b5d7-bf3d986b8507"
	avB34OtterOracle      = "1b115051-a3c8-4c5f-9ad2-82b5a6897c51"
	avB34SokkaOracle      = "7dd518f9-1ddd-4950-add3-fc21f675f4dc"
	avB34TundraOracle     = "68df4c46-513a-4d75-9d4b-ac95d45be255"
	avB34VillagerOracle   = "549a3f53-f90e-4af0-bbc5-2c4b9df30008"
	avB34DenyOracle       = "a7dad8fe-b0b8-4c6d-ad11-76b402952799"
	avB34DolphinOracle    = "59fa8323-bba9-4c51-833d-2ab966599b62"
	avB34LostOracle       = "a37adc18-ea11-4972-8d04-936cf3fe2613"
	avB34SerpentOracle    = "771bf51d-cb32-4240-936c-f9d1e0c90cf0"
	avB34SealsOracle      = "e0172cad-8b93-4d1a-b442-5eced3212b17"
	avB34WhipOracle       = "2015f455-988e-444c-93df-23fab179ef87"
	avB34RatOracle        = "84557d9d-ee22-46c9-8f4a-62a440fb302f"
	avB34AmbushersOracle  = "8ed27504-f86b-4ce5-9395-7029825176ab"
	avB34SentinelsOracle  = "4ba60296-5612-41ed-bf53-a061783ed3b9"
	avB34GilacornOracle   = "1551cd95-1e68-404f-be1d-cc9c7f447210"
	avB34VultureOracle    = "169d24f3-a308-4a2e-ae25-85dae448994e"
	avB34PentapusOracle   = "37d2e83d-f5fd-48d2-b091-3c9167522530"
	avB34CapitalOracle    = "aff20717-265d-4214-a8f5-f520bb243cb5"
	avB34ArchersOracle    = "2e735849-1930-4f43-86b4-da660ade15bc"
	avB34IrohOracle       = "8e03797d-dd61-439f-b4ce-996cd71bdcf5"
	avB34SageOracle       = "0c052420-2d8b-4bbf-8197-b8a36e7208c9"
	avB34MasteryOracle    = "e5659c45-e064-4690-b90a-24ffe3793d4d"
	avB34ScoutOracle      = "f07e6ba7-42e8-4b4d-9831-00a856059060"
)

func TestAvatarB34EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB34SalvagersOracle:  "Fire Nation Salvagers",
		avB34NylaOracle:       "Nyla, Shirshu Sleuth",
		avB34RokuOracle:       "Avatar Roku, Firebender",
		avB34ChongOracle:      "Chong and Lily, Nomads",
		avB34FangOracle:       "Fang, Roku's Companion",
		avB34IrohDragonOracle: "Iroh, Dragon of the West",
		avB34VictoryOracle:    "Overwhelming Victory",
		avB34BlazeOracle:      "Reckless Blaze",
		avB34SmellerbeeOracle: "Smellerbee, Rebel Fighter",
		avB34KyoshiOracle:     "Avatar Kyoshi, Earthbender",
		avB34HeiBaiOracle:     "Hei Bai, Forest Guardian",
		avB34GuidanceOracle:   "Master's Guidance",
		avB34TaleOracle:       "Tale of Katara and Toph",
		avB34TeamworkOracle:   "Allied Teamwork",
		avB34AppaOracle:       "Appa, Aang's Companion",
		avB34KataraOracle:     "Katara, Heroic Healer",
		avB34GuardOracle:      "Kyoshi Warrior Guard",
		avB34MomoOracle:       "Momo, Rambunctious Rascal",
		avB34OtterOracle:      "Sledding Otter-Penguin",
		avB34SokkaOracle:      "Sokka, Wolf Cove's Protector",
		avB34TundraOracle:     "Tundra Wall",
		avB34VillagerOracle:   "Wolf Cove Villager",
		avB34DenyOracle:       "Deny Entry",
		avB34DolphinOracle:    "Flying Dolphin-Fish",
		avB34LostOracle:       "Lost in the Spirit World",
		avB34SerpentOracle:    "The Terror of Serpent's Pass",
		avB34SealsOracle:      "Turtle-Seals",
		avB34WhipOracle:       "Water Whip",
		avB34RatOracle:        "Elephant-Rat",
		avB34AmbushersOracle:  "Fire Nation Ambushers",
		avB34SentinelsOracle:  "Fire Nation Sentinels",
		avB34GilacornOracle:   "Gilacorn",
		avB34VultureOracle:    "Lion Vulture",
		avB34PentapusOracle:   "Purple Pentapus",
		avB34CapitalOracle:    "Capital Guard",
		avB34ArchersOracle:    "Fire Nation Archers",
		avB34IrohOracle:       "Iroh, Firebending Instructor",
		avB34SageOracle:       "Loyal Fire Sage",
		avB34MasteryOracle:    "Roku's Mastery",
		avB34ScoutOracle:      "Warship Scout",
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

// --- the vanilla and keyword-only creatures -------------------------------------

func TestAvatarB34KeywordAndVanillaCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, k := range []struct {
		name, typeLine, oracle, kw string
	}{
		{"Sokka, Wolf Cove's Protector", "Legendary Creature — Human Warrior Ally", avB34SokkaOracle, "vigilance"},
		{"Tundra Wall", "Creature — Wall", avB34TundraOracle, "defender"},
		{"Flying Dolphin-Fish", "Creature — Whale Fish", avB34DolphinOracle, "flying"},
		{"The Terror of Serpent's Pass", "Legendary Creature — Serpent", avB34SerpentOracle, "hexproof"},
		{"Turtle-Seals", "Creature — Turtle Seal", avB34SealsOracle, "vigilance"},
		{"Elephant-Rat", "Creature — Elephant Rat", avB34RatOracle, "menace"},
		{"Fire Nation Ambushers", "Creature — Human Soldier", avB34AmbushersOracle, "flash"},
		{"Gilacorn", "Creature — Lizard", avB34GilacornOracle, "deathtouch"},
	} {
		effectstest.AssertKeywords(t, g, effectstest.B12Push(g, me.ID, k.name, k.typeLine, k.oracle, 2, 2), k.kw)
	}
	for _, v := range []struct{ name, typeLine, oracle string }{
		{"Kyoshi Warrior Guard", "Creature — Human Warrior Ally", avB34GuardOracle},
		{"Capital Guard", "Creature — Human Soldier", avB34CapitalOracle},
		{"Warship Scout", "Creature — Human Scout", avB34ScoutOracle},
	} {
		id := effectstest.B12Push(g, me.ID, v.name, v.typeLine, v.oracle, 2, 2)
		if c := effectstest.LayeredCard(t, g, id); len(c.Keywords) != 0 {
			t.Errorf("%s is vanilla but has %v", v.name, c.Keywords)
		}
		if spec, _ := effects.Lookup(v.oracle); len(spec.Triggered) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 {
			t.Errorf("%s carries abilities", v.name)
		}
	}
}

func TestAvatarB34WolfCoveVillagerEntersTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.AvB1ResolveCreature(t, g, me, "Wolf Cove Villager", "Creature — Human Peasant", avB34VillagerOracle)
	if !effectstest.BattlefieldCardPtr(t, g, id).Tapped {
		t.Error("Wolf Cove Villager entered untapped")
	}
}

// --- Appa, Aang's Companion ----------------------------------------------------------

func TestAvatarB34AppaGivesAnotherAttackerWithoutFlyingFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	appa := effectstest.B12Push(g, me.ID, "Appa, Aang's Companion", "Legendary Creature — Bison Ally", avB34AppaOracle, 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, appa, bear)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, bad := range []uuid.UUID{appa, home} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Errorf("%s was accepted: Appa himself or a creature not attacking", bad)
		}
	}
	effectstest.AvB34Pick(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AvB22Has(t, g, bear, "flying") {
		t.Error("the attacking Bear did not gain flying")
	}
}

// --- Katara, Heroic Healer -------------------------------------------------------------

func TestAvatarB34KataraPutsACounterOnEachOtherCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	katara := effectstest.AvB1ResolveCreature(t, g, me, "Katara, Heroic Healer", "Legendary Creature — Human Warrior Ally", avB34KataraOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("your other creature has %d counters, want 1", n)
	}
	if effectstest.CountersOn(g, katara, game.CounterPlusOne) != 0 || effectstest.CountersOn(g, theirs, game.CounterPlusOne) != 0 {
		t.Error("Katara or an opponent's creature got a counter")
	}
	effectstest.AssertKeywords(t, g, katara, "lifelink")
}

// --- Momo, Rambunctious Rascal -----------------------------------------------------------

func TestAvatarB34MomoDealsFourToATappedOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	big := effectstest.B12Creature(g, opp.ID, "Tapped Giant", "Creature — Giant", 6, 6)
	effectstest.BattlefieldCardPtr(t, g, big).Tapped = true
	untapped := effectstest.B12Creature(g, opp.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	momo := effectstest.CastAndResolveCreature(t, g, "Momo, Rambunctious Rascal", "Legendary Creature — Lemur Bat Ally", avB34MomoOracle)
	effectstest.AssertKeywords(t, g, momo, "flying")
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: untapped}); err == nil {
		t.Error("an untapped creature was accepted")
	}
	effectstest.AvB34Pick(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCardPtr(t, g, big).DamageMarked; got != 4 {
		t.Errorf("damage = %d, want 4", got)
	}
}

// --- Sledding Otter-Penguin --------------------------------------------------------------

func TestAvatarB34OtterPenguinGrowsForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	otter := effectstest.B12Push(g, me.ID, "Sledding Otter-Penguin", "Creature — Otter Bird", avB34OtterOracle, 2, 3)
	effectstest.B09AddColorless(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, otter, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for two")
	}
	effectstest.B09AddColorless(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, otter, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, otter, game.CounterPlusOne); n != 1 {
		t.Errorf("counters = %d, want 1", n)
	}
}

// --- Allied Teamwork ---------------------------------------------------------------------

func TestAvatarB34AlliedTeamworkMakesAnAllyAndPumpsAllies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirAlly := effectstest.B12Creature(g, opp.ID, "Their Ally", "Creature — Human Ally", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Allied Teamwork", "Enchantment", avB34TeamworkOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	allies := effectstest.AvB34Tokens(g, me.ID, "Ally")
	if len(allies) != 1 {
		t.Fatalf("Ally tokens = %d, want 1", len(allies))
	}
	if c := effectstest.LayeredCard(t, g, allies[0].InstanceID); c.CurrentPower() != 2 || c.CurrentToughness() != 2 || !c.HasColor("W") {
		t.Errorf("the Ally is %d/%d %v, want a white 1/1 pumped to 2/2", c.CurrentPower(), c.CurrentToughness(), c.Colors)
	}
	if c := effectstest.LayeredCard(t, g, bear); c.CurrentPower() != 2 {
		t.Errorf("a non-Ally got pumped to %d", c.CurrentPower())
	}
	if c := effectstest.LayeredCard(t, g, theirAlly); c.CurrentPower() != 2 {
		t.Errorf("an opponent's Ally got pumped to %d", c.CurrentPower())
	}
}

// --- Deny Entry ------------------------------------------------------------------------

func TestAvatarB34DenyEntryCountersACreatureSpellAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetPlayer(me.ID))
	if err := effectstest.CastCatalogSpellErr(t, g, "Deny Entry", "Instant", avB34DenyOracle, effectstest.B16TargetCard(bolt)); err == nil {
		t.Fatal("Deny Entry targeted an instant spell")
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	bear := effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.CastCatalogSpell(t, g, "Deny Entry", "Instant", avB34DenyOracle, effectstest.B16TargetCard(bear))
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Fatal("the creature spell was not countered")
	}
	if me.Hand.Size() != hand+1 || effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Errorf("drew %d with %d to discard, want 1 and 1", me.Hand.Size()-hand, effectstest.DiscardOwed(g, me.ID))
	}
}

// --- Lost in the Spirit World ----------------------------------------------------------

func TestAvatarB34LostInTheSpiritWorldBouncesAndMakesASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lost in the Spirit World", "Sorcery", avB34LostOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the creature did not go to its owner's hand")
	}
	spirits := effectstest.AvB34Tokens(g, me.ID, "Spirit")
	if len(spirits) != 1 || spirits[0].OracleID != effects.KurukSpiritToken().OracleID {
		t.Fatalf("Spirits = %+v, want one of Avatar Kuruk's printed row", spirits)
	}

	// With no target chosen it still makes the Spirit.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Lost in the Spirit World", "Sorcery", avB34LostOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.AvB34Tokens(g, me.ID, "Spirit")) != 1 {
		t.Error("no Spirit with no target chosen")
	}
}

// --- Water Whip --------------------------------------------------------------------------

func TestAvatarB34WaterWhipWaterbendsFiveBouncesTwoAndDrawsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	tappers := effectstest.PushTapCostSoldiers(g, me.ID, 3)
	effectstest.FillPoolColored(me, "U", 2)
	effectstest.B09AddColorless(me, 1)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Water Whip", TypeLine: "Sorcery — Lesson", ManaCost: "{U}{U}",
			OracleID: avB34WhipOracle, Owner: me.ID, Controller: me.ID})
	})
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: tappers, Targets: targets}); err == nil {
		t.Fatal("cast with only four of the five waterbend paid")
	}
	effectstest.B09AddColorless(me, 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, TapIDs: tappers, Targets: targets}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(a) || !opp.Hand.Contains(b) {
		t.Error("the two creatures were not returned")
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("drew %d, want 2", me.Hand.Size()-hand)
	}
}

// --- Fire Nation Sentinels -----------------------------------------------------------------

func TestAvatarB34SentinelsGrowTheTeamWhenAnOpposingNontokenCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sentinels := effectstest.B12Push(g, me.ID, "Fire Nation Sentinels", "Creature — Human Soldier", avB34SentinelsOracle, 4, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Token", TypeLine: "Token Creature — Bear",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	effectstest.B27Kill(g, token)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, bear, game.CounterPlusOne) != 0 {
		t.Fatal("a token's death grew the team")
	}
	mine := effectstest.B12Creature(g, me.ID, "My Doomed Bear", "Creature — Bear", 2, 2)
	effectstest.B27Kill(g, mine)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, bear, game.CounterPlusOne) != 0 {
		t.Fatal("your own creature's death grew the team")
	}
	effectstest.B27Kill(g, theirs)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, sentinels, game.CounterPlusOne) != 1 {
		t.Errorf("counters = %d and %d, want 1 on each creature you control",
			effectstest.CountersOn(g, bear, game.CounterPlusOne), effectstest.CountersOn(g, sentinels, game.CounterPlusOne))
	}
}

// --- Lion Vulture ---------------------------------------------------------------------------

func TestAvatarB34LionVultureGrowsAndDrawsOnlyAfterAnOpponentLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vulture := effectstest.B12Push(g, me.ID, "Lion Vulture", "Creature — Cat Bird", avB34VultureOracle, 2, 2)
	effectstest.AssertKeywords(t, g, vulture, "flying")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, vulture, game.CounterPlusOne) != 0 {
		t.Fatal("it grew with no opponent having lost life")
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vulture = effectstest.B12Push(g, me.ID, "Lion Vulture", "Creature — Cat Bird", avB34VultureOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })
	effectstest.AdvanceTo(t, g, game.StepEnd)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, vulture, game.CounterPlusOne) != 1 || me.Hand.Size() != hand+1 {
		t.Errorf("counters %d, drew %d; want 1 and 1", effectstest.CountersOn(g, vulture, game.CounterPlusOne), me.Hand.Size()-hand)
	}
}

// --- Purple Pentapus ------------------------------------------------------------------------

func TestAvatarB34PentapusSurveilsAndReturnsFromTheGraveyardTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AvB1ResolveCreature(t, g, me, "Purple Pentapus", "Creature — Octopus Starfish", avB34PentapusOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID); c == nil {
		t.Fatal("no surveil prompt")
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Purple Pentapus", TypeLine: "Creature — Octopus Starfish",
			OracleID: avB34PentapusOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	effectstest.FillPoolColored(me, "B", 1)
	effectstest.B09AddColorless(me, 2)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no untapped creature to tap")
	}
	helper := effectstest.B12Creature(g, me.ID, "Helper", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{helper}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) || !effectstest.BattlefieldCardPtr(t, g, id).Tapped {
		t.Error("the Pentapus did not return tapped")
	}
	if !effectstest.BattlefieldCardPtr(t, g, helper).Tapped {
		t.Error("the creature tapped for the cost is untapped")
	}
}

// --- Fire Nation Archers --------------------------------------------------------------------

func TestAvatarB34ArchersPingEachOpponentAndMakeASoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	archers := effectstest.B12Push(g, me.ID, "Fire Nation Archers", "Creature — Human Archer", avB34ArchersOracle, 3, 4)
	effectstest.AssertKeywords(t, g, archers, "reach")
	effectstest.B09AddColorless(me, 5)
	if err := g.ActivateCatalogAbility(me.ID, archers, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := 40
		if p.ID != me.ID {
			want = 38
		}
		if p.Life != want {
			t.Errorf("%s life %d, want %d", p.Name, p.Life, want)
		}
	}
	soldiers := effectstest.AvB34Tokens(g, me.ID, "Soldier")
	if len(soldiers) != 1 || soldiers[0].Power != 2 || soldiers[0].Toughness != 2 || !soldiers[0].HasColor("R") {
		t.Errorf("soldiers = %+v, want one 2/2 red Soldier", soldiers)
	}
	if def, _ := effects.Lookup(avB34ArchersOracle); len(def.Triggered) != 0 {
		t.Error("the Archers have a trigger")
	}
}

// --- Iroh, Firebending Instructor -----------------------------------------------------------

func TestAvatarB34IrohInstructorPumpsTheAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	iroh := effectstest.B12Push(g, me.ID, "Iroh, Firebending Instructor", "Legendary Creature — Human Noble Ally", avB34IrohOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, iroh, bear)
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{iroh: 3, bear: 3, home: 2} {
		if c := effectstest.LayeredCard(t, g, id); c.CurrentPower() != want || c.CurrentToughness() != want {
			t.Errorf("%s is %d/%d, want %d/%d", c.Name, c.CurrentPower(), c.CurrentToughness(), want, want)
		}
	}
}

// --- Loyal Fire Sage ------------------------------------------------------------------------

func TestAvatarB34LoyalFireSageFirebendsOneAndMakesAnAlly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sage := effectstest.B12Push(g, me.ID, "Loyal Fire Sage", "Creature — Human Cleric Ally", avB34SageOracle, 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, sage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 1})
	effectstest.MkAdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.MkWantPool(t, me, map[string]int{})
	effectstest.B09AddColorless(me, 5)
	if err := g.ActivateCatalogAbility(me.ID, sage, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.AvB34Tokens(g, me.ID, "Ally")) != 1 {
		t.Error("no Ally token")
	}
}

// --- Roku's Mastery -------------------------------------------------------------------------

func TestAvatarB34RokusMasteryScriesOnlyAtFourOrMore(t *testing.T) {
	for _, x := range []int{3, 4} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.AdvanceToMain(t, g)
		big := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 6, 6)
		effectstest.FillPoolColored(me, "R", 2)
		effectstest.B09AddColorless(me, x)
		effectstest.CastFromHandForTest(t, g, me, "Roku's Mastery", "Instant", "{X}{R}{R}", avB34MasteryOracle, game.CastSpellParams{
			Strict: true, XValue: x, Targets: effectstest.B16TargetCard(big),
		})
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.BattlefieldCardPtr(t, g, big).DamageMarked; got != x {
			t.Errorf("X=%d: damage %d", x, got)
		}
		scry := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID)
		if (scry != nil) != (x >= 4) {
			t.Errorf("X=%d: scry prompt %v", x, scry != nil)
		}
		if scry != nil && len(scry.ScryCards) != 2 {
			t.Errorf("scry %d cards, want 2", len(scry.ScryCards))
		}
	}
}

// --- Avatar Roku, Firebender ------------------------------------------------------------

func TestAvatarB34RokuAddsSixRedWhenAnyPlayerAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Roku is the opponent's: the active player's attack triggers it too.
	effectstest.B12Push(g, opp.ID, "Avatar Roku, Firebender", "Legendary Creature — Human Avatar", avB34RokuOracle, 6, 6)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, opp, map[string]int{"R": 6})
	effectstest.MkWantPool(t, me, map[string]int{})
	effectstest.MkAdvanceTo(t, g, game.StepEndCombat)
	effectstest.MkWantPool(t, opp, map[string]int{"R": 6})
	effectstest.MkAdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.MkWantPool(t, opp, map[string]int{})
}

func TestAvatarB34RokuPumpsATargetForThreeRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	roku := effectstest.B12Push(g, me.ID, "Avatar Roku, Firebender", "Legendary Creature — Human Avatar", avB34RokuOracle, 6, 6)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FillPoolColored(me, "R", 2)
	effectstest.B09AddColorless(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, roku, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Fatal("activated without three red")
	}
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, roku, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, bear); c.CurrentPower() != 5 || c.CurrentToughness() != 2 {
		t.Errorf("the Bear is %d/%d, want 5/2", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Master's Guidance -------------------------------------------------------------------

func TestAvatarB34MastersGuidanceNeedsTwoLegendaryAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Master's Guidance", "Enchantment", avB34GuidanceOracle, 0, 0)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, legend, bear)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("one legendary attacker triggered it")
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Master's Guidance", "Enchantment", avB34GuidanceOracle, 0, 0)
	l1 := effectstest.B12Creature(g, me.ID, "Legend One", "Legendary Creature — Human", 2, 2)
	l2 := effectstest.B12Creature(g, me.ID, "Legend Two", "Legendary Creature — Human", 2, 2)
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, l1, l2, bear)
	effectstest.AvB34Pick(t, g, me.ID, l1, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, l1, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, l2, game.CounterPlusOne) != 0 {
		t.Errorf("counters %d/%d/%d, want 1 on each of the two targets", effectstest.CountersOn(g, l1, game.CounterPlusOne),
			effectstest.CountersOn(g, bear, game.CounterPlusOne), effectstest.CountersOn(g, l2, game.CounterPlusOne))
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a second trigger for the same declaration")
	}
}

func TestAvatarB34MastersGuidanceDrawsWithAFourPowerCreature(t *testing.T) {
	for _, power := range []int{3, 4} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.B12Push(g, me.ID, "Master's Guidance", "Enchantment", avB34GuidanceOracle, 0, 0)
		effectstest.B12Creature(g, me.ID, "Beast", "Creature — Beast", power, power)
		effectstest.AdvanceToMain(t, g)
		effectstest.AdvanceTo(t, g, game.StepEnd)
		hand := me.Hand.Size()
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if power >= 4 {
			want = 1
		}
		if got := me.Hand.Size() - hand; got != want {
			t.Errorf("power %d: drew %d, want %d", power, got, want)
		}
	}
}

// --- Fire Nation Salvagers ---------------------------------------------------------------

func TestAvatarB34SalvagersCounterOnEntryAndReanimateFromTheDamagedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirCorpse := effectstest.AvB24Graveyard(opp, game.Card{Name: "Their Dead Giant", TypeLine: "Creature — Giant", Power: 5, Toughness: 5})
	otherSeat := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	elsewhere := effectstest.AvB24Graveyard(otherSeat, game.Card{Name: "Someone Else's Dead", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	salv := effectstest.AvB1ResolveCreature(t, g, me, "Fire Nation Salvagers", "Creature — Human Soldier", avB34SalvagersOracle)
	effectstest.AssertKeywords(t, g, salv, "menace")
	effectstest.AvB34Pick(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Fatal("no counter from the enters trigger")
	}
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, bear, plain)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no reanimation target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: elsewhere}); err == nil {
		t.Error("a card from another player's graveyard was accepted")
	}
	effectstest.AvB34Pick(t, g, me.ID, theirCorpse)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirCorpse) || effectstest.BattlefieldCardPtr(t, g, theirCorpse).Controller != me.ID {
		t.Error("the damaged player's creature card did not come back under your control")
	}
}

func TestAvatarB34SalvagersNeedACreatureWithACounterToConnect(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Fire Nation Salvagers", "Creature — Human Soldier", avB34SalvagersOracle, 3, 3)
	effectstest.AvB24Graveyard(opp, game.Card{Name: "Their Dead Giant", TypeLine: "Creature — Giant", Power: 5, Toughness: 5})
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a creature with no counters triggered the reanimation")
	}
}

// --- Nyla, Shirshu Sleuth -----------------------------------------------------------------

func TestAvatarB34NylaTradesLifeForCluesThenReturnsTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dead := effectstest.AvB24Graveyard(me, game.Card{Name: "Dead Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}", Power: 3, Toughness: 3})
	life := me.Life
	effectstest.AvB1ResolveCreature(t, g, me, "Nyla, Shirshu Sleuth", "Legendary Creature — Mole Beast", avB34NylaOracle)
	effectstest.AvB34Pick(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(dead); z == nil || z.Kind != game.ZoneExile {
		t.Fatal("the creature card was not exiled")
	}
	clues := effectstest.AvB34Tokens(g, me.ID, "Clue")
	if me.Life != life-3 || len(clues) != 3 {
		t.Fatalf("life %d (want %d), Clues %d (want 3)", me.Life, life-3, len(clues))
	}
	// With Clues on the battlefield the end-step trigger does nothing.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(dead) {
		t.Fatal("returned the card while you controlled Clues")
	}
	for _, c := range clues {
		effectstest.B27Kill(g, c.InstanceID)
	}
	g.RunStateChecksForTest()
	effectstest.B39NextTurnOf(t, g, g.Turn.ActiveSeat)
	for g.Turn.ActiveSeat != 0 {
		effectstest.B39NextTurnOf(t, g, g.Turn.ActiveSeat)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.AvB34Pick(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the exiled card did not return to its owner's hand")
	}
}

// --- Chong and Lily, Nomads ---------------------------------------------------------------

func TestAvatarB34ChongAndLilyLoreCountersOrPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chong := effectstest.B12Push(g, me.ID, "Chong and Lily, Nomads", "Legendary Creature — Human Bard Ally", avB34ChongOracle, 3, 3)
	saga := effectstest.PushPermanent(g, me.ID, game.Card{Name: "A Saga", TypeLine: "Enchantment — Saga", Counters: map[string]int{game.CounterLore: 2}})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// The precombat main phase adds a lore counter of its own (CR 714.3b).
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	lore := effectstest.CountersOn(g, saga, game.CounterLore)
	effectstest.DeclareAttack(t, g, opp.ID, chong, bear)
	mode := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, bear); c.CurrentPower() != 2+lore {
		t.Errorf("the Bear has power %d, want 2 + %d lore counters", c.CurrentPower(), lore)
	}
}

func TestAvatarB34ChongAndLilyPutsLoreCountersOnTargetSagas(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chong := effectstest.B12Push(g, me.ID, "Chong and Lily, Nomads", "Legendary Creature — Human Bard Ally", avB34ChongOracle, 3, 3)
	s1 := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Saga One", TypeLine: "Enchantment — Saga", Counters: map[string]int{game.CounterLore: 1}})
	s2 := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Saga Two", TypeLine: "Enchantment — Saga", Counters: map[string]int{game.CounterLore: 1}})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	l1, l2 := effectstest.CountersOn(g, s1, game.CounterLore), effectstest.CountersOn(g, s2, game.CounterLore)
	effectstest.DeclareAttack(t, g, opp.ID, chong)
	mode := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mode == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	effectstest.AvB34Pick(t, g, me.ID, s1, s2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, s1, game.CounterLore) != l1+1 || effectstest.CountersOn(g, s2, game.CounterLore) != l2+1 {
		t.Errorf("lore counters %d and %d, want one more on each", effectstest.CountersOn(g, s1, game.CounterLore), effectstest.CountersOn(g, s2, game.CounterLore))
	}
}

func TestAvatarB34ChongAndLilyIgnoreAnAttackWithoutABard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Chong and Lily, Nomads", "Legendary Creature — Human Bard Ally", avB34ChongOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) != nil {
		t.Error("an attack with no Bard triggered it")
	}
}

// --- Fang, Roku's Companion ----------------------------------------------------------------

func TestAvatarB34FangPumpsAnotherLegendByHisPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fang := effectstest.B12Push(g, me.ID, "Fang, Roku's Companion", "Legendary Creature — Dragon", avB34FangOracle, 4, 4)
	effectstest.AssertKeywords(t, g, fang, "flying")
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, fang)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, bad := range []uuid.UUID{fang, bear} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Errorf("%s accepted: Fang himself or a nonlegendary creature", bad)
		}
	}
	effectstest.AvB34Pick(t, g, me.ID, legend)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, legend); c.CurrentPower() != 6 || c.CurrentToughness() != 2 {
		t.Errorf("the legend is %d/%d, want 6/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestAvatarB34FangReturnsOnceAsASpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fang := effectstest.B12Push(g, me.ID, "Fang, Roku's Companion", "Legendary Creature — Dragon", avB34FangOracle, 4, 4)
	effectstest.B27Kill(g, fang)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fang) {
		t.Fatal("Fang did not return")
	}
	c := effectstest.LayeredCard(t, g, fang)
	if !c.HasSubtype("Spirit") || !c.HasSubtype("Dragon") {
		t.Fatalf("the returned Fang is %q, want a Dragon Spirit", c.TypeLine)
	}
	effectstest.B27Kill(g, fang)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fang) || !me.Graveyard.Contains(fang) {
		t.Error("Fang the Spirit returned again")
	}
}

// --- Iroh, Dragon of the West --------------------------------------------------------------

func TestAvatarB34IrohDragonGivesFirebendingToCreaturesWithCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	iroh := effectstest.B12Push(g, me.ID, "Iroh, Dragon of the West", "Legendary Creature — Human Noble Ally", avB34IrohDragonOracle, 4, 4)
	effectstest.AssertKeywords(t, g, iroh, "haste")
	countered := effectstest.B12Creature(g, me.ID, "Countered Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _, _ = g.AddCounterMustSettleNowForEffect(countered, game.CounterPlusOne, 1) })
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	// Only the countered Bear gains firebending 2; mentor may pick either
	// Bear (both have less power than Iroh).
	effectstest.DeclareAttack(t, g, opp.ID, iroh, countered, plain)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.AvB34Pick(t, g, me.ID, plain)
	}
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	// One firebending 2: the countered Bear's. Iroh and the plain Bear
	// had no counter as the trigger resolved.
	effectstest.MkWantPool(t, me, map[string]int{"R": 2})
	// Mentor put one counter on an attacker with lesser power.
	if got := effectstest.CountersOn(g, plain, game.CounterPlusOne) + effectstest.CountersOn(g, countered, game.CounterPlusOne); got != 2 {
		t.Errorf("counters among the two Bears = %d, want the 1 they had plus mentor's 1", got)
	}
}

// --- Overwhelming Victory ------------------------------------------------------------------

func TestAvatarB34OverwhelmingVictoryTurnsExcessIntoPowerAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Other Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Overwhelming Victory", "Instant — Lesson", avB34VictoryOracle, effectstest.B16TargetCard(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 2/2 survived 5 damage")
	}
	if c := effectstest.LayeredCard(t, g, mine); c.CurrentPower() != 5 || !game.HasKeyword(&c, "trample") {
		t.Errorf("your creature is %d power, trample %v; want +3/+0 and trample", c.CurrentPower(), game.HasKeyword(&c, "trample"))
	}
	if c := effectstest.LayeredCard(t, g, theirs); c.CurrentPower() != 2 || game.HasKeyword(&c, "trample") {
		t.Error("an opponent's creature was pumped")
	}
}

func TestAvatarB34OverwhelmingVictoryWithoutExcessStillGivesTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Giant", "Creature — Giant", 6, 6)
	effectstest.CastCatalogSpell(t, g, "Overwhelming Victory", "Instant — Lesson", avB34VictoryOracle, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, mine); c.CurrentPower() != 2 || !game.HasKeyword(&c, "trample") {
		t.Errorf("your creature is %d power, trample %v; want 2 and trample", c.CurrentPower(), game.HasKeyword(&c, "trample"))
	}
}

// --- Reckless Blaze ---------------------------------------------------------------------------

func TestAvatarB34RecklessBlazeAddsRedForYourCreaturesThatDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	survivor := effectstest.B12Creature(g, me.ID, "Giant", "Creature — Giant", 7, 7)
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Reckless Blaze", "Sorcery — Lesson", avB34BlazeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 2})
	effectstest.MkWantPool(t, opp, map[string]int{})
	if got := effectstest.BattlefieldCardPtr(t, g, survivor).DamageMarked; got != 5 {
		t.Errorf("the 7/7 has %d damage, want 5", got)
	}
	effectstest.B27Kill(g, survivor)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MkWantPool(t, me, map[string]int{"R": 3})
}

// --- Smellerbee, Rebel Fighter -----------------------------------------------------------------

func TestAvatarB34SmellerbeeHastesTheTeamAndRedrawsForTheAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	smellerbee := effectstest.B12Push(g, me.ID, "Smellerbee, Rebel Fighter", "Legendary Creature — Human Rebel Ally", avB34SmellerbeeOracle, 3, 3)
	effectstest.AssertKeywords(t, g, smellerbee, "first strike")
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !effectstest.AvB22Has(t, g, a, "haste") || effectstest.AvB22Has(t, g, theirs, "haste") {
		t.Error("haste went to the wrong creatures")
	}
	if c := effectstest.LayeredCard(t, g, smellerbee); game.HasKeyword(&c, "haste") {
		t.Error("Smellerbee gave herself haste")
	}
	effectstest.DeclareAttack(t, g, opp.ID, smellerbee, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if hand == 0 {
		t.Fatal("the test needs a hand to discard")
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand %d after discarding %d, want 3 (one per attacker)", got, hand)
	}
}

// --- Avatar Kyoshi, Earthbender -----------------------------------------------------------------

func TestAvatarB34KyoshiHexproofOnYourTurnAndEarthbendsEight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kyoshi := effectstest.B12Push(g, me.ID, "Avatar Kyoshi, Earthbender", "Legendary Creature — Human Avatar", avB34KyoshiOracle, 6, 6)
	if !effectstest.AvB22Has(t, g, kyoshi, "hexproof") {
		t.Error("no hexproof during your turn")
	}
	land := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.BattlefieldCardPtr(t, g, land).Tapped = true
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AvB34Pick(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, land)
	if !c.IsCreature() || effectstest.CountersOn(g, land, game.CounterPlusOne) != 8 {
		t.Fatalf("the land is creature %v with %d counters, want an 8-counter land creature", c.IsCreature(), effectstest.CountersOn(g, land, game.CounterPlusOne))
	}
	if effectstest.BattlefieldCardPtr(t, g, land).Tapped {
		t.Error("the land was not untapped")
	}
	for g.Turn.ActiveSeat == 0 {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if effectstest.AvB22Has(t, g, kyoshi, "hexproof") {
		t.Error("hexproof on an opponent's turn")
	}
}

// --- Hei Bai, Forest Guardian ----------------------------------------------------------------------

func TestAvatarB34HeiBaiFindsAShrineAndMakesSpiritsPerLegendaryEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shrine := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: shrine, Name: "A Shrine", TypeLine: "Legendary Enchantment — Shrine", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "On Top", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	})
	hei := effectstest.AvB1ResolveCreature(t, g, me, "Hei Bai, Forest Guardian", "Legendary Creature — Bear Spirit", avB34HeiBaiOracle)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no put prompt: %+v", g.PendingChoices)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, shrine)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(shrine) {
		t.Fatal("the Shrine was not put onto the battlefield")
	}
	if z := g.FindCardZoneForEffect(hei); z == nil || z.Kind != game.ZoneBattlefield {
		t.Fatal("Hei Bai left")
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "A Shrine", TypeLine: "Legendary Enchantment — Shrine"})
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Another Legend", TypeLine: "Legendary Enchantment"})
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Plain Enchantment", TypeLine: "Enchantment"})
	hei = effectstest.B12Push(g, me.ID, "Hei Bai, Forest Guardian", "Legendary Creature — Bear Spirit", avB34HeiBaiOracle, 4, 4)
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		effectstest.FillPoolColored(me, c, 1)
	}
	if err := g.ActivateCatalogAbility(me.ID, hei, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.AvB34Tokens(g, me.ID, "Spirit")); n != 2 {
		t.Errorf("Spirits = %d, want 2 (one per legendary enchantment)", n)
	}
}

// --- Tale of Katara and Toph -------------------------------------------------------------------------

func TestAvatarB34TaleCountersTheFirstTapOnYourTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Tale of Katara and Toph", "Enchantment", avB34TaleOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Fatalf("the first tap gave %d counters, want 1", effectstest.CountersOn(g, bear, game.CounterPlusOne))
	}
	g.WithWriteLock(func() {
		_ = g.UntapTargetForEffect(bear)
		_ = g.TapTargetForEffect(bear)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Error("a second tap this turn gave another counter")
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, theirs, game.CounterPlusOne) != 0 {
		t.Error("an opponent's creature got a counter")
	}
}
