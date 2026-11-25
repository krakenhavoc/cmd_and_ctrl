package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsco "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/co"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b302_cards_test.go — Commander 2015, batch 302 (S82 older sets,
// #5546). Every test plays the real catalog entry through the engine.

const (
	oldB302DisasterRadius = "3322c865-bb2a-4201-9115-979883cd7894"
	oldB302LoneRevenant   = "5d814b0b-82e8-4145-87e1-4fed44d16465"
	oldB302FallOfHammer   = "7e3fe8e4-e56a-4109-a314-96c44bbe7d22"
	oldB302Cloudthresher  = "ab0e2aa0-1ddd-4a4c-b7cd-4063caaa8bcd"
	oldB302Thundercloud   = "eb812ea7-3301-4774-86a2-64c4ed6766fc"
	oldB302SteamAugury    = "0aa556a6-66aa-42f1-ba41-0001e10c20a4"
	oldB302Blustersquall  = "2dfad8c3-1973-4fc8-971d-f66cacd88070"
	oldB302Dominate       = "77db8db4-9fe4-478e-b919-5348b0b87b0d"
	oldB302CobraTrap      = "f3953f48-01c0-4647-9e4a-d70708d01e33"
)

func TestOldB302EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB302DisasterRadius, oldB302LoneRevenant, oldB302FallOfHammer, oldB302Cloudthresher,
		oldB302Thundercloud, oldB302SteamAugury, oldB302Blustersquall, oldB302Dominate, oldB302CobraTrap} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// oldB302Cast puts the card in the active seat's hand and casts it strictly.
func oldB302Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.D5681Spell(g, me, uuid.New(), name, typeLine, "", oracle)
	var id uuid.UUID
	g.ReadSnapshot(func() { id = me.Hand.Cards[len(me.Hand.Cards)-1].InstanceID })
	params.Strict = true
	return g.CastSpell(me.ID, id, params)
}

func TestOldB302DisasterRadiusDealsTheRevealedManaValueToEachOpposingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, opp.ID, "Their Small", "Creature — Bear", 2, 3)
	big := effectstest.B12Creature(g, opp.ID, "Their Big", "Creature — Bear", 2, 5)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	revealed := game.NewCard("Revealed Beast", me.ID)
	revealed.TypeLine = "Creature — Beast"
	revealed.ManaCost = "{2}{G}{G}"
	me.Hand.PushTop(revealed)
	notCreature := game.NewCard("A Sorcery", me.ID)
	notCreature.TypeLine = "Sorcery"
	me.Hand.PushTop(notCreature)

	if err := oldB302Cast(t, g, "Disaster Radius", "Sorcery", oldB302DisasterRadius, game.CastSpellParams{}); err == nil {
		t.Fatal("cast without revealing a creature card")
	}
	if err := oldB302Cast(t, g, "Disaster Radius", "Sorcery", oldB302DisasterRadius, game.CastSpellParams{RevealIDs: []uuid.UUID{notCreature.InstanceID}}); err == nil {
		t.Fatal("revealed a noncreature card")
	}
	if err := oldB302Cast(t, g, "Disaster Radius", "Sorcery", oldB302DisasterRadius, game.CastSpellParams{RevealIDs: []uuid.UUID{revealed.InstanceID}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 2/3 survived 4 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 2/5 died to 4 damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own creature was hit")
	}
	if !me.Hand.Contains(revealed.InstanceID) {
		t.Error("the revealed card left the hand")
	}
}

func TestOldB302LoneRevenantLooksOnlyWhenItIsYourOnlyCreature(t *testing.T) {
	for _, tc := range []struct {
		name  string
		other bool
		want  bool
	}{{"alone", false, true}, {"with another creature", true, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := g.Seats[0], g.Seats[1]
			ids := effectstest.SeedFactOrFictionLibrary(me, "L1", "L2", "L3", "L4", "L5")
			rev := effectstest.B12Push(g, me.ID, "Lone Revenant", "Creature — Spirit", oldB302LoneRevenant, 4, 4)
			if tc.other {
				effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 1, 1)
			}
			hand := me.Hand.Size()
			effectstest.DealCombatDamageToPlayer(g, rev, opp.ID, 4)
			effectstest.PassPriorityAroundTable(t, g)
			pick := effectstest.LatestChooseCardsFor(g, me.ID)
			if !tc.want {
				if pick != nil || me.Hand.Size() != hand {
					t.Fatalf("triggered with another creature: pick %v hand %d→%d", pick, hand, me.Hand.Size())
				}
				return
			}
			if pick == nil {
				t.Fatalf("no card to pick: %+v", g.PendingChoices)
			}
			if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}
			effectstest.AutoAnswerAll(t, g)
			if !me.Hand.Contains(ids[2]) || me.Hand.Size() != hand+1 {
				t.Errorf("the chosen card is not the only one added to hand: %d→%d", hand, me.Hand.Size())
			}
			if me.Library.Size() < 5 {
				t.Errorf("the rest should be on the bottom of the library")
			}
		})
	}
}

func TestOldB302FallOfTheHammerDealsDamageEqualToItsPowerToAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	biter := effectstest.B12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	other := effectstest.B12Creature(g, opp.ID, "Their Other", "Creature — Bear", 2, 4)
	if err := oldB302Cast(t, g, "Fall of the Hammer", "Instant", oldB302FallOfHammer,
		game.CastSpellParams{Targets: effectstest.CardRefs(biter, biter)}); err == nil {
		t.Fatal("the same creature was both targets")
	}
	if err := oldB302Cast(t, g, "Fall of the Hammer", "Instant", oldB302FallOfHammer,
		game.CastSpellParams{Targets: effectstest.CardRefs(biter, victim)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the target survived 3 damage")
	}
	if !g.Battlefield.Contains(biter) || !g.Battlefield.Contains(other) {
		t.Error("only the second target takes damage")
	}
}

func oldB302Damaged(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.DamageMarked
			}
		}
	})
	return n
}

func TestOldB302CloudthresherDamagesEveryFlyerAndEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	flyer := effectstest.B30Flyer(g, opp.ID, "Their Flyer", "Creature — Bird", 1, 4)
	ground := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 4)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.CastHoldCreature(t, g, "Cloudthresher", "Creature — Elemental", oldB302Cloudthresher, 7, 7)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB302Damaged(g, flyer); got != 2 {
		t.Errorf("the flyer has %d damage, want 2", got)
	}
	if got := oldB302Damaged(g, ground); got != 0 {
		t.Errorf("the ground creature has %d damage, want 0", got)
	}
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-2 {
			t.Errorf("a player is at %d, want %d", p.Life, lives[p.ID]-2)
		}
	}
	_ = me
}

func TestOldB302CloudthresherEvokedStillDealsTheDamageThenIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.FloatMana(t, g, me, "{G}{G}{G}{G}")
	id := effectstest.CastWithAltCost(t, g, "Cloudthresher", "Creature — Elemental", oldB302Cloudthresher, "evoke")
	effectstest.PassUntilOnBattlefield(t, g, id)
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("the evoked creature stayed on the battlefield")
	}
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want -2 even when evoked", life, me.Life)
	}
}

func TestOldB302ThundercloudShamanCountsGiantsAndSparesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.B12Push(g, me.ID, "My Giant", "Creature — Giant Warrior", "", 5, 5)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	theirGiant := effectstest.B12Push(g, opp.ID, "Their Giant", "Creature — Giant", "", 3, 3)
	effectstest.CastHoldCreature(t, g, "Thundercloud Shaman", "Creature — Giant Shaman", oldB302Thundercloud, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	// Two Giants: my Giant and the Shaman. 2 damage to each non-Giant.
	if g.Battlefield.Contains(mine) {
		t.Error("my 2/2 Bear survived 2 damage")
	}
	if !g.Battlefield.Contains(theirs) || oldB302Damaged(g, theirs) != 2 {
		t.Errorf("their 2/3 Bear should survive with 2 damage, has %d", oldB302Damaged(g, theirs))
	}
	if oldB302Damaged(g, giant) != 0 || oldB302Damaged(g, theirGiant) != 0 {
		t.Error("a Giant was damaged")
	}
}

func TestOldB302SteamAuguryGivesTheChosenPileToHandAndTheRestToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	revealed := effectstest.SeedFactOrFictionLibrary(me, "SA One", "SA Two", "SA Three", "SA Four", "SA Five")
	hand, graves := me.Hand.Size(), me.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Steam Augury", "Instant", oldB302SteamAugury, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	split := effectstest.LatestRevealPickFor(g, me.ID)
	if split == nil {
		t.Fatalf("the caster separates the piles: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, revealed[:2]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	pick := effectstest.LatestOptionPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the named opponent chooses a pile: %+v", g.PendingChoices)
	}
	// Take the pile of three: two cards are binned.
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d → %d, want +3", hand, me.Hand.Size())
	}
	if me.Graveyard.Size() != graves+3 {
		t.Errorf("graveyard %d → %d, want +3 (two binned cards and the spell)", graves, me.Graveyard.Size())
	}
}

func TestOldB302BlustersquallTapsOneCreatureOrEachWhenOverloaded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	if err := oldB302Cast(t, g, "Blustersquall", "Instant", oldB302Blustersquall, game.CastSpellParams{Targets: effectstest.CardTarget(mine)}); err == nil {
		t.Fatal("targeted my own creature")
	}
	if err := oldB302Cast(t, g, "Blustersquall", "Instant", oldB302Blustersquall, game.CastSpellParams{Targets: effectstest.CardTarget(a)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, a) || effectstest.B16Tapped(t, g, b) {
		t.Error("only the targeted creature should be tapped")
	}
	effectstest.FloatMana(t, g, me, "{U}{U}{U}{U}")
	if err := oldB302Cast(t, g, "Blustersquall", "Instant", oldB302Blustersquall, game.CastSpellParams{AlternativeCost: "overload"}); err != nil {
		t.Fatalf("overload cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, b) {
		t.Error("overload left an opposing creature untapped")
	}
	if effectstest.B16Tapped(t, g, mine) {
		t.Error("overload tapped my own creature")
	}
}

func TestOldB302DominateTakesACreatureWithManaValueXOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := effectstest.Pr7bCreature(g, opp.ID, "Cheap", 2, "{1}{G}", nil)
	dear := effectstest.Pr7bCreature(g, opp.ID, "Dear", 5, "{3}{G}{G}", nil)
	if err := oldB302Cast(t, g, "Dominate", "Instant", oldB302Dominate, game.CastSpellParams{XValue: 3, Targets: effectstest.CardTarget(dear)}); err == nil {
		t.Fatal("took a mana value 5 creature with X = 3")
	}
	if err := oldB302Cast(t, g, "Dominate", "Instant", oldB302Dominate, game.CastSpellParams{XValue: 3, Targets: effectstest.CardTarget(cheap)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	live := effectstest.JmpB4Live(t, g, cheap)
	if live.Controller != me.ID {
		t.Errorf("controller %v, want me", live.Controller)
	}
}

func TestOldB302CobraTrapIsCheapOnlyAfterAnOpponentDestroyedYourNoncreaturePermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[g.Turn.ActiveSeat]
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "My Rock", TypeLine: "Artifact"})
	golem := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "My Golem", TypeLine: "Artifact Creature — Golem", Power: 2, Toughness: 2})
	if oldB302TrapOffered(t, g, me) {
		t.Fatal("offered before anything was destroyed")
	}
	// My own destroy doesn't count; neither does an opposing creature kill.
	effectstest.CastCatalogSpell(t, g, "Shatter", "Instant", effectstest.LeaA1ShatterOracle, []game.TargetRef{{Kind: game.TargetCard, ID: golem}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) {
		t.Fatal("setup: the golem was not destroyed")
	}
	if oldB302TrapOffered(t, g, me) {
		t.Error("offered after only a creature was destroyed")
	}
	effectstest.CastCatalogSpell(t, g, "Shatter", "Instant", effectstest.LeaA1ShatterOracle, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Fatal("setup: the artifact was not destroyed")
	}
	if !oldB302TrapOffered(t, g, me) {
		t.Error("not offered after an opponent destroyed my artifact")
	}
	if oldB302TrapOffered(t, g, opp) {
		t.Error("offered to the player who did the destroying")
	}
}

// oldB302TrapOffered reads Cobra Trap's trap condition for a player.
func oldB302TrapOffered(t *testing.T, g *game.Game, player *game.Player) bool {
	t.Helper()
	var ok bool
	g.ReadSnapshot(func() { ok = cardsco.ANoncreaturePermanentOfYoursWasDestroyedByAnOpponentThisTurn(g, player.ID) })
	return ok
}

func TestOldB302CobraTrapMakesFourSnakes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := oldB302Cast(t, g, "Cobra Trap", "Instant — Trap", oldB302CobraTrap, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	found := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.Name == "Snake" && c.Controller == me.ID {
			found++
			if c.Power != 1 || c.Toughness != 1 || !c.HasColor("G") {
				t.Errorf("Snake is %d/%d", c.Power, c.Toughness)
			}
		}
	}
	if found != 4 {
		t.Errorf("%d Snake tokens, want 4", found)
	}
}
