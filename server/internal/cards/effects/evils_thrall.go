package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Evil's Thrall — Sorcery {2}{R} (Marvel Super Heroes, #4455):
//
//	"Gain control of target creature until end of turn. If you control a
//	 Villain with greater mana value than that creature, gain control of
//	 that creature until the end of your next turn instead. Untap that
//	 creature. It gains haste until end of turn."
//
// Act of Treason with a longer theft. The duration is chosen as the spell
// resolves (CR 608.2c): "until end of turn", or ADR 0063's "until the end
// of your next turn" when you control a Villain whose mana value is greater
// than the target's. "Villain" is read through the spell's text (a changed
// creature-type word applies), and both mana values are the permanents'
// current ones (a token is 0; an {X} spell on the battlefield counts X as 0,
// CR 202.3e). The haste is only until end of turn, as printed, even when
// control lasts longer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d7562f89-4808-44ab-8dc9-fc4078e8a47b",
		Name:         "Evil's Thrall",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			duration := DurationUntilEndOfTurn(ctx)
			label := "Evil's Thrall — gain control until end of turn"
			if mshA3ControlsGreaterVillain(ctx, target) {
				duration = ctx.Game.UntilEndOfYourNextTurnDuration(ctx.Controller())
				label = "Evil's Thrall — gain control until the end of your next turn"
			}
			if err := (GainControl{Target: target, Duration: duration, Label: label}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Target:   target,
				Keywords: []string{"haste"},
				Label:    "Evil's Thrall — haste until end of turn",
			}.Apply(ctx)
		},
	})
}

// mshA3ControlsGreaterVillain reports whether the resolving item's
// controller controls a Villain with greater mana value than the target.
func mshA3ControlsGreaterVillain(ctx *Context, target uuid.UUID) bool {
	victim, ok := ctx.Game.LookupCardForEffect(target)
	if !ok {
		return false
	}
	villain := ctx.TextSubtype("Villain")
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.HasSubtype(villain) && c.ManaValue() > victim.ManaValue() {
			return true
		}
	}
	return false
}
