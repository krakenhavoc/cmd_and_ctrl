package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rona, Disciple of Gix — Legendary Creature — Human Artificer {1}{U}{B}, 2/2
// (Dominaria, S82 OLD, #5530):
//
//	"When Rona enters, you may exile target historic card from your graveyard.
//	 You may cast spells from among cards exiled with Rona.
//	 {4}, {T}: Exile the top card of your library.
//	 (Artifacts, legendaries, and Sagas are historic.)"
//
// Both exiles are linked to this Rona object (CR 607.2a, 400.7): the enters
// trigger's "you may exile" is an optional targeted trigger, the activated
// ability takes whatever is on top as it resolves, and a Rona that left and
// came back has no claim on either. The permission is Intrepid Paleontologist's
// standing cast permission over cards exiled with the source, for spells only
// (a land exiled this way can't be played), at the ordinary cost and timing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "409305b3-5373-4376-8096-c04a73184cef",
		Name:         "Rona, Disciple of Gix",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.Targeting(
					effects.WhenThisEnters("Rona — you may exile target historic card from your graveyard", effects.AvB7ExileTargetLinked),
					effects.TargetCardInGraveyard("target historic card from your graveyard", effects.B14Historic(), effects.YouOwn())),
				"Rona, Disciple of Gix — exile target historic card from your graveyard?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Exile the top card of your library.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect:  effects.S82OldB212ExileTopCardLinked,
		}},
		CastPermissions: []game.CastPermission{{
			Zone:     game.ZoneExile,
			Filter:   game.PermissionFilter{ExiledWithSource: true},
			CastOnly: true,
			Label:    "Cast a spell from among cards exiled with Rona, Disciple of Gix",
		}},
	})
}
