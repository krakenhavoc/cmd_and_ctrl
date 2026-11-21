package testsr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestPacifismStopsAttacksAndBlocksAndLetsGoWhenDestroyed is the
// acceptance card. The second half matters as much as the first: the
// restriction is a continuous effect read off the attachment, so
// killing the Aura has to give the creature back with no bookkeeping.
func TestPacifismStopsAttacksAndBlocksAndLetsGoWhenDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushRestrictionBear(g, me, "Grizzly Bears")

	aura := effectstest.CastCatalogSpell(t, g, "Pacifism", "Enchantment — Aura", effectstest.PacifismOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AssertRestrictions(t, g, bear, game.CantAttackOrBlock)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(aura); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	effectstest.AssertRestrictions(t, g, bear, 0)
}

// TestArrestStopsEveryActivationIncludingMana — Arrest's third
// clause has no carve-out, which is exactly what separates it from
// Faith's Fetters.
func TestArrestStopsEveryActivationIncludingMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	feeder := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Carrion Feeder",
		TypeLine:   "Creature — Zombie",
		Power:      1,
		Toughness:  1,
		OracleID:   effectstest.CarrionFeederOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})

	effectstest.CastCatalogSpell(t, g, "Arrest", "Enchantment — Aura", effectstest.ArrestOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: feeder}})
	effectstest.PassPriorityAroundTable(t, g)

	// Carrion Feeder prints its own "can't block", so the set is the
	// union of the two sources — which is the whole reason the field
	// is OR-accumulated rather than assigned.
	effectstest.AssertRestrictions(t, g, feeder,
		game.CantAttackOrBlock|game.CantActivate|game.CantActivateMana)

	if err := g.ActivateCatalogAbility(me.ID, feeder, 0, game.ActivateAbilityParams{}); err != game.ErrCantActivate {
		t.Errorf("an Arrested sac outlet activated: %v", err)
	}
}

// TestFaithsFettersGainsLifeAndSparesManaAbilities covers the three
// clauses that make Fetters different from Arrest: the life, the
// "enchant permanent" reach, and the mana carve-out.
func TestFaithsFettersGainsLifeAndSparesManaAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// A LAND, to prove "enchant permanent" is not "enchant
	// creature" — and Rogue's Passage carries one ability of each
	// kind, which is what the carve-out is about.
	passage := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Rogue's Passage",
		TypeLine:   "Land",
		OracleID:   effectstest.RoguesPassageOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	startLife := me.Life

	effectstest.CastCatalogSpell(t, g, "Faith's Fetters", "Enchantment — Aura", effectstest.FaithsFettersOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: passage}})
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Life - startLife; got != 4 {
		t.Errorf("life gain = %d, want 4", got)
	}
	effectstest.AssertRestrictions(t, g, passage, game.CantAttackOrBlock|game.CantActivate)

	if err := g.ActivateManaAbility(me.ID, passage, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf(`Fetters says "unless they're mana abilities", but the mana ability was refused: %v`, err)
	}
}

// TestWhispersilkCloakMakesTheCarrierUnblockable is the caveat this
// PR retires. The Cloak's two halves are checked together because
// they are the card: shroud in the targeting gate, unblockable in
// the block gate.
func TestWhispersilkCloakMakesTheCarrierUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	carrier := effectstest.PushRestrictionBear(g, me, "Carrier")
	blocker := effectstest.PushRestrictionBear(g, them, "Blocker")
	cloak := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Whispersilk Cloak",
		TypeLine:   "Artifact — Equipment",
		OracleID:   effectstest.WhispersilkOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(cloak, game.TargetRef{Kind: game.TargetCard, ID: carrier}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})

	effectstest.AssertRestrictions(t, g, carrier, game.CantBeBlocked)
	if !effectstest.HasEffectiveKeyword(t, g, carrier, "shroud") {
		t.Error("the Cloak stopped granting shroud")
	}

	effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(carrier, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, carrier); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("a cloaked attacker was blocked: %v", err)
	}
}

// TestCarrionFeederCannotBlock is the other retired caveat, and the
// only card here whose restriction is printed on itself.
func TestCarrionFeederCannotBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := effectstest.PushRestrictionBear(g, me, "Attacker")
	feeder := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Carrion Feeder",
		TypeLine:   "Creature — Zombie",
		Power:      1,
		Toughness:  1,
		OracleID:   effectstest.CarrionFeederOracle,
		Owner:      them.ID,
		Controller: them.ID,
	})

	effectstest.AssertRestrictions(t, g, feeder, game.CantBlock)

	effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(feeder, attacker); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("the Feeder blocked: %v", err)
	}
	// It still attacks — "can't block" is one bit, not both.
	if game.Restricted(restrictionCardByID(t, g, feeder), game.CantAttack) {
		t.Error("Carrion Feeder may attack; only blocking is restricted")
	}
}

// TestRoguesPassageUnblockableLastsOnlyTheTurn is the turn-scoped
// half of the vocabulary: the same bit, with a duration instead of
// an attachment, and gone after cleanup (CR 514.2).
func TestRoguesPassageUnblockableLastsOnlyTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushRestrictionBear(g, me, "Grizzly Bears")
	passage := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Rogue's Passage",
		TypeLine:   "Land",
		OracleID:   effectstest.RoguesPassageOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	for i := 0; i < 4; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(),
			Name:       "Forest",
			TypeLine:   "Basic Land — Forest",
			Owner:      me.ID,
			Controller: me.ID,
		})
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)

	if err := g.ActivateCatalogAbility(me.ID, passage, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		AutoTap: true,
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AssertRestrictions(t, g, bear, game.CantBeBlocked)

	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AssertRestrictions(t, g, bear, 0)
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("ScopedEffects = %d after cleanup, want 0", n)
	}
}

// restrictionCardByID returns a pointer to a battlefield card for a
// read that has already forced a recompute.
func restrictionCardByID(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	var out *game.Card
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				out = &g.Battlefield.Cards[i]
				return
			}
		}
	})
	if out == nil {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return out
}
