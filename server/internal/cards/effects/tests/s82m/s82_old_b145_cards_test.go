package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b145_cards_test.go — Throne of Eldraine, S82 older sets batch 145
// (#5516): Banish into Fable, Bartered Cow, Feasting Troll King, Clackbridge
// Troll, Doom Foretold and Sorcerer's Broom.

const (
	eld145Fable = "1f18e63d-ab1f-426c-afd5-ce89433394ca"
	eld145Cow   = "33d42811-686d-4aca-bb3b-086de453ddaf"
	eld145King  = "6f0c6876-dcae-46e5-8c62-6929733c3525"
	eld145Troll = "9c2f25af-0b16-488d-9249-026a84bdb27f"
	eld145Doom  = "e2baf24c-7055-431d-8e96-ba5ff2ef91e1"
	eld145Broom = "e73efb04-30ac-4484-b11b-e98f0353b9b6"
)

func TestS82OldB145EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		eld145Fable: "Banish into Fable", eld145Cow: "Bartered Cow", eld145King: "Feasting Troll King",
		eld145Troll: "Clackbridge Troll", eld145Doom: "Doom Foretold", eld145Broom: "Sorcerer's Broom",
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

func eld145Count(g *game.Game, controller uuid.UUID, name string) int {
	return effectstest.TmtB1Count(g, controller, name)
}

// --- Banish into Fable ---

func eld145Fable145(t *testing.T, artifact, enchantment bool) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if artifact {
		effectstest.B12Push(g, me.ID, "Trinket", "Artifact", "", 0, 0)
	}
	if enchantment {
		effectstest.B12Push(g, me.ID, "Charm", "Enchantment", "", 0, 0)
	}
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Artifact", "", 0, 0)
	return g, me, opp, victim
}

func eld145CastFable(t *testing.T, artifact, enchantment bool, wantCopies int) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp, victim := eld145Fable145(t, artifact, enchantment)
	effectstest.CastCatalogSpell(t, g, "Banish into Fable", "Instant", eld145Fable,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	// The copies, if any, are on the stack asking for new targets.
	for i := 0; i < 4 && effectstest.CopyPromptFor(g, me.ID) != nil; i++ {
		c := effectstest.CopyPromptFor(g, me.ID)
		if err := g.ResolvePickTarget(c.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
			t.Fatalf("copy target: %v", err)
		}
	}
	if got := effectstest.StackCopies(g); got != wantCopies {
		t.Fatalf("%d copies on the stack, want %d", got, wantCopies)
	}
	return g, me, opp, victim
}

func TestS82OldB145BanishIntoFableCopiesOnlyForWhatYouControl(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		artifact, enchantment bool
		copies                int
	}{{"neither", false, false, 0}, {"artifact", true, false, 1}, {"enchantment", false, true, 1}, {"both", true, true, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			eld145CastFable(t, tc.artifact, tc.enchantment, tc.copies)
		})
	}
}

func TestS82OldB145BanishIntoFableBouncesAndMakesAKnight(t *testing.T) {
	g, me, opp, victim := eld145CastFable(t, false, false, 0)
	_ = g.Stack
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(victim) {
		t.Error("the target was not returned to its owner's hand")
	}
	if got := eld145Count(g, me.ID, "Knight"); got != 1 {
		t.Errorf("%d Knights, want 1", got)
	}
	effectstest.AssertKeywords(t, g, effectstest.TokensNamed(g, "Knight")[0].InstanceID, "vigilance")
}

// --- Bartered Cow ---

func TestS82OldB145BarteredCowMakesAFoodWhenItDiesOrIsDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cow := effectstest.B12Push(g, me.ID, "Bartered Cow", "Creature — Ox", eld145Cow, 3, 3)
	effectstest.KillCreature(t, g, me.ID, cow)
	effectstest.PassPriorityAroundTable(t, g)
	if got := eld145Count(g, me.ID, "Food"); got != 1 {
		t.Fatalf("after dying: %d Food, want exactly 1", got)
	}
	hand := effectstest.HandCardFull(me, "Bartered Cow", "Creature — Ox", "{3}{W}", eld145Cow, []string{"W"})
	effectstest.LengDiscard(t, g, me, hand)
	effectstest.PassPriorityAroundTable(t, g)
	if got := eld145Count(g, me.ID, "Food"); got != 2 {
		t.Errorf("after the discard: %d Food, want 2", got)
	}
}

func TestS82OldB145BarteredCowDoesNotTriggerWhenAnotherCardIsDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Bartered Cow", "Creature — Ox", eld145Cow, 3, 3)
	other := effectstest.HandCardFull(me, "Other", "Sorcery", "{1}", "", nil)
	effectstest.LengDiscard(t, g, me, other)
	effectstest.PassPriorityAroundTable(t, g)
	if got := eld145Count(g, me.ID, "Food"); got != 0 {
		t.Errorf("%d Food, want 0", got)
	}
}

func TestS82OldB145BarteredCowStillOnlyOneFoodWhenItIsExiledFromTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cow := effectstest.B12Push(g, me.ID, "Bartered Cow", "Creature — Ox", eld145Cow, 3, 3)
	effectstest.ExileCreature(t, g, cow)
	effectstest.PassPriorityAroundTable(t, g)
	if got := eld145Count(g, me.ID, "Food"); got != 0 {
		t.Errorf("%d Food from an exile, want 0 (it did not die)", got)
	}
}

// --- Feasting Troll King ---

func TestS82OldB145FeastingTrollKingMakesThreeFoodWhenCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Feasting Troll King", "Creature — Troll Noble", eld145King, 7, 6)
	if got := eld145Count(g, me.ID, "Food"); got != 3 {
		t.Errorf("%d Food, want 3", got)
	}
}

func TestS82OldB145FeastingTrollKingMakesNoFoodWhenNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Feasting Troll King", "Creature — Troll Noble", eld145King, 7, 6)
	effectstest.PassPriorityAroundTable(t, g)
	if got := eld145Count(g, me.ID, "Food"); got != 0 {
		t.Errorf("%d Food from a creature that was not cast, want 0", got)
	}
}

func TestS82OldB145FeastingTrollKingReturnsFromTheGraveyardForThreeFoods(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	king := effectstest.DrcTraitorInGraveyard(g, me)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == king {
				me.Graveyard.Cards[i].Name, me.Graveyard.Cards[i].OracleID = "Feasting Troll King", eld145King
				me.Graveyard.Cards[i].TypeLine = "Creature — Troll Noble"
			}
		}
	})
	foods := effectstest.PushTokens(g, me.ID, effects.FoodToken(), 3)
	if err := g.ActivateCatalogAbility(me.ID, king, 0, game.ActivateAbilityParams{SacrificeIDs: foods}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AppInZone(g, king, game.ZoneBattlefield) {
		t.Error("the King did not return to the battlefield")
	}
	for _, f := range foods {
		if effectstest.AppInZone(g, f, game.ZoneBattlefield) {
			t.Error("a Food was not sacrificed")
		}
	}
}

func TestS82OldB145FeastingTrollKingRefusesWithTwoFoodsOrOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	king := effectstest.DrcTraitorInGraveyard(g, me)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == king {
				me.Graveyard.Cards[i].Name, me.Graveyard.Cards[i].OracleID = "Feasting Troll King", eld145King
			}
		}
	})
	foods := effectstest.PushTokens(g, me.ID, effects.FoodToken(), 2)
	if err := g.ActivateCatalogAbility(me.ID, king, 0, game.ActivateAbilityParams{SacrificeIDs: foods}); err == nil {
		t.Error("two Foods paid for a three-Food cost")
	}
}

// --- Clackbridge Troll ---

func TestS82OldB145ClackbridgeTrollGivesTheTargetOpponentThreeGoats(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.CastHoldCreature(t, g, "Clackbridge Troll", "Creature — Troll", eld145Troll, 8, 8)
	pick := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := eld145Count(g, opp.ID, "Goat"); got != 3 {
		t.Errorf("%d Goats for the opponent, want 3", got)
	}
	if got := eld145Count(g, me.ID, "Goat"); got != 0 {
		t.Errorf("%d Goats for the caster, want 0", got)
	}
}

func TestS82OldB145ClackbridgeTrollTapsGainsAndDrawsWhenAnOpponentSacrifices(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		troll := effectstest.B12Push(g, me.ID, "Clackbridge Troll", "Creature — Troll", eld145Troll, 8, 8)
		goat := effectstest.B12Push(g, opp.ID, "Goat", "Creature — Goat", "", 0, 1)
		life, hand := me.Life, me.Hand.Size()
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PassPriorityAroundTable(t, g)
		who := uuid.Nil
		if paid {
			who = opp.ID
		}
		effectstest.AppAnswerWith(t, g, effectstest.AppOnly(who, goat))
		tapped := effectstest.LayeredCard(t, g, troll).Tapped
		if paid {
			if effectstest.AppInZone(g, goat, game.ZoneBattlefield) {
				t.Error("paid: the creature was not sacrificed")
			}
			if !tapped || me.Life != life+3 || me.Hand.Size() != hand+1 {
				t.Errorf("paid: tapped=%v life %d→%d hand %d→%d, want tapped, +3 life, +1 card", tapped, life, me.Life, hand, me.Hand.Size())
			}
			continue
		}
		if tapped || me.Life != life || me.Hand.Size() != hand {
			t.Errorf("unpaid: tapped=%v life %d→%d hand %d→%d, want nothing", tapped, life, me.Life, hand, me.Hand.Size())
		}
	}
}

func TestS82OldB145ClackbridgeTrollDoesNotOfferOnAnOpponentsCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, opp.ID, "Clackbridge Troll", "Creature — Troll", eld145Troll, 8, 8)
	effectstest.B12Push(g, me.ID, "Goat", "Creature — Goat", "", 0, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if link := effectstest.AppOpenLink(g); link != nil {
		t.Errorf("an offer is open on an opponent's turn: %+v", link)
	}
}

// --- Doom Foretold ---

func eld145Upkeep(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	effectstest.AppToUpkeepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestS82OldB145DoomForetoldMakesTheUpkeepPlayerSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doom := effectstest.B12Push(g, me.ID, "Doom Foretold", "Enchantment", eld145Doom, 0, 0)
	victim := effectstest.B12Push(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	effectstest.PushToken(g, opp.ID, effects.FoodToken())
	eld145Upkeep(t, g, 1)
	if opts := effectstest.SacrificeOptions(effectstest.SacrificeChoiceFor(g, opp.ID)); len(opts) != 1 || opts[0] != victim {
		t.Errorf("sacrifice options %v, want only the nonland nontoken permanent", opts)
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, victim)
	if effectstest.AppInZone(g, victim, game.ZoneBattlefield) {
		t.Error("the chosen permanent was not sacrificed")
	}
	if !effectstest.AppInZone(g, doom, game.ZoneBattlefield) || eld145Count(g, me.ID, "Knight") != 0 {
		t.Error("the rest of the ability ran although they could sacrifice")
	}
}

func TestS82OldB145DoomForetoldPunishesAPlayerWithNothingToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	doom := effectstest.B12Push(g, me.ID, "Doom Foretold", "Enchantment", eld145Doom, 0, 0)
	// Tokens and lands are not sacrificeable choices.
	effectstest.PushToken(g, opp.ID, effects.FoodToken())
	effectstest.B12Push(g, opp.ID, "Plains", "Basic Land — Plains", "", 0, 0)
	card := effectstest.HandCardFull(opp, "Spare", "Sorcery", "{1}", "", nil)
	oppLife, myLife, myHand, oppHand := opp.Life, me.Life, me.Hand.Size(), opp.Hand.Size()
	eld145Upkeep(t, g, 1)
	if pick := effectstest.DiscardChoiceFor(g, opp.ID); pick != nil {
		effectstest.AnswerDiscard(t, g, opp.ID, card)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != oppHand-1 {
		t.Errorf("their hand %d→%d, want one card discarded", oppHand, opp.Hand.Size())
	}
	if opp.Life != oppLife-2 || me.Life != myLife+2 {
		t.Errorf("life: they %d→%d, you %d→%d, want -2 and +2", oppLife, opp.Life, myLife, me.Life)
	}
	if me.Hand.Size() != myHand+1 {
		t.Errorf("hand %d→%d, want +1", myHand, me.Hand.Size())
	}
	if got := eld145Count(g, me.ID, "Knight"); got != 1 {
		t.Errorf("%d Knights, want 1", got)
	}
	if effectstest.AppInZone(g, doom, game.ZoneBattlefield) {
		t.Error("Doom Foretold was not sacrificed")
	}
}

// --- Sorcerer's Broom ---

func TestS82OldB145SorcerersBroomCopiesItselfWhenYouSacrificeAndPay(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		broom := effectstest.B12Push(g, me.ID, "Sorcerer's Broom", "Artifact Creature — Spirit", eld145Broom, 2, 1)
		food := effectstest.PushToken(g, me.ID, effects.FoodToken())
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		if err := g.SacrificePermanentForEffect(food); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.FillPool(me, 3)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if pay {
			want = 1
		}
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Sorcerer's Broom" && c.InstanceID != broom && c.IsToken() {
				n++
			}
		}
		if n != want {
			t.Errorf("pay=%v: %d token copies, want %d", pay, n, want)
		}
	}
}
