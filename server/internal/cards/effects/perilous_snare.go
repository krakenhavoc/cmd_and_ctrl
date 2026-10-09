package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	exile := exileChosenTargetUntilThisLeaves("Perilous Snare — the exiled card returns when Perilous Snare leaves the battlefield")
	Register(Spec{
		OracleID:        "cff7499a-44b0-4a2a-b96c-ed2cafb1a90d",
		Name:            "Perilous Snare",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Targets:   TargetPermanent("target nonland permanent an opponent controls", Nonland(), OpponentControls()),
				Key:       "Perilous Snare — exile target nonland permanent an opponent controls until it leaves",
				Effect:    exile,
			},
		},
		Activated: []ActivatedAbility{
			MaxSpeedActivated(ActivatedAbility{
				Label:        "Max speed — {T}: Put a +1/+1 counter on target creature or Vehicle you control. Activate only as a sorcery.",
				Cost:         TapCost(),
				Targets:      TargetPermanent("target creature or Vehicle you control", Or(Creature(), OfSubtype("Vehicle")), YouControl()),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			}),
		},
	})
}
