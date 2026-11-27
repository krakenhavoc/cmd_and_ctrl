package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b574_cards_test.go — Ravnica: City of Guilds cards, S82 OLD batch
// 574 (#5600), played through the real catalog entries.

const (
	o574Cat        = "13d301d1-c1e1-4656-8700-3f1c1c7c4485"
	o574Sandsower  = "1dce1f39-eacb-4d29-899b-55569284c7ed"
	o574Moloch     = "48f78611-13a7-4b5b-a52a-317ecf658ea3"
	o574Informant  = "4ad09b75-6d7a-46c7-95a2-d7f90995fdd1"
	o574Witch      = "6f78d21b-1959-4edc-9f28-d7ada539abef"
	o574Minion     = "80b4a9f7-008d-4979-a3df-7637fae68167"
	o574Mortipede  = "a5122d52-cebc-435b-a077-cc64376d7a86"
	o574IvyDancer  = "f2e081be-87ef-457a-a5cb-2c2ff10cbcf5"
	o574Commando   = "f44a8a72-2c51-429e-8447-7bd8155b6f72"
	o574Corrupter  = "faddaa9d-3f70-49a2-81e6-22b16cf8e9fe"
	o574CreatureTL = "Creature — Bear"
)

func TestS82OldB574CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o574Cat: "Sabertooth Alley Cat", o574Sandsower: "Sandsower", o574Moloch: "Torpid Moloch",
		o574Informant: "Lurking Informant", o574Witch: "Thoughtpicker Witch", o574Minion: "Tidewater Minion",
		o574Mortipede: "Mortipede", o574IvyDancer: "Ivy Dancer", o574Commando: "Ordruun Commando",
		o574Corrupter: "Woodwraith Corrupter",
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

func TestS82OldB574SabertoothAlleyCatMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	cat := effectstest.B12Push(g, me.ID, "Sabertooth Alley Cat", "Creature — Cat", o574Cat, 2, 1)
	effectstest.AdvanceToDeclareAttackersOf(t, g, seat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != cat {
		t.Fatalf("pass with the Cat home = %v, want a refusal naming it", err)
	}
	if err := g.DeclareAttacker(cat, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
}

func TestS82OldB574SabertoothAlleyCatOnlyDefendersMayBlockAfterTheAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	cat := effectstest.B12Push(g, me.ID, "Sabertooth Alley Cat", "Creature — Cat", o574Cat, 2, 1)
	bear := effectstest.B12Creature(g, them.ID, "Their Bear", o574CreatureTL, 2, 2)
	wall := effectstest.TypedCreature(g, them.ID, "Their Wall", "Creature — Wall", "", 0, 4, "defender")
	effectstest.RavB2MustActivate(t, g, me.ID, cat, 0, game.ActivateAbilityParams{}, "C", "R")
	effectstest.AttackIntoBlocks(t, g, them.ID, cat)
	if r := effectstest.BlockRefusal(t, g, bear, cat); r == "" {
		t.Error("a creature without defender blocked the Cat after the ability")
	}
	if r := effectstest.BlockRefusal(t, g, wall, cat); r != "" {
		t.Errorf("a creature with defender may not block the Cat: %q", r)
	}
}

func TestS82OldB574SabertoothAlleyCatCanBeBlockedByAnyoneWithoutTheAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	cat := effectstest.B12Push(g, me.ID, "Sabertooth Alley Cat", "Creature — Cat", o574Cat, 2, 1)
	bear := effectstest.B12Creature(g, them.ID, "Their Bear", o574CreatureTL, 2, 2)
	effectstest.AttackIntoBlocks(t, g, them.ID, cat)
	if r := effectstest.BlockRefusal(t, g, bear, cat); r != "" {
		t.Errorf("without the ability the Cat is unblockable by a Bear: %q", r)
	}
}

func TestS82OldB574SandsowerTapsATargetCreatureForThreeTappedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.B12Push(g, me.ID, "Sandsower", "Creature — Spirit", o574Sandsower, 1, 3)
	a := effectstest.B12Creature(g, me.ID, "A", o574CreatureTL, 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", o574CreatureTL, 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", o574CreatureTL, 2, 2)
	tgt := effectstest.CardRefs(victim)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{src, a}, Targets: tgt, Strict: true}); err == nil {
		t.Fatal("two creatures paid a cost of three")
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{src, a, b}, Targets: tgt, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"the Sandsower": src, "A": a, "B": b, "the victim": victim} {
		if !effectstest.TappedForTest(t, g, id) {
			t.Errorf("%s is untapped", name)
		}
	}
}

func TestS82OldB574TorpidMolochLosesDefenderForThreeLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	moloch := effectstest.B12Push(g, me.ID, "Torpid Moloch", "Creature — Lizard", o574Moloch, 3, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == moloch {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "defender")
			}
		}
	})
	lands := []uuid.UUID{
		effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain"),
		effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain"),
		effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain"),
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, moloch, "defender") {
		t.Fatal("the Moloch has no defender")
	}
	if err := effectstest.RavB2Activate(t, g, me.ID, moloch, 0, game.ActivateAbilityParams{SacrificeIDs: lands[:2]}); err == nil {
		t.Fatal("two lands paid a cost of three")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, moloch, 0, game.ActivateAbilityParams{SacrificeIDs: lands})
	for _, l := range lands {
		if g.Battlefield.Contains(l) {
			t.Error("a land was not sacrificed")
		}
	}
	if effectstest.HasKeywordOnBattlefield(t, g, moloch, "defender") {
		t.Error("the Moloch still has defender")
	}
}

func TestS82OldB574LurkingInformantMayMillTheTopCard(t *testing.T) {
	for _, mill := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		informant := effectstest.B12Push(g, me.ID, "Lurking Informant", "Creature — Human Rogue", o574Informant, 1, 2)
		top := effectstest.SsCard(opp.Library, opp.ID, "Their Top", "Sorcery", 0, 0)
		if err := effectstest.RavB2Activate(t, g, me.ID, informant, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}, "C", "C"); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		if prompt == nil {
			t.Fatal("the controller was not asked")
		}
		var picks []uuid.UUID
		if mill {
			picks = []uuid.UUID{top}
		}
		if err := g.ResolveChooseCards(prompt.ID, me.ID, picks); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := opp.Graveyard.Contains(top); got != mill {
			t.Errorf("mill=%v: the card is in the graveyard = %v", mill, got)
		}
		if opp.Library.Contains(top) == mill {
			t.Errorf("mill=%v: the card is still in the library = %v", mill, opp.Library.Contains(top))
		}
	}
}

func TestS82OldB574ThoughtpickerWitchExilesOneOfTheTopTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	witch := effectstest.B12Push(g, me.ID, "Thoughtpicker Witch", "Creature — Human Wizard", o574Witch, 1, 1)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", o574CreatureTL, 1, 1)
	second := effectstest.SsCard(opp.Library, opp.ID, "Second", "Sorcery", 0, 0)
	top := effectstest.SsCard(opp.Library, opp.ID, "Top", "Sorcery", 0, 0)
	params := game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}
	if err := effectstest.RavB2Activate(t, g, me.ID, witch, 0, params, "C"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if prompt == nil {
		t.Fatal("the controller was not asked which card to exile")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{second}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if opp.Library.Contains(second) {
		t.Error("the chosen card is still in the library")
	}
	if !opp.Library.Contains(top) {
		t.Error("the card that was not chosen left the library")
	}
}

func TestS82OldB574TidewaterMinionLosesDefenderAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	minion := effectstest.B12Push(g, me.ID, "Tidewater Minion", "Creature — Elemental Minion", o574Minion, 4, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == minion {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "defender")
			}
		}
	})
	rock := effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rock {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	effectstest.RavB2MustActivate(t, g, me.ID, minion, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rock)})
	if effectstest.TappedForTest(t, g, rock) {
		t.Error("the target permanent is still tapped")
	}
	if !effectstest.TappedForTest(t, g, minion) {
		t.Error("the untap ability does not tap the Minion")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, minion, 0, game.ActivateAbilityParams{}, "C", "C", "C", "C")
	if effectstest.HasKeywordOnBattlefield(t, g, minion, "defender") {
		t.Error("the Minion still has defender")
	}
}

func TestS82OldB574MortipedeLuresEveryBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mort := effectstest.ReqCreature(g, me.ID, "Mortipede", o574Mortipede, 4, 1, "haste")
	a := effectstest.ReqCreature(g, opp.ID, "Wall A", "", 0, 4)
	b := effectstest.ReqCreature(g, opp.ID, "Wall B", "", 0, 4)
	effectstest.RavB2MustActivate(t, g, me.ID, mort, 0, game.ActivateAbilityParams{}, "C", "C", "G")
	found := false
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModAddBlockRequirement && m.Text == string(game.BlockRequirementLure) {
				found = true
				if e.Duration.Kind != game.UntilEndOfTurn {
					t.Errorf("duration %v, want until end of turn", e.Duration.Kind)
				}
			}
		}
	}
	if !found {
		t.Fatal("Mortipede registered no lure")
	}
	effectstest.ReqToBlockers(t, g, mort)
	effectstest.ReqRefusal(t, g.PassPriority())
	if err := effectstest.ReqBlock(t, g, mort, a, b); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both walls blocking: %v", err)
	}
}

func TestS82OldB574IvyDancerGrantsForestwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dancer := effectstest.B12Push(g, me.ID, "Ivy Dancer", "Creature — Dryad Shaman", o574IvyDancer, 1, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", o574CreatureTL, 2, 2)
	effectstest.RavB2MustActivate(t, g, me.ID, dancer, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "forestwalk") {
		t.Error("the target did not gain forestwalk")
	}
	if !effectstest.TappedForTest(t, g, dancer) {
		t.Error("the ability taps the Dancer")
	}
}

func TestS82OldB574OrdruunCommandoPreventsTheNextDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cmd := effectstest.B12Push(g, me.ID, "Ordruun Commando", "Creature — Minotaur Soldier", o574Commando, 4, 1)
	effectstest.RavB2MustActivate(t, g, me.ID, cmd, 0, game.ActivateAbilityParams{}, "W")
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), cmd, 3) })
	if got := effectstest.DamageMarkedOn(g, cmd); got != 2 {
		t.Errorf("3 damage into a 1-shield: marked %d, want 2", got)
	}
}

func TestS82OldB574WoodwraithCorrupterMakesAForestAFourByFourCreatureForGood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	corr := effectstest.B12Push(g, me.ID, "Woodwraith Corrupter", "Creature — Elemental Horror", o574Corrupter, 3, 6)
	forest := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	island := effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	if err := effectstest.RavB2Activate(t, g, me.ID, corr, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(island)}, "C", "B", "G"); err == nil {
		t.Fatal("the ability targeted an Island")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, corr, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(forest)}, "C", "B", "G")
	c := effectstest.LayeredCard(t, g, forest)
	eff := c.Effective()
	if !c.IsCreature() || !c.IsLand() {
		t.Errorf("the Forest is creature=%v land=%v, want both", c.IsCreature(), c.IsLand())
	}
	if eff.Power != 4 || eff.Toughness != 4 {
		t.Errorf("the Forest is %d/%d, want 4/4", eff.Power, eff.Toughness)
	}
	if len(eff.Colors) != 2 {
		t.Errorf("colors %v, want black and green", eff.Colors)
	}
	if !c.HasSubtype("Elemental") || !c.HasSubtype("Horror") || !c.HasSubtype("Forest") {
		t.Errorf("subtypes %v, want Elemental, Horror and Forest", eff.Subtypes)
	}
	// It lasts past the turn: the effect has no end of its own.
	effectstest.EndTurn(t, g)
	if c := effectstest.LayeredCard(t, g, forest); !c.IsCreature() {
		t.Error("the Forest stopped being a creature when the turn ended")
	}
}
