package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreamcaller Siren — Creature — Siren Pirate {2}{U}{U}, 3/3 (Ixalan, S82
// older sets):
//
//	"Flash
//	 Flying
//	 This creature can block only creatures with flying.
//	 When this creature enters, if you control another Pirate, tap up to two
//	 target nonland permanents."
//
// "Can block only creatures with flying" is Welkin Tern's block rule. The
// trigger's "if" is an intervening if (CR 603.4): another Pirate is read
// through the layered view, so a changeling counts, and is checked as the
// trigger would go on the stack and again as it resolves. The targets are
// chosen as it goes on the stack (CR 603.3d), none to two of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8ed7791d-6d75-4cd7-9428-c731f623b241",
		Name:            "Dreamcaller Siren",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.WithoutKeyword("flying")), "it can block only creatures with flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.InterveningIf(effects.CondBoard{
					Name: "you control another Pirate",
					Test: func(ctx *effects.Context) bool {
						return effects.MshB4ControlsAnotherHero(ctx.Game, ctx.Controller(), ctx.Source(), ctx.TextSubtype("Pirate"))
					},
				}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return effects.Self(ev, source, lki, g) && effects.MshB4ControlsAnotherHero(g, source.Controller, source.InstanceID, source.TextSubtype("Pirate"))
				}, "Dreamcaller Siren — tap up to two target nonland permanents",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range item.Targets {
							if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
								continue
							}
							if err := (effects.TapTarget{Target: t.ID}).Apply(ctx.AsGroupMember()); err != nil {
								return err
							}
						}
						return nil
					})),
				effects.TargetPermanent("up to two target nonland permanents", effects.Nonland()).WithCount(0, 2)),
		},
	})
}
