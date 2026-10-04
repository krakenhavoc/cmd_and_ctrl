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
// Sandbox simplifications, declared: a card a replacement sent
// somewhere other than a graveyard (Rest in Peace) stays there, where
// the ruling would still put it onto the battlefield; and an Aura comes
// back unattached and is put into the graveyard, as for Carmen, Cruel
// Skymarcher. Both are weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "1ea3ba2d-e88f-4879-92f2-681637094b45",
		Name:         "Nazgûl Battle-Mace",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A sacrificed card that went somewhere other than a graveyard isn't put onto the battlefield.",
			"An Aura returned this way comes back unattached and is put into the graveyard — you don't get to choose what it enchants.",
		},
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
	take := func(ctx *Context) error {
		if z := ctx.Game.FindCardZoneForEffect(card); card == uuid.Nil || z == nil || z.Kind != game.ZoneGraveyard {
			return nil
		}
		return ReturnFromGraveyard{Target: card, Dest: game.ZoneBattlefield, Controller: controller}.Apply(ctx)
	}
	p := g.PlayerByIDForEffect(victim)
	if p == nil || p.Eliminated || p.Life < 3 {
		return take(ctx)
	}
	return PickOption{
		Player:   victim,
		Question: "Nazgûl Battle-Mace — pay 3 life, or they put the card you sacrificed onto the battlefield?",
		Options:  []game.ChoiceOption{{Label: "Pay 3 life", LifeCost: 3}, {Label: "Don't pay"}},
		Then: func(ctx *Context, index int) error {
			if index == 0 {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -3)
			}
			return take(ctx)
		},
	}.Apply(ctx)
}
