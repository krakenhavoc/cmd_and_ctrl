package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gisa and Geralf — Legendary Creature — Human Wizard {2}{U}{B}, 4/4:
//
//	"When Gisa and Geralf enters, mill four cards.
//	 Once during each of your turns, you may cast a Zombie creature
//	 spell from your graveyard."
//
// The mill is an ordinary entry trigger. The cast is a STANDING
// graveyard permission (ADR 0066) derived off the battlefield, with
// Muldrotha's per-type budget (CastPermission.PerType, #2167) holding
// the one type "creature": one cast per turn per Gisa and Geralf
// OBJECT, so a second copy that arrives later in the turn opens one
// more, as the rulings say. TimingYourTurnOnly closes it off your turn
// and leaves the creature's own timing in force, and the spell's costs
// are paid as usual. A Zombie whose own text lets it be cast from the
// graveyard is cast by that text and spends nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "20e94d0f-6887-45f8-a137-877011d786c6",
		Name:         "Gisa and Geralf",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gisa and Geralf — mill four cards", Do(MillCards{N: 4})),
		},
		CastPermissions: []game.CastPermission{{
			Zone:     game.ZoneGraveyard,
			CastOnly: true,
			Filter:   game.PermissionFilter{CreatureOnly: true, CreatureType: "Zombie"},
			PerType:  []string{game.PermissionTypeCreature},
			Timing:   game.TimingYourTurnOnly,
			Label:    "Once during each of your turns, cast a Zombie creature spell from your graveyard (Gisa and Geralf)",
		}},
	})
}
