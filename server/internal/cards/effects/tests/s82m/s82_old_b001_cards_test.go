package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b001_cards_test.go — S82 OLD batch 1 (#5462, Secret Lair Drop):
// card draw and library. Each test plays the real catalog entry.

const (
	oldB001Tinybones   = "008eda1f-b913-4f04-8c29-5e87c270149e"
	oldB001Gregor      = "0b64da77-ca1c-427b-b48b-b919f0fe08e9"
	oldB001Study       = "115bc79e-a109-42d7-ac58-2ee8dc0a959f"
	oldB001ShaiHulud   = "155ee70b-898c-472f-a5e7-71f14da94894"
	oldB001Accumulated = "1ca25379-d194-4670-a7bf-525ad11cf436"
	oldB001Baron       = "24e3d781-88a2-485f-b86a-3078594ba331"
	oldB001Reprocess   = "27177cb0-1670-488e-82be-4610da5a7de3"
)

func TestOldB001EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB001Tinybones:   "Tinybones, Trinket Thief",
		oldB001Gregor:      "Gregor, Shrewd Magistrate",
		oldB001Study:       "Stinging Study",
		oldB001ShaiHulud:   "Shai-Hulud",
		oldB001Accumulated: "Accumulated Knowledge",
		oldB001Baron:       "Baron Vladimir Harkonnen",
		oldB001Reprocess:   "Reprocess",
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

// oldB001Library replaces p's library with n basic lands.
func oldB001Library(p *game.Player, n int) {
	p.Library.Cards = nil
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: p.ID, Controller: p.ID})
	}
}

// --- Reprocess ---------------------------------------------------------------

func TestOldB001ReprocessSacrificesTheChosenPermanentsAndDrawsForEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB001Library(me, 6)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	wolf := effectstest.SeedCreature(g, "Wolf", me.ID)
	keep := effectstest.SeedCreature(g, "Keeper", me.ID)
	oppBear := effectstest.SeedCreature(g, "Their Bear", g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID)
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Reprocess", "Sorcery", oldB001Reprocess, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatalf("Reprocess asked nothing: %+v", g.PendingChoices)
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{bear, wolf}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(wolf) {
		t.Error("a chosen creature was not sacrificed")
	}
	if !g.Battlefield.Contains(oppBear) {
		t.Error("an opponent's creature was sacrificed")
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("an unchosen creature was sacrificed")
	}
	// The spell left the hand when cast; two cards were drawn.
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand = %d, want %d (two draws)", got, before+2)
	}
}

func TestOldB001ReprocessMayChooseNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB001Library(me, 3)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	effectstest.CastCatalogSpell(t, g, "Reprocess", "Sorcery", oldB001Reprocess, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("Reprocess asked nothing")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || me.Hand.Size() != hand {
		t.Errorf("choosing nothing sacrificed or drew: bear on battlefield %v, hand %d (was %d)", g.Battlefield.Contains(bear), me.Hand.Size(), hand)
	}
}

// --- Accumulated Knowledge ---------------------------------------------------

func TestOldB001AccumulatedKnowledgeDrawsOnePlusOnePerCopyInAnyGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB001Library(me, 8)
	effectstest.PushGraveyardCardForTest(me, "Accumulated Knowledge")
	effectstest.PushGraveyardCardForTest(opp, "Accumulated Knowledge")
	effectstest.PushGraveyardCardForTest(opp, "Something Else")
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Accumulated Knowledge", "Instant", oldB001Accumulated, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// 1 + 2 copies already in graveyards; the spell itself is still on the stack.
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("drew %d net of the cast, want 3", got)
	}
}

// --- Stinging Study ----------------------------------------------------------

func TestOldB001StingingStudyUsesTheCommandersManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB001Library(me, 8)
	g.WithWriteLock(func() {
		me.Command.PushTop(game.Card{InstanceID: uuid.New(), Name: "Test Commander", TypeLine: "Legendary Creature — Test",
			ManaCost: "{3}{G}{G}", IsCommander: true, Owner: me.ID, Controller: me.ID})
	})
	before, life := me.Hand.Size(), me.Life
	effectstest.CastCatalogSpell(t, g, "Stinging Study", "Instant", oldB001Study, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 5 {
		t.Errorf("drew %d net of the cast, want 5", got)
	}
	if got := life - me.Life; got != 5 {
		t.Errorf("lost %d life, want 5", got)
	}
}

func TestOldB001StingingStudyWithoutACommanderDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB001Library(me, 4)
	before, life := me.Hand.Size(), me.Life
	effectstest.CastCatalogSpell(t, g, "Stinging Study", "Instant", oldB001Study, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before || me.Life != life {
		t.Errorf("with no commander: hand %d (was %d), life %d (was %d)", me.Hand.Size(), before, me.Life, life)
	}
}

// --- Gregor ------------------------------------------------------------------

func TestOldB001GregorDrawsCardsEqualToItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB001Library(me, 8)
	id := effectstest.BfCardWithOracle(g, me, "Gregor, Shrewd Magistrate", "Legendary Creature — Human Advisor", oldB001Gregor)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Power = 3
			}
		}
	})
	if spec, _ := effects.Lookup(oldB001Gregor); len(spec.PrintedKeywords) == 0 || spec.PrintedKeywords[0] != "skulk" {
		t.Errorf("Gregor lost skulk: %v", spec.PrintedKeywords)
	}
	before := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, id, opp.ID, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 3 {
		t.Errorf("drew %d, want 3 (its power)", got)
	}
}

// --- Tinybones ---------------------------------------------------------------

func TestOldB001TinybonesDrawsAtEndStepOnlyIfAnOpponentDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB001Library(me, 8)
	effectstest.BfCardWithOracle(g, me, "Tinybones, Trinket Thief", "Legendary Creature — Skeleton Rogue", oldB001Tinybones)
	before, life := me.Hand.Size(), me.Life
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginEndStep, Actor: me.ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before || me.Life != life {
		t.Fatalf("triggered with no opponent discard: hand %d (was %d), life %d (was %d)", me.Hand.Size(), before, me.Life, life)
	}
	// An opponent discards.
	oppCard := game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID}
	g.WithWriteLock(func() { opp.Hand.PushTop(oppCard) })
	var err error
	g.WithWriteLock(func() { err = g.DiscardRandomForEffect(opp.ID, 1) })
	if err != nil {
		t.Fatalf("discard: %v", err)
	}
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginEndStep, Actor: me.ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 || me.Life != life-1 {
		t.Errorf("after an opponent's discard: hand %d (want %d), life %d (want %d)", me.Hand.Size(), before+1, me.Life, life-1)
	}
}

func TestOldB001TinybonesActivatedAbilityHitsOnlyEmptyHandedOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a, b := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)], g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	id := effectstest.BfCardWithOracle(g, me, "Tinybones, Trinket Thief", "Legendary Creature — Skeleton Rogue", oldB001Tinybones)
	g.WithWriteLock(func() {
		a.Hand.Cards = nil
		b.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Held", TypeLine: "Sorcery", Owner: b.ID, Controller: b.ID})
	})
	la, lb := a.Life, b.Life
	effectstest.Pr7Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	if got := la - a.Life; got != 10 {
		t.Errorf("the empty-handed opponent lost %d, want 10", got)
	}
	if b.Life != lb {
		t.Errorf("an opponent holding a card lost %d", lb-b.Life)
	}
}

// --- Baron Vladimir Harkonnen ------------------------------------------------

func TestOldB001BaronDrawsWhenAnOpponentSacrificesANontokenCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB001Library(me, 8)
	effectstest.BfCardWithOracle(g, me, "Baron Vladimir Harkonnen", "Legendary Creature — Human Noble", oldB001Baron)
	token := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	real := effectstest.SeedCreature(g, "Bear", opp.ID)
	mine := effectstest.SeedCreature(g, "Mine", me.ID)
	sacrifice := func(id uuid.UUID) {
		g.WithWriteLock(func() { g.SacrificePermanentForEffect(id) })
		effectstest.PassPriorityAroundTable(t, g)
	}
	before := me.Hand.Size()
	sacrifice(token)
	if me.Hand.Size() != before {
		t.Error("drew for a sacrificed token")
	}
	sacrifice(mine)
	if me.Hand.Size() != before {
		t.Error("drew for its own controller's sacrifice")
	}
	sacrifice(real)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand = %d, want %d after an opponent sacrificed a nontoken creature", me.Hand.Size(), before+1)
	}
}

func TestOldB001BaronEdictsANontokenCreatureAndDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB001Library(me, 8)
	baron := effectstest.BfCardWithOracle(g, me, "Baron Vladimir Harkonnen", "Legendary Creature — Human Noble", oldB001Baron)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == baron {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	fodder := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Fodder", TypeLine: "Legendary Creature — Test", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	nonLegend := effectstest.SeedCreature(g, "Plain", me.ID)
	token := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Goblin", TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	real := effectstest.SeedCreature(g, "Bear", opp.ID)
	// Only another legendary creature may be sacrificed.
	if err := g.ActivateCatalogAbility(me.ID, baron, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, SacrificeIDs: []uuid.UUID{nonLegend}}); err == nil {
		t.Error("a nonlegendary creature paid the cost")
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, baron, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, real)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(real) || !g.Battlefield.Contains(token) {
		t.Errorf("the nontoken creature should be sacrificed and the token kept: real on bf %v, token on bf %v", g.Battlefield.Contains(real), g.Battlefield.Contains(token))
	}
	if got := life - opp.Life; got != 3 {
		t.Errorf("the opponent lost %d life, want 3", got)
	}
}

// --- Shai-Hulud --------------------------------------------------------------

func TestOldB001ShaiHuludSweepsAndFetchesOneLandPerPlayerWhoLostAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)], g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	me.Library.Cards = nil
	for _, n := range []string{"Forest", "Forest", "Plains", "Island"} {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: n, TypeLine: "Basic Land — " + n, Owner: me.ID, Controller: me.ID})
	}
	desert := game.Card{InstanceID: uuid.New(), Name: "Test Desert", TypeLine: "Land — Desert", Owner: me.ID, Controller: me.ID}
	me.Library.PushTop(desert)
	rock := effectstest.PushPermanent(g, a.ID, game.Card{Name: "Rock", TypeLine: "Artifact", Owner: a.ID, Controller: a.ID})
	rock2 := effectstest.PushPermanent(g, a.ID, game.Card{Name: "Rock Two", TypeLine: "Artifact", Owner: a.ID, Controller: a.ID})
	ench := effectstest.PushPermanent(g, b.ID, game.Card{Name: "Aura", TypeLine: "Enchantment", Owner: b.ID, Controller: b.ID})
	bear := effectstest.SeedCreature(g, "Bear", a.ID)
	effectstest.BfCardWithOracle(g, me, "Shai-Hulud", "Creature — Wurm", oldB001ShaiHulud)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{rock, rock2, ench} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived the sweep", id)
		}
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was destroyed")
	}
	// Two players (a and b) lost permanents: one search for two cards.
	var picks []uuid.UUID
	for _, c := range me.Library.Cards {
		if c.Name == "Forest" || c.InstanceID == desert.InstanceID {
			picks = append(picks, c.InstanceID)
		}
		if len(picks) == 2 {
			break
		}
	}
	effectstest.AnswerSearchByID(t, g, me.ID, picks...)
	effectstest.PassPriorityAroundTable(t, g)
	lands := 0
	for _, id := range picks {
		if c, ok := effectstest.BattlefieldCard(g, id); ok {
			lands++
			if !c.Tapped {
				t.Errorf("%s entered untapped", c.Name)
			}
		}
	}
	if lands != 2 {
		t.Errorf("%d fetched lands entered, want 2", lands)
	}
	// A third card is refused: only as many as players who lost a permanent.
	for _, c := range me.Library.Cards {
		if c.Name == "Plains" && g.Battlefield.Contains(c.InstanceID) {
			t.Error("a third land entered")
		}
	}
}
