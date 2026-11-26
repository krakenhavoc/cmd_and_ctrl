package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b534_cards_test.go — Shadowmoor: Auras, batch 534 (#5591, S82 OLD).
// Every test plays the real catalog entry through the engine.

const (
	oldB534Fists   = "0e6f9fc3-5cc7-42cc-ae83-9be875a1579f"
	oldB534Tether  = "1ea20ecb-1919-4be7-b0bb-d79d2dec98fd"
	oldB534Torture = "389b6d77-47ab-43a3-99d1-dcc762939cc4"
	oldB534Fever   = "3a018a54-6094-42bb-af6a-ef436c09f0b4"
	oldB534Helm    = "66fa528f-0e3a-450f-9035-1ba5a8a921e6"
	oldB534Runes   = "674b9469-3213-4cb0-be7e-db209bc09f92"
	oldB534Sinking = "8a5e925f-320d-4bb7-9374-0384e7ac41e7"
	oldB534Steel   = "90d2e092-6eb8-4d13-bdad-19e7325a8b36"
	oldB534Dust    = "da1ca3ec-8bef-4b7e-896e-4d50e8cffbe5"
)

func TestOldB534CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB534Fists:   "Fists of the Demigod",
		oldB534Tether:  "Biting Tether",
		oldB534Torture: "Torture",
		oldB534Fever:   "Bloodshed Fever",
		oldB534Helm:    "Helm of the Ghastlord",
		oldB534Runes:   "Runes of the Deus",
		oldB534Sinking: "Sinking Feeling",
		oldB534Steel:   "Steel of the Godhead",
		oldB534Dust:    "Torpor Dust",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

// oldB534Attach puts the real Aura on the battlefield attached to host.
func oldB534Attach(g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	aura := effectstest.PushCatalogPermanent(g, owner, name, effectstest.AuraTypeLine, oracle, false)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, aura).AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	})
	return aura
}

// oldB534Cast casts the Aura from hand at host, with its mana floated.
func oldB534Cast(t *testing.T, g *game.Game, name, oracle, cost, mana string, colors []string, host uuid.UUID) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.HandCardFull(me, name, effectstest.AuraTypeLine, cost, oracle, colors)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, mana)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(host)})
}

func oldB534Has(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	return effectstest.EffectiveAbilitiesContain(t, g, id, kw)
}

// oldB534ColourAura runs the shared check of the four colour-conditional Auras:
// each colour of the host switches its own half on, a host of both colours gets
// both, and a host of neither gets nothing.
func oldB534ColourAura(t *testing.T, name, oracle string, c1, k1 string, c2, k2 string) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		label  string
		colors []string
		bonus  int
		kws    map[string]bool
	}{
		{"first colour only", []string{c1}, 1, map[string]bool{k1: true, k2: false}},
		{"second colour only", []string{c2}, 1, map[string]bool{k1: false, k2: true}},
		{"both colours", []string{c1, c2}, 2, map[string]bool{k1: true, k2: true}},
		{"neither colour", nil, 0, map[string]bool{k1: false, k2: false}},
	} {
		host := effectstest.B43Creature(g, me.ID, tc.label, "Creature — Bear", 2, 2, tc.colors...)
		oldB534Attach(g, me.ID, name, oracle, host)
		if p, tough := effectstest.PtOf(t, g, host); p != 2+tc.bonus || tough != 2+tc.bonus {
			t.Errorf("%s: %s: host is %d/%d, want %d/%d", name, tc.label, p, tough, 2+tc.bonus, 2+tc.bonus)
		}
		for kw, want := range tc.kws {
			if got := oldB534Has(t, g, host, kw); got != want {
				t.Errorf("%s: %s: has %q = %v, want %v", name, tc.label, kw, got, want)
			}
		}
	}
}

func TestOldB534FistsOfTheDemigodGivesWitherToBlackAndFirstStrikeToRed(t *testing.T) {
	oldB534ColourAura(t, "Fists of the Demigod", oldB534Fists, "B", "wither", "R", "first strike")
}

func TestOldB534RunesOfTheDeusGivesDoubleStrikeToRedAndTrampleToGreen(t *testing.T) {
	oldB534ColourAura(t, "Runes of the Deus", oldB534Runes, "R", "double strike", "G", "trample")
}

func TestOldB534SteelOfTheGodheadGivesLifelinkToWhiteAndUnblockabilityToBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	white := effectstest.B43Creature(g, me.ID, "White Bear", "Creature — Bear", 2, 2, "W")
	blue := effectstest.B43Creature(g, me.ID, "Blue Bear", "Creature — Bear", 2, 2, "U")
	oldB534Attach(g, me.ID, "Steel of the Godhead", oldB534Steel, white)
	oldB534Attach(g, me.ID, "Steel of the Godhead", oldB534Steel, blue)
	for id, want := range map[uuid.UUID]int{white: 3, blue: 3} {
		if p, tough := effectstest.PtOf(t, g, id); p != want || tough != want {
			t.Errorf("host is %d/%d, want %d/%d", p, tough, want, want)
		}
	}
	if !oldB534Has(t, g, white, "lifelink") || oldB534Has(t, g, blue, "lifelink") {
		t.Error("lifelink should be on the white host only")
	}
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{white, blue} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.BrRefusal(t, g.DeclareBlocker(blocker, blue), game.BlockReasonCantBeBlocked)
	if err := g.DeclareBlocker(blocker, white); err != nil {
		t.Errorf("the white creature could not be blocked: %v", err)
	}
}

func TestOldB534HelmOfTheGhastlordDrawsForBlueAndDiscardsForBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blue := effectstest.B43Creature(g, me.ID, "Blue Bear", "Creature — Bear", 2, 2, "U")
	black := effectstest.B43Creature(g, me.ID, "Black Bear", "Creature — Bear", 2, 2, "B")
	oldB534Attach(g, me.ID, "Helm of the Ghastlord", oldB534Helm, blue)
	oldB534Attach(g, me.ID, "Helm of the Ghastlord", oldB534Helm, black)
	for _, id := range []uuid.UUID{blue, black} {
		if p, tough := effectstest.PtOf(t, g, id); p != 3 || tough != 3 {
			t.Errorf("host is %d/%d, want 3/3", p, tough)
		}
	}
	effectstest.SeedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	effectstest.DealNoncombatDamage(g, me.ID, blue, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: the blue creature's damage draws a card", me.Hand.Size(), hand+1)
	}
	effectstest.SeedLibrary(opp, "X", "Y")
	for opp.Hand.Size() < 3 {
		effectstest.HandCardFull(opp, "Spare", "Sorcery", "{1}", "", nil)
	}
	oppHand := opp.Hand.Size()
	effectstest.DealNoncombatDamage(g, me.ID, black, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != oppHand-1 {
		t.Errorf("opponent's hand %d, want %d: the black creature's damage makes them discard", opp.Hand.Size(), oppHand-1)
	}
}

func TestOldB534BitingTetherStealsTheCreatureAndShrinksItEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := oldB534Cast(t, g, "Biting Tether", oldB534Tether, "{4}{U}", "CCCCU", []string{"U"}, host); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, host) != me.ID {
		t.Fatal("you do not control the enchanted creature")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, host, game.CounterMinusOne); got != 1 {
		t.Errorf("%d -1/-1 counters after your upkeep, want 1", got)
	}
}

func TestOldB534TortureActivatesToPutAMinusOneCounterOnTheEnchantedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	aura := oldB534Attach(g, me.ID, "Torture", oldB534Torture, host)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CB")
	effectstest.P7Activate(t, g, me, aura, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, host, game.CounterMinusOne); got != 1 {
		t.Errorf("%d -1/-1 counters, want 1", got)
	}
	if p, tough := effectstest.PtOf(t, g, host); p != 2 || tough != 2 {
		t.Errorf("the host is %d/%d, want 2/2", p, tough)
	}
}

func TestOldB534BloodshedFeverForcesTheCreatureToAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	oldB534Attach(g, me.ID, "Bloodshed Fever", oldB534Fever, host)
	if len(effectstest.GaLayered(t, g, host).Effective().AttackRequirements) == 0 {
		t.Error("the enchanted creature does not have to attack each combat")
	}
}

func TestOldB534SinkingFeelingKeepsTheCreatureTappedUnlessItPaysACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	oldB534Attach(g, me.ID, "Sinking Feeling", oldB534Sinking, host)
	effectstest.TapForTest(t, g, host)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.IsTapped(g, host) {
		t.Fatal("the enchanted creature untapped during its controller's untap step")
	}
	effectstest.FloatForTest(g, me, "C")
	effectstest.P7Activate(t, g, me, host, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.IsTapped(g, host) {
		t.Error("the creature did not untap")
	}
	if got := effectstest.CountersOn(g, host, game.CounterMinusOne); got != 1 {
		t.Errorf("%d -1/-1 counters, want 1", got)
	}
}

func TestOldB534TorporDustCanBeCastAtInstantSpeedAndTakesThreePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	// Cast in the upkeep, outside a main phase and with the stack empty: only flash allows it.
	id := effectstest.HandCardFull(me, "Torpor Dust", effectstest.AuraTypeLine, "{2}{U/B}", oldB534Dust, []string{"U", "B"})
	effectstest.FloatForTest(g, me, "CCU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(host)}); err != nil {
		t.Fatalf("cast in the upkeep: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, host); p != 1 || tough != 4 {
		t.Errorf("the host is %d/%d, want 1/4", p, tough)
	}
}
