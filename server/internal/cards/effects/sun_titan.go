package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sun Titan — 6/6 Creature — Giant for {4}{W}{W}:
//
//	"Vigilance
//	Whenever this creature enters or attacks, you may return target
//	permanent card with mana value 3 or less from your graveyard to
//	the battlefield."
//
// S22 attack triggers, and the first card from the Aang decklist's
// attack-trigger group to land in the catalog.
//
// "Enters or attacks" is ONE printed ability with two trigger
// conditions, so it is one TriggeredAbility watching two event kinds
// rather than two declarations — EventETB and EventAttack both carry
// the permanent in CardID, so a single AppliesTo covers both. A Sun
// Titan that enters and then attacks the same turn fires twice, once
// per condition, which is correct.
//
// Reanimation is real: ReturnFromGraveyard with Dest ZoneBattlefield
// moves the card and emits its EventETB, so the returned permanent's
// own enters-trigger fires. The card returns under its OWNER's
// control, which matches the printed text ("from YOUR graveyard").
func init() {
	Register(Spec{
		OracleID:        "b2e950fb-cb7e-40a0-a311-5bbdd0477b29",
		Name:            "Sun Titan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB, game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetCardInGraveyard(
				"target permanent card with mana value 3 or less in your graveyard",
				YouOwn(), Permanent(), ManaValueLE(3),
			),
			Key: "Sun Titan — return a permanent card to the battlefield",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				return ReturnFromGraveyard{
					Target: item.Targets[0].ID,
					Dest:   game.ZoneBattlefield,
				}.Apply(NewContext(g, item))
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Sun Titan — return a permanent card from your graveyard to the battlefield?",
			},
		}},
	})
}
