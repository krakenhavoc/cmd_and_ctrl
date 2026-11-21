package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardshu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/hu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82m09Coalition = "30ab812f-a055-4c68-ad81-9dd23e38827f"
	s82m09Xerex     = "2e150f97-5688-4559-94b8-86c33fda28a0"
	s82m09Marshal   = "9ad01fb2-2ba7-447a-a5ef-e57d46421635"
	s82m09Protocol  = "ac935fad-2fa7-4afc-9db4-3cd30ec4aabd"
	s82m09Cavalier  = "18aa69d5-9611-49fe-8aa4-7b91d5980dbc"
	s82m09Lancer    = "5adcd0cb-9305-4e77-a117-31e9dca5c96d"
	s82m09Cavalry   = "a867be6f-46ce-4a12-a827-9429bd94243f"
	s82m09Locthwain = "a9baa117-ab06-4af5-9a09-4be423d7cbbd"
	s82m09Herald    = "c9b40f24-c08e-4e05-a141-0a8a670d37c9"
	s82m09Wave      = "148145ef-a368-4caf-921d-e3809adf03de"
)

// s82m09CastKnight casts a free Human Knight from hand so that it enters
// through the real cast path and fires enters triggers.
func s82m09CastKnight(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandSpell(p, "Friend", "Creature — Human Knight", "")
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast a Knight: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestS82Mom09Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		s82m09Coalition: "Knight of the New Coalition", s82m09Xerex: "Xerex Strobe-Knight",
		s82m09Marshal: "Marshal of Zhalfir", s82m09Protocol: "Protocol Knight",
		s82m09Cavalier: "Swordsworn Cavalier", s82m09Lancer: "Zhalfirin Lancer",
		s82m09Cavalry: "Exsanguinator Cavalry", s82m09Locthwain: "Locthwain Lancer",
		s82m09Herald: "Herald of Hoofbeats", s82m09Wave: "Vodalian Wave-Knight",
		cardshu.InvasionOfBelenonOracleID: "Invasion of Belenon", cardshu.InvasionOfBelenonOracleID + "#1": "Belenon War Anthem",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%+v", name, ok, spec.Name)
		}
	}
}

func s82m09Knights(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.IsCreature() && c.HasSubtype("Knight") {
			n++
		}
	}
	return n
}

func TestS82Mom09KnightOfTheNewCoalitionMakesAKnightToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	effectstest.CastFromSeat(t, g, me, "Knight of the New Coalition", "Creature — Human Knight", s82m09Coalition, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82m09Knights(g, me.ID); got != 2 {
		t.Fatalf("Knights = %d, want 2 (the creature and its token)", got)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Knight"); n != 1 {
		t.Errorf("Knight tokens = %d, want 1", n)
	}
}

func TestS82Mom09XerexNeedsTwoSpellsThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	x := effectstest.B12Push(g, me.ID, "Xerex Strobe-Knight", "Creature — Human Knight", s82m09Xerex, 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == x {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, x, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no spells cast this turn")
	}
	effectstest.ToMainForCost(t, g)
	effectstest.CastCatalogSpell(t, g, "Bear A", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, x, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated after only one spell")
	}
	effectstest.CastCatalogSpell(t, g, "Bear B", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, x, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate after two spells: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Knight"); n != 1 {
		t.Errorf("Knight tokens = %d, want 1", n)
	}
}

func TestS82Mom09MarshalPumpsOtherKnightsAndTapsAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	m := effectstest.B12Push(g, me.ID, "Marshal of Zhalfir", "Creature — Human Knight", s82m09Marshal, 2, 2)
	k := effectstest.B12Push(g, me.ID, "Other Knight", "Creature — Human Knight", "", 2, 2)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if c := effectstest.E2Card(t, g, k); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("other Knight %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.E2Card(t, g, bear); c.CurrentPower() != 2 {
		t.Errorf("a Bear got pumped: %d", c.CurrentPower())
	}
	if c := effectstest.E2Card(t, g, m); c.CurrentPower() != 2 {
		t.Errorf("the Marshal pumped itself: %d", c.CurrentPower())
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == m {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, m, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: m}}}); err == nil {
		t.Error("targeted itself with 'another target creature'")
	}
	if err := g.ActivateCatalogAbility(me.ID, m, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedForTest(t, g, theirs) {
		t.Error("target was not tapped")
	}
}

func TestS82Mom09ProtocolKnightStunsOnlyWithAnotherKnight(t *testing.T) {
	for _, withKnight := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if withKnight {
			effectstest.B12Push(g, me.ID, "Friend", "Creature — Human Knight", "", 2, 2)
		}
		theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
		effectstest.ToMainForCost(t, g)
		effectstest.CastFromSeat(t, g, me, "Protocol Knight", "Creature — Human Knight", s82m09Protocol, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPickTargetIfOpen(t, g, theirs)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.TappedForTest(t, g, theirs) {
			t.Errorf("withKnight=%v: not tapped", withKnight)
		}
		want := 0
		if withKnight {
			want = 1
		}
		if got := effectstest.CounterCount(g, theirs, game.CounterStun); got != want {
			t.Errorf("withKnight=%v: stun counters %d, want %d", withKnight, got, want)
		}
	}
}

func TestS82Mom09SwordswornCavalierFirstStrikeOnlyAfterAnotherKnightEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := effectstest.B12Push(g, me.ID, "Swordsworn Cavalier", "Creature — Human Knight", s82m09Cavalier, 3, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, c, "first strike") {
		t.Fatal("first strike from its own entry")
	}
	effectstest.B12Push(g, g.Seats[1].ID, "Their Knight", "Creature — Human Knight", "", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, c, "first strike") {
		t.Fatal("first strike from an opponent's Knight")
	}
	effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, c, "first strike") {
		t.Fatal("first strike from a non-Knight")
	}
	s82m09CastKnight(t, g, me)
	if !effectstest.HasEffectiveKeyword(t, g, c, "first strike") {
		t.Fatal("no first strike after another Knight entered")
	}
}

func TestS82Mom09ZhalfirinLancerGrowsWhenAnotherKnightEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	l := effectstest.B12Push(g, me.ID, "Zhalfirin Lancer", "Creature — Human Knight", s82m09Lancer, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.E2Card(t, g, l).CurrentPower(); p != 3 {
		t.Fatalf("a Bear changed power to %d", p)
	}
	s82m09CastKnight(t, g, me)
	c := effectstest.E2Card(t, g, l)
	if c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !effectstest.HasEffectiveKeyword(t, g, l, "vigilance") {
		t.Errorf("after a Knight: %d/%d", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestS82Mom09HeraldGivesOtherKnightsHorsemanship(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	h := effectstest.B12Push(g, me.ID, "Herald of Hoofbeats", "Creature — Human Knight", s82m09Herald, 3, 3)
	k := effectstest.B12Push(g, me.ID, "Friend", "Creature — Human Knight", "", 2, 2)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, h, "horsemanship") || !effectstest.HasEffectiveKeyword(t, g, k, "horsemanship") {
		t.Error("Herald and the other Knight should have horsemanship")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "horsemanship") {
		t.Error("a Bear has horsemanship")
	}
}

func TestS82Mom09WaveKnightCountersOnEachOtherMerfolkOrKnightPerDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	w := effectstest.B12Push(g, me.ID, "Vodalian Wave-Knight", "Creature — Merfolk Knight", s82m09Wave, 3, 3)
	mk := effectstest.B12Push(g, me.ID, "Merfolk Knight", "Creature — Merfolk Knight", "", 1, 1)
	m := effectstest.B12Push(g, me.ID, "Fish", "Creature — Merfolk", "", 1, 1)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 1, 1)
	opp := effectstest.B12Push(g, g.Seats[1].ID, "Their Knight", "Creature — Human Knight", "", 1, 1)
	effectstest.StdB21Library(me, game.Card{Name: "A"}, game.Card{Name: "B"})
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"merfolk knight": mk, "merfolk": m} {
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("%s counters = %d, want 1", name, got)
		}
	}
	for name, id := range map[string]uuid.UUID{"self": w, "bear": bear, "opponent's knight": opp} {
		if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 0 {
			t.Errorf("%s counters = %d, want 0", name, got)
		}
	}
}

func TestS82Mom09ExsanguinatorCavalryCountersAndBloodOnAnyKnightDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Exsanguinator Cavalry", "Creature — Vampire Knight", s82m09Cavalry, 2, 3)
	k := effectstest.B12Push(g, me.ID, "Friend", "Creature — Human Knight", "", 2, 2)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 0 {
		t.Fatalf("a Bear got %d counters", got)
	}
	effectstest.DealCombatDamageToPlayer(g, k, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, k, game.CounterPlusOne); got != 1 {
		t.Errorf("the Knight got %d counters, want 1", got)
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Blood"); n != 1 {
		t.Errorf("Blood tokens = %d, want 1", n)
	}
	// Damage to a creature is not damage to a player.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: k, Target: bear, Amount: 2, Combat: true})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, k, game.CounterPlusOne); got != 1 {
		t.Errorf("damage to a creature added counters: %d", got)
	}
}

func TestS82Mom09LocthwainLancerDrainsOnNontokenKnightDeathsIncludingItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	l := effectstest.B12Push(g, me.ID, "Locthwain Lancer", "Creature — Human Knight", s82m09Locthwain, 5, 5)
	k := effectstest.B12Push(g, me.ID, "Friend", "Creature — Human Knight", "", 2, 2)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Knight", "Creature — Human Knight", "", 2, 2)
	effectstest.StdB21Library(me, game.Card{Name: "A"}, game.Card{Name: "B"}, game.Card{Name: "C"})
	life, hand := opp.Life, me.Hand.Size()
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life || me.Hand.Size() != hand {
		t.Fatalf("a non-Knight or an opponent's Knight triggered it: life %d, hand %d", opp.Life, me.Hand.Size())
	}
	effectstest.KillCreature(t, g, me.ID, k)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 || me.Hand.Size() != hand+1 {
		t.Fatalf("after a Knight died: life %d (want %d), hand %d (want %d)", opp.Life, life-1, me.Hand.Size(), hand+1)
	}
	effectstest.KillCreature(t, g, me.ID, l)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 || me.Hand.Size() != hand+2 {
		t.Errorf("the Lancer's own death: life %d (want %d), hand %d (want %d)", opp.Life, life-2, me.Hand.Size(), hand+2)
	}
}

func TestS82Mom09BelenonWarAnthemPumpsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Belenon War Anthem", "Enchantment", cardshu.InvasionOfBelenonOracleID+"#1", 0, 0)
	mine := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if c := effectstest.E2Card(t, g, mine); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("own creature %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.E2Card(t, g, theirs); c.CurrentPower() != 2 {
		t.Errorf("opponent's creature pumped: %d", c.CurrentPower())
	}
}
