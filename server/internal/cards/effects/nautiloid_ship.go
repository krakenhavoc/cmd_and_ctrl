package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nautiloid Ship — Artifact — Vehicle {4}, 5/5:
//
//	"Flying
//	 When this Vehicle enters, exile target player's graveyard.
//	 Whenever this Vehicle deals combat damage to a player, you may put
//	 a creature card exiled with this Vehicle onto the battlefield under
//	 your control.
//	 Crew 3"
//
// The two triggers are linked (CR 607.2a): "exiled with this Vehicle"
// means the cards the entry trigger exiled and that are still in exile.
// b27ExiledWith reads that off the event log by the entry trigger's
// label (Bag of Holding's and Unlicensed Hearse's shape). A Ship that
// leaves and comes back is a new object (CR 400.7) and has exiled
// nothing (the ruling), which b27ExiledWith's reset on re-entry gives.
//
// The entry trigger is Bojuka Bog's (exileTargetPlayersGraveyard). A
// commander exiled this way whose owner moves it to the command zone
// (CR 903.9a) is no longer in exile and is never offered.
//
// The damage trigger offers the creature cards among them, read as
// cards in exile (printed types). The "may" is the prompt's zero
// minimum. The chosen card enters under the Ship's controller's
// control, whoever owns it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "613a8774-165e-4cf6-ad43-124f9ffc9980",
		Name:            "Nautiloid Ship",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Nautiloid Ship"),
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters(nautiloidShipExileLabel, exileTargetPlayersGraveyard),
				TargetPlayer("target player")),
			WheneverThisDealsCombatDamageToAPlayer(
				"Nautiloid Ship — you may put a creature card exiled with it onto the battlefield under your control",
				nautiloidShipReturnACreature),
		},
	})
}

// nautiloidShipExileLabel is the entry trigger's label, the key its
// exiled-with record is read by.
const nautiloidShipExileLabel = "Nautiloid Ship — exile target player's graveyard"

// nautiloidShipReturnACreature is the combat-damage trigger's body.
func nautiloidShipReturnACreature(g *game.Game, item *game.StackItem) error {
	var creatures []uuid.UUID
	for _, id := range b27ExiledWith(g, item.SourceCardID, nautiloidShipExileLabel) {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() {
			creatures = append(creatures, id)
		}
	}
	if len(creatures) == 0 {
		return nil
	}
	controller := item.Controller
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   item.SourceCardID,
		Question: "Nautiloid Ship — you may put a creature card exiled with it onto the battlefield under your control",
		Cards:    creatures,
		Zone:     game.ZoneExile,
		Min:      0,
		Max:      1,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			_, err := g.ReturnFromExileToBattlefieldForEffect(picked[0], controller, false)
			return err
		},
	})
	return nil
}
