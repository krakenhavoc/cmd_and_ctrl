package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 older sets, Theros heroic, batch 388 (#5561).

const (
	oldB388GiantGrowthOracle = "5748ebf1-24e3-499d-ab7c-c2cebd462a24"
	oldB388PhalanxOracle     = "857db7e0-254a-4920-a71d-9bab6555342f"
	oldB388TormentedOracle   = "47fbb4f0-565b-4682-9971-0735210c76e5"
	oldB388ArtisanOracle     = "3e8abec7-1e18-460a-9d88-30969755ebd2"
	oldB388BattlewiseOracle  = "416de3dd-c15d-4596-9386-5cac53302724"
	oldB388AnaxOracle        = "9c4219db-1a7a-4d61-bb8b-7998cfed1faf"
	oldB388TritonOracle      = "2a7e68ed-156c-4680-851b-ae8de4a808a5"
	oldB388AnthousaOracle    = "141f952b-9e9f-48c5-9c88-b1504dfa73bc"
	oldB388PriestOracle      = "e2886098-0309-4da5-b2e5-a06af5cfe5b8"
	oldB388AgentOracle       = "45f88a6a-ac8a-4a53-a96b-f09ca33c79ba"
)

func TestOldB388EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB388PhalanxOracle:    "Phalanx Leader",
		oldB388TormentedOracle:  "Tormented Hero",
		oldB388ArtisanOracle:    "Artisan of Forms",
		oldB388BattlewiseOracle: "Battlewise Hoplite",
		oldB388AnaxOracle:       "Anax and Cymede",
		oldB388TritonOracle:     "Wavecrash Triton",
		oldB388AnthousaOracle:   "Anthousa, Setessan Hero",
		oldB388PriestOracle:     "Setessan Battle Priest",
		oldB388AgentOracle:      "Agent of the Fates",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB388Heroic casts Giant Growth on `target` (a heroic trigger for the
// creature's controller) and settles the stack, answering a target pick the
// trigger asks.
func oldB388Heroic(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", oldB388GiantGrowthOracle, effectstest.CardTarget(target))
	for i := 0; i < 6; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, me.ID) == nil {
			return
		}
		effectstest.AnswerFirstPickTarget(t, g)
	}
}

func oldB388Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB388Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.BattlefieldCardCopy(t, g, id).Counters[game.CounterPlusOne]
}

func TestOldB388PhalanxLeaderPutsACounterOnEachCreatureYouControl(t *testing.T) {
	g, me, opp := oldB388Table(t)
	leader := effectstest.B12Push(g, me.ID, "Phalanx Leader", "Creature — Human Soldier", oldB388PhalanxOracle, 1, 1)
	mine := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	oldB388Heroic(t, g, me, leader)
	if oldB388Counters(t, g, leader) != 1 || oldB388Counters(t, g, mine) != 1 {
		t.Error("each creature I control should have one +1/+1 counter")
	}
	if oldB388Counters(t, g, theirs) != 0 {
		t.Error("an opponent's creature got a counter")
	}
}

func TestOldB388HeroicIgnoresASpellThatTargetsSomethingElse(t *testing.T) {
	g, me, _ := oldB388Table(t)
	leader := effectstest.B12Push(g, me.ID, "Phalanx Leader", "Creature — Human Soldier", oldB388PhalanxOracle, 1, 1)
	mine := effectstest.SeedBear(g, me.ID)
	oldB388Heroic(t, g, me, mine)
	if oldB388Counters(t, g, leader) != 0 || oldB388Counters(t, g, mine) != 0 {
		t.Error("a spell targeting another creature triggered heroic")
	}
}

func TestOldB388TormentedHeroDrainsEachOpponentAndEntersTapped(t *testing.T) {
	g, me, _ := oldB388Table(t)
	hero := effectstest.CastCatalogSpell(t, g, "Tormented Hero", "Creature — Human Warrior", oldB388TormentedOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldCardCopy(t, g, hero).Tapped {
		t.Error("Tormented Hero did not enter tapped")
	}
	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID] = p.Life
	}
	oldB388Heroic(t, g, me, hero)
	opps := len(g.Seats) - 1
	if me.Life != life[me.ID]+opps {
		t.Errorf("my life = %d, want %d", me.Life, life[me.ID]+opps)
	}
	for _, p := range g.Seats {
		if p.ID != me.ID && p.Life != life[p.ID]-1 {
			t.Errorf("%s life = %d, want %d", p.ID, p.Life, life[p.ID]-1)
		}
	}
}

func TestOldB388BattlewiseHopliteGetsACounterAndScries(t *testing.T) {
	g, me, _ := oldB388Table(t)
	hop := effectstest.B12Push(g, me.ID, "Battlewise Hoplite", "Creature — Human Soldier", oldB388BattlewiseOracle, 2, 2)
	oldB388Heroic(t, g, me, hop)
	if oldB388Counters(t, g, hop) != 1 {
		t.Error("no +1/+1 counter")
	}
	if effectstest.PendingOfKind(g, game.PendingChoiceScry) == nil {
		t.Error("no scry prompt was opened")
	}
}

func TestOldB388AnaxPumpsAndGivesTrampleToCreaturesYouControl(t *testing.T) {
	g, me, opp := oldB388Table(t)
	anax := effectstest.B12Push(g, me.ID, "Anax and Cymede", "Legendary Creature — Human Soldier", oldB388AnaxOracle, 3, 2)
	mine := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	oldB388Heroic(t, g, me, anax)
	// Giant Growth itself adds +3/+3 to Anax; read my bear and theirs.
	if p, _ := effectstest.PtOf(t, g, mine); p != 3 {
		t.Errorf("my bear has power %d, want 3", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("my bear lacks trample")
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature was affected")
	}
}

func TestOldB388WavecrashTritonTapsAndFreezesAnOpponentsCreature(t *testing.T) {
	g, me, opp := oldB388Table(t)
	triton := effectstest.B12Push(g, me.ID, "Wavecrash Triton", "Creature — Merfolk Wizard", oldB388TritonOracle, 1, 4)
	theirs := effectstest.SeedBear(g, opp.ID)
	oldB388Heroic(t, g, me, triton)
	c := effectstest.BattlefieldCardCopy(t, g, theirs)
	if !c.Tapped {
		t.Error("the opponent's creature was not tapped")
	}
}

func TestOldB388AnthousaMakesUpToThreeLandsWarriors(t *testing.T) {
	g, me, _ := oldB388Table(t)
	anthousa := effectstest.B12Push(g, me.ID, "Anthousa, Setessan Hero", "Legendary Creature — Human Warrior", oldB388AnthousaOracle, 4, 5)
	land := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	oldB388Heroic(t, g, me, anthousa)
	if p, tough := effectstest.PtOf(t, g, land); p != 2 || tough != 2 {
		t.Errorf("the land is %d/%d, want a 2/2 Warrior", p, tough)
	}
}

func TestOldB388SetessanBattlePriestGainsTwo(t *testing.T) {
	g, me, _ := oldB388Table(t)
	priest := effectstest.B12Push(g, me.ID, "Setessan Battle Priest", "Creature — Human Cleric", oldB388PriestOracle, 1, 3)
	life := me.Life
	oldB388Heroic(t, g, me, priest)
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
}

func TestOldB388AgentOfTheFatesMakesEachOpponentSacrifice(t *testing.T) {
	g, me, opp := oldB388Table(t)
	agent := effectstest.B12Push(g, me.ID, "Agent of the Fates", "Creature — Human Assassin", oldB388AgentOracle, 3, 2)
	theirs := effectstest.SeedBear(g, opp.ID)
	oldB388Heroic(t, g, me, agent)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's only creature was not sacrificed")
	}
	if !g.Battlefield.Contains(agent) {
		t.Error("my creature was sacrificed")
	}
}

func TestOldB388ArtisanOfFormsBecomesACopyAndKeepsTheAbility(t *testing.T) {
	g, me, opp := oldB388Table(t)
	artisan := effectstest.B12Push(g, me.ID, "Artisan of Forms", "Creature — Human Wizard", oldB388ArtisanOracle, 1, 1)
	target := effectstest.B12Push(g, opp.ID, "Big Beast", "Creature — Beast", "", 6, 6)
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", oldB388GiantGrowthOracle, effectstest.CardTarget(artisan))
	for i := 0; i < 8; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			effectstest.PickCard(t, g, me.ID, target)
			continue
		}
		if effectstest.PendingOfKind(g, game.PendingChoiceTriggerPrompt) != nil {
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			continue
		}
		break
	}
	if c := effectstest.BattlefieldCardCopy(t, g, artisan); c.Name != "Big Beast" {
		t.Fatalf("the Artisan is %q, want a copy of Big Beast", c.Name)
	}
	// The copy has the ability too: another targeting spell asks for a target again.
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", oldB388GiantGrowthOracle, effectstest.CardTarget(artisan))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) == nil && effectstest.PendingOfKind(g, game.PendingChoiceTriggerPrompt) == nil {
		t.Error("the copy did not keep the heroic ability")
	}
}
