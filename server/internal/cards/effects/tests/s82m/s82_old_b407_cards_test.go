package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b407_cards_test.go — Gatecrash, batch 407 (S82 older sets, #5565):
// fourteen Auras. Every test plays the real catalog entry through the engine.

const (
	oldB407DyingWish   = "302407c8-0842-4f8e-a5aa-5c11a69851c4"
	oldB407Visitation  = "3d95e8b1-0ddb-4bd3-b4cc-de55b6f8c403"
	oldB407Madcap      = "4a9dc9ef-a603-467a-aeed-32cc111394b0"
	oldB407Market      = "5102b7c8-6886-4004-9992-416d22b4fce7"
	oldB407Authority   = "538fa2e0-d225-4232-b243-38868b695e26"
	oldB407Pulpit      = "541c6564-ec39-4b14-8c02-dd6a5d122ed3"
	oldB407Skygames    = "5eeb1495-4f0c-4920-8cd6-9c9257327e33"
	oldB407Mantle      = "7ed8f5bc-022b-4b1b-8592-57694bf59a0b"
	oldB407Agoraphobia = "b268fe15-2860-4730-865e-eed7f308fc45"
	oldB407Thief       = "b43ba23b-ff92-4f52-a4f4-2d0825abb7ed"
	oldB407Lashes      = "d3b259e9-f9c7-4830-91fa-6e3d9f4cc1e3"
	oldB407Ward        = "daff56a6-a5b6-4726-8fee-676b846e0c6b"
	oldB407Gift        = "dcb6d317-5c55-4973-8e3d-2e98211dc30e"
	oldB407Ransom      = "f4ae18b5-2cf1-4720-8237-40a0c7b040ae"

	oldB407Bear = "Creature — Bear"
)

var oldB407Cards = map[string]string{
	oldB407DyingWish: "Dying Wish", oldB407Visitation: "Primal Visitation", oldB407Madcap: "Madcap Skills",
	oldB407Market: "Tin Street Market", oldB407Authority: "Alpha Authority", oldB407Pulpit: "Debtor's Pulpit",
	oldB407Skygames: "Skygames", oldB407Mantle: "Holy Mantle", oldB407Agoraphobia: "Agoraphobia",
	oldB407Thief: "Way of the Thief", oldB407Lashes: "One Thousand Lashes", oldB407Ward: "Guildscorn Ward",
	oldB407Gift: "Gift of Orzhova", oldB407Ransom: "Soul Ransom",
}

func TestOldB407EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB407Cards {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB407Enchanted casts the Aura on a fresh 2/2 Bear of the active seat and
// returns the Bear.
func oldB407Enchanted(t *testing.T, g *game.Game, name, oracle string) uuid.UUID {
	t.Helper()
	bear := effectstest.B12Creature(g, g.Seats[g.Turn.ActiveSeat].ID, "Bear", oldB407Bear, 2, 2)
	effectstest.AuraCast(t, g, name, oracle, bear)
	return bear
}

func TestOldB407StaticAurasPumpAndGrantWhileAttached(t *testing.T) {
	cases := []struct {
		name, oracle string
		power, tough int
		keywords     []string
	}{
		{"Primal Visitation", oldB407Visitation, 5, 5, []string{"haste"}},
		{"Madcap Skills", oldB407Madcap, 5, 2, []string{"menace"}},
		{"Holy Mantle", oldB407Mantle, 4, 4, []string{"protection from creatures"}},
		{"Guildscorn Ward", oldB407Ward, 2, 2, []string{"protection from multicolored"}},
		{"Gift of Orzhova", oldB407Gift, 3, 3, []string{"flying", "lifelink"}},
		{"Alpha Authority", oldB407Authority, 2, 2, []string{"hexproof"}},
		{"Agoraphobia", oldB407Agoraphobia, -3, 2, nil},
		{"Way of the Thief", oldB407Thief, 4, 4, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			bear := oldB407Enchanted(t, g, tc.name, tc.oracle)
			if got := effectstest.EffectivePower(t, g, bear); got != tc.power {
				t.Errorf("power = %d, want %d", got, tc.power)
			}
			if got := effectstest.EffectiveToughness(t, g, bear); got != tc.tough {
				t.Errorf("toughness = %d, want %d", got, tc.tough)
			}
			for _, kw := range tc.keywords {
				if !effectstest.HasEffectiveKeyword(t, g, bear, kw) {
					t.Errorf("the enchanted creature has no %q", kw)
				}
			}
		})
	}
}

func TestOldB407AlphaAuthorityLetsOnlyOneCreatureBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := oldB407Enchanted(t, g, "Alpha Authority", oldB407Authority)
	first := effectstest.B12Creature(g, opp.ID, "First", oldB407Bear, 2, 2)
	second := effectstest.B12Creature(g, opp.ID, "Second", oldB407Bear, 2, 2)
	_ = me
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(first, bear); err != nil {
		t.Fatalf("the first blocker: %v", err)
	}
	if err := g.DeclareBlocker(second, bear); err == nil {
		t.Error("a second creature blocked an attacker that can't be blocked by more than one")
	}
}

func TestOldB407WayOfTheThiefIsUnblockableOnlyWhileYouControlAGate(t *testing.T) {
	for _, withGate := range []bool{true, false} {
		name := "without a Gate"
		if withGate {
			name = "with a Gate"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			bear := oldB407Enchanted(t, g, "Way of the Thief", oldB407Thief)
			blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB407Bear, 2, 2)
			if withGate {
				effectstest.B12Push(g, me.ID, "Gate", "Land — Gate", "", 0, 0)
			}
			effectstest.DeclareAttack(t, g, opp.ID, bear)
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			err := g.DeclareBlocker(blocker, bear)
			if withGate && err == nil {
				t.Error("the enchanted creature was blocked while its Aura's controller controlled a Gate")
			}
			if !withGate && err != nil {
				t.Errorf("without a Gate the creature should be blockable: %v", err)
			}
		})
	}
}

func TestOldB407AgoraphobiaReturnsToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB407Bear, 2, 2)
	aura := effectstest.AuraCast(t, g, "Agoraphobia", oldB407Agoraphobia, bear)
	effectstest.FloatMana(t, g, me, "{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("the Aura did not return to its owner's hand")
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("power = %d after the Aura left, want 2", got)
	}
}

func TestOldB407OneThousandLashesLocksTheCreatureAndDrainsItsControllerEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB407Bear, 2, 2)
	effectstest.AuraCast(t, g, "One Thousand Lashes", oldB407Lashes, theirs)
	if got := effectstest.RestrictionsOf(t, g, theirs); got&(game.CantAttackOrBlock|game.CantActivate) != game.CantAttackOrBlock|game.CantActivate {
		t.Errorf("restrictions = %v, want can't attack, block or activate", got.Names())
	}
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-1 {
		t.Errorf("the enchanted creature's controller lost %d life, want 1", life-got)
	}
	if got := effectstest.LifeOf(g, g.Seats[0].ID); got != 40 {
		t.Errorf("the Aura's controller's life changed to %d", got)
	}
}

func TestOldB407DyingWishDrainsForTheDeadCreaturesPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.AuraCast(t, g, "Dying Wish", oldB407DyingWish, bear)
	life, mine := effectstest.LifeOf(g, opp.ID), effectstest.LifeOf(g, me.ID)
	effectstest.P1Destroy(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life-3 {
		t.Errorf("target player's life %d → %d, want -3", life, got)
	}
	if got := effectstest.LifeOf(g, me.ID); got != mine+3 {
		t.Errorf("my life %d → %d, want +3", mine, got)
	}
}

func TestOldB407DyingWishRefusesACreatureYouDoNotControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB407Bear, 2, 2)
	err := effectstest.CastCatalogSpellErr(t, g, "Dying Wish", effectstest.AuraTypeLine, oldB407DyingWish, effectstest.CardTarget(theirs))
	if err == nil {
		t.Error("Dying Wish was cast on a creature its caster does not control")
	}
}

func oldB407Land(t *testing.T, g *game.Game, name, oracle string) (land, aura uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	land = effectstest.B12Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
	effectstest.ToMainForCost(t, g)
	aura = effectstest.AuraCast(t, g, name, oracle, land)
	return land, aura
}

func TestOldB407TinStreetMarketLootsFromTheEnchantedLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land, _ := oldB407Land(t, g, "Tin Street Market", oldB407Market)
	effectstest.SeedLibrary(me, "Top", "Next")
	discard := effectstest.HandCardFull(me, "Junk", "Sorcery", "", "", nil)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand = %d, want %d (discarded one, drew one)", got, hand)
	}
	if me.Hand.Contains(discard) {
		t.Error("the discarded card is still in hand")
	}
	if !effectstest.TappedNow(g, land) {
		t.Error("the land was not tapped")
	}
}

func TestOldB407DebtorsPulpitTapsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land, _ := oldB407Land(t, g, "Debtor's Pulpit", oldB407Pulpit)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB407Bear, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, bear) {
		t.Error("the target creature was not tapped")
	}
}

func TestOldB407SkygamesGrantsFlyingAtSorcerySpeedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land, _ := oldB407Land(t, g, "Skygames", oldB407Skygames)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB407Bear, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate in my main phase: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target did not gain flying")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	other := effectstest.B12Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)
	_ = other
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("the granted ability was activated outside a main phase")
	}
}

func TestOldB407SoulRansomStealsAndOpponentsMayBuyItBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB407Bear, 2, 2)
	aura := effectstest.AuraCast(t, g, "Soul Ransom", oldB407Ransom, theirs)
	if c, ok := g.LookupCardForEffect(theirs); !ok || c.Controller != me.ID {
		t.Fatal("the Aura's controller does not control the enchanted creature")
	}
	// The Aura's controller can't activate it.
	d1 := effectstest.HandCardFull(me, "Junk A", "Sorcery", "", "", nil)
	d2 := effectstest.HandCardFull(me, "Junk B", "Sorcery", "", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{d1, d2}}); err == nil {
		t.Fatal("the Aura's controller activated an opponents-only ability")
	}
	e1 := effectstest.HandCardFull(opp, "Junk C", "Sorcery", "", "", nil)
	e2 := effectstest.HandCardFull(opp, "Junk D", "Sorcery", "", "", nil)
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	oppHand := opp.Hand.Size()
	if err := g.ActivateCatalogAbility(opp.ID, aura, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{e1, e2}}); err != nil {
		t.Fatalf("an opponent could not activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("the Aura's controller's hand %d → %d, want +2", hand, got)
	}
	if got := opp.Hand.Size(); got != oppHand-2 {
		t.Errorf("the activator's hand %d → %d, want -2", oppHand, got)
	}
	if c, ok := g.LookupCardForEffect(theirs); !ok || c.Controller != opp.ID {
		t.Error("the creature did not go back to its owner")
	}
}
