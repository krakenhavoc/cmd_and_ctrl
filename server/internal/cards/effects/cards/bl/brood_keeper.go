package cardsbl

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brood Keeper — Creature — Human Shaman {3}{R}, 2/3 (Magic 2015, #5556):
//
//	"Whenever an Aura becomes attached to this creature, create a 2/2 red
//	 Dragon creature token with flying. It has '{R}: This creature gets +1/+0
//	 until end of turn.'"
//
// Eriette's attach watch (CR 303.4f): an Aura that becomes attached to this
// creature, whoever controls the Aura, once per attachment; re-attaching to
// the host it is already on does nothing (CR 701.3b). The token is the
// Dragon the set prints.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c123f45d-a8db-405f-93ef-7227e826790c",
		Name:         "Brood Keeper",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttach, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventAttach && ev.Label == "Aura" && !ev.AlreadyAttached &&
					source != nil && ev.Target != uuid.Nil && ev.Target == source.InstanceID
			}, "Brood Keeper — create a 2/2 red Dragon creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.OldB356DragonToken(), N: 1})),
		},
	})
}
