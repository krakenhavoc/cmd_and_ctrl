package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b714_cards_test.go — Portal Three Kingdoms, batch 714 (S82, #5622).
// Each test plays the real catalog entry through the engine.

const (
	oldB714MourningOracle   = "31a331a0-3065-4ecb-9e6a-ab11e241833b"
	oldB714DefeatOracle     = "8b119579-38e1-4509-9dcb-019d5a160ac2"
	oldB714SageOracle       = "b8ffd53b-5b10-4866-83fe-630eee58c514"
	oldB714ReturnOracle     = "cf6676b7-b40a-46bd-b166-bd3cbb49d80d"
	oldB714FarmersOracle    = "174cf7cd-3e8c-4f90-abc1-a78a0ce832d2"
	oldB714CaravanOracle    = "587e9f15-1d79-458f-a073-229b828b2393"
	oldB714MarshalOracle    = "f82c628c-c01a-4a1c-af9a-e296ba50f12f"
	oldB714StrikeOracle     = "250f8642-9754-48fd-8f09-70ed13d7a42c"
	oldB714VictoryOracle    = "8df38d95-1707-4f83-9d64-41f0be2a3b3b"
	oldB714WuSpyOracle      = "4b681770-eea3-459c-ad5b-138834213f8c"
	oldB714AssassinsOracle  = "e87dd341-416c-4e93-8eb4-1519ecd2674c"
	oldB714ZuoCiOracle      = "ae8773a3-05f2-4074-9a53-033b0c127235"
	oldB714CouncilOracle    = "ff427657-173a-4845-89af-6f0a93467130"
	oldB714CreatureTypeLine = "Creature — Human Soldier"
)

func TestOldB714EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB714MourningOracle: "False Mourning", oldB714DefeatOracle: "False Defeat",
		oldB714SageOracle: "Sage's Knowledge", oldB714ReturnOracle: "Return to Battle",
		oldB714FarmersOracle: "Shu Soldier-Farmers", oldB714CaravanOracle: "Shu Grain Caravan",
		oldB714MarshalOracle: "Marshaling the Troops", oldB714StrikeOracle: "Preemptive Strike",
		oldB714VictoryOracle: "Champion's Victory", oldB714WuSpyOracle: "Wu Spy",
		oldB714AssassinsOracle: "Wei Assassins", oldB714ZuoCiOracle: "Zuo Ci, the Mocking Sage",
		oldB714CouncilOracle: "Council of Advisors",
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

func oldB714Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle, mana string, targets ...game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	effectstest.FloatForTest(g, p, mana)
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	})
	return id, g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

func oldB714Card(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

func TestOldB714FalseMourningPutsYourOwnCardOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.PushGraveyardCardTyped(me, "Mine", "Sorcery")
	theirs := effectstest.PushGraveyardCardTyped(opp, "Theirs", "Sorcery")
	if _, err := oldB714Cast(t, g, me, "False Mourning", "Sorcery", "{G}", oldB714MourningOracle, "G", oldB714Card(theirs)); err == nil {
		t.Fatal("False Mourning targeted another player's graveyard")
	}
	if _, err := oldB714Cast(t, g, me, "False Mourning", "Sorcery", "{G}", oldB714MourningOracle, "G", oldB714Card(mine)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Library.Contains(mine) || me.Graveyard.Contains(mine) {
		t.Error("the card did not move to the library")
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != mine && me.Library.Cards[0].InstanceID != mine {
		t.Error("the card is not on an end of the library")
	}
}

func TestOldB714FalseDefeatReturnsACreatureToTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	spell := effectstest.PushGraveyardCardTyped(me, "Spell", "Sorcery")
	if _, err := oldB714Cast(t, g, me, "False Defeat", "Sorcery", "{3}{W}", oldB714DefeatOracle, "CCCW", oldB714Card(spell)); err == nil {
		t.Fatal("False Defeat targeted a noncreature card")
	}
	if _, err := oldB714Cast(t, g, me, "False Defeat", "Sorcery", "{3}{W}", oldB714DefeatOracle, "CCCW", oldB714Card(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Error("the creature did not return to the battlefield")
	}
}

func TestOldB714SagesKnowledgeReturnsOnlyASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sorcery := effectstest.PushGraveyardCardTyped(me, "Sorcery", "Sorcery")
	instant := effectstest.PushGraveyardCardTyped(me, "Instant", "Instant")
	if _, err := oldB714Cast(t, g, me, "Sage's Knowledge", "Sorcery", "{2}{U}", oldB714SageOracle, "CCU", oldB714Card(instant)); err == nil {
		t.Fatal("Sage's Knowledge targeted an instant")
	}
	if _, err := oldB714Cast(t, g, me, "Sage's Knowledge", "Sorcery", "{2}{U}", oldB714SageOracle, "CCU", oldB714Card(sorcery)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(sorcery) || !me.Graveyard.Contains(instant) {
		t.Error("the sorcery did not return to hand, or the instant moved")
	}
}

func TestOldB714ReturnToBattleReturnsACreatureToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	spell := effectstest.PushGraveyardCardTyped(me, "Spell", "Sorcery")
	if _, err := oldB714Cast(t, g, me, "Return to Battle", "Sorcery", "{B}", oldB714ReturnOracle, "B", oldB714Card(spell)); err == nil {
		t.Fatal("Return to Battle targeted a noncreature card")
	}
	if _, err := oldB714Cast(t, g, me, "Return to Battle", "Sorcery", "{B}", oldB714ReturnOracle, "B", oldB714Card(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature did not return to hand")
	}
}

func TestOldB714EntersLifeAndDrawCreatures(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		life         int
	}{
		{"Shu Soldier-Farmers", oldB714FarmersOracle, 4},
		{"Shu Grain Caravan", oldB714CaravanOracle, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			before := me.Life
			effectstest.CastHoldCreature(t, g, tc.name, oldB714CreatureTypeLine, tc.oracle, 2, 2)
			if me.Life != before+tc.life {
				t.Errorf("life %d, want %d", me.Life, before+tc.life)
			}
		})
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibraryTop(me, "Filler", "Sorcery", "{1}")
	before := effectstest.HandCount(g, me)
	effectstest.CastHoldCreature(t, g, "Council of Advisors", "Creature — Human Advisor", oldB714CouncilOracle, 1, 1)
	if got := effectstest.HandCount(g, me); got != before+1 {
		t.Errorf("hand %d, want %d (the cast nets zero, one drawn)", got, before+1)
	}
}

func TestOldB714MarshalingTheTroopsGainsFourForEachCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	before := me.Life
	if _, err := oldB714Cast(t, g, me, "Marshaling the Troops", "Sorcery", "{1}{G}", oldB714MarshalOracle, "CG"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedOnBattlefield(t, g, a) || !effectstest.TappedOnBattlefield(t, g, b) || effectstest.TappedOnBattlefield(t, g, c) {
		t.Error("the wrong creatures were tapped")
	}
	if me.Life != before+8 {
		t.Errorf("life %d, want %d", me.Life, before+8)
	}
}

func TestOldB714MarshalingTheTroopsMayTapNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	before := me.Life
	if _, err := oldB714Cast(t, g, me, "Marshaling the Troops", "Sorcery", "{1}{G}", oldB714MarshalOracle, "CG"); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("life %d, want %d", me.Life, before)
	}
}

func TestOldB714PreemptiveStrikeCountersOnlyACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[g.Turn.ActiveSeat]
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	sorcery := effectstest.CastFromHandForTest(t, g, opp, "Spell", "Sorcery", "", "", game.CastSpellParams{})
	if _, err := oldB714Cast(t, g, me, "Preemptive Strike", "Instant", "{1}{U}", oldB714StrikeOracle, "CU", oldB714Card(sorcery)); err == nil {
		t.Fatal("Preemptive Strike targeted a noncreature spell")
	}
	effectstest.PassPriorityAroundTable(t, g)
	bear := effectstest.CastFromHandForTest(t, g, opp, "Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	if _, err := oldB714Cast(t, g, me, "Preemptive Strike", "Instant", "{1}{U}", oldB714StrikeOracle, "CU", oldB714Card(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature spell was not countered")
	}
}

func TestOldB714ChampionsVictoryBouncesAnAttackerAfterYouWereAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.SfPush(g, me.ID, "Attacker", "Creature — Test", 3, []string{"R"})
	home := effectstest.SfPush(g, me.ID, "Home", "Creature — Test", 1, []string{"R"})
	if _, err := oldB714Cast(t, g, opp, "Champion's Victory", "Instant", "{U}", oldB714VictoryOracle, "U", oldB714Card(atk)); err == nil {
		t.Fatal("cast outside the declare attackers step")
	}
	effectstest.SfAttackHold(t, g, opp.ID, atk)
	if _, err := oldB714Cast(t, g, opp, "Champion's Victory", "Instant", "{U}", oldB714VictoryOracle, "U", oldB714Card(home)); err == nil {
		t.Fatal("targeted a creature that isn't attacking")
	}
	if _, err := oldB714Cast(t, g, g.Seats[2], "Champion's Victory", "Instant", "{U}", oldB714VictoryOracle, "U", oldB714Card(atk)); err == nil {
		t.Fatal("cast by a player nobody attacked")
	}
	if _, err := oldB714Cast(t, g, opp, "Champion's Victory", "Instant", "{U}", oldB714VictoryOracle, "U", oldB714Card(atk)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(atk) || !me.Hand.Contains(atk) {
		t.Error("the attacker was not returned to its owner's hand")
	}
}

func TestOldB714WuSpyPutsOneOfTheTopTwoIntoTheirGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	second := effectstest.SsCard(opp.Library, opp.ID, "Second", "Sorcery", 0, 0)
	top := effectstest.SsCard(opp.Library, opp.ID, "Top", "Sorcery", 0, 0)
	effectstest.CastHoldCreature(t, g, "Wu Spy", "Creature — Human Soldier Rogue", oldB714WuSpyOracle, 1, 1)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if prompt == nil {
		t.Fatal("the controller was not asked which card to put into the graveyard")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{second}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(second) || opp.Library.Contains(second) {
		t.Error("the chosen card is not in their graveyard")
	}
	if !opp.Library.Contains(top) || opp.Graveyard.Contains(top) {
		t.Error("the card that was not chosen left the library")
	}
}

func TestOldB714WeiAssassinsLetsTheOpponentChooseWhichCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	keep := effectstest.B12Creature(g, opp.ID, "Keep", "Creature — Bear", 2, 2)
	lose := effectstest.B12Creature(g, opp.ID, "Lose", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Wei Assassins", "Creature — Human Soldier Assassin", oldB714AssassinsOracle, 3, 2)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opp.ID)
	if prompt == nil {
		t.Fatal("the opponent was not asked which creature to lose")
	}
	if err := g.ResolveChooseCards(prompt.ID, opp.ID, []uuid.UUID{lose}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lose) {
		t.Error("the chosen creature survived")
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("the creature that was not chosen was destroyed")
	}
}

func TestOldB714WeiAssassinsADoomedLoneCreatureIsForced(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	only := effectstest.B12Creature(g, opp.ID, "Only", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Wei Assassins", "Creature — Human Soldier Assassin", oldB714AssassinsOracle, 3, 2)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(only) {
		t.Error("the lone creature survived")
	}
}

func TestOldB714ZuoCiHasHexproofAndCantBeBlockedByHorsemanship(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zuo := effectstest.B12Push(g, me.ID, "Zuo Ci, the Mocking Sage", "Legendary Creature — Human Advisor", oldB714ZuoCiOracle, 1, 2)
	if !effectstest.HasEffectiveKeyword(t, g, zuo, "hexproof") {
		t.Error("no hexproof")
	}
	rider := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Rider", TypeLine: "Creature — Soldier", Keywords: []string{"horsemanship"},
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	foot := effectstest.B12Creature(g, opp.ID, "Foot", "Creature — Soldier", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(zuo, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(rider, zuo); err == nil {
		t.Error("a creature with horsemanship blocked Zuo Ci")
	}
	if err := g.DeclareBlocker(foot, zuo); err != nil {
		t.Errorf("a creature without horsemanship could not block: %v", err)
	}
}
