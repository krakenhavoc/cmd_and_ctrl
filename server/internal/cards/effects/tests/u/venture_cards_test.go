package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// vcVentured settles the stack, entering Lost Mine of Phandelver when
// asked for a dungeon, and fails unless `player` ended in its first room.
func vcVentured(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	if left := effectstest.FrSettle(t, g, effects.LostMineOfPhandelverKey); len(left) > 0 {
		t.Fatalf("no venture: the dungeon choice was never asked")
	}
	if room := effectstest.InitRoom(g, player); room != "cave-entrance" {
		t.Fatalf("room = %q, want cave-entrance", room)
	}
}

// vcNotVentured settles and fails if `player` was asked for a dungeon.
func vcNotVentured(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	effectstest.FrSettle(t, g)
	if c := effectstest.LatestOptionPickFor(g, player); c != nil {
		t.Fatalf("ventured: asked %q", c.Reason)
	}
	if room := effectstest.InitRoom(g, player); room != "" {
		t.Fatalf("ventured into %q", room)
	}
}

// vcCompleteADungeon walks Lost Mine of Phandelver to its end for
// `player`, completing it.
func vcCompleteADungeon(t *testing.T, g *game.Game, player uuid.UUID) {
	t.Helper()
	effectstest.FrVenture(t, g, player, effects.LostMineOfPhandelverKey)
	effectstest.FrVenture(t, g, player, "goblin-lair")
	effectstest.FrVenture(t, g, player, "dark-pool")
	effectstest.FrVenture(t, g, player)
	if len(effectstest.InitCompleted(g, player)) == 0 {
		t.Fatalf("the dungeon was not completed")
	}
}

// --- enters ----------------------------------------------------------

func TestVentureCreaturesVentureOnEntry(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Veteran Dungeoneer", "Creature — Human Warrior", effectstest.VeteranDungeoneerOracle},
		{"Displacer Beast", "Creature — Cat Beast", effectstest.DisplacerBeastOracle},
		{"Cloister Gargoyle", "Artifact Creature — Gargoyle", effectstest.CloisterGargoyleOracle},
		{"Nadaar, Selfless Paladin", "Legendary Creature — Dragon Knight", effectstest.NadaarOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.CastAndResolveCreature(t, g, tc.name, tc.typeLine, tc.oracle)
			vcVentured(t, g, me.ID)
		})
	}
}

// Displacement returns the Beast to its owner's hand.
func TestDisplacerBeastReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	beast := effectstest.PushDiesCreatureForTest(g, me.ID, "Displacer Beast", effectstest.DisplacerBeastOracle, "Creature — Cat Beast", 3, 2)
	effectstest.ConduitMana(t, g, me, "{U}{C}{C}{C}")
	if err := effectstest.ActivateOn(t, g, me.ID, beast, 0, uuid.Nil); err != nil {
		t.Fatalf("Displacement: %v", err)
	}
	effectstest.FrSettle(t, g)
	if z := effectstest.InitZone(g, beast); z != game.ZoneHand {
		t.Fatalf("the Beast is in %v, want hand", z)
	}
}

// --- dies ------------------------------------------------------------

func TestClatteringSkeletonsVenturesWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushDiesCreatureForTest(g, me.ID, "Clattering Skeletons", effectstest.ClatteringSkeletonsOracle, "Creature — Skeleton", 4, 3)
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
		{"Planar Ally", "Creature — Angel", effectstest.PlanarAllyOracle},
		{"Triumphant Adventurer", "Creature — Human Knight", effectstest.TriumphantAdventurerOracle},
		{"Nadaar, Selfless Paladin", "Legendary Creature — Dragon Knight", effectstest.NadaarOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, them := g.Seats[0], g.Seats[1]
			id := effectstest.PushDiesCreatureForTest(g, me.ID, tc.name, tc.oracle, tc.typeLine, 3, 3)
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
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
		{"Yuan-Ti Malison", "Creature — Snake Rogue", effectstest.YuanTiMalisonOracle},
		{"Shortcut Seeker", "Creature — Human Rogue", effectstest.ShortcutSeekerOracle},
		{"Yuan-Ti Fang-Blade", "Creature — Snake Rogue", effectstest.YuanTiFangBladeOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, them := g.Seats[0], g.Seats[1]
			id := effectstest.PushDiesCreatureForTest(g, me.ID, tc.name, tc.oracle, tc.typeLine, 2, 2)
			effectstest.AttackWith(t, g, them.ID, id)
			vcVentured(t, g, me.ID)
		})
	}
}

// Midnight Pathlighter: two creatures connecting with one player is one
// venture (CR 603.2c); and a nonlegendary creature can't block them.
func TestMidnightPathlighterVenturesOncePerPlayerAndEvades(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Midnight Pathlighter", effectstest.MidnightPathlighterOracle, "Creature — Human Wizard", 2, 3)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	plain := effectstest.PushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	legend := effectstest.PushDiesCreatureForTest(g, them.ID, "Their Legend", "", "Legendary Creature — Human", 2, 2)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, them.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(plain, a); err == nil {
		t.Errorf("a nonlegendary creature blocked")
	}
	if err := g.DeclareBlocker(legend, a); err != nil {
		t.Errorf("a legendary creature could not block: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	vcVentured(t, g, me.ID)
	if c := effectstest.LatestOptionPickFor(g, me.ID); c != nil {
		t.Fatalf("a second venture for the same player: %q", c.Reason)
	}
}

// --- statics -----------------------------------------------------------

// Cloister Gargoyle is a 0/4 until its controller completes a dungeon,
// then a 3/4 flier.
func TestCloisterGargoyleGrowsOnceADungeonIsCompleted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cloister Gargoyle", OracleID: effectstest.CloisterGargoyleOracle,
		TypeLine: "Artifact Creature — Gargoyle", Power: 0, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	if p := effectstest.EffectivePower(t, g, id); p != 0 || effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, id), "flying") {
		t.Fatalf("before: power %d, abilities %v", p, effectstest.EffectiveAbilities(t, g, id))
	}
	vcCompleteADungeon(t, g, me.ID)
	if p := effectstest.EffectivePower(t, g, id); p != 3 {
		t.Errorf("after: power %d, want 3", p)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, id), "flying") {
		t.Errorf("after: no flying")
	}
}

func TestGloomStalkerHasDoubleStrikeOnceADungeonIsCompleted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gloom Stalker", OracleID: effectstest.GloomStalkerOracle,
		TypeLine: "Creature — Dwarf Ranger", Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, id), "double strike") {
		t.Fatalf("double strike before completing a dungeon")
	}
	vcCompleteADungeon(t, g, me.ID)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, id), "double strike") {
		t.Errorf("no double strike after completing a dungeon")
	}
}

// Nadaar's anthem: other creatures you control, not Nadaar, not an
// opponent's.
func TestNadaarPumpsOtherCreaturesOnceADungeonIsCompleted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	nadaar := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Nadaar, Selfless Paladin", OracleID: effectstest.NadaarOracle,
		TypeLine: "Legendary Creature — Dragon Knight", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Fatalf("before: bear power %d", p)
	}
	vcCompleteADungeon(t, g, me.ID)
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("my bear has power %d, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, nadaar); p != 3 {
		t.Errorf("Nadaar pumped itself: power %d", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 2 {
		t.Errorf("their bear pumped: power %d", p)
	}
}

// Triumphant Adventurer has first strike on its controller's turn only.
func TestTriumphantAdventurerHasFirstStrikeOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Triumphant Adventurer", OracleID: effectstest.TriumphantAdventurerOracle,
		TypeLine: "Creature — Human Knight", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Triumphant Adventurer", OracleID: effectstest.TriumphantAdventurerOracle,
		TypeLine: "Creature — Human Knight", Power: 1, Toughness: 1, Owner: them.ID, Controller: them.ID,
	})
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, mine), "first strike") {
		t.Errorf("no first strike on its controller's turn")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, theirs), "first strike") {
		t.Errorf("first strike on another player's turn")
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, mine), "deathtouch") {
		t.Errorf("no deathtouch")
	}
}

// --- activated --------------------------------------------------------

func TestDungeonMapAndSecretDoorVentureAtSorcerySpeed(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Dungeon Map", "Artifact", effectstest.DungeonMapOracle},
		{"Secret Door", "Artifact Creature — Wall", effectstest.SecretDoorOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.PushDiesCreatureForTest(g, me.ID, tc.name, tc.oracle, tc.typeLine, 0, 4)
			effectstest.ConduitMana(t, g, me, "{U}{C}{C}{C}{C}")
			if err := effectstest.ActivateOn(t, g, me.ID, id, 0, uuid.Nil); err == nil {
				t.Fatalf("activated outside a main phase")
			}
			effectstest.AdvanceToMain(t, g)
			effectstest.ConduitMana(t, g, me, "{U}{C}{C}{C}{C}")
			if err := effectstest.ActivateOn(t, g, me.ID, id, 0, uuid.Nil); err != nil {
				t.Fatalf("activate: %v", err)
			}
			vcVentured(t, g, me.ID)
		})
	}
}

func TestRangersHawkTapsAnotherCreatureToVenture(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hawk := effectstest.PushDiesCreatureForTest(g, me.ID, "Ranger's Hawk", effectstest.RangersHawkOracle, "Creature — Bird", 1, 1)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.ConduitMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, hawk, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.FindBattlefieldCardForTest(g, bear).Tapped || !effectstest.FindBattlefieldCardForTest(g, hawk).Tapped {
		t.Errorf("the Hawk and the Bear were not both tapped")
	}
	vcVentured(t, g, me.ID)
}

func TestDungeonDescentEntersTappedAndTapsALegendToVenture(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushDiesCreatureForTest(g, me.ID, "Dungeon Descent", effectstest.DungeonDescentOracle, "Land", 0, 0)
	plain := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	legend := effectstest.PushDiesCreatureForTest(g, me.ID, "Legend", "", "Legendary Creature — Human", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.ConduitMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{plain}}); err == nil {
		t.Errorf("a nonlegendary creature paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{legend}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	vcVentured(t, g, me.ID)

	// The land enters tapped.
	effectstest.CastCatalogSpell(t, g, "Dungeon Descent", "Land", effectstest.DungeonDescentOracle, nil)
	effectstest.FrSettle(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Zombie Ogre", effectstest.ZombieOgreOracle, "Creature — Zombie Ogre", 3, 5)
	effectstest.AdvanceToEndStep(t, g)
	vcNotVentured(t, g, me.ID)

	g2 := effectstest.NewCatalogGame(t)
	me2, them2 := g2.Seats[0], g2.Seats[1]
	effectstest.PushDiesCreatureForTest(g2, me2.ID, "Zombie Ogre", effectstest.ZombieOgreOracle, "Creature — Zombie Ogre", 3, 5)
	theirs := effectstest.PushVanillaCreature(g2, them2.ID, "Their Bear", 2, 2)
	g2.WithWriteLock(func() {
		if err := g2.DestroyPermanentForEffect(theirs); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.FrSettle(t, g2)
	effectstest.AdvanceToEndStep(t, g2)
	vcVentured(t, g2, me2.ID)
	_ = them
}

func TestWanderingTroubadourVenturesOnlyAfterALandEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Wandering Troubadour", effectstest.WanderingTroubadourOracle, "Creature — Dragon Bard", 4, 2)
	effectstest.AdvanceToEndStep(t, g)
	vcNotVentured(t, g, me.ID)

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.PushDiesCreatureForTest(g2, me2.ID, "Wandering Troubadour", effectstest.WanderingTroubadourOracle, "Creature — Dragon Bard", 4, 2)
	effectstest.CastCatalogSpell(t, g2, "Forest", "Basic Land — Forest", "", nil)
	effectstest.FrSettle(t, g2)
	effectstest.AdvanceToEndStep(t, g2)
	vcVentured(t, g2, me2.ID)
}

// --- dealt damage --------------------------------------------------------

func TestZaltoVenturesWhenDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	zalto := effectstest.PushDiesCreatureForTest(g, me.ID, "Zalto, Fire Giant Duke", effectstest.ZaltoOracle, "Legendary Creature — Giant Barbarian", 7, 3)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, zalto, 1); err != nil {
			t.Fatal(err)
		}
	})
	vcVentured(t, g, me.ID)
}

// --- spells -------------------------------------------------------------

func TestBarTheGateCountersACreatureSpellAndVentures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, effectstest.B13HandCard(me, "Bar the Gate", "Instant", effectstest.BarTheGateOracle),
		game.CastSpellParams{Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Fatalf("Bar the Gate: %v", err)
	}
	vcVentured(t, g, me.ID)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature spell was not countered")
	}
}

func TestFatesReversalReturnsACreatureCardAndVentures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	effectstest.CastCatalogSpell(t, g, "Fates' Reversal", "Sorcery", effectstest.FatesReversalOracle, effectstest.B16TargetCard(dead))
	vcVentured(t, g, me.ID)
	if z := effectstest.InitZone(g, dead); z != game.ZoneHand {
		t.Fatalf("the creature card is in %v, want hand", z)
	}

	// With no target chosen it still ventures.
	g2 := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g2, "Fates' Reversal", "Sorcery", effectstest.FatesReversalOracle, nil)
	vcVentured(t, g2, g2.Seats[0].ID)
}

func TestYouFindACursedIdolStealsItsEyes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastModal(t, g, "You Find a Cursed Idol", "Sorcery", effectstest.CursedIdolOracle, []int{2}, nil)
	vcVentured(t, g, me.ID)
	if n := effectstest.FrTokens(g, me.ID, "Treasure"); n != 1 {
		t.Errorf("Treasures = %d, want 1", n)
	}

	g2 := effectstest.NewCatalogGame(t)
	rock := effectstest.B12Permanent(g2, g2.Seats[1].ID, "Their Rock", "Artifact")
	effectstest.CastModal(t, g2, "You Find a Cursed Idol", "Sorcery", effectstest.CursedIdolOracle, []int{0}, effectstest.B16TargetCard(rock))
	vcNotVentured(t, g2, g2.Seats[0].ID)
	if g2.Battlefield.Contains(rock) {
		t.Error("Smash It did not destroy the artifact")
	}
}

func TestKickInTheDoorCountersHastesBarsWallsAndVentures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	wall := effectstest.PushDiesCreatureForTest(g, them.ID, "Their Wall", "", "Creature — Wall", 0, 4)
	other := effectstest.PushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Kick in the Door", "Sorcery", effectstest.KickInTheDoorOracle, effectstest.B16TargetCard(bear))
	vcVentured(t, g, me.ID)
	if got := effectstest.FindBattlefieldCardForTest(g, bear).Counters["+1/+1"]; got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "haste") {
		t.Errorf("no haste")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, them.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err == nil {
		t.Errorf("a Wall blocked")
	}
	if err := g.DeclareBlocker(other, bear); err != nil {
		t.Errorf("a non-Wall could not block: %v", err)
	}
}
