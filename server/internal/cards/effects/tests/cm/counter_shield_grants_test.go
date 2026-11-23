package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cgCounter tries to counter a spell the way every catalog counterspell
// does, and reports whether it is still on the stack afterwards.
func cgCounter(t *testing.T, g *game.Game, spell uuid.UUID) (survived bool) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(spell); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	return g.Stack.Contains(spell)
}

func TestInsistPromisesTheNextCreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Insist", "Sorcery", effectstest.CgInsistOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Insist draws a card: hand %d → %d", hand, me.Hand.Size())
	}
	if got := effectstest.CgShields(g, me); len(got) != 1 || !got[0].NextOnly || got[0].SourceName != "Insist" {
		t.Fatalf("the unspent promise is on the caster's panel: %+v", got)
	}

	sorcery := effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	if cgCounter(t, g, sorcery) {
		t.Error("a sorcery doesn't spend a creature promise, and can be countered")
	}
	bear := effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	if len(effectstest.CgShields(g, me)) != 0 {
		t.Error("the first creature spell spends the promise")
	}
	if !cgCounter(t, g, bear) {
		t.Error("the creature spell that spent it can't be countered")
	}
	effectstest.PassPriorityAroundTable(t, g)
	second := effectstest.CastCatalogSpell(t, g, "Second Bear", "Creature — Bear", "", nil)
	if cgCounter(t, g, second) {
		t.Error("the promise is spent once")
	}
}

func TestOvermasterPromisesTheNextInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Overmaster", "Sorcery", effectstest.CgOvermasterOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Overmaster draws a card: hand %d → %d", hand, me.Hand.Size())
	}
	bear := effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	if cgCounter(t, g, bear) {
		t.Error("a creature spell doesn't spend it")
	}
	sorcery := effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	if !cgCounter(t, g, sorcery) {
		t.Error("the next sorcery can't be countered")
	}
	if len(effectstest.CgShields(g, me)) != 0 {
		t.Error("and it spent the promise")
	}
}

func TestMistriseVillagePromisesTheNextSpellOfAnyKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	village := effectstest.PushCatalogPermanent(g, me.ID, "Mistrise Village", "Land", effectstest.CgMistriseVillageOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, village, 0, game.ActivateAbilityParams{})
	if !effectstest.B16Tapped(t, g, village) {
		t.Error("{T} is part of the cost")
	}
	if got := effectstest.CgShields(g, me); len(got) != 1 || !got[0].NextOnly {
		t.Fatalf("the promise: %+v", got)
	}
	bear := effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	if !cgCounter(t, g, bear) {
		t.Error("any kind of spell spends it")
	}
}

func TestMistriseVillageEntersTappedUnlessYouControlAMountainOrForest(t *testing.T) {
	for _, tc := range []struct {
		name, land string
		tapped     bool
	}{
		{"no Mountain or Forest", "Basic Land — Island", true},
		{"a Forest", "Basic Land — Forest", false},
		{"a Mountain", "Basic Land — Mountain", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.SeedLandOnBattlefield(g, me.ID, "Other Land", tc.land)
			village := effectstest.PlayLandFromHand(t, g, "Mistrise Village", effectstest.CgMistriseVillageOracle)
			if got := effectstest.B16Tapped(t, g, village); got != tc.tapped {
				t.Errorf("tapped = %v, want %v", got, tc.tapped)
			}
		})
	}
}

func TestVexingShusherMarksTargetSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shusher := effectstest.PushCatalogPermanent(g, me.ID, "Vexing Shusher", "Creature — Goblin Shaman", effectstest.CgVexingShusherOracle, false)
	sorcery := effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, shusher, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(sorcery)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.CgResolveTop(t, g)
	marks := g.StackMeta[sorcery].CantBeCountered
	if len(marks) != 1 || marks[0].SourceName != "Vexing Shusher" || marks[0].Source != shusher {
		t.Fatalf("the mark: %+v", marks)
	}
	if !g.SpellCantBeCounteredForEffect(sorcery) {
		t.Error("the stack chip reads the mark")
	}
	if !cgCounter(t, g, sorcery) {
		t.Error("the marked spell can't be countered")
	}
}

// Vexing Shusher's own rider is the printed one, and its ability can
// mark an opponent's spell too.
func TestVexingShusherCantBeCounteredAndMarksAnyonesSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := effectstest.CastCatalogSpell(t, g, "Vexing Shusher", "Creature — Goblin Shaman", effectstest.CgVexingShusherOracle, nil)
	if !cgCounter(t, g, spell) {
		t.Error("This spell can't be countered.")
	}
	effectstest.PassPriorityAroundTable(t, g)
	theirs := effectstest.CsPushSpell(g, effectstest.CsSpell{Card: effectstest.CsBlueInstant, Caster: 1, Controller: 1})
	if err := g.ActivateCatalogAbility(me.ID, spell, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.CgResolveTop(t, g)
	if !cgCounter(t, g, theirs) {
		t.Errorf("an opponent's (%s) spell can be marked too", opp.Name)
	}
}

func TestVeilOfSummerShieldsYourSpellsThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	waiting := effectstest.CsPushSpell(g, effectstest.CsSpell{Card: effectstest.CsSorcery})
	effectstest.CastCatalogSpell(t, g, "Veil of Summer", "Instant", effectstest.CgVeilOfSummerOracle, nil)
	hand := me.Hand.Size()
	effectstest.CgResolveTop(t, g)
	if me.Hand.Size() != hand {
		t.Error("no opponent has cast a blue or black spell, so no card")
	}
	if !cgCounter(t, g, waiting) {
		t.Error("a spell already on the stack is covered")
	}
	theirs := effectstest.CsPushSpell(g, effectstest.CsSpell{Card: effectstest.CsBlueInstant, Caster: 1, Controller: 1})
	if cgCounter(t, g, theirs) {
		t.Error("an opponent's spell isn't")
	}
	effectstest.PassPriorityAroundTable(t, g)
	later := effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	if !cgCounter(t, g, later) {
		t.Error("a spell cast later this turn is covered (CR 611.2c)")
	}
	if got := effectstest.CgShields(g, me); len(got) != 1 || got[0].NextOnly || got[0].SourceName != "Veil of Summer" {
		t.Errorf("the grant on the panel: %+v", got)
	}
}

// The draw reads the turn's casts: an opponent's blue or black spell,
// not a green one, and not your own blue one.
func TestVeilOfSummerDrawsAfterAnOpponentsBlueOrBlackSpell(t *testing.T) {
	for _, tc := range []struct {
		name  string
		color string
		seat  int
		draws bool
	}{
		{"an opponent's black spell", "B", 1, true},
		{"an opponent's blue spell", "U", 2, true},
		{"an opponent's green spell", "G", 1, false},
		{"your own blue spell", "U", 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			caster := g.Seats[tc.seat]
			id := uuid.New()
			caster.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell", TypeLine: "Instant",
				ManaCost: "{" + tc.color + "}", Colors: []string{tc.color}, Owner: caster.ID, Controller: caster.ID})
			g.WithWriteLock(func() {
				g.EmitEvent(game.Event{Kind: game.EventCast, Actor: caster.ID, CardID: id, Source: id})
			})
			effectstest.CastCatalogSpell(t, g, "Veil of Summer", "Instant", effectstest.CgVeilOfSummerOracle, nil)
			hand := me.Hand.Size()
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Hand.Size() == hand+1; got != tc.draws {
				t.Errorf("drew = %v, want %v", got, tc.draws)
			}
		})
	}
}

func TestDomriPlusOneShieldsCreatureSpellsYouCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	domri := effectstest.PushCatalogPermanent(g, me.ID, "Domri, Anarch of Bolas", "Legendary Planeswalker — Domri", effectstest.CgDomriOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(domri, game.CounterLoyalty, 3) })
	bear := effectstest.SeedBear(g, me.ID)
	if got := effectstest.B39CurrentPower(t, g, bear); got != 3 {
		t.Errorf("creatures you control get +1/+0: power %d", got)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, domri, 0, game.ActivateAbilityParams{})
	var pick *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
			pick = c
		}
	}
	if pick == nil || len(pick.ColorOptions) != 2 {
		t.Fatalf("Add {R} or {G}: %+v", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := effectstest.CounterOn(g, domri, game.CounterLoyalty); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
	creature := effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	if !cgCounter(t, g, creature) {
		t.Error("a creature spell you cast this turn can't be countered")
	}
	effectstest.PassPriorityAroundTable(t, g)
	sorcery := effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	if cgCounter(t, g, sorcery) {
		t.Error("a noncreature spell can")
	}
	another := effectstest.CastCatalogSpell(t, g, "Another Bear", "Creature — Bear", "", nil)
	if !cgCounter(t, g, another) {
		t.Error("it is not a one-use promise: every creature spell this turn")
	}
}

func TestDomriMinusTwoFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	domri := effectstest.PushCatalogPermanent(g, me.ID, "Domri, Anarch of Bolas", "Legendary Planeswalker — Domri", effectstest.CgDomriOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(domri, game.CounterLoyalty, 3) })
	mine := effectstest.B39Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B39Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 4)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, domri, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs, mine)}); err == nil {
		t.Error("the clauses are in order: a creature you control first")
	}
	effectstest.B16Activate(t, g, me.ID, domri, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mine, theirs)})
	// My Bear is 3/2 under Domri's anthem: it deals 3 to the Ox and takes 3.
	if g.Battlefield.Contains(mine) {
		t.Error("the Ox deals 3 to my 3/2")
	}
	ox, ok := effectstest.BattlefieldCard(g, theirs)
	if !ok || ox.DamageMarked != 3 {
		t.Errorf("my Bear deals 3 (2 + Domri's +1) to the Ox: %+v", ox.DamageMarked)
	}
	if got := effectstest.CounterOn(g, domri, game.CounterLoyalty); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

func cgBoundDeterminedCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   effectstest.CgBoundDeterminedOracle,
		Layout:     game.LayoutSplit,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Bound", TypeLine: "Instant", ManaCost: "{3}{B}{G}",
				OracleText: "Sacrifice a creature. Return up to X cards from your graveyard to your hand, where X is the number of colors that creature was. Exile this card."},
			{Name: "Determined", TypeLine: "Instant", ManaCost: "{G}{U}",
				OracleText: "Other spells you control can't be countered this turn.\nDraw a card."},
		},
	}
	c.SettleImported()
	return c
}

func TestDeterminedShieldsYourOtherSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	waiting := effectstest.CsPushSpell(g, effectstest.CsSpell{Card: effectstest.CsSorcery})
	c := cgBoundDeterminedCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Determined: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.CgResolveTop(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("Determined draws a card: %d → %d", hand, me.Hand.Size())
	}
	if !cgCounter(t, g, waiting) {
		t.Error("another spell you control can't be countered")
	}
	got := effectstest.CgShields(g, me)
	if len(got) != 1 || got[0].SourceName != "Determined" {
		t.Fatalf("the grant: %+v", got)
	}
	// Determined itself is excepted by object; cast again from the
	// graveyard it would be a new one. The except names the card.
	var except game.ObjectRef
	for _, s := range me.Statics {
		except = s.CantBeCountered.Except
	}
	if except.ID != c.InstanceID {
		t.Errorf("the except names Determined: %+v", except)
	}
}

func TestBoundSacrificesReturnsByColorsAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	gold := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Two-Color Bear",
		TypeLine: "Creature — Bear", ManaCost: "{B}{G}", Colors: []string{"B", "G"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	var yard []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Old Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		yard = append(yard, id)
	}
	c := cgBoundDeterminedCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Bound: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, gold)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no return prompt")
	}
	if pick.ChooseMax != 2 {
		t.Errorf("X is the two colours the creature was: max %d", pick.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, gold, yard[0])
	if !me.Hand.Contains(gold) || !me.Hand.Contains(yard[0]) {
		t.Error("the picks return to hand, the sacrificed creature among them")
	}
	if me.Graveyard.Contains(c.InstanceID) || !g.Exile.Contains(c.InstanceID) {
		t.Error("Bound exiles itself")
	}
}
