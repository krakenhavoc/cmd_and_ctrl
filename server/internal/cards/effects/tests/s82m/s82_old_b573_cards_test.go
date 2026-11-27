package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b573_cards_test.go — Guildpact cards, S82 OLD batch 573 (#5599),
// played through the real catalog entries.

const (
	o573Mimeofacture  = "79b978b8-44c8-4ca8-bdd5-10490145398d"
	o573Predatory     = "c5866507-ef3b-446a-99ea-d31d974bdb01"
	o573YoreTiller    = "7e9d61d3-6ba7-45ac-a55c-a2f36595c5d2"
	o573Moratorium    = "bb0d89e7-4ae6-4ac2-a0e6-51ab08226069"
	o573KillerInst    = "bbf90463-280a-413e-b8a3-4577dcf11e54"
	o573Leyline       = "bc9f159b-984d-4b9f-8904-b38b5cb79636"
	o573ConjurersBan  = "5e6cbe03-6d6a-413e-9f3c-2770d408cc79"
	o573DuneBrood     = "634bd800-8caa-47ae-8b70-2c66baf9a355"
	o573SanguinePrae  = "af6124cf-e8c5-44f3-964a-4ac0caa6f22b"
	o573CreatureTL    = "Creature — Bear"
	o573LegendaryBear = "Legendary Creature — Bear"
)

func TestS82OldB573CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o573Mimeofacture: "Mimeofacture", o573Predatory: "Predatory Focus", o573YoreTiller: "Yore-Tiller Nephilim",
		o573Moratorium: "Moratorium Stone", o573KillerInst: "Killer Instinct", o573Leyline: "Leyline of Singularity",
		o573ConjurersBan: "Conjurer's Ban", o573DuneBrood: "Dune-Brood Nephilim", o573SanguinePrae: "Sanguine Praetor",
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

func TestS82OldB573MimeofactureTakesTheNamedCardFromTheirLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", o573CreatureTL, 2, 2)
	twin := effectstest.SsCard(opp.Library, opp.ID, "Grizzly Bears", o573CreatureTL, 2, 2)
	other := effectstest.SsCard(opp.Library, opp.ID, "Other Bear", o573CreatureTL, 2, 2)
	spec, _ := effects.Lookup(o573Mimeofacture)
	if len(spec.OptionalCosts) != 1 || len(spec.Triggered) != 1 {
		t.Fatalf("Mimeofacture should declare replicate: %d optional costs, %d triggers", len(spec.OptionalCosts), len(spec.Triggered))
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Mimeofacture", "Sorcery", o573Mimeofacture, nil); err == nil {
		t.Fatal("Mimeofacture cast without a target")
	}
	effectstest.CastCatalogSpell(t, g, "Mimeofacture", "Sorcery", o573Mimeofacture, effectstest.CardRefs(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, twin)
	c, ok := effectstest.BattlefieldCard(g, twin)
	if !ok {
		t.Fatal("the same-named card did not enter the battlefield")
	}
	if c.Controller != me.ID {
		t.Errorf("the card is controlled by %v, want the caster", c.Controller)
	}
	if opp.Library.Contains(twin) || !opp.Library.Contains(other) {
		t.Error("the search took the wrong card out of their library")
	}
}

func TestS82OldB573PredatoryFocusLetsYourCreaturesAssignAsThoughUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", o573CreatureTL, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", o573CreatureTL, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Predatory Focus", "Sorcery", o573Predatory, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.ApaLive(g, mine).Effective().AssignsAsThoughUnblocked {
		t.Error("my creature may not assign its combat damage as though it weren't blocked")
	}
	if effectstest.ApaLive(g, theirs).Effective().AssignsAsThoughUnblocked {
		t.Error("their creature got the option")
	}
	late := effectstest.B12Creature(g, me.ID, "Late Bear", o573CreatureTL, 2, 2)
	if effectstest.ApaLive(g, late).Effective().AssignsAsThoughUnblocked {
		t.Error("a creature that arrived after resolution got the option (CR 611.2c)")
	}
}

func TestS82OldB573YoreTillerNephilimReturnsACreatureTappedAndAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.B12Push(g, me.ID, "Yore-Tiller Nephilim", "Creature — Nephilim", o573YoreTiller, 2, 2)
	dead := effectstest.SsCard(me.Graveyard, me.ID, "Dead Bear", o573CreatureTL, 4, 4)
	spell := effectstest.SsCard(me.Graveyard, me.ID, "Dead Spell", "Sorcery", 0, 0)
	effectstest.DeclareAttack(t, g, opp.ID, id)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerTriggerTargets(t, g, me.ID, dead)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestOptionPickFor(g, me.ID); pick != nil {
		effectstest.AnswerOptionPick(t, g, me.ID, 0)
	}
	c, ok := effectstest.BattlefieldCard(g, dead)
	if !ok {
		t.Fatal("the creature card did not return")
	}
	if !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("tapped=%v attacking=%v, want tapped and attacking", c.Tapped, c.AttackingTarget)
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("a noncreature card left the graveyard")
	}
}

func TestS82OldB573MoratoriumStoneExilesOneCardFromAnyGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stone := effectstest.B12Push(g, me.ID, "Moratorium Stone", "Artifact", o573Moratorium, 0, 0)
	mine := effectstest.SsCard(me.Graveyard, me.ID, "Mine", "Sorcery", 0, 0)
	theirs := effectstest.SsCard(opp.Graveyard, opp.ID, "Theirs", "Sorcery", 0, 0)
	effectstest.RavB2MustActivate(t, g, me.ID, stone, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}, "C", "C")
	if opp.Graveyard.Contains(theirs) || !me.Graveyard.Contains(mine) {
		t.Error("the first ability did not exile exactly the targeted card")
	}
	if !effectstest.TappedForTest(t, g, stone) {
		t.Error("the first ability taps the Stone")
	}
}

func TestS82OldB573MoratoriumStoneSecondAbilityExilesEveryCopyByName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stone := effectstest.B12Push(g, me.ID, "Moratorium Stone", "Artifact", o573Moratorium, 0, 0)
	target := effectstest.SsCard(opp.Graveyard, opp.ID, "Grizzly Bears", o573CreatureTL, 2, 2)
	twinMine := effectstest.SsCard(me.Graveyard, me.ID, "Grizzly Bears", o573CreatureTL, 2, 2)
	other := effectstest.SsCard(me.Graveyard, me.ID, "Other Bear", o573CreatureTL, 2, 2)
	onField := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", o573CreatureTL, 2, 2)
	otherField := effectstest.B12Creature(g, opp.ID, "Other Bear", o573CreatureTL, 2, 2)
	land := effectstest.SsCard(opp.Graveyard, opp.ID, "Forest", "Basic Land — Forest", 0, 0)
	if err := effectstest.RavB2Activate(t, g, me.ID, stone, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(land)}, "C", "C", "W", "B"); err == nil {
		t.Fatal("the second ability targeted a land card")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, stone, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(target)}, "C", "C", "W", "B")
	if g.Battlefield.Contains(stone) {
		t.Error("the Stone is sacrificed as a cost")
	}
	for name, id := range map[string]uuid.UUID{"target": target, "twin in my graveyard": twinMine} {
		if opp.Graveyard.Contains(id) || me.Graveyard.Contains(id) {
			t.Errorf("the %s was not exiled", name)
		}
	}
	if g.Battlefield.Contains(onField) {
		t.Error("the permanent with that name was not exiled")
	}
	if !me.Graveyard.Contains(other) || !g.Battlefield.Contains(otherField) {
		t.Error("a card with another name was exiled")
	}
}

func TestS82OldB573KillerInstinctPutsAHastyCreatureOntoTheBattlefieldAndSacrificesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_ = effectstest.B12Push(g, me.ID, "Killer Instinct", "Enchantment", o573KillerInst, 0, 0)
	top := effectstest.SsCard(me.Library, me.ID, "Top Bear", o573CreatureTL, 3, 3)
	g.WithWriteLock(func() {
		for i := range me.Library.Cards {
			if me.Library.Cards[i].InstanceID == top {
				me.Library.Cards[i].TypeLine = o573CreatureTL
			}
		}
	})
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(top) {
		t.Fatal("the revealed creature card was not put onto the battlefield")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, top, "haste") {
		t.Error("the creature did not gain haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(top) || !me.Graveyard.Contains(top) {
		t.Error("the creature was not sacrificed at the next end step")
	}
}

func TestS82OldB573KillerInstinctLeavesANoncreatureOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_ = effectstest.B12Push(g, me.ID, "Killer Instinct", "Enchantment", o573KillerInst, 0, 0)
	top := effectstest.SsCard(me.Library, me.ID, "Top Spell", "Sorcery", 0, 0)
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(top) || !me.Library.Contains(top) {
		t.Error("a noncreature card left the top of the library")
	}
}

func TestS82OldB573LeylineOfSingularityMakesNonlandPermanentsLegendary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spec, _ := effects.Lookup(o573Leyline)
	if spec.OpeningHand == nil {
		t.Error("Leyline of Singularity has no opening-hand clause")
	}
	mine := effectstest.B12Creature(g, me.ID, "My Bear", o573CreatureTL, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", o573CreatureTL, 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if effectstest.LayeredCard(t, g, mine).HasSupertype("Legendary") {
		t.Fatal("legendary before the Leyline is out")
	}
	ley := effectstest.B12Push(g, me.ID, "Leyline of Singularity", "Enchantment", o573Leyline, 0, 0)
	for name, id := range map[string]uuid.UUID{"mine": mine, "theirs": theirs, "the Leyline": ley} {
		if !effectstest.LayeredCard(t, g, id).HasSupertype("Legendary") {
			t.Errorf("%s is not legendary", name)
		}
	}
	if effectstest.LayeredCard(t, g, land).HasSupertype("Legendary") {
		t.Error("a land became legendary")
	}
}

func TestS82OldB573LeylineOfSingularityTriggersTheLegendRule(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	_ = effectstest.B12Push(g, me.ID, "Leyline of Singularity", "Enchantment", o573Leyline, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "Twin Bear", o573CreatureTL, 2, 2)
	effectstest.B12Creature(g, me.ID, "Twin Bear", o573CreatureTL, 2, 2)
	g.RunStateChecksForTest()
	if pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceLegendRule, me.ID); pick == nil {
		t.Fatalf("two same-named nonland permanents under the Leyline raised no legend-rule choice (first %v)", a)
	}
}

func TestS82OldB573ConjurersBanBansTheNameForEveryPlayerUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], effectstest.Top100Opponent(g)
	effectstest.SsCard(me.Library, me.ID, "Library Card", "Sorcery", 0, 0)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Conjurer's Ban", "Sorcery", o573ConjurersBan, nil)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatal("the caster was not asked for a card name")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Forest"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: Conjurer's Ban draws a card", me.Hand.Size(), hand+1)
	}
	forest := game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest"}
	plains := game.Card{InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains"}
	for _, p := range []*game.Player{me, opp} {
		if err := g.LandPlayGateFor(p.ID, forest, game.ZoneHand); !errors.Is(err, game.ErrCantPlayLand) {
			t.Errorf("%v may play a land with the chosen name: %v", p.ID, err)
		}
		if err := g.LandPlayGateFor(p.ID, plains, game.ZoneHand); errors.Is(err, game.ErrCantPlayLand) {
			t.Errorf("a land with another name was banned: %v", err)
		}
	}
	// A spell with the chosen name can't be cast by anyone, the caster included.
	for _, p := range []*game.Player{me, opp} {
		var forbidden bool
		g.WithWriteLock(func() {
			err := g.CastGateLocked(p.ID, game.Card{Name: "Forest", TypeLine: "Sorcery"}, game.ZoneHand, game.CastSpellParams{})
			forbidden = errors.Is(err, game.ErrCantCast)
		})
		if !forbidden {
			t.Errorf("%v may cast a spell with the chosen name", p.ID)
		}
	}
	for i := 0; i < len(g.Seats); i++ {
		effectstest.EndTurn(t, g)
		if g.Seats[g.Turn.ActiveSeat].ID == me.ID {
			break
		}
	}
	if err := g.LandPlayGateFor(me.ID, forest, game.ZoneHand); errors.Is(err, game.ErrCantPlayLand) {
		t.Errorf("the land ban outlived the caster's next turn: %v", err)
	}
}

func TestS82OldB573DuneBroodNephilimMakesASandPerLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.B12Push(g, me.ID, "Dune-Brood Nephilim", "Creature — Nephilim", o573DuneBrood, 3, 3)
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.AttackWith(t, g, opp.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	n, _ := effectstest.CountTokensNamed(g, me.ID, "Sand")
	if n != 3 {
		t.Errorf("%d Sand tokens, want 3 (one for each land I control)", n)
	}
}

func TestS82OldB573SanguinePraetorDestroysEveryCreatureOfTheSacrificedValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	praetor := effectstest.B12Push(g, me.ID, "Sanguine Praetor", "Creature — Avatar Praetor", o573SanguinePrae, 7, 5)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == praetor {
				g.Battlefield.Cards[i].ManaCost = "{6}{B}{B}"
			}
		}
	})
	two := func(owner uuid.UUID, name string, cost string) uuid.UUID {
		id := effectstest.B12Creature(g, owner, name, o573CreatureTL, 2, 2)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].ManaCost = cost
				}
			}
		})
		return id
	}
	fodder := two(me.ID, "Fodder", "{1}{G}")
	sameMine := two(me.ID, "Same Mine", "{2}")
	sameTheirs := two(opp.ID, "Same Theirs", "{G}{G}")
	differentTheirs := two(opp.ID, "Different", "{3}")
	params := game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}
	effectstest.RavB2MustActivate(t, g, me.ID, praetor, 0, params, "B")
	for name, id := range map[string]uuid.UUID{"the sacrificed creature": fodder, "my same-value creature": sameMine, "their same-value creature": sameTheirs} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", name)
		}
	}
	if !g.Battlefield.Contains(differentTheirs) || !g.Battlefield.Contains(praetor) {
		t.Error("a creature of another mana value was destroyed")
	}
}
