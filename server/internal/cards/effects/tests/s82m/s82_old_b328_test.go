package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b328_test.go — Dragons of Tarkir, batch 328 (#5551): the exploit
// creatures. Each test plays the real catalog entry through the engine and
// asserts what its text says.

const (
	oldB328Gravecaller = "1b57e9f8-68af-4f83-9062-6d715857dcef"
	oldB328Sidisi      = "3406db41-6b95-48fc-b58e-4476220fd34d"
	oldB328Faithful    = "45de1caf-46c2-4acc-8be2-08e1840a3b98"
	oldB328Minister    = "5d5b63a4-c9d9-4971-9d4e-2eab86f251e2"
	oldB328Drowner     = "9839f80b-d28b-49c7-a31a-40c245af2edc"
	oldB328Sorcerer    = "be5c87d3-492c-4b73-9d79-067bd3c3e738"
	oldB328Profaner    = "e6bfac15-ec3c-4a55-946a-0b4514eabafd"
)

func TestOldB328CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB328Gravecaller: "Rakshasa Gravecaller", oldB328Sidisi: "Sidisi, Undead Vizier", oldB328Faithful: "Sidisi's Faithful",
		oldB328Minister: "Minister of Pain", oldB328Drowner: "Gurmag Drowner", oldB328Sorcerer: "Silumgar Sorcerer",
		oldB328Profaner: "Profaner of the Dead",
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

func TestOldB328PrintedKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle string
		kws                    []string
	}{
		{"Sidisi, Undead Vizier", "Legendary Creature — Zombie Snake", oldB328Sidisi, []string{"deathtouch"}},
		{"Silumgar Sorcerer", "Creature — Human Wizard Sorcerer", oldB328Sorcerer, []string{"flash", "flying"}},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
		for _, kw := range tc.kws {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

// oldB328Exploit casts the creature, takes the exploit on `fodder` and leaves
// the payoff on the stack or pending.
func oldB328Exploit(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string, fodder uuid.UUID) {
	t.Helper()
	effectstest.ExploitCast(t, g, me, name, typeLine, oracle)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
}

func TestOldB328RakshasaGravecallerMakesTwoZombies(t *testing.T) {
	g, me, _, fodder := effectstest.VowExploitTable(t)
	oldB328Exploit(t, g, me, "Rakshasa Gravecaller", "Creature — Demon", oldB328Gravecaller, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the exploited creature is still on the battlefield")
	}
	zombies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.HasSubtype("Zombie") && c.Power == 2 && c.Toughness == 2 {
			zombies++
		}
	}
	if zombies != 2 {
		t.Errorf("%d 2/2 Zombie tokens, want 2", zombies)
	}
}

func TestOldB328RakshasaGravecallerSacrificingItselfStillMakesTwoZombies(t *testing.T) {
	g, me, _, _ := effectstest.VowExploitTable(t)
	id := effectstest.ExploitCast(t, g, me, "Rakshasa Gravecaller", "Creature — Demon", oldB328Gravecaller)
	effectstest.AnswerOwnPermanents(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	zombies := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Zombie") {
			zombies++
		}
	}
	if g.Battlefield.Contains(id) || zombies != 2 {
		t.Errorf("Gravecaller on battlefield %v, %d Zombies, want it gone and 2 Zombies", g.Battlefield.Contains(id), zombies)
	}
}

func TestOldB328RakshasaGravecallerDeclinedExploitMakesNoZombies(t *testing.T) {
	g, me, _, _ := effectstest.VowExploitTable(t)
	effectstest.VowDeclined(t, g, me, "Rakshasa Gravecaller", "Creature — Demon", oldB328Gravecaller)
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() {
			t.Error("a token appeared although the exploit was declined")
		}
	}
}

func TestOldB328SidisiSearchesForAnyCard(t *testing.T) {
	g, me, _, fodder := effectstest.VowExploitTable(t)
	ids := effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"), game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	oldB328Exploit(t, g, me, "Sidisi, Undead Vizier", "Legendary Creature — Zombie Snake", oldB328Sidisi, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) {
		t.Error("the searched card is not in hand")
	}
}

func TestOldB328SidisiMayDeclineTheSearchAndADeclinedExploitSearchesNot(t *testing.T) {
	g, me, _, fodder := effectstest.VowExploitTable(t)
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"), game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	hand := me.Hand.Size()
	oldB328Exploit(t, g, me, "Sidisi, Undead Vizier", "Legendary Creature — Zombie Snake", oldB328Sidisi, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the optional search was not offered")
	}
	if err := effectstest.AnswerSearchApply(g, c, me.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after declining the search, want %d", me.Hand.Size(), hand)
	}
	g2, me2, _, _ := effectstest.VowExploitTable(t)
	effectstest.SeedSearchLibrary(me2, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.VowDeclined(t, g2, me2, "Sidisi, Undead Vizier", "Legendary Creature — Zombie Snake", oldB328Sidisi)
	if effectstest.SearchChoiceFor(g2, me2.ID) != nil {
		t.Error("a declined exploit still searched")
	}
}

func TestOldB328SidisisFaithfulBouncesATargetCreature(t *testing.T) {
	g, me, opp, fodder := effectstest.VowExploitTable(t)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	oldB328Exploit(t, g, me, "Sidisi's Faithful", "Creature — Snake Wizard", oldB328Faithful, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Errorf("the Bear should be back in its owner's hand: battlefield %v, hand %v", g.Battlefield.Contains(bear), opp.Hand.Contains(bear))
	}
}

func TestOldB328MinisterOfPainShrinksOnlyOpponentsCreatures(t *testing.T) {
	g, me, opp, fodder := effectstest.VowExploitTable(t)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	oldB328Exploit(t, g, me, "Minister of Pain", "Creature — Human Shaman", oldB328Minister, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the opponent's 1/1 survived -1/-1")
	}
	if p, tou := effectstest.PtOf(t, g, big); p != 2 || tou != 2 {
		t.Errorf("the opponent's 3/3 is %d/%d, want 2/2", p, tou)
	}
	if p, tou := effectstest.PtOf(t, g, mine); p != 3 || tou != 3 {
		t.Errorf("my 3/3 is %d/%d, want 3/3", p, tou)
	}
}

func TestOldB328GurmagDrownerPutsOneOfTheTopFourInHandAndMillsTheRest(t *testing.T) {
	g, me, _, fodder := effectstest.VowExploitTable(t)
	top := effectstest.SeedLibrary(me, "Four", "Three", "Two", "One")
	hand, gy := me.Hand.Size(), me.Graveyard.Size()
	oldB328Exploit(t, g, me, "Gurmag Drowner", "Creature — Snake Wizard", oldB328Drowner, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no card was offered")
	}
	if len(pick.ChooseCards) != 4 {
		t.Errorf("%d cards offered, want the top four", len(pick.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top[len(top)-1])
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if got := me.Graveyard.Size() - gy; got != 4 {
		// three milled plus the exploited creature
		t.Errorf("graveyard grew by %d, want 4 (three cards and the fodder)", got)
	}
}

func TestOldB328SilumgarSorcererCountersACreatureSpell(t *testing.T) {
	g, me, _, fodder := effectstest.VowExploitTable(t)
	bait := effectstest.HandCardForTest(me, "Bait", "Creature — Bear", "")
	if err := g.CastSpell(me.ID, bait, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the bait: %v", err)
	}
	effectstest.ExploitCast(t, g, me, "Silumgar Sorcerer", "Creature — Human Wizard Sorcerer", oldB328Sorcerer)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bait)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bait) {
		t.Error("the countered creature spell resolved")
	}
	if !me.Graveyard.Contains(bait) {
		t.Error("the countered spell is not in its owner's graveyard")
	}
}

func TestOldB328SilumgarSorcererCannotTargetANoncreatureSpell(t *testing.T) {
	g, me, _, fodder := effectstest.VowExploitTable(t)
	spell := effectstest.HandCardForTest(me, "Bolt", "Instant", "")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}
	effectstest.ExploitCast(t, g, me, "Silumgar Sorcerer", "Creature — Human Wizard Sorcerer", oldB328Sorcerer)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Error("the payoff asked for a target although no creature spell was on the stack")
	}
}

func TestOldB328ProfanerBouncesOpponentsCreaturesWithLessToughness(t *testing.T) {
	g, me, opp, _ := effectstest.VowExploitTable(t)
	fodder := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 3)
	two := effectstest.B12Creature(g, opp.ID, "Two", "Creature — Bear", 2, 2)
	one := effectstest.B12Creature(g, opp.ID, "One", "Creature — Bear", 5, 1)
	three := effectstest.B12Creature(g, opp.ID, "Three", "Creature — Bear", 1, 3)
	four := effectstest.B12Creature(g, opp.ID, "Four", "Creature — Bear", 1, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	oldB328Exploit(t, g, me, "Profaner of the Dead", "Creature — Snake Wizard", oldB328Profaner, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{two, one} {
		if g.Battlefield.Contains(id) || !opp.Hand.Contains(id) {
			t.Error("a creature with toughness below 3 was not returned to its owner's hand")
		}
	}
	for _, id := range []uuid.UUID{three, four} {
		if !g.Battlefield.Contains(id) {
			t.Error("a creature with toughness 3 or more was returned")
		}
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the controller's own creature was returned")
	}
}
