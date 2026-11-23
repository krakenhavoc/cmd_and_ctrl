package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	bnsB38ShieldMare  = "e07b6988-9ee7-4f20-8aa5-7dfa87ff507b"
	bnsB38ShoreKeeper = "9625773b-1976-49b3-88d2-93a4befb595c"
	bnsB38Starfish    = "94a4f70a-8aac-412a-99f0-933e3ff87fe0"
	bnsB38SkyEel      = "c90ea9d5-82f7-4b92-8cb1-10d0a7378e8c"
	bnsB38Bloodletter = "41684707-5118-45cf-892d-46b66045d026"
	bnsB38Piper       = "be0c9947-3e7c-4b53-b87c-ff16ffd668ea"
	bnsB38Gnarlid     = "eeba70f5-336c-46e6-90db-50818f88e8a5"
	bnsB38Dragon      = "1817acd5-5ed0-47b8-8e56-4715a488867e"
	bnsB38Tyrranax    = "2d09e689-c163-48c5-af5d-766f110d041e"
	bnsB38Thallid     = "d0c2ca8a-c5d4-4684-bacb-46c04f47c276"
)

func TestBnsB38Registered(t *testing.T) {
	for _, id := range []string{bnsB38ShieldMare, bnsB38ShoreKeeper, bnsB38Starfish, bnsB38SkyEel, bnsB38Bloodletter,
		bnsB38Piper, bnsB38Gnarlid, bnsB38Dragon, bnsB38Tyrranax, bnsB38Thallid} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v completeness=%v caveats=%v", id, ok, spec.Completeness, spec.Caveats)
		}
	}
}

func TestBnsB38ShieldMareCantBeBlockedByRedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	mare := effectstest.TmpB1Push(g, me.ID, "Shield Mare", "Creature — Horse", bnsB38ShieldMare, 2, 3, "W")
	red := effectstest.TmpB1Push(g, them.ID, "Red blocker", "Creature — Goblin", "", 2, 2, "R")
	green := effectstest.TmpB1Push(g, them.ID, "Green blocker", "Creature — Bear", "", 2, 2, "G")
	effectstest.AttackIntoBlocks(t, g, them.ID, mare)
	if got := effectstest.BlockRefusal(t, g, red, mare); got != game.BlockReasonCantBeBlockedBy {
		t.Errorf("red blocker: refusal %q, want %q", got, game.BlockReasonCantBeBlockedBy)
	}
	if got := effectstest.BlockRefusal(t, g, green, mare); got != "" {
		t.Errorf("green blocker refused: %q", got)
	}
}

func TestBnsB38ShieldMareGainsThreeOnEnteringAndWhenAnOpponentTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	before := me.Life
	effectstest.Ecl6Cast(t, g, me, "Shield Mare", "Creature — Horse", bnsB38ShieldMare, 2, 3, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+3 {
		t.Fatalf("life after entering = %d, want %d", me.Life, before+3)
	}
}

func TestBnsB38ShieldMareTargetedByAnOpponentGainsLifeOnceAndOwnSpellDoesNot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1] // seat 0 is active and casts
	mare := effectstest.TmpB1Push(g, opp.ID, "Shield Mare", "Creature — Horse", bnsB38ShieldMare, 2, 3, "W")
	effectstest.AdvanceToMain(t, g)
	before := opp.Life
	bolt := effectstest.PushCatalogHandCard(me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: effectstest.CardRefs(mare)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.B06AnswerAllTriggerPrompts(t, g, opp.ID, true)
	if opp.Life != before+3 {
		t.Errorf("life = %d, want %d after an opponent targeted Shield Mare", opp.Life, before+3)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	mare2 := effectstest.TmpB1Push(g2, me2.ID, "Shield Mare", "Creature — Horse", bnsB38ShieldMare, 2, 3, "W")
	effectstest.AdvanceToMain(t, g2)
	before2 := me2.Life
	pump := effectstest.PushCatalogHandCard(me2, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle)
	if err := g2.CastSpell(me2.ID, pump, game.CastSpellParams{Targets: effectstest.CardRefs(mare2)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.B06AnswerAllTriggerPrompts(t, g2, me2.ID, true)
	if me2.Life != before2 {
		t.Errorf("own spell changed life %d -> %d", before2, me2.Life)
	}
}

func TestBnsB38ShoreKeeperSacrificesToDrawThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	k := effectstest.B12Push(g, me.ID, "Shore Keeper", "Creature — Trilobite", bnsB38ShoreKeeper, 0, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.Mm2Library(me, effectstest.Mm2Spell("A"), effectstest.Mm2Spell("B"), effectstest.Mm2Spell("C"), effectstest.Mm2Spell("D"))
	hand := len(me.Hand.Cards)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{U}")
	effectstest.B16Activate(t, g, me.ID, k, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
	if !me.Graveyard.Contains(k) {
		t.Error("Shore Keeper was not sacrificed")
	}
}

func TestBnsB38ShoreKeeperNeedsEightMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	k := effectstest.B12Push(g, me.ID, "Shore Keeper", "Creature — Trilobite", bnsB38ShoreKeeper, 0, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, k, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with seven mana")
	}
}

func TestBnsB38SigiledStarfishScrysOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	s := effectstest.B12Push(g, me.ID, "Sigiled Starfish", "Creature — Starfish", bnsB38Starfish, 0, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.Mm2Library(me, effectstest.Mm2Spell("A"), effectstest.Mm2Spell("B"))
	effectstest.B16Activate(t, g, me.ID, s, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) == 0 {
		t.Fatal("no scry prompt opened")
	}
}

func TestBnsB38SkyEelSchoolLootsOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.Mm2Library(me, effectstest.Mm2Spell("A"), effectstest.Mm2Spell("B"))
	hand := len(me.Hand.Cards)
	effectstest.Ecl6Cast(t, g, me, "Sky-Eel School", "Creature — Fish", bnsB38SkyEel, 3, 3, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, firstOnBattlefield(g, bnsB38SkyEel), "flying") {
		t.Error("no flying")
	}
	if len(g.PendingChoices) == 0 {
		t.Fatal("no discard prompt after drawing")
	}
	if got := len(me.Hand.Cards) - hand; got != 1 {
		t.Errorf("hand grew by %d before discarding, want 1 (cast card left, one drawn)", got+1)
	}
}

func firstOnBattlefield(g *game.Game, oracle string) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == oracle {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func TestBnsB38SkymarchBloodletterDrainsOneOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	myLife, theirLife := me.Life, opp.Life
	effectstest.Ecl6Cast(t, g, me, "Skymarch Bloodletter", "Creature — Vampire Soldier", bnsB38Bloodletter, 2, 2, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-1 || me.Life != myLife+1 {
		t.Errorf("life me %d->%d opp %d->%d, want +1 / -1", myLife, me.Life, theirLife, opp.Life)
	}
}

func TestBnsB38SlimyPiperAttacksAsThreeTwoOrWithFourCreaturesFourThreeIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := effectstest.SeatsForCombat(g)
	p := effectstest.B12Push(g, me.ID, "Slimy Piper", "Creature — Fungus Bard", bnsB38Piper, 2, 1)
	effectstest.DeclareAttack(t, g, them.ID, p)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, p); c.CurrentPower() != 3 || c.CurrentToughness() != 2 {
		t.Errorf("with fewer than four creatures it is %d/%d, want 3/2", c.CurrentPower(), c.CurrentToughness())
	}
	if effectstest.HasEffectiveKeyword(t, g, p, "indestructible") {
		t.Error("indestructible with fewer than four creatures")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, them2 := effectstest.SeatsForCombat(g2)
	p2 := effectstest.B12Push(g2, me2.ID, "Slimy Piper", "Creature — Fungus Bard", bnsB38Piper, 2, 1)
	for i := 0; i < 3; i++ {
		effectstest.B12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	}
	effectstest.DeclareAttack(t, g2, them2.ID, p2)
	effectstest.PassPriorityAroundTable(t, g2)
	if c := effectstest.GaLayered(t, g2, p2); c.CurrentPower() != 4 || c.CurrentToughness() != 3 {
		t.Errorf("with four creatures it is %d/%d, want 4/3", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.HasEffectiveKeyword(t, g2, p2, "indestructible") {
		t.Error("no indestructible with four creatures")
	}
}

func TestBnsB38SnappingGnarlidGrowsOnLandfallUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Snapping Gnarlid", "Creature — Beast", bnsB38Gnarlid, 2, 2)
	effectstest.ToMainForCost(t, g)
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, id); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("after a land it is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.AangAdvanceToMain(t, g, 1)
	if c := effectstest.GaLayered(t, g, id); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("next turn it is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestBnsB38SparktongueDragonPaysThreeForThreeDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.Ecl6Cast(t, g, me, "Sparktongue Dragon", "Creature — Dragon", bnsB38Dragon, 3, 3, game.CastSpellParams{})
	effectstest.B06AddMana(me, "R", "R", "R")
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3", life, opp.Life)
	}
	if len(effectstest.PoolColors(me)) != 0 {
		t.Errorf("pool = %v after paying {2}{R}", effectstest.PoolColors(me))
	}
}

func TestBnsB38SparktongueDragonDecliningDealsNoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Ecl6Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.Ecl6Cast(t, g, me, "Sparktongue Dragon", "Creature — Dragon", bnsB38Dragon, 3, 3, game.CastSpellParams{})
	effectstest.B06AddMana(me, "R", "R", "R")
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("opponent life %d -> %d after declining", life, opp.Life)
	}
	if len(effectstest.PoolColors(me)) != 3 {
		t.Errorf("the mana was spent anyway: pool = %v", effectstest.PoolColors(me))
	}
}

func bnsB38TyrranaxAtCombat(t *testing.T, g *game.Game) (me *game.Player, bear uuid.UUID) {
	t.Helper()
	me = g.Seats[0]
	effectstest.B12Push(g, me.ID, "Spined Tyrranax", "Creature — Dinosaur Beast", bnsB38Tyrranax, 5, 5)
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B06AddMana(me, "G", "G", "G")
	effectstest.PassPriorityAroundTable(t, g)
	return me, bear
}

func TestBnsB38SpinedTyrranaxPaysForACounterAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bear := bnsB38TyrranaxAtCombat(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("Bear has %d counters, want 1", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("Bear has no trample")
	}
	if len(effectstest.PoolColors(me)) != 0 {
		t.Errorf("pool = %v after paying {2}{G}", effectstest.PoolColors(me))
	}
}

func TestBnsB38SpinedTyrranaxDecliningDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bear := bnsB38TyrranaxAtCombat(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("Bear has %d counters after declining", n)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("Bear gained trample after declining")
	}
}

func TestBnsB38SporecrownThallidPumpsOtherFungusAndSaprolingYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	th := effectstest.B12Push(g, me.ID, "Sporecrown Thallid", "Creature — Fungus", bnsB38Thallid, 2, 2)
	fungus := effectstest.B12Creature(g, me.ID, "Fungus", "Creature — Fungus", 1, 1)
	sap := effectstest.B12Creature(g, me.ID, "Saproling", "Creature — Saproling", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Fungus", "Creature — Fungus", 1, 1)
	effectstest.AdvanceToMain(t, g)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		p, t int
	}{{"itself", th, 2, 2}, {"Fungus", fungus, 2, 2}, {"Saproling", sap, 2, 2}, {"Bear", bear, 2, 2}, {"opponent's Fungus", theirs, 1, 1}} {
		if c := effectstest.GaLayered(t, g, tc.id); c.CurrentPower() != tc.p || c.CurrentToughness() != tc.t {
			t.Errorf("%s is %d/%d, want %d/%d", tc.name, c.CurrentPower(), c.CurrentToughness(), tc.p, tc.t)
		}
	}
}
