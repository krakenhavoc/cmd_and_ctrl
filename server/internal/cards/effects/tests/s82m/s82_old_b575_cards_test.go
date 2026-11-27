package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b575_cards_test.go — Ravnica: City of Guilds cards, S82 OLD batch
// 575 (#5600), played through the real catalog entries.

const (
	o575Crown      = "316c265b-67e2-48b3-aac8-07a898dbb20f"
	o575FireFiend  = "61f84dbb-e9bb-45b1-a30e-c631be6ad25c"
	o575RootKin    = "99d0876f-174b-4cfe-aa52-f3d5434f6cba"
	o575Groodion   = "b027a80b-30cf-4e7b-8902-efd584e0edf7"
	o575Seeds      = "dd3d4d96-1e88-49e1-8276-7f62c22ad91a"
	o575Agrus      = "e758aecd-6ce1-42f6-9716-32871e3bfa41"
	o575Ursapine   = "fcf92abb-6c06-42d2-a3d1-f44ec997d56d"
	o575CreatureTL = "Creature — Bear"
)

func TestS82OldB575CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o575Crown: "Crown of Convergence", o575FireFiend: "Goblin Fire Fiend", o575RootKin: "Root-Kin Ally",
		o575Groodion: "Drooling Groodion", o575Seeds: "Seeds of Strength", o575Agrus: "Agrus Kos, Wojek Veteran",
		o575Ursapine: "Ursapine",
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

// o575Colored sets the printed colors of a card already placed in a zone.
func o575Colored(g *game.Game, z *game.Zone, id uuid.UUID, colors ...string) {
	g.WithWriteLock(func() {
		for i := range z.Cards {
			if z.Cards[i].InstanceID == id {
				z.Cards[i].Colors = colors
			}
		}
	})
}

func TestS82OldB575CrownOfConvergencePumpsCreaturesSharingAColorWithTheTopCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spec, _ := effects.Lookup(o575Crown)
	if spec.LibraryTopVisible != game.LibraryTopRevealed {
		t.Error("the top card of the library is not played revealed")
	}
	red := effectstest.B12Creature(g, me.ID, "Red Bear", o575CreatureTL, 2, 2)
	green := effectstest.B12Creature(g, me.ID, "Green Bear", o575CreatureTL, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Red Bear", o575CreatureTL, 2, 2)
	o575Colored(g, g.Battlefield, red, "R")
	o575Colored(g, g.Battlefield, green, "G")
	o575Colored(g, g.Battlefield, theirs, "R")
	effectstest.B12Push(g, me.ID, "Crown of Convergence", "Artifact", o575Crown, 0, 0)
	top := effectstest.SsCard(me.Library, me.ID, "Top Spell", "Sorcery", 0, 0)
	o575Colored(g, me.Library, top, "R")
	if got := effectstest.EffectivePower(t, g, red); got != 2 {
		t.Errorf("a noncreature top card pumped: %d, want 2", got)
	}
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top Goblin", TypeLine: "Creature — Goblin",
			Colors: []string{"R"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	if got := effectstest.EffectivePower(t, g, red); got != 3 {
		t.Errorf("my red creature is %d, want 3 under a red creature card", got)
	}
	if got := effectstest.EffectivePower(t, g, green); got != 2 {
		t.Errorf("my green creature is %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 {
		t.Errorf("their red creature is %d, want 2 (it is not mine)", got)
	}
}

func TestS82OldB575CrownOfConvergencePutsTheTopCardOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	crown := effectstest.B12Push(g, me.ID, "Crown of Convergence", "Artifact", o575Crown, 0, 0)
	under := effectstest.SsCard(me.Library, me.ID, "Under", "Sorcery", 0, 0)
	top := effectstest.SsCard(me.Library, me.ID, "Top", "Sorcery", 0, 0)
	effectstest.RavB2MustActivate(t, g, me.ID, crown, 0, game.ActivateAbilityParams{}, "G", "W")
	var gotTop uuid.UUID
	g.ReadSnapshot(func() {
		first, _ := me.Library.Top()
		gotTop = first.InstanceID
	})
	if gotTop == top {
		t.Error("the top card is still on top")
	}
	if gotTop != under {
		t.Errorf("the new top card is not the one under the old one")
	}
	if !me.Library.Contains(top) {
		t.Error("the card left the library")
	}
}

func TestS82OldB575GoblinFireFiendHasHasteMustBeBlockedAndFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	fiend := effectstest.B12Push(g, me.ID, "Goblin Fire Fiend", "Creature — Goblin Berserker", o575FireFiend, 1, 1)
	if !effectstest.HasKeywordOnBattlefield(t, g, fiend, "haste") {
		t.Error("the Fiend has no haste")
	}
	wall := effectstest.ReqCreature(g, opp.ID, "Wall", "", 0, 4)
	effectstest.RavB2MustActivate(t, g, me.ID, fiend, 0, game.ActivateAbilityParams{}, "R")
	effectstest.RavB2MustActivate(t, g, me.ID, fiend, 0, game.ActivateAbilityParams{}, "R")
	if got := effectstest.EffectivePower(t, g, fiend); got != 3 {
		t.Errorf("after two activations the Fiend has power %d, want 3", got)
	}
	effectstest.ReqToBlockers(t, g, fiend)
	effectstest.ReqRefusal(t, g.PassPriority())
	if err := effectstest.ReqBlock(t, g, fiend, wall); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Fiend blocked: %v", err)
	}
}

func TestS82OldB575RootKinAllyConvokesAndPumpsForTwoTappedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spec, _ := effects.Lookup(o575RootKin)
	if spec.TapCost == nil {
		t.Error("Root-Kin Ally has no convoke")
	}
	ally := effectstest.B12Push(g, me.ID, "Root-Kin Ally", "Creature — Elemental Warrior", o575RootKin, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "A", o575CreatureTL, 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", o575CreatureTL, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, ally, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a}, Strict: true}); err == nil {
		t.Fatal("one creature paid a cost of two")
	}
	if err := g.ActivateCatalogAbility(me.ID, ally, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, ally), effectstest.EffectiveToughness(t, g, ally); p != 5 || tough != 5 {
		t.Errorf("the Ally is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.TappedForTest(t, g, a) || !effectstest.TappedForTest(t, g, b) {
		t.Error("the two creatures were not tapped")
	}
	if effectstest.TappedForTest(t, g, ally) {
		t.Error("the Ally tapped though it was not one of the two")
	}
}

func TestS82OldB575DroolingGroodionPumpsOneAndShrinksAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	groodion := effectstest.B12Push(g, me.ID, "Drooling Groodion", "Creature — Beast", o575Groodion, 4, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", o575CreatureTL, 1, 1)
	friend := effectstest.B12Creature(g, me.ID, "Friend", o575CreatureTL, 2, 2)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", o575CreatureTL, 3, 3)
	same := effectstest.RavB2Activate(t, g, me.ID, groodion, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.CardRefs(friend, friend)}, "C", "C", "B", "G")
	if same == nil {
		t.Fatal("the same creature was chosen for both clauses")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, groodion, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.CardRefs(friend, foe)}, "C", "C", "B", "G")
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if p := effectstest.EffectivePower(t, g, friend); p != 4 {
		t.Errorf("my creature has power %d, want 4", p)
	}
	if p, tough := effectstest.EffectivePower(t, g, foe), effectstest.EffectiveToughness(t, g, foe); p != 1 || tough != 1 {
		t.Errorf("their creature is %d/%d, want 1/1", p, tough)
	}
}

func TestS82OldB575SeedsOfStrengthGivesThreePlusOnesToThreeTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", o575CreatureTL, 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", o575CreatureTL, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Seeds of Strength", "Instant", o575Seeds, effectstest.CardRefs(a, a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, a), effectstest.EffectiveToughness(t, g, a); p != 4 || tough != 4 {
		t.Errorf("the creature named twice is %d/%d, want 4/4", p, tough)
	}
	if p, tough := effectstest.EffectivePower(t, g, b), effectstest.EffectiveToughness(t, g, b); p != 3 || tough != 3 {
		t.Errorf("the creature named once is %d/%d, want 3/3", p, tough)
	}
}

func TestS82OldB575AgrusKosPumpsAttackingRedAndWhiteCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	agrus := effectstest.B12Push(g, me.ID, "Agrus Kos, Wojek Veteran", "Legendary Creature — Human Soldier", o575Agrus, 3, 3)
	red := effectstest.B12Creature(g, me.ID, "Red Bear", o575CreatureTL, 2, 2)
	white := effectstest.B12Creature(g, me.ID, "White Bear", o575CreatureTL, 2, 2)
	green := effectstest.B12Creature(g, me.ID, "Green Bear", o575CreatureTL, 2, 2)
	homeRed := effectstest.B12Creature(g, me.ID, "Home Red Bear", o575CreatureTL, 2, 2)
	for id, colors := range map[uuid.UUID][]string{agrus: {"R", "W"}, red: {"R"}, white: {"W"}, green: {"G"}, homeRed: {"R"}} {
		o575Colored(g, g.Battlefield, id, colors...)
	}
	effectstest.DeclareAttack(t, g, opp.ID, agrus, red, white, green)
	effectstest.PassPriorityAroundTable(t, g)
	check := func(name string, id uuid.UUID, wp, wt int) {
		t.Helper()
		if p, tough := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != wp || tough != wt {
			t.Errorf("%s is %d/%d, want %d/%d", name, p, tough, wp, wt)
		}
	}
	check("Agrus Kos (red and white)", agrus, 5, 5)
	check("the attacking red creature", red, 4, 2)
	check("the attacking white creature", white, 2, 4)
	check("the attacking green creature", green, 2, 2)
	check("the red creature at home", homeRed, 2, 2)
}

func TestS82OldB575UrsapinePumpsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ursa := effectstest.B12Push(g, me.ID, "Ursapine", "Creature — Beast", o575Ursapine, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", o575CreatureTL, 2, 2)
	effectstest.RavB2MustActivate(t, g, me.ID, ursa, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}, "G")
	if p, tough := effectstest.EffectivePower(t, g, theirs), effectstest.EffectiveToughness(t, g, theirs); p != 3 || tough != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tough)
	}
}
