package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b003_cards_test.go — S82 OLD batch 3 (#5462, Secret Lair Drop): card
// draw, tutors and win conditions. Each test plays the real catalog entry.

const (
	oldB003Contract  = "6858702a-e563-4dcb-a50d-65406cea7c86"
	oldB003Dire      = "7545adab-b6f8-496d-8c25-416c7ec85fa2"
	oldB003Song      = "75b73eb2-dafd-43db-9e96-7e1995d9849f"
	oldB003Jin       = "76c43ac8-a878-4239-b0e3-f0254ef63e50"
	oldB003Petition  = "79acdb37-8d52-4bc8-a7be-a7e92b6166e5"
	oldB003Happily   = "7a4cc57b-9f85-446c-9ca0-98c72b9bb732"
	oldB003Irulan    = "85f20d9d-588a-4b18-8cc3-3796150630ba"
	oldB003Karametra = "984e763c-5a83-4231-8ec4-1be76846479a"
	oldB003Yidris    = "9efe8aff-9a7d-4397-b5fd-c1a0fad7c15f"
)

func TestOldB003EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB003Contract:  "Liliana's Contract",
		oldB003Dire:      "Dire Undercurrents",
		oldB003Song:      "Song of Creation",
		oldB003Jin:       "Jin Sakai, Ghost of Tsushima",
		oldB003Petition:  "Dark Petition",
		oldB003Happily:   "Happily Ever After",
		oldB003Irulan:    "Princess Irulan Corrino",
		oldB003Karametra: "Karametra, God of Harvests",
		oldB003Yidris:    "Yidris, Maelstrom Wielder",
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

func oldB003Library(p *game.Player, n int, name, typeLine string) []uuid.UUID {
	p.Library.Cards = nil
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		id := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
		ids = append(ids, id)
	}
	return ids
}

func oldB003Emit(g *game.Game, ev game.Event) {
	g.WithWriteLock(func() { g.EmitEvent(ev) })
}

// --- Liliana's Contract ------------------------------------------------------

func TestOldB003ContractDrawsFourLosesFourAndWinsWithFourDifferentDemons(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB003Library(me, 10, "Forest", "Basic Land — Forest")
	hand, life := me.Hand.Size(), me.Life
	effectstest.CastWithCost(t, g, "Liliana's Contract", "Enchantment", "{3}{B}{B}", oldB003Contract)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d net of the cast, want 4", got)
	}
	if got := life - me.Life; got != 4 {
		t.Errorf("lost %d life, want 4", got)
	}
	// Three Demons are not enough; a duplicate name does not count twice.
	for _, n := range []string{"Demon A", "Demon B", "Demon A"} {
		effectstest.B12Push(g, me.ID, n, "Creature — Demon", "", 3, 3)
	}
	oldB003Emit(g, game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if g.State != game.StateActive {
		t.Fatal("won with only two different Demon names")
	}
	effectstest.B12Push(g, me.ID, "Demon C", "Creature — Demon", "", 3, 3)
	effectstest.B12Push(g, me.ID, "Demon D", "Creature — Demon", "", 3, 3)
	oldB003Emit(g, game.Event{Kind: game.EventBeginUpkeep, Actor: g.Seats[1].ID})
	effectstest.PassPriorityAroundTable(t, g)
	if g.State != game.StateActive {
		t.Fatal("won at an opponent's upkeep")
	}
	oldB003Emit(g, game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if g.State != game.StateEnded || g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Errorf("four Demons with different names did not win: state %v outcome %+v", g.State, g.Outcome)
	}
}

// --- Happily Ever After ------------------------------------------------------

func TestOldB003HappilyEverAfterGainsAndDrawsForEachPlayerThenWins(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB003Library(me, 10, "Forest", "Basic Land — Forest")
	for _, p := range g.Seats {
		if p != me {
			oldB003Library(p, 5, "Forest", "Basic Land — Forest")
		}
	}
	hands, lives := map[uuid.UUID]int{}, map[uuid.UUID]int{}
	for _, p := range g.Seats {
		hands[p.ID], lives[p.ID] = p.Hand.Size(), p.Life
	}
	effectstest.CastWithCost(t, g, "Happily Ever After", "Enchantment", "{2}{W}", oldB003Happily)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID] + 5
		if p.Life != want || p.Hand.Size() != hands[p.ID]+1 {
			t.Errorf("%s: life %d (want %d), hand %d (want %d)", p.Name, p.Life, want, p.Hand.Size(), hands[p.ID]+1)
		}
	}
	// Five colours among permanents, six card types among permanents and the graveyard.
	board := []game.Card{
		{Name: "W", TypeLine: "Creature — Test", Colors: []string{"W"}},
		{Name: "U", TypeLine: "Artifact Creature — Test", Colors: []string{"U"}},
		{Name: "B", TypeLine: "Enchantment", Colors: []string{"B"}},
		{Name: "R", TypeLine: "Land", Colors: []string{"R"}},
		{Name: "G", TypeLine: "Planeswalker — Test", Colors: []string{"G"}},
	}
	for _, c := range board {
		c.Owner, c.Controller, c.InstanceID = me.ID, me.ID, uuid.New()
		effectstest.PushBattlefieldCardWithTimestamp(g, c)
	}
	check := func(wantWin bool, why string) {
		t.Helper()
		oldB003Emit(g, game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID})
		effectstest.PassPriorityAroundTable(t, g)
		if won := g.State == game.StateEnded && g.Outcome != nil && g.Outcome.Winner == me.ID; won != wantWin {
			t.Fatalf("%s: won = %v, want %v", why, won, wantWin)
		}
	}
	// Types on the board: creature, artifact, enchantment, land, planeswalker = 5. Not enough.
	check(false, "five card types")
	effectstest.PushGraveyardPermanent(me, "Old Bolt", "Instant", "{R}")
	// Sixth type, but life is below the starting total.
	g.WithWriteLock(func() { me.Life = g.Settings.StartingLife - 1 })
	check(false, "life below the starting total")
	g.WithWriteLock(func() { me.Life = g.Settings.StartingLife })
	check(true, "all three conditions")
}

// --- Dire Undercurrents ------------------------------------------------------

func TestOldB003DireUndercurrentsDrawsForBlueAndDiscardsForBlackCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB003Library(opp, 8, "Forest", "Basic Land — Forest")
	effectstest.PushCatalogPermanent(g, me.ID, "Dire Undercurrents", "Enchantment", oldB003Dire, false)
	enter := func(colors ...string) {
		id := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Guest", TypeLine: "Creature — Test", Colors: colors, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		oldB003Emit(g, game.Event{Kind: game.EventETB, CardID: id, Actor: me.ID})
	}
	before := opp.Hand.Size()
	enter("U")
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - before; got != 1 {
		t.Fatalf("a blue creature: the target drew %d, want 1", got)
	}
	// A black creature makes the target discard; the choice is theirs.
	g.WithWriteLock(func() {
		opp.Hand.Cards = nil
		opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	})
	enter("B")
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		effectstest.AnswerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID)
	}
	if opp.Hand.Size() != 0 {
		t.Errorf("a black creature: the target still holds %d cards, want 0", opp.Hand.Size())
	}
	// A red creature does neither.
	before = opp.Hand.Size()
	enter("R")
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != before || len(g.PendingChoices) != 0 {
		t.Errorf("a red creature triggered it: hand %d, %d prompts", opp.Hand.Size(), len(g.PendingChoices))
	}
}

// --- Song of Creation --------------------------------------------------------

func TestOldB003SongOfCreationDrawsTwoPerSpellAndDiscardsTheHandAtEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB003Library(me, 12, "Forest", "Basic Land — Forest")
	effectstest.PushCatalogPermanent(g, me.ID, "Song of Creation", "Enchantment", oldB003Song, false)
	before := me.Hand.Size()
	effectstest.CastWithCost(t, g, "Any Spell", "Sorcery", "{1}{R}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d for a cast, want 2", got)
	}
	// The end step of an opponent's turn does nothing; mine empties the hand.
	oldB003Emit(g, game.Event{Kind: game.EventBeginEndStep, Actor: g.Seats[1].ID})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() == 0 {
		t.Fatal("discarded the hand at an opponent's end step")
	}
	oldB003Emit(g, game.Event{Kind: game.EventBeginEndStep, Actor: me.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 0 {
		t.Errorf("hand = %d after the end step, want 0", me.Hand.Size())
	}
	if spec, _ := effects.Lookup(oldB003Song); spec.AdditionalLandPlays != 1 {
		t.Errorf("AdditionalLandPlays = %d, want 1", spec.AdditionalLandPlays)
	}
}

// --- Princess Irulan Corrino -------------------------------------------------

func TestOldB003IrulanPunishesAnOpponentsSecondDrawAndRewardsYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB003Library(me, 10, "Forest", "Basic Land — Forest")
	oldB003Library(opp, 10, "Forest", "Basic Land — Forest")
	effectstest.PushCatalogPermanent(g, me.ID, "Princess Irulan Corrino", "Legendary Creature — Human Noble Advisor", oldB003Irulan, false)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1) // their turn: the draw step is their first draw
	life, oppLife := me.Life, opp.Life
	effectstest.DrawN(g, opp, 1) // their second card
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLife - opp.Life; got != 0 {
		// the draw step has not drawn yet at upkeep: this is their first card
		t.Logf("first draw of the turn cost %d", got)
	}
	effectstest.DrawN(g, opp, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLife - opp.Life; got != 2 {
		t.Errorf("the opponent lost %d on their second card, want 2", got)
	}
	effectstest.DrawN(g, opp, 1) // third: nothing
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLife - opp.Life; got != 2 {
		t.Errorf("the opponent's third card cost life: %d total", got)
	}
	// My own second card each turn gains 2.
	effectstest.DrawN(g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("my first card changed my life to %d", me.Life)
	}
	effectstest.DrawN(g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("my second card: life %d, want %d", me.Life, life+2)
	}
}

func TestOldB003IrulanTapDrawsTwoForYouAndTheTargetOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB003Library(me, 10, "Forest", "Basic Land — Forest")
	oldB003Library(opp, 10, "Forest", "Basic Land — Forest")
	irulan := effectstest.PushCatalogPermanent(g, me.ID, "Princess Irulan Corrino", "Legendary Creature — Human Noble Advisor", oldB003Irulan, false)
	mine, theirs := me.Hand.Size(), opp.Hand.Size()
	// Only an opponent may be targeted.
	if err := g.ActivateCatalogAbility(me.ID, irulan, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("targeted yourself")
	}
	effectstest.Pr7Activate(t, g, me.ID, irulan, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if me.Hand.Size()-mine < 2 || opp.Hand.Size()-theirs != 2 {
		t.Errorf("drew %d and %d, want at least 2 and exactly 2", me.Hand.Size()-mine, opp.Hand.Size()-theirs)
	}
}

// --- Karametra, God of Harvests ----------------------------------------------

func TestOldB003KarametraFetchesAForestOrPlainsTappedOnACreatureCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Karametra, God of Harvests", "Legendary Enchantment Creature — God", oldB003Karametra, false)
	me.Library.Cards = nil
	plains := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: plains, Name: "Temple Plains", TypeLine: "Land — Plains Forest", Owner: me.ID, Controller: me.ID})
	// A noncreature spell does nothing.
	effectstest.CastWithCost(t, g, "A Sorcery", "Sorcery", "{1}{R}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID) != nil {
		t.Fatal("a noncreature spell triggered Karametra")
	}
	effectstest.CastWithCost(t, g, "A Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AnswerSearchByID(t, g, me.ID, plains)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, plains)
	if !ok || !c.Tapped {
		t.Errorf("the fetched land: on battlefield %v, tapped %v; want both", ok, c.Tapped)
	}
}

// --- Dark Petition -----------------------------------------------------------

func TestOldB003DarkPetitionAddsBlackManaOnlyWithSpellMastery(t *testing.T) {
	for _, mastery := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		wanted := uuid.New()
		me.Library.Cards = nil
		me.Library.PushTop(game.Card{InstanceID: wanted, Name: "Wanted", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Other", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		effectstest.PushGraveyardPermanent(me, "Bolt", "Instant", "{R}")
		if mastery {
			effectstest.PushGraveyardPermanent(me, "Wrath", "Sorcery", "{2}{W}{W}")
		} else {
			effectstest.PushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
		}
		effectstest.CastWithCost(t, g, "Dark Petition", "Sorcery", "{3}{B}{B}", oldB003Petition)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerSearchByID(t, g, me.ID, wanted)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(wanted) {
			t.Fatalf("mastery %v: the tutored card is not in hand", mastery)
		}
		black := me.ManaPool.CountColor("B")
		want := 0
		if mastery {
			want = 3
		}
		if black != want {
			t.Errorf("mastery %v: %d black mana in the pool, want %d", mastery, black, want)
		}
	}
}

// --- Jin Sakai ---------------------------------------------------------------

func TestOldB003JinSakaiDrawsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB003Library(me, 6, "Forest", "Basic Land — Forest")
	jin := effectstest.PushCatalogPermanent(g, me.ID, "Jin Sakai, Ghost of Tsushima", "Legendary Creature — Human Samurai", oldB003Jin, false)
	before := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, jin, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// oldB003JinTable puts Jin Sakai and n attackers under seat 0 and walks to
// seat 0's declare-attackers step.
func oldB003JinTable(t *testing.T, n int) (*game.Game, *game.Player, *game.Player, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Jin Sakai, Ghost of Tsushima", "Legendary Creature — Human Samurai", oldB003Jin, false)
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2))
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	return g, me, opp, ids
}

func TestOldB003JinSakaiDoesNotTriggerForTwoAttackersOnOnePlayer(t *testing.T) {
	g, me, opp, ids := oldB003JinTable(t, 2)
	for _, id := range ids {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	if effectstest.ModePickChoiceFor(g, me.ID) != nil {
		t.Error("two attackers on one player triggered Jin Sakai")
	}
}

func TestOldB003JinSakaiStandoffGivesDoubleStrikeToALoneAttacker(t *testing.T) {
	g, me, opp, ids := oldB003JinTable(t, 1)
	if err := g.DeclareAttacker(ids[0], opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.Cmdr2S43ChooseMode(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, ids[0], "double strike") {
		t.Error("Standoff did not give double strike")
	}
	if effectstest.AuraRestrictions(t, g, ids[0])&game.CantBeBlocked != 0 {
		t.Error("Standoff also made it unblockable")
	}
}

func TestOldB003JinSakaiGhostMakesALoneAttackerUnblockable(t *testing.T) {
	g, me, opp, ids := oldB003JinTable(t, 1)
	if err := g.DeclareAttacker(ids[0], opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.Cmdr2S43ChooseMode(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraRestrictions(t, g, ids[0])&game.CantBeBlocked == 0 {
		t.Error("Ghost did not make it unblockable")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, ids[0], "double strike") {
		t.Error("Ghost also gave double strike")
	}
}

func TestOldB003JinSakaiIgnoresAnAttackOnAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Jin Sakai, Ghost of Tsushima", "Legendary Creature — Human Samurai", oldB003Jin, false)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLoyalty: 3}})
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(attacker, walker); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	if effectstest.ModePickChoiceFor(g, me.ID) != nil {
		t.Error("an attack on a planeswalker triggered Jin Sakai")
	}
}

// --- Yidris, Maelstrom Wielder -----------------------------------------------

func TestOldB003YidrisGivesCascadeToSpellsFromHandForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	yidris := effectstest.PushCatalogPermanent(g, me.ID, "Yidris, Maelstrom Wielder", "Legendary Creature — Ogre Wizard", oldB003Yidris, false)
	effectstest.SeedCheapLibrary(g, me, 8)
	// Before any combat damage a spell has no cascade.
	effectstest.CastWithCost(t, g, "Big Before", "Sorcery", "{4}{R}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 0 {
		t.Fatalf("a spell cast before Yidris connected got %d cascade offers", got)
	}
	effectstest.DealCombatDamageToPlayer(g, yidris, opp.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastWithCost(t, g, "Big After", "Sorcery", "{4}{R}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Errorf("a spell cast after Yidris connected got %d cascade offers, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	// A second spell cascades too: the grant lasts all turn.
	effectstest.SeedCheapLibrary(g, me, 8)
	effectstest.CastWithCost(t, g, "Second Big", "Sorcery", "{4}{R}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Errorf("a second spell got %d cascade offers, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
}
