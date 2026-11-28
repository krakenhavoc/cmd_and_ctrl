package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b791_cards_test.go — Visions, S82 OLD batch 791 (#5631). Every test
// plays the real catalog entry through the engine.

const (
	oldB791Brood     = "7a22d694-29d1-4e7b-8fca-d7008aede489"
	oldB791Wand      = "9d32da71-60e5-4eb3-97ca-ae76e2f5df57"
	oldB791Djinn     = "050dac46-9ba0-4b8a-b61b-1c7ec6f3723a"
	oldB791Champion  = "9352e742-9a22-4463-b41f-3117ff1540ed"
	oldB791Assassin  = "db43b866-13e0-4336-a580-b64fe1ee5954"
	oldB791Breeze    = "138439ad-76c6-4d75-8025-e68b94f68fa3"
	oldB791Efreet    = "34f52d00-301d-4065-a86b-8393b4c6e99d"
	oldB791Winds     = "70446c56-07de-4d7f-94ae-4da88fba41cf"
	oldB791Shoreline = "01fd889f-1839-44b9-b936-0e62c8f8a730"
	oldB791Mundungu  = "3fb4161d-a5b9-411d-942c-2fdb039e7bde"
)

func TestOldB791EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB791Brood: "Brood of Cockroaches", oldB791Wand: "Wand of Denial", oldB791Djinn: "Waterspout Djinn",
		oldB791Champion: "Talruum Champion", oldB791Assassin: "Suq'Ata Assassin", oldB791Breeze: "Breezekeeper",
		oldB791Efreet: "Shimmering Efreet", oldB791Winds: "Katabatic Winds", oldB791Shoreline: "Flooded Shoreline",
		oldB791Mundungu: "Mundungu",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB791KeywordCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, c := range []struct {
		name, tl, oracle string
		kws              []string
	}{
		{"Breezekeeper", "Creature — Djinn", oldB791Breeze, []string{"flying", game.KeywordPhasing}},
		{"Shimmering Efreet", "Creature — Efreet", oldB791Efreet, []string{"flying", game.KeywordPhasing}},
		{"Suq'Ata Assassin", "Creature — Human Assassin", oldB791Assassin, []string{"fear"}},
		{"Talruum Champion", "Creature — Minotaur", oldB791Champion, []string{"first strike"}},
		{"Waterspout Djinn", "Creature — Djinn", oldB791Djinn, []string{"flying"}},
	} {
		id := effectstest.B12Push(g, me.ID, c.name, c.tl, c.oracle, 2, 2)
		for _, kw := range c.kws {
			if !effectstest.P1g2HasKeyword(g, id, kw) {
				t.Errorf("%s has no %s", c.name, kw)
			}
		}
	}
}

func TestOldB791BroodOfCockroachesReturnsAtTheEndStepForALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	brood := effectstest.B12Push(g, me.ID, "Brood of Cockroaches", "Creature — Insect", oldB791Brood, 1, 1)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(brood); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(brood) {
		t.Fatal("the Brood is not in the graveyard")
	}
	life := effectstest.LifeOf(g, me.ID)
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(brood) {
		t.Error("the Brood did not return to its owner's hand")
	}
	if got := effectstest.LifeOf(g, me.ID); got != life-1 {
		t.Errorf("life %d → %d, want -1", life, got)
	}
}

func TestOldB791BroodOfCockroachesStaysGoneIfItLeftTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	brood := effectstest.B12Push(g, me.ID, "Brood of Cockroaches", "Creature — Insect", oldB791Brood, 1, 1)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(brood); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _, _ = me.Graveyard.Remove(brood) })
	life := effectstest.LifeOf(g, me.ID)
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(brood) {
		t.Error("an exiled Brood came back")
	}
	if got := effectstest.LifeOf(g, me.ID); got != life-1 {
		t.Errorf("life %d → %d, want -1 even though the card is gone", life, got)
	}
}

func TestOldB791WandOfDenialMillsANonlandCardForTwoLife(t *testing.T) {
	setup := func(t *testing.T, top game.Card) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		wand := effectstest.PushCatalogPermanent(g, me.ID, "Wand of Denial", "Artifact", oldB791Wand, false)
		top.Owner, top.Controller = opp.ID, opp.ID
		top.InstanceID = uuid.New()
		id := top.InstanceID
		opp.Library.PushTop(top)
		effectstest.B16Activate(t, g, me.ID, wand, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
		effectstest.PassPriorityAroundTable(t, g)
		return g, me, opp, wand, id
	}
	t.Run("paid", func(t *testing.T) {
		g, me, opp, _, id := setup(t, game.Card{Name: "Fireball", TypeLine: "Sorcery"})
		life := effectstest.LifeOf(g, me.ID)
		prompt := effectstest.AnswerPayUnless(t, g, me.ID, true)
		_ = prompt
		effectstest.PassPriorityAroundTable(t, g)
		if !opp.Graveyard.Contains(id) {
			t.Error("the nonland card was not put into its owner's graveyard")
		}
		if got := effectstest.LifeOf(g, me.ID); got != life-2 {
			t.Errorf("life %d → %d, want -2", life, got)
		}
	})
	t.Run("declined", func(t *testing.T) {
		g, me, opp, _, id := setup(t, game.Card{Name: "Fireball", TypeLine: "Sorcery"})
		effectstest.AnswerPayUnless(t, g, me.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Graveyard.Contains(id) {
			t.Error("the card was milled without paying")
		}
	})
	t.Run("a land is never offered", func(t *testing.T) {
		g, me, _, _, _ := setup(t, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
		if effectstest.HasPayUnlessFor(g, me.ID) {
			t.Error("a prompt was raised for a land card")
		}
	})
}

func TestOldB791WaterspoutDjinnReturnsAnUntappedIslandOrIsSacrificed(t *testing.T) {
	t.Run("returns an Island", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[1]
		djinn := effectstest.B12Push(g, me.ID, "Waterspout Djinn", "Creature — Djinn", oldB791Djinn, 4, 4)
		island := effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		prompt := effectstest.LatestChooseCards(g, me.ID)
		if prompt == nil {
			t.Fatal("no Island prompt")
		}
		if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{island}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(djinn) {
			t.Error("the Djinn was sacrificed although an Island was returned")
		}
		if !me.Hand.Contains(island) {
			t.Error("the Island did not return to its owner's hand")
		}
	})
	t.Run("no untapped Island", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[1]
		djinn := effectstest.B12Push(g, me.ID, "Waterspout Djinn", "Creature — Djinn", oldB791Djinn, 4, 4)
		tapped := effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		if err := g.TapCard(tapped, true); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(djinn) {
			t.Error("the Djinn survived with only a tapped Island")
		}
	})
}

func TestOldB791TalruumChampionStripsFirstStrikeFromTheCreatureItFaces(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champ := effectstest.B12Push(g, me.ID, "Talruum Champion", "Creature — Minotaur", oldB791Champion, 3, 3)
	striker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Striker", TypeLine: "Creature — Knight", Power: 2, Toughness: 2,
		Keywords: []string{"first strike"}, Owner: opp.ID, Controller: opp.ID,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: me.ID, CardID: champ, Target: striker, Amount: 1})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.P1g2HasKeyword(g, striker, "first strike") {
		t.Error("the attacker the Champion blocks kept first strike")
	}
	if !effectstest.P1g2HasKeyword(g, champ, "first strike") {
		t.Error("the Champion lost its own first strike")
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	champ = effectstest.B12Push(g, me.ID, "Talruum Champion", "Creature — Minotaur", oldB791Champion, 3, 3)
	blocker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Blocker", TypeLine: "Creature — Knight", Power: 2, Toughness: 2,
		Keywords: []string{"first strike"}, Owner: opp.ID, Controller: opp.ID,
	})
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: opp.ID, CardID: blocker, Target: champ, Amount: 1})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.P1g2HasKeyword(g, blocker, "first strike") {
		t.Error("the creature blocking the Champion kept first strike")
	}
}

func TestOldB791SuqAtaAssassinPoisonsTheDefenderWhenUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	assassin := effectstest.B12Push(g, me.ID, "Suq'Ata Assassin", "Creature — Human Assassin", oldB791Assassin, 1, 1)
	effectstest.BrAttack(t, g, assassin)
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlayerCounterFor(g, opp.ID, game.CounterPoison); n != 1 {
		t.Errorf("defending player has %d poison counters, want 1", n)
	}
}

func TestOldB791ShimmeringEfreetPhasesATargetOutWhenItPhasesIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	efreet := effectstest.B12Push(g, me.ID, "Shimmering Efreet", "Creature — Efreet", oldB791Efreet, 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventPhaseIn, Actor: me.ID, CardID: efreet})
	})
	if pick := effectstest.PassUntilPickTarget(t, g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.PhasedOutInCatalogGame(g, victim) {
		t.Error("the targeted creature did not phase out")
	}
}

func TestOldB791KatabaticWindsGroundsFlyersAndStopsTheirTapAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Katabatic Winds", "Enchantment", oldB791Winds, false)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flyer", TypeLine: "Creature — Bird", Power: 2, Toughness: 2,
		Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	ground := effectstest.B12Creature(g, me.ID, "Ground", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(flyer, opp.ID); err == nil {
		t.Error("a creature with flying attacked under Katabatic Winds")
	}
	if err := g.DeclareAttacker(ground, opp.ID); err != nil {
		t.Errorf("a creature without flying could not attack: %v", err)
	}
}

func TestOldB791KatabaticWindsOnlyStopsAbilitiesWithTapInTheCost(t *testing.T) {
	r := effects.Old791FlyingCreaturesCantActivateTapAbilities("Katabatic Winds")
	flyer := game.Card{InstanceID: uuid.New(), TypeLine: "Creature — Bird", Keywords: []string{"flying"}}
	ground := game.Card{InstanceID: uuid.New(), TypeLine: "Creature — Bear"}
	tapAbility := game.ActivationAbility{Label: "{T}: Do a thing.", TapCost: true}
	freeAbility := game.ActivationAbility{Label: "{1}: Do a thing."}
	cases := []struct {
		name   string
		card   game.Card
		ab     game.ActivationAbility
		forbid bool
	}{
		{"flyer with a {T} ability", flyer, tapAbility, true},
		{"flyer with a {T} mana ability", flyer, game.ActivationAbility{Label: "{T}: Add {G}.", Mana: true, TapCost: true}, true},
		{"flyer without {T}", flyer, freeAbility, false},
		{"ground creature with {T}", ground, tapAbility, false},
	}
	for _, c := range cases {
		got := r.Forbids(game.ActivationQuery{Card: c.card, FromZone: game.ZoneBattlefield, Ability: c.ab})
		if got != c.forbid {
			t.Errorf("%s: forbidden=%v, want %v", c.name, got, c.forbid)
		}
	}
}

func TestOldB791FloodedShorelineReturnsTwoIslandsToBounceACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	shore := effectstest.PushCatalogPermanent(g, me.ID, "Flooded Shoreline", "Enchantment", oldB791Shoreline, false)
	a := effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	b := effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	forest := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, shore, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear), ReturnIDs: []uuid.UUID{a, forest}}); err == nil {
		t.Fatal("a Forest paid for the Islands")
	}
	effectstest.B16Activate(t, g, me.ID, shore, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear), ReturnIDs: []uuid.UUID{a, b}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the targeted creature was not returned to its owner's hand")
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the Islands were not returned")
	}
	if !g.Battlefield.Contains(forest) {
		t.Error("the Forest left the battlefield")
	}
}

func TestOldB791MundunguCountersUnlessOneManaAndOneLifeIsPaid(t *testing.T) {
	board := func(t *testing.T) (*game.Game, uuid.UUID, *game.Player) {
		g := effectstest.NewCatalogGame(t)
		caster, them := g.Seats[0], g.Seats[1]
		effectstest.ToMainForCost(t, g)
		victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
		mundungu := effectstest.PushCatalogPermanent(g, them.ID, "Mundungu", "Creature — Human Wizard", oldB791Mundungu, false)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority to the responder: %v", err)
		}
		if err := g.ActivateCatalogAbility(them.ID, mundungu, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err != nil {
			t.Fatalf("activating Mundungu: %v", err)
		}
		effectstest.PassUntilTaxed(t, g, caster.ID)
		return g, victim, caster
	}
	g, victim, caster := board(t)
	effectstest.AnswerPayUnless(t, g, caster.ID, false)
	if g.Stack.Contains(victim) {
		t.Error("a declined payment must counter the spell")
	}

	g, victim, caster = board(t)
	effectstest.FillPool(caster, 1)
	life := effectstest.LifeOf(g, caster.ID)
	effectstest.AnswerPayUnless(t, g, caster.ID, true)
	if !g.Stack.Contains(victim) {
		t.Error("a paid tax countered the spell anyway")
	}
	if got := effectstest.LifeOf(g, caster.ID); got != life-1 {
		t.Errorf("life %d → %d, want -1", life, got)
	}
}
