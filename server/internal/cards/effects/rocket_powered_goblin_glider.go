package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rocket-Powered Goblin Glider — Artifact — Equipment {3}:
//
//	"When this Equipment enters, if it was cast from your graveyard,
//	 attach it to target creature you control.
//	 Equipped creature gets +2/+0 and has flying and haste.
//	 Equip {2}
//	 Mayhem {2} (You may cast this card from your graveyard for {2} if
//	 you discarded it this turn. Timing rules still apply.)"
//
// Embercleave's attach trigger behind Null Summoner's provenance test:
// the intervening "if" reads Card.Provenance, the zone the spell that
// became this permanent was cast from (CR 400.7d), so the trigger is not
// put on the stack for an Equipment cast from hand and has no target to
// choose then. Mayhem (CR 702.187, #4628).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "5db9121d-934a-44f1-b178-b389c6595af7",
		Name:             "Rocket-Powered Goblin Glider",
		Completeness:     CompletenessFull,
		CastableZonesIf:  []game.ConditionalCastZone{MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{Mayhem("{2}")},
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			GrantToAttached("flying", "haste"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, wasCastFromYourGraveyard),
					"Rocket-Powered Goblin Glider — attach it to target creature you control", AttachSourceToTarget),
				PermanentYouControl("target creature you control", Creature()),
			),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

// wasCastFromYourGraveyard is "if it was cast from your graveyard":
// the permanent's provenance is a cast out of a graveyard. A card is
// cast from its owner's graveyard unless a permission opened another
// player's, which mayhem never does.
func wasCastFromYourGraveyard(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && source.Provenance.FromZone == game.ZoneGraveyard
}
