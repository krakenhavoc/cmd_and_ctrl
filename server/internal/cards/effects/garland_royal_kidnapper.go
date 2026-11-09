package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Garland, Royal Kidnapper — Legendary Creature — Human Knight {2}{U}{B},
// 3/4 (Final Fantasy, #4462):
//
//	"When Garland enters, target opponent becomes the monarch.
//	 Whenever an opponent becomes the monarch, gain control of target
//	 creature that player controls for as long as they're the monarch.
//	 Creatures you control but don't own get +2/+2 and can't be
//	 sacrificed."
//
// The entry is Jared Carthalion's crowning of a targeted opponent. The
// steal is a layer-2 control change (CR 613.1b) pinned to the creature and
// lasting WhilePlayerIsMonarch (game.ForAsLongAsPlayerIsMonarchDuration, an
// additive duration condition): it ends when the crown leaves that player
// for anyone, and if they are no longer the monarch as the ability resolves
// it does nothing. The creature is chosen from those the new monarch
// controls (ControlledByThatEventActor, Black Bolt's clause), and "target
// opponent" can be you only through a crown that moved to another opponent.
// The static is two layered abilities over the same set: a creature under
// your control whose owner is someone else gets +2/+2 (layer 7c) and can't
// be sacrificed (layer 6), the stolen creature and a token an opponent
// made for you included, and a creature you own and control is untouched.
//
// No simplification.
func init() {
	notOwned := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller && target.Owner != source.Controller
	}
	Register(Spec{
		OracleID:     "ab914053-4384-415f-b1e8-09fecb5a30b8",
		Name:         "Garland, Royal Kidnapper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Garland, Royal Kidnapper — target opponent becomes the monarch",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer {
								return BecomeTheMonarch{Player: t.ID}.Apply(ctx)
							}
						}
						return nil
					}),
				TargetPlayer("target opponent", Opponent())),
			Targeting(
				On(game.EventMonarchChanged, AnOpponentBecameTheMonarch,
					"Garland, Royal Kidnapper — gain control of target creature that player controls for as long as they're the monarch",
					garlandSteal),
				ControlledByThatEventActor(TargetCreature("target creature that player controls"))),
		},
		Static: []game.StaticAbility{
			{
				Label:     "Creatures you control but don't own get +2/+2.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: notOwned,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
			{
				Label:     "Creatures you control but don't own can't be sacrificed.",
				Layer:     game.Layer6Ability,
				AppliesTo: notOwned,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.CantBeSacrificed = true
				},
			},
		},
	})
}

// garlandSteal takes the target for as long as the player who just became
// the monarch stays the monarch.
func garlandSteal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	d, ok := g.ForAsLongAsPlayerIsMonarchDuration(item.Trigger.Event.Actor, target)
	if !ok {
		return nil
	}
	return GainControl{Target: target, Controller: item.Controller, Duration: d,
		Label: "Garland, Royal Kidnapper — gain control while they're the monarch"}.Apply(ctx)
}
