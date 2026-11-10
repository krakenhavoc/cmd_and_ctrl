package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thieving Amalgam — Creature — Ape Snake {5}{B}{B} 6/7 (Outlaws of
// Thunder Junction Commander, #5429):
//
//	"At the beginning of each opponent's upkeep, you manifest the top card
//	 of that player's library. (Put it onto the battlefield face down as a
//	 2/2 creature. Turn it face up any time for its mana cost if it's a
//	 creature card.)
//	 Whenever a creature you control but don't own dies, its owner loses 2
//	 life and you gain 2 life."
//
// The manifest is the two-player door (ManifestTopOfLibrary,
// game/manifest_library.go): the upkeep's player is the library's owner,
// read off the triggering event, and the Amalgam's controller controls the
// face-down 2/2 and is the only player who may look at it (CR 708.5). An
// empty library manifests nothing.
//
// The dies trigger reads the creature as it last existed (CR 603.10a): the
// controller it left under and the type it left with are the exit's own
// stamps (Edea's predicate), so a manifested 2/2 that was an instant card
// counts, and a creature its owner took back before it died does not. The
// owner is the snapshot's, so a stolen token, which has ceased to exist by
// the time the trigger resolves, still names the player who created it.
// The life loss and the life gain are separate: you gain 2 whether or not
// the owner could lose any.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f60cd1d8-eeb6-41a3-af32-3e2a54937ca2",
		Name:         "Thieving Amalgam",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtEachOpponentsUpkeep("Thieving Amalgam — manifest the top card of that player's library",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return ManifestTopOfLibrary{LibraryOwner: item.Trigger.Event.Actor}.Apply(NewContext(g, item))
				}),
			On(game.EventLTB, edeaCreatureYouControlButDontOwnDied,
				"Thieving Amalgam — its owner loses 2 life and you gain 2 life", thievingAmalgamDrain),
		},
	})
}

// thievingAmalgamDrain is "its owner loses 2 life and you gain 2 life",
// the owner read off the dead creature's last-known information.
func thievingAmalgamDrain(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if obj := ctx.Trigger().Object; obj != nil && obj.Owner != uuid.Nil {
		if err := g.ChangePlayerLifeForEffect(item.SourceCardID, obj.Owner, -2); err != nil {
			return err
		}
	}
	return GainLife{Amount: 2}.Apply(ctx)
}
