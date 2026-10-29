package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nazgûl Battle-Mace — Artifact — Equipment {5}:
//
//	"Equipped creature has menace, deathtouch, annihilator 1, and
//	 "Whenever an opponent sacrifices a nontoken permanent, put that
//	 card onto the battlefield under your control unless that player
//	 pays 3 life."
//	 Equip {3}"
//
// The keywords are layer-6 grants; annihilator 1 goes through the
// cumulative append, so a creature that already has annihilator gets a
// second instance, and the engine turns each into an attack trigger
// (ADR 0113 §2). The quoted trigger is a granted ability bundle (ADR
// 0093): the equipped creature has it, so "you" is that creature's
// controller. It triggers on any sacrifice of a nontoken permanent by
// an opponent (the 2023-11-03 ruling), read as the sacrifice is
// announced. As it resolves, the player who sacrificed may pay 3 life
// (offered only if their life total is at least 3, CR 119.4); if they
// don't, the card is
// put onto the battlefield under your control if it is still in a
// graveyard.
//
// "That card" is the sacrificed card wherever the sacrifice left it
// (#3859): a graveyard, exile under Rest in Peace, a library, or the
// command zone. An Aura put onto the battlefield chooses what it enchants
// as it enters (CR 303.4f, #3158).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ea3ba2d-e88f-4879-92f2-681637094b45",
		Name:         "Nazgûl Battle-Mace",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: nazgulBattleMaceGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventSacrifice, AnOpponentSacrificedANontokenPermanent, nazgulBattleMaceLabel, nazgulBattleMaceTakeIt),
			},
			Text: "Whenever an opponent sacrifices a nontoken permanent, put that card onto the battlefield under your control unless that player pays 3 life.",
		}},
		Static: []game.StaticAbility{
			GrantToAttached("menace", "deathtouch", "annihilator 1"),
			GrantAbilitiesToAttached(nazgulBattleMaceGrant),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}

const (
	nazgulBattleMaceGrant = "nazgul-battle-mace/take-sacrificed"
	nazgulBattleMaceLabel = "Nazgûl Battle-Mace — put that card onto the battlefield unless its player pays 3 life"
)

// nazgulBattleMaceTakeIt asks the sacrificing player whether to pay 3
// life, and otherwise puts the card onto the battlefield under the
// trigger's controller.
func nazgulBattleMaceTakeIt(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ev := ctx.Trigger().Event
	victim, card, controller := ev.Actor, ev.CardID, item.Controller
	p := g.PlayerByIDForEffect(victim)
	if p == nil || p.Eliminated || p.Life < 3 {
		return putSacrificedCardOntoBattlefield(ctx, card, controller)
	}
	return PickOption{
		Player:   victim,
		Question: "Nazgûl Battle-Mace — pay 3 life, or they put the card you sacrificed onto the battlefield?",
		Options:  []game.ChoiceOption{{Label: "Pay 3 life", LifeCost: 3}, {Label: "Don't pay"}},
		ThenKey:  nazgulBattleMaceAnswer,
		Carry:    []uuid.UUID{card},
	}.Apply(ctx)
}

// nazgulBattleMaceAnswer is the sacrificing player's answer, keyed (#4253):
// the sacrificed card rides the prompt as data, the Mace's controller and
// the Mace are the Context's, and the player is the chooser.
var nazgulBattleMaceAnswer = OptionPickThen("option-pick/nazgul-battle-mace-pay-or-take", func(ctx *Context, r game.OptionPicked) error {
	if r.Index == 0 {
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Chooser, -3)
	}
	if len(r.Carry) == 0 {
		return nil
	}
	return putSacrificedCardOntoBattlefield(ctx, r.Carry[0], ctx.Controller())
})

// putSacrificedCardOntoBattlefield is "put that card onto the battlefield
// under your control" for a card an opponent just sacrificed, from
// wherever the sacrifice left it (#3859): the graveyard in the ordinary
// case, exile under Rest in Peace, the library under a shuffle-in
// replacement, the command zone for a commander (CR 903.9). The card was
// sacrificed all the same (CR 701.17a); a replacement only rewrote where
// it went, and "that card" is the card in that zone. A card that is
// somewhere a sacrifice cannot have sent it, or no longer exists, is left
// alone.
func putSacrificedCardOntoBattlefield(ctx *Context, card, controller uuid.UUID) error {
	if card == uuid.Nil {
		return nil
	}
	z := ctx.Game.FindCardZoneForEffect(card)
	if z == nil {
		return nil
	}
	switch z.Kind {
	case game.ZoneGraveyard:
		return ReturnFromGraveyard{Target: card, Dest: game.ZoneBattlefield, Controller: controller}.Apply(ctx)
	case game.ZoneExile:
		return ReturnFromExile{Target: card, Controller: controller}.Apply(ctx)
	case game.ZoneLibrary, game.ZoneCommand, game.ZoneHand:
		return putCardsFromZoneTogether(ctx, z.Kind, []uuid.UUID{card}, controller, false)
	}
	return nil
}
