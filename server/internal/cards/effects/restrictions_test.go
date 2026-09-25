package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// restrictions_test.go — the S24 restriction cards, driven through a
// real cast or a real activation. The vocabulary itself is pinned in
// server/internal/game/restrictions_test.go and the "the bot is never
// offered a move the engine refuses" invariant in
// server/internal/legal/restrictions_test.go; these are the cards.

const (
	pacifismOracle      = "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330"
	arrestOracle        = "81728b98-8cf9-4734-a318-69184bb4d15c"
	faithsFettersOracle = "2b2d76f5-4c9b-49dc-b202-68095e2d9b29"
	roguesPassageOracle = "f29dc596-2121-4421-8463-15f6c2e8b9b3"
)

// pushRestrictionBear is a 2/2 on the battlefield under `owner`,
// stamped with a battlefield timestamp so the layer engine sorts it.
func pushRestrictionBear(g *game.Game, owner *game.Player, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
}

// restrictionsOf reads a battlefield card's post-layer restriction
// set, forcing a recompute first — the same read every engine gate
// makes.
func restrictionsOf(t *testing.T, g *game.Game, id uuid.UUID) game.Restriction {
	t.Helper()
	var out game.Restriction
	found := false
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				out = game.RestrictionsOn(&g.Battlefield.Cards[i])
				found = true
				return
			}
		}
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return out
}

func assertRestrictions(t *testing.T, g *game.Game, id uuid.UUID, want game.Restriction) {
	t.Helper()
	if got := restrictionsOf(t, g, id); got != want {
		t.Errorf("restrictions = %v, want %v", got.Names(), want.Names())
	}
}

// TestPacifismStopsAttacksAndBlocksAndLetsGoWhenDestroyed is the
// acceptance card. The second half matters as much as the first: the
// restriction is a continuous effect read off the attachment, so
// killing the Aura has to give the creature back with no bookkeeping.
func TestPacifismStopsAttacksAndBlocksAndLetsGoWhenDestroyed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushRestrictionBear(g, me, "Grizzly Bears")

	aura := castCatalogSpell(t, g, "Pacifism", "Enchantment — Aura", pacifismOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	assertRestrictions(t, g, bear, game.CantAttackOrBlock)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(aura); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	assertRestrictions(t, g, bear, 0)
}

// TestArrestStopsEveryActivationIncludingMana — Arrest's third
// clause has no carve-out, which is exactly what separates it from
// Faith's Fetters.
func TestArrestStopsEveryActivationIncludingMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	feeder := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Carrion Feeder",
		TypeLine:   "Creature — Zombie",
		Power:      1,
		Toughness:  1,
		OracleID:   carrionFeederOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})

	castCatalogSpell(t, g, "Arrest", "Enchantment — Aura", arrestOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: feeder}})
	passPriorityAroundTable(t, g)

	// Carrion Feeder prints its own "can't block", so the set is the
	// union of the two sources — which is the whole reason the field
	// is OR-accumulated rather than assigned.
	assertRestrictions(t, g, feeder,
		game.CantAttackOrBlock|game.CantActivate|game.CantActivateMana)

	if err := g.ActivateCatalogAbility(me.ID, feeder, 0, game.ActivateAbilityParams{}); err != game.ErrCantActivate {
		t.Errorf("an Arrested sac outlet activated: %v", err)
	}
}

// TestFaithsFettersGainsLifeAndSparesManaAbilities covers the three
// clauses that make Fetters different from Arrest: the life, the
// "enchant permanent" reach, and the mana carve-out.
func TestFaithsFettersGainsLifeAndSparesManaAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// A LAND, to prove "enchant permanent" is not "enchant
	// creature" — and Rogue's Passage carries one ability of each
	// kind, which is what the carve-out is about.
	passage := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Rogue's Passage",
		TypeLine:   "Land",
		OracleID:   roguesPassageOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	startLife := me.Life

	castCatalogSpell(t, g, "Faith's Fetters", "Enchantment — Aura", faithsFettersOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: passage}})
	passPriorityAroundTable(t, g)

	if got := me.Life - startLife; got != 4 {
		t.Errorf("life gain = %d, want 4", got)
	}
	assertRestrictions(t, g, passage, game.CantAttackOrBlock|game.CantActivate)

	if err := g.ActivateManaAbility(me.ID, passage, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf(`Fetters says "unless they're mana abilities", but the mana ability was refused: %v`, err)
	}
}

// TestWhispersilkCloakMakesTheCarrierUnblockable is the caveat this
// PR retires. The Cloak's two halves are checked together because
// they are the card: shroud in the targeting gate, unblockable in
// the block gate.
func TestWhispersilkCloakMakesTheCarrierUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	carrier := pushRestrictionBear(g, me, "Carrier")
	blocker := pushRestrictionBear(g, them, "Blocker")
	cloak := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Whispersilk Cloak",
		TypeLine:   "Artifact — Equipment",
		OracleID:   whispersilkOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(cloak, game.TargetRef{Kind: game.TargetCard, ID: carrier}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})

	assertRestrictions(t, g, carrier, game.CantBeBlocked)
	if !hasEffectiveKeyword(t, g, carrier, "shroud") {
		t.Error("the Cloak stopped granting shroud")
	}

	advanceToStepInTurn(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(carrier, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceToStepInTurn(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, carrier); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("a cloaked attacker was blocked: %v", err)
	}
}

// TestCarrionFeederCannotBlock is the other retired caveat, and the
// only card here whose restriction is printed on itself.
func TestCarrionFeederCannotBlock(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := pushRestrictionBear(g, me, "Attacker")
	feeder := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Carrion Feeder",
		TypeLine:   "Creature — Zombie",
		Power:      1,
		Toughness:  1,
		OracleID:   carrionFeederOracle,
		Owner:      them.ID,
		Controller: them.ID,
	})

	assertRestrictions(t, g, feeder, game.CantBlock)

	advanceToStepInTurn(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceToStepInTurn(t, g, game.StepDeclareBlockers)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushRestrictionBear(g, me, "Grizzly Bears")
	passage := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Rogue's Passage",
		TypeLine:   "Land",
		OracleID:   roguesPassageOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	for i := 0; i < 4; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(),
			Name:       "Forest",
			TypeLine:   "Basic Land — Forest",
			Owner:      me.ID,
			Controller: me.ID,
		})
	}
	advanceToStepInTurn(t, g, game.StepPrecombatMain)

	if err := g.ActivateCatalogAbility(me.ID, passage, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		AutoTap: true,
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)

	assertRestrictions(t, g, bear, game.CantBeBlocked)

	advanceToNextSeatsTurn(t, g)
	assertRestrictions(t, g, bear, 0)
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

// advanceToStepInTurn walks the cursor forward to `step` within the
// current turn.
func advanceToStepInTurn(t *testing.T, g *game.Game, step game.Step) {
	t.Helper()
	for i := 0; i < 30; i++ {
		if g.Turn.Step == step {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to %v: %v", step, err)
		}
	}
	t.Fatalf("never reached %v (at %v)", step, g.Turn.Step)
}
