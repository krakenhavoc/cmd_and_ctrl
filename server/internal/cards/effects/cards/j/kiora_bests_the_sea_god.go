package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kiora Bests the Sea God — Enchantment — Saga {5}{U}{U} (Theros Beyond Death, #5514):
//
//	"I — Create an 8/8 blue Kraken creature token with hexproof.
//	 II — Tap all nonland permanents target opponent controls. They don't untap
//	 during their controller's next untap step.
//	 III — Gain control of target permanent an opponent controls. Untap it."
//
// Chapter II taps every nonland permanent the targeted opponent controls as it
// resolves and freezes exactly those (TapAndFreeze: "they" are the ones tapped).
// Chapter III takes the permanent with no duration (CR 611.2a), so it stays with
// its new controller after the Saga is sacrificed, and untaps it once the
// control change has happened.
//
// No simplification.
func init() {
	const name = "Kiora Bests the Sea God"
	effects.Register(effects.Spec{
		OracleID:     "84d0c616-2fce-47f4-bfdd-783741a63fd2",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create an 8/8 blue Kraken creature token with hexproof"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("8/8 blue Kraken with hexproof"), N: 1}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "tap all nonland permanents target opponent controls; they don't untap during their controller's next untap step"),
				effects.TargetPlayer("target opponent", effects.Opponent()),
				kioraTapAndFreeze),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "gain control of target permanent an opponent controls; untap it"),
				effects.TargetPermanent("target permanent an opponent controls", effects.OpponentControls()),
				kioraStealAndUntap),
		},
	})
}

func kioraTapAndFreeze(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		var ids []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == t.ID && !c.IsLand() {
				ids = append(ids, c.InstanceID)
			}
		}
		return effects.TapAndFreeze{Targets: ids, Player: t.ID, Label: "Kiora Bests the Sea God — they don't untap during their controller's next untap step"}.Apply(ctx)
	}
	return nil
}

func kioraStealAndUntap(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.GainControl{Target: t.ID, Duration: game.IndefiniteDuration(), Label: "Kiora Bests the Sea God — gain control"}).Apply(ctx); err != nil {
			return err
		}
		return effects.UntapTarget{Target: t.ID}.Apply(ctx)
	}
	return nil
}
