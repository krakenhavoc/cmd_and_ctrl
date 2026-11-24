package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b052_cards_test.go — Strixhaven: School of Mages, batch 52 (#5502).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	oldB52Vortex      = "5383a1c1-662f-46fe-a0e5-0f38efe1a37a"
	oldB52Spellbinder = "8c4d7cbc-9cfc-4f0a-8844-e458ab806056"
	oldB52Titan       = "a52e73bb-47cb-4912-9b81-4ae40641eabb"
	oldB52Masterpiece = "e9f5f87d-a485-4e69-9beb-49a51230bd53"
	oldB52Redcap      = "b14bd895-1d4e-4bd5-9ac1-fed56afb6c90"
)

func TestOldB52AllRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB52Vortex:      "Detention Vortex",
		oldB52Spellbinder: "Elite Spellbinder",
		oldB52Titan:       "Daemogoth Titan",
		oldB52Masterpiece: "Elemental Masterpiece",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func oldB52OnBattlefield(g *game.Game, id uuid.UUID) bool {
	z := g.FindCardZoneForEffect(id)
	return z != nil && z.Kind == game.ZoneBattlefield
}

// --- Detention Vortex ---------------------------------------------------------

func TestOldB52DetentionVortexStopsAttackBlockAndActivationsAndOpponentsMayDestroyIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	var owner *game.Player
	for _, s := range g.Seats {
		if s.ID != active.ID {
			owner = s
			break
		}
	}
	effectstest.AdvanceToMain(t, g)
	host := effectstest.B12Push(g, active.ID, "Weaselback Redcap", "Creature — Goblin Knight", oldB52Redcap, 1, 1)
	aura := effectstest.PushCatalogPermanent(g, owner.ID, "Detention Vortex", "Enchantment — Aura", oldB52Vortex, false)
	effectstest.AttachForTest(t, g, aura, host)
	if err := effectstest.HobB4Activate(t, g, active.ID, host, 0, "{R}{C}", nil); err == nil {
		t.Error("an activated ability of the enchanted permanent was activated")
	}
	active.ManaPool = nil
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(host, owner.ID); err == nil {
		t.Error("the enchanted creature attacked")
	}
	effectstest.AdvanceToMain(t, g)
	// Only the Aura's opponents may destroy it: its controller may not.
	if err := effectstest.HobB4Activate(t, g, owner.ID, aura, 0, "{C}{C}{C}", nil); err == nil {
		t.Error("the Aura's controller activated its destroy ability")
	}
	owner.ManaPool = nil
	if err := effectstest.HobB4Activate(t, g, active.ID, aura, 0, "{C}{C}", nil); err == nil {
		t.Error("destroyed the Aura for {2}")
	}
	active.ManaPool = nil
	if err := effectstest.HobB4Activate(t, g, active.ID, aura, 0, "{C}{C}{C}", nil); err != nil {
		t.Fatalf("an opponent could not pay {3} to destroy the Aura: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB52OnBattlefield(g, aura) {
		t.Error("the Aura is still on the battlefield")
	}
}

func TestOldB52DetentionVortexStopsBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Detention Vortex", "Enchantment — Aura", oldB52Vortex, effectstest.CardTargets(blocker))
	effectstest.PassPriorityAroundTable(t, g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err == nil {
		t.Error("the enchanted creature blocked")
	}
}

func TestOldB52DetentionVortexCannotEnchantALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.SeedLand(g, opp.ID, "Forest", "Basic Land — Forest", "")
	if err := effectstest.CastCatalogSpellErr(t, g, "Detention Vortex", "Enchantment — Aura", oldB52Vortex, effectstest.CardTargets(land)); err == nil {
		t.Error("Detention Vortex enchanted a land")
	}
}

// --- Elite Spellbinder ---------------------------------------------------------

func TestOldB52EliteSpellbinderExilesANonlandCardThatItsOwnerMayPlayForTwoMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Hand.Cards = nil
	land := effectstest.SocB17Hand(opp, "Forest", "Basic Land — Forest", "", "")
	spell := effectstest.SocB17Hand(opp, "Test Instant", "Instant", "{U}", "")
	effectstest.CastCatalogSpell(t, g, "Elite Spellbinder", "Creature — Human Cleric", oldB52Spellbinder, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceRevealedHandPick, me.ID)
	if pick == nil {
		t.Fatal("no pick among the opponent's nonland cards")
	}
	for _, id := range pick.DiscardOptions {
		if id == land {
			t.Error("a land card was offered")
		}
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("take the card: %v", err)
	}
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOfCard(g, spell); z != game.ZoneExile {
		t.Fatalf("the chosen card is in %v, want exile", z)
	}
	if !opp.Hand.Contains(land) {
		t.Error("the land left the hand")
	}
	// Its owner may play it for as long as it remains exiled, for {2} more.
	effectstest.FloatForTest(g, opp, "U")
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("cast the exiled card for its printed cost")
	}
	opp.ManaPool = nil
	effectstest.FloatForTest(g, opp, "UUU")
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Errorf("could not cast the exiled card for its cost plus {2}: %v", err)
	}
}

func TestOldB52EliteSpellbinderMayExileNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Hand.Cards = nil
	spell := effectstest.SocB17Hand(opp, "Test Instant", "Instant", "{U}", "")
	effectstest.CastCatalogSpell(t, g, "Elite Spellbinder", "Creature — Human Cleric", oldB52Spellbinder, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceRevealedHandPick, me.ID)
	if pick == nil || !pick.PickOptional {
		t.Fatalf("the pick should be optional: %+v", pick)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(spell) {
		t.Error("the card left the hand though none was chosen")
	}
}

// --- Daemogoth Titan -----------------------------------------------------------

func TestOldB52DaemogothTitanSacrificesACreatureWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := effectstest.B12Push(g, me.ID, "Daemogoth Titan", "Creature — Demon", oldB52Titan, 11, 10)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(titan, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no sacrifice prompt")
	}
	if len(pick.ChooseCards) != 2 {
		t.Errorf("%d candidates, want 2 (the Titan may sacrifice itself)", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOfCard(g, bear); z != game.ZoneGraveyard {
		t.Errorf("the sacrificed Bear is in %v, want the graveyard", z)
	}
	if !oldB52OnBattlefield(g, titan) {
		t.Error("the Titan was sacrificed instead")
	}
}

func TestOldB52DaemogothTitanSacrificesItselfWhenAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := effectstest.B12Push(g, me.ID, "Daemogoth Titan", "Creature — Demon", oldB52Titan, 11, 10)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(titan, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestChooseCards(g, me.ID); pick != nil {
		if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{titan}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if oldB52OnBattlefield(g, titan) {
		t.Error("a lone Titan survived its own trigger")
	}
}

func TestOldB52DaemogothTitanSacrificesWhenItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	titan := effectstest.B12Push(g, opp.ID, "Daemogoth Titan", "Creature — Demon", oldB52Titan, 11, 10)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(titan, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, opp.ID)
	if pick == nil {
		t.Fatal("no sacrifice prompt after blocking")
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB52OnBattlefield(g, bear) {
		t.Error("the Bear was not sacrificed")
	}
}

// --- Elemental Masterpiece -----------------------------------------------------

func TestOldB52ElementalMasterpieceMakesTwoElementals(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Elemental Masterpiece", "Sorcery", oldB52Masterpiece, nil)
	effectstest.PassPriorityAroundTable(t, g)
	elementals := effectstest.SocB17Names(g, me.ID, "Elemental")
	if len(elementals) != 2 {
		t.Fatalf("%d Elementals, want 2", len(elementals))
	}
	for _, e := range elementals {
		if p, tough := effectstest.EffectivePower(t, g, e.InstanceID), effectstest.EffectiveToughness(t, g, e.InstanceID); p != 4 || tough != 4 {
			t.Errorf("an Elemental is %d/%d, want 4/4", p, tough)
		}
	}
}

func TestOldB52ElementalMasterpieceChannelsFromHandForATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Elemental Masterpiece", "Sorcery", oldB52Masterpiece)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{R}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.SocB17Names(g, me.ID, "Treasure")); n != 1 {
		t.Errorf("%d Treasures, want 1", n)
	}
	if !effectstest.InGraveyard(me, id) {
		t.Error("the discarded card is not in the graveyard")
	}
	if len(effectstest.SocB17Names(g, me.ID, "Elemental")) != 0 {
		t.Error("channeling cast the spell")
	}
}
