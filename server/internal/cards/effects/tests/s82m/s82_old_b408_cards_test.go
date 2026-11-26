package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b408_cards_test.go — Gatecrash, batch 408 (S82 older sets, #5565):
// ten cards of life gain and loss. Every test plays the real catalog entry
// through the engine.

const (
	oldB408Skirmisher = "29e07d57-0f05-441c-844f-cf1d71102b2d"
	oldB408Parasite   = "577eb51d-7979-4f97-a8b1-e31242720611"
	oldB408Obzedat    = "6c3df703-e96a-4ee4-b449-b8e1e3460bd3"
	oldB408Devour     = "6c40f825-836e-4b76-8e43-267b4abf013b"
	oldB408Rapport    = "7121b877-d262-4627-acd8-a159f8a779b1"
	oldB408Confessor  = "9db7cb7f-1309-4f1d-9c9d-423c91c0896c"
	oldB408Purge      = "a0f29566-bca8-4fe8-86ce-4a051ece6d5a"
	oldB408AlmsBeast  = "b221939c-93e7-4ffc-b995-5b67e173ee15"
	oldB408Luminate   = "c1bc819a-f291-4b55-94f7-269f160ce76b"
	oldB408Guildmage  = "ec44ca05-6b10-4be2-b4df-967535967fce"

	oldB408Bear = "Creature — Bear"
)

var oldB408Cards = map[string]string{
	oldB408Skirmisher: "Angelic Skirmisher", oldB408Parasite: "Thrull Parasite", oldB408Obzedat: "Obzedat, Ghost Council",
	oldB408Devour: "Devour Flesh", oldB408Rapport: "Predator's Rapport", oldB408Confessor: "Vizkopa Confessor",
	oldB408Purge: "Purge the Profane", oldB408AlmsBeast: "Alms Beast", oldB408Luminate: "Luminate Primordial",
	oldB408Guildmage: "Duskmantle Guildmage",
}

func TestOldB408EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB408Cards {
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

func TestOldB408AngelicSkirmisherGivesTheChosenAbilityToYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skirmisher := effectstest.B12Push(g, me.ID, "Angelic Skirmisher", "Creature — Angel", oldB408Skirmisher, 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB408Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB408Bear, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, skirmisher, "flying") {
		t.Error("the Skirmisher has no flying")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BlbSTPickOption(t, g, me.ID, "Lifelink")
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{skirmisher, mine} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "lifelink") {
			t.Error("a creature I control did not gain lifelink")
		}
		if effectstest.HasEffectiveKeyword(t, g, id, "vigilance") {
			t.Error("a creature gained an ability that was not chosen")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "lifelink") {
		t.Error("an opponent's creature gained lifelink")
	}
}

func TestOldB408ThrullParasiteRemovesACounterForTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	parasite := effectstest.B12Push(g, me.ID, "Thrull Parasite", "Creature — Thrull", oldB408Parasite, 1, 1)
	target := effectstest.B12Creature(g, me.ID, "Target", oldB408Bear, 2, 2)
	effectstest.PutCounters(t, g, target, game.CounterPlusOne, 2)
	land := effectstest.B12Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(land)}); err == nil {
		t.Error("an unrelated permanent's ability was activated as the Parasite's")
	}
	life := effectstest.LifeOf(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, parasite, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(target)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, target, game.CounterPlusOne); got != 1 {
		t.Errorf("counters = %d, want 1", got)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life-2 {
		t.Errorf("life %d → %d, want -2", life, got)
	}
	if !effectstest.TappedNow(g, parasite) {
		t.Error("the Parasite was not tapped")
	}
}

func TestOldB408ThrullParasiteAsksWhichKindWhenThereAreSeveral(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	parasite := effectstest.B12Push(g, me.ID, "Thrull Parasite", "Creature — Thrull", oldB408Parasite, 1, 1)
	target := effectstest.B12Creature(g, me.ID, "Target", oldB408Bear, 2, 2)
	effectstest.PutCounters(t, g, target, game.CounterPlusOne, 1)
	effectstest.PutCounters(t, g, target, "charge", 1)
	if err := g.ActivateCatalogAbility(me.ID, parasite, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(target)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BlbSTPickOption(t, g, me.ID, "charge")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, target, "charge") != 0 || effectstest.CounterOn(g, target, game.CounterPlusOne) != 1 {
		t.Errorf("counters now +1/+1 %d, charge %d; want 1 and 0",
			effectstest.CounterOn(g, target, game.CounterPlusOne), effectstest.CounterOn(g, target, "charge"))
	}
}

func TestOldB408ObzedatDrainsOnEntryAndLeavesAndReturnsWithHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	life, mine := effectstest.LifeOf(g, opp.ID), effectstest.LifeOf(g, me.ID)
	effectstest.CastHoldCreature(t, g, "Obzedat, Ghost Council", "Legendary Creature — Spirit Advisor", oldB408Obzedat, 5, 5)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("opponent life %d → %d, want -2", life, got)
	}
	if got := effectstest.LifeOf(g, me.ID); got != mine+2 {
		t.Errorf("my life %d → %d, want +2", mine, got)
	}
	var obzedat uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == oldB408Obzedat {
			obzedat = c.InstanceID
		}
	}
	if obzedat == uuid.Nil {
		t.Fatal("Obzedat is not on the battlefield")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(obzedat) || !g.Exile.Contains(obzedat) {
		t.Fatal("Obzedat was not exiled at my end step")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, me.Seat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	back := uuid.Nil
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == oldB408Obzedat {
			back = c.InstanceID
		}
	}
	if back == uuid.Nil {
		t.Fatal("Obzedat did not return at my next upkeep")
	}
	if !effectstest.HasEffectiveKeyword(t, g, back, "haste") {
		t.Error("the returned Obzedat has no haste")
	}
}

func TestOldB408DevourFleshMakesThemSacrificeAndGainItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := effectstest.B12Creature(g, opp.ID, "Big", oldB408Bear, 1, 5)
	life, mine := effectstest.LifeOf(g, opp.ID), effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Devour Flesh", "Instant", oldB408Devour, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SacrificeChoiceFor(g, opp.ID); c != nil {
		if err := g.ResolveSacrificeChoice(c.ID, opp.ID, big); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(big) {
		t.Error("the creature was not sacrificed")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life+5 {
		t.Errorf("the target player's life %d → %d, want +5 (the sacrificed creature's toughness)", life, got)
	}
	if got := effectstest.LifeOf(g, me.ID); got != mine {
		t.Errorf("the caster's life changed to %d", got)
	}
}

func TestOldB408PredatorsRapportGainsPowerPlusToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB408Bear, 3, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB408Bear, 2, 2)
	hand := effectstest.HandCardFull(me, "Predator's Rapport", "Instant", "", oldB408Rapport, nil)
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(theirs)}); err == nil {
		t.Fatal("an opponent's creature was accepted")
	}
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Predator's Rapport", "Instant", oldB408Rapport, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+7 {
		t.Errorf("life %d → %d, want +7", life, got)
	}
}

func TestOldB408VizkopaConfessorPaysLifeAndExilesOneOfThatManyCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp.Hand.Cards = nil
	var hand []uuid.UUID
	for i := 0; i < 4; i++ {
		id := uuid.New()
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Their Card", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
		hand = append(hand, id)
	}
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastHoldCreature(t, g, "Vizkopa Confessor", "Creature — Human Cleric", oldB408Confessor, 1, 3)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life-2 {
		t.Errorf("life %d → %d, want -2", life, got)
	}
	ask := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if ask == nil || ask.ChooseMin != 2 || ask.ChooseMax != 2 {
		t.Fatalf("the opponent should reveal exactly 2 cards, got %+v", ask)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, hand[0], hand[1])
	rp := effectstest.RevealPickFor(g, me.ID)
	if rp == nil || len(rp.ChooseCards) != 2 {
		t.Fatalf("the controller should choose among 2 revealed cards, got %+v", rp)
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{hand[1]}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if !g.Exile.Contains(hand[1]) || opp.Hand.Size() != 3 {
		t.Errorf("the chosen card was not exiled (hand %d)", opp.Hand.Size())
	}
}

func TestOldB408PurgeTheProfaneDiscardsTwoAndGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.HandCardFull(opp, "A", "Instant", "", "", nil)
	b := effectstest.HandCardFull(opp, "B", "Instant", "", "", nil)
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Purge the Profane", "Sorcery", oldB408Purge, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(a) || !opp.Graveyard.Contains(b) {
		t.Error("the opponent did not discard both cards")
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+2 {
		t.Errorf("life %d → %d, want +2", life, got)
	}
}

func TestOldB408PurgeTheProfaneStillGainsAgainstAnEmptyHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp.Hand.Cards = nil
	life := effectstest.LifeOf(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Purge the Profane", "Sorcery", oldB408Purge, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+2 {
		t.Errorf("life %d → %d, want +2", life, got)
	}
}

func TestOldB408AlmsBeastGivesLifelinkToItsCombatPartners(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	beast := effectstest.B12Push(g, me.ID, "Alms Beast", "Creature — Beast", oldB408AlmsBeast, 6, 6)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB408Bear, 2, 9)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", oldB408Bear, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, blocker, "lifelink") {
		t.Fatal("lifelink before any combat")
	}
	effectstest.DeclareAttack(t, g, opp.ID, beast)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, beast); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, blocker, "lifelink") {
		t.Error("the creature blocking the Beast has no lifelink")
	}
	if effectstest.HasEffectiveKeyword(t, g, bystander, "lifelink") {
		t.Error("a creature that is not in combat with the Beast has lifelink")
	}
}

func TestOldB408LuminatePrimordialExilesOnePerOpponentAndTheyGainItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)], g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	big := effectstest.B12Creature(g, opp.ID, "Big", oldB408Bear, 5, 5)
	small := effectstest.B12Creature(g, third.ID, "Small", oldB408Bear, 2, 2)
	oppLife, thirdLife := effectstest.LifeOf(g, opp.ID), effectstest.LifeOf(g, third.ID)
	effectstest.CastHoldCreature(t, g, "Luminate Primordial", "Creature — Avatar", oldB408Luminate, 4, 7)
	for _, pick := range []uuid.UUID{big, small} {
		ch := effectstest.PickTargetPrompt(t, g)
		if err := g.ResolvePickTargets(ch.ID, ch.Chooser, effectstest.CardTarget(pick)); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		ch := effectstest.PickTargetPrompt(t, g)
		if err := g.ResolvePickTargets(ch.ID, ch.Chooser, nil); err != nil {
			t.Fatalf("declining the third opponent's clause: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || g.Battlefield.Contains(small) {
		t.Fatal("a targeted creature was not exiled")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != oppLife+5 {
		t.Errorf("first opponent's life %d → %d, want +5", oppLife, got)
	}
	if got := effectstest.LifeOf(g, third.ID); got != thirdLife+2 {
		t.Errorf("second opponent's life %d → %d, want +2", thirdLife, got)
	}
}

func TestOldB408DuskmantleGuildmageDrainsForEachCardThatReachesAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	guildmage := effectstest.B12Push(g, me.ID, "Duskmantle Guildmage", "Creature — Human Wizard", oldB408Guildmage, 2, 2)
	effectstest.SeedLibrary(opp, "A", "B", "C", "D")
	effectstest.FloatMana(t, g, me, "{C}{U}{B}")
	if err := g.ActivateCatalogAbility(me.ID, guildmage, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the first ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.FloatMana(t, g, me, "{C}{C}{U}{B}")
	if err := g.ActivateCatalogAbility(me.ID, guildmage, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.PlayerRefs(opp.ID)}); err != nil {
		t.Fatalf("activate the mill: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("two cards milled: opponent life %d → %d, want -2", life, got)
	}
	mine := effectstest.LifeOf(g, me.ID)
	effectstest.P1Destroy(t, g, effectstest.B12Creature(g, me.ID, "Mine", oldB408Bear, 1, 1))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != mine {
		t.Errorf("a card in my own graveyard cost me life: %d → %d", mine, got)
	}
}
