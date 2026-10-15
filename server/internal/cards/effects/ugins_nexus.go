package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ugin's Nexus — Legendary Artifact {5}:
//
//	"If a player would begin an extra turn, that player skips that turn
//	 instead.
//	 If Ugin's Nexus would be put into a graveyard from the battlefield,
//	 instead exile it and take an extra turn after this one."
//
// ADR 0188 Decision 2 (#3560). Two replacement effects, and the second
// one gives the card away: it is a way to take an extra turn that the
// first would skip, which is why the card exiles itself on the way out.
//
//   - "If a player would begin an extra turn, that player skips that
//     turn instead" is SkipExtraTurns(), Trouble in Pairs' replacement
//     without the opponent test: it skips the Nexus controller's own
//     extra turns as well. The skip happens when the turn WOULD BEGIN,
//     so (the 2014-09-20 ruling) an extra turn queued while the Nexus
//     is on the battlefield is taken if the Nexus has left by then.
//   - "Instead exile it and take an extra turn" is a self-replacement on
//     the move from the battlefield to a graveyard (the Grim Repriser
//     shape): it redirects the move to exile, and declares the extra turn
//     for the Nexus' controller with
//     TakeExtraTurnAfterReplacedMoveForEffect. The turn is queued after
//     the exile, not inside the callback (a Replace callback emits no
//     event), so the log reads the exile first and the extra turn after
//     it. The Nexus is in exile by the time the turn would begin, so its
//     first ability no longer applies and the turn is taken, unless
//     another skip does.
//
// Because the card is exiled and not put into a graveyard, it never
// died: dies triggers do not see it (CR 700.4). With two Nexuses
// destroyed together each is exiled with its own extra turn queued, and
// the turn is skipped only if a Nexus is still on the battlefield when
// it would begin (the other Nexus, if it was not itself destroyed).
// Rest in Peace is a second graveyard replacement on the same move: the
// controller of the Nexus chooses the CR 616 order, and choosing Rest in
// Peace first exiles the Nexus without the extra turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fe2afa18-54d0-4595-af27-256398793a42",
		Name:         "Ugin's Nexus",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			SkipExtraTurns(),
			{
				Watches: []game.EventKind{game.EventZoneMove},
				AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
					return ev.Kind == game.RepEventMove && ev.OldZone == game.ZoneBattlefield &&
						ev.NewZone == game.ZoneGraveyard && src != nil && ev.CardID == src.InstanceID
				},
				Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
					ev.NewZone = game.ZoneExile
					ev.NewZoneOwner = uuid.Nil
					return g.TakeExtraTurnAfterReplacedMoveForEffect(ev, src.InstanceID, src.Controller)
				},
				Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
					return src.Controller
				},
				Label: "Ugin's Nexus: exile instead of a graveyard, and take an extra turn",
			},
		},
	})
}
