package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// War of the Spark, batch 170 (#5520).

const (
	oldB170BondInsight    = "3a877718-ff90-4d6e-9f95-47b06adad7f8"
	oldB170BondRevival    = "62cca25f-3bf7-4b67-97d5-280bf2544722"
	oldB170SpellkeeperW   = "91e09399-3f09-4323-8533-51536ecb7869"
	oldB170AidFallen      = "acb0084d-3b09-4d2a-a4c1-848c92588845"
	oldB170Planewide      = "af23d348-adc9-451e-af08-bf9b1986c591"
	oldB170PriceBetrayal  = "43a2bfe4-9705-4d0f-901a-21f9803b50f1"
	oldB170AngrathRampage = "850a8369-94ab-43b6-8cfe-286e3623b19d"
	oldB170DeliverEvil    = "9e73f938-0aa4-4639-a97a-f7e68f645623"
	oldB170GideonTriumph  = "daec03b6-e3f8-4e44-9766-c64dc23fa3dc"
)

func TestOldB170CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		oldB170BondInsight: "Bond of Insight", oldB170BondRevival: "Bond of Revival", oldB170SpellkeeperW: "Spellkeeper Weird",
		oldB170AidFallen: "Aid the Fallen", oldB170Planewide: "Planewide Celebration", oldB170PriceBetrayal: "Price of Betrayal",
		oldB170AngrathRampage: "Angrath's Rampage", oldB170DeliverEvil: "Deliver Unto Evil", oldB170GideonTriumph: "Gideon's Triumph",
	}
	for oracle, name := range want {
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

// oldB170Pick answers the open card pick of `who` with `want`.
func oldB170Pick(t *testing.T, g *game.Game, who uuid.UUID, want ...uuid.UUID) []uuid.UUID {
	t.Helper()
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, who)
	if p == nil {
		t.Fatal("no card pick is open")
	}
	shown := append([]uuid.UUID(nil), p.ChooseCards...)
	if err := g.ResolveChooseCards(p.ID, who, want); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	return shown
}

func oldB170Has(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

// oldB170Walker puts a planeswalker with loyalty onto the battlefield, so the
// state-based actions leave it alone.
func oldB170Walker(g *game.Game, owner uuid.UUID, name, subtype string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Planeswalker — " + subtype,
		Counters: map[string]int{"loyalty": 5}, Owner: owner, Controller: owner,
	})
}

func oldB170Library(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

// --- Bond of Insight ----------------------------------------------------------

func TestOldB170BondOfInsightMillsEveryoneThenReturnsTwoSpellsAndExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range g.Seats {
		oldB170Library(p, 8)
	}
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Library.Size()
	}
	a := effectstest.PushGraveyardCardTyped(me, "Old Bolt", "Instant")
	b := effectstest.PushGraveyardCardTyped(me, "Old Wrath", "Sorcery")
	c := effectstest.PushGraveyardCardTyped(me, "Old Third", "Sorcery")
	bear := effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	spell := effectstest.CastCatalogSpell(t, g, "Bond of Insight", "Sorcery", oldB170BondInsight, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Library.Size() != before[p.ID]-4 {
			t.Errorf("library %d -> %d, want -4", before[p.ID], p.Library.Size())
		}
	}
	shown := oldB170Pick(t, g, me.ID, a, b)
	if oldB170Has(shown, bear) {
		t.Error("a creature card was offered")
	}
	if oldB170Has(shown, spell) {
		t.Error("the resolving spell was offered")
	}
	if !oldB170Has(shown, c) {
		t.Error("a third spell card was not offered")
	}
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) || me.Hand.Contains(c) {
		t.Errorf("hand has a=%v b=%v c=%v, want the two picked", me.Hand.Contains(a), me.Hand.Contains(b), me.Hand.Contains(c))
	}
	if !effectstest.InExile(g, spell) {
		t.Error("Bond of Insight was not exiled")
	}
}

func TestOldB170BondOfInsightMayReturnNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range g.Seats {
		oldB170Library(p, 8)
	}
	a := effectstest.PushGraveyardCardTyped(me, "Old Bolt", "Instant")
	effectstest.CastCatalogSpell(t, g, "Bond of Insight", "Sorcery", oldB170BondInsight, nil)
	effectstest.PassPriorityAroundTable(t, g)
	oldB170Pick(t, g, me.ID)
	if me.Hand.Contains(a) {
		t.Error("a card returned though none was chosen")
	}
}

// --- Bond of Revival ----------------------------------------------------------

func TestOldB170BondOfRevivalReturnsACreatureWithHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	rock := effectstest.PushGraveyardCardTyped(me, "Old Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Bond of Revival", "Sorcery", oldB170BondRevival, effectstest.CardTarget(rock)); err == nil {
		t.Fatal("a noncreature card was targeted")
	}
	effectstest.CastCatalogSpell(t, g, "Bond of Revival", "Sorcery", oldB170BondRevival, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature did not return to the battlefield")
	}
	if !effectstest.HasKeywordOf(g, bear, "haste") {
		t.Error("the creature has no haste")
	}
}

// --- Spellkeeper Weird --------------------------------------------------------

func TestOldB170SpellkeeperWeirdSacrificesToReturnASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	weird := effectstest.SeedPermanentWithOracle(g, me.ID, "Spellkeeper Weird", "Creature — Weird", oldB170SpellkeeperW)
	bolt := effectstest.PushGraveyardCardTyped(me, "Old Bolt", "Instant")
	bear := effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, weird, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("a creature card was accepted; the target is an instant or sorcery card")
	}
	effectstest.B16Activate(t, g, me.ID, weird, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bolt)})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant did not return")
	}
	if g.Battlefield.Contains(weird) {
		t.Error("the Weird was not sacrificed")
	}
}

// --- Aid the Fallen -----------------------------------------------------------

func TestOldB170AidTheFallenReturnsEitherOrBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	pw := effectstest.PushGraveyardCardTyped(me, "Old Walker", "Legendary Planeswalker — Test")
	effectstest.CastModal(t, g, "Aid the Fallen", "Sorcery", oldB170AidFallen, []int{0, 1}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, bear, 0, 0), effectstest.ModeRef(game.TargetCard, pw, 1, 0),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) || !me.Hand.Contains(pw) {
		t.Errorf("hand has bear=%v walker=%v, want both", me.Hand.Contains(bear), me.Hand.Contains(pw))
	}
}

func TestOldB170AidTheFallenOneModeLeavesTheOtherCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	pw := effectstest.PushGraveyardCardTyped(me, "Old Walker", "Legendary Planeswalker — Test")
	effectstest.CastModal(t, g, "Aid the Fallen", "Sorcery", oldB170AidFallen, []int{1}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, pw, 0, 0),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(pw) || me.Hand.Contains(bear) {
		t.Errorf("hand has walker=%v bear=%v, want only the walker", me.Hand.Contains(pw), me.Hand.Contains(bear))
	}
	if _, err := effectstest.CastModalErr(t, g, "Aid the Fallen", "Sorcery", oldB170AidFallen, []int{0}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, pw, 0, 0),
	}); err == nil {
		t.Error("a planeswalker card was accepted for the creature mode")
	}
}

// --- Planewide Celebration ----------------------------------------------------

func TestOldB170PlanewideCelebrationRunsFourModesAndMayRepeat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushGraveyardCardTyped(me, "Old Rock", "Artifact")
	life := me.Life
	effectstest.CastModal(t, g, "Planewide Celebration", "Sorcery", oldB170Planewide, []int{0, 0, 1, 3}, []game.TargetRef{
		effectstest.ModeRef(game.TargetCard, rock, 2, 0),
	})
	effectstest.PassPriorityAroundTable(t, g)
	toks := effectstest.TokensNamed(g, "Citizen")
	if len(toks) != 2 {
		t.Fatalf("%d Citizens, want 2", len(toks))
	}
	if len(toks[0].Effective().Colors) != 5 || toks[0].Power != 2 || toks[0].Toughness != 2 {
		t.Errorf("Citizen is %v %d/%d, want 2/2 in all five colors", toks[0].Effective().Colors, toks[0].Power, toks[0].Toughness)
	}
	if !me.Hand.Contains(rock) {
		t.Error("the permanent card did not return")
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

func TestOldB170PlanewideCelebrationNeedsExactlyFourModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	if _, err := effectstest.CastModalErr(t, g, "Planewide Celebration", "Sorcery", oldB170Planewide, []int{0, 0, 3}, nil); err == nil {
		t.Error("three modes were accepted")
	}
}

// --- Price of Betrayal --------------------------------------------------------

func TestOldB170PriceOfBetrayalRemovesUpToFiveCountersFromAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Beast", 2, 2)
	effectstest.AddCountersForTest(g, foe, game.CounterPlusOne, 7)
	land := effectstest.SeedPermanentWithOracle(g, opp.ID, "Plains", "Basic Land — Plains", uuid.NewString())
	if err := effectstest.CastCatalogSpellErr(t, g, "Price of Betrayal", "Sorcery", oldB170PriceBetrayal, effectstest.CardTarget(land)); err == nil {
		t.Fatal("a land was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Price of Betrayal", "Sorcery", oldB170PriceBetrayal, effectstest.CardTarget(foe))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, g.Seats[g.Turn.ActiveSeat].ID, 5)
	var got int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == foe {
				got = c.Counters[game.CounterPlusOne]
			}
		}
	})
	if got != 2 {
		t.Errorf("+1/+1 counters left: %d, want 2", got)
	}
}

func TestOldB170PriceOfBetrayalRemovesCountersFromAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if err := g.AddPlayerCounterByForEffect(opp.ID, opp.ID, game.CounterEnergy, 3); err != nil {
		t.Fatal(err)
	}
	effectstest.CastCatalogSpell(t, g, "Price of Betrayal", "Sorcery", oldB170PriceBetrayal, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, g.Seats[g.Turn.ActiveSeat].ID, 2)
	if got := opp.Counters[game.CounterEnergy]; got != 1 {
		t.Errorf("energy counters left: %d, want 1", got)
	}
}

func TestOldB170PriceOfBetrayalCannotTargetYourself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := effectstest.CastCatalogSpellErr(t, g, "Price of Betrayal", "Sorcery", oldB170PriceBetrayal, []game.TargetRef{effectstest.PlayerTarget(me.ID)}); err == nil {
		t.Error("the caster was a legal target; the clause names an opponent")
	}
}

// --- Angrath's Rampage --------------------------------------------------------

func TestOldB170AngrathsRampageEachModeTakesItsKind(t *testing.T) {
	for _, tc := range []struct {
		mode     int
		typeLine string
	}{{0, "Artifact"}, {1, "Creature — Bear"}, {2, "Legendary Planeswalker — Test"}} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		var want uuid.UUID
		if tc.mode == 2 {
			want = oldB170Walker(g, opp.ID, "Wanted", "Test")
		} else {
			want = effectstest.SeedPermanentWithOracle(g, opp.ID, "Wanted", tc.typeLine, uuid.NewString())
		}
		other := effectstest.SeedPermanentWithOracle(g, opp.ID, "Other", "Enchantment", uuid.NewString())
		effectstest.CastModal(t, g, "Angrath's Rampage", "Sorcery", oldB170AngrathRampage, []int{tc.mode}, []game.TargetRef{
			effectstest.ModeRef(game.TargetPlayer, opp.ID, 0, 0),
		})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerSacrifice(t, g, opp.ID, want)
		if g.Battlefield.Contains(want) {
			t.Errorf("mode %d: the %s was not sacrificed", tc.mode, tc.typeLine)
		}
		if !g.Battlefield.Contains(other) {
			t.Errorf("mode %d: the enchantment was sacrificed", tc.mode)
		}
	}
}

func TestOldB170AngrathsRampageCanTargetYouAndNothingIsLostWithoutOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.SeedPermanentWithOracle(g, me.ID, "Mine", "Creature — Bear", uuid.NewString())
	effectstest.CastModal(t, g, "Angrath's Rampage", "Sorcery", oldB170AngrathRampage, []int{2}, []game.TargetRef{
		effectstest.ModeRef(game.TargetPlayer, me.ID, 0, 0),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("a creature was sacrificed to the planeswalker mode")
	}
}

// --- Deliver Unto Evil --------------------------------------------------------

func oldB170Four(me *game.Player) []uuid.UUID {
	return []uuid.UUID{
		effectstest.PushGraveyardCardTyped(me, "One", "Creature — Bear"),
		effectstest.PushGraveyardCardTyped(me, "Two", "Artifact"),
		effectstest.PushGraveyardCardTyped(me, "Three", "Sorcery"),
		effectstest.PushGraveyardCardTyped(me, "Four", "Land"),
	}
}

func TestOldB170DeliverUntoEvilAnOpponentChoosesTwoToLeave(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cards := oldB170Four(me)
	spell := effectstest.CastCatalogSpell(t, g, "Deliver Unto Evil", "Sorcery", oldB170DeliverEvil, effectstest.CardTargets(cards...))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	shown := oldB170Pick(t, g, opp.ID, cards[0], cards[2])
	if len(shown) != 4 {
		t.Errorf("%d cards offered, want 4", len(shown))
	}
	if me.Hand.Contains(cards[0]) || me.Hand.Contains(cards[2]) {
		t.Error("a chosen card went to hand")
	}
	if !me.Hand.Contains(cards[1]) || !me.Hand.Contains(cards[3]) {
		t.Error("an unchosen card did not go to hand")
	}
	if !effectstest.InExile(g, spell) {
		t.Error("Deliver Unto Evil was not exiled")
	}
}

func TestOldB170DeliverUntoEvilWithABolasReturnsEverything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB170Walker(g, me.ID, "Nicol Bolas, Dragon-God", "Bolas")
	cards := oldB170Four(me)
	effectstest.CastCatalogSpell(t, g, "Deliver Unto Evil", "Sorcery", oldB170DeliverEvil, effectstest.CardTargets(cards...))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range cards {
		if !me.Hand.Contains(id) {
			t.Errorf("card %s did not return", id)
		}
	}
}

func TestOldB170DeliverUntoEvilWithOneTargetTheOpponentLeavesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	card := effectstest.PushGraveyardCardTyped(me, "Lone", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Deliver Unto Evil", "Sorcery", oldB170DeliverEvil, effectstest.CardTarget(card))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	oldB170Pick(t, g, opp.ID, card)
	if me.Hand.Contains(card) {
		t.Error("the only card was returned; the opponent chooses up to two of them to leave")
	}
}

// --- Gideon's Triumph ---------------------------------------------------------

func TestOldB170GideonsTriumphSacrificesACreatureThatAttackedOrBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	g.EmitEvent(game.Event{Kind: game.EventAttack, Actor: opp.ID, CardID: attacker, Target: me.ID})
	g.EmitEvent(game.Event{Kind: game.EventBlock, Actor: opp.ID, Source: blocker, CardID: blocker, Target: attacker})
	effectstest.CastCatalogSpell(t, g, "Gideon's Triumph", "Instant", oldB170GideonTriumph, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	opts := effectstest.SacrificeOptions(c)
	if oldB170Has(opts, idle) || !oldB170Has(opts, attacker) || !oldB170Has(opts, blocker) {
		t.Errorf("options %v: want the attacker and the blocker, not the idle creature", opts)
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, blocker)
	if g.Battlefield.Contains(blocker) || !g.Battlefield.Contains(attacker) || !g.Battlefield.Contains(idle) {
		t.Error("only the chosen blocker should have been sacrificed")
	}
}

func TestOldB170GideonsTriumphWithAGideonTakesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB170Walker(g, me.ID, "Gideon Blackblade", "Gideon")
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{a, b, c} {
		g.EmitEvent(game.Event{Kind: game.EventAttack, Actor: opp.ID, CardID: id, Target: me.ID})
	}
	effectstest.CastCatalogSpell(t, g, "Gideon's Triumph", "Instant", oldB170GideonTriumph, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, a, b)
	left := 0
	for _, id := range []uuid.UUID{a, b, c} {
		if g.Battlefield.Contains(id) {
			left++
		}
	}
	if left != 1 || !g.Battlefield.Contains(idle) {
		t.Errorf("%d attackers left, idle on battlefield %v; want 1 and true", left, g.Battlefield.Contains(idle))
	}
}

func TestOldB170GideonsTriumphSacrificesNothingWhenNoCreatureAttackedOrBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Gideon's Triumph", "Instant", oldB170GideonTriumph, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(idle) {
		t.Error("a creature that neither attacked nor blocked was sacrificed")
	}
}
