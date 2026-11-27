package testss82m

import (
	"github.com/google/uuid"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b625_test.go — Judgment cards, S82 OLD batch 625 (#5611), played
// through the real catalog entries.

const (
	o625ArcaneTeachings = "0b633017-bfdf-47cd-a60f-c93816b47e9c"
	o625Exoskeletal     = "29358dcd-b26a-4720-9daa-7baa648e3540"
	o625InfectiousRage  = "ed0402d7-02cb-4763-a661-82280d6c9af8"
	o625Cagemail        = "fe1f40fe-011b-4686-b865-fed00c56d30a"
	o625GiantWarthog    = "0eb41c36-5910-41fd-97fc-1cae2332e5c5"
	o625Filth           = "411484ac-a87d-49e8-a7ef-0cd055a872b6"
	o625Eesha           = "b9b9edb2-62f5-499c-b2fc-359dc673c4b4"
	o625BorderPatrol    = "fa5e2575-51bb-448e-9aaf-0e0f5956ac98"
)

func TestS82OldB625CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o625ArcaneTeachings: "Arcane Teachings", o625Exoskeletal: "Exoskeletal Armor", o625InfectiousRage: "Infectious Rage",
		o625Cagemail: "Cagemail", o625GiantWarthog: "Giant Warthog", o625Filth: "Filth",
		o625Eesha: "Commander Eesha", o625BorderPatrol: "Border Patrol",
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

func TestS82OldB625ArcaneTeachingsPumpsAndGrantsAPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Arcane Teachings", o625ArcaneTeachings, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("enchanted Bear = %d/%d, want 4/4", p, tough)
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life = %d, want %d", opp.Life, life-1)
	}
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("the tapped creature activated the granted ability again")
	}
}

func TestS82OldB625ExoskeletalArmorCountsCreatureCardsInEveryGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Exoskeletal Armor", o625Exoskeletal, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Fatalf("with empty graveyards the Bear = %d/%d, want 2/2", p, tough)
	}
	effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Dead Bear", "Creature — Bear")
	effectstest.PushGraveyardCardWithTypeLine(g, opp.ID, "Dead Wolf", "Creature — Wolf")
	effectstest.PushGraveyardCardWithTypeLine(g, opp.ID, "Dead Spell", "Sorcery")
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("with two creature cards in graveyards the Bear = %d/%d, want 4/4", p, tough)
	}
}

func TestS82OldB625CagemailPumpsAndStopsAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Cagemail", o625Cagemail, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("enchanted Bear = %d/%d, want 4/4", p, tough)
	}
	r := effectstest.AuraRestrictions(t, g, bear)
	if r&game.CantAttack == 0 {
		t.Error("the enchanted creature can still attack")
	}
	if r&game.CantBlock != 0 {
		t.Error("the enchanted creature can't block, but Cagemail only stops attacks")
	}
}

func TestS82OldB625InfectiousRageShrinksAndReturnsToAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	other := effectstest.AuraBear(g, g.Seats[1].ID)
	aura := effectstest.AuraCast(t, g, "Infectious Rage", o625InfectiousRage, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 1 {
		t.Fatalf("enchanted Bear = %d/%d, want 4/1", p, tough)
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, aura) == nil {
		t.Fatal("the Aura did not return to the battlefield")
	}
	if !effectstest.AttachedToNow(t, g, aura, other) {
		t.Error("the Aura did not return attached to the only creature left")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, other); p != 4 || tough != 1 {
		t.Errorf("the new host = %d/%d, want 4/1", p, tough)
	}
}

func TestS82OldB625InfectiousRageStaysInTheGraveyardWithNoCreatureToEnchant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	aura := effectstest.AuraCast(t, g, "Infectious Rage", o625InfectiousRage, bear)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, aura) != nil {
		t.Error("the Aura returned with no creature to enchant")
	}
	if !me.Graveyard.Contains(aura) {
		t.Error("the Aura left the graveyard")
	}
}

func TestS82OldB625VanillaKeywordCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	warthog := effectstest.B12Push(g, me.ID, "Giant Warthog", "Creature — Boar Beast", o625GiantWarthog, 5, 5)
	patrol := effectstest.B12Push(g, me.ID, "Border Patrol", "Creature — Human Nomad", o625BorderPatrol, 1, 6)
	eesha := effectstest.B12Push(g, me.ID, "Commander Eesha", "Legendary Creature — Bird Soldier", o625Eesha, 2, 4)
	for _, c := range []struct {
		id uuid.UUID
		kw string
	}{{warthog, "trample"}, {patrol, "vigilance"}, {eesha, "flying"}, {eesha, "protection from creatures"}} {
		if !effectstest.HasKeywordOnBattlefield(t, g, c.id, c.kw) {
			t.Errorf("%v lacks %s", c.id, c.kw)
		}
	}
}

func TestS82OldB625FilthGivesSwampwalkFromTheGraveyardOnlyWithASwamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.GraveyardCardOf(g, me, "Filth", "Creature — Incarnation", "{3}{B}", o625Filth)
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "swampwalk") {
		t.Fatal("swampwalk without a Swamp")
	}
	effectstest.B12Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "swampwalk") {
		t.Error("no swampwalk with Filth in the graveyard and a Swamp")
	}
	filth := effectstest.B12Push(g, me.ID, "Filth", "Creature — Incarnation", o625Filth, 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, filth, "swampwalk") {
		t.Error("the battlefield Filth lacks its own printed swampwalk")
	}
}
