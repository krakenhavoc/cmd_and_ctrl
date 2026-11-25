package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b184_test.go — Game Night leftovers (S82 older sets, batch 184).

const (
	s82Old184HydrolashOracle  = "31279757-499a-4d52-aa81-76b6f1983b53"
	s82Old184RevelationOracle = "ba9cd7bc-846f-4b5c-a1dc-9763a60dafc9"
	s82Old184SoothsayerOracle = "eee801ea-e2dc-4b03-a9f9-e87566cb2ae8"
	s82Old184SeekOracle       = "3865d219-44b7-49ce-ab26-b833acb39765"
	s82Old184DjinnOracle      = "a598e6ce-21ee-49c1-989e-ee864208262f"
	s82Old184FateOracle       = "cf874feb-00fe-4e48-b57a-cf8b9855c057"
)

func TestS82Old184CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82Old184HydrolashOracle:  "Hydrolash",
		s82Old184RevelationOracle: "Nissa's Revelation",
		s82Old184SoothsayerOracle: "Thallid Soothsayer",
		s82Old184SeekOracle:       "Seek the Wilds",
		s82Old184DjinnOracle:      "Soulblade Djinn",
		s82Old184FateOracle:       "Gruesome Fate",
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

func TestS82Old184HydrolashShrinksAttackersAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B16Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3, "G")
	idle := effectstest.B16Creature(g, opp.ID, "Idle", "Creature — Bear", 3, 3, "G")
	top := effectstest.SeedLibrary(me, "Top Card")[0]
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Hydrolash", TypeLine: "Instant", OracleID: s82Old184HydrolashOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast Hydrolash: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, attacker); got != 1 {
		t.Errorf("attacker power %d, want 1", got)
	}
	if got := effectstest.EffectiveToughness(t, g, attacker); got != 3 {
		t.Errorf("attacker toughness %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, idle); got != 3 {
		t.Errorf("a non-attacking creature has power %d, want 3", got)
	}
	if !me.Hand.Contains(top) {
		t.Error("the spell draws a card")
	}
}

func TestS82Old184NissasRevelationDrawsPowerAndGainsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 6; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID})
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Nissa's Revelation", "Sorcery", s82Old184RevelationOracle, nil)
	hand++ // the spell is in hand until cast
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	if got := me.Life; got != life+4 {
		t.Errorf("life %d → %d, want +4 (the Ogre's toughness)", life, got)
	}
	if got := me.Hand.Size(); got != hand-1+3 {
		t.Errorf("hand %d, want %d: three cards drawn (the Ogre's power)", got, hand-1+3)
	}
}

func TestS82Old184NissasRevelationDoesNothingForANonCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 7; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Nissa's Revelation", "Sorcery", s82Old184RevelationOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	if me.Life != life || me.Hand.Size() != hand {
		t.Errorf("life %d → %d, hand %d → %d: a non-creature does nothing", life, me.Life, hand, me.Hand.Size())
	}
}

func TestS82Old184ThallidSoothsayerSacrificesACreatureToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	soothsayer := effectstest.B21Push(g, me.ID, "Thallid Soothsayer", "Creature — Fungus", s82Old184SoothsayerOracle, 2, 3)
	fodder := effectstest.B16Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1, "G")
	top := effectstest.SeedLibrary(me, "Top Card")[0]
	effectstest.AdvanceToMain(t, g)
	params := game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}
	if err := g.ActivateCatalogAbility(me.ID, soothsayer, 0, params); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, soothsayer, 0, params)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if !me.Hand.Contains(top) {
		t.Error("the ability draws a card")
	}
}

func TestS82Old184SeekTheWildsTakesACreatureOrLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var top [5]uuid.UUID
	cards := []struct{ name, typeLine string }{
		{"Shock", "Instant"},
		{"Bear", "Creature — Bear"},
		{"Forest", "Basic Land — Forest"},
		{"Ring", "Artifact"},
		{"Fifth", "Sorcery"},
	}
	for i := len(cards) - 1; i >= 0; i-- {
		top[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: top[i], Name: cards[i].name, TypeLine: cards[i].typeLine, Owner: me.ID, Controller: me.ID})
	}
	effectstest.CastCatalogSpell(t, g, "Seek the Wilds", "Sorcery", s82Old184SeekOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick opened")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if len(offered) != 2 || !offered[top[1]] || !offered[top[2]] || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the creature and the land, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[2]) {
		t.Fatal("the land is not in hand")
	}
	if effectstest.PutInLibraryChoiceFor(g, me.ID) == nil {
		t.Fatal("no order prompt for the rest on the bottom")
	}
}

func TestS82Old184SoulbladeDjinnPumpsOnANoncreatureSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	djinn := effectstest.B21Push(g, me.ID, "Soulblade Djinn", "Creature — Djinn", s82Old184DjinnOracle, 4, 3)
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Grizzly Bears", "Creature — Bear", "", nil); err != nil {
		t.Fatalf("creature spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("a creature spell pumped the team: power %d, want 2", got)
	}
	effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("bear power %d, want 3", got)
	}
	if got := effectstest.EffectiveToughness(t, g, djinn); got != 4 {
		t.Errorf("djinn toughness %d, want 4", got)
	}
}

func TestS82Old184GruesomeFateDrainsEachOpponentPerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	}
	effectstest.B16Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Theirs", "Creature — Bear", 2, 2, "G")
	lives := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		lives[s.ID] = s.Life
	}
	effectstest.CastCatalogSpell(t, g, "Gruesome Fate", "Sorcery", s82Old184FateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := lives[s.ID] - 3
		if s.ID == me.ID {
			want = lives[s.ID]
		}
		if s.Life != want {
			t.Errorf("seat %v life %d → %d, want %d", s.ID, lives[s.ID], s.Life, want)
		}
	}
}
