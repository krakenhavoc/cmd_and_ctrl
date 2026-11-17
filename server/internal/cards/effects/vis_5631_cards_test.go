package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vis_5631_cards_test.go — a batch of Visions cards (#5631). Every test
// plays the real catalog entry through the engine.

const (
	visArchangelOracle             = "9971697b-2acc-4bc2-a44e-074d03a51df7"
	visFreewindFalconOracle        = "8cdda36f-5e82-4c2f-8eaa-75b9bbc93a02"
	visLongbowArcherOracle         = "51531db1-ba59-4ccc-9a91-54de2887fc8b"
	visPantherWarriorsOracle       = "ef7788af-8edc-46df-a5b6-895c734ea423"
	visPythonOracle                = "d8a0c3ff-7042-4b52-a216-15e170c8094f"
	visScalebanesEliteOracle       = "c04dc6fe-06fc-4e5b-87da-6b8b262a214f"
	visTempestDrakeOracle          = "ff0086e0-706f-4474-9b5e-a1591235bf9b"
	visWarthogOracle               = "ae1f22a5-89cf-440c-b90d-fed194b2dee5"
	visSuqAtaLancerOracle          = "f348b326-8a73-4df9-8faf-e271a36ce373"
	visFallenAskariOracle          = "4956008e-d912-48ba-9b82-176258281060"
	visHulkingCyclopsOracle        = "04936918-7ef3-42fa-b0cd-87d18cb69d2c"
	visRiverBoaOracle              = "86187aed-77ce-4b2d-aab0-0a8f807a9451"
	visWakeOfVulturesOracle        = "2948b1d5-c12c-45c4-a3e8-aa470fcee1af"
	visDarajaGriffinOracle         = "daa0d720-cac1-487a-93e2-33e18f91f19b"
	visViashivanDragonOracle       = "86fb1fc1-7784-4c94-a80c-21ad87e50fc4"
	visKeeperOfKookusOracle        = "97417270-e525-4261-b7d9-0dde3360e666"
	visJamuraanLionOracle          = "b558541f-ae57-41f7-a5bd-e23fa35e68b4"
	visInfantryVeteranOracle       = "42798e2b-9a5c-4f93-8f77-b7bb7a916d07"
	visSisaysRingOracle            = "8f5822ae-651f-410e-9316-5522eeb52d72"
	visRainbowEfreetOracle         = "7c158405-d3ef-4c10-a504-7b061ef71a74"
	visTeferisHonorGuardOracle     = "8a3a8a02-b7da-4648-a772-c9efb7c332da"
	visCreepingMoldOracle          = "59180e94-ccdf-4d9f-9a4a-fe55497d0d63"
	visElvenCacheOracle            = "8a922366-ee2a-4ddb-a904-683dcf4f191a"
	visInspirationOracle           = "8f32ceb2-92c2-4dde-bf73-40bb79c3fcef"
	visUndoOracle                  = "b446c3bd-3c7d-4d59-ba59-3dd80818465e"
	visWarriorsHonorOracle         = "c3c9da49-ef4f-48c7-99c8-0f226b485285"
	visMiraculousRecoveryOracle    = "cea73a4b-e10a-4517-a92b-24a08321009f"
	visRetributionOfTheMeekOracle  = "61f2083b-3018-43a4-8852-8e6d2a97d5d6"
	visTremorOracle                = "4281a153-665b-44cf-bcff-9cf88c14c2b5"
	visManOWarOracle               = "67a3541c-8408-40c8-b44f-90035b860f57"
	visUktabiOrangutanOracle       = "6cc07a7e-e70c-45c5-83c5-a2704e57dbc5"
	visNekrataalOracle             = "d40adf5d-5edf-48a0-98f3-fcdbf1b5f25c"
	visCoralAtollOracle            = "3f347ebf-e0d2-4ae0-ad84-df7a460404e0"
	visDormantVolcanoOracle        = "38ba1956-5505-4a7a-b6af-e75715b1401f"
	visEvergladesOracle            = "ef3b8b0c-cea7-4bae-934c-9c65fd64245d"
	visJungleBasinOracle           = "f922f90a-b1a2-4630-9266-40726ca89f74"
	visKarooOracle                 = "d4e875d9-2245-470d-aa2f-1dfe66ce2d15"
	visHumanKnightTypeLine         = "Creature — Human Knight"
	visShieldedBoaTypeLine         = "Creature — Snake"
	visLightningBoltForKeeperLabel = "Lightning Bolt"
)

// visShield gives a River Boa pushed at the given size a regeneration
// shield from its own ability.
func visShield(t *testing.T, g *game.Game, owner *game.Player, power, toughness int) uuid.UUID {
	t.Helper()
	boa := b12Push(g, owner.ID, "River Boa", visShieldedBoaTypeLine, visRiverBoaOracle, power, toughness)
	hmlSetColors(g, boa, "G")
	apaMana(owner, "G")
	if err := g.ActivateCatalogAbility(owner.ID, boa, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("River Boa's regeneration: %v", err)
	}
	passPriorityAroundTable(t, g)
	return boa
}

// visBattlefieldNamed is the first permanent named `name` that `owner`
// controls, or false.
func visBattlefieldNamed(g *game.Game, owner uuid.UUID, name string) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}

// --- every card is in the catalog ----------------------------------------

func TestVis5631EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		visArchangelOracle, visFreewindFalconOracle, visLongbowArcherOracle, visPantherWarriorsOracle,
		visPythonOracle, visScalebanesEliteOracle, visTempestDrakeOracle, visWarthogOracle,
		visSuqAtaLancerOracle, visFallenAskariOracle, visHulkingCyclopsOracle, visRiverBoaOracle,
		visWakeOfVulturesOracle, visDarajaGriffinOracle, visViashivanDragonOracle, visKeeperOfKookusOracle,
		visJamuraanLionOracle, visInfantryVeteranOracle, visSisaysRingOracle, visRainbowEfreetOracle,
		visTeferisHonorGuardOracle, visCreepingMoldOracle, visElvenCacheOracle, visInspirationOracle,
		visUndoOracle, visWarriorsHonorOracle, visMiraculousRecoveryOracle, visRetributionOfTheMeekOracle,
		visTremorOracle, visManOWarOracle, visUktabiOrangutanOracle, visNekrataalOracle,
		visCoralAtollOracle, visDormantVolcanoOracle, visEvergladesOracle, visJungleBasinOracle, visKarooOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- the keyword bodies ----------------------------------------------------

func TestVis5631KeywordCreatures(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Archangel", "Creature — Angel", visArchangelOracle, []string{"flying", "vigilance"}},
		{"Tempest Drake", "Creature — Drake", visTempestDrakeOracle, []string{"flying", "vigilance"}},
		{"Freewind Falcon", "Creature — Bird", visFreewindFalconOracle, []string{"flying", "protection from red"}},
		{"Longbow Archer", "Creature — Human Soldier Archer", visLongbowArcherOracle, []string{"reach", "first strike"}},
		{"Scalebane's Elite", "Creature — Human Soldier", visScalebanesEliteOracle, []string{"protection from black"}},
		{"Warthog", "Creature — Boar", visWarthogOracle, []string{"swampwalk"}},
		{"Suq'Ata Lancer", visHumanKnightTypeLine, visSuqAtaLancerOracle, []string{"haste"}},
		{"River Boa", "Creature — Snake", visRiverBoaOracle, []string{"islandwalk"}},
		{"Wake of Vultures", "Creature — Bird", visWakeOfVulturesOracle, []string{"flying"}},
		{"Daraja Griffin", "Creature — Griffin", visDarajaGriffinOracle, []string{"flying"}},
		{"Viashivan Dragon", "Creature — Dragon", visViashivanDragonOracle, []string{"flying"}},
		{"Rainbow Efreet", "Creature — Efreet", visRainbowEfreetOracle, []string{"flying"}},
		{"Nekrataal", "Creature — Human Assassin", visNekrataalOracle, []string{"first strike"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id := b12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			abilities := effectiveAbilities(t, g, id)
			for _, kw := range tc.keywords {
				if !hasAbility(abilities, kw) {
					t.Errorf("%s has no %s", tc.name, kw)
				}
			}
		})
	}
}

// Panther Warriors and Python are vanilla: no keyword and no ability rows.
func TestVis5631VanillaBodies(t *testing.T) {
	for _, oracle := range []string{visPantherWarriorsOracle, visPythonOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not in the catalog", oracle)
		}
		if len(spec.PrintedKeywords) != 0 || len(spec.Activated) != 0 || len(spec.Static) != 0 || len(spec.Triggered) != 0 {
			t.Errorf("%s declares rules: %+v", spec.Name, spec)
		}
	}
}

// Freewind Falcon's protection from red: a red spell can't target it.
func TestVis5631FreewindFalconCantBeTargetedByRed(t *testing.T) {
	g := newCatalogGame(t)
	falcon := b12Push(g, g.Seats[1].ID, "Freewind Falcon", "Creature — Bird", visFreewindFalconOracle, 1, 1)
	hmlSetColors(g, falcon, "W")
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	visBoltFails(t, g, me, falcon)
}

// visBoltFails casts a Lightning Bolt from `me` at `target` and wants
// the cast refused.
func visBoltFails(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: visLightningBoltForKeeperLabel, TypeLine: "Instant", OracleID: lightningBoltOracle,
		ManaCost: "{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID,
	})
	me.ManaPool.AddMana(game.ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: cardTarget(target)}); err == nil {
		t.Error("Lightning Bolt targeted a creature with protection from red")
	}
}

// --- can't block, and flanking ----------------------------------------------

func TestVis5631CantBlockBodies(t *testing.T) {
	cases := []struct{ name, typeLine, oracle string }{
		{"Hulking Cyclops", "Creature — Cyclops", visHulkingCyclopsOracle},
		{"Fallen Askari", visHumanKnightTypeLine, visFallenAskariOracle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			attacker := b12Creature(g, me.ID, "Raider", "Creature — Human", 2, 2)
			body := b12Push(g, opp.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			other := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
			brAttack(t, g, attacker)
			brRefusal(t, g.DeclareBlocker(body, attacker), game.BlockReasonCantBlock)
			if err := g.DeclareBlocker(other, attacker); err != nil {
				t.Fatalf("another creature can't block: %v", err)
			}
		})
	}
}

// Flanking (CR 702.25a): each blocker without flanking gets -1/-1 until
// end of turn; one with flanking does not.
func TestVis5631Flanking(t *testing.T) {
	cases := []struct{ name, oracle string }{
		{"Fallen Askari", visFallenAskariOracle},
		{"Suq'Ata Lancer", visSuqAtaLancerOracle},
		{"Teferi's Honor Guard", visTeferisHonorGuardOracle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			knight := apaPush(g, me.ID, me.ID, game.Card{Name: tc.name, OracleID: tc.oracle, TypeLine: visHumanKnightTypeLine, Power: 2, Toughness: 2, Keywords: []string{"flanking"}})
			plain := apaPush(g, opp.ID, opp.ID, apaCreature("Plain", "", 2, 2))
			flanker := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Flanker", TypeLine: "Creature — Test", Power: 2, Toughness: 2, Keywords: []string{"flanking"}})
			g.WithWriteLock(func() {
				for i, b := range []uuid.UUID{plain, flanker} {
					g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: opp.ID, CardID: b, Target: knight, Amount: i + 1})
				}
			})
			passPriorityAroundTable(t, g)
			if got := apaLive(g, plain).CurrentToughness(); got != 1 {
				t.Errorf("the blocker without flanking has toughness %d, want 1", got)
			}
			if got := apaLive(g, flanker).CurrentToughness(); got != 2 {
				t.Errorf("the blocker with flanking has toughness %d, want 2", got)
			}
		})
	}
}

// --- regeneration ----------------------------------------------------------------

func TestVis5631RiverBoaRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	boa := visShield(t, g, me, 2, 1)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(boa); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(boa) {
		t.Fatal("the shielded Boa was destroyed")
	}
	if !hmlTapped(t, g, boa) {
		t.Error("regeneration did not tap the Boa")
	}
}

func TestVis5631WakeOfVulturesSacrificesToRegenerate(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	wake := b12Push(g, me.ID, "Wake of Vultures", "Creature — Bird", visWakeOfVulturesOracle, 3, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	apaMana(me, "C", "B")
	apaActivate(t, g, me, wake, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}})
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear was not sacrificed to pay the cost")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(wake); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(wake) {
		t.Error("the shielded Vultures were destroyed")
	}
}

// --- the other activated abilities --------------------------------------------

func TestVis5631DarajaGriffinDestroysOnlyABlackCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	griffin := b12Push(g, me.ID, "Daraja Griffin", "Creature — Griffin", visDarajaGriffinOracle, 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	hmlSetColors(g, bear, "G")
	zombie := b12Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2)
	hmlSetColors(g, zombie, "B")
	if err := g.ActivateCatalogAbility(me.ID, griffin, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(bear)}); err == nil {
		t.Fatal("the Griffin targeted a green creature")
	}
	if !g.Battlefield.Contains(griffin) {
		t.Fatal("a refused activation sacrificed the Griffin")
	}
	apaActivate(t, g, me, griffin, 0, game.ActivateAbilityParams{Targets: cardTarget(zombie)})
	if g.Battlefield.Contains(griffin) {
		t.Error("the Griffin was not sacrificed")
	}
	if g.Battlefield.Contains(zombie) {
		t.Error("the black creature survived")
	}
}

func TestVis5631ViashivanDragonPumpsEachWay(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	dragon := b12Push(g, me.ID, "Viashivan Dragon", "Creature — Dragon", visViashivanDragonOracle, 4, 4)
	apaMana(me, "R")
	apaActivate(t, g, me, dragon, 0, game.ActivateAbilityParams{})
	apaMana(me, "G")
	apaActivate(t, g, me, dragon, 1, game.ActivateAbilityParams{})
	apaMana(me, "G")
	apaActivate(t, g, me, dragon, 1, game.ActivateAbilityParams{})
	if p, tt := ptOf(t, g, dragon); p != 5 || tt != 6 {
		t.Errorf("after the pumps %d/%d, want 5/6", p, tt)
	}
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if p, tt := ptOf(t, g, dragon); p != 4 || tt != 4 {
		t.Errorf("next turn %d/%d, want 4/4", p, tt)
	}
}

func TestVis5631KeeperOfKookusGainsProtectionFromRed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	keeper := b12Push(g, me.ID, "Keeper of Kookus", "Creature — Goblin", visKeeperOfKookusOracle, 1, 1)
	if hasEffectiveKeyword(t, g, keeper, "protection from red") {
		t.Fatal("protection before the ability")
	}
	apaMana(me, "R")
	apaActivate(t, g, me, keeper, 0, game.ActivateAbilityParams{})
	if !hasEffectiveKeyword(t, g, keeper, "protection from red") {
		t.Fatal("no protection from red after the ability")
	}
	visBoltFails(t, g, me, keeper)
	advanceToMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if hasEffectiveKeyword(t, g, keeper, "protection from red") {
		t.Error("protection outlasted the turn")
	}
}

func TestVis5631JamuraanLionStopsATargetFromBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	lion := pushCatalogPermanent(g, me.ID, "Jamuraan Lion", "Creature — Cat", visJamuraanLionOracle, false)
	attacker := b12Creature(g, me.ID, "Raider", "Creature — Human", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	other := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	apaMana(me, "W")
	apaActivate(t, g, me, lion, 0, game.ActivateAbilityParams{Targets: cardTarget(wall)})
	if !hmlTapped(t, g, lion) {
		t.Error("the Lion was not tapped to pay the cost")
	}
	brAttack(t, g, attacker)
	brRefusal(t, g.DeclareBlocker(wall, attacker), game.BlockReasonCantBlock)
	if err := g.DeclareBlocker(other, attacker); err != nil {
		t.Fatalf("an untargeted creature can't block: %v", err)
	}
}

func TestVis5631InfantryVeteranPumpsOnlyAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	veteran := pushCatalogPermanent(g, me.ID, "Infantry Veteran", "Creature — Human Soldier", visInfantryVeteranOracle, false)
	attacker := b12Creature(g, me.ID, "Raider", "Creature — Human", 2, 2)
	idle := b12Creature(g, me.ID, "Idle", "Creature — Human", 2, 2)
	brAttack(t, g, attacker)
	if err := g.ActivateCatalogAbility(me.ID, veteran, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(idle)}); err == nil {
		t.Fatal("the Veteran targeted a creature that isn't attacking")
	}
	apaActivate(t, g, me, veteran, 0, game.ActivateAbilityParams{Targets: cardTarget(attacker)})
	if p, tt := ptOf(t, g, attacker); p != 3 || tt != 3 {
		t.Errorf("the attacker is %d/%d, want 3/3", p, tt)
	}
}

func TestVis5631SisaysRingAddsTwoColorless(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	ring := pushCatalogPermanent(g, me.ID, "Sisay's Ring", "Artifact", visSisaysRingOracle, false)
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if n := me.ManaPool.CountColor("C"); n != 2 || me.ManaPool.Total() != 2 {
		t.Errorf("pool holds %d colorless of %d, want 2 of 2", n, me.ManaPool.Total())
	}
	if !hmlTapped(t, g, ring) {
		t.Error("the Ring did not tap")
	}
}

// Rainbow Efreet and Teferi's Honor Guard phase themselves out, and come
// back in their controller's next untap step.
func TestVis5631ThisCreaturePhasesOut(t *testing.T) {
	cases := []struct{ name, typeLine, oracle string }{
		{"Rainbow Efreet", "Creature — Efreet", visRainbowEfreetOracle},
		{"Teferi's Honor Guard", visHumanKnightTypeLine, visTeferisHonorGuardOracle},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			seat := g.Turn.ActiveSeat
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
			apaMana(me, "U", "U")
			apaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if g.Battlefield.Contains(id) || !g.PhasedOut.Contains(id) {
				t.Fatal("it did not phase out")
			}
			advanceToMainOf(t, g, (seat+1)%len(g.Seats))
			if !g.PhasedOut.Contains(id) {
				t.Fatal("it phased in during another player's untap step")
			}
			advanceToMainOf(t, g, seat)
			if !g.Battlefield.Contains(id) {
				t.Error("it did not phase in during its controller's untap step")
			}
		})
	}
}

// --- the spells ------------------------------------------------------------------

func TestVis5631CreepingMoldDestroysALandButNotACreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	land := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	mold := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mold, Name: "Creeping Mold", TypeLine: "Sorcery", OracleID: visCreepingMoldOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, mold, game.CastSpellParams{Targets: cardTarget(bear)}); err == nil {
		t.Fatal("Creeping Mold targeted a creature")
	}
	castCatalogSpell(t, g, "Creeping Mold", "Sorcery", visCreepingMoldOracle, cardTarget(land))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land survived Creeping Mold")
	}
}

func TestVis5631ElvenCacheReturnsACardToHand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	card := pushGraveyardCardForTest(me, "Old Friend")
	castCatalogSpell(t, g, "Elven Cache", "Sorcery", visElvenCacheOracle, cardTarget(card))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) || me.Graveyard.Contains(card) {
		t.Error("the card did not return to hand")
	}
}

func TestVis5631InspirationMakesTargetPlayerDrawTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := top100Opponent(g)
	before := opp.Hand.Size()
	castCatalogSpell(t, g, "Inspiration", "Instant", visInspirationOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != before+2 {
		t.Errorf("the target's hand is %d, want %d", got, before+2)
	}
}

func TestVis5631UndoBouncesTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Undo", "Sorcery", visUndoOracle, cardTargets(mine, theirs))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("the two creatures are not in their owners' hands")
	}
}

func TestVis5631WarriorsHonorPumpsOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Warrior's Honor", "Instant", visWarriorsHonorOracle, nil)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, mine); p != 3 || tt != 3 {
		t.Errorf("yours is %d/%d, want 3/3", p, tt)
	}
	if p, tt := ptOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("theirs is %d/%d, want 2/2", p, tt)
	}
}

func TestVis5631MiraculousRecoveryReturnsWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	card := pushGraveyardCardForTest(me, "Fallen Hero")
	castCatalogSpell(t, g, "Miraculous Recovery", "Instant", visMiraculousRecoveryOracle, cardTarget(card))
	passPriorityAroundTable(t, g)
	hero, ok := visBattlefieldNamed(g, me.ID, "Fallen Hero")
	if !ok {
		t.Fatal("the creature card did not return to the battlefield")
	}
	if got := hero.Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("%d +1/+1 counters on it, want 1", got)
	}
}

func TestVis5631RetributionOfTheMeekDestroysOnlyTheBig(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := top100ActiveSeat(g), top100Opponent(g)
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	shielded := visShield(t, g, me, 5, 5)
	castCatalogSpell(t, g, "Retribution of the Meek", "Sorcery", visRetributionOfTheMeekOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("a 4-power creature survived")
	}
	if g.Battlefield.Contains(shielded) {
		t.Error("a regeneration shield saved a 5-power creature")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("a 3-power creature was destroyed")
	}
}

func TestVis5631TremorSparesFlyers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := b12Creature(g, opp.ID, "Walker", "Creature — Human", 1, 1)
	flyer := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	life := me.Life
	castCatalogSpell(t, g, "Tremor", "Sorcery", visTremorOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the creature without flying survived")
	}
	if !g.Battlefield.Contains(flyer) {
		t.Error("the flyer was dealt damage")
	}
	if me.Life != life {
		t.Errorf("the caster lost life: %d -> %d", life, me.Life)
	}
}

// --- the enters triggers ------------------------------------------------------

func TestVis5631ManOWarBouncesATargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	jelly := castHoldCreature(t, g, "Man-o'-War", "Creature — Jellyfish", visManOWarOracle, 2, 2)
	b04WaitForPick(t, g, me.ID)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the target was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(jelly) {
		t.Error("the Man-o'-War left the battlefield")
	}
}

func TestVis5631UktabiOrangutanDestroysAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Uktabi Orangutan", "Creature — Ape", visUktabiOrangutanOracle, 2, 2)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a creature was accepted as the artifact target")
	}
	answerPickTarget(t, g, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestVis5631NekrataalDestroysPastRegeneration(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	shielded := visShield(t, g, opp, 2, 1)
	zombie := b12Creature(g, opp.ID, "Zombie", "Creature — Zombie", 2, 2)
	hmlSetColors(g, zombie, "B")
	golem := b12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	castHoldCreature(t, g, "Nekrataal", "Creature — Human Assassin", visNekrataalOracle, 2, 1)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	for _, bad := range []uuid.UUID{zombie, golem} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as a target", bad)
		}
	}
	answerPickTarget(t, g, shielded)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(shielded) {
		t.Error("a regeneration shield saved Nekrataal's target")
	}
}

// --- the land cycle ---------------------------------------------------------------

type visKaroo struct {
	name, oracle, landType, color string
}

var visKaroos = []visKaroo{
	{"Coral Atoll", visCoralAtollOracle, "Island", "U"},
	{"Dormant Volcano", visDormantVolcanoOracle, "Mountain", "R"},
	{"Everglades", visEvergladesOracle, "Swamp", "B"},
	{"Jungle Basin", visJungleBasinOracle, "Forest", "G"},
	{"Karoo", visKarooOracle, "Plains", "W"},
}

func TestVis5631KarooTapsForColorlessAndItsColor(t *testing.T) {
	for _, tc := range visKaroos {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := top100ActiveSeat(g)
			land := pushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
				t.Fatal(err)
			}
			if me.ManaPool.CountColor("C") != 1 || me.ManaPool.CountColor(tc.color) != 1 || me.ManaPool.Total() != 2 {
				t.Errorf("pool holds %d, want {C}{%s}", me.ManaPool.Total(), tc.color)
			}
		})
	}
}

// Only an untapped land of the type is offered; returning it keeps the
// land, which entered tapped.
func TestVis5631KarooReturnsAnUntappedLandOfItsType(t *testing.T) {
	for _, tc := range visKaroos {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			basic := seedLandOnBattlefield(g, me.ID, tc.landType, "Basic Land — "+tc.landType)
			tapped := seedLandOnBattlefield(g, me.ID, tc.landType, "Basic Land — "+tc.landType)
			g.WithWriteLock(func() { findBattlefieldCardForTest(g, tapped).Tapped = true })
			wrong := seedLandOnBattlefield(g, me.ID, "Wastes", "Basic Land")
			id := playLandFromHand(t, g, tc.name, tc.oracle)
			top100AssertEnteredTapped(t, g, id, tc.name)
			passPriorityAroundTable(t, g)
			pick := latestChooseCardsFor(g, me.ID)
			if pick == nil {
				t.Fatal("no prompt to return a land")
			}
			if !slices.Contains(pick.ChooseCards, basic) || slices.Contains(pick.ChooseCards, tapped) || slices.Contains(pick.ChooseCards, wrong) || slices.Contains(pick.ChooseCards, id) {
				t.Fatalf("offered %v; want only the untapped %s", pick.ChooseCards, tc.landType)
			}
			if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{basic}); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) {
				t.Errorf("%s was sacrificed after the land was returned", tc.name)
			}
			if !me.Hand.Contains(basic) {
				t.Errorf("the %s is not in its owner's hand", tc.landType)
			}
		})
	}
}

// Declining, or having no untapped land of the type, sacrifices it.
func TestVis5631KarooIsSacrificedWithoutThePayment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	island := seedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	atoll := playLandFromHand(t, g, "Coral Atoll", visCoralAtollOracle)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to return an Island")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(atoll) {
		t.Error("Coral Atoll stayed after the payment was declined")
	}
	if !g.Battlefield.Contains(island) {
		t.Error("the Island moved")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	tapped := seedLandOnBattlefield(g2, me2.ID, "Plains", "Basic Land — Plains")
	g2.WithWriteLock(func() { findBattlefieldCardForTest(g2, tapped).Tapped = true })
	karoo := playLandFromHand(t, g2, "Karoo", visKarooOracle)
	passPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(karoo) {
		t.Error("Karoo stayed with only a tapped Plains")
	}
}
