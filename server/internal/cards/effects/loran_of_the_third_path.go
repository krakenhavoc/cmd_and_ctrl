package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loran of the Third Path — Legendary Creature — Human Artificer
// {2}{W}, 2/1 (EDHREC rank 344):
//
//	"Vigilance
//	 When Loran enters, destroy up to one target artifact or
//	 enchantment.
//	 {T}: You and target opponent each draw a card."
//
// White's Reclamation Sage with a political tap ability stapled on.
// Three abilities, all real:
//
//   - Vigilance is printed and enforced.
//   - The ETB's "up to one target" is modelled as a "you may" trigger
//     with a single target clause — the same shape Sun Titan and
//     Reclamation Sage use. Observably identical: with no legal
//     target the trigger is removed without a prompt (CR 603.3d),
//     with one the controller may decline. The destroy goes through
//     the S20 pick_target prompt and the CR 608.2b re-check.
//   - The tap ability targets an opponent and draws one card for the
//     controller and one for the target. CR 302.1 summoning sickness
//     applies (a tap cost on a creature), as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b3d81980-76f2-44e2-b1c9-01e30c726312",
		Name:            "Loran of the Third Path",
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Loran of the Third Path — destroy up to one target artifact or enchantment?",
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Loran of the Third Path — destroy target artifact or enchantment",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						return DestroyTarget{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
					})
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}: You and target opponent each draw a card.",
			Cost:    TapCost(),
			Targets: TargetPlayer("target opponent", Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
					return err
				}
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				return DrawCards{Player: item.Targets[0].ID, N: 1}.Apply(ctx)
			},
		}},
	})
}
