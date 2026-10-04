package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s58_pr9_cards_test.go — S58 PR 9, "Drawn To Death (2): wheels,
// tutors and the rest": ten cards through the real catalog.

const (
	s58ArchfiendOfIfnir     = "535e2af0-7b08-4026-b1bf-87626407dd42"
	s58AwakenErstwhile      = "06fa6719-aa4f-4857-8029-e42d01232645"
	s58BeholdTheBeyond      = "b2402ac0-8ab8-413c-bac4-9f4ba2163c58"
	s58CurseOfFoolsWisdom   = "93b2a685-8267-4cbd-ac55-db6ff95fe98d"
	s58DemonicCollusion     = "53cbc12d-e182-4cde-8967-9b2664daa817"
	s58DoomsdayExcruciator  = "ad0db433-a406-4ef0-8ffa-416af610c4e7"
	s58IllGottenGains       = "c88eff74-def5-42ea-873c-b3b479f6fe18"
	s58SchemingSilvertongue = "44443716-8356-40cd-a879-35264b29108c"
	s58StarvingRevenant     = "2ca969eb-3d79-4d1f-8d9d-7b8204ad166a"
	s58MasterOfLakeTown     = "89c3c81b-8960-4f79-b2ec-f13560058071"
)

func TestS58PR9Registered(t *testing.T) {
	want := map[string]string{
		s58ArchfiendOfIfnir:            "Archfiend of Ifnir",
		s58AwakenErstwhile:             "Awaken the Erstwhile",
		s58BeholdTheBeyond:             "Behold the Beyond",
		s58CurseOfFoolsWisdom:          "Curse of Fool's Wisdom",
		s58DemonicCollusion:            "Demonic Collusion",
		s58DoomsdayExcruciator:         "Doomsday Excruciator",
		s58IllGottenGains:              "Ill-Gotten Gains",
		s58SchemingSilvertongue:        "Scheming Silvertongue",
		s58StarvingRevenant:            "Starving Revenant",
		s58MasterOfLakeTown:            "The Master of Lake-town",
		s58SchemingSilvertongue + "#1": "Sign in Blood",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s: registered=%v name=%q", name, ok, spec.Name)
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want full", name, spec.Completeness)
		}
	}
}

// s58HandCards puts n filler cards in p's hand and returns their IDs.
func s58HandCards(p *game.Player, n int) []uuid.UUID {
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = uuid.New()
		p.Hand.PushTop(game.Card{InstanceID: ids[i], Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
	return ids
}

func s58Zombies(g *game.Game, controller uuid.UUID) int {
	return b27CountNamed(g, controller, "Zombie")
}

// Awaken the Erstwhile: every hand goes, and each player makes as many
// Zombies as THEY discarded, not as many as the most anyone discarded.
func TestS58AwakenTheErstwhile(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	s58HandCards(opp, 3)
	third.Hand.Cards = nil
	meBefore := me.Hand.Size()
	oppBefore := opp.Hand.Size()
	castCatalogSpell(t, g, "Awaken the Erstwhile", "Sorcery", s58AwakenErstwhile, nil)
	passPriorityAroundTable(t, g)

	for _, p := range g.Seats {
		if p.Hand.Size() != 0 {
			t.Errorf("seat %s still holds %d cards", p.Name, p.Hand.Size())
		}
	}
	if got := s58Zombies(g, me.ID); got != meBefore {
		t.Errorf("caster: %d Zombies, want %d", got, meBefore)
	}
	if got := s58Zombies(g, opp.ID); got != oppBefore {
		t.Errorf("opponent: %d Zombies, want %d", got, oppBefore)
	}
	if got := s58Zombies(g, third.ID); got != 0 {
		t.Errorf("a player with no hand made %d Zombies", got)
	}
}

// Behold the Beyond: discard the hand, then take three cards.
func TestS58BeholdTheBeyond(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	held := s58HandCards(me, 2)
	var lib []uuid.UUID
	for _, name := range []string{"Alpha", "Bravo", "Charlie", "Delta"} {
		id := uuid.New()
		lib = append(lib, id)
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	castCatalogSpell(t, g, "Behold the Beyond", "Sorcery", s58BeholdTheBeyond, nil)
	passPriorityAroundTable(t, g)

	for _, id := range held {
		if !me.Graveyard.Contains(id) {
			t.Error("a card in hand was not discarded")
		}
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, lib[:3]); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if me.Hand.Size() != 3 {
		t.Errorf("hand after the search: %d, want 3", me.Hand.Size())
	}
}

// Demonic Collusion: the buyback discards two cards and the spell
// returns to hand; without it the spell goes to the graveyard.
func TestS58DemonicCollusionBuyback(t *testing.T) {
	for _, bought := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		pile := s58HandCards(me, 2)
		tutored := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: tutored, Name: "Prize", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})

		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Demonic Collusion", TypeLine: "Sorcery", OracleID: s58DemonicCollusion, Owner: me.ID, Controller: me.ID})
		advanceToMain(t, g)
		params := game.CastSpellParams{}
		if bought {
			params.OptionalCosts = []int{0}
			params.DiscardIDs = pile
		}
		if err := g.CastSpell(me.ID, id, params); err != nil {
			t.Fatalf("CastSpell (buyback=%v): %v", bought, err)
		}
		if bought {
			for _, d := range pile {
				if !me.Graveyard.Contains(d) {
					t.Error("the buyback's discard did not happen")
				}
			}
		}
		passPriorityAroundTable(t, g)
		if c := searchChoiceFor(g, me.ID); c != nil {
			if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{tutored}); err != nil {
				t.Fatalf("ResolveSearchLibrary: %v", err)
			}
		}
		if !me.Hand.Contains(tutored) {
			t.Errorf("buyback=%v: the tutored card is not in hand", bought)
		}
		if me.Hand.Contains(id) != bought || me.Graveyard.Contains(id) == bought {
			t.Errorf("buyback=%v: Collusion hand=%v graveyard=%v", bought, me.Hand.Contains(id), me.Graveyard.Contains(id))
		}
	}
}

// Unbought, the spell cannot be cast with the discard half: only the
// two-card discard makes it a buyback.
func TestS58DemonicCollusionBuybackNeedsTwoCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	one := s58HandCards(me, 1)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Demonic Collusion", TypeLine: "Sorcery", OracleID: s58DemonicCollusion, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{OptionalCosts: []int{0}, DiscardIDs: one}); err == nil {
		t.Fatal("a buyback with one card discarded must be refused")
	}
}

// Ill-Gotten Gains exiles itself, every hand is discarded, then EACH
// player picks up to three cards from THEIR OWN graveyard.
func TestS58IllGottenGains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The other two seats hold nothing, so they have no graveyard to
	// pick from and are not asked.
	for _, p := range g.Seats {
		p.Hand.Cards = nil
	}
	mine := s58HandCards(me, 2)
	theirs := s58HandCards(opp, 4)
	old := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})

	spell := castCatalogSpell(t, g, "Ill-Gotten Gains", "Sorcery", s58IllGottenGains, nil)
	passPriorityAroundTable(t, g)

	if me.Graveyard.Contains(spell) || !g.Exile.Contains(spell) {
		t.Fatal("Ill-Gotten Gains must exile itself")
	}
	myPrompt := pendingChooseCards(t, g, me.ID)
	if len(myPrompt.ChooseCards) != 2 || myPrompt.ChooseMax != 2 || myPrompt.ChooseMin != 0 {
		t.Errorf("caster's prompt: %d candidates, bounds %d..%d", len(myPrompt.ChooseCards), myPrompt.ChooseMin, myPrompt.ChooseMax)
	}
	oppPrompt := pendingChooseCards(t, g, opp.ID)
	if len(oppPrompt.ChooseCards) != 5 {
		t.Errorf("opponent's prompt offers %d cards, want their own 5", len(oppPrompt.ChooseCards))
	}
	for _, id := range oppPrompt.ChooseCards {
		if id == mine[0] || id == mine[1] {
			t.Error("the opponent was offered the caster's graveyard")
		}
	}
	// A third player has an empty graveyard and is not asked.
	if c := chooseCardsChoiceFor(g, g.Seats[2].ID); c != nil {
		t.Error("a player with no graveyard was asked")
	}

	answerChooseCards(t, g, me.ID, mine[0])
	answerChooseCards(t, g, opp.ID, theirs[0], theirs[1], theirs[2])
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != 1 || !me.Hand.Contains(mine[0]) {
		t.Errorf("caster's hand: %d cards, want just the one picked", me.Hand.Size())
	}
	if opp.Hand.Size() != 3 {
		t.Errorf("opponent's hand: %d cards, want the three picked", opp.Hand.Size())
	}
	if opp.Graveyard.Contains(theirs[0]) || !opp.Graveyard.Contains(theirs[3]) || !opp.Graveyard.Contains(old) {
		t.Error("only the picked cards leave the opponent's graveyard")
	}
}

// Doomsday Excruciator: cast, each player keeps only the bottom six
// cards of their library, exiled face down; an Excruciator that
// arrives uncast exiles nothing. The upkeep draw works either way.
func TestS58DoomsdayExcruciator(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sizes := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		sizes[i] = p.Library.Size()
	}
	exileBefore := g.Exile.Size()
	advanceToMain(t, g)
	castFromHandForTest(t, g, me, "Doomsday Excruciator", "Creature — Demon", "", s58DoomsdayExcruciator, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Library.Size() != 6 {
			t.Errorf("seat %d keeps %d cards, want 6 (had %d)", i, p.Library.Size(), sizes[i])
		}
	}
	want := 0
	for _, s := range sizes {
		want += s - 6
	}
	if got := g.Exile.Size() - exileBefore; got != want {
		t.Errorf("exiled %d cards, want %d", got, want)
	}
	for _, c := range g.Exile.Cards {
		if c.Name == "basic-filler" && !c.FaceDown {
			t.Error("a library card was exiled face up")
			break
		}
	}
}

func TestS58DoomsdayExcruciatorUncastExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Library.Size()
	b27Push(g, me.ID, "Doomsday Excruciator", "Creature — Demon", s58DoomsdayExcruciator, "{B}{B}{B}{B}{B}{B}", 6, 6, "B")
	passPriorityAroundTable(t, g)
	if me.Library.Size() != before {
		t.Errorf("an Excruciator that was not cast exiled %d cards", before-me.Library.Size())
	}
}

func TestS58DoomsdayExcruciatorDrawsInUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Doomsday Excruciator", "Creature — Demon", s58DoomsdayExcruciator, "{B}{B}{B}{B}{B}{B}", 6, 6, "B")
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("upkeep draw: hand %d → %d", hand, me.Hand.Size())
	}
}

// Archfiend of Ifnir: a discard puts a -1/-1 counter on each of the
// opponents' creatures and none on yours; a card CYCLED is one trigger,
// not two.
func TestS58ArchfiendOfIfnir(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Archfiend of Ifnir", "Creature — Demon", s58ArchfiendOfIfnir, "{3}{B}{B}", 5, 4, "B")
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	a := b12Creature(g, opp.ID, "Theirs A", "Creature — Bear", 3, 3)
	b := b12Creature(g, opp.ID, "Theirs B", "Creature — Bear", 3, 3)
	s58HandCards(me, 2)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if counterCount(g, a, game.CounterMinusOne) != 1 || counterCount(g, b, game.CounterMinusOne) != 1 {
		t.Fatalf("opponent creatures: %d and %d counters, want 1 each",
			counterCount(g, a, game.CounterMinusOne), counterCount(g, b, game.CounterMinusOne))
	}
	if counterCount(g, mine, game.CounterMinusOne) != 0 {
		t.Error("the controller's own creature got a counter")
	}

	// An opponent's discard is not "you".
	s58HandCards(opp, 1)
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if counterCount(g, a, game.CounterMinusOne) != 1 {
		t.Error("an opponent's discard triggered the Archfiend")
	}
}

func TestS58ArchfiendOfIfnirCyclingIsOneTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Archfiend of Ifnir", "Creature — Demon", s58ArchfiendOfIfnir, "{3}{B}{B}", 5, 4, "B")
	a := b12Creature(g, opp.ID, "Theirs A", "Creature — Bear", 3, 3)
	cycleFromHand(t, g, "Archfiend of Ifnir", "Creature — Demon", s58ArchfiendOfIfnir, "{C}{C}")
	passPriorityAroundTable(t, g)
	if got := counterCount(g, a, game.CounterMinusOne); got != 1 {
		t.Errorf("cycling a card put %d counters on each creature, want 1", got)
	}
}

// Curse of Fool's Wisdom: each card the enchanted player draws costs
// them 2 life and gains the Curse's controller 2; nobody else's draw
// does anything.
func TestS58CurseOfFoolsWisdom(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	castCatalogSpell(t, g, "Curse of Fool's Wisdom", "Enchantment — Aura Curse", s58CurseOfFoolsWisdom,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	myLife, oppLife, thirdLife := me.Life, opp.Life, third.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 3) })
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-6 || me.Life != myLife+6 {
		t.Errorf("three cards drawn: enchanted player %d→%d, controller %d→%d", oppLife, opp.Life, myLife, me.Life)
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(third.ID, 2) })
	passPriorityAroundTable(t, g)
	if third.Life != thirdLife || me.Life != myLife+6 {
		t.Error("an unenchanted player's draw triggered the Curse")
	}
	if spec, _ := Lookup(s58CurseOfFoolsWisdom); spec.Madness != "{3}{B}" {
		t.Errorf("madness cost %q", spec.Madness)
	}
}

// s58Library stacks named cards on top of p's library; the LAST name is
// the top card.
func s58Library(p *game.Player, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	for i, n := range names {
		ids[i] = uuid.New()
		p.Library.PushTop(game.Card{InstanceID: ids[i], Name: n, TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
	return ids
}

func s58CastRevenant(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	advanceToMain(t, g)
	castFromHandForTest(t, g, me, "Starving Revenant", "Creature — Spirit Horror", "", s58StarvingRevenant, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
}

// Starving Revenant: a card draw and 3 life per card put back on top.
func TestS58StarvingRevenantSurveil(t *testing.T) {
	for _, tc := range []struct {
		name string
		keep func(top []uuid.UUID) (graveyard, onTop []uuid.UUID)
		kept int
	}{
		{"keep both", func(top []uuid.UUID) ([]uuid.UUID, []uuid.UUID) { return nil, top }, 2},
		{"keep one", func(top []uuid.UUID) ([]uuid.UUID, []uuid.UUID) { return top[:1], top[1:] }, 1},
		{"keep none", func(top []uuid.UUID) ([]uuid.UUID, []uuid.UUID) { return top, nil }, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			s58Library(me, "Deep", "Next", "Top")
			hand, life := me.Hand.Size(), me.Life
			s58CastRevenant(t, g, me)
			c := surveilChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no surveil prompt")
			}
			bin, top := tc.keep(c.ScryCards)
			if err := g.ResolveSurveil(c.ID, me.ID, bin, top); err != nil {
				t.Fatalf("ResolveSurveil: %v", err)
			}
			passPriorityAroundTable(t, g)
			// The Revenant itself left the hand when cast.
			if got := me.Hand.Size() - (hand - 0); got != tc.kept {
				t.Errorf("cards drawn: %d, want %d", got, tc.kept)
			}
			if me.Life != life-3*tc.kept {
				t.Errorf("life %d→%d, want a loss of %d", life, me.Life, 3*tc.kept)
			}
		})
	}
}

// Descend 8: once eight permanent cards are in the graveyard, each draw
// is a drain of one; with seven it is nothing.
func TestS58StarvingRevenantDescend(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Starving Revenant", "Creature — Spirit Horror", s58StarvingRevenant, "{2}{B}{B}", 4, 4, "B")
	for i := 0; i < 7; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Permanent", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	}
	myLife, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	if me.Life != myLife || opp.Life != oppLife {
		t.Fatal("seven permanent cards must not turn descend on")
	}
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Permanent", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	if len(g.PendingChoices) > 0 {
		answerPickTargetPlayer(t, g, opp.ID)
	}
	passPriorityAroundTable(t, g)
	if me.Life != myLife+1 || opp.Life != oppLife-1 {
		t.Errorf("descend drain: me %d→%d, opponent %d→%d", myLife, me.Life, oppLife, opp.Life)
	}
	// An opponent's draw is not "you draw".
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	if me.Life != myLife+1 {
		t.Error("an opponent's draw triggered descend")
	}
}

// The Master of Lake-town: whoever loses life mills that many cards,
// damage included; and it draws per big graveyard when it dies.
func TestS58MasterOfLakeTownMillsTheLoser(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := b27Push(g, me.ID, "The Master of Lake-town", "Legendary Creature — Human Advisor", s58MasterOfLakeTown, "{1}{B}{B}", 3, 2, "B")
	assertKeywords(t, g, master, "deathtouch")

	oppLib, myLib := opp.Library.Size(), me.Library.Size()
	s58LoseLife(g, master, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != oppLib-3 || me.Library.Size() != myLib {
		t.Errorf("life loss: opponent library %d→%d, mine %d→%d", oppLib, opp.Library.Size(), myLib, me.Library.Size())
	}

	s58LoseLife(g, master, me.ID, 2) // its own controller too
	passPriorityAroundTable(t, g)
	if me.Library.Size() != myLib-2 {
		t.Errorf("the controller's loss milled %d, want 2", myLib-me.Library.Size())
	}

	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(master, opp.ID, 4) })
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != oppLib-3-4 {
		t.Errorf("damage milled %d, want 4", oppLib-3-opp.Library.Size())
	}
}

func TestS58MasterOfLakeTownDrawsPerBigGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	master := b27Push(g, me.ID, "The Master of Lake-town", "Legendary Creature — Human Advisor", s58MasterOfLakeTown, "{1}{B}{B}", 3, 2, "B")
	for _, p := range []*game.Player{opp, third} {
		for i := 0; i < 7; i++ {
			p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bin", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
		}
	}
	// A sixth-card graveyard on the fourth seat does not count; mine
	// holds only the Master itself.
	for i := 0; i < 6; i++ {
		g.Seats[3].Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bin", TypeLine: "Sorcery", Owner: g.Seats[3].ID, Controller: g.Seats[3].ID})
	}
	hand := me.Hand.Size()
	b27Kill(g, master)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d→%d, want two cards (two graveyards with seven or more)", hand, me.Hand.Size())
	}
}

// Scheming Silvertongue: gaining 2 life this turn makes it prepared at
// your second main phase, and the prepared copy is Sign in Blood.
func TestS58SchemingSilvertonguePrepares(t *testing.T) {
	for _, tc := range []struct {
		gain int
		want bool
	}{{1, false}, {2, true}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		card := preparationCard(me.ID, s58SchemingSilvertongue, []string{"flying", "lifelink"},
			game.Face{Name: "Scheming Silvertongue", TypeLine: "Creature — Vampire Warlock", ManaCost: "{1}{B}", Colors: []string{"B"}, Power: 1, Toughness: 1},
			game.Face{Name: "Sign in Blood", TypeLine: "Sorcery", ManaCost: "{B}{B}", Colors: []string{"B"}})
		advanceToMain(t, g)
		castFromHandAndResolve(t, g, me, card)
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(card.InstanceID, me.ID, tc.gain) })
		advanceTo(t, g, game.StepPostcombatMain)
		passPriorityAroundTable(t, g)
		if got := g.IsPreparedForEffect(card.InstanceID); got != tc.want {
			t.Errorf("gained %d: prepared = %v, want %v", tc.gain, got, tc.want)
		}
		if tc.want {
			if _, ok := prepareCopyOf(g, "Sign in Blood"); !ok {
				t.Error("no copy of Sign in Blood in exile")
			}
		}
	}
}

// Sign in Blood, cast as the prepare spell, is the same card: two
// cards and 2 life for the target.
func TestS58SchemingSilvertongueSignInBlood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spec, ok := Lookup(s58SchemingSilvertongue + "#1")
	if !ok || spec.OnResolve == nil {
		t.Fatal("the prepare spell is not registered")
	}
	hand, life := opp.Hand.Size(), opp.Life
	castCatalogSpell(t, g, "Sign in Blood", "Sorcery", s58SchemingSilvertongue+"#1",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+2 || opp.Life != life-2 {
		t.Errorf("target: hand %d→%d, life %d→%d", hand, opp.Hand.Size(), life, opp.Life)
	}
	_ = me
}
