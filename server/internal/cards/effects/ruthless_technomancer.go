package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruthless Technomancer — Creature — Human Wizard {3}{B}, 2/4 (EDHREC
// rank 2087):
//
//	"When this creature enters, you may sacrifice another creature
//	 you control. If you do, create a number of Treasure tokens equal
//	 to that creature's power.
//	 {2}{B}, Sacrifice X artifacts: Return target creature card with
//	 power X or less from your graveyard to the battlefield. X can't
//	 be 0."
//
// A Fling into Treasures, and a reanimator that eats them back. The
// ETB is Springbloom Druid's shape: the "you may" is the trigger's
// optional prompt, the creature is chosen as the trigger's target
// (any other creature you control — the picker the engine has for a
// choice among your own permanents), its power is read as it stands
// (counters and anthems included) just before the sacrifice, and
// that many Treasures follow.
//
// Two sandbox simplifications, declared, both weaker than printed:
//
//   - The creature is chosen when the trigger goes on the stack, not
//     on resolution (Springbloom's caveat): an opponent who removes
//     it in response fizzles the trigger, where printed you would
//     pick another. A second Technomancer cannot be chosen — the
//     "another" is by name, the same read Noxious Gearhulk uses.
//   - The activated ability is not implemented. "Sacrifice X
//     artifacts" is a cost with a variable count, and AbilityCost has
//     no X (the roadmap's "X on an activated ability" seam); the
//     target's "power X or less" depends on it too. Nothing is
//     charged for nothing: the ability is simply absent.
func init() {
	Register(Spec{
		OracleID:     "4e58ad76-37c7-4531-b207-6890b39a2679",
		Name:         "Ruthless Technomancer",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You pick the creature to sacrifice when the enter trigger goes on the stack rather than on resolution, so opponents can respond to the choice.",
			"The reanimation ability isn't implemented — you can't sacrifice X artifacts to return a creature card from your graveyard.",
		},
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      b06SelfETB,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Ruthless Technomancer — sacrifice another creature for Treasures equal to its power?"},
			Targets:        TargetCreature("another creature you control", YouControl(), b03NotNamed("Ruthless Technomancer")),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Ruthless Technomancer — sacrifice a creature, Treasures equal to its power",
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
						if err := (SacrificePermanent{Target: target}).Apply(ctx); err != nil {
							return err
						}
						if z := g.FindCardZoneForEffect(target); z != nil && z.Kind == game.ZoneBattlefield {
							return nil // not sacrificed: no Treasures
						}
						if power <= 0 {
							return nil
						}
						return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: power}.Apply(ctx)
					})
			},
		}},
	})
}
