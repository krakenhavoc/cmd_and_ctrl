package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iroh, Grand Lotus — Legendary Creature — Human Noble Ally {3}{G}{U}{R},
// 5/5:
//
//	"Firebending 2
//	 During your turn, each non-Lesson instant and sorcery card in your
//	 graveyard has flashback. The flashback cost is equal to that card's
//	 mana cost.
//	 During your turn, each Lesson card in your graveyard has flashback
//	 {1}."
//
// Firebending is Avatar Roku's (two red that last until end of combat).
// The two graveyard lines are Lier, Disciple of the Drowned's static
// flashback grant (CR 702.34a: cast for the flashback cost, exiled
// whenever it would leave the stack) gated to your own turn by
// TimingYourTurnOnly, which leaves each card's own timing in force, so a
// sorcery is still cast only when a sorcery could be (the ruling). The
// first is narrowed to non-Lesson instants and sorceries and costs the
// card's mana cost; the second covers every Lesson card for {1}. They
// are derived from the battlefield on every query, so a card that
// reaches the graveyard during your turn can be cast at once.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "ee8e6ef3-eb96-48bd-8619-8a9112b3d4f0",
		Name:         "Iroh, Grand Lotus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Iroh, Grand Lotus — firebending 2", Do(firebending(2))),
		},
		CastPermissions: []game.CastPermission{
			{
				Zone:              game.ZoneGraveyard,
				Filter:            game.PermissionFilter{CardInZone: true, InstantOrSorceryOnly: true, NotSubtype: "Lesson"},
				AltCostKey:        "flashback",
				ExileOnResolution: true,
				Timing:            game.TimingYourTurnOnly,
				Label:             "Flashback — its mana cost (Iroh, Grand Lotus)",
			},
			{
				Zone:              game.ZoneGraveyard,
				Filter:            game.PermissionFilter{CardInZone: true, Subtype: "Lesson"},
				AltCostKey:        "flashback",
				Cost:              "{1}",
				ExileOnResolution: true,
				Timing:            game.TimingYourTurnOnly,
				Label:             "Flashback {1} (Iroh, Grand Lotus)",
			},
		},
	})
}
