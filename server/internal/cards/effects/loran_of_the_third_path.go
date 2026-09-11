package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loran of the Third Path — Legendary Creature — Human Artificer,
// {2}{W}, 2/1:
//
//	"Vigilance"
//	"When Loran enters, destroy up to one target artifact or
//	 enchantment."
//	"{T}: You and target opponent each draw a card."
//
// A three-mana Disenchant on a body, and then a repeatable draw that
// is politically cheap because the opponent draws too. Played as much
// for being a white creature that draws cards as for the removal.
//
// # "Up to one target"
//
// Min 0, Max 1 on the target clause. That is not decoration: it is
// what lets Loran be cast when there is no artifact or enchantment on
// the table at all. A Min-1 clause on a trigger with no legal target
// is removed from the stack, and a Min-1 clause at CAST time would
// make the creature uncastable — which is exactly the bug "up to"
// exists to prevent. The ETB is a trigger rather than an OnETB hook
// so the destroy can be responded to, as printed.
//
// The target clause is deliberately NOT narrowed to opponents'
// permanents: printed, Loran can blow up your own Ichor Wellspring,
// and that is a real line.
//
// # The draw ability
//
// "You and target opponent each draw a card" — the OPPONENT is the
// target, the controller is not, and the controller draws whether or
// not the opponent still can. Both draws happen on resolution; a
// target that left the table between announce and resolution fizzles
// the whole ability (CR 608.2b), which is printed behaviour for a
// single-target ability.
//
// {T} on a creature carries summoning sickness through
// AbilityCost.Tap, so Loran cannot draw the turn she lands. Vigilance
// does not change that and is not supposed to — it is one of the
// twelve keywords the combat code honours, so it does its real job
// of letting her attack and still tap for the draw later.
//
// No simplifications.
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
			Targets: TargetPermanent("up to one target artifact or enchantment",
				Or(Artifact(), Enchantment())).WithCount(0, 1),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Loran of the Third Path — destroy up to one artifact or enchantment",
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
