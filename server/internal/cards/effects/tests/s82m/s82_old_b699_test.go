package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b699_test.go — Mercadian Masques, "Mana and ramp, Token makers,
// Graveyard recursion" (S82 OLD batch 699, #5619): each card driven through a
// real activation, cast or trigger.

const (
	oldB699Horn     = "3b9bcf88-7304-48c5-bd46-3e37a93967a5"
	oldB699Heart    = "4c774c6e-c5a0-4018-b494-d3c521d2cac3"
	oldB699Eye      = "a1bdea9f-56d0-411a-9da8-601dd6dc6d32"
	oldB699Skull    = "b5ae2532-e642-47d1-bb5f-53f408e2fdc2"
	oldB699Tooth    = "fa4c57b3-6eaa-4938-b41a-0dad3e774d49"
	oldB699SnakePit = "115bd33d-4875-4014-ae9f-0052f5acdbb2"
	oldB699Ritual   = "389ad507-8a14-4856-8d19-c90e3104aca8"
	oldB699Ants     = "9c3c9a22-e092-4845-b661-047590746498"
	oldB699Thug     = "a4d46c1b-7742-4ac8-9794-b669ee4d2cc1"
	oldB699Cross    = "e9aac98a-6fad-463e-89fa-3ad3a7c7f53a"
	oldB699Horror   = "f7920a2c-d14a-4ed6-8341-96751322f4e7"
)

func TestS82OldB699AllAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{oldB699Horn, oldB699Heart, oldB699Eye, oldB699Skull, oldB699Tooth, oldB699SnakePit,
		oldB699Ritual, oldB699Ants, oldB699Thug, oldB699Cross, oldB699Horror} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82OldB699RamosArtifactsTapOrSacrificeForTheirColour(t *testing.T) {
	for _, tc := range []struct{ name, oracle, colour string }{
		{"Horn of Ramos", oldB699Horn, "G"}, {"Heart of Ramos", oldB699Heart, "R"}, {"Eye of Ramos", oldB699Eye, "U"},
		{"Skull of Ramos", oldB699Skull, "B"}, {"Tooth of Ramos", oldB699Tooth, "W"},
	} {
		t.Run(tc.name+" taps", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.B12Push(g, me.ID, tc.name, "Artifact", tc.oracle, 0, 0)
			effectstest.AdvanceToMain(t, g)
			effectstest.ActivateManaFor(t, g, me.ID, id, 0, game.ManaAbilityParams{})
			if got := effectstest.PoolColours(me); got[tc.colour] != 1 || me.ManaPool.Total() != 1 {
				t.Errorf("pool %v, want one {%s}", got, tc.colour)
			}
			if !g.Battlefield.Contains(id) || !effectstest.TappedForTest(t, g, id) {
				t.Error("the tap ability should tap it and keep it")
			}
		})
		t.Run(tc.name+" sacrifices", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.B12Push(g, me.ID, tc.name, "Artifact", tc.oracle, 0, 0)
			effectstest.AdvanceToMain(t, g)
			effectstest.ActivateManaFor(t, g, me.ID, id, 1, game.ManaAbilityParams{})
			if got := effectstest.PoolColours(me); got[tc.colour] != 1 || me.ManaPool.Total() != 1 {
				t.Errorf("pool %v, want one {%s}", got, tc.colour)
			}
			if g.Battlefield.Contains(id) {
				t.Error("the sacrifice ability left it on the battlefield")
			}
		})
	}
}

func TestS82OldB699SnakePitMakesASnakeWhenAnOpponentCastsBlueOrBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Snake Pit", oldB699SnakePit, "Enchantment")
	cast := func(name string, colors ...string) {
		effectstest.B09OpponentCastsCard(t, g, opp, game.Card{Name: name, TypeLine: "Instant", Colors: colors})
	}
	cast("Blue Instant", "U")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Snake")); n != 1 {
		t.Fatalf("%d Snakes after a blue spell, want 1", n)
	}
	cast("Black Instant", "B")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Snake")); n != 1 {
		t.Fatalf("%d Snakes after a declined black spell, want 1", n)
	}
	cast("Red Instant", "R")
	if effectstest.AssistCardTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a red spell asked Snake Pit's question")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Snake")); n != 1 {
		t.Errorf("%d Snakes after a red spell, want 1", n)
	}
	// Its own controller's blue spell does not trigger it.
	effectstest.AdvanceToMain(t, g)
	mine := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mine, Name: "My Blue", TypeLine: "Instant", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if effectstest.AssistCardTriggerPrompt(g, me.ID) != nil {
		t.Error("the controller's own blue spell asked the question")
	}
}

func TestS82OldB699MidnightRitualExilesXCreatureCardsForZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var bears []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		bears = append(bears, id)
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Dead Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := effectstest.CastCatalogSpellErr(t, g, "Midnight Ritual", "Sorcery", oldB699Ritual, effectstest.DgCardRef(spell)); err == nil {
		t.Error("targeted a noncreature card")
	}
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: bears[0]}, {Kind: game.TargetCard, ID: bears[1]}}
	effectstest.CastXSpell(t, g, "Midnight Ritual", "Sorcery", oldB699Ritual, "{X}{2}{B}", 2, targets)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range bears {
		if me.Graveyard.Contains(id) || !g.Exile.Contains(id) {
			t.Error("a targeted creature card was not exiled")
		}
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" && c.Controller == me.ID && c.Power == 2 && c.Toughness == 2 {
			n++
		}
	}
	if n != 2 {
		t.Errorf("%d Zombies, want 2", n)
	}
}

func TestS82OldB699StrongarmThugReturnsAMercenary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	merc, bear := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: merc, Name: "Dead Merc", TypeLine: "Creature — Human Mercenary", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.CastHoldCreatureNoResolve(t, g, "Strongarm Thug", "Creature — Human Mercenary", oldB699Thug, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.PickTargetPrompt(t, g)
	for _, id := range p.PickTargetCards {
		if id == bear {
			t.Error("a non-Mercenary card was offered")
		}
	}
	effectstest.AnswerPickTarget(t, g, merc)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(merc) || me.Graveyard.Contains(merc) {
		t.Error("the Mercenary card did not return to hand")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the Bear left the graveyard")
	}
}

func TestS82OldB699HauntedCrossroadsPutsACreatureCardOnTopOfTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	cross := effectstest.PushCatalogPermanent(g, me.ID, "Haunted Crossroads", "Enchantment", oldB699Cross, false)
	bear, spell := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Dead Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, cross, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.DgCardRef(bear)}); err == nil {
		t.Fatal("activated without paying {B}")
	}
	effectstest.FloatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, cross, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.DgCardRef(spell)}); err == nil {
		t.Fatal("targeted a noncreature card")
	}
	if err := g.ActivateCatalogAbility(me.ID, cross, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.DgCardRef(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(bear) {
		t.Error("the creature card is still in the graveyard")
	}
	if names := effectstest.LibraryTopNames(me, 1); len(names) != 1 || names[0] != "Dead Bear" {
		t.Errorf("library top %v, want Dead Bear", names)
	}
}

func TestS82OldB699SaberAntsMakeThatManyInsects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	ants := effectstest.B12Push(g, me.ID, "Saber Ants", "Creature — Insect", oldB699Ants, 2, 5)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, ants, 2) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	oldB699AnswerYesNo(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Insect")); n != 2 {
		t.Fatalf("%d Insects, want 2", n)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, ants, 1) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	oldB699AnswerYesNo(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Insect")); n != 2 {
		t.Errorf("%d Insects after declining, want 2", n)
	}
}

// oldB699AnswerYesNo answers the open "you may" prompt, whichever kind the
// engine raised (a trigger prompt, or a confirm from the effect).
func oldB699AnswerYesNo(t *testing.T, g *game.Game, who uuid.UUID, yes bool) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c == nil || c.Chooser != who {
			continue
		}
		switch c.Kind {
		case game.PendingChoiceTriggerPrompt:
			if err := g.ResolveTriggerPrompt(c.ID, who, yes); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			return
		case game.PendingChoiceConfirm:
			if err := g.ResolveConfirm(c.ID, who, yes); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			return
		}
	}
	t.Fatalf("no yes/no prompt for %s: %+v", who, g.PendingChoices)
}

func TestS82OldB699EnslavedHorrorLetsEachOtherPlayerReanimate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	push := func(p *game.Player, name string) uuid.UUID {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
		return id
	}
	mine := push(me, "My Dead")
	aDead := push(a, "A Dead")
	bDead := push(b, "B Dead")
	_ = c
	effectstest.CastHoldCreature(t, g, "Enslaved Horror", "Creature — Horror", oldB699Horror, 4, 4)
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID); p != nil {
		t.Fatal("the controller was asked")
	}
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, a.ID)
	if p == nil {
		t.Fatal("the first other player was not asked")
	}
	if err := g.ResolveChooseCards(p.ID, a.ID, []uuid.UUID{aDead}); err != nil {
		t.Fatalf("a's pick: %v", err)
	}
	p = effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, b.ID)
	if p == nil {
		t.Fatal("the second other player was not asked")
	}
	if err := g.ResolveChooseCards(p.ID, b.ID, nil); err != nil {
		t.Fatalf("b declines: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aDead) {
		t.Error("the first player's creature did not return")
	}
	if cd, ok := effectstest.BattlefieldCard(g, aDead); !ok || cd.Controller != a.ID {
		t.Errorf("the returned creature is controlled by %v, want its owner", cd.Controller)
	}
	if g.Battlefield.Contains(bDead) || !b.Graveyard.Contains(bDead) {
		t.Error("the second player declined but the creature moved")
	}
	if g.Battlefield.Contains(mine) || !me.Graveyard.Contains(mine) {
		t.Error("the controller's graveyard was used")
	}
}
