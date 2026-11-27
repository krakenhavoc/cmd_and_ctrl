package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b585_cards_test.go — Saviors of Kamigawa, S82 older sets, batch 585
// (#5601). Each test plays the real catalog entry.

const (
	oldB585Pains   = "4608c663-e088-40ca-a915-b6e39253eb98"
	oldB585Vines   = "64ab6dfe-ed87-47ec-afd5-a7f2e7e6bbd2"
	oldB585Scroll  = "6974c433-a522-40bb-8791-375fcc242234"
	oldB585Ideas   = "6b4d34bc-fa64-4c42-90e0-190671a5b988"
	oldB585Soramar = "b11f96a0-5f70-45d6-90cc-2edbead3672e"
	oldB585Curtain = "d6efbbf9-1449-45ce-b187-30041eb36d93"
	oldB585Zubera  = "d99b44bc-375a-4cf1-9148-ef6fe615cb10"
)

func TestOldB585EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB585Pains:   "Pain's Reward",
		oldB585Vines:   "Rending Vines",
		oldB585Scroll:  "Scroll of Origins",
		oldB585Ideas:   "Ideas Unbound",
		oldB585Soramar: "Soramaro, First to Dream",
		oldB585Curtain: "Curtain of Light",
		oldB585Zubera:  "Rushing-Tide Zubera",
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

// oldB585Cast puts a free spell of `typeLine` in the caster's hand and casts it.
func oldB585Cast(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracle string, targets ...uuid.UUID) error {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	var refs []game.TargetRef
	for _, tg := range targets {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: tg})
	}
	return g.CastSpell(p.ID, id, game.CastSpellParams{Targets: refs, Strict: true})
}

func oldB585Library(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.SeedLibrary(p, "Card")
	}
}

// --- Pain's Reward ----------------------------------------------------

func oldB585OpenChoice(g *game.Game, kind game.PendingChoiceKind) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		if c := g.PendingChoices[i]; c != nil && c.Kind == kind {
			return c
		}
	}
	return nil
}

func oldB585Bid(t *testing.T, g *game.Game, who uuid.UUID, n int) {
	t.Helper()
	c := oldB585OpenChoice(g, game.PendingChoicePayAmount)
	if c == nil || c.Chooser != who {
		t.Fatalf("want a bid asked of %s, got %+v", who, c)
	}
	if err := g.ResolvePayAmount(c.ID, who, n); err != nil {
		t.Fatalf("bid %d: %v", n, err)
	}
}

func oldB585Answer(t *testing.T, g *game.Game, who uuid.UUID, option int) {
	t.Helper()
	c := oldB585OpenChoice(g, game.PendingChoiceOptionPick)
	if c == nil || c.Chooser != who {
		t.Fatalf("want %s asked to top the bid, got %+v", who, c)
	}
	if err := g.ResolveOptionPick(c.ID, who, option); err != nil {
		t.Fatalf("answer %d: %v", option, err)
	}
}

func TestOldB585PainsRewardHighBidderLosesTheBidAndDrawsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	for _, p := range g.Seats {
		oldB585Library(p, 6)
	}
	if err := oldB585Cast(t, g, me, "Pain's Reward", "Sorcery", oldB585Pains); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	oldB585Bid(t, g, me.ID, 3)
	life, hand := me.Life, me.Hand.Size()
	others := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		others[p.ID] = p.Hand.Size()
		oldB585Answer(t, g, p.ID, 0) // let it stand
	}
	if me.Life != life-3 || me.Hand.Size() != hand+4 {
		t.Fatalf("life %d→%d, hand %d→%d: want -3 life and +4 cards", life, me.Life, hand, me.Hand.Size())
	}
	for _, p := range g.Seats[1:] {
		if p.Life != 40 || p.Hand.Size() != others[p.ID] {
			t.Errorf("%s was not the high bidder yet life %d, hand %d", p.Name, p.Life, p.Hand.Size())
		}
	}
}

func TestOldB585PainsRewardAnOpponentWhoToppedItWinsAndPays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	for _, p := range g.Seats {
		oldB585Library(p, 6)
	}
	if err := oldB585Cast(t, g, me, "Pain's Reward", "Sorcery", oldB585Pains); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	oppHand, myHand := opp.Hand.Size(), me.Hand.Size()
	oldB585Bid(t, g, me.ID, 2)
	oldB585Answer(t, g, opp.ID, 1) // top it
	oldB585Bid(t, g, opp.ID, 5)
	oldB585Answer(t, g, g.Seats[2].ID, 0)
	oldB585Answer(t, g, g.Seats[3].ID, 0)
	oldB585Answer(t, g, me.ID, 0)
	if opp.Life != 35 || opp.Hand.Size() != oppHand+4 {
		t.Errorf("the winner has life %d, hand %d: want 35 and four more cards", opp.Life, opp.Hand.Size())
	}
	if me.Life != 40 || me.Hand.Size() != myHand {
		t.Errorf("the loser of the bidding has life %d, hand %d: want 40 and no change", me.Life, me.Hand.Size())
	}
}

// --- Rending Vines ----------------------------------------------------

func TestOldB585RendingVinesDestroysOnlyWithEnoughCardsInHand(t *testing.T) {
	for _, tc := range []struct {
		hand      int
		destroyed bool
	}{{2, false}, {3, true}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := g.Seats[0], g.Seats[1]
		oldB585Library(me, 3)
		rock := effectstest.B12Push(g, opp.ID, "Three Rock", "Artifact", "", 0, 0)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == rock {
					g.Battlefield.Cards[i].ManaCost = "{3}"
				}
			}
		})
		effectstest.D2dSetHand(me, tc.hand)
		if err := oldB585Cast(t, g, me, "Rending Vines", "Instant — Arcane", oldB585Vines, rock); err != nil {
			t.Fatalf("cast: %v", err)
		}
		// The spell has left the hand: the count is what it left behind.
		effectstest.PassPriorityAroundTable(t, g)
		onField := effects.OnBattlefield(g, rock)
		if onField == tc.destroyed {
			t.Errorf("hand %d: artifact still on the battlefield = %v, want destroyed = %v", tc.hand, onField, tc.destroyed)
		}
		if me.Hand.Size() != tc.hand+1 {
			t.Errorf("hand %d: hand is %d after the draw, want %d", tc.hand, me.Hand.Size(), tc.hand+1)
		}
	}
}

// --- Scroll of Origins ------------------------------------------------

func TestOldB585ScrollOfOriginsDrawsOnlyWithSevenCards(t *testing.T) {
	for _, tc := range []struct{ hand, drawn int }{{6, 0}, {7, 1}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[0]
		oldB585Library(me, 3)
		scroll := effectstest.PushCatalogPermanent(g, me.ID, "Scroll of Origins", "Artifact", oldB585Scroll, false)
		effectstest.D2dSetHand(me, tc.hand)
		me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
		if err := g.ActivateCatalogAbility(me.ID, scroll, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("hand %d: activate: %v", tc.hand, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Size() - tc.hand; got != tc.drawn {
			t.Errorf("hand %d: drew %d, want %d", tc.hand, got, tc.drawn)
		}
	}
}

// --- Ideas Unbound ----------------------------------------------------

func TestOldB585IdeasUnboundDrawsThreeThenDiscardsThreeAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	oldB585Library(me, 6)
	effectstest.D2dSetHand(me, 0)
	if err := oldB585Cast(t, g, me, "Ideas Unbound", "Sorcery — Arcane", oldB585Ideas); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 3 {
		t.Fatalf("hand is %d after the draw, want 3", me.Hand.Size())
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.DiscardChoiceFor(g, me.ID); c == nil || c.ChooseMax != 3 {
		t.Fatalf("no three-card discard prompt for the caster at the end step: %+v", c)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != 0 {
		t.Errorf("hand is %d after discarding three of three", me.Hand.Size())
	}
}

// --- Curtain of Light -------------------------------------------------

func TestOldB585CurtainOfLightBlocksAnUnblockedAttackerAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	oldB585Library(them, 3)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := oldB585Cast(t, g, them, "Curtain of Light", "Instant", oldB585Curtain, bear); err == nil {
		t.Fatal("cast in the main phase was accepted")
	} else {
		var cant *game.CantCastError
		if !errors.As(err, &cant) {
			t.Fatalf("main phase: %v, want the timing refusal", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.FinishBlockDeclarations(t, g)
	hand := them.Hand.Size()
	if err := oldB585Cast(t, g, them, "Curtain of Light", "Instant", oldB585Curtain, bear); err != nil {
		t.Fatalf("cast after blockers: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	blocked := false
	g.ReadSnapshot(func() { blocked = g.BlockedAttackerForEffect(bear) })
	if !blocked {
		t.Error("the attacker did not become blocked")
	}
	if them.Hand.Size() != hand+1 {
		t.Errorf("hand %d→%d, want a card drawn", hand, them.Hand.Size())
	}
	// A blocked attacker is no longer a legal target.
	if err := oldB585Cast(t, g, them, "Curtain of Light", "Instant", oldB585Curtain, bear); err == nil {
		t.Error("a blocked attacker was a legal target")
	}
}

// --- Soramaro, First to Dream -----------------------------------------

func TestOldB585SoramaroIsAsBigAsYourHandAndDrawsForALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	oldB585Library(me, 3)
	sora := effectstest.PushCatalogPermanent(g, me.ID, "Soramaro, First to Dream", "Legendary Creature — Spirit", oldB585Soramar, false)
	land := effectstest.CcLand(g, me.ID, "Island", "Basic Land — Island")
	effectstest.D2dSetHand(me, 4)
	if p, tough := effectstest.EffectivePTOf(t, g, sora); p != 4 || tough != 4 {
		t.Fatalf("Soramaro is %d/%d with four cards in hand, want 4/4", p, tough)
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, sora, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.InHandOf(me, land) {
		t.Fatal("the land was not returned as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != 6 {
		t.Fatalf("hand is %d, want 4 + the land + the drawn card", me.Hand.Size())
	}
	if p, tough := effectstest.EffectivePTOf(t, g, sora); p != 6 || tough != 6 {
		t.Errorf("Soramaro is %d/%d with six cards in hand, want 6/6", p, tough)
	}
}

// --- Rushing-Tide Zubera ----------------------------------------------

func TestOldB585RushingTideZuberaDrawsThreeOnlyAfterFourDamage(t *testing.T) {
	for _, tc := range []struct{ damage, drawn int }{{3, 0}, {4, 3}} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := g.Seats[0], g.Seats[1]
		oldB585Library(me, 6)
		zub := effectstest.PushCatalogPermanent(g, me.ID, "Rushing-Tide Zubera", "Creature — Zubera Spirit", oldB585Zubera, false)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == zub {
					g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 3, 3
				}
			}
		})
		src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Bear", 5, 5)
		hand := me.Hand.Size()
		// Two hits, so the total is what counts: 2+1 is lethal on a 3/3 with 3 damage, 2+2 with 4.
		effectstest.DealDamageTo(t, g, src, zub, 2)
		effectstest.DealDamageTo(t, g, src, zub, tc.damage-2)
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Size() - hand; got != tc.drawn {
			t.Errorf("%d damage: drew %d, want %d", tc.damage, got, tc.drawn)
		}
	}
}
