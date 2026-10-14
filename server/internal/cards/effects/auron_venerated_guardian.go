package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Auron, Venerated Guardian — Legendary Creature — Human Spirit Samurai
// {3}{W}, 2/5:
//
//	"Vigilance
//	 Shooting Star — Whenever Auron attacks, put a +1/+1 counter on it.
//	 When you do, exile target creature defending player controls with
//	 power less than Auron's power until Auron leaves the battlefield."
//
// The counter first, then a CR 603.12 reflexive trigger, so the target
// is chosen with the counter already on Auron (a 3-power Auron can take a
// 2-power creature on its first attack). "When you do" happens only if
// the counter was put: an Auron that left before the attack trigger
// resolved gets none and exiles nothing. The defending player rides the
// reflexive trigger's params, and "power less than Auron's" is the
// RelativeToSource clause, judged against Auron's power as the target is
// chosen and again as it resolves. The exile lasts until this Auron
// object leaves the battlefield (ExileUntil.ThisLeaves); the card then
// returns under its owner's control (CR 610.3c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5e166b42-252e-4f34-9610-dcddc04a62ff",
		Name:            "Auron, Venerated Guardian",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Shooting Star — put a +1/+1 counter on Auron", auronShootingStar),
		},
	})
}

var auronExileBody = game.ReflexiveBody("auron-venerated-guardian/exile",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		ts := ctx.LegalTargets()
		if len(ts) == 0 {
			return nil
		}
		return ExileUntil{Target: ts[0].ID, ThisLeaves: true,
			Label: "Auron, Venerated Guardian — the exiled creature returns when Auron leaves the battlefield"}.Apply(ctx)
	}),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return RelativeToSource(TargetCreature("target creature defending player controls with power less than Auron's power",
			ControlledBy(p.Player)), LesserPower())
	})

func auronShootingStar(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if self, ok := ctx.SourcePermanent(); !ok || self.Left {
		return nil
	}
	if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return ReflexiveTrigger{
		Label:  "Shooting Star — exile target creature defending player controls with power less than Auron's power until Auron leaves the battlefield",
		Body:   auronExileBody,
		Params: game.EffectParams{Player: b17DefendingPlayer(g, item.Trigger.Event)},
	}.Apply(ctx)
}
