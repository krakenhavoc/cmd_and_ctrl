package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// You Look Upon the Tarrasque — Instant {4}{G}:
//
//	"Choose one —
//	 • Run and Hide — Prevent all combat damage that would be dealt to
//	   you and creatures you control this turn.
//	 • Gather Your Courage — Target creature gets +5/+5 and gains
//	   indestructible until end of turn. All creatures your opponents
//	   control able to block that creature this turn do so."
//
// Run and Hide is ADR 0108 §7's not-one-use shield on you and the
// creatures you control, combat damage only (Take the Bait's recipient
// set with a different type). Gather Your Courage is the boost and the
// keyword grant, then a Lure record that spares the caster's own
// creatures (#2050, ADR 0045's 2026-10-08 amendment): when the target is
// an opponent's creature attacking you, your creatures are not required
// to block it.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "231a6f8a-624a-4bff-8648-0b4ef910a6aa",
		Name:         "You Look Upon the Tarrasque",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Run and Hide — Prevent all combat damage that would be dealt to you and creatures you control this turn."),
			Mode("Gather Your Courage — Target creature gets +5/+5 and gains indestructible until end of turn. All creatures your opponents control able to block that creature this turn do so.",
				TargetCreature("target creature")),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			switch {
			case ctx.HasMode(0):
				return PreventDamageFromSource{Protect: ShieldYouAndYourPermanents("creature"), CombatOnly: true}.Apply(ctx)
			case ctx.HasMode(1):
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (BoostUntilEOT{Target: t.ID, Power: 5, Toughness: 5, Label: "Gather Your Courage — +5/+5"}).Apply(ctx); err != nil {
						return err
					}
					if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"}, Label: "Gather Your Courage — indestructible"}).Apply(ctx); err != nil {
						return err
					}
					return BlockRequirementUntilEOT{
						Target:           t.ID,
						Kind:             game.BlockRequirementLure,
						Label:            "Gather Your Courage — your opponents' creatures able to block it do so",
						ExceptController: item.Controller,
					}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
