package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b05_test.go — the casualty spells of SNC / NCC, S82 batch 5.

const (
	Snc5MakeDisappear   = "5f12f665-f083-48ae-b61a-541051c996ec"
	Snc5ALittleChat     = "1167fe30-32b8-4381-8067-a05cd345977b"
	Snc5LightEmUp       = "c6932bcf-d66b-42c1-8ca4-6a4573e0f6bf"
	Snc5GrislySigil     = "b9d9d118-a073-4e73-90fe-026462d1d895"
	Snc5DigUpTheBody    = "144e3d90-8ed8-4c21-941d-fc59cec042a2"
	Snc5RobTheArchives  = "df45a344-5937-4a3f-a80f-581ea456ea05"
	Snc5RooftopNuisance = "8843faa9-0cd5-4ef3-8e7b-b8833db9dc84"
	Snc5JoinTheMaestros = "849a7ae8-2b4e-4fa5-9607-2bfc961fee40"
	Snc5IllicitShipment = "81d62f0a-b2ba-4e87-9003-e5472bd1101f"
	Snc5CutYourLosses   = "54530a84-8278-4c93-a2cb-f5c86619a053"
	Snc5CutOfTheProfits = "8ce63f60-0211-46c7-8fb6-c5484ebfa029"
	Snc5AudaciousSwap   = "40b0b93d-4075-42d6-bcd8-b2f9910cc7d2"
)

// snc5Hand puts a card in the active seat's hand at main phase.
func Snc5Hand(t *testing.T, g *game.Game, name, typeLine, cost, oracle string) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	return me, id
}

// snc5Drive passes priority and answers every prompt until the stack is
// empty: pick_target with `target`, the first card of a choose_cards or a
// search, a cast offer with castOffers, and pay_unless with pay.
func Snc5Drive(t *testing.T, g *game.Game, target game.TargetRef, castOffers, pay bool) {
	t.Helper()
	for i := 0; i < 80; i++ {
		var pc *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				pc = c
				break
			}
		}
		if pc == nil {
			if StackFullyEmpty(g) {
				return
			}
			PassPriorityAroundTable(t, g)
			continue
		}
		var err error
		switch pc.Kind {
		case game.PendingChoicePickTarget:
			err = g.ResolvePickTarget(pc.ID, pc.Chooser, target)
		case game.PendingChoiceChooseCards:
			err = g.ResolveChooseCards(pc.ID, pc.Chooser, pc.ChooseCards[:1])
		case game.PendingChoiceSearchLibrary:
			err = g.ResolveSearchLibrary(pc.ID, pc.Chooser, pc.SearchCards[:1])
		case game.PendingChoicePayUnless:
			err = g.ResolvePayUnless(pc.ID, pc.Chooser, pay)
		case game.PendingChoiceResolutionCast:
			if !castOffers {
				err = g.ResolvePendingChoice(pc.ID, pc.Chooser, nil)
			} else {
				card := KeywordCastCard(pc)
				err = g.CastForResolutionChoice(pc.Chooser, pc.ID, card, game.CastSpellParams{FromZone: string(OfferedZone(g, pc.Chooser, card))})
			}
		default:
			t.Fatalf("unhandled prompt %s", pc.Kind)
		}
		if err != nil {
			t.Fatalf("answering %s: %v", pc.Kind, err)
		}
	}
	t.Fatal("the stack never emptied")
}

func Snc5Sac(g *game.Game, owner uuid.UUID, power int) uuid.UUID {
	return PushVanillaCreature(g, owner, "Fodder", power, 1)
}
