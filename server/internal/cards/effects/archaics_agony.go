package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archaic's Agony — Sorcery {4}{R} (Secrets of Strixhaven):
//
//	"Converge — Archaic's Agony deals X damage to target creature, where
//	 X is the number of colors of mana spent to cast this spell. Exile
//	 cards from the top of your library equal to the excess damage dealt
//	 to that creature this way. You may play those cards until the end of
//	 your next turn."
//
// X is the colours spent, read back at resolution (Painful Truths'
// converge: WantsDistinctColors and ReadsManaSpent). The excess is CR
// 120.4a's: what actually landed beyond the lethal damage the creature
// still needed (toughness less damage already marked), measured off the
// damage event so a prevention shield lowers it with the damage
// (socB20DealDamageWithExcess: Hell to Pay's reading plus the deathtouch
// clause, where 1 damage from a deathtouch source is lethal). With no
// excess nothing is exiled. The exiled cards are played, lands included, by the caster
// until the end of the caster's next turn (ADR 0063 duration).
func init() {
	Register(Spec{
		OracleID:            "55d1a788-fe9a-4e19-a524-6709f728ea28",
		Name:                "Archaic's Agony",
		Completeness:        CompletenessFull,
		ReadsManaSpent:      game.ReadsOwnPayment,
		WantsDistinctColors: true,
		Targets:             TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				excess, err := socB20DealDamageWithExcess(ctx, t.ID, ctx.ColorsSpentCount())
				if err != nil || excess <= 0 {
					return err
				}
				window := ctx.Game.UntilEndOfYourNextTurnDuration(item.Controller)
				return ExileTopWithPermission{
					From:    item.Controller,
					GrantTo: item.Controller,
					N:       excess,
					Window:  &window,
				}.Apply(ctx)
			}
			return nil
		},
	})
}
