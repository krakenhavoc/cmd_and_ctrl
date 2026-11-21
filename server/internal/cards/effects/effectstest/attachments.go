package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attachments_test.go covers the S24 Equipment and Aura batch. The
// engine-side relation is pinned in game/attach_test.go; everything
// here is the catalog half — that equip is an ordinary activated
// ability, that "equipped creature" statics reach the host, and that
// the attachment-conditioned triggers fire.

const (
	BonesplitterOracle   = "452e3f5f-ce17-4682-966b-5cc100210aee"
	SkullclampOracle     = "65986c1b-8e51-4604-b685-d82fa7d1263a"
	GreavesOracle        = "ca204b66-8d0c-431a-8d34-282f7c2d17da"
	BootsOracle          = "c8b143ad-43ec-4e0d-a440-e348daa31391"
	SwordFeastOracle     = "d0901053-6de0-46d0-9ee3-8d40510236c1"
	SwordFireOracle      = "2ccdc60a-49a9-44b9-a7af-0ebf18b26785"
	RancorOracle         = "9d2d6479-531c-4ce1-b52b-00e36fa63b64"
	CurseOpulenceOracle  = "ba0d3df2-3acf-46d7-8d64-8d67d1579adc"
	EquipTypeLine        = "Artifact — Equipment"
	AuraTypeLine         = "Enchantment — Aura"
	CurseAuraTypeLine    = "Enchantment — Aura Curse"
	TestCreatureTypeLine = "Creature — Bear"
)

// advanceToMain walks the turn to a main phase, which is where the
// sorcery-speed equip gate opens (CR 702.6a).
func AdvanceToMain(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("never reached a main phase")
}

// equipTo activates the equipment's equip ability (always the last
// declared ability on these cards) targeting `creature`, then passes
// priority so it resolves.
func EquipTo(t *testing.T, g *game.Game, controller, equipment, creature uuid.UUID) {
	t.Helper()
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: OracleOfBattlefieldCard(t, g, equipment)})
	idx := len(abilities) - 1
	if idx < 0 {
		t.Fatal("equipment has no activated ability")
	}
	if err := g.ActivateCatalogAbility(controller, equipment, idx, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}},
	}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	PassPriorityAroundTable(t, g)
}

func OracleOfBattlefieldCard(t *testing.T, g *game.Game, id uuid.UUID) string {
	t.Helper()
	var out string
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c.OracleID, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

func AttachmentHostOf(t *testing.T, g *game.Game, id uuid.UUID) game.TargetRef {
	t.Helper()
	var out game.TargetRef
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c.AttachedTo, true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", id)
	}
	return out
}

// --- The Swords: attachment-conditioned combat triggers -----------

// dealCombatDamageToPlayer fakes the combat-damage step's event for
// one creature, which is all the Swords' triggers watch. Driving a
// full combat would test the combat engine, not the attachment.
func DealCombatDamageToPlayer(g *game.Game, creature, victim uuid.UUID, amount int) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind:   game.EventDealDamage,
			Source: creature,
			Target: victim,
			Amount: amount,
			Combat: true,
		})
	})
}
