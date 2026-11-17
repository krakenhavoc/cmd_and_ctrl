package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Siege-Gang Lieutenant — Creature — Goblin {3}{R}, 2/2 (Modern Horizons
// 3 Commander, #5468):
//
//	"Lieutenant — At the beginning of combat on your turn, if you
//	 control your commander, create two 1/1 red Goblin creature tokens.
//	 Those tokens gain haste until end of turn.
//	 {2}, Sacrifice a Goblin: This creature deals 1 damage to any
//	 target."
//
// Lieutenant is an ability word, so the condition is an intervening if
// (CR 603.4), Loyal Apprentice's: checked as the combat step begins and
// again as the trigger resolves (b18ControlsYourCommander, a commander
// you own and control). The haste is scoped to the two tokens just made,
// for the one turn.
//
// The activated ability is Siege-Gang Commander's at its own numbers: any
// Goblin you control may be sacrificed, the Lieutenant itself included,
// and the damage is then dealt by it as it last existed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e9d538b-a9e5-4d59-b7c1-208fe18110e3",
		Name:         "Siege-Gang Lieutenant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return StepBegan(game.StepBeginCombat, true)(ev, source, lki, g) &&
					b18ControlsYourCommander(g, source.Controller)
			}, "Siege-Gang Lieutenant — create two 1/1 red Goblins that gain haste",
				siegeGangLieutenantGoblins),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice a Goblin: This creature deals 1 damage to any target.",
			Cost:    Plus(ManaCost("{2}"), game.AbilityCost{SacrificeOther: b12SacrificeAGoblin()}),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return ctx.Game.DamageInstanceForEffect(func() error {
					for _, t := range ctx.LegalTargets() {
						if err := (DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			},
		}},
	})
}

// siegeGangLieutenantGoblins re-checks the lieutenant condition, creates
// the two Goblins and gives those two tokens haste until end of turn.
func siegeGangLieutenantGoblins(g *game.Game, item *game.StackItem) error {
	if !b18ControlsYourCommander(g, item.Controller) {
		return nil
	}
	ctx := NewContext(g, item)
	cursor := b25LastEventSeq(g)
	if err := (CreateToken{Template: RedGoblinToken(), N: 2}).Apply(ctx); err != nil {
		return err
	}
	for _, id := range b27TokensCreatedByAfter(g, item.Controller, cursor) {
		if err := (GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"haste"},
			Label:    "Siege-Gang Lieutenant — the Goblins gain haste",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
