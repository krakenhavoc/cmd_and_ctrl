package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glacierwood Siege — Enchantment {1}{G}{U}:
//
//	"As this enchantment enters, choose Temur or Sultai.
//	 • Temur — Whenever you cast an instant or sorcery spell, target
//	   player mills four cards.
//	 • Sultai — You may play lands from your graveyard."
//
// #1647: one of the three remaining Sieges the #1572 builder left
// unverified — both modes are ordinary machinery.
//
//   - Temur is a WheneverYouCast trigger with a target clause, exactly
//     the shape Targeting() was built for: the spell being cast is
//     never touched, only the announced player.
//   - Sultai is Courser of Kruphix's standing permission one zone
//     over — graveyard rather than the top of the library, so no
//     LibraryTopVisible is needed (a graveyard is a public zone, CR
//     400.2) — gated to the Sultai word so a Temur Siege grants
//     nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "295a831c-490e-42ba-afc1-dab3524a4f0c",
		Name:         "Glacierwood Siege",
		Completeness: CompletenessFull,
		AsEnters:     ChooseOptionAsEnters("Glacierwood Siege", "Temur", "Sultai"),
		Triggered: []game.TriggeredAbility{
			WhenChosen("Temur", Targeting(
				WheneverYouCast(Or(Instant(), Sorcery()),
					"Glacierwood Siege — target player mills four cards", glacierwoodSiegeMill),
				TargetPlayer("target player"))),
		},
		GatedCastPermissions: []game.CastPermissionGate{{
			ActiveWhen: ChosenIs("Sultai"),
			Permission: game.CastPermission{
				Zone:   game.ZoneGraveyard,
				Filter: game.PermissionFilter{LandsOnly: true},
				Label:  "Play a land from your graveyard (Glacierwood Siege)",
			},
		}},
	})
}

// glacierwoodSiegeMill is the Temur body.
func glacierwoodSiegeMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		return MillCards{Player: t.ID, N: 4}.Apply(ctx)
	}
	return nil
}
