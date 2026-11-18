package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Price of Loyalty — Sorcery {2}{R} (Adventures in the Forgotten Realms):
//
//	"Gain control of target creature until end of turn. Untap that creature.
//	 It gains haste until end of turn. If mana from a Treasure was spent to
//	 cast this spell, that creature gets +2/+0 until end of turn."
//
// Act of Treason's three clauses, then the Treasure rider read from the spell's
// payment record as it resolves (Devour Intellect's reading).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "f25b97f9-8ca2-40f7-9d55-47fa8bb2054e",
		Name:           "Price of Loyalty",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
		Targets:        TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			if err := (GainControl{
				Target:   target,
				Duration: DurationUntilEndOfTurn(ctx),
				Label:    "Price of Loyalty — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{Target: target, Keywords: []string{"haste"},
				Label: "Price of Loyalty — haste until end of turn"}).Apply(ctx); err != nil {
				return err
			}
			if ctx.ManaSpent().FromTreasure() {
				return BoostUntilEOT{Target: target, Power: 2, Label: "Price of Loyalty — +2/+0 until end of turn"}.Apply(ctx)
			}
			return nil
		},
	})
}
