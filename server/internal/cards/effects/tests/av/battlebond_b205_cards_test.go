package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// battlebond_b205_cards_test.go — Battlebond, S82 older-sets batch 205
// (#5529): counters, card draw and enter/die triggers.

const (
	bbdB205HydraOracle    = "1a840210-204d-41cd-afc0-46abf7ea5f87"
	bbdB205GwafaOracle    = "7ba31f45-c178-4bac-ab7b-5ba4d050133c"
	bbdB205SanctionOracle = "b594f7f1-1db7-46e9-bbc7-42ccc16ce9a1"
	bbdB205QuestOracle    = "c5f92560-dc82-4bd2-ae6b-4c3942d458b7"
	bbdB205ReversalOracle = "3db00947-d568-45d7-82cd-b54b4fcb1312"
	bbdB205AzraOracle     = "457470db-e664-4ca3-84ec-6cec2d4ae906"
	bbdB205KissOracle     = "779d3e58-de06-44be-98f8-f05f7711ecb2"
	bbdB205FuryOracle     = "c3575cf9-847a-472c-94b3-ef70712c297d"
	bbdB205MummyOracle    = "414c97a3-cf31-4145-a431-3b1ad4343947"
	bbdB205WumpusOracle   = "698c9882-4117-4919-8a19-b1cbffc16692"
	bbdB205DeadOracle     = "9ae376d5-c68f-4a4b-851c-ebfd385b86dd"
	bbdB205VictorOracle   = "fa58591b-240f-4527-8308-7b87263e784f"
)

func TestBbdB205EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bbdB205HydraOracle:    "Apocalypse Hydra",
		bbdB205GwafaOracle:    "Gwafa Hazid, Profiteer",
		bbdB205SanctionOracle: "Regna's Sanction",
		bbdB205QuestOracle:    "Quest for the Gravelord",
		bbdB205ReversalOracle: "Stunning Reversal",
		bbdB205AzraOracle:     "Azra Oddsmaker",
		bbdB205KissOracle:     "Kiss of the Amesha",
		bbdB205FuryOracle:     "Khorvath's Fury",
		bbdB205MummyOracle:    "Sparring Mummy",
		bbdB205WumpusOracle:   "Hunted Wumpus",
		bbdB205DeadOracle:     "Tenacious Dead",
		bbdB205VictorOracle:   "Enthralling Victor",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v, want Full with no caveats", name, spec, ok)
		}
	}
}

func TestBbdB205ApocalypseHydraEntersWithXCountersAndDoublesAtFive(t *testing.T) {
	for _, tc := range []struct{ x, want int }{{3, 3}, {4, 4}, {5, 10}, {6, 12}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := effectstest.Top100ActiveSeat(g)
		effectstest.FillPool(me, tc.x+2)
		effectstest.FillPoolColored(me, "R", 1)
		effectstest.FillPoolColored(me, "G", 1)
		id := effectstest.CastFromHandForTest(t, g, me, "Apocalypse Hydra", "Creature — Hydra", "{X}{R}{G}", bbdB205HydraOracle,
			game.CastSpellParams{Strict: true, XValue: tc.x})
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.PlusOneCounters(g, id); got != tc.want {
			t.Errorf("X=%d: %d +1/+1 counters, want %d", tc.x, got, tc.want)
		}
	}
}

func TestBbdB205ApocalypseHydraPingsByRemovingACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := effectstest.Top100Opponent(g)
	hydra := effectstest.B12Push(g, me.ID, "Apocalypse Hydra", "Creature — Hydra", bbdB205HydraOracle, 0, 0)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(hydra, game.CounterPlusOne, 3) })
	life := opp.Life
	effectstest.ApaMana(me, "R", "C")
	effectstest.ApaActivate(t, g, me, hydra, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
	if got := effectstest.PlusOneCounters(g, hydra); got != 2 {
		t.Errorf("%d counters left, want 2", got)
	}
}

func TestBbdB205GwafaHazidBribesAndBarsCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := effectstest.Top100Opponent(g)
	gwafa := effectstest.B12Push(g, me.ID, "Gwafa Hazid, Profiteer", "Legendary Creature — Human Rogue", bbdB205GwafaOracle, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, gwafa, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mine)}); err == nil {
		t.Fatal("activated on a creature you control")
	}
	hand := opp.Hand.Size()
	effectstest.ApaMana(me, "W", "U")
	effectstest.ApaActivate(t, g, me, gwafa, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, theirs, "bribery"); got != 1 {
		t.Fatalf("%d bribery counters, want 1", got)
	}
	if got := opp.Hand.Size(); got != hand+1 {
		t.Errorf("its controller's hand %d, want %d", got, hand+1)
	}
	card := effectstest.FindCardAnywhere(t, g, theirs)
	if !card.Effective().Restrictions.Has(game.CantAttack) || !card.Effective().Restrictions.Has(game.CantBlock) {
		t.Errorf("a creature with a bribery counter can still attack or block: %v", card.Effective().Restrictions)
	}
	if mineCard := effectstest.FindCardAnywhere(t, g, mine); mineCard.Effective().Restrictions.Has(game.CantAttack) {
		t.Error("a creature without a bribery counter can't attack")
	}
}

func TestBbdB205QuestForTheGravelordCountsDeathsThenMakesAZombieGiant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := effectstest.Top100Opponent(g)
	quest := effectstest.B12Push(g, me.ID, "Quest for the Gravelord", "Enchantment", bbdB205QuestOracle, 0, 0)
	for i := 0; i < 3; i++ {
		var victim uuid.UUID
		if i%2 == 0 {
			victim = effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 1, 1)
			effectstest.KillCreature(t, g, opp.ID, victim)
		} else {
			victim = effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 1, 1)
			effectstest.KillCreature(t, g, me.ID, victim)
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CountersOn(g, quest, "quest"); got != 3 {
		t.Fatalf("%d quest counters, want 3", got)
	}
	effectstest.ApaActivate(t, g, me, quest, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(quest) {
		t.Error("the enchantment was not sacrificed")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Zombie Giant" && c.IsToken() {
			found = true
			if p, tough := effectstest.PtOf(t, g, c.InstanceID); p != 5 || tough != 5 {
				t.Errorf("the token is %d/%d, want 5/5", p, tough)
			}
		}
	}
	if !found {
		t.Error("no Zombie Giant token")
	}
}

func TestBbdB205StunningReversalReplacesTheNextLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "B", 1)
	spell := effectstest.CastFromHandForTest(t, g, me, "Stunning Reversal", "Instant", "{3}{B}", bbdB205ReversalOracle, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(spell) {
		t.Error("Stunning Reversal did not exile itself")
	}
	hand := me.Hand.Size()
	g.WithWriteLock(func() { me.Life = 0 })
	g.RunStateChecksForTest()
	if me.Eliminated {
		t.Fatal("the loss was not replaced")
	}
	if me.Life != 1 {
		t.Errorf("life %d, want 1", me.Life)
	}
	if got := me.Hand.Size(); got != hand+7 {
		t.Errorf("hand %d, want %d", got, hand+7)
	}
	g.WithWriteLock(func() { me.Life = 0 })
	g.RunStateChecksForTest()
	if !me.Eliminated {
		t.Error("a second loss the same turn was replaced")
	}
}

func TestBbdB205KissOfTheAmeshaGivesLifeAndCardsToTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := effectstest.Top100Opponent(g)
	effectstest.FillPool(me, 4)
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.FillPoolColored(me, "U", 1)
	life, hand := opp.Life, opp.Hand.Size()
	effectstest.CastFromHandForTest(t, g, me, "Kiss of the Amesha", "Sorcery", "{4}{W}{U}", bbdB205KissOracle,
		game.CastSpellParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+7 {
		t.Errorf("life %d, want %d", opp.Life, life+7)
	}
	if got := opp.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
}

func TestBbdB205SparringMummyUntapsATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	effectstest.CastRingCreature(t, g, "Sparring Mummy", "Creature — Zombie", bbdB205MummyOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AvB34Pick(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindCardAnywhere(t, g, bear).Tapped {
		t.Error("the Bear stayed tapped")
	}
}

func TestBbdB205EnthrallingVictorStealsASmallCreatureForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 3, 3)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, small).Tapped = true })
	effectstest.CastRingCreature(t, g, "Enthralling Victor", "Creature — Human Warrior", bbdB205VictorOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Error("a creature with power 3 was accepted")
	}
	effectstest.AvB34Pick(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindCardAnywhere(t, g, small)
	if c.Controller != me.ID {
		t.Error("you did not gain control")
	}
	if c.Tapped {
		t.Error("the creature was not untapped")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, small, "haste") {
		t.Error("the creature did not gain haste")
	}
}

func TestBbdB205TenaciousDeadReturnsTappedWhenYouPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dead := effectstest.B12Push(g, me.ID, "Tenacious Dead", "Creature — Skeleton Warrior", bbdB205DeadOracle, 1, 1)
	effectstest.FillPool(me, 1)
	effectstest.FillPoolColored(me, "B", 1)
	effectstest.KillCreature(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("it did not return")
	}
	if !effectstest.FindCardAnywhere(t, g, dead).Tapped {
		t.Error("it returned untapped")
	}
}

func TestBbdB205TenaciousDeadStaysDeadWhenYouDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dead := effectstest.B12Push(g, me.ID, "Tenacious Dead", "Creature — Skeleton Warrior", bbdB205DeadOracle, 1, 1)
	effectstest.FillPool(me, 1)
	effectstest.FillPoolColored(me, "B", 1)
	effectstest.KillCreature(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dead) || !me.Graveyard.Contains(dead) {
		t.Error("it came back although the cost was declined")
	}
}

func TestBbdB205HuntedWumpusLetsEachOtherPlayerPutACreatureInPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	creature := effectstest.PushCatalogHandCard(opp, "Hand Bear", "Creature — Bear", "")
	effectstest.PushCatalogHandCard(opp, "Hand Spell", "Sorcery", "")
	mine := effectstest.PushCatalogHandCard(me, "My Bear", "Creature — Bear", "")
	effectstest.CastRingCreature(t, g, "Hunted Wumpus", "Creature — Beast", bbdB205WumpusOracle, 6, 6)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, creature)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(creature) {
		t.Error("the opponent's creature card did not enter")
	}
	if g.Battlefield.Contains(mine) {
		t.Error("the Wumpus's controller put a creature in too")
	}
}

// friendFoeAnswers answers the caster's "friend or foe" question for each
// seated player in turn order, naming the seats that are friends.
func bbdB205FriendFoe(t *testing.T, g *game.Game, caster *game.Player, friends map[uuid.UUID]bool) {
	t.Helper()
	for range g.Seats {
		c := effectstest.LatestOptionPickFor(g, caster.ID)
		if c == nil {
			t.Fatalf("no friend-or-foe question: %+v", g.PendingChoices)
		}
		who := c.OptionCarry[1]
		idx := 1
		if friends[who] {
			idx = 0
		}
		effectstest.AnswerOptionPick(t, g, caster.ID, idx)
	}
}

func TestBbdB205RegnasSanctionCountersFriendsAndTapsAllButOneForFoes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	foe := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	friend := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	friendly := effectstest.B12Creature(g, friend.ID, "Friendly", "Creature — Bear", 2, 2)
	keep := effectstest.B12Creature(g, foe.ID, "Keep", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, foe.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.FillPool(me, 3)
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.CastFromHandForTest(t, g, me, "Regna's Sanction", "Sorcery", "{3}{W}", bbdB205SanctionOracle, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	friends := map[uuid.UUID]bool{me.ID: true, friend.ID: true}
	for _, s := range g.Seats {
		if s.ID != me.ID && s.ID != friend.ID && s.ID != foe.ID {
			friends[s.ID] = false
		}
	}
	bbdB205FriendFoe(t, g, me, friends)
	effectstest.AnswerChooseCards(t, g, foe.ID, keep)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, friendly} {
		if got := effectstest.PlusOneCounters(g, id); got != 1 {
			t.Errorf("a friend's creature has %d counters, want 1", got)
		}
	}
	if got := effectstest.PlusOneCounters(g, keep) + effectstest.PlusOneCounters(g, other); got != 0 {
		t.Errorf("a foe's creature got a counter")
	}
	if effectstest.FindCardAnywhere(t, g, keep).Tapped {
		t.Error("the chosen creature was tapped")
	}
	if !effectstest.FindCardAnywhere(t, g, other).Tapped {
		t.Error("the other creature was not tapped")
	}
}

func TestBbdB205KhorvathsFuryWheelsFriendsAndBurnsFoes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	foe := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.FillPool(me, 4)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.CastFromHandForTest(t, g, me, "Khorvath's Fury", "Sorcery", "{4}{R}", bbdB205FuryOracle, game.CastSpellParams{Strict: true})
	handBefore := me.Hand.Size()
	foeHand, foeLife := foe.Hand.Size(), foe.Life
	effectstest.PassPriorityAroundTable(t, g)
	friends := map[uuid.UUID]bool{me.ID: true}
	bbdB205FriendFoe(t, g, me, friends)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("friend's hand %d, want %d (discarded all, drew that many plus one)", got, handBefore+1)
	}
	if want := foeLife - foeHand; foe.Life != want {
		t.Errorf("a foe's life %d, want %d (%d cards in hand)", foe.Life, want, foeHand)
	}
}

func TestBbdB205AzraOddsmakerDrawsTwoWhenTheChosenCreatureConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Azra Oddsmaker", "Creature — Azra Warrior", bbdB205AzraOracle, 3, 3)
	chosen := effectstest.B12Creature(g, me.ID, "Chosen", "Creature — Bear", 2, 2)
	fodder := effectstest.PushCatalogHandCard(me, "Fodder", "Sorcery", "")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.B24OpenPayPrompt(g, me)
	if prompt == nil {
		t.Fatal("no discard prompt")
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, []uuid.UUID{fodder}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, chosen)
	hand := me.Hand.Size()
	effectstest.DealCombatDamage(g, me.ID, chosen, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
}
