package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 362 (Conspiracy Auras, #5557): Guardian Zendikon, Necromantic
// Thirst, Traveler's Cloak, Flowstone Blade, Stasis Cell, Unhallowed Pact and
// Power of Fire.

const (
	oldB362Zendikon = "44154004-e50e-45d1-bde0-ba6e3354067c"
	oldB362Thirst   = "494895db-3da1-4a19-b7a5-ea47817f4e4f"
	oldB362Cloak    = "5241182d-e2ca-4b31-809f-fbff2822dbf6"
	oldB362Blade    = "77a77622-dc5b-4ff2-a98b-732d44c8856e"
	oldB362Cell     = "8b2ba6f8-1060-4167-9941-3ab7bd4e0ca8"
	oldB362Pact     = "da1dd850-9118-42b9-8846-e6bd5042775c"
	oldB362Fire     = "df4903cb-50fa-4620-8470-6a6a398fc1ef"
)

func TestOldB362EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB362Zendikon: "Guardian Zendikon", oldB362Thirst: "Necromantic Thirst", oldB362Cloak: "Traveler's Cloak",
		oldB362Blade: "Flowstone Blade", oldB362Cell: "Stasis Cell", oldB362Pact: "Unhallowed Pact", oldB362Fire: "Power of Fire",
	} {
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

func oldB362Bear(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Creature(g, owner, "My Bear", "Creature — Bear", 2, 2)
}

func oldB362Cast(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := effectstest.CastCatalogSpell(t, g, name, "Enchantment — Aura", oracle, effectstest.CardTarget(host))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatalf("%s did not resolve", name)
	}
	return aura
}

func TestOldB362GuardianZendikonMakesALandAWhiteWallAndReturnsItWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	oldB362Cast(t, g, "Guardian Zendikon", oldB362Zendikon, land)
	c := effectstest.B12Card(t, g, land)
	if !c.IsLand() || !c.IsCreature() || !c.HasSubtype("Wall") {
		t.Fatalf("the enchanted land is %q, want a land Wall creature", c.TypeLine)
	}
	if p, tt := effectstest.PtOf(t, g, land); p != 2 || tt != 6 {
		t.Errorf("the land is %d/%d, want 2/6", p, tt)
	}
	if got := effectstest.EffectiveColors(t, g, land); len(got) != 1 || got[0] != "W" {
		t.Errorf("colours %v, want white", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, land, "defender") {
		t.Error("the Wall lacks defender")
	}
	effectstest.KillCreature(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, land) {
		t.Error("the land was not returned to its owner's hand")
	}
}

func TestOldB362NecromanticThirstReturnsACreatureCardOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Ons2Opponent(g)
	bear := oldB362Bear(g, me.ID)
	oldB362Cast(t, g, "Necromantic Thirst", oldB362Thirst, bear)
	dead := effectstest.PushGraveyardCardTyped(me, "Fallen Bear", "Creature — Bear")
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID)
	if pick == nil {
		t.Fatal("no target pick for the trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: dead}); err != nil {
		t.Fatalf("pick the creature card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, dead) {
		t.Error("the creature card was not returned to hand")
	}
}

func TestOldB362TravelersCloakDrawsAndGrantsLandwalkOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AuraSeedLibrary(me, 3)
	bear := oldB362Bear(g, me.ID)
	hand := len(me.Hand.Cards)
	effectstest.CastCatalogSpell(t, g, "Traveler's Cloak", "Enchantment — Aura", oldB362Cloak, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	// Gate is the eighth of CR 205.3i's land types (index 3).
	effectstest.AnswerOptionPick(t, g, me.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, len(me.Hand.Cards))
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "gatewalk") {
		t.Error("the enchanted creature has no gatewalk")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "islandwalk") {
		t.Error("the enchanted creature also has islandwalk")
	}
}

func TestOldB362FlowstoneBladePumpsTheEnchantedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := oldB362Bear(g, me.ID)
	aura := oldB362Cast(t, g, "Flowstone Blade", oldB362Blade, bear)
	effectstest.FillPoolColored(me, "R", 1)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 1 {
		t.Errorf("bear is %d/%d, want 3/1", p, tt)
	}
	// An empty pool can't pay.
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no mana")
	}
}

func TestOldB362StasisCellLocksTheCreatureAndMovesAtInstantSpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Ons2Opponent(g)
	mine := oldB362Bear(g, me.ID)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	aura := oldB362Cast(t, g, "Stasis Cell", oldB362Cell, mine)
	if h := effectstest.AttachmentHostOf(t, g, aura); h.ID != mine {
		t.Fatalf("attached to %v, want my bear", h.ID)
	}
	effectstest.FillPoolColored(me, "U", 4)
	if err := effectstest.ActivateOn(t, g, me.ID, aura, 0, theirs); err != nil {
		t.Fatalf("move the Aura: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if h := effectstest.AttachmentHostOf(t, g, aura); h.ID != theirs {
		t.Errorf("attached to %v after the move, want their bear", h.ID)
	}
}

func TestOldB362UnhallowedPactReturnsTheCreatureUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Ons2Opponent(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oldB362Cast(t, g, "Unhallowed Pact", oldB362Pact, theirs)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Their Bear" {
			n++
			if c.Controller != me.ID {
				t.Errorf("the Bear returned under %s, want the Aura's controller", c.Controller)
			}
		}
	}
	if n != 1 {
		t.Errorf("%d Bears on the battlefield, want 1", n)
	}
}

func TestOldB362PowerOfFireGivesTheCreatureAPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Ons2Opponent(g)
	bear := oldB362Bear(g, me.ID)
	oldB362Cast(t, g, "Power of Fire", oldB362Fire, bear)
	// Granted abilities follow the creature (ADR 0093), so it can tap to ping.
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("ping: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, opp.Life)
	}
}

func TestOldB362GuardianZendikonEnchantsOnlyALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := oldB362Bear(g, me.ID)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Guardian Zendikon", "Enchantment — Aura", oldB362Zendikon, effectstest.CardTarget(bear)); err == nil {
		t.Error("Guardian Zendikon was cast on a creature that is not a land")
	}
}
