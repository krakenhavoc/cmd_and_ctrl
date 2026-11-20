package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// khm_5505_b1_cards_test.go — Kaldheim, batch 1 (#5505). Every test
// plays the real catalog entry through the engine.

const (
	khmBattlefieldRaptorOracle   = "68a99e50-0e98-480b-ad25-aa2c226d24ca"
	khmBreakneckBerserkerOracle  = "9f18a580-1c30-4f97-be06-098a763dfda8"
	khmGnottvoldRecluseOracle    = "e7e65cc5-9ceb-43b6-acde-ae492b71daf6"
	khmStorySeekerOracle         = "fd7abcee-d4d3-4cf6-8385-c6ad7ac218e4"
	khmGrizzledOutriderOracle    = "0d3d0da2-73fe-4f8a-b13c-23981eb7e2ee"
	khmAuguryRavenOracle         = "ee05fc58-1356-4b2f-b9ce-bda8cc716aee"
	khmGodsHallGuardianOracle    = "7d22e071-4af4-4182-b387-ecca873d00cc"
	khmVengefulReaperOracle      = "5d776be0-cece-4496-a3b0-b7424eca77fd"
	khmScornEffigyOracle         = "9ef2a604-4053-4ce0-bc08-43d9cc0b15cc"
	khmSulfurousMireOracle       = "77cf536e-246e-4c18-8e04-c904fcad3f40"
	khmVolatileFjordOracle       = "96262c15-9130-409c-8242-71f84be719b7"
	khmDemonBoltOracle           = "6e786f0b-e2cb-40ee-96bc-81945a0b35a3"
	khmDepartTheRealmOracle      = "ae069ba4-8ca6-4eda-8f5d-fc82072d7d40"
	khmDoomskarOracle            = "38ece0fa-2c94-4eae-b6ab-74b783862f3e"
	khmFeedTheSerpentOracle      = "ad404b46-dee7-41bb-969d-35b4cbaadece"
	khmIronVerdictOracle         = "70528ca9-195c-49a5-9d83-d51df3b0591f"
	khmKayasOnslaughtOracle      = "230811bf-52e7-451e-96ec-9d57c864ac5d"
	khmMammothGrowthOracle       = "bd433a27-1f3f-4d69-8b7e-97239fdf165b"
	khmWarhornBlastOracle        = "8106606c-9a1c-44b2-a9bf-3f189ce050ef"
	khmWingsOfTheCosmosOracle    = "afa9cf24-cd02-41f5-9141-f6ff6f27d528"
	khmSquashOracle              = "54bd7b7c-4052-437d-b2e4-311339e8e16e"
	khmTergridsShadowOracle      = "692260bb-e01c-4550-832a-9611856a6ff0"
	khmDwarvenReinforcementsOrcl = "b38c567e-fd79-4a74-ab5f-df4a6a8f7e00"
	khmPoisonTheCupOracle        = "58020778-554c-4af6-8484-0b90f64bc5d9"
	khmBeskirShieldmateOracle    = "a1090a50-2347-48d0-bb67-379accb8e02d"
	khmElderleafMentorOracle     = "3795fc31-2369-47d0-b9c9-120f44b8f114"
	khmRavenousLindwurmOracle    = "3583ec24-4f26-45ea-8fb4-a45bdd7bfc79"
	khmImmersturmRaiderOracle    = "e3179dd6-5558-4385-8883-f78f9d1b1c4f"
	khmBrinebarrowIntruderOracle = "eba73313-5ec5-4004-8040-658861b71a8f"
	khmGuardianGladewalkerOracle = "03a3f54e-9289-4d99-beab-bfa0988a050e"
	khmSarulfsPackmateOracle     = "f32caf11-abc1-4cae-a311-10210176ea3a"
	khmDoomskarTitanOracle       = "86692047-813f-47b1-9f4e-c8b9d5b17b1c"
	khmKarfellKennelMasterOracle = "2cde1ca7-d126-4d3a-82a6-0acfd7c85cdb"
	khmElderfangRitualistOracle  = "29c1f80e-f7e2-425e-a225-e88351fcbaab"
	khmClarionSpiritOracle       = "dec20b18-afe6-405e-b47d-a757a5645c63"
	khmDoomskarOracleOracle      = "c1d8e97b-f34b-4e52-9795-a4433f26725a"
	khmMistwalkerOracle          = "ceb302df-d67e-4e5d-aca0-f53bcdea658d"
	khmCravenHulkOracle          = "3cd13364-3ee8-4a00-986a-489b0ccf90c7"
	khmHagiMobOracle             = "c19aaf86-b098-4e7d-af50-9e02ae3cda55"
	khmGoldmawChampionOracle     = "9cbc8765-1735-48ca-bbc2-411aa7da2ed3"
	khmDoggedPursuitOracle       = "987f8c13-0164-4536-b505-d9fa799222fb"
	khmFuneralLongboatOracle     = "a994de7b-0128-4f03-922f-c8cd80cb5852"
	khmBearType                  = "Creature — Bear"
	khmGiantType                 = "Creature — Giant Berserker"
)

// khmBear seeds a vanilla 2/2 for `owner`.
func khmBear(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Creature(g, owner, "Bear", khmBearType, 2, 2)
}

// khmHandCreature puts a catalog creature with a body into p's hand.
func khmHandCreature(p *game.Player, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// khmTap taps a permanent through the effect API.
func khmTap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
}

// --- every card is in the catalog ----------------------------------------

func TestKhm5505EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		khmBattlefieldRaptorOracle: "Battlefield Raptor", khmBreakneckBerserkerOracle: "Breakneck Berserker",
		khmGnottvoldRecluseOracle: "Gnottvold Recluse", khmStorySeekerOracle: "Story Seeker",
		khmGrizzledOutriderOracle: "Grizzled Outrider", khmAuguryRavenOracle: "Augury Raven",
		khmGodsHallGuardianOracle: "Gods' Hall Guardian", khmVengefulReaperOracle: "Vengeful Reaper",
		khmScornEffigyOracle: "Scorn Effigy", khmSulfurousMireOracle: "Sulfurous Mire",
		khmVolatileFjordOracle: "Volatile Fjord", khmDemonBoltOracle: "Demon Bolt",
		khmDepartTheRealmOracle: "Depart the Realm", khmDoomskarOracle: "Doomskar",
		khmFeedTheSerpentOracle: "Feed the Serpent", khmIronVerdictOracle: "Iron Verdict",
		khmKayasOnslaughtOracle: "Kaya's Onslaught", khmMammothGrowthOracle: "Mammoth Growth",
		khmWarhornBlastOracle: "Warhorn Blast", khmWingsOfTheCosmosOracle: "Wings of the Cosmos",
		khmSquashOracle: "Squash", khmTergridsShadowOracle: "Tergrid's Shadow",
		khmDwarvenReinforcementsOrcl: "Dwarven Reinforcements", khmPoisonTheCupOracle: "Poison the Cup",
		khmBeskirShieldmateOracle: "Beskir Shieldmate", khmElderleafMentorOracle: "Elderleaf Mentor",
		khmRavenousLindwurmOracle: "Ravenous Lindwurm", khmImmersturmRaiderOracle: "Immersturm Raider",
		khmBrinebarrowIntruderOracle: "Brinebarrow Intruder", khmGuardianGladewalkerOracle: "Guardian Gladewalker",
		khmSarulfsPackmateOracle: "Sarulf's Packmate", khmDoomskarTitanOracle: "Doomskar Titan",
		khmKarfellKennelMasterOracle: "Karfell Kennel-Master", khmElderfangRitualistOracle: "Elderfang Ritualist",
		khmClarionSpiritOracle: "Clarion Spirit", khmDoomskarOracleOracle: "Doomskar Oracle",
		khmMistwalkerOracle: "Mistwalker", khmCravenHulkOracle: "Craven Hulk",
		khmHagiMobOracle: "Hagi Mob", khmGoldmawChampionOracle: "Goldmaw Champion",
		khmDoggedPursuitOracle: "Dogged Pursuit", khmFuneralLongboatOracle: "Funeral Longboat",
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

// --- printed keywords and foretell ----------------------------------------

func TestKhm5505KeywordBodies(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Battlefield Raptor", "Creature — Bird", khmBattlefieldRaptorOracle, []string{"flying", "first strike"}},
		{"Breakneck Berserker", "Creature — Dwarf Berserker", khmBreakneckBerserkerOracle, []string{"haste"}},
		{"Gnottvold Recluse", "Creature — Spider", khmGnottvoldRecluseOracle, []string{"reach"}},
		{"Story Seeker", "Creature — Dwarf Cleric", khmStorySeekerOracle, []string{"lifelink"}},
		{"Augury Raven", "Creature — Bird", khmAuguryRavenOracle, []string{"flying"}},
		{"Gods' Hall Guardian", "Creature — Cat", khmGodsHallGuardianOracle, []string{"vigilance"}},
		{"Vengeful Reaper", "Creature — Angel Cleric", khmVengefulReaperOracle, []string{"flying", "deathtouch", "haste"}},
		{"Brinebarrow Intruder", "Creature — Human Rogue", khmBrinebarrowIntruderOracle, []string{"flash"}},
		{"Guardian Gladewalker", "Creature — Shapeshifter", khmGuardianGladewalkerOracle, []string{game.KeywordChangeling}},
		{"Mistwalker", "Creature — Shapeshifter", khmMistwalkerOracle, []string{game.KeywordChangeling, "flying"}},
		{"Funeral Longboat", "Artifact — Vehicle", khmFuneralLongboatOracle, []string{"vigilance"}},
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

func TestKhm5505ForetellCosts(t *testing.T) {
	for _, tc := range []struct{ name, oracle, cost string }{
		{"Augury Raven", khmAuguryRavenOracle, "{1}{U}"},
		{"Gods' Hall Guardian", khmGodsHallGuardianOracle, "{3}{W}"},
		{"Vengeful Reaper", khmVengefulReaperOracle, "{1}{B}"},
		{"Scorn Effigy", khmScornEffigyOracle, "{0}"},
		{"Demon Bolt", khmDemonBoltOracle, "{R}"},
		{"Depart the Realm", khmDepartTheRealmOracle, "{U}"},
		{"Doomskar", khmDoomskarOracle, "{1}{W}{W}"},
		{"Iron Verdict", khmIronVerdictOracle, "{W}"},
		{"Kaya's Onslaught", khmKayasOnslaughtOracle, "{W}"},
		{"Mammoth Growth", khmMammothGrowthOracle, "{G}"},
		{"Warhorn Blast", khmWarhornBlastOracle, "{2}{W}"},
		{"Tergrid's Shadow", khmTergridsShadowOracle, "{2}{B}{B}"},
		{"Dwarven Reinforcements", khmDwarvenReinforcementsOrcl, "{1}{R}"},
		{"Poison the Cup", khmPoisonTheCupOracle, "{1}{B}"},
		{"Sarulf's Packmate", khmSarulfsPackmateOracle, "{1}{G}"},
		{"Doomskar Titan", khmDoomskarTitanOracle, "{4}{R}"},
		{"Doomskar Oracle", khmDoomskarOracleOracle, "{W}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sa := foretellOf(tc.oracle)
			if sa == nil {
				t.Fatalf("%s declares no foretell", tc.name)
			}
			if sa.CastCost != tc.cost {
				t.Errorf("foretell cost %q, want %q", sa.CastCost, tc.cost)
			}
			if sa.Cost != game.ForetellExileCost {
				t.Errorf("special-action cost %q, want the keyword's %q", sa.Cost, game.ForetellExileCost)
			}
		})
	}
}

// --- the snow duals ---------------------------------------------------------

func TestKhm5505SnowDualsEnterTappedAndTapForEitherColour(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle, a, b string }{
		{"Sulfurous Mire", "Snow Land — Swamp Mountain", khmSulfurousMireOracle, "B", "R"},
		{"Volatile Fjord", "Snow Land — Island Mountain", khmVolatileFjordOracle, "U", "R"},
	} {
		for _, colour := range []string{tc.a, tc.b} {
			t.Run(tc.name+" "+colour, func(t *testing.T) {
				g := newCatalogGame(t)
				me := g.Seats[0]
				land := b12PlayFromHand(t, g, tc.name, tc.typeLine, tc.oracle, game.CastSpellParams{})
				if !b16Tapped(t, g, land) {
					t.Fatal("the snow dual did not enter tapped")
				}
				g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
				b28TapForMana(t, g, me.ID, land, colour)
				if got := poolColors(me); len(got) != 1 || got[0] != colour {
					t.Errorf("tapped for {%s}: pool %v", colour, got)
				}
			})
		}
	}
}

// --- removal and burn ---------------------------------------------------------

func TestKhm5505DemonBoltDealsFourToACreatureOrPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := castCatalogSpellErr(t, g, "Demon Bolt", "Instant", khmDemonBoltOracle, cardTarget(rock)); err == nil {
		t.Fatal("Demon Bolt targeted an artifact")
	}
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	castCatalogSpell(t, g, "Demon Bolt", "Instant", khmDemonBoltOracle, cardTarget(wall))
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, wall).DamageMarked; got != 4 {
		t.Errorf("the creature has %d damage, want 4", got)
	}
}

func TestKhm5505SquashCostsThreeLessWithAGiant(t *testing.T) {
	for _, withGiant := range []bool{false, true} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me, opp := top100ActiveSeat(g), top100Opponent(g)
		if withGiant {
			b12Creature(g, me.ID, "Giant", khmGiantType, 3, 3)
		}
		wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 7)
		id := handCardFull(me, "Squash", "Instant", "{4}{R}", khmSquashOracle, []string{"R"})
		apaMana(me, "C", "R")
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardTarget(wall)})
		if !withGiant {
			if err == nil {
				t.Error("Squash was cast for {1}{R} without a Giant")
			}
			continue
		}
		if err != nil {
			t.Fatalf("Squash for {1}{R} with a Giant: %v", err)
		}
		passPriorityAroundTable(t, g)
		if got := findCardAnywhere(t, g, wall).DamageMarked; got != 6 {
			t.Errorf("the creature has %d damage, want 6", got)
		}
	}
}

func TestKhm5505IronVerdictHitsOnlyATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 7)
	if err := castCatalogSpellErr(t, g, "Iron Verdict", "Instant", khmIronVerdictOracle, cardTarget(wall)); err == nil {
		t.Fatal("Iron Verdict targeted an untapped creature")
	}
	khmTap(g, wall)
	castCatalogSpell(t, g, "Iron Verdict", "Instant", khmIronVerdictOracle, cardTarget(wall))
	passPriorityAroundTable(t, g)
	if got := findCardAnywhere(t, g, wall).DamageMarked; got != 5 {
		t.Errorf("the creature has %d damage, want 5", got)
	}
}

func TestKhm5505FeedTheSerpentExiles(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	bear := khmBear(g, opp.ID)
	castCatalogSpell(t, g, "Feed the Serpent", "Instant", khmFeedTheSerpentOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || opp.Graveyard.Contains(bear) || !g.Exile.Contains(bear) {
		t.Error("the creature was not exiled")
	}
}

func TestKhm5505DepartTheRealmBouncesANonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	land := b12Permanent(g, opp.ID, "Their Land", "Basic Land — Island")
	if err := castCatalogSpellErr(t, g, "Depart the Realm", "Instant", khmDepartTheRealmOracle, cardTarget(land)); err == nil {
		t.Fatal("Depart the Realm targeted a land")
	}
	bear := khmBear(g, opp.ID)
	castCatalogSpell(t, g, "Depart the Realm", "Instant", khmDepartTheRealmOracle, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
}

func TestKhm5505DoomskarDestroysAllCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine, theirs := khmBear(g, me.ID), khmBear(g, opp.ID)
	land := b12Permanent(g, opp.ID, "Their Land", "Basic Land — Plains")
	castCatalogSpell(t, g, "Doomskar", "Sorcery", khmDoomskarOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a creature survived Doomskar")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("Doomskar destroyed a land")
	}
}

func TestKhm5505TergridsShadowEachPlayerSacrificesTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	a, b, keep := khmBear(g, me.ID), khmBear(g, me.ID), khmBear(g, me.ID)
	x, y := khmBear(g, opp.ID), khmBear(g, opp.ID)
	castCatalogSpell(t, g, "Tergrid's Shadow", "Instant", khmTergridsShadowOracle, nil)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, a, b)
	answerSacrifice(t, g, opp.ID, sacrificeOptions(sacrificeChoiceFor(g, opp.ID))...)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, x, y} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", id)
		}
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("a third creature was sacrificed")
	}
}

func TestKhm5505PoisonTheCupDestroysAndScriesOnlyWhenForetold(t *testing.T) {
	t.Run("cast from hand", func(t *testing.T) {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		bear := khmBear(g, top100Opponent(g).ID)
		castCatalogSpell(t, g, "Poison the Cup", "Instant", khmPoisonTheCupOracle, cardTarget(bear))
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(bear) {
			t.Error("the creature survived")
		}
		if latestChoiceOfKind(g, game.PendingChoiceScry) != nil {
			t.Error("a spell that was not foretold scried")
		}
	})
	t.Run("foretold", func(t *testing.T) {
		g := newCatalogGame(t)
		me := top100ActiveSeat(g)
		advanceTo(t, g, game.StepPrecombatMain)
		card := handCardForTest(me, "Poison the Cup", "Instant", khmPoisonTheCupOracle)
		me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
		if err := g.PerformSpecialAction(me.ID, card, game.SpecialActionForetell, game.SpecialActionParams{Strict: true}); err != nil {
			t.Fatalf("foretell Poison the Cup: %v", err)
		}
		g.WithWriteLock(func() { g.Turn.Seq++ })
		bear := khmBear(g, top100Opponent(g).ID)
		me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "B"})
		if err := g.CastSpell(me.ID, card, game.CastSpellParams{
			Strict:          true,
			FromZone:        "exile",
			AlternativeCost: game.AltCostKeyForetell,
			Targets:         cardTarget(bear),
		}); err != nil {
			t.Fatalf("cast the foretold Poison the Cup: %v", err)
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(bear) {
			t.Error("the creature survived")
		}
		scry := latestChoiceOfKind(g, game.PendingChoiceScry)
		if scry == nil {
			t.Fatal("the foretold spell did not scry")
		}
		if scry.Count != 2 {
			t.Errorf("scry %d, want 2", scry.Count)
		}
	})
}

// --- pumps and tokens ----------------------------------------------------------

func TestKhm5505TargetedPumps(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		p, tough     int
		keyword      string
	}{
		{"Kaya's Onslaught", khmKayasOnslaughtOracle, 3, 3, "double strike"},
		{"Mammoth Growth", khmMammothGrowthOracle, 6, 6, ""},
		{"Wings of the Cosmos", khmWingsOfTheCosmosOracle, 3, 5, "flying"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			bear := khmBear(g, top100ActiveSeat(g).ID)
			khmTap(g, bear)
			castCatalogSpell(t, g, tc.name, "Instant", tc.oracle, cardTarget(bear))
			passPriorityAroundTable(t, g)
			if p, tough := ptOf(t, g, bear); p != tc.p || tough != tc.tough {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, tc.p, tc.tough)
			}
			if tc.keyword != "" && !hasEffectiveKeyword(t, g, bear, tc.keyword) {
				t.Errorf("the Bear has no %s", tc.keyword)
			}
			if tapped := hmlTapped(t, g, bear); tapped != (tc.name != "Wings of the Cosmos") {
				t.Errorf("tapped = %v after %s", tapped, tc.name)
			}
		})
	}
}

func TestKhm5505WarhornBlastPumpsOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	a, b, theirs := khmBear(g, me.ID), khmBear(g, me.ID), khmBear(g, opp.ID)
	castCatalogSpell(t, g, "Warhorn Blast", "Instant", khmWarhornBlastOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, tough := ptOf(t, g, id); p != 4 || tough != 3 {
			t.Errorf("your Bear is %d/%d, want 4/3", p, tough)
		}
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their Bear is %d/%d, want 2/2", p, tough)
	}
}

func TestKhm5505DwarvenReinforcementsMakesTwoDwarves(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	castCatalogSpell(t, g, "Dwarven Reinforcements", "Sorcery", khmDwarvenReinforcementsOrcl, nil)
	passPriorityAroundTable(t, g)
	n, tok := b14Tokens(g, me.ID, "Dwarf Berserker")
	if n != 2 {
		t.Fatalf("%d Dwarf Berserker tokens, want 2", n)
	}
	if tok.Power != 2 || tok.Toughness != 1 {
		t.Errorf("token is %d/%d, want 2/1", tok.Power, tok.Toughness)
	}
}

// --- enters and dies triggers ----------------------------------------------------

func TestKhm5505EntersTriggers(t *testing.T) {
	t.Run("Elderleaf Mentor", func(t *testing.T) {
		g := newCatalogGame(t)
		me := top100ActiveSeat(g)
		castHoldCreature(t, g, "Elderleaf Mentor", "Creature — Elf Warrior", khmElderleafMentorOracle, 3, 2)
		passPriorityAroundTable(t, g)
		if n, _ := b14Tokens(g, me.ID, "Elf Warrior"); n != 1 {
			t.Errorf("%d Elf Warrior tokens, want 1", n)
		}
	})
	t.Run("Ravenous Lindwurm", func(t *testing.T) {
		g := newCatalogGame(t)
		me := top100ActiveSeat(g)
		life := me.Life
		castHoldCreature(t, g, "Ravenous Lindwurm", "Creature — Wurm", khmRavenousLindwurmOracle, 6, 6)
		passPriorityAroundTable(t, g)
		if me.Life != life+4 {
			t.Errorf("life %d, want %d", me.Life, life+4)
		}
	})
	t.Run("Sarulf's Packmate", func(t *testing.T) {
		g := newCatalogGame(t)
		me := top100ActiveSeat(g)
		advanceToMain(t, g)
		hand := me.Hand.Size()
		castHoldCreature(t, g, "Sarulf's Packmate", "Creature — Wolf", khmSarulfsPackmateOracle, 3, 3)
		passPriorityAroundTable(t, g)
		if got := me.Hand.Size(); got != hand+1 {
			t.Errorf("hand is %d, want %d", got, hand+1)
		}
	})
}

func TestKhm5505ImmersturmRaiderRummages(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	castHoldCreature(t, g, "Immersturm Raider", "Creature — Demon Berserker", khmImmersturmRaiderOracle, 2, 1)
	passPriorityAroundTable(t, g)
	if n := discardOwed(g, me.ID); n != 1 {
		t.Fatalf("may discard %d, want up to 1", n)
	}
	hand := me.Hand.Size()
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand is %d after discarding one and drawing one, want %d", got, hand)
	}
}

func TestKhm5505BrinebarrowIntruderFlashesInAndShrinks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	bear := khmBear(g, opp.ID)
	id := khmHandCreature(me, "Brinebarrow Intruder", "Creature — Human Rogue", khmBrinebarrowIntruderOracle, 1, 2)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("Brinebarrow Intruder at the beginning of combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerPickTargetIfOpen(t, g, bear)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Brinebarrow Intruder did not resolve")
	}
	if p, tough := ptOf(t, g, bear); p != 0 || tough != 2 {
		t.Errorf("the opponent's Bear is %d/%d, want 0/2", p, tough)
	}
}

func TestKhm5505GuardianGladewalkerPutsACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	bear := khmBear(g, me.ID)
	castHoldCreature(t, g, "Guardian Gladewalker", "Creature — Shapeshifter", khmGuardianGladewalkerOracle, 1, 1)
	answerPickTargetIfOpen(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, bear); got != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", got)
	}
}

func TestKhm5505DoomskarTitanPumpsAndHastesYourTeam(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	mine, theirs := khmBear(g, me.ID), khmBear(g, opp.ID)
	titan := castHoldCreature(t, g, "Doomskar Titan", khmGiantType, khmDoomskarTitanOracle, 4, 4)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 3 || tough != 2 {
		t.Errorf("your Bear is %d/%d, want 3/2", p, tough)
	}
	for _, id := range []uuid.UUID{mine, titan} {
		if !hasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%s has no haste", id)
		}
	}
	if p, _ := ptOf(t, g, theirs); p != 2 || hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature was pumped")
	}
}

func TestKhm5505KarfellKennelMasterShieldsUpToTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	a, b := khmBear(g, me.ID), khmBear(g, me.ID)
	castHoldCreature(t, g, "Karfell Kennel-Master", "Creature — Zombie Berserker", khmKarfellKennelMasterOracle, 4, 4)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(a, b)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if pw, tough := ptOf(t, g, id); pw != 3 || tough != 2 {
			t.Errorf("a Bear is %d/%d, want 3/2", pw, tough)
		}
		if !hasEffectiveKeyword(t, g, id, "indestructible") {
			t.Error("a Bear has no indestructible")
		}
	}
}

func TestKhm5505BeskirShieldmateLeavesAWarrior(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Beskir Shieldmate", "Creature — Human Warrior", khmBeskirShieldmateOracle, 2, 1)
	b27Kill(g, id)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	n, tok := b14Tokens(g, me.ID, "Human Warrior")
	if n != 1 {
		t.Fatalf("%d Human Warrior tokens, want 1", n)
	}
	if tok.Power != 1 || tok.Toughness != 1 {
		t.Errorf("token is %d/%d, want 1/1", tok.Power, tok.Toughness)
	}
}

func TestKhm5505ElderfangRitualistReturnsAnotherElf(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := wthGraveyardCard(me, "Old Elf", "Creature — Elf Warrior")
	goblin := wthGraveyardCard(me, "Old Goblin", "Creature — Goblin")
	id := b12Push(g, me.ID, "Elderfang Ritualist", "Creature — Elf Cleric", khmElderfangRitualistOracle, 3, 1)
	b27Kill(g, id)
	g.RunStateChecksForTest()
	answerPickTargetIfOpen(t, g, elf)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(elf) {
		t.Error("the Elf card did not return to hand")
	}
	if me.Hand.Contains(goblin) || me.Hand.Contains(id) {
		t.Error("a card other than another Elf returned")
	}
}

// --- second spell each turn --------------------------------------------------------

func TestKhm5505SecondSpellTriggers(t *testing.T) {
	t.Run("Clarion Spirit", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		b12Push(g, me.ID, "Clarion Spirit", "Creature — Spirit", khmClarionSpiritOracle, 2, 2)
		for i, want := range []int{0, 1, 1} {
			b24CastNoncreature(t, g, "Spell", "Instant", "{U}")
			passPriorityAroundTable(t, g)
			n, tok := b14Tokens(g, me.ID, "Spirit")
			if n != want {
				t.Fatalf("after spell %d, %d Spirit tokens, want %d", i+1, n, want)
			}
			if n > 0 && !hasEffectiveKeyword(t, g, tok.InstanceID, "flying") {
				t.Error("the Spirit has no flying")
			}
		}
	})
	t.Run("Doomskar Oracle", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		b12Push(g, me.ID, "Doomskar Oracle", "Creature — Human Cleric", khmDoomskarOracleOracle, 3, 2)
		life := me.Life
		for i, want := range []int{0, 2, 2} {
			b24CastNoncreature(t, g, "Spell", "Instant", "{U}")
			passPriorityAroundTable(t, g)
			if got := me.Life - life; got != want {
				t.Fatalf("after spell %d, gained %d, want %d", i+1, got, want)
			}
		}
	})
}

// --- activated abilities, boast, crew ----------------------------------------------

func TestKhm5505MistwalkerPumps(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	id := b12Push(g, me.ID, "Mistwalker", "Creature — Shapeshifter", khmMistwalkerOracle, 1, 4)
	apaMana(me, "C", "U")
	apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if p, tough := ptOf(t, g, id); p != 2 || tough != 3 {
		t.Errorf("Mistwalker is %d/%d, want 2/3", p, tough)
	}
}

func TestKhm5505HagiMobBoastsForOneDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mob := pushCatalogPermanent(g, me.ID, "Hagi Mob", "Creature — Troll Berserker", khmHagiMobOracle, false)
	advanceToMain(t, g)
	err := g.ActivateCatalogAbility(me.ID, mob, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if !errors.Is(err, game.ErrBoastNotAttacked) {
		t.Fatalf("before attacking: %v, want ErrBoastNotAttacked", err)
	}
	attackWith(t, g, opp.ID, mob)
	life := opp.Life
	apaMana(me, "C", "R")
	apaActivate(t, g, me, mob, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
	if opp.Life != life-1 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-1)
	}
}

func TestKhm5505GoldmawChampionBoastsToTap(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champ := pushCatalogPermanent(g, me.ID, "Goldmaw Champion", "Creature — Dwarf Warrior", khmGoldmawChampionOracle, false)
	bear := khmBear(g, opp.ID)
	attackWith(t, g, opp.ID, champ)
	apaMana(me, "C", "W")
	apaActivate(t, g, me, champ, 0, game.ActivateAbilityParams{Targets: cardTarget(bear)})
	if !hmlTapped(t, g, bear) {
		t.Error("the target creature is untapped")
	}
}

func TestKhm5505FuneralLongboatCrewsWithOnePower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	boat := b12Push(g, me.ID, "Funeral Longboat", "Artifact — Vehicle", khmFuneralLongboatOracle, 3, 3)
	crew := b12Creature(g, me.ID, "Crew", "Creature — Test", 1, 1)
	advanceToMain(t, g)
	if isCreatureNow(g, boat) {
		t.Fatal("the Vehicle is a creature before it is crewed")
	}
	if err := g.ActivateCatalogAbility(me.ID, boat, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("crew 1: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !isCreatureNow(g, boat) {
		t.Error("the crewed Vehicle is not a creature")
	}
}

// --- statics and the end step ----------------------------------------------------

func TestKhm5505CravenHulkCantBlockAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a1 := pushVanillaCreature(g, me.ID, "Raider One", 2, 2)
	a2 := pushVanillaCreature(g, me.ID, "Raider Two", 2, 2)
	hulk := apaPush(g, opp.ID, opp.ID, apaCreature("Craven Hulk", khmCravenHulkOracle, 4, 4))
	bear := apaPush(g, opp.ID, opp.ID, apaCreature("Bear", "", 2, 2))
	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: a1, Target: opp.ID}, {Attacker: a2, Target: opp.ID}}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	err := g.DeclareBlocker(hulk, a1)
	var re *game.BlockRefusedError
	if !errors.As(err, &re) || re.Reason != game.BlockReasonCantBlockAlone {
		t.Fatalf("blocking alone: %v, want cant_block_alone", err)
	}
	if err := g.DeclareBlocker(bear, a1); err != nil {
		t.Fatalf("the Bear block: %v", err)
	}
	if err := g.DeclareBlocker(hulk, a2); err != nil {
		t.Fatalf("the Hulk beside the Bear: %v", err)
	}
}

func TestKhm5505DoggedPursuitDrainsEachOpponentAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Dogged Pursuit", "Enchantment", khmDoggedPursuitOracle, false)
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	walkToEndStepOfSeat0(t, g)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := lives[i] - 1
		if i == 0 {
			want = lives[i] + 1
		}
		if p.Life != want {
			t.Errorf("seat %d is at %d, want %d", i, p.Life, want)
		}
	}
}
