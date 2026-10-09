package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Muldrotha, the Gravetide — Legendary Creature — Elemental Avatar
// {3}{B}{G}{U}, 6/6 (#2167):
//
//	"During each of your turns, you may play a land and cast a permanent
//	 spell of each permanent type from your graveyard. (If a card has
//	 multiple permanent types, choose one as you play it.)"
//
// A STANDING graveyard permission (ADR 0066), derived off the
// battlefield on every query, with a per-type budget (#2167,
// CastPermission.PerType): the six permanent types, one play or cast
// each. A land is played and spends "land", beside the turn's land drop
// (CR 305.2), so it needs a land play left. A spell spends one of its
// permanent types, the caster's choice when it has two (the cast's
// `permission_type`). The rulings (2020-11-10) fall out:
//
//   - "Use the type of the card as it's played or cast": the type is
//     judged against the face being cast (and a face-down cast's 2/2
//     creature), not the card in the graveyard.
//   - "You may cast an artifact creature spell as your artifact spell
//     and cast another artifact creature spell as your creature spell":
//     each spends the one type chosen.
//   - "If you … then have a new Muldrotha come under your control in the
//     same turn, you may play another land or spell of that type": the
//     spent types are kept per granting OBJECT (CR 400.7).
//   - "You must follow the normal timing permissions": TimingYourTurnOnly
//     closes the permission off your own turn and leaves each card's own
//     timing in force; a land still needs your main phase and an empty
//     stack.
//   - "You must pay the costs": the permission names no price, so the
//     printed cost (or any alternative cost the card offers) is paid.
//   - "Once you begin to cast a spell, losing control of Muldrotha won't
//     affect the spell": nothing about the spell reads the permission
//     once it is cast.
//   - "If multiple effects allow you to play a card from your graveyard,
//     you must announce which permission you're using": a card whose own
//     text opens the graveyard is cast by that text and spends nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4625704-1d52-44e4-804f-2f45644d76ac",
		Name:         "Muldrotha, the Gravetide",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:    game.ZoneGraveyard,
			PerType: game.PermanentPermissionTypes,
			Timing:  game.TimingYourTurnOnly,
			Label:   "Play a land and cast a permanent spell of each permanent type from your graveyard (Muldrotha, the Gravetide)",
		}},
	})
}
