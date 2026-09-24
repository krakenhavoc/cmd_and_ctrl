package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disciple of Freyalise // Garden of Freyalise — the FRONT face,
// Creature — Elf Druid {3}{G}{G}{G}, 3/3:
//
//	"When this creature enters, you may sacrifice another creature.
//	 If you do, you gain X life and draw X cards, where X is that
//	 creature's power."
//
// The land back (pay 3 life or enter tapped; {T}: Add {G}) is the
// mdfc_lands.go row under "<oracle>#1"; this is face 0, which keeps
// the bare oracle ID (game.CatalogKey).
//
// Ruthless Technomancer's shape exactly, with the payoff swapped from
// Treasures to life and cards: the "you may" is the trigger's
// optional prompt, the sacrificed creature is chosen as the trigger's
// target when it goes on the stack (so an opponent can respond to the
// choice — the same declared sandbox timing Ruthless Technomancer and
// Springbloom Druid carry), its power is read just before the
// sacrifice, and SacrificePermanent.Then gates the payoff on the
// sacrifice's own answer (#993) rather than a live board re-read —
// which is what keeps a sacrificed COMMANDER's life and cards arriving
// even though it is still on the battlefield while its owner answers
// CR 903.9.
func init() {
	Register(Spec{
		OracleID:     "2699005b-a471-429f-a9d8-fbf2077ee2fd",
		Name:         "Disciple of Freyalise",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You pick the creature to sacrifice when the enter trigger goes on the stack rather than on resolution, so opponents can respond to the choice.",
		},
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      b06SelfETB,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Disciple of Freyalise — sacrifice another creature to gain that much life and draw that many cards?"},
			Targets:        TargetCreature("another creature you control", YouControl(), b03NotNamed("Disciple of Freyalise")),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Disciple of Freyalise — sacrifice a creature, gain life and draw equal to its power",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						target := item.Targets[0].ID
						if target == item.SourceCardID {
							return nil
						}
						ctx := NewContext(g, item)
						if !ctx.IsTargetLegal(item.Targets[0]) {
							return nil
						}
						victim, ok := g.LookupCardForEffect(target)
						if !ok {
							return nil
						}
						power := victim.CurrentPower()
						controller := item.Controller
						return SacrificePermanent{
							Target: target,
							Then: func(ctx *Context, sacrificed bool) error {
								if !sacrificed || power <= 0 {
									return nil
								}
								if err := (GainLife{Player: controller, Amount: power}).Apply(ctx); err != nil {
									return err
								}
								return DrawCards{Player: controller, N: power}.Apply(ctx)
							},
						}.Apply(ctx)
					})
			},
		}},
	})
}
