package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b121_cards_test.go — Tales of Middle-earth Commander sorceries
// (S82 GAP batch 121, #5948). Every test plays the real catalog entry
// through the engine.

const (
	gapB121ForthEorlingasOracle  = "35807c14-4647-482f-9ba1-6d640fd5e8d8"
	gapB121RallyGaladhrimOracle  = "84b2d0f6-ad98-4a71-a3d5-645b0f2db6c6"
	gapB121MistsOfLorienOracle   = "65259ba4-956d-46a3-abd9-791d02015b41"
	gapB121CallForAidOracle      = "6d650eef-1156-4e3b-8fa8-90a6ee04d9a8"
	gapB121SorcerousSquallOracle = "a986d83d-22f1-45c3-bc2d-f5b210c539fb"
	gapB121FellBeastsShriekOracl = "d660f6a0-3642-4445-b877-fda2a2950680"
	gapB121SubjugateOracle       = "eb56141a-eb43-40e9-8ff1-e0ef5424f6a4"
)

func TestGapB121EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB121ForthEorlingasOracle:  "Forth Eorlingas!",
		gapB121RallyGaladhrimOracle:  "Rally the Galadhrim",
		gapB121MistsOfLorienOracle:   "Mists of Lórien",
		gapB121CallForAidOracle:      "Call for Aid",
		gapB121SorcerousSquallOracle: "Sorcerous Squall",
		gapB121FellBeastsShriekOracl: "Fell Beast's Shriek",
		gapB121SubjugateOracle:       "Subjugate the Hobbits",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

// gapB121Costed pushes a creature with a printed mana cost, so its mana
// value is real.
func gapB121Costed(g *game.Game, owner uuid.UUID, name, cost string, commander bool) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", ManaCost: cost,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner, IsCommander: commander,
	})
}

func gapB121Controller(g *game.Game, id uuid.UUID) uuid.UUID {
	c, _ := effectstest.BattlefieldCard(g, id)
	return c.Controller
}

func TestGapB121SubjugateTheHobbitsTakesNoncommanderCreaturesOfManaValueThreeOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	cheap := gapB121Costed(g, opp.ID, "Cheap", "{1}{G}", false)
	three := gapB121Costed(g, g.Seats[2].ID, "Three", "{1}{G}{G}", false)
	four := gapB121Costed(g, opp.ID, "Four", "{2}{G}{G}", false)
	cmdr := gapB121Costed(g, opp.ID, "Their Commander", "{G}", true)

	effectstest.TdCastFromHand(t, g, game.Card{Name: "Subjugate the Hobbits", TypeLine: "Sorcery", OracleID: gapB121SubjugateOracle, ManaCost: "{5}{U}{U}"},
		game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{cheap, three} {
		if gapB121Controller(g, id) != me.ID {
			t.Errorf("%s was not taken", id)
		}
	}
	if gapB121Controller(g, four) == me.ID {
		t.Error("a mana value 4 creature was taken")
	}
	if gapB121Controller(g, cmdr) == me.ID {
		t.Error("a commander was taken")
	}
	// No stated duration: the theft survives the turn.
	effectstest.EndTurn(t, g)
	if gapB121Controller(g, cheap) != me.ID {
		t.Error("the control change ended with the turn")
	}
}

func TestGapB121CallForAidTakesEverythingAndRestrictsTheCaster(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[g.Turn.ActiveSeat], g.Seats[1], g.Seats[2]
	a := gapB121Costed(g, opp.ID, "A", "{1}", false)
	b := gapB121Costed(g, opp.ID, "B", "{2}", false)
	bystander := gapB121Costed(g, other.ID, "Bystander", "{1}", false)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(a) })

	effectstest.TdCastFromHand(t, g, game.Card{Name: "Call for Aid", TypeLine: "Sorcery", OracleID: gapB121CallForAidOracle, ManaCost: "{4}{R}"},
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{a, b} {
		c, _ := effectstest.BattlefieldCard(g, id)
		if c.Controller != me.ID || c.Tapped || !game.HasKeyword(&c, "haste") {
			t.Errorf("%s: controller %v tapped %v haste %v", c.Name, c.Controller, c.Tapped, game.HasKeyword(&c, "haste"))
		}
		if !g.CantBeSacrificedForEffect(id) {
			t.Errorf("%s can be sacrificed", c.Name)
		}
	}
	if gapB121Controller(g, bystander) != other.ID {
		t.Error("another opponent's creature was taken")
	}
	// "You can't attack that player this turn": the stolen creature may
	// attack someone else, never the victim.
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err == nil {
		t.Error("the caster attacked the player whose creatures were taken")
	}
	if err := g.DeclareAttacker(a, other.ID); err != nil {
		t.Errorf("attacking a different player was refused: %v", err)
	}
	effectstest.EndTurn(t, g)
	if gapB121Controller(g, a) != opp.ID || gapB121Controller(g, b) != opp.ID {
		t.Error("control did not return at end of turn")
	}
}

func TestGapB121ForthEorlingasMakesXKnightsAndCrownsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Forth Eorlingas!", TypeLine: "Sorcery", OracleID: gapB121ForthEorlingasOracle, ManaCost: "{X}{R}{W}"},
		game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)

	var knights []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Human Knight" && c.Controller == me.ID {
			if c.Power != 2 || c.Toughness != 2 || !game.HasKeyword(&c, "trample") || !game.HasKeyword(&c, "haste") {
				t.Errorf("token is %d/%d trample %v haste %v", c.Power, c.Toughness, game.HasKeyword(&c, "trample"), game.HasKeyword(&c, "haste"))
			}
			knights = append(knights, c.InstanceID)
		}
	}
	if len(knights) != 3 {
		t.Fatalf("%d Human Knight tokens, want 3", len(knights))
	}
	if effects.YoureTheMonarch(g, me.ID) {
		t.Fatal("the caster was crowned before any combat damage")
	}
	effectstest.DeclareAttack(t, g, opp.ID, knights[0])
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.YoureTheMonarch(g, me.ID) {
		t.Error("combat damage to a player did not make the caster the monarch")
	}
}

func TestGapB121ForthEorlingasWithXZeroStillWatchesForCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Forth Eorlingas!", TypeLine: "Sorcery", OracleID: gapB121ForthEorlingasOracle, ManaCost: "{X}{R}{W}"},
		game.CastSpellParams{XValue: 0})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.YoureTheMonarch(g, me.ID) {
		t.Error("a creature already on the battlefield did not crown the caster")
	}
}

func TestGapB121RallyTheGaladhrimCopiesAndConspireDoublesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Elf", TypeLine: "Creature — Elf",
		ManaCost: "{G}", Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Rally the Galadhrim", TypeLine: "Sorcery", OracleID: gapB121RallyGaladhrimOracle,
		ManaCost: "{2}{G}{U}", Colors: []string{"G", "U"}},
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: elf}}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Green Elf")); n != 2 {
		t.Fatalf("%d Green Elves, want the original and one token copy", n)
	}
}

func TestGapB121RallyTheGaladhrimCanBeConspired(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mk := func(name string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Elf",
			ManaCost: "{G}", Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	}
	elf, tapA, tapB := mk("Green Elf"), mk("Tapper A"), mk("Tapper B")
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Rally the Galadhrim", TypeLine: "Sorcery", OracleID: gapB121RallyGaladhrimOracle,
		ManaCost: "{2}{G}{U}", Colors: []string{"G", "U"}},
		game.CastSpellParams{
			Targets:       []game.TargetRef{{Kind: game.TargetCard, ID: elf}},
			OptionalCosts: []int{0},
			TeamworkIDs:   []uuid.UUID{tapA, tapB},
		})
	if !effectstest.IsTapped(g, tapA) || !effectstest.IsTapped(g, tapB) {
		t.Fatal("conspire did not tap the two creatures")
	}
	sqResolveAllCards(t, g, me, game.TargetRef{Kind: game.TargetCard, ID: elf})
	if n := len(effectstest.BattlefieldIDsNamed(g, "Green Elf")); n != 3 {
		t.Errorf("%d Green Elves, want the original and two token copies", n)
	}
}

// sqResolveAllCards passes until the stack is empty, keeping the target
// when a copy asks for new ones.
func sqResolveAllCards(t *testing.T, g *game.Game, me *game.Player, keep game.TargetRef) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if effectstest.StackFullyEmpty(g) && effectstest.LatestPickTarget(g, me.ID) == nil {
			return
		}
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, keep); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
}

func TestGapB121MistsOfLorienReturnsEveryNonlandPermanentOfTheSameManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	target := gapB121Costed(g, opp.ID, "Target Two", "{1}{G}", false)
	sameMine := gapB121Costed(g, me.ID, "Same Two", "{1}{U}", false)
	sameTheirs := gapB121Costed(g, g.Seats[2].ID, "Same Two Too", "{2}", false)
	other := gapB121Costed(g, opp.ID, "Three", "{2}{G}", false)
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
		Owner: opp.ID, Controller: opp.ID})

	effectstest.TdCastFromHand(t, g, game.Card{Name: "Mists of Lórien", TypeLine: "Sorcery", OracleID: gapB121MistsOfLorienOracle, ManaCost: "{2}{U}"},
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{target, sameMine, sameTheirs} {
		if effectstest.BattlefieldHas(g, id) {
			t.Errorf("%s was not returned", id)
		}
	}
	if !effectstest.BattlefieldHas(g, other) {
		t.Error("a mana value 3 permanent was returned")
	}
	if !effectstest.BattlefieldHas(g, land) {
		t.Error("a land was returned")
	}
}

func TestGapB121MistsOfLorienReplicates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	one := gapB121Costed(g, opp.ID, "One", "{1}", false)
	two := gapB121Costed(g, g.Seats[2].ID, "Two", "{2}", false)
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Mists of Lórien", TypeLine: "Sorcery", OracleID: gapB121MistsOfLorienOracle, ManaCost: "{2}{U}"},
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: one}}, OptionalCosts: []int{0}})
	for i := 0; i < 64; i++ {
		if effectstest.StackFullyEmpty(g) && effectstest.LatestPickTarget(g, me.ID) == nil {
			break
		}
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			// The copy takes a new target: the mana value 2 creature.
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: two}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if effectstest.BattlefieldHas(g, one) || effectstest.BattlefieldHas(g, two) {
		t.Errorf("the replicate copy did not return its own target (one on battlefield %v, two %v)", effectstest.BattlefieldHas(g, one), effectstest.BattlefieldHas(g, two))
	}
}

func TestGapB121SorcerousSquallMillsNineThenCastsAnInstantFromTheirGraveyardAndExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.StockLibraryFor(g, opp, 12)
	bolt := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
			OracleID: effectstest.LightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
	})
	before, gyBefore := opp.Life, opp.Graveyard.Size()

	effectstest.TdCastFromHand(t, g, game.Card{Name: "Sorcerous Squall", TypeLine: "Sorcery", OracleID: gapB121SorcerousSquallOracle, ManaCost: "{6}{U}{U}{U}"},
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - gyBefore; got != 9 {
		t.Fatalf("the opponent milled %d cards, want 9", got)
	}
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, bolt)
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, bolt, game.CastSpellParams{
		Strict: true, AlternativeCost: game.AltCostKeyGrantedCost, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("free cast from their graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("the Bolt dealt %d, want 3", before-opp.Life)
	}
	if !effectstest.InExile(g, bolt) {
		t.Error("the spell was not exiled instead of going to the graveyard")
	}
}

func TestGapB121SorcerousSquallOffersOnlyInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
			Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	})
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Sorcerous Squall", TypeLine: "Sorcery", OracleID: gapB121SorcerousSquallOracle, ManaCost: "{6}{U}{U}{U}"},
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a creature card was offered for the free cast")
	}
}

func TestGapB121FellBeastsShriekEachOpponentChoosesThenTapAndGoad(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	o1, o2, o3 := g.Seats[1], g.Seats[2], g.Seats[3]
	a1 := effectstest.PushVanillaCreature(g, o1.ID, "One A", 2, 2)
	b1 := effectstest.PushVanillaCreature(g, o1.ID, "One B", 3, 3)
	only := effectstest.PushVanillaCreature(g, o2.ID, "Two Only", 1, 1)
	_ = o3

	effectstest.TdCastFromHand(t, g, game.Card{Name: "Fell Beast's Shriek", TypeLine: "Sorcery", OracleID: gapB121FellBeastsShriekOracl, ManaCost: "{U}{R}"},
		game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, o1.ID)
	if pick == nil {
		t.Fatal("the first opponent was not asked which creature to give up")
	}
	if effectstest.IsTapped(g, only) {
		t.Error("a creature was tapped before every opponent had chosen")
	}
	if err := g.ResolveChooseCards(pick.ID, o1.ID, []uuid.UUID{b1}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{b1, only} {
		if !effectstest.IsTapped(g, id) {
			t.Errorf("%s is not tapped", id)
		}
		if effectstest.B33Goaded(t, g, id) != me.ID {
			t.Errorf("%s is not goaded by the caster", id)
		}
	}
	if effectstest.IsTapped(g, a1) || effectstest.B33Goaded(t, g, a1) != uuid.Nil {
		t.Error("an unchosen creature was tapped or goaded")
	}
}

func TestGapB121FellBeastsShriekSplicesOntoAnInstant(t *testing.T) {
	spec, ok := effects.Lookup(gapB121FellBeastsShriekOracl)
	if !ok || spec.Splice == nil {
		t.Fatal("Fell Beast's Shriek has no splice")
	}
	if spec.Splice.Onto != game.SpliceOntoInstantOrSorcery || spec.Splice.ManaCost != "{2}{U}{R}" {
		t.Errorf("splice = %+v, want onto instant or sorcery {2}{U}{R}", spec.Splice)
	}
}
