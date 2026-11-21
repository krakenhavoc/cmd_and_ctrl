package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Perilous Snare — Artifact {2}{W}:
//
//	"Start your engines!
//	 When this artifact enters, exile target nonland permanent an
//	 opponent controls until this artifact leaves the battlefield.
//	 Max speed — {T}: Put a +1/+1 counter on target creature or Vehicle
//	 you control. Activate only as a sorcery."
//
// ADR 0138 (#2122), for the Sami Whammy deck (#2190). The exile is
// Static Prison's "until this leaves the battlefield" (CR 610.3): one
// exile, returned as the Snare leaves; a Snare removed before its
// trigger resolves exiles nothing (CR 610.3b). The max-speed ability is
// an ordinary sorcery-speed activated ability whose activation also
// needs max speed (CR 702.178a): greyed until then.
//
// No simplification.
func init() {
	exile := effects.ExileChosenTargetUntilThisLeaves("Perilous Snare — the exiled card returns when Perilous Snare leaves the battlefield")
	effects.Register(effects.Spec{
		OracleID:        "cff7499a-44b0-4a2a-b96c-ed2cafb1a90d",
		Name:            "Perilous Snare",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.B06SelfETB,
				Targets:   effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls()),
				Key:       "Perilous Snare — exile target nonland permanent an opponent controls until it leaves",
				Effect:    exile,
			},
		},
		Activated: []effects.ActivatedAbility{
			effects.MaxSpeedActivated(effects.ActivatedAbility{
				Label:        "Max speed — {T}: Put a +1/+1 counter on target creature or Vehicle you control. Activate only as a sorcery.",
				Cost:         effects.TapCost(),
				Targets:      effects.TargetPermanent("target creature or Vehicle you control", effects.Or(effects.Creature(), effects.OfSubtype("Vehicle")), effects.YouControl()),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			}),
		},
	})
}
