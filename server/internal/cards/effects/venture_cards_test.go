package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// venture_cards_test.go pins the "venture into the dungeon" cards (ADR
// 0203 amendment of 2026-10-19, #4100): one test per card, each checking
// that the printed event ventures (the player is asked for a dungeon and
// enters its first room) and every other clause the card prints.

const (
	clatteringSkeletonsOracle  = "9e107382-f683-44ed-ada1-775d3ac79dd0"
	yuanTiMalisonOracle        = "f0e4dda1-2ffb-4353-a292-dcbc966967b4"
	shortcutSeekerOracle       = "27cfa916-282d-476c-8ee9-ca7f0d84f551"
	yuanTiFangBladeOracle      = "8ef06396-861d-4599-9976-6e5c04575456"
	planarAllyOracle           = "9ae30bd5-b638-4647-b59d-ffe08e77c182"
	triumphantAdventurerOracle = "6aac0ae4-d9a7-4c40-a39e-229186c84d74"
	veteranDungeoneerOracle    = "1ce7b966-2a10-4250-8c0d-525e73efe2f4"
	displacerBeastOracle       = "7f24c0b7-1d39-48b6-969f-ee97fdf99305"
	cloisterGargoyleOracle     = "e1e84b14-c521-4eb7-92ed-bde1131e8662"
	gloomStalkerOracle         = "16201410-86e9-40a7-88a6-c6c2fb2ea6bd"
	nadaarOracle               = "0e3f9345-d986-4d60-a343-45598ab59aff"
	dungeonMapOracle           = "20f1bba5-d398-4d6e-a76a-75e29df42aec"
	secretDoorOracle           = "a53baf90-8d0c-4900-9f24-893ad31628ec"
	rangersHawkOracle          = "44924f1e-edc4-481b-93f3-78606d4fd691"
	dungeonDescentOracle       = "f086a63c-0c62-4674-bd27-82e7aed12b1a"
	zombieOgreOracle           = "0898fb77-d304-496c-8f2e-1cfe50cc5507"
	wanderingTroubadourOracle  = "1a304703-6b48-4f3f-bd41-7f10ce0a751c"
	zaltoOracle                = "fb1c507d-ef94-438e-bee5-29eb48dc8e70"
	midnightPathlighterOracle  = "e49228d6-6488-4cbc-aa53-303afd3a549b"
	barTheGateOracle           = "11f55ec5-b9a3-4670-8afb-31d29478fb59"
	fatesReversalOracle        = "8cd88e87-6585-437a-bcad-b494fffc0544"
	cursedIdolOracle           = "aa7b41ba-08eb-40b8-a8d5-e67dc09c88ba"
	kickInTheDoorOracle        = "d3278ddb-c198-450d-946e-8547e5c96dd1"
)

// vcVentured settles the stack, entering Lost Mine of Phandelver when
// asked for a dungeon, and fails unless `player` ended in its first room.
func vcVentured(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	if left := frSettle(t, g, LostMineOfPhandelverKey); len(left) > 0 {
		t.Fatalf("no venture: the dungeon choice was never asked")
	}
	if room := initRoom(g, player); room != "cave-entrance" {
		t.Fatalf("room = %q, want cave-entrance", room)
	}
}

// vcNotVentured settles and fails if `player` was asked for a dungeon.
func vcNotVentured(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	frSettle(t, g)
	if c := latestOptionPickFor(g, player); c != nil {
		t.Fatalf("ventured: asked %q", c.Reason)
	}
	if room := initRoom(g, player); room != "" {
		t.Fatalf("ventured into %q", room)
	}
}

// vcCompleteADungeon walks Lost Mine of Phandelver to its end for
// `player`, completing it.
func vcCompleteADungeon(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	frVenture(t, g, player, LostMineOfPhandelverKey)
	frVenture(t, g, player, "goblin-lair")
	frVenture(t, g, player, "dark-pool")
	frVenture(t, g, player)
	if len(initCompleted(g, player)) == 0 {
		t.Fatalf("the dungeon was not completed")
	}
}

// --- enters ----------------------------------------------------------

func TestVentureCreaturesVentureOnEntry(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Veteran Dungeoneer", "Creature — Human Warrior", veteranDungeoneerOracle},
		{"Displacer Beast", "Creature — Cat Beast", displacerBeastOracle},
		{"Cloister Gargoyle", "Artifact Creature — Gargoyle", cloisterGargoyleOracle},
		{"Nadaar, Selfless Paladin", "Legendary Creature — Dragon Knight", nadaarOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			castAndResolveCreature(t, g, tc.name, tc.typeLine, tc.oracle)
			vcVentured(t, g, me.ID)
		})
	}
}

// Displacement returns the Beast to its owner's hand.
func TestDisplacerBeastReturnsToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	beast := pushDiesCreatureForTest(g, me.ID, "Displacer Beast", displacerBeastOracle, "Creature — Cat Beast", 3, 2)
	conduitMana(t, g, me, "{U}{C}{C}{C}")
	if err := activateOn(t, g, me.ID, beast, 0, uuid.Nil); err != nil {
		t.Fatalf("Displacement: %v", err)
	}
	frSettle(t, g)
	if z := initZone(g, beast); z != game.ZoneHand {
		t.Fatalf("the Beast is in %v, want hand", z)
	}
}

// --- dies ------------------------------------------------------------

func TestClatteringSkeletonsVenturesWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushDiesCreatureForTest(g, me.ID, "Clattering Skeletons", clatteringSkeletonsOracle, "Creature — Skeleton", 4, 3)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	vcVentured(t, g, me.ID)
}

// --- attacks ---------------------------------------------------------

func TestVentureCreaturesVentureWhenTheyAttack(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Planar Ally", "Creature — Angel", planarAllyOracle},
		{"Triumphant Adventurer", "Creature — Human Knight", triumphantAdventurerOracle},
		{"Nadaar, Selfless Paladin", "Legendary Creature — Dragon Knight", nadaarOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, them := g.Seats[0], g.Seats[1]
			id := pushDiesCreatureForTest(g, me.ID, tc.name, tc.oracle, tc.typeLine, 3, 3)
			advanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(id, them.ID); err != nil {
				t.Fatal(err)
			}
			// Leaving the declaration puts the attack trigger on the stack
			// and resolves it as far as the dungeon choice.
			_, _ = g.AdvanceStep()
			vcVentured(t, g, me.ID)
		})
	}
}

// --- combat damage to a player -----------------------------------------

func TestVentureCreaturesVentureOnCombatDamage(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Yuan-Ti Malison", "Creature — Snake Rogue", yuanTiMalisonOracle},
		{"Shortcut Seeker", "Creature — Human Rogue", shortcutSeekerOracle},
		{"Yuan-Ti Fang-Blade", "Creature — Snake Rogue", yuanTiFangBladeOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, them := g.Seats[0], g.Seats[1]
			id := pushDiesCreatureForTest(g, me.ID, tc.name, tc.oracle, tc.typeLine, 2, 2)
			attackWith(t, g, them.ID, id)
			vcVentured(t, g, me.ID)
		})
	}
}

// Midnight Pathlighter: two creatures connecting with one player is one
// venture (CR 603.2c); and a nonlegendary creature can't block them.
func TestMidnightPathlighterVenturesOncePerPlayerAndEvades(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushDiesCreatureForTest(g, me.ID, "Midnight Pathlighter", midnightPathlighterOracle, "Creature — Human Wizard", 2, 3)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	plain := pushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	legend := pushDiesCreatureForTest(g, them.ID, "Their Legend", "", "Legendary Creature — Human", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, them.ID); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(plain, a); err == nil {
		t.Errorf("a nonlegendary creature blocked")
	}
	if err := g.DeclareBlocker(legend, a); err != nil {
		t.Errorf("a legendary creature could not block: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	vcVentured(t, g, me.ID)
	if c := latestOptionPickFor(g, me.ID); c != nil {
		t.Fatalf("a second venture for the same player: %q", c.Reason)
	}
}

// --- statics -----------------------------------------------------------

// Cloister Gargoyle is a 0/4 until its controller completes a dungeon,
// then a 3/4 flier.
func TestCloisterGargoyleGrowsOnceADungeonIsCompleted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cloister Gargoyle", OracleID: cloisterGargoyleOracle,
		TypeLine: "Artifact Creature — Gargoyle", Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	if p := effectivePower(t, g, id); p != 0 || hasAbility(effectiveAbilities(t, g, id), "flying") {
		t.Fatalf("before: power %d, abilities %v", p, effectiveAbilities(t, g, id))
	}
	vcCompleteADungeon(t, g, me.ID)
	if p := effectivePower(t, g, id); p != 3 {
		t.Errorf("after: power %d, want 3", p)
	}
	if !hasAbility(effectiveAbilities(t, g, id), "flying") {
		t.Errorf("after: no flying")
	}
}

func TestGloomStalkerHasDoubleStrikeOnceADungeonIsCompleted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gloom Stalker", OracleID: gloomStalkerOracle,
		TypeLine: "Creature — Dwarf Ranger", Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	if hasAbility(effectiveAbilities(t, g, id), "double strike") {
		t.Fatalf("double strike before completing a dungeon")
	}
	vcCompleteADungeon(t, g, me.ID)
	if !hasAbility(effectiveAbilities(t, g, id), "double strike") {
		t.Errorf("no double strike after completing a dungeon")
	}
}

// Nadaar's anthem: other creatures you control, not Nadaar, not an
// opponent's.
func TestNadaarPumpsOtherCreaturesOnceADungeonIsCompleted(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	nadaar := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Nadaar, Selfless Paladin", OracleID: nadaarOracle,
		TypeLine: "Legendary Creature — Dragon Knight", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	if p := effectivePower(t, g, bear); p != 2 {
		t.Fatalf("before: bear power %d", p)
	}
	vcCompleteADungeon(t, g, me.ID)
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("my bear has power %d, want 3", p)
	}
	if p := effectivePower(t, g, nadaar); p != 3 {
		t.Errorf("Nadaar pumped itself: power %d", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("their bear pumped: power %d", p)
	}
}

// Triumphant Adventurer has first strike on its controller's turn only.
func TestTriumphantAdventurerHasFirstStrikeOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Triumphant Adventurer", OracleID: triumphantAdventurerOracle,
		TypeLine: "Creature — Human Knight", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Triumphant Adventurer", OracleID: triumphantAdventurerOracle,
		TypeLine: "Creature — Human Knight", Power: 1, Toughness: 1, Owner: them.ID, Controller: them.ID,
	})
	if !hasAbility(effectiveAbilities(t, g, mine), "first strike") {
		t.Errorf("no first strike on its controller's turn")
	}
	if hasAbility(effectiveAbilities(t, g, theirs), "first strike") {
		t.Errorf("first strike on another player's turn")
	}
	if !hasAbility(effectiveAbilities(t, g, mine), "deathtouch") {
		t.Errorf("no deathtouch")
	}
}

// --- activated --------------------------------------------------------

func TestDungeonMapAndSecretDoorVentureAtSorcerySpeed(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Dungeon Map", "Artifact", dungeonMapOracle},
		{"Secret Door", "Artifact Creature — Wall", secretDoorOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := pushDiesCreatureForTest(g, me.ID, tc.name, tc.oracle, tc.typeLine, 0, 4)
			conduitMana(t, g, me, "{U}{C}{C}{C}{C}")
			if err := activateOn(t, g, me.ID, id, 0, uuid.Nil); err == nil {
				t.Fatalf("activated outside a main phase")
			}
			advanceToMain(t, g)
			conduitMana(t, g, me, "{U}{C}{C}{C}{C}")
			if err := activateOn(t, g, me.ID, id, 0, uuid.Nil); err != nil {
				t.Fatalf("activate: %v", err)
			}
			vcVentured(t, g, me.ID)
		})
	}
}

func TestRangersHawkTapsAnotherCreatureToVenture(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hawk := pushDiesCreatureForTest(g, me.ID, "Ranger's Hawk", rangersHawkOracle, "Creature — Bird", 1, 1)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceToMain(t, g)
	conduitMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, hawk, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !findBattlefieldCardForTest(g, bear).Tapped || !findBattlefieldCardForTest(g, hawk).Tapped {
		t.Errorf("the Hawk and the Bear were not both tapped")
	}
	vcVentured(t, g, me.ID)
}

func TestDungeonDescentEntersTappedAndTapsALegendToVenture(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushDiesCreatureForTest(g, me.ID, "Dungeon Descent", dungeonDescentOracle, "Land", 0, 0)
	plain := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	legend := pushDiesCreatureForTest(g, me.ID, "Legend", "", "Legendary Creature — Human", 2, 2)
	advanceToMain(t, g)
	conduitMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{plain}}); err == nil {
		t.Errorf("a nonlegendary creature paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{legend}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	vcVentured(t, g, me.ID)

	// The land enters tapped.
	castCatalogSpell(t, g, "Dungeon Descent", "Land", dungeonDescentOracle, nil)
	frSettle(t, g)
	tapped := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Dungeon Descent" && c.InstanceID != land && c.Tapped {
			tapped++
		}
	}
	if tapped != 1 {
		t.Errorf("the played Dungeon Descent entered untapped")
	}
}

// --- end step ----------------------------------------------------------

func TestZombieOgreVenturesOnlyIfACreatureDied(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushDiesCreatureForTest(g, me.ID, "Zombie Ogre", zombieOgreOracle, "Creature — Zombie Ogre", 3, 5)
	advanceToEndStep(t, g)
	vcNotVentured(t, g, me.ID)

	g2 := newCatalogGame(t)
	me2, them2 := g2.Seats[0], g2.Seats[1]
	pushDiesCreatureForTest(g2, me2.ID, "Zombie Ogre", zombieOgreOracle, "Creature — Zombie Ogre", 3, 5)
	theirs := pushVanillaCreature(g2, them2.ID, "Their Bear", 2, 2)
	g2.WithWriteLock(func() {
		if err := g2.DestroyPermanentForEffect(theirs); err != nil {
			t.Fatal(err)
		}
	})
	frSettle(t, g2)
	advanceToEndStep(t, g2)
	vcVentured(t, g2, me2.ID)
	_ = them
}

func TestWanderingTroubadourVenturesOnlyAfterALandEntered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushDiesCreatureForTest(g, me.ID, "Wandering Troubadour", wanderingTroubadourOracle, "Creature — Dragon Bard", 4, 2)
	advanceToEndStep(t, g)
	vcNotVentured(t, g, me.ID)

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	pushDiesCreatureForTest(g2, me2.ID, "Wandering Troubadour", wanderingTroubadourOracle, "Creature — Dragon Bard", 4, 2)
	castCatalogSpell(t, g2, "Forest", "Basic Land — Forest", "", nil)
	frSettle(t, g2)
	advanceToEndStep(t, g2)
	vcVentured(t, g2, me2.ID)
}

// --- dealt damage --------------------------------------------------------

func TestZaltoVenturesWhenDealtDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	zalto := pushDiesCreatureForTest(g, me.ID, "Zalto, Fire Giant Duke", zaltoOracle, "Legendary Creature — Giant Barbarian", 7, 3)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, zalto, 1); err != nil {
			t.Fatal(err)
		}
	})
	vcVentured(t, g, me.ID)
}

// --- spells -------------------------------------------------------------

func TestBarTheGateCountersACreatureSpellAndVentures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, b13HandCard(me, "Bar the Gate", "Instant", barTheGateOracle),
		game.CastSpellParams{Targets: b16TargetCard(bear)}); err != nil {
		t.Fatalf("Bar the Gate: %v", err)
	}
	vcVentured(t, g, me.ID)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature spell was not countered")
	}
}

func TestFatesReversalReturnsACreatureCardAndVentures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	castCatalogSpell(t, g, "Fates' Reversal", "Sorcery", fatesReversalOracle, b16TargetCard(dead))
	vcVentured(t, g, me.ID)
	if z := initZone(g, dead); z != game.ZoneHand {
		t.Fatalf("the creature card is in %v, want hand", z)
	}

	// With no target chosen it still ventures.
	g2 := newCatalogGame(t)
	castCatalogSpell(t, g2, "Fates' Reversal", "Sorcery", fatesReversalOracle, nil)
	vcVentured(t, g2, g2.Seats[0].ID)
}

func TestYouFindACursedIdolStealsItsEyes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castModal(t, g, "You Find a Cursed Idol", "Sorcery", cursedIdolOracle, []int{2}, nil)
	vcVentured(t, g, me.ID)
	if n := frTokens(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("Treasures = %d, want 1", n)
	}

	g2 := newCatalogGame(t)
	rock := b12Permanent(g2, g2.Seats[1].ID, "Their Rock", "Artifact")
	castModal(t, g2, "You Find a Cursed Idol", "Sorcery", cursedIdolOracle, []int{0}, b16TargetCard(rock))
	vcNotVentured(t, g2, g2.Seats[0].ID)
	if g2.Battlefield.Contains(rock) {
		t.Error("Smash It did not destroy the artifact")
	}
}

func TestKickInTheDoorCountersHastesBarsWallsAndVentures(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	wall := pushDiesCreatureForTest(g, them.ID, "Their Wall", "", "Creature — Wall", 0, 4)
	other := pushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	castCatalogSpell(t, g, "Kick in the Door", "Sorcery", kickInTheDoorOracle, b16TargetCard(bear))
	vcVentured(t, g, me.ID)
	if got := findBattlefieldCardForTest(g, bear).Counters["+1/+1"]; got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
	if !hasAbility(effectiveAbilities(t, g, bear), "haste") {
		t.Errorf("no haste")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, them.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err == nil {
		t.Errorf("a Wall blocked")
	}
	if err := g.DeclareBlocker(other, bear); err != nil {
		t.Errorf("a non-Wall could not block: %v", err)
	}
}
