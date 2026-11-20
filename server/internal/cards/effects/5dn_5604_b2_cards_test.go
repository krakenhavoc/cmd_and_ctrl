package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// 5dn_5604_b2_cards_test.go — Fifth Dawn, batch 2 (#5604, S82). Every
// non-vanilla card is played through the engine.

const (
	dn5B2BlindCreeper         = "02d4c3a8-c733-45c1-bca5-51eed47c9230"
	dn5B2BlinkmothInfusion    = "8788f555-8c18-4842-9716-385afae2b9a4"
	dn5B2BringerGreenDawn     = "f0548af1-d1cc-44d1-b5a0-f013563f29de"
	dn5B2CacklingImp          = "adc8c55c-8fbe-4e07-b1be-8cb796db3d8a"
	dn5B2ChannelTheSuns       = "d1b1be96-b2aa-4386-97d2-eb0a8cc47210"
	dn5B2CompositeGolem       = "784970de-ef8c-4672-b5d0-24a4b0a978a3"
	dn5B2DevourInShadow       = "9086d8ff-e375-4149-996d-2a88a7596c1c"
	dn5B2DrossCrocodile       = "028fa430-722d-45aa-b415-3f84076091a9"
	dn5B2EarlyFrost           = "9f9cdbc2-d2f2-4390-8506-350fabf3f89a"
	dn5B2EbonDrake            = "db4242ab-fdf3-487a-bc77-42bd2ef973c3"
	dn5B2FangrenPathcutter    = "36d1a3f4-7903-41aa-9fb7-401d8789b00c"
	dn5B2FeedbackBolt         = "2fc5ebd9-e4af-4e2b-a2cc-a13fe051c5cb"
	dn5B2FerociousCharge      = "a7796af7-d41c-42a6-9033-a4d79dee84d6"
	dn5B2Fleshgrafter         = "04b1fb9e-4aef-46d9-b163-f9beff198bad"
	dn5B2FurnaceWhelp         = "8422f1e0-00ca-4ffb-a6b2-e2c9f96d7f23"
	dn5B2GemstoneArray        = "91303465-4b98-4ecc-82f4-cfd2f9a0f8a7"
	dn5B2GraftedWargear       = "c48cf069-eb20-4a61-b58c-b031ec22a5f0"
	dn5B2Granulate            = "fdcec649-6651-4ba0-bd3f-84f16e65397d"
	dn5B2HornedHelm           = "127b438d-e5c2-47ba-9554-d9d5af018d09"
	dn5B2IntoThinAir          = "06df923f-916e-49f4-960a-7b3fb108c704"
	dn5B2IronBarbHellion      = "6ec9f9a9-9e66-4d81-ae22-dc0de596acd5"
	dn5B2JoinerAdept          = "139b098c-d4be-4dca-9bde-9eb8ea97c7bf"
	dn5B2KrarkClanEngineers   = "0964288c-0397-428a-9a84-784aaea4e040"
	dn5B2KrarkClanOgre        = "5de63529-7dea-4018-bf46-afb4bc8b69fd"
	dn5B2LeoninSquire         = "fa19c9d7-fe93-4e93-a5ef-9af9c6cd1a5e"
	dn5B2LoseHope             = "4f6dbd90-f7fe-4adf-a00f-27573c0bd5c4"
	dn5B2LoxodonAnchorite     = "c70dded8-658b-4250-8057-54894350455c"
	dn5B2LoxodonStalwart      = "6b2eb774-08f3-46cb-aeca-2ab82fdd5f03"
	dn5B2MagmaJet             = "2f292253-64d3-4cf2-881f-a3eea4fda388"
	dn5B2NeurokStealthsuit    = "cee06ea6-8bad-4858-b78e-04879a825bbe"
	dn5B2NimGrotesque         = "ed4d4084-739b-43ac-bd2b-a86d1ecd7bbd"
	dn5B2PlasmaElemental      = "419c501f-c107-4166-b56d-631672f2a5cc"
	dn5B2Qumulox              = "cec11318-4578-4c66-a9be-1704df0c296c"
	dn5B2ScreamingFury        = "eaded717-e0e2-4def-be7e-6bfb137628c2"
	dn5B2SkyhunterProwler     = "5e0d13b0-d5d0-4c5f-945e-5cda7d980670"
	dn5B2SkyhunterSkirmisher  = "0c07d09e-e127-4573-b827-6c50246f7a31"
	dn5B2SparkElemental       = "50608b7e-b6b2-40ca-91ec-db3e03dbb1af"
	dn5B2SparringCollar       = "81bfbc5a-c448-4124-ac25-79bd0944136d"
	dn5B2StandFirm            = "c0406e70-8131-4e13-b1d5-4e943ad296b8"
	dn5B2SylvokExplorer       = "cf7371c8-2fd8-4e70-abde-27495af85011"
	dn5B2TelJiladJustice      = "6c853b95-3d98-44fd-b05a-02d05d5eb0b8"
	dn5B2TelJiladLifebreather = "1ef49064-eb2a-49ae-917e-3d441ba14bf9"
	dn5B2ThermalNavigator     = "355c2840-2cb8-431a-bcdc-f4b8824a1f5d"
	dn5B2ThoughtCourier       = "06ea02f5-f296-4334-bb6f-9cb018e1e150"
	dn5B2Tyrranax             = "28ecc209-d78b-4b37-8411-fb8079f6e088"
	dn5B2Vanquish             = "683ec6d2-4ea7-46a7-911e-12c83e4c4a06"
	dn5B2VedalkenMastermind   = "1ba85537-6dc2-459e-8850-619b41c0a881"
	dn5B2ViridianLorebearers  = "080c0557-59ee-4c46-afd9-5baabf0dc3d3"
	dn5B2ViridianScout        = "c5b4bc25-0f32-4e1d-b6fe-f122140f2207"
	dn5B2VulshokSorcerer      = "6f858405-dfe4-4ca2-b2e3-80ce8cda5522"
)

func TestDn5B2CardsAreRegisteredFull(t *testing.T) {
	for name, oracle := range map[string]string{
		"Blind Creeper": dn5B2BlindCreeper, "Blinkmoth Infusion": dn5B2BlinkmothInfusion,
		"Bringer of the Green Dawn": dn5B2BringerGreenDawn, "Cackling Imp": dn5B2CacklingImp,
		"Channel the Suns": dn5B2ChannelTheSuns, "Composite Golem": dn5B2CompositeGolem,
		"Devour in Shadow": dn5B2DevourInShadow, "Dross Crocodile": dn5B2DrossCrocodile,
		"Early Frost": dn5B2EarlyFrost, "Ebon Drake": dn5B2EbonDrake,
		"Fangren Pathcutter": dn5B2FangrenPathcutter, "Feedback Bolt": dn5B2FeedbackBolt,
		"Ferocious Charge": dn5B2FerociousCharge, "Fleshgrafter": dn5B2Fleshgrafter,
		"Furnace Whelp": dn5B2FurnaceWhelp, "Gemstone Array": dn5B2GemstoneArray,
		"Grafted Wargear": dn5B2GraftedWargear, "Granulate": dn5B2Granulate,
		"Horned Helm": dn5B2HornedHelm, "Into Thin Air": dn5B2IntoThinAir,
		"Iron-Barb Hellion": dn5B2IronBarbHellion, "Joiner Adept": dn5B2JoinerAdept,
		"Krark-Clan Engineers": dn5B2KrarkClanEngineers, "Krark-Clan Ogre": dn5B2KrarkClanOgre,
		"Leonin Squire": dn5B2LeoninSquire, "Lose Hope": dn5B2LoseHope,
		"Loxodon Anchorite": dn5B2LoxodonAnchorite, "Loxodon Stalwart": dn5B2LoxodonStalwart,
		"Magma Jet": dn5B2MagmaJet, "Neurok Stealthsuit": dn5B2NeurokStealthsuit,
		"Nim Grotesque": dn5B2NimGrotesque, "Plasma Elemental": dn5B2PlasmaElemental,
		"Qumulox": dn5B2Qumulox, "Screaming Fury": dn5B2ScreamingFury,
		"Skyhunter Prowler": dn5B2SkyhunterProwler, "Skyhunter Skirmisher": dn5B2SkyhunterSkirmisher,
		"Spark Elemental": dn5B2SparkElemental, "Sparring Collar": dn5B2SparringCollar,
		"Stand Firm": dn5B2StandFirm, "Sylvok Explorer": dn5B2SylvokExplorer,
		"Tel-Jilad Justice": dn5B2TelJiladJustice, "Tel-Jilad Lifebreather": dn5B2TelJiladLifebreather,
		"Thermal Navigator": dn5B2ThermalNavigator, "Thought Courier": dn5B2ThoughtCourier,
		"Tyrranax": dn5B2Tyrranax, "Vanquish": dn5B2Vanquish,
		"Vedalken Mastermind": dn5B2VedalkenMastermind, "Viridian Lorebearers": dn5B2ViridianLorebearers,
		"Viridian Scout": dn5B2ViridianScout, "Vulshok Sorcerer": dn5B2VulshokSorcerer,
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

// dn5B2Artifact puts a plain artifact with a mana cost onto the battlefield.
func dn5B2Artifact(g *game.Game, owner uuid.UUID, name, manaCost string) uuid.UUID {
	return mrd2Catalog(g, owner, name, "Artifact", "", manaCost, 0, 0)
}

// dn5B2Library gives a player three cards to scry or draw.
func dn5B2Library(p *game.Player) {
	mm2Library(p, mm2Spell("Library A"), mm2Spell("Library B"), mm2Spell("Library C"))
}

// --- printed keywords and restrictions ----------------------------------------

func TestDn5B2KeywordCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Iron-Barb Hellion", "Creature — Hellion Beast", dn5B2IronBarbHellion, []string{"haste"}},
		{"Qumulox", "Creature — Beast", dn5B2Qumulox, []string{"flying"}},
		{"Skyhunter Prowler", "Creature — Cat Knight", dn5B2SkyhunterProwler, []string{"flying", "vigilance"}},
		{"Skyhunter Skirmisher", "Creature — Cat Knight", dn5B2SkyhunterSkirmisher, []string{"flying", "double strike"}},
		{"Spark Elemental", "Creature — Elemental", dn5B2SparkElemental, []string{"trample", "haste"}},
		{"Furnace Whelp", "Creature — Dragon", dn5B2FurnaceWhelp, []string{"flying"}},
		{"Loxodon Stalwart", "Creature — Elephant Soldier", dn5B2LoxodonStalwart, []string{"vigilance"}},
		{"Vulshok Sorcerer", "Creature — Human Shaman Sorcerer", dn5B2VulshokSorcerer, []string{"haste"}},
		{"Cackling Imp", "Creature — Imp", dn5B2CacklingImp, []string{"flying"}},
		{"Ebon Drake", "Creature — Drake", dn5B2EbonDrake, []string{"flying"}},
		{"Bringer of the Green Dawn", "Creature — Bringer", dn5B2BringerGreenDawn, []string{"trample"}},
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

func TestDn5B2IronBarbHellionCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	id := b12Push(g, g.Seats[0].ID, "Iron-Barb Hellion", "Creature — Hellion Beast", dn5B2IronBarbHellion, 5, 4)
	assertRestrictions(t, g, id, game.CantBlock)
}

func TestDn5B2PlasmaElementalCantBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	id := b12Push(g, g.Seats[0].ID, "Plasma Elemental", "Creature — Elemental", dn5B2PlasmaElemental, 4, 1)
	assertRestrictions(t, g, id, game.CantBeBlocked)
}

func TestDn5B2AffinitySpellsDeclareTheirCostModifier(t *testing.T) {
	for _, oracle := range []string{dn5B2Qumulox, dn5B2IntoThinAir, dn5B2BlinkmothInfusion} {
		spec, _ := Lookup(oracle)
		if len(spec.SelfCostModifiers) != 1 {
			t.Errorf("%s: %d self cost modifiers, want affinity for artifacts", spec.Name, len(spec.SelfCostModifiers))
		}
	}
}

func TestDn5B2SparkElementalIsSacrificedAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := b12Push(g, me.ID, "Spark Elemental", "Creature — Elemental", dn5B2SparkElemental, 3, 1)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
		t.Fatal("the Elemental was not sacrificed at the end step")
	}
}

// --- self pumps -------------------------------------------------------------------

func TestDn5B2SelfPumps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		mana                   []string
		wantP, wantT           int
	}{
		{"Furnace Whelp", "Creature — Dragon", dn5B2FurnaceWhelp, 2, 2, []string{"R"}, 3, 2},
		{"Loxodon Stalwart", "Creature — Elephant Soldier", dn5B2LoxodonStalwart, 3, 3, []string{"W"}, 3, 4},
		{"Tyrranax", "Creature — Dinosaur Beast", dn5B2Tyrranax, 5, 4, []string{"C", "G"}, 4, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			ravB2MustActivate(t, g, me.ID, id, 0, game.ActivateAbilityParams{}, tc.mana...)
			if p, tough := effectivePT(t, g, id); p != tc.wantP || tough != tc.wantT {
				t.Errorf("%s is %d/%d, want %d/%d", tc.name, p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestDn5B2FleshgrafterDiscardsAnArtifactCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	graft := b12Push(g, me.ID, "Fleshgrafter", "Creature — Human Warrior", dn5B2Fleshgrafter, 2, 2)
	rock := pushCatalogHandCard(me, "Spare Rock", "Artifact", "")
	sorcery := pushCatalogHandCard(me, "Spare Sorcery", "Sorcery", "")
	if err := g.ActivateCatalogAbility(me.ID, graft, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{sorcery}}); err == nil {
		t.Fatal("a sorcery card was accepted as the discarded artifact card")
	}
	if err := g.ActivateCatalogAbility(me.ID, graft, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePT(t, g, graft); p != 4 || tough != 4 {
		t.Errorf("Fleshgrafter is %d/%d, want 4/4", p, tough)
	}
	if !me.Graveyard.Contains(rock) {
		t.Error("the artifact card is discarded as the cost")
	}
}

func TestDn5B2ThermalNavigatorSacrificesAnArtifactForFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	nav := b12Push(g, me.ID, "Thermal Navigator", "Artifact Creature — Construct", dn5B2ThermalNavigator, 2, 2)
	rock := dn5B2Artifact(g, me.ID, "Spare Rock", "{1}")
	ravB2MustActivate(t, g, me.ID, nav, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is sacrificed as the cost")
	}
	if !hasKeywordOnBattlefield(t, g, nav, "flying") {
		t.Error("the Navigator should have flying until end of turn")
	}
}

// --- tap and sacrifice abilities ------------------------------------------------

func TestDn5B2LoxodonAnchoritePreventsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	anchorite := b12Push(g, me.ID, "Loxodon Anchorite", "Creature — Elephant Cleric", dn5B2LoxodonAnchorite, 2, 3)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 5)
	ravB2MustActivate(t, g, me.ID, anchorite, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	if got := damageMarkedOn(g, bear); got != 1 {
		t.Errorf("3 damage into a 2-shield: marked %d, want 1", got)
	}
}

func TestDn5B2ThoughtCourierLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dn5B2Library(me)
	courier := b12Push(g, me.ID, "Thought Courier", "Creature — Human Wizard", dn5B2ThoughtCourier, 1, 1)
	hand := me.Hand.Size()
	ravB2MustActivate(t, g, me.ID, courier, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after the draw", me.Hand.Size(), hand+1)
	}
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt after the draw")
	}
}

func TestDn5B2PingersAtAPlayer(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Vulshok Sorcerer", "Creature — Human Shaman Sorcerer", dn5B2VulshokSorcerer},
		{"Cackling Imp", "Creature — Imp", dn5B2CacklingImp},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			life := opp.Life
			ravB2MustActivate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{playerTarget(opp.ID)}})
			if opp.Life != life-1 {
				t.Errorf("opponent's life %d, want %d", opp.Life, life-1)
			}
			if !tappedForTest(t, g, id) {
				t.Error("the ability taps its source")
			}
		})
	}
}

func TestDn5B2SylvokExplorerTapsForAnOpponentsLandColour(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	explorer := b12Push(g, me.ID, "Sylvok Explorer", "Creature — Human Druid", dn5B2SylvokExplorer, 1, 1)
	seedManaLand(g, them.ID, "Forest", "Basic Land — Forest", "G")
	seedManaLand(g, me.ID, "Island", "Basic Land — Island", "U")
	if err := g.ActivateManaAbility(me.ID, explorer, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "G" {
		t.Errorf("pool %v, want the opponent's Forest's {G}", poolColors(me))
	}
}

func TestDn5B2CompositeGolemSacrificesForFiveColours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	golem := b12Push(g, me.ID, "Composite Golem", "Artifact Creature — Golem", dn5B2CompositeGolem, 4, 4)
	if err := g.ActivateManaAbility(me.ID, golem, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	got := map[string]int{}
	for _, c := range poolColors(me) {
		got[c]++
	}
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		if got[c] != 1 {
			t.Errorf("pool %v, want one of each colour", poolColors(me))
			break
		}
	}
	if g.Battlefield.Contains(golem) || !me.Graveyard.Contains(golem) {
		t.Error("the Golem is sacrificed by its own ability")
	}
}

func TestDn5B2VedalkenMastermindReturnsAPermanentYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mastermind := b12Push(g, me.ID, "Vedalken Mastermind", "Creature — Vedalken Wizard", dn5B2VedalkenMastermind, 1, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := ravB2Activate(t, g, me.ID, mastermind, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}, "U"); err == nil {
		t.Fatal("an opponent's creature was accepted as a permanent you control")
	}
	me.ManaPool.EmptyPool()
	ravB2MustActivate(t, g, me.ID, mastermind, 0, game.ActivateAbilityParams{Targets: cardRefs(mine)}, "U")
	if !handHasID(me, mine) {
		t.Error("the creature should return to its owner's hand")
	}
}

func TestDn5B2ViridianScoutShootsOnlyAFlier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scout := b12Push(g, me.ID, "Viridian Scout", "Creature — Elf Warrior Scout", dn5B2ViridianScout, 1, 2)
	walker := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	flier := b12Push(g, opp.ID, "Furnace Whelp", "Creature — Dragon", dn5B2FurnaceWhelp, 2, 2)
	if err := ravB2Activate(t, g, me.ID, scout, 0, game.ActivateAbilityParams{Targets: cardRefs(walker)}, "C", "C", "G"); err == nil {
		t.Fatal("a creature without flying was accepted as the target")
	}
	me.ManaPool.EmptyPool()
	ravB2MustActivate(t, g, me.ID, scout, 0, game.ActivateAbilityParams{Targets: cardRefs(flier)}, "C", "C", "G")
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(flier) {
		t.Error("2 damage should destroy the 2/2 flier")
	}
	if g.Battlefield.Contains(scout) {
		t.Error("the Scout is sacrificed as the cost")
	}
}

func TestDn5B2KrarkClanEngineersSacrificeTwoArtifactsToDestroyOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	engineers := b12Push(g, me.ID, "Krark-Clan Engineers", "Creature — Goblin Artificer", dn5B2KrarkClanEngineers, 2, 2)
	a := dn5B2Artifact(g, me.ID, "Rock A", "{1}")
	b := dn5B2Artifact(g, me.ID, "Rock B", "{1}")
	target := dn5B2Artifact(g, opp.ID, "Their Rock", "{3}")
	ravB2MustActivate(t, g, me.ID, engineers, 0, game.ActivateAbilityParams{
		Targets: cardRefs(target), SacrificeIDs: []uuid.UUID{a, b},
	}, "R")
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("both artifacts are sacrificed as the cost")
	}
	if g.Battlefield.Contains(target) {
		t.Error("the target artifact should be destroyed")
	}
}

func TestDn5B2KrarkClanOgreStopsABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := b12Push(g, me.ID, "Krark-Clan Ogre", "Creature — Ogre", dn5B2KrarkClanOgre, 3, 3)
	rock := dn5B2Artifact(g, me.ID, "Spare Rock", "{1}")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ravB2MustActivate(t, g, me.ID, ogre, 0, game.ActivateAbilityParams{
		Targets: cardRefs(bear), SacrificeIDs: []uuid.UUID{rock},
	}, "R")
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is sacrificed as the cost")
	}
	assertRestrictions(t, g, bear, game.CantBlock)
}

func TestDn5B2TelJiladLifebreatherRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	troll := b12Push(g, me.ID, "Tel-Jilad Lifebreather", "Creature — Troll Shaman", dn5B2TelJiladLifebreather, 3, 2)
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ravB2MustActivate(t, g, me.ID, troll, 0, game.ActivateAbilityParams{
		Targets: cardRefs(bear), SacrificeIDs: []uuid.UUID{forest},
	}, "G")
	if g.Battlefield.Contains(forest) {
		t.Error("the Forest is sacrificed as the cost")
	}
	p1Destroy(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the regenerated creature should survive being destroyed")
	}
}

func TestDn5B2ViridianLorebearersCountsOpponentsArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := b12Push(g, me.ID, "Viridian Lorebearers", "Creature — Elf Shaman", dn5B2ViridianLorebearers, 3, 3)
	dn5B2Artifact(g, opp.ID, "Their Rock", "{1}")
	dn5B2Artifact(g, g.Seats[2].ID, "Another Rock", "{1}")
	dn5B2Artifact(g, me.ID, "My Rock", "{1}")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ravB2MustActivate(t, g, me.ID, elf, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}, "C", "C", "C", "G")
	if p, tough := effectivePT(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the bear is %d/%d, want 4/4 for the opponents' two artifacts", p, tough)
	}
}

func TestDn5B2GemstoneArrayBanksAndSpendsACharge(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	array := b12Push(g, me.ID, "Gemstone Array", "Artifact", dn5B2GemstoneArray, 0, 0)
	ravB2MustActivate(t, g, me.ID, array, 0, game.ActivateAbilityParams{}, "C", "C")
	if got := counterCount(g, array, game.CounterCharge); got != 1 {
		t.Fatalf("charge counters %d, want 1", got)
	}
	if err := g.ActivateManaAbility(me.ID, array, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if latestChoiceOfKind(g, game.PendingChoiceMana) != nil {
		resolveManaPickAnyColor(t, g, me.ID)
	}
	if got := counterCount(g, array, game.CounterCharge); got != 0 {
		t.Errorf("charge counters %d, want 0 after the mana ability", got)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool %v, want one mana", poolColors(me))
	}
}

// --- Equipment ------------------------------------------------------------------------

func TestDn5B2ColouredAttachEquipment(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, mana string
		keyword            string
		wantP, wantT       int
	}{
		{"Horned Helm", dn5B2HornedHelm, "G", "trample", 3, 3},
		{"Sparring Collar", dn5B2SparringCollar, "R", "first strike", 2, 2},
		{"Neurok Stealthsuit", dn5B2NeurokStealthsuit, "U", "shroud", 2, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			eq := seedEquipment(g, me.ID, tc.name, tc.oracle)
			bear := seedBear(g, me.ID)
			ravB2MustActivate(t, g, me.ID, eq, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}, tc.mana, tc.mana)
			if host := attachmentHostOf(t, g, eq); host.ID != bear {
				t.Fatalf("the Equipment is attached to %+v, want the bear", host)
			}
			if !hasKeywordOnBattlefield(t, g, bear, tc.keyword) {
				t.Errorf("the equipped bear has no %s", tc.keyword)
			}
			if p, tough := effectivePT(t, g, bear); p != tc.wantP || tough != tc.wantT {
				t.Errorf("the equipped bear is %d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
		})
	}
}

func TestDn5B2GraftedWargearSacrificesTheCreatureItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	wargear := seedEquipment(g, me.ID, "Grafted Wargear", dn5B2GraftedWargear)
	first := b12Creature(g, me.ID, "First Bear", "Creature — Bear", 2, 2)
	second := b12Creature(g, me.ID, "Second Bear", "Creature — Bear", 2, 2)
	ravB2MustActivate(t, g, me.ID, wargear, 0, game.ActivateAbilityParams{Targets: cardRefs(first)})
	if p, tough := effectivePT(t, g, first); p != 5 || tough != 4 {
		t.Fatalf("the equipped bear is %d/%d, want 5/4", p, tough)
	}
	ravB2MustActivate(t, g, me.ID, wargear, 0, game.ActivateAbilityParams{Targets: cardRefs(second)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(first) {
		t.Error("the creature the Wargear became unattached from should be sacrificed")
	}
	if !g.Battlefield.Contains(second) {
		t.Error("the newly equipped creature stays")
	}
}

// --- statics --------------------------------------------------------------------------

func TestDn5B2NimGrotesqueCountsYourArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nim := b12Push(g, me.ID, "Nim Grotesque", "Creature — Zombie", dn5B2NimGrotesque, 3, 6)
	dn5B2Artifact(g, me.ID, "Rock A", "{1}")
	dn5B2Artifact(g, me.ID, "Rock B", "{1}")
	dn5B2Artifact(g, opp.ID, "Their Rock", "{1}")
	if p, tough := effectivePT(t, g, nim); p != 5 || tough != 6 {
		t.Errorf("Nim Grotesque is %d/%d, want 5/6 for my two artifacts", p, tough)
	}
}

func TestDn5B2JoinerAdeptGivesYourLandsAnyColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	gaPut(g, me.ID, "Joiner Adept", "Creature — Elf Druid", dn5B2JoinerAdept)
	if n := grantedManaRowsOf(g, forest); n != 1 {
		t.Fatalf("the Forest has %d granted rows, want 1", n)
	}
	if err := tapGrantedMana(t, g, me.ID, forest, "R"); err != nil {
		t.Fatalf("tap the Forest: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
}

// --- triggers -------------------------------------------------------------------------

func TestDn5B2EbonDrakeLosesLifeForEachSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Ebon Drake", "Creature — Drake", dn5B2EbonDrake, 3, 3)
	life := me.Life
	castCatalogSpell(t, g, "Channel the Suns", "Sorcery", dn5B2ChannelTheSuns, nil)
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

func TestDn5B2BlindCreeperShrinksForEachSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	creeper := b12Push(g, me.ID, "Blind Creeper", "Creature — Zombie Beast", dn5B2BlindCreeper, 3, 3)
	castCatalogSpell(t, g, "Channel the Suns", "Sorcery", dn5B2ChannelTheSuns, nil)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePT(t, g, creeper); p != 2 || tough != 2 {
		t.Errorf("Blind Creeper is %d/%d, want 2/2", p, tough)
	}
}

func TestDn5B2LeoninSquireReturnsACheapArtifactCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cheap := dn5GraveyardCard(me, "Cheap Trinket", "Artifact", "{1}")
	dn5GraveyardCard(me, "Dear Trinket", "Artifact", "{2}")
	castCatalogSpell(t, g, "Leonin Squire", "Creature — Cat Soldier", dn5B2LeoninSquire, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if !handHasID(me, cheap) {
		t.Error("the mana value 1 artifact card should return to hand")
	}
}

func TestDn5B2BringerOfTheGreenDawnMayMakeABeast(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	b12Push(g, owner.ID, "Bringer of the Green Dawn", "Creature — Bringer", dn5B2BringerGreenDawn, 5, 5)
	advanceToUpkeepOf(t, g, 1)
	answerLatestTriggerPrompt(t, g, owner.ID, true)
	passPriorityAroundTable(t, g)
	if got := tokensNamedControlledBy(g, "Beast", owner.ID); got != 1 {
		t.Errorf("%d Beasts, want 1", got)
	}
	if spec, _ := Lookup(dn5B2BringerGreenDawn); len(spec.AlternativeCosts) != 1 {
		t.Error("the Bringer offers its {W}{U}{B}{R}{G} alternative cost")
	}
}

func TestDn5B2FangrenPathcutterGivesAttackersTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fangren := b12Push(g, me.ID, "Fangren Pathcutter", "Creature — Beast", dn5B2FangrenPathcutter, 4, 6)
	bear := b12Creature(g, me.ID, "Attacking Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, fangren, bear)
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "trample") || !hasKeywordOnBattlefield(t, g, fangren, "trample") {
		t.Error("each attacking creature should gain trample")
	}
	if hasKeywordOnBattlefield(t, g, idle, "trample") {
		t.Error("a creature that isn't attacking gains nothing")
	}
}

// --- spells --------------------------------------------------------------------------

func TestDn5B2BoostThenScrySpells(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		wantP, wantT int
	}{
		{"Ferocious Charge", dn5B2FerociousCharge, 6, 7},
		{"Stand Firm", dn5B2StandFirm, 3, 4},
		{"Lose Hope", dn5B2LoseHope, 1, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			dn5B2Library(me)
			bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 3)
			castCatalogSpell(t, g, tc.name, "Instant", tc.oracle, cardTarget(bear))
			passPriorityAroundTable(t, g)
			if p, tough := effectivePT(t, g, bear); p != tc.wantP || tough != tc.wantT {
				t.Errorf("the bear is %d/%d, want %d/%d", p, tough, tc.wantP, tc.wantT)
			}
			if scryChoiceFor(g, me.ID) == nil {
				t.Fatal("no scry 2 prompt")
			}
			answerScryKeepAll(t, g, me.ID)
		})
	}
}

func TestDn5B2ScreamingFuryPumpsAndGrantsHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Screaming Fury", "Sorcery", dn5B2ScreamingFury, cardTarget(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePT(t, g, bear); p != 7 || tough != 2 {
		t.Errorf("the bear is %d/%d, want 7/2", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "haste") {
		t.Error("the bear should gain haste")
	}
}

func TestDn5B2MagmaJetDealsTwoAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	dn5B2Library(me)
	life := opp.Life
	castCatalogSpell(t, g, "Magma Jet", "Instant", dn5B2MagmaJet, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent's life %d, want %d", opp.Life, life-2)
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry 2 prompt")
	}
	answerScryKeepAll(t, g, me.ID)
}

func TestDn5B2TelJiladJusticeDestroysAnArtifactAndScries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	dn5B2Library(me)
	rock := dn5B2Artifact(g, opp.ID, "Their Rock", "{2}")
	castCatalogSpell(t, g, "Tel-Jilad Justice", "Instant", dn5B2TelJiladJustice, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact should be destroyed")
	}
	if scryChoiceFor(g, me.ID) == nil {
		t.Fatal("no scry 2 prompt")
	}
	answerScryKeepAll(t, g, me.ID)
}

func TestDn5B2ChannelTheSunsAddsOneOfEachColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Channel the Suns", "Sorcery", dn5B2ChannelTheSuns, nil)
	passPriorityAroundTable(t, g)
	got := map[string]int{}
	for _, c := range poolColors(me) {
		got[c]++
	}
	if len(got) != 5 || me.ManaPool.Total() != 5 {
		t.Errorf("pool %v, want {W}{U}{B}{R}{G}", poolColors(me))
	}
}

func TestDn5B2DevourInShadowCostsLifeEqualToToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	victim := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 1, 3)
	life := me.Life
	castCatalogSpell(t, g, "Devour in Shadow", "Instant", dn5B2DevourInShadow, cardTarget(victim))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature should be destroyed")
	}
	if me.Life != life-3 {
		t.Errorf("life %d, want %d", me.Life, life-3)
	}
}

func TestDn5B2IntoThinAirBouncesAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	rock := dn5B2Artifact(g, opp.ID, "Their Rock", "{2}")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Into Thin Air", "Instant", dn5B2IntoThinAir, cardTarget(bear)); err == nil {
		t.Fatal("a creature was accepted as the target artifact")
	}
	castCatalogSpell(t, g, "Into Thin Air", "Instant", dn5B2IntoThinAir, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if !handHasID(opp, rock) {
		t.Error("the artifact should return to its owner's hand")
	}
}

func TestDn5B2BlinkmothInfusionUntapsEveryArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	mine := dn5B2Artifact(g, me.ID, "My Rock", "{1}")
	theirs := dn5B2Artifact(g, opp.ID, "Their Rock", "{1}")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{mine, theirs, bear} {
		arnTap(t, g, id)
	}
	castCatalogSpell(t, g, "Blinkmoth Infusion", "Instant", dn5B2BlinkmothInfusion, nil)
	passPriorityAroundTable(t, g)
	if arnTapped(t, g, mine) || arnTapped(t, g, theirs) {
		t.Error("every artifact should untap")
	}
	if !arnTapped(t, g, bear) {
		t.Error("a creature that isn't an artifact stays tapped")
	}
}

func TestDn5B2EarlyFrostTapsUpToThreeLands(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := seatsForCombat(g)
	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		lands = append(lands, b12Permanent(g, opp.ID, "Island", "Basic Land — Island"))
	}
	castCatalogSpell(t, g, "Early Frost", "Instant", dn5B2EarlyFrost, cardRefs(lands...))
	passPriorityAroundTable(t, g)
	for _, id := range lands {
		if !arnTapped(t, g, id) {
			t.Error("each target land should be tapped")
		}
	}
}

func TestDn5B2VanquishDestroysOnlyABlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	if err := castCatalogSpellErrStrict(t, g, "Vanquish", "Instant", dn5B2Vanquish, cardRefs(blocker)); err == nil {
		t.Fatal("Vanquish targeted a creature that is not blocking")
	}
	attackIntoBlocks(t, g, opp.ID, attacker)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	g2CastFrom(t, g, me, "Vanquish", "Instant", dn5B2Vanquish, game.CastSpellParams{Targets: cardRefs(blocker)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocker) {
		t.Error("the blocking creature should be destroyed")
	}
}

func TestDn5B2FeedbackBoltCountsYourArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	for _, n := range []string{"Rock A", "Rock B", "Rock C"} {
		dn5B2Artifact(g, me.ID, n, "{1}")
	}
	dn5B2Artifact(g, opp.ID, "Their Rock", "{1}")
	life := opp.Life
	castCatalogSpell(t, g, "Feedback Bolt", "Instant", dn5B2FeedbackBolt, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent's life %d, want %d for my three artifacts", opp.Life, life-3)
	}
}

func TestDn5B2GranulateDestroysCheapNonlandArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seatsForCombat(g)
	four := dn5B2Artifact(g, opp.ID, "Four Rock", "{4}")
	one := dn5B2Artifact(g, me.ID, "One Rock", "{1}")
	five := dn5B2Artifact(g, opp.ID, "Five Rock", "{5}")
	land := mrd2Catalog(g, opp.ID, "Artifact Land", "Artifact Land", "", "", 0, 0)
	castCatalogSpell(t, g, "Granulate", "Sorcery", dn5B2Granulate, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(four) || g.Battlefield.Contains(one) {
		t.Error("each nonland artifact with mana value 4 or less should be destroyed")
	}
	if !g.Battlefield.Contains(five) {
		t.Error("a mana value 5 artifact survives")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("an artifact land survives")
	}
}
