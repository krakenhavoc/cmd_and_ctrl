package testsa

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// any_player_pays_cards_test.go — the seven mana-offer cards of ADR
// 0150 PR 1 (#1903): each card's effect happens only when nobody pays,
// and the offer goes round the table from the active player.

const (
	appRhysticTutor  = "1c6da9a9-6395-460b-8150-ea46d7299dc2"
	appWildMight     = "3f8fa982-73ee-410a-8873-0afc8a6f11fc"
	appRhysticShield = "a905dafa-1818-471a-a485-d84be447263b"
	appSoulStrings   = "fac310c7-c1e5-42f9-8140-17472be21d0d"
	appNakayaShade   = "ce969713-fdaa-4c31-b164-76a9b0b9ff81"
	appRhysticCircle = "bffb7d30-b9a4-4529-a3fc-c82d6cd4b599"
	appIcyPrison     = "d4e2cc1a-4212-4284-9ebe-d3c40931de50"
)

func appSeats(g *game.Game) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(g.Seats))
	for _, s := range g.Seats {
		out = append(out, s.ID)
	}
	return out
}

func TestAnyPlayerPaysRhysticTutorSearchesOnlyWhenNobodyPays(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		needle := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Needle", TypeLine: "Artifact"})
		effectstest.CastCatalogSpell(t, g, "Rhystic Tutor", "Sorcery", appRhysticTutor, nil)
		effectstest.AppMana(opp, 2)
		effectstest.PassPriorityAroundTable(t, g)
		var payers []uuid.UUID
		if paid {
			payers = append(payers, opp.ID)
		}
		asked := effectstest.AppAnswerOffer(t, g, payers...)
		search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
		if paid {
			if !slices.Equal(asked, appSeats(g)[:2]) {
				t.Errorf("paid: asked %v, want seat 0 then seat 1", asked)
			}
			if search != nil {
				t.Error("paid: the search happened anyway")
			}
			if len(opp.ManaPool) != 0 {
				t.Errorf("paid: seat 1 kept %d mana", len(opp.ManaPool))
			}
			continue
		}
		if !slices.Equal(asked, appSeats(g)) {
			t.Errorf("unpaid: asked %v, want every seat from the active player", asked)
		}
		if search == nil {
			t.Fatal("unpaid: no search")
		}
		if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{needle}); err != nil {
			t.Fatalf("search: %v", err)
		}
		if !me.Hand.Contains(needle) {
			t.Error("unpaid: the searched card is not in hand")
		}
	}
}

func TestAnyPlayerPaysWildMightGivesTheExtraFourOnlyWhenNobodyPays(t *testing.T) {
	for _, tc := range []struct {
		paid bool
		want int
	}{{false, 7}, {true, 3}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
		effectstest.CastCatalogSpell(t, g, "Wild Might", "Instant", appWildMight, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
		effectstest.AppMana(opp, 2)
		effectstest.PassPriorityAroundTable(t, g)
		// The +1/+1 is already there while the offer is open.
		if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 3 {
			t.Errorf("paid=%v: power %d while the offer is open, want 3", tc.paid, p)
		}
		var payers []uuid.UUID
		if tc.paid {
			payers = append(payers, opp.ID)
		}
		effectstest.AppAnswerOffer(t, g, payers...)
		if p, tough := effectstest.EffectivePTOf(t, g, bear); p != tc.want || tough != tc.want {
			t.Errorf("paid=%v: %d/%d, want %d/%d", tc.paid, p, tough, tc.want, tc.want)
		}
	}
}

func TestAnyPlayerPaysRhysticShieldGivesTheExtraTwoOnlyWhenNobodyPays(t *testing.T) {
	for _, tc := range []struct {
		paid bool
		want int
	}{{false, 5}, {true, 3}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
		theirs := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Other", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
		effectstest.CastCatalogSpell(t, g, "Rhystic Shield", "Instant", appRhysticShield, nil)
		effectstest.AppMana(opp, 2)
		effectstest.PassPriorityAroundTable(t, g)
		var payers []uuid.UUID
		if tc.paid {
			payers = append(payers, opp.ID)
		}
		effectstest.AppAnswerOffer(t, g, payers...)
		if _, tough := effectstest.EffectivePTOf(t, g, bear); tough != tc.want {
			t.Errorf("paid=%v: toughness %d, want %d", tc.paid, tough, tc.want)
		}
		if _, tough := effectstest.EffectivePTOf(t, g, theirs); tough != 2 {
			t.Errorf("paid=%v: an opponent's creature got toughness %d", tc.paid, tough)
		}
	}
}

func TestAnyPlayerPaysSoulStringsAsksForXAndReturnsBothWhenNobodyPays(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		a := effectstest.PushGraveyardCardWithTypeLineID(me, "Dead One", "Creature — Zombie")
		b := effectstest.PushGraveyardCardWithTypeLineID(me, "Dead Two", "Creature — Zombie")
		active := g.Seats[g.Turn.ActiveSeat]
		id := uuid.New()
		active.Hand.PushTop(game.Card{InstanceID: id, Name: "Soul Strings", TypeLine: "Sorcery", OracleID: appSoulStrings, Owner: active.ID, Controller: active.ID})
		for g.Turn.Step != game.StepPrecombatMain {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatal(err)
			}
		}
		if err := g.CastSpell(active.ID, id, game.CastSpellParams{XValue: 3, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.AppMana(opp, 3)
		effectstest.PassPriorityAroundTable(t, g)
		var link *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.AnyPlayerOffer != nil {
				link = c
			}
		}
		if link == nil || link.PayCost != "{3}" {
			t.Fatalf("paid=%v: the offer is %+v, want {3}", paid, link)
		}
		var payers []uuid.UUID
		if paid {
			payers = append(payers, opp.ID)
		}
		effectstest.AppAnswerOffer(t, g, payers...)
		inHand := me.Hand.Contains(a) && me.Hand.Contains(b)
		if inHand == paid {
			t.Errorf("paid=%v: both cards in hand = %v", paid, inHand)
		}
	}
}

func TestAnyPlayerPaysNakayaShadePumpsOnlyWhenNobodyPays(t *testing.T) {
	for _, tc := range []struct {
		paid bool
		want int
	}{{false, 2}, {true, 1}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		shade := effectstest.PushCatalogPermanent(g, me.ID, "Nakaya Shade", "Creature — Shade", appNakayaShade, false)
		effectstest.AppMana(opp, 2)
		if err := g.ActivateCatalogAbility(me.ID, shade, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		var payers []uuid.UUID
		if tc.paid {
			payers = append(payers, opp.ID)
		}
		effectstest.AppAnswerOffer(t, g, payers...)
		if p, _ := effectstest.EffectivePTOf(t, g, shade); p != tc.want {
			t.Errorf("paid=%v: power %d, want %d", tc.paid, p, tc.want)
		}
	}
}

func TestAnyPlayerPaysRhysticCircleShieldsOnlyWhenNobodyPays(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		circle := effectstest.PushCatalogPermanent(g, me.ID, "Rhystic Circle", "Enchantment", appRhysticCircle, false)
		dragon := effectstest.Pr7Creature(g, opp.ID, "Dragon", 5, "R")
		effectstest.AppMana(opp, 1)
		if err := g.ActivateCatalogAbility(me.ID, circle, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		var payers []uuid.UUID
		if paid {
			payers = append(payers, opp.ID)
		}
		effectstest.AppAnswerOffer(t, g, payers...)
		var prompt *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceChooseSource {
				prompt = c
			}
		}
		if paid {
			if prompt != nil {
				t.Error("paid: the shield's source was asked for anyway")
			}
			continue
		}
		if prompt == nil {
			t.Fatal("unpaid: no source choice")
		}
		effectstest.Pr7Choose(t, g, me.ID, dragon)
		life := me.Life
		effectstest.Pr7Hit(t, g, dragon, me.ID, 5)
		if me.Life != life {
			t.Errorf("unpaid: life %d → %d, want the dragon's damage prevented", life, me.Life)
		}
	}
}

// Icy Prison asks its controller first in their own upkeep; nobody
// paying sacrifices it, and the payment keeps it.
func TestAnyPlayerPaysIcyPrisonUpkeepOfferKeepsOrSacrificesIt(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		owner := g.Seats[1]
		prison := effectstest.PushCatalogPermanent(g, owner.ID, "Icy Prison", "Enchantment", appIcyPrison, false)
		for i := 0; g.Turn.ActiveSeat != 1 || g.Turn.Step != game.StepUpkeep; i++ {
			if i > 40 {
				t.Fatalf("never reached seat 1's upkeep (at seat %d %v)", g.Turn.ActiveSeat, g.Turn.Step)
			}
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		var link *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.AnyPlayerOffer != nil {
				link = c
			}
		}
		if link == nil || link.Chooser != owner.ID || !link.AnyPlayerOffer.HelpsController {
			t.Fatalf("paid=%v: the first link is %+v, want the Prison's controller and a payment that helps them", paid, link)
		}
		effectstest.AppMana(owner, 3)
		var payers []uuid.UUID
		if paid {
			payers = append(payers, owner.ID)
		}
		effectstest.AppAnswerOffer(t, g, payers...)
		onField := false
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == prison {
				onField = true
			}
		}
		if onField != paid {
			t.Errorf("paid=%v: Icy Prison on the battlefield = %v", paid, onField)
		}
	}
}
