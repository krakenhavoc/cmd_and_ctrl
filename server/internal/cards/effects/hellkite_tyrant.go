package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hellkite Tyrant — Creature — Dragon {4}{R}{R}, 6/5:
//
//	"Flying, trample
//	 Whenever this creature deals combat damage to a player, gain
//	 control of all artifacts that player controls.
//	 At the beginning of your upkeep, if you control twenty or more
//	 artifacts, you win the game."
//
// The steal is Agent of Treachery's indefinite control change (the
// printed text states no duration), applied to the set the damaged
// player controls as the trigger RESOLVES (CR 611.2c): an artifact that
// player gains in response is not taken, and one that left is skipped.
// The damaged player is read off the triggering damage event, not a
// captured value.
//
// The win is Felidar Sovereign's shape: an intervening-if (CR 603.4)
// checked on the way onto the stack and again on resolution, then
// WinTheGame, which a Platinum Angel or similar gate can still stop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d9b066ff-9519-415c-ae17-bfea703c9889",
		Name:            "Hellkite Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(
				"Hellkite Tyrant — gain control of all artifacts that player controls",
				hellkiteTyrantSteal),
			On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && controlsTwentyArtifacts(g, source.Controller)
			}, "Hellkite Tyrant — if you control twenty or more artifacts, you win the game",
				func(g *game.Game, item *game.StackItem) error {
					if !controlsTwentyArtifacts(g, item.Controller) {
						return nil
					}
					return WinTheGame{Player: item.Controller}.Apply(NewContext(g, item))
				}),
		},
	})
}

// controlsTwentyArtifacts is Hellkite Tyrant's intervening-if.
func controlsTwentyArtifacts(g *game.Game, player uuid.UUID) bool {
	return len(permanentsControlledByMatching(g, player, Artifact())) >= 20
}

// hellkiteTyrantSteal takes every artifact the damaged player controls
// now, for no stated duration.
func hellkiteTyrantSteal(g *game.Game, item *game.StackItem) error {
	victim := item.Trigger.Event.Target
	ctx := NewContext(g, item)
	for _, id := range permanentsControlledByMatching(g, victim, Artifact()) {
		if err := (GainControl{
			Target:     id,
			Controller: item.Controller,
			Duration:   game.IndefiniteDuration(),
			Label:      "Hellkite Tyrant — gain control of all artifacts",
		}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
