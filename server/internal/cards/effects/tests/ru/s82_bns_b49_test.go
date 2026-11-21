package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b49_test.go — S82 bonus-sheet pass, batch 49 (Foundations
// Jumpstart): eleven cards through the real catalog.

const (
	bns49SubtleStrike  = "08122109-6287-4d78-9242-2da8a25022b0"
	bns49SupremeWill   = "75325029-e6f1-44af-88b2-53cc2b7233c0"
	bns49Tenacity      = "7dfe701f-6572-4733-9da5-09397600bba7"
	bns49Ulcerate      = "4a9e73b7-8c8b-4489-b670-ce49c2c8a46a"
	bns49UncagedFury   = "7aac51ef-ae03-4a9e-9a48-46223787f291"
	bns49Chill         = "d6596336-8177-4475-95b9-a21858563276"
	bns49VampiricFury  = "1e23457f-1ed9-4597-9f73-6f5b7206f5a1"
	bns49Outburst      = "c0f69063-9079-4e7a-98dc-c29da8171d33"
	bns49Whirlwind     = "ba9e129d-a3cf-431c-93af-f2e4902d40a0"
	bns49CaveTemption  = "75540897-53f6-433b-bd70-9851551df6ef"
	bns49KabiraCrossrd = "b3dbb16f-fa8f-4406-bcf3-e647e4337619"
)

func bns49PT(g *game.Game, id uuid.UUID) (p, tough int) {
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, _ := g.LookupCardForEffect(id)
		p, tough = c.PowerForComparison(), c.CurrentToughness()
	})
	return
}

func TestBns49Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		bns49SubtleStrike: "Subtle Strike", bns49SupremeWill: "Supreme Will",
		bns49Tenacity: "Tenacity", bns49Ulcerate: "Ulcerate",
		bns49UncagedFury: "Uncaged Fury", bns49Chill: "Uncomfortable Chill",
		bns49VampiricFury: "Vampiric Fury", bns49Outburst: "Violent Outburst",
		bns49Whirlwind: "Whirlwind Denial", bns49CaveTemption: "Cave of Temptation",
		bns49KabiraCrossrd: "Kabira Crossroads",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestBns49SubtleStrikeBothModesNameTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	foe := effectstest.BlbB1Push(g, opp.ID, "Foe", "Creature — Bear", "", 3, 3)
	mine := effectstest.BlbB1Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(me, "Subtle Strike", "Instant", "{1}{B}", bns49SubtleStrike, []string{"B"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}, Targets: effectstest.CardTargets(foe, mine)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := bns49PT(g, foe); p != 2 || th != 2 {
		t.Errorf("foe %d/%d, want 2/2", p, th)
	}
	if got := effectstest.CountersOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on mine, want 1", got)
	}
}

func TestBns49SubtleStrikeOneModeOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.BlbB1Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(me, "Subtle Strike", "Instant", "{1}{B}", bns49SubtleStrike, []string{"B"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardTargets(mine)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := bns49PT(g, mine); p != 3 || th != 3 {
		t.Errorf("mine %d/%d, want 3/3", p, th)
	}
}

func TestBns49UlcerateShrinksAndCostsThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	foe := effectstest.BlbB1Push(g, opp.ID, "Foe", "Creature — Bear", "", 4, 4)
	before := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Ulcerate", "Instant", bns49Ulcerate, effectstest.CardTarget(foe))
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := bns49PT(g, foe); p != 1 || th != 1 {
		t.Errorf("foe %d/%d, want 1/1", p, th)
	}
	if got := effectstest.LifeOf(g, me.ID); got != before-3 {
		t.Errorf("life %d -> %d, want -3", before, got)
	}
}

func TestBns49UlcerateRefusesANonCreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	land := effectstest.BlbB1Push(g, me.ID, "Land", "Land", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(me, "Ulcerate", "Instant", "{B}", bns49Ulcerate, []string{"B"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(land)}); err == nil {
		t.Error("Ulcerate accepted a land")
	}
}

func TestBns49TenacityPumpsLifelinksAndUntapsYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	tapped := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID, Tapped: true,
	})
	effectstest.CastCatalogSpell(t, g, "Tenacity", "Instant", bns49Tenacity, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := bns49PT(g, tapped); p != 3 || th != 3 {
		t.Errorf("mine %d/%d, want 3/3", p, th)
	}
	if !effectstest.P1g2HasKeyword(g, tapped, "lifelink") {
		t.Error("lifelink was not granted")
	}
	if effectstest.P1g2HasKeyword(g, theirs, "lifelink") {
		t.Error("an opponent's creature gained lifelink")
	}
	var mineTapped, theirsTapped bool
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(tapped)
		mineTapped = c.Tapped
		c, _ = g.LookupCardForEffect(theirs)
		theirsTapped = c.Tapped
	})
	if mineTapped {
		t.Error("my creature stayed tapped")
	}
	if !theirsTapped {
		t.Error("an opponent's creature was untapped")
	}
}

func TestBns49UncagedFuryGrantsDoubleStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.BlbB1Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Uncaged Fury", "Instant", bns49UncagedFury, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := bns49PT(g, mine); p != 3 || th != 3 {
		t.Errorf("%d/%d, want 3/3", p, th)
	}
	if !effectstest.P1g2HasKeyword(g, mine, "double strike") {
		t.Error("no double strike")
	}
}

func TestBns49UncomfortableChillShrinksOpponentsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	mine := effectstest.BlbB1Push(g, me.ID, "Mine", "Creature — Bear", "", 3, 3)
	a := effectstest.BlbB1Push(g, opp.ID, "A", "Creature — Bear", "", 4, 4)
	b := effectstest.BlbB1Push(g, other.ID, "B", "Creature — Bear", "", 3, 3)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Uncomfortable Chill", "Instant", bns49Chill, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := bns49PT(g, a); p != 2 {
		t.Errorf("A power %d, want 2", p)
	}
	if p, _ := bns49PT(g, b); p != 1 {
		t.Errorf("B power %d, want 1", p)
	}
	if p, _ := bns49PT(g, mine); p != 3 {
		t.Errorf("my power %d, want 3", p)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1 (the spell leaves, a card is drawn)", hand, me.Hand.Size())
	}
}

func TestBns49VampiricFuryOnlyVampires(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	vamp := effectstest.BlbB1Push(g, me.ID, "Vamp", "Creature — Vampire", "", 2, 2)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.BlbB1Push(g, opp.ID, "TheirVamp", "Creature — Vampire", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Vampiric Fury", "Instant", bns49VampiricFury, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := bns49PT(g, vamp); p != 4 || !effectstest.P1g2HasKeyword(g, vamp, "first strike") {
		t.Errorf("my Vampire: power %d, first strike %v", p, effectstest.P1g2HasKeyword(g, vamp, "first strike"))
	}
	if p, _ := bns49PT(g, bear); p != 2 || effectstest.P1g2HasKeyword(g, bear, "first strike") {
		t.Error("a non-Vampire was affected")
	}
	if p, _ := bns49PT(g, theirs); p != 2 || effectstest.P1g2HasKeyword(g, theirs, "first strike") {
		t.Error("an opponent's Vampire was affected")
	}
}

func TestBns49ViolentOutburstCascadesAndPumps(t *testing.T) {
	spec, _ := effects.Lookup(bns49Outburst)
	if len(spec.Triggered) != 1 || spec.Triggered[0].Keyword != effects.KeywordCascade {
		t.Fatalf("want exactly one cascade trigger, got %+v", spec.Triggered)
	}
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.BlbB1Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Violent Outburst", "Instant", bns49Outburst, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := bns49PT(g, mine); p != 3 {
		t.Errorf("mine power %d, want 3", p)
	}
	if p, _ := bns49PT(g, theirs); p != 2 {
		t.Errorf("theirs power %d, want 2", p)
	}
}

// bns49WhirlwindBoard has seat 0 cast a sorcery and seat 1 answer it with
// Whirlwind Denial, stopping at the tax prompt.
func bns49WhirlwindBoard(t *testing.T) (g *game.Game, victim uuid.UUID, caster *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	caster, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim = effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	denial := effectstest.HandCardFull(them, "Whirlwind Denial", "Instant", "", bns49Whirlwind, []string{"U"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if err := g.CastSpell(them.ID, denial, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting Whirlwind Denial: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, caster.ID)
	return g, victim, caster
}

func TestBns49WhirlwindDenialCountersUnlessFourPaid(t *testing.T) {
	g, victim, caster := bns49WhirlwindBoard(t)
	prompt := effectstest.AnswerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{4}" {
		t.Errorf("the tax is %q, want {4}", prompt.PayCost)
	}
	if g.Stack.Contains(victim) {
		t.Error("a declined tax must counter the spell")
	}
}

func TestBns49WhirlwindDenialLetsAPaidSpellResolve(t *testing.T) {
	g, victim, caster := bns49WhirlwindBoard(t)
	for i := 0; i < 4; i++ {
		caster.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	effectstest.AnswerPayUnless(t, g, caster.ID, true)
	if !g.Stack.Contains(victim) {
		t.Error("a paid tax countered the spell anyway")
	}
}

func TestBns49SupremeWillCountersUnlessThreePaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	will := effectstest.HandCardFull(them, "Supreme Will", "Instant", "", bns49SupremeWill, []string{"U"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if err := g.CastSpell(them.ID, will, game.CastSpellParams{Modes: []int{0}, Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("casting Supreme Will: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, caster.ID)
	prompt := effectstest.AnswerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{3}" {
		t.Errorf("the tax is %q, want {3}", prompt.PayCost)
	}
	if g.Stack.Contains(victim) {
		t.Error("a declined tax must counter the spell")
	}
}

func TestBns49SupremeWillLooksAtFourTakesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	seeded := effectstest.SeedLibrary(me, "L1", "L2", "L3", "L4", "Filler")
	libraryBefore := me.Library.Size()
	id := effectstest.HandCardFull(me, "Supreme Will", "Instant", "", bns49SupremeWill, []string{"U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if choice == nil {
		t.Fatal("no choose-cards prompt")
	}
	if len(choice.ChooseCards) != 4 {
		t.Errorf("offered %d cards, want the top four", len(choice.ChooseCards))
	}
	if err := g.ResolveChooseCards(choice.ID, me.ID, []uuid.UUID{seeded[0]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(seeded[0]) {
		t.Error("the chosen card did not reach the hand")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || len(order.ScryCards) != 3 || order.LibraryPlacement != game.LibraryPlaceBottom {
		t.Fatalf("want a bottom-order prompt over three cards, got %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{seeded[3], seeded[2], seeded[1]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := me.Library.Size(); got != libraryBefore-1 {
		t.Errorf("library %d -> %d, want -1", libraryBefore, got)
	}
}

func TestBns49CaveOfTemptationSacrificesForTwoCountersAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cave := effectstest.PushCatalogPermanent(g, me.ID, "Cave of Temptation", "Land", bns49CaveTemption, false)
	target := effectstest.PushBattlefieldCreature(g, me.ID)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, cave, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.CardRefs(target),
	}); err == nil {
		t.Fatal("the sacrifice was legal outside a main phase")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	cave2 := effectstest.PushCatalogPermanent(g2, me2.ID, "Cave of Temptation", "Land", bns49CaveTemption, false)
	target2 := effectstest.PushBattlefieldCreature(g2, me2.ID)
	effectstest.AdvanceToMain(t, g2)
	if err := g2.AddManaForEffect(me2.ID, uuid.Nil, "{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g2.ActivateCatalogAbility(me2.ID, cave2, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(target2),
	}); err != nil {
		t.Fatalf("{4}, {T}, Sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !me2.Graveyard.Contains(cave2) {
		t.Error("the Cave was not sacrificed")
	}
	if got := effectstest.CountersOn(g2, target2, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2", got)
	}
}

func TestBns49KabiraCrossroadsEntersTappedGainsTwoAndTapsForWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Life
	id := effectstest.PlayLandFromHand(t, g, "Kabira Crossroads", bns49KabiraCrossrd)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Kabira Crossroads")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d -> %d, want +2", before, me.Life)
	}
	spec, _ := effects.Lookup(bns49KabiraCrossrd)
	if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{W}" {
		t.Errorf("mana abilities %+v, want one producing {W}", spec.ManaAbilities)
	}
}
