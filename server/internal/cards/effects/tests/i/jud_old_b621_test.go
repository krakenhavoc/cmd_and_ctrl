package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jud_old_b621_test.go — S82 OLD batch 621 (#5611): Judgment, graveyard
// recursion. Each test plays the real catalog card through the engine.

const (
	oldB621Grip       = "0144d9bb-d3cb-4c23-b133-6653d7c3a76d"
	oldB621Tracer     = "4fd567cb-0b8e-41dd-a2b7-87a196de294a"
	oldB621Disdain    = "68f330ca-8279-46c5-900d-a9b443676104"
	oldB621Grave      = "435c8d98-9fa3-40d8-9b7c-32474099e08e"
	oldB621Mythmaker  = "591ce7dc-8407-43ed-acfc-e3d8fd547125"
	oldB621Pulsemage  = "06e4355b-02d4-4238-a6e6-b03a5773346c"
	oldB621Exorcist   = "78e875b1-e621-4bf0-ad81-24ed9e2b43f8"
	oldB621ArcaneFlig = "40eee955-bec6-455a-addf-68cdb84c76b9"
	oldB621GrizzlyFat = "f3073e4f-ca7e-4fde-88a4-bd7b6ebfbc8c"
	oldB621DefyGrav   = "5a295dc1-7f60-4581-b0f9-61e8b14339bf"
)

func TestOldB621EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB621Grip: "Grip of Amnesia", oldB621Tracer: "Nantuko Tracer", oldB621Disdain: "Laquatus's Disdain",
		oldB621Grave: "Grave Consequences", oldB621Mythmaker: "Nomad Mythmaker",
		oldB621Pulsemage: "Pulsemage Advocate", oldB621Exorcist: "Selfless Exorcist",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB621Activate(t *testing.T, g *game.Game, me *game.Player, source uuid.UUID, targets ...game.TargetRef) {
	t.Helper()
	effectstest.B16Activate(t, g, me.ID, source, 0, game.ActivateAbilityParams{Targets: targets})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB621NantukoTracerPutsACardOnTheBottomOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	theirs := effectstest.PushGraveyardCardTyped(opp, "Theirs", "Creature — Bear")
	lib := opp.Library.Size()
	effectstest.B12PlayFromHand(t, g, "Nantuko Tracer", "Creature — Insect Druid", oldB621Tracer, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(theirs) || opp.Library.Size() != lib+1 {
		t.Fatalf("the card is in the graveyard=%v, library %d (want %d)", opp.Graveyard.Contains(theirs), opp.Library.Size(), lib+1)
	}
	if me.Library.Contains(theirs) {
		t.Error("the card went to the wrong library")
	}
}

func TestOldB621NantukoTracerDeclinedLeavesTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := effectstest.BlbB1Seats(g)
	mine := effectstest.PushGraveyardCardTyped(me, "Mine", "Creature — Bear")
	effectstest.B12PlayFromHand(t, g, "Nantuko Tracer", "Creature — Insect Druid", oldB621Tracer, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(mine) {
		t.Error("a declined Tracer moved the card")
	}
}

func oldB621CastFate(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return effectstest.CastCatalogSpell(t, g, "Grizzly Fate", "Sorcery", oldB621GrizzlyFat, nil)
}

func oldB621Counter(t *testing.T, g *game.Game, opp *game.Player, name, oracle string, target uuid.UUID) {
	t.Helper()
	id := effectstest.PushCatalogHandCard(opp, name, "Instant", oracle)
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: effectstest.CardTarget(target)}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
}

func TestOldB621GripOfAmnesiaCountersUnlessTheControllerExilesTheirWholeGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbMPFillGraveyard(g, me, 3)
	fate := oldB621CastFate(t, g)
	hand := opp.Hand.Size()
	oldB621Counter(t, g, opp, "Grip of Amnesia", oldB621Grip, fate)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.AnswerPayUnless(t, g, me.ID, false)
	if c == nil {
		t.Fatal("the spell's controller was not asked")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Bear"); got != 0 {
		t.Errorf("a spell whose controller declined made %d Bears", got)
	}
	if !me.Graveyard.Contains(fate) {
		t.Error("the countered spell is not in its owner's graveyard")
	}
	if opp.Hand.Size() != hand+1 {
		t.Errorf("Grip's controller hand %d, want %d (the draw)", opp.Hand.Size(), hand+1)
	}
}

func TestOldB621GripOfAmnesiaPaidExilesEveryCardAndTheSpellResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	var ids []uuid.UUID
	for _, n := range []string{"A", "B", "C"} {
		ids = append(ids, effectstest.PushGraveyardCardTyped(me, n, "Creature — Bear"))
	}
	fate := oldB621CastFate(t, g)
	oldB621Counter(t, g, opp, "Grip of Amnesia", oldB621Grip, fate)
	effectstest.PassPriorityAroundTable(t, g)
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == me.ID {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatal("no pay prompt")
	}
	if err := g.ResolvePayUnlessWithCards(prompt.ID, me.ID, true, ids); err != nil {
		t.Fatalf("pay: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids {
		if me.Graveyard.Contains(id) {
			t.Error("a card is still in the graveyard")
		}
	}
	if got := effectstest.B43TokensNamed(g, me.ID, "Bear"); got != 2 {
		t.Errorf("the spell made %d Bears, want 2", got)
	}
}

func TestOldB621GripOfAmnesiaAnEmptyGraveyardPaysForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	fate := oldB621CastFate(t, g)
	oldB621Counter(t, g, opp, "Grip of Amnesia", oldB621Grip, fate)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.B43TokensNamed(g, me.ID, "Bear"); got != 2 {
		t.Errorf("a spell whose controller has no graveyard to exile made %d Bears, want 2", got)
	}
}

func TestOldB621LaquatussDisdainOnlyTargetsASpellCastFromAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	fate := oldB621CastFate(t, g)
	id := effectstest.PushCatalogHandCard(opp, "Laquatus's Disdain", "Instant", oldB621Disdain)
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(fate)}); err == nil {
		t.Fatal("a spell cast from hand was a legal target")
	}
	effectstest.PassPriorityAroundTable(t, g)

	defy := effectstest.SeedGraveyardCard(t, g, "Defy Gravity", "Instant", oldB621DefyGrav)
	bear := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.DcMana(t, g, me.ID, effectstest.OdyB1Pool("{U}"))
	if err := g.CastSpell(me.ID, defy, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	hand := opp.Hand.Size()
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(defy)}); err != nil {
		t.Fatalf("a flashback spell was refused as a target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the countered flashback spell still gave flying")
	}
	if opp.Hand.Size() != hand {
		// the Disdain left the hand (-1) and drew a card (+1)
		t.Errorf("hand %d, want %d after casting and drawing", opp.Hand.Size(), hand)
	}
}

func TestOldB621GraveConsequencesLosesLifePerCardLeftAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mine := []uuid.UUID{
		effectstest.PushGraveyardCardTyped(me, "M1", "Creature — Bear"),
		effectstest.PushGraveyardCardTyped(me, "M2", "Creature — Bear"),
		effectstest.PushGraveyardCardTyped(me, "M3", "Creature — Bear"),
	}
	theirs := []uuid.UUID{
		effectstest.PushGraveyardCardTyped(opp, "T1", "Creature — Bear"),
		effectstest.PushGraveyardCardTyped(opp, "T2", "Creature — Bear"),
	}
	myLife, oppLife, hand := me.Life, opp.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Grave Consequences", "Instant", oldB621Grave, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// APNAP: the active player (me) first, then the opponent. I exile two.
	effectstest.AnswerChooseCards(t, g, me.ID, mine[0], mine[1])
	effectstest.AnswerChooseCards(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(mine[0]) || me.Graveyard.Contains(mine[1]) || !me.Graveyard.Contains(mine[2]) {
		t.Error("the two picked cards should be exiled and the third left")
	}
	for _, id := range theirs {
		if !opp.Graveyard.Contains(id) {
			t.Error("the opponent declined, but a card left their graveyard")
		}
	}
	// The spell is still on the stack as it resolves, so it is not counted.
	if got := myLife - me.Life; got != 1 {
		t.Errorf("I lost %d life, want 1 (one card left)", got)
	}
	if got := oppLife - opp.Life; got != 2 {
		t.Errorf("the opponent lost %d life, want 2", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestOldB621NomadMythmakerReturnsAnAuraOntoYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	mythmaker := effectstest.BlbB1Push(g, me.ID, "Nomad Mythmaker", "Creature — Human Nomad Cleric", oldB621Mythmaker, 2, 2)
	mine := effectstest.BlbB1Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.BlbB1Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	aura := effectstest.PushGraveyardCardTyped(opp, "Arcane Flight", "Enchantment — Aura")
	setOracle(g, opp, aura, oldB621ArcaneFlig)
	effectstest.B06AddMana(me, "W")
	oldB621Activate(t, g, me, mythmaker, effectstest.CardTarget(aura)...)
	if p := effectstest.AuraHostChoiceFor(g, me.ID); p != nil {
		for _, c := range p.PickTargetCards {
			if c == theirs {
				t.Error("an opponent's creature was offered as the host")
			}
		}
		effectstest.AnswerAuraHost(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine})
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not return to the battlefield")
	}
	if c, ok := effectstest.BattlefieldCard(g, aura); !ok || c.Controller != me.ID {
		t.Errorf("the Aura's controller is not me: %+v", c.Controller)
	}
	if p := effectstest.EffectivePower(t, g, mine); p != 3 {
		t.Errorf("the enchanted Bear has power %d, want 3", p)
	}
}

func TestOldB621PulsemageAdvocateReturnsThreeCardsAndReanimates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	adv := effectstest.BlbB1Push(g, me.ID, "Pulsemage Advocate", "Creature — Human Cleric", oldB621Pulsemage, 1, 3)
	var theirs []uuid.UUID
	for _, n := range []string{"T1", "T2", "T3", "T4"} {
		theirs = append(theirs, effectstest.PushGraveyardCardTyped(opp, n, "Creature — Bear"))
	}
	mine := effectstest.PushGraveyardCardTyped(me, "Mine", "Creature — Bear")
	hand := opp.Hand.Size()
	oldB621Activate(t, g, me, adv,
		game.TargetRef{Kind: game.TargetCard, ID: theirs[0], Slot: 0},
		game.TargetRef{Kind: game.TargetCard, ID: theirs[1], Slot: 0},
		game.TargetRef{Kind: game.TargetCard, ID: theirs[2], Slot: 0},
		game.TargetRef{Kind: game.TargetCard, ID: mine, Slot: 1})
	if got := opp.Hand.Size(); got != hand+3 {
		t.Errorf("the opponent's hand is %d, want %d", got, hand+3)
	}
	if !opp.Graveyard.Contains(theirs[3]) {
		t.Error("an untargeted card left the graveyard")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("the creature card did not return to the battlefield")
	}
}

func TestOldB621PulsemageAdvocateNeedsThreeCardsFromOneGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	adv := effectstest.BlbB1Push(g, me.ID, "Pulsemage Advocate", "Creature — Human Cleric", oldB621Pulsemage, 1, 3)
	a1 := effectstest.PushGraveyardCardTyped(opp, "A1", "Creature — Bear")
	a2 := effectstest.PushGraveyardCardTyped(opp, "A2", "Creature — Bear")
	b1 := effectstest.PushGraveyardCardTyped(other, "B1", "Creature — Bear")
	mine := effectstest.PushGraveyardCardTyped(me, "Mine", "Creature — Bear")
	for name, cards := range map[string][]uuid.UUID{"two cards": {a1, a2}, "two graveyards": {a1, a2, b1}} {
		targets := []game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 1}}
		for _, id := range cards {
			targets = append(targets, game.TargetRef{Kind: game.TargetCard, ID: id, Slot: 0})
		}
		if err := g.ActivateCatalogAbility(me.ID, adv, 0, game.ActivateAbilityParams{Targets: targets}); err == nil {
			t.Errorf("%s: the activation was accepted", name)
		}
	}
}

func TestOldB621SelflessExorcistExilesTheCardAndTakesItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	ex := effectstest.BlbB1Push(g, me.ID, "Selfless Exorcist", "Creature — Human Cleric", oldB621Exorcist, 3, 4)
	small := effectstest.PushGraveyardCardTyped(opp, "Small", "Creature — Bear")
	setPT(g, opp, small, 3, 3)
	oldB621Activate(t, g, me, ex, effectstest.CardTarget(small)...)
	if opp.Graveyard.Contains(small) || !g.Exile.Contains(small) {
		t.Fatal("the creature card was not exiled")
	}
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(ex) {
		t.Fatal("the Exorcist died to 3 damage")
	}
	if dmg := effectstest.DamageOn(t, g, ex); dmg != 3 {
		t.Errorf("the Exorcist has %d damage, want 3", dmg)
	}
}

func TestOldB621SelflessExorcistRefusesANonCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	ex := effectstest.BlbB1Push(g, me.ID, "Selfless Exorcist", "Creature — Human Cleric", oldB621Exorcist, 3, 4)
	spell := effectstest.PushGraveyardCardTyped(opp, "Bolt", "Instant")
	if err := g.ActivateCatalogAbility(me.ID, ex, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(spell)}); err == nil {
		t.Error("an instant card was accepted as the target creature card")
	}
}

func setOracle(g *game.Game, p *game.Player, id uuid.UUID, oracle string) {
	for i := range p.Graveyard.Cards {
		if p.Graveyard.Cards[i].InstanceID == id {
			p.Graveyard.Cards[i].OracleID = oracle
		}
	}
}

func setPT(g *game.Game, p *game.Player, id uuid.UUID, power, toughness int) {
	for i := range p.Graveyard.Cards {
		if p.Graveyard.Cards[i].InstanceID == id {
			p.Graveyard.Cards[i].Power, p.Graveyard.Cards[i].Toughness = power, toughness
		}
	}
}
