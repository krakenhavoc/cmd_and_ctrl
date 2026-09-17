package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

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

// blockRefusal declares `blocker` against `attacker` and returns the
// engine's reason, "" for an accepted block. It runs on a clone, so a
// test can ask several times about the same pair.
func blockRefusal(t *testing.T, g *game.Game, blocker, attacker uuid.UUID) game.BlockReason {
	t.Helper()
	err := g.Clone().DeclareBlocker(blocker, attacker)
	if err == nil {
		return ""
	}
	var br *game.BlockRefusedError
	if !errors.As(err, &br) || !errors.Is(err, game.ErrIllegalBlock) {
		t.Fatalf("DeclareBlocker = %v, want nil or a block refusal", err)
	}
	return br.Reason
}

// seatsForCombat returns the active seat and the next one, which the
// tests attack.
func seatsForCombat(g *game.Game) (me, them *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func attackIntoBlocks(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceToStepInTurn(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceToStepInTurn(t, g, game.StepDeclareBlockers)
}

// TestLordOfAtlantisIslandwalkIsEnforced retires the card's caveat:
// the granted islandwalk stops a block while the defender controls an
// Island, and stops mattering the moment the Lord leaves.
func TestLordOfAtlantisIslandwalkIsEnforced(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	merfolk := pushTribalCreature(g, me.ID, "Merfolk Looter", "Creature — Merfolk Rogue", 1, 1)
	lord := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lord of Atlantis",
		TypeLine: "Creature — Merfolk", Power: 2, Toughness: 2,
		OracleID: lordOfAtlantisOracle, Owner: me.ID, Controller: me.ID,
	})
	bear := pushTribalCreature(g, them.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	seedLand(g, them.ID, "Island", "Basic Land — Island", "")

	if !hasAbility(effectiveAbilities(t, g, merfolk), "islandwalk") {
		t.Fatal("the Lord did not grant islandwalk")
	}
	attackIntoBlocks(t, g, them.ID, merfolk)

	if r := blockRefusal(t, g, bear, merfolk); r != game.BlockReasonLandwalk {
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
	if r := blockRefusal(t, g, bear, merfolk); r != "" {
		t.Errorf("the Lord is gone, so islandwalk is too, but the block was refused: %q", r)
	}
}

// TestLandwalkLordsGrantTheirLandwalk — Elvish Champion's forestwalk and
// Goblin King's mountainwalk, which S26 left out. Both lords have no
// "you control" clause, so an opponent's creature of the type gets the
// landwalk too, and "other" keeps it off the lord itself.
func TestLandwalkLordsGrantTheirLandwalk(t *testing.T) {
	for _, tc := range []struct {
		lord, oracle, tribe, keyword, land string
	}{
		{"Elvish Champion", elvishChampionOracle, "Elf", "forestwalk", "Forest"},
		{"Goblin King", goblinKingOracle, "Goblin", "mountainwalk", "Mountain"},
	} {
		t.Run(tc.lord, func(t *testing.T) {
			g := newCatalogGame(t)
			me, them := seatsForCombat(g)
			mine := pushTribalCreature(g, me.ID, "My "+tc.tribe, "Creature — "+tc.tribe, 1, 1)
			theirs := pushTribalCreature(g, them.ID, "Their "+tc.tribe, "Creature — "+tc.tribe, 1, 1)
			lord := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: tc.lord, TypeLine: "Creature — " + tc.tribe,
				Power: 2, Toughness: 2, OracleID: tc.oracle, Owner: me.ID, Controller: me.ID,
			})
			blocker := pushTribalCreature(g, them.ID, "Wall", "Creature — Wall", 0, 4)

			for _, id := range []uuid.UUID{mine, theirs} {
				if !hasAbility(effectiveAbilities(t, g, id), tc.keyword) {
					t.Errorf("%s did not grant %s to every other %s", tc.lord, tc.keyword, tc.tribe)
				}
				if got := effectivePower(t, g, id); got != 2 {
					t.Errorf("%s power = %d, want 2", tc.tribe, got)
				}
			}
			if hasAbility(effectiveAbilities(t, g, lord), tc.keyword) {
				t.Errorf("%s granted %s to itself", tc.lord, tc.keyword)
			}

			attackIntoBlocks(t, g, them.ID, mine)
			if r := blockRefusal(t, g, blocker, mine); r != "" {
				t.Fatalf("no %s under the defender, but the block was refused: %q", tc.land, r)
			}
			seedLand(g, them.ID, tc.land, "Basic Land — "+tc.land, "")
			if r := blockRefusal(t, g, blocker, mine); r != game.BlockReasonLandwalk {
				t.Errorf("the defender controls a %s: reason %q, want landwalk", tc.land, r)
			}
		})
	}
}

// TestMasterOfThePearlTridentIsYoursOnly — the modern lord says "you
// control", so an opponent's Merfolk get neither half.
func TestMasterOfThePearlTridentIsYoursOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	mine := pushTribalCreature(g, me.ID, "Merfolk Looter", "Creature — Merfolk Rogue", 1, 1)
	theirs := pushTribalCreature(g, them.ID, "Lullmage Mentor", "Creature — Merfolk Wizard", 2, 2)
	master := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Master of the Pearl Trident",
		TypeLine: "Creature — Merfolk", Power: 2, Toughness: 2,
		OracleID: masterOfThePearlTridentOracle, Owner: me.ID, Controller: me.ID,
	})
	seedLand(g, them.ID, "Island", "Basic Land — Island", "")

	if got := effectivePower(t, g, mine); got != 2 {
		t.Errorf("own Merfolk power = %d, want 2", got)
	}
	if !hasAbility(effectiveAbilities(t, g, mine), "islandwalk") {
		t.Error("own Merfolk did not gain islandwalk")
	}
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("opponent's Merfolk power = %d, want 2 — the Master says \"you control\"", got)
	}
	if hasAbility(effectiveAbilities(t, g, theirs), "islandwalk") {
		t.Error("opponent's Merfolk gained islandwalk")
	}
	if hasAbility(effectiveAbilities(t, g, master), "islandwalk") || effectivePower(t, g, master) != 2 {
		t.Error("the Master pumped itself — \"other\" excludes it")
	}

	attackIntoBlocks(t, g, them.ID, mine)
	if r := blockRefusal(t, g, theirs, mine); r != game.BlockReasonLandwalk {
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
			g := newCatalogGame(t)
			me, them := seatsForCombat(g)
			selkie := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Cold-Eyed Selkie", TypeLine: "Creature — Merfolk Rogue",
				Power: 3, Toughness: 1, OracleID: coldEyedSelkieOracle, Owner: me.ID, Controller: me.ID,
			})
			bear := pushTribalCreature(g, them.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
			seedLand(g, them.ID, "Island", "Basic Land — Island", "")
			if !hasAbility(effectiveAbilities(t, g, selkie), "islandwalk") {
				t.Fatal("Cold-Eyed Selkie has no islandwalk")
			}

			attackIntoBlocks(t, g, them.ID, selkie)
			if r := blockRefusal(t, g, bear, selkie); r != game.BlockReasonLandwalk {
				t.Fatalf("blocking the Selkie against an Island: reason %q, want landwalk", r)
			}

			handBefore := me.Hand.Size()
			advanceToStepInTurn(t, g, game.StepCombatDamage)
			answerLatestTriggerPrompt(t, g, me.ID, tc.accept)
			passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	carrier := pushRestrictionBear(g, me, "Carrier")
	plain := pushRestrictionBear(g, me, "Plain")
	blocker := pushRestrictionBear(g, them, "Blocker")
	boots := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Trailblazer's Boots", TypeLine: "Artifact — Equipment",
		OracleID: trailblazersBootsOracle, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(boots, game.TargetRef{Kind: game.TargetCard, ID: carrier}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
	if !hasAbility(effectiveAbilities(t, g, carrier), "nonbasic landwalk") {
		t.Fatal("the equipped creature has no nonbasic landwalk")
	}
	if hasAbility(effectiveAbilities(t, g, plain), "nonbasic landwalk") {
		t.Fatal("the Boots granted landwalk to a creature they aren't attached to")
	}
	seedLand(g, them.ID, "Island", "Basic Land — Island", "")

	attackIntoBlocks(t, g, them.ID, carrier, plain)
	if r := blockRefusal(t, g, blocker, carrier); r != "" {
		t.Fatalf("only a basic land under the defender, but the block was refused: %q", r)
	}
	seedLand(g, them.ID, "Command Tower", "Land", "")
	if r := blockRefusal(t, g, blocker, carrier); r != game.BlockReasonLandwalk {
		t.Errorf("the defender controls a nonbasic land: reason %q, want landwalk", r)
	}
	if r := blockRefusal(t, g, blocker, plain); r != "" {
		t.Errorf("the unequipped creature was refused: %q", r)
	}
}

// TestUrborgSwitchesSwampwalkOn — land types are read after layer 4.
// A defender with only a Forest can block a swampwalker until Urborg,
// Tomb of Yawgmoth makes the Forest a Swamp too, whoever controls
// Urborg.
func TestUrborgSwitchesSwampwalkOn(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	walker := pushTribalCreature(g, me.ID, "Bog Wraith", "Creature — Wraith", 3, 3, "swampwalk")
	blocker := pushTribalCreature(g, them.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	seedLand(g, them.ID, "Forest", "Basic Land — Forest", "")

	attackIntoBlocks(t, g, them.ID, walker)
	if r := blockRefusal(t, g, blocker, walker); r != "" {
		t.Fatalf("a plain Forest, but the block was refused: %q", r)
	}
	seedLand(g, me.ID, "Urborg, Tomb of Yawgmoth", "Legendary Land", urborgOracle)
	if r := blockRefusal(t, g, blocker, walker); r != game.BlockReasonLandwalk {
		t.Errorf("Urborg makes the defender's Forest a Swamp: reason %q, want landwalk", r)
	}
}
