package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Collector's Case — Artifact {1}{U} (Marvel Super Heroes Commander,
// #4455):
//
//	"When this artifact enters, tap up to one target creature and put two
//	 stun counters on it. (If a permanent with a stun counter would
//	 become untapped, remove one from it instead.)
//	 {3}{U}, {T}: Tap target creature."
//
// The enters trigger names up to one creature (Cryogen Relic's count
// shape), taps it and puts two stun counters on it (Alchemax Slayer-Bots'
// tap-and-stun with a second counter); with no target it does nothing. The
// activated ability is Arashin Sunshield's tap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a347123d-191a-4758-aab1-9cc7769eac66",
		Name:         "Collector's Case",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Collector's Case — tap up to one target creature and put two stun counters on it",
				msc5TapAndStunTwice),
				TargetCreature("up to one target creature", Creature()).WithCount(0, 1)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{3}{U}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}

// msc5TapAndStunTwice taps each target still legal and puts two stun
// counters on it.
func msc5TapAndStunTwice(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
		return AddCounter{Target: t.ID, Kind: game.CounterStun, N: 2}.Apply(ctx)
	}
	return nil
}
