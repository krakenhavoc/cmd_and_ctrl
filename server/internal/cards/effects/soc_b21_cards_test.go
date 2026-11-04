package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b21_cards_test.go — Secrets of Strixhaven, batch 21 (#4456). Each test
// plays the real catalog entry through the engine and asserts what the
// printed text says.

const (
	socB21FiresingerOracle   = "1bf505a7-292e-4642-83b6-8ff41ebe5d51"
	socB21LumaretOracle      = "0109432d-5a2a-456f-ad39-b75cb4c73420"
	socB21KnowledgeOracle    = "0d6507b0-c986-49d1-aea3-77b2cf2a7433"
	socB21BrushOffOracle     = "873694c0-c100-4d24-a74f-1ab26442dd18"
	socB21BanemakerOracle    = "c531801c-acf0-45ae-a112-054cc5ccb6ed"
	socB21BarrageOracle      = "c47caa45-e7fc-414a-8a5d-59b58f1751f3"
	socB21ComposerOracle     = "9b3155c1-c662-4ff1-80e2-4064e6233dc5"
	socB21StrifeknightOracle = "b0b9a7c1-f515-4627-a050-b595d6641611"
	socB21InspirationOracle  = "9c055ef0-1164-4a9f-80de-23405423aed5"
	socB21OsteomancerOracle  = "fd01229f-1c4c-4ad1-bc6f-4023f24d8c71"
)

func TestSocB21EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB21FiresingerOracle:         "Blazing Firesinger",
		socB21FiresingerOracle + "#1":  "Seething Song",
		socB21LumaretOracle:            "Bogwater Lumaret",
		socB21KnowledgeOracle:          "Borrowed Knowledge",
		socB21BrushOffOracle:           "Brush Off",
		socB21BanemakerOracle:          "Burrog Banemaker",
		socB21BarrageOracle:            "Burrog Barrage",
		socB21ComposerOracle:           "Campus Composer",
		socB21ComposerOracle + "#1":    "Aqueous Aria",
		socB21StrifeknightOracle:       "Charging Strifeknight",
		socB21InspirationOracle:        "Chase Inspiration",
		socB21OsteomancerOracle:        "Cheerful Osteomancer",
		socB21OsteomancerOracle + "#1": "Raise Dead",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Blazing Firesinger // Seething Song ------------------------------------

func TestSocB21FiresingerEntersPreparedAndSeethingSongAddsFiveRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB21FiresingerOracle, nil, "Blazing Firesinger", "Creature — Dwarf Bard", "{2}{R}", 2, 3,
		"Seething Song", "Instant", "{2}{R}")
	castFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Blazing Firesinger did not enter prepared")
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Fatalf("%d mana in the pool before the copy was cast", n)
	}
	fraBCastCopy(t, g, me, "Seething Song")
	if n := me.ManaPool.CountColor("R"); n != 5 || me.ManaPool.Total() != 5 {
		t.Errorf("Seething Song added %d red of %d mana, want {R}{R}{R}{R}{R}", n, me.ManaPool.Total())
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
}

// --- Bogwater Lumaret --------------------------------------------------------

func TestSocB21LumaretGainsLifeForItselfAndEachOtherCreatureOfYours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	life := me.Life
	lumaret := game.Card{
		InstanceID: uuid.New(), Name: "Bogwater Lumaret", TypeLine: "Creature — Spirit Frog", OracleID: socB21LumaretOracle,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	}
	castFromHandAndResolve(t, g, me, lumaret)
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Fatalf("Lumaret entering: life %d -> %d, want +1", life, me.Life)
	}
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Fatalf("another creature of mine entering: life %d, want %d", me.Life, life+2)
	}
}

func TestSocB21LumaretIgnoresOpponentsCreaturesAndNoncreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Bogwater Lumaret", "Creature — Spirit Frog", socB21LumaretOracle, 2, 2)
	life := me.Life
	castAndResolveCreature(t, g, "Rock", "Artifact", "")
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("a noncreature entering gained life: %d -> %d", life, me.Life)
	}
	b13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("an opponent's creature entering gained life: %d -> %d", life, me.Life)
	}
}

// --- Borrowed Knowledge ------------------------------------------------------

func socB21SetHand(p *game.Player, n int) {
	p.Hand.Cards = nil
	for i := 0; i < n; i++ {
		c := game.NewCard("Hand Card", p.ID)
		c.TypeLine = "Sorcery"
		p.Hand.PushTop(c)
	}
}

func socB21CastKnowledge(t *testing.T, g *game.Game, mode int, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Borrowed Knowledge", TypeLine: "Sorcery", OracleID: socB21KnowledgeOracle,
		Owner: me.ID, Controller: me.ID,
	})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{mode}, Targets: targets})
}

func TestSocB21KnowledgeDrawsAsManyAsTheOpponentsHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	socB21SetHand(me, 3)
	socB21SetHand(opp, 6)
	gy := me.Graveyard.Size()
	if err := socB21CastKnowledge(t, g, 0, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 6 {
		t.Errorf("hand %d, want the opponent's 6 cards drawn", me.Hand.Size())
	}
	if me.Graveyard.Size() != gy+3+1 {
		t.Errorf("graveyard grew by %d, want the 3 hand cards and the spell", me.Graveyard.Size()-gy)
	}
	if opp.Hand.Size() != 6 {
		t.Errorf("the opponent's hand changed to %d", opp.Hand.Size())
	}
}

func TestSocB21KnowledgeFirstModeWithAnEmptyHandStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	socB21SetHand(me, 0)
	socB21SetHand(opp, 4)
	if err := socB21CastKnowledge(t, g, 0, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 4 {
		t.Errorf("hand %d, want 4", me.Hand.Size())
	}
}

func TestSocB21KnowledgeFirstModeTakesAnOpponentOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	socB21SetHand(me, 2)
	err := socB21CastKnowledge(t, g, 0, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("targeting yourself: err = %v, want ErrIllegalTarget", err)
	}
}

func TestSocB21KnowledgeSecondModeDrawsWhatItDiscarded(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	socB21SetHand(me, 4)
	socB21SetHand(opp, 7)
	gy := me.Graveyard.Size()
	if err := socB21CastKnowledge(t, g, 1, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 4 {
		t.Errorf("hand %d, want the 4 discarded cards redrawn", me.Hand.Size())
	}
	if me.Graveyard.Size() != gy+4+1 {
		t.Errorf("graveyard grew by %d, want the 4 hand cards and the spell", me.Graveyard.Size()-gy)
	}
}

func TestSocB21KnowledgeSecondModeWithAnEmptyHandDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	socB21SetHand(me, 0)
	if err := socB21CastKnowledge(t, g, 1, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 {
		t.Errorf("hand %d, want 0", me.Hand.Size())
	}
}

// --- Brush Off ---------------------------------------------------------------

func socB21StackSpell(t *testing.T, g *game.Game, name, typeLine string) uuid.UUID {
	t.Helper()
	return castCatalogSpell(t, g, name, typeLine, "", nil)
}

func TestSocB21BrushOffCostsOneBlueLessAtAnInstantOrSorcery(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		generic, blue  int
	}{
		{"instant", "Instant", 1, 1},
		{"sorcery", "Sorcery", 1, 1},
		{"creature", "Creature — Bear", 2, 2},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := socB21StackSpell(t, g, "Test Spell", tc.typeLine)
		c := selfPriced(t, g, me, socB21BrushOffOracle, "Instant", "{2}{U}{U}", game.ZoneHand,
			[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
		if c.Generic != tc.generic || len(c.Required) != tc.blue {
			t.Errorf("at a %s spell: generic %d, %d coloured, want generic %d and %d blue", tc.name, c.Generic, len(c.Required), tc.generic, tc.blue)
		}
	}
	g := newCatalogGame(t)
	if c := selfPriced(t, g, g.Seats[0], socB21BrushOffOracle, "Instant", "{2}{U}{U}", game.ZoneHand, nil); c.ManaValue() != 4 {
		t.Errorf("with no target: mana value %d, want 4", c.ManaValue())
	}
}

func TestSocB21BrushOffCounterCastForTwoMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := socB21StackSpell(t, g, "Test Instant", "Instant")
	b06AddMana(me, "U", "G")
	castFromHandForTest(t, g, me, "Brush Off", "Instant", "{2}{U}{U}", socB21BrushOffOracle, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left in the pool, the discounted cost is two", n)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(target) {
		t.Error("the targeted spell was not countered into the graveyard")
	}
}

func TestSocB21BrushOffAtACreatureSpellPaysFullPrice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := socB21StackSpell(t, g, "Test Bear", "Creature — Bear")
	b06AddMana(me, "U", "G")
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Brush Off", TypeLine: "Instant", ManaCost: "{2}{U}{U}", OracleID: socB21BrushOffOracle,
		Owner: me.ID, Controller: me.ID,
	})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	if err == nil {
		t.Fatal("two mana paid for Brush Off at a creature spell")
	}
}

// --- Burrog Banemaker --------------------------------------------------------

func TestSocB21BanemakerHasDeathtouchAndPumpsForOneBlackEachTime(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	frog := b12Push(g, me.ID, "Burrog Banemaker", "Creature — Frog Warlock", socB21BanemakerOracle, 1, 1)
	if !hasEffectiveKeyword(t, g, frog, "deathtouch") {
		t.Fatal("no deathtouch")
	}
	advanceTo(t, g, game.StepPrecombatMain)
	for i := 1; i <= 2; i++ {
		b06AddMana(me, "B", "B")
		if err := g.ActivateCatalogAbility(me.ID, frog, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		passPriorityAroundTable(t, g)
		if p, tg := socB14PT(t, g, frog); p != 1+i || tg != 1+i {
			t.Fatalf("after %d activations: %d/%d, want %d/%d", i, p, tg, 1+i, 1+i)
		}
	}
}

func TestSocB21BanemakerPumpIsRefusedWithoutItsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	frog := b12Push(g, me.ID, "Burrog Banemaker", "Creature — Frog Warlock", socB21BanemakerOracle, 1, 1)
	advanceTo(t, g, game.StepPrecombatMain)
	b06AddMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, frog, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("paid {1}{B} with no black mana")
	}
	if p, tg := socB14PT(t, g, frog); p != 1 || tg != 1 {
		t.Errorf("%d/%d after a refused activation", p, tg)
	}
}

// --- Burrog Barrage ----------------------------------------------------------

func socB21CastBarrage(t *testing.T, g *game.Game, targets ...uuid.UUID) error {
	t.Helper()
	refs := make([]game.TargetRef, len(targets))
	for i, id := range targets {
		refs[i] = game.TargetRef{Kind: game.TargetCard, ID: id}
	}
	id := uuid.New()
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Burrog Barrage", TypeLine: "Instant", OracleID: socB21BarrageOracle,
		Owner: me.ID, Controller: me.ID,
	})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Targets: refs})
}

func TestSocB21BarrageDealsDamageEqualToPowerWithoutAnEarlierSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Frog", "Creature — Frog", 2, 2)
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	if err := socB21CastBarrage(t, g, mine, theirs); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := socB14PT(t, g, mine); p != 2 {
		t.Errorf("power %d without another instant or sorcery, want 2", p)
	}
	c, ok := g.LookupCardForEffect(theirs)
	if !ok || !g.Battlefield.Contains(theirs) || c.DamageMarked != 2 {
		t.Errorf("the Ogre should carry exactly 2 damage (alive %v, damage %d)", g.Battlefield.Contains(theirs), c.DamageMarked)
	}
}

func TestSocB21BarrageGetsPlusOneAfterAnotherInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Frog", "Creature — Frog", 2, 2)
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	socB21StackSpell(t, g, "Test Instant", "Instant")
	passPriorityAroundTable(t, g)
	if err := socB21CastBarrage(t, g, mine, theirs); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := socB14PT(t, g, mine); p != 3 {
		t.Errorf("power %d, want 3 after another instant", p)
	}
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) {
		c, _ := g.LookupCardForEffect(theirs)
		t.Errorf("3 damage did not kill the 3/3 Ogre (damage %d)", c.DamageMarked)
	}
}

func TestSocB21BarrageIgnoresAnEarlierNonInstantSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Frog", "Creature — Frog", 2, 2)
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if err := socB21CastBarrage(t, g, mine, theirs); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := socB14PT(t, g, mine); p != 2 {
		t.Errorf("power %d after only a creature spell, want 2", p)
	}
}

func TestSocB21BarrageWithNoSecondTargetStillPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Frog", "Creature — Frog", 2, 2)
	socB21StackSpell(t, g, "Test Sorcery", "Sorcery")
	passPriorityAroundTable(t, g)
	if err := socB21CastBarrage(t, g, mine); err != nil {
		t.Fatalf("cast with one target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := socB14PT(t, g, mine); p != 3 {
		t.Errorf("power %d, want 3: the pump does not need the second target", p)
	}
}

func TestSocB21BarrageRefusesItsOwnCreatureAsTheSecondTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Frog", "Creature — Frog", 2, 2)
	other := b12Creature(g, me.ID, "Toad", "Creature — Frog", 2, 2)
	err := socB21CastBarrage(t, g, mine, other)
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("second target one of mine: err = %v, want ErrIllegalTarget", err)
	}
}

func TestSocB21BarrageRefusesAnOpponentsCreatureAsTheFirstTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	err := socB21CastBarrage(t, g, theirs)
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("first target an opponent's: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Campus Composer // Aqueous Aria -----------------------------------------

func TestSocB21ComposerEntersPreparedAndAriaMakesAFlyingElemental(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB21ComposerOracle, nil, "Campus Composer", "Creature — Merfolk Bard", "{3}{U}", 3, 4,
		"Aqueous Aria", "Sorcery", "{4}{U}")
	castFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Campus Composer did not enter prepared")
	}
	fraBCastCopy(t, g, me, "Aqueous Aria")
	var token game.Card
	for _, b := range g.Battlefield.Cards {
		if b.IsToken() && b.Name == "Elemental" && b.Controller == me.ID {
			token = b
		}
	}
	if token.InstanceID == uuid.Nil {
		t.Fatal("no Elemental token")
	}
	if p, tg := socB14PT(t, g, token.InstanceID); p != 3 || tg != 3 {
		t.Errorf("token %d/%d, want 3/3", p, tg)
	}
	if !hasEffectiveKeyword(t, g, token.InstanceID, "flying") {
		t.Error("the token has no flying")
	}
	if len(token.Colors) != 2 {
		t.Errorf("token colours %v, want blue and red", token.Colors)
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
}

func TestSocB21ComposerWardTwoCountersAnUnpaidSpell(t *testing.T) {
	g, _, opp, composer := wardTable(t, "Campus Composer", "Creature — Merfolk Bard", socB21ComposerOracle)
	bolt := boltWardedCreature(t, g, opp, composer)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatal("no ward payment prompt for the opponent's spell")
	}
	answerPayUnless(t, g, opp.ID, false)
	passPriorityAroundTable(t, g)
	if boltResolved(g, composer) || !opp.Graveyard.Contains(bolt) {
		t.Fatal("a declined ward did not counter the spell")
	}
}

// --- Charging Strifeknight ---------------------------------------------------

func TestSocB21StrifeknightLootsByDiscardingACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	knight := b12Push(g, me.ID, "Charging Strifeknight", "Creature — Spirit Knight", socB21StrifeknightOracle, 3, 3)
	if !hasEffectiveKeyword(t, g, knight, "haste") {
		t.Fatal("no haste")
	}
	advanceTo(t, g, game.StepPrecombatMain)
	socB21SetHand(me, 2)
	pitched := me.Hand.Cards[0].InstanceID
	lib := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitched}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitched) {
		t.Error("the discarded card is not in the graveyard")
	}
	if me.Hand.Size() != 2 || me.Library.Size() != lib-1 {
		t.Errorf("hand %d, library %d -> %d: a card should have been drawn", me.Hand.Size(), lib, me.Library.Size())
	}
	if !b13Tapped(t, g, knight) {
		t.Error("the Strifeknight is not tapped")
	}
}

func TestSocB21StrifeknightRefusesWithNoCardToDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	knight := b12Push(g, me.ID, "Charging Strifeknight", "Creature — Spirit Knight", socB21StrifeknightOracle, 3, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	socB21SetHand(me, 0)
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty hand")
	}
	if b13Tapped(t, g, knight) {
		t.Error("a refused activation tapped the Strifeknight")
	}
}

// --- Chase Inspiration -------------------------------------------------------

func TestSocB21InspirationGivesPlusZeroPlusThreeAndHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Frog", "Creature — Frog", 2, 2)
	castCatalogSpell(t, g, "Chase Inspiration", "Instant", socB21InspirationOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)
	if p, tg := socB14PT(t, g, mine); p != 2 || tg != 5 {
		t.Errorf("%d/%d, want 2/5", p, tg)
	}
	if !hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("no hexproof")
	}
}

func TestSocB21InspirationTakesOnlyYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	err := castCatalogSpellErr(t, g, "Chase Inspiration", "Instant", socB21InspirationOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("an opponent's creature: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Cheerful Osteomancer // Raise Dead --------------------------------------

func TestSocB21OsteomancerEntersPreparedAndRaiseDeadReturnsACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB21OsteomancerOracle, nil, "Cheerful Osteomancer", "Creature — Orc Warlock", "{3}{B}", 4, 2,
		"Raise Dead", "Sorcery", "{B}")
	castFromHandAndResolve(t, g, me, c)
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatal("Cheerful Osteomancer did not enter prepared")
	}
	gone := pushGraveyardCardForTest(me, "Fallen Frog")
	fraBCastCopy(t, g, me, "Raise Dead", game.TargetRef{Kind: game.TargetCard, ID: gone})
	if !me.Hand.Contains(gone) {
		t.Error("Raise Dead did not return the creature card")
	}
	if g.IsPreparedForEffect(c.InstanceID) {
		t.Error("still prepared after casting the copy")
	}
}

func TestSocB21RaiseDeadRefusesANoncreatureCardAndAnOpponentsGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	c := sos1Prep(me.ID, socB21OsteomancerOracle, nil, "Cheerful Osteomancer", "Creature — Orc Warlock", "{3}{B}", 4, 2,
		"Raise Dead", "Sorcery", "{B}")
	castFromHandAndResolve(t, g, me, c)
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	theirs := pushGraveyardCardForTest(opp, "Their Frog")
	cp, ok := prepareCopyOf(g, "Raise Dead")
	if !ok {
		t.Fatal("no Raise Dead copy in exile")
	}
	for name, id := range map[string]uuid.UUID{"a sorcery card": spell, "an opponent's creature card": theirs} {
		err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{
			FromZone: string(game.ZoneExile), Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
		})
		if !errors.Is(err, game.ErrIllegalTarget) {
			t.Errorf("%s: err = %v, want ErrIllegalTarget", name, err)
		}
	}
}
