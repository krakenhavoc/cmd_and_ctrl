package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// landwalk_cards_test.go — #705's first wave (ADR 0045 addendum, owner
// question 3 option (b)): the three lords whose landwalk was inert or
// omitted, the two cards that were waiting on landwalk alone, and the
// Urborg interaction that proves land types are read after layer 4.
// The engine half (every basic type, nonbasic, two landwalks, the
// defending player of a planeswalker or battle attack) is in
// game/landwalk_test.go.

const (
	masterOfThePearlTridentOracle = "9f6ea9fe-eee5-4dca-a34e-c66d3e218716"
	coldEyedSelkieOracle          = "f33cd975-ab70-4209-a6d8-e0d727772bf2"
	trailblazersBootsOracle       = "634d5009-cbf3-44cb-8c15-7057f501a210"
	goblinKingOracle              = "d236b3fc-0d3f-4d99-875d-e32a33fe5767"
)

// TestLordOfAtlantisIslandwalkIsEnforced retires the card's caveat:
// the granted islandwalk stops a block while the defender controls an
// Island, and stops mattering the moment the Lord leaves.
func TestLordOfAtlantisIslandwalkIsEnforced(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	merfolk := effectstest.PushTribalCreature(g, me.ID, "Merfolk Looter", "Creature — Merfolk Rogue", 1, 1)
	lord := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lord of Atlantis",
		TypeLine: "Creature — Merfolk", Power: 2, Toughness: 2,
		OracleID: effectstest.LordOfAtlantisOracle, Owner: me.ID, Controller: me.ID,
	})
	bear := effectstest.PushTribalCreature(g, them.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.SeedLand(g, them.ID, "Island", "Basic Land — Island", "")

	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, merfolk), "islandwalk") {
		t.Fatal("the Lord did not grant islandwalk")
	}
	effectstest.AttackIntoBlocks(t, g, them.ID, merfolk)

	if r := effectstest.BlockRefusal(t, g, bear, merfolk); r != game.BlockReasonLandwalk {
		t.Fatalf("blocking a granted islandwalker against an Island: reason %q, want landwalk", r)
	}
	if g.SeatOwesBlockDecision(them.ID) {
		t.Error("the defender's only creature can't block, yet the #328 signal says a block decision is owed")
	}

	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(lord); err != nil {
			t.Fatalf("exile the Lord: %v", err)
		}
	})
	// #1501: the defender had no legal block as the step began, so
	// their declaration completed then — "declared, none" — and a block
	// now is late, whatever the Lord's leaving changed.
	if r := effectstest.BlockRefusal(t, g, bear, merfolk); r != game.BlockReasonBlocksDeclared {
		t.Errorf("a block after the declaration completed: reason %q, want blocks_declared", r)
	}
	// What the Lord's leaving changes is the PAIR: islandwalk no longer
	// refuses it.
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		atk, _ := g.LookupCardForEffect(merfolk)
		blk, _ := g.LookupCardForEffect(bear)
		if r := g.BlockPairRefusalLocked(&atk, &blk); !r.Legal() {
			t.Errorf("the Lord is gone, so islandwalk is too, but the pair is refused: %q", r.Reason)
		}
	})
}

// TestLandwalkLordsGrantTheirLandwalk — Elvish Champion's forestwalk and
// Goblin King's mountainwalk, which S26 left out. Both lords have no
// "you control" clause, so an opponent's creature of the type gets the
// landwalk too, and "other" keeps it off the lord itself.
func TestLandwalkLordsGrantTheirLandwalk(t *testing.T) {
	for _, tc := range []struct {
		lord, oracle, tribe, keyword, land string
	}{
		{"Elvish Champion", effectstest.ElvishChampionOracle, "Elf", "forestwalk", "Forest"},
		{"Goblin King", goblinKingOracle, "Goblin", "mountainwalk", "Mountain"},
	} {
		t.Run(tc.lord, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, them := effectstest.SeatsForCombat(g)
			mine := effectstest.PushTribalCreature(g, me.ID, "My "+tc.tribe, "Creature — "+tc.tribe, 1, 1)
			theirs := effectstest.PushTribalCreature(g, them.ID, "Their "+tc.tribe, "Creature — "+tc.tribe, 1, 1)
			lord := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: tc.lord, TypeLine: "Creature — " + tc.tribe,
				Power: 2, Toughness: 2, OracleID: tc.oracle, Owner: me.ID, Controller: me.ID,
			})
			blocker := effectstest.PushTribalCreature(g, them.ID, "Wall", "Creature — Wall", 0, 4)

			for _, id := range []uuid.UUID{mine, theirs} {
				if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, id), tc.keyword) {
					t.Errorf("%s did not grant %s to every other %s", tc.lord, tc.keyword, tc.tribe)
				}
				if got := effectstest.EffectivePower(t, g, id); got != 2 {
					t.Errorf("%s power = %d, want 2", tc.tribe, got)
				}
			}
			if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, lord), tc.keyword) {
				t.Errorf("%s granted %s to itself", tc.lord, tc.keyword)
			}

			effectstest.AttackIntoBlocks(t, g, them.ID, mine)
			if r := effectstest.BlockRefusal(t, g, blocker, mine); r != "" {
				t.Fatalf("no %s under the defender, but the block was refused: %q", tc.land, r)
			}
			effectstest.SeedLand(g, them.ID, tc.land, "Basic Land — "+tc.land, "")
			if r := effectstest.BlockRefusal(t, g, blocker, mine); r != game.BlockReasonLandwalk {
				t.Errorf("the defender controls a %s: reason %q, want landwalk", tc.land, r)
			}
		})
	}
}

// TestMasterOfThePearlTridentIsYoursOnly — the modern lord says "you
// control", so an opponent's Merfolk get neither half.
func TestMasterOfThePearlTridentIsYoursOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	mine := effectstest.PushTribalCreature(g, me.ID, "Merfolk Looter", "Creature — Merfolk Rogue", 1, 1)
	theirs := effectstest.PushTribalCreature(g, them.ID, "Lullmage Mentor", "Creature — Merfolk Wizard", 2, 2)
	master := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Master of the Pearl Trident",
		TypeLine: "Creature — Merfolk", Power: 2, Toughness: 2,
		OracleID: masterOfThePearlTridentOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.SeedLand(g, them.ID, "Island", "Basic Land — Island", "")

	if got := effectstest.EffectivePower(t, g, mine); got != 2 {
		t.Errorf("own Merfolk power = %d, want 2", got)
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, mine), "islandwalk") {
		t.Error("own Merfolk did not gain islandwalk")
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("opponent's Merfolk power = %d, want 2 — the Master says \"you control\"", got)
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, theirs), "islandwalk") {
		t.Error("opponent's Merfolk gained islandwalk")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, master), "islandwalk") || effectstest.EffectivePower(t, g, master) != 2 {
		t.Error("the Master pumped itself — \"other\" excludes it")
	}

	effectstest.AttackIntoBlocks(t, g, them.ID, mine)
	if r := effectstest.BlockRefusal(t, g, theirs, mine); r != game.BlockReasonLandwalk {
		t.Errorf("blocking an islandwalking Merfolk against an Island: reason %q, want landwalk", r)
	}
}

// TestColdEyedSelkie — printed islandwalk, and "you may draw that many
// cards" for the combat damage it deals to a player: one prompt, the
// whole amount on yes, nothing on no.
func TestColdEyedSelkie(t *testing.T) {
	for _, tc := range []struct {
		name   string
		accept bool
		drawn  int
	}{{"accepted", true, 3}, {"declined", false, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, them := effectstest.SeatsForCombat(g)
			selkie := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Cold-Eyed Selkie", TypeLine: "Creature — Merfolk Rogue",
				Power: 3, Toughness: 1, OracleID: coldEyedSelkieOracle, Owner: me.ID, Controller: me.ID,
			})
			bear := effectstest.PushTribalCreature(g, them.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
			effectstest.SeedLand(g, them.ID, "Island", "Basic Land — Island", "")
			if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, selkie), "islandwalk") {
				t.Fatal("Cold-Eyed Selkie has no islandwalk")
			}

			effectstest.AttackIntoBlocks(t, g, them.ID, selkie)
			if r := effectstest.BlockRefusal(t, g, bear, selkie); r != game.BlockReasonLandwalk {
				t.Fatalf("blocking the Selkie against an Island: reason %q, want landwalk", r)
			}

			handBefore := me.Hand.Size()
			effectstest.AdvanceToStepInTurn(t, g, game.StepCombatDamage)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, tc.accept)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() - handBefore; got != tc.drawn {
				t.Errorf("drew %d cards, want %d", got, tc.drawn)
			}
		})
	}
}

// TestTrailblazersBootsGrantNonbasicLandwalk — the Boots' landwalk
// names a supertype: a basic Island leaves the carrier blockable, a
// nonbasic land does not, and unequipped the creature is ordinary.
func TestTrailblazersBootsGrantNonbasicLandwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	carrier := effectstest.PushRestrictionBear(g, me, "Carrier")
	plain := effectstest.PushRestrictionBear(g, me, "Plain")
	blocker := effectstest.PushRestrictionBear(g, them, "Blocker")
	boots := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Trailblazer's Boots", TypeLine: "Artifact — Equipment",
		OracleID: trailblazersBootsOracle, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(boots, game.TargetRef{Kind: game.TargetCard, ID: carrier}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, carrier), "nonbasic landwalk") {
		t.Fatal("the equipped creature has no nonbasic landwalk")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, plain), "nonbasic landwalk") {
		t.Fatal("the Boots granted landwalk to a creature they aren't attached to")
	}
	effectstest.SeedLand(g, them.ID, "Island", "Basic Land — Island", "")

	effectstest.AttackIntoBlocks(t, g, them.ID, carrier, plain)
	if r := effectstest.BlockRefusal(t, g, blocker, carrier); r != "" {
		t.Fatalf("only a basic land under the defender, but the block was refused: %q", r)
	}
	effectstest.SeedLand(g, them.ID, "Command Tower", "Land", "")
	if r := effectstest.BlockRefusal(t, g, blocker, carrier); r != game.BlockReasonLandwalk {
		t.Errorf("the defender controls a nonbasic land: reason %q, want landwalk", r)
	}
	if r := effectstest.BlockRefusal(t, g, blocker, plain); r != "" {
		t.Errorf("the unequipped creature was refused: %q", r)
	}
}

// TestUrborgSwitchesSwampwalkOn — land types are read after layer 4.
// A defender with only a Forest can block a swampwalker until Urborg,
// Tomb of Yawgmoth makes the Forest a Swamp too, whoever controls
// Urborg.
func TestUrborgSwitchesSwampwalkOn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	walker := effectstest.PushTribalCreature(g, me.ID, "Bog Wraith", "Creature — Wraith", 3, 3, "swampwalk")
	blocker := effectstest.PushTribalCreature(g, them.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.SeedLand(g, them.ID, "Forest", "Basic Land — Forest", "")

	effectstest.AttackIntoBlocks(t, g, them.ID, walker)
	if r := effectstest.BlockRefusal(t, g, blocker, walker); r != "" {
		t.Fatalf("a plain Forest, but the block was refused: %q", r)
	}
	effectstest.SeedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", effectstest.UrborgOracle)
	if r := effectstest.BlockRefusal(t, g, blocker, walker); r != game.BlockReasonLandwalk {
		t.Errorf("Urborg makes the defender's Forest a Swamp: reason %q, want landwalk", r)
	}
}
