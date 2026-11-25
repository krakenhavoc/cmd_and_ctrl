package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Evra, Halcyon Witness — Legendary Creature — Avatar {4}{W}{W}, 4/4
// (Dominaria, S82 OLD batch 214, #5530):
//
//	"Lifelink
//	 {4}: Exchange your life total with Evra's power."
//
// Tree of Redemption's exchange on power (CR 701.12g): Evra's power is read as
// the ability resolves (modifiers included) beside your life total. Your life
// becomes the old power by the necessary gain or loss, through the ordinary
// life-change pipeline; Evra's power is set to the old life total by a
// layer 7b base-power effect pinned to this object, with counters and Auras
// applying after it, a newer exchange overriding an older by timestamp. An
// Evra that has left the battlefield, or is a new object, does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cf4b538d-7115-406d-94bf-bfa25197b6ef",
		Name:            "Evra, Halcyon Witness",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}: Exchange your life total with Evra's power.",
			Cost:    effects.ManaCost("{4}"),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if g.AbilitySourceGoneForEffect(item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				p := g.PlayerByIDForEffect(item.Controller)
				if p == nil || !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				g.RecomputeLayersIfStaleLocked()
				evra, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				power, life := evra.CurrentPower(), p.Life
				if err := (effects.ScopedEffectFor{
					Target:   item.SourceCardID,
					Mods:     []game.Mod{game.SetBasePowerMod(life)},
					Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
					Label:    "Evra, Halcyon Witness — power becomes your former life total",
				}).Apply(ctx); err != nil {
					return err
				}
				if delta := power - life; delta != 0 {
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, delta)
				}
				return nil
			},
		}},
	})
}
