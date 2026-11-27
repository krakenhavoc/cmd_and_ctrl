package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b586_cards_test.go — Saviors of Kamigawa, S82 older sets, batch 586
// (#5601). Each test plays the real catalog entry.

const (
	oldB586Oni      = "024ea14f-48b7-485d-8426-317d59f8262d"
	oldB586Kirin    = "31d572b8-2c5e-4819-b2a6-e7ef90872d9e"
	oldB586Choice   = "506714a8-b9c2-443a-a9e8-ef9130e674ba"
	oldB586Measure  = "8f1d65ae-7ce5-4bb6-8c00-eace25ff837c"
	oldB586Netsuke  = "ab151429-0890-4231-a3a7-34e8455dac70"
	oldB586Clock    = "ef987ddb-25f6-4895-8f83-045254745ea4"
	oldB586Spirit   = "Creature — Spirit"
	oldB586Bear     = "Creature — Bear"
	oldB586StartLif = 40
)

func TestOldB586EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB586Oni:     "Raving Oni-Slave",
		oldB586Kirin:   "Bounteous Kirin",
		oldB586Choice:  "Choice of Damnations",
		oldB586Measure: "Measure of Wickedness",
		oldB586Netsuke: "Ivory Crane Netsuke",
		oldB586Clock:   "Blood Clock",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB586PickPlayer(t *testing.T, g *game.Game, chooser, player uuid.UUID) {
	t.Helper()
	pick := effectstest.LatestPickTarget(g, chooser)
	if pick == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(pick.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// oldB586AnswerTrigger answers the "you may" prompt of a trigger.
func oldB586AnswerTrigger(t *testing.T, g *game.Game, who uuid.UUID, apply bool) {
	t.Helper()
	c := effectstest.ConfirmForTriggerPrompt(g, who)
	if c == nil {
		t.Fatalf("no trigger prompt for %s: %+v", who, g.PendingChoices)
	}
	if err := g.ResolveTriggerPrompt(c.ID, who, apply); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
}

func oldB586Library(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.SeedLibrary(p, "Card")
	}
}

// --- Raving Oni-Slave --------------------------------------------------

func TestOldB586RavingOniSlaveLosesLifeOnEnteringAndLeavingUnlessYouControlADemon(t *testing.T) {
	for _, demon := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[0]
		if demon {
			effectstest.PushCatalogPermanent(g, me.ID, "Pit Demon", "Creature — Demon", "", false)
		}
		want := me.Life
		if !demon {
			want -= 3
		}
		oni := effectstest.CastAndResolveCreature(t, g, "Raving Oni-Slave", "Creature — Ogre Warrior", oldB586Oni)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != want {
			t.Fatalf("demon %v: life after entering = %d, want %d", demon, me.Life, want)
		}
		effectstest.KillCreature(t, g, me.ID, oni)
		effectstest.PassPriorityAroundTable(t, g)
		if !demon {
			want -= 3
		}
		if me.Life != want {
			t.Errorf("demon %v: life after leaving = %d, want %d", demon, me.Life, want)
		}
	}
}

// --- Bounteous Kirin ---------------------------------------------------

func TestOldB586BounteousKirinGainsLifeEqualToASpiritOrArcaneSpellsManaValue(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		gain     int
	}{{oldB586Spirit, 3}, {"Instant — Arcane", 3}, {oldB586Bear, 0}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Bounteous Kirin", "Legendary Creature — Kirin Spirit", oldB586Kirin, false)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell", TypeLine: tc.typeLine, ManaCost: "{3}", Owner: me.ID, Controller: me.ID})
		me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("%s: cast: %v", tc.typeLine, err)
		}
		life := me.Life
		effectstest.PassPriorityAroundTable(t, g) // the trigger resolves first
		if tc.gain > 0 {
			oldB586AnswerTrigger(t, g, me.ID, true)
			effectstest.PassPriorityAroundTable(t, g)
		} else if effectstest.ConfirmForTriggerPrompt(g, me.ID) != nil {
			t.Fatalf("%s: the Kirin asked about a spell that is neither Spirit nor Arcane", tc.typeLine)
		}
		if me.Life != life+tc.gain {
			t.Errorf("%s: life %d→%d, want +%d", tc.typeLine, life, me.Life, tc.gain)
		}
	}
}

func TestOldB586BounteousKirinMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Bounteous Kirin", "Legendary Creature — Kirin Spirit", oldB586Kirin, false)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell", TypeLine: oldB586Spirit, ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	oldB586AnswerTrigger(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d→%d after declining", life, me.Life)
	}
}

// --- Choice of Damnations ----------------------------------------------

func oldB586CastChoice(t *testing.T, g *game.Game, me, opp *game.Player, number int) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Choice of Damnations", TypeLine: "Sorcery — Arcane", OracleID: oldB586Choice, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, opp.ID, number)
}

func TestOldB586ChoiceOfDamnationsLifeLossWhenTheCasterAcceptsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	oldB586CastChoice(t, g, me, opp, 7)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if opp.Life != oldB586StartLif-7 {
		t.Errorf("life = %d, want %d", opp.Life, oldB586StartLif-7)
	}
	if !effects.OnBattlefield(g, rock) {
		t.Error("a permanent was sacrificed although the caster took the life loss")
	}
}

func TestOldB586ChoiceOfDamnationsSacrificesAllButTheChosenNumberOfTheirChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	for i := 0; i < 5; i++ {
		ids = append(ids, effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false))
	}
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Mine", "Artifact", "", false)
	oldB586CastChoice(t, g, me, opp, 2)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, opp.ID)
	if pick == nil || pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("keep prompt = %+v, want exactly two to keep", pick)
	}
	if err := g.ResolveOwnPermanents(pick.ID, opp.ID, []uuid.UUID{ids[1], ids[3]}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	for i, id := range ids {
		kept := i == 1 || i == 3
		if effects.OnBattlefield(g, id) != kept {
			t.Errorf("permanent %d on the battlefield = %v, want %v", i, !kept, kept)
		}
	}
	if !effects.OnBattlefield(g, mine) {
		t.Error("the caster's own permanent was sacrificed")
	}
	if opp.Life != oldB586StartLif {
		t.Errorf("life = %d, the sacrifice branch must not cost life", opp.Life)
	}
}

func TestOldB586ChoiceOfDamnationsZeroSacrificesEverythingAndBigKeepsEverything(t *testing.T) {
	for _, tc := range []struct {
		number int
		left   int
	}{{0, 0}, {3, 3}, {9, 3}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		var ids []uuid.UUID
		for i := 0; i < 3; i++ {
			ids = append(ids, effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false))
		}
		oldB586CastChoice(t, g, me, opp, tc.number)
		effectstest.AnswerMayChoice(t, g, me.ID, false)
		left := 0
		for _, id := range ids {
			if effects.OnBattlefield(g, id) {
				left++
			}
		}
		if left != tc.left {
			t.Errorf("number %d: %d permanents left, want %d", tc.number, left, tc.left)
		}
	}
}

// --- Measure of Wickedness ---------------------------------------------

func TestOldB586MeasureOfWickednessSacrificesItselfAndCostsEightLifeAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	m := effectstest.PushCatalogPermanent(g, me.ID, "Measure of Wickedness", "Enchantment", oldB586Measure, false)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, m) {
		t.Error("Measure of Wickedness was not sacrificed")
	}
	if me.Life != oldB586StartLif-8 {
		t.Errorf("life = %d, want %d", me.Life, oldB586StartLif-8)
	}
}

func TestOldB586MeasureOfWickednessChangesHandsOncePerCardPutIntoYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	m := effectstest.PushCatalogPermanent(g, me.ID, "Measure of Wickedness", "Enchantment", oldB586Measure, false)
	// A token dying is not a card, and an opponent's card is not in your graveyard.
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Soldier", TypeLine: "Token Creature — Soldier",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB586Bear, 2, 2)
	effectstest.KillCreature(t, g, me.ID, token)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || controllerOf(g, m) != me.ID {
		t.Fatal("a token, or a card that went to an opponent's graveyard, triggered it")
	}
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB586Bear, 2, 2)
	effectstest.KillCreature(t, g, me.ID, mine)
	oldB586PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := controllerOf(g, m); got != opp.ID {
		t.Fatalf("controller = %v (me %v), want the targeted opponent; stack %d, choices %d", got, me.ID, g.Stack.Size(), len(g.PendingChoices))
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestPickTarget(g, opp.ID) != nil {
		t.Error("a second trigger was queued for the one card")
	}
}

func controllerOf(g *game.Game, id uuid.UUID) uuid.UUID {
	return effectstest.ApaLive(g, id).Controller
}

// --- Ivory Crane Netsuke -----------------------------------------------

func TestOldB586IvoryCraneNetsukeGainsFourWithSevenCardsInHand(t *testing.T) {
	for _, tc := range []struct{ hand, gain int }{{6, 0}, {7, 4}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		oldB586Library(me, 3)
		effectstest.PushCatalogPermanent(g, me.ID, "Ivory Crane Netsuke", "Artifact", oldB586Netsuke, false)
		effectstest.D2dSetHand(me, tc.hand)
		g.Turn.Step = game.StepUntap
		effectstest.AdvanceTo(t, g, game.StepUpkeep)
		life := me.Life
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != life+tc.gain {
			t.Errorf("hand %d: life %d→%d, want +%d", tc.hand, life, me.Life, tc.gain)
		}
	}
}

// --- Blood Clock --------------------------------------------------------

func TestOldB586BloodClockLetsTheUpkeepPlayerPayLifeOrReturnAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB586Library(me, 3)
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Blood Clock", "Artifact", oldB586Clock, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "My Bear", oldB586Bear, "", false)
	g.Turn.Step = game.StepUntap
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	// The active player, not the clock's controller, answers.
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if pick == nil {
		t.Fatalf("no option prompt for the upkeep player: %+v", g.PendingChoices)
	}
	life := me.Life
	if err := g.ResolveOptionPick(pick.ID, me.ID, 1); err != nil { // pay 2 life
		t.Fatalf("pay: %v", err)
	}
	if me.Life != life-2 || !effects.OnBattlefield(g, bear) {
		t.Fatalf("paying: life %d→%d, bear on battlefield %v", life, me.Life, effects.OnBattlefield(g, bear))
	}
}

func TestOldB586BloodClockReturnsAChosenPermanentWhenTheLifeIsNotPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB586Library(me, 3)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "My Bear", oldB586Bear, "", false)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "My Rock", "Artifact", "", false)
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Blood Clock", "Artifact", oldB586Clock, false)
	g.Turn.Step = game.StepUntap
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if pick == nil {
		t.Fatal("no option prompt")
	}
	life := me.Life
	if err := g.ResolveOptionPick(pick.ID, me.ID, 0); err != nil { // return a permanent
		t.Fatalf("return: %v", err)
	}
	own := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if own == nil || own.ChooseMin != 1 || own.ChooseMax != 1 {
		t.Fatalf("permanent prompt = %+v", own)
	}
	if err := g.ResolveOwnPermanents(own.ID, me.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if !me.Hand.Contains(rock) || !effects.OnBattlefield(g, bear) || me.Life != life {
		t.Errorf("rock in hand %v, bear on battlefield %v, life %d→%d", me.Hand.Contains(rock), effects.OnBattlefield(g, bear), life, me.Life)
	}
}
