package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Breya, Etherium Shaper — Legendary Artifact Creature — Human
// {W}{U}{B}{R}, 4/4 (EDHREC rank 2027):
//
//	"When Breya enters, create two 1/1 blue Thopter artifact creature
//	 tokens with flying.
//	 {2}, Sacrifice two artifacts: Choose one —
//	 • Breya deals 3 damage to target player or planeswalker.
//	 • Target creature gets -4/-4 until end of turn.
//	 • You gain 5 life."
//
// The four-colour artifact commander, and her own fodder: the enters
// trigger makes the two Thopters the first activation eats.
//
// The modal ability is three activated abilities with the same cost,
// one per bullet — Insidious Fungus's shape, because a CR 602 ability
// has no mode picker and the client's ability menu is one; the mode is
// chosen at activation either way (CR 700.2). Each cost is a mana
// component plus a sacrifice clause with a count of two (#747,
// SacrificeN). Breya is an artifact, so she may be one of the two; the
// damage mode then still deals its damage, with Breya's last known
// information as the source.
//
// No simplification.
func init() {
	cost := Plus(ManaCost("{2}"), SacrificeN(2, "two artifacts", Artifact()))
	Register(Spec{
		OracleID:     "d460a9e2-5a7d-4562-880e-45174be19a9d",
		Name:         "Breya, Etherium Shaper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Breya, Etherium Shaper — create two Thopters",
				Do(CreateToken{Template: TokenCard("1/1 blue Thopter artifact with flying"), N: 2})),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, Sacrifice two artifacts: Breya deals 3 damage to target player or planeswalker",
				Cost:    cost,
				Targets: targetPlayerOrPlaneswalker("target player or planeswalker"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "{2}, Sacrifice two artifacts: Target creature gets -4/-4 until end of turn",
				Cost:    cost,
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return BoostUntilEOT{
								Target:    t.ID,
								Power:     -4,
								Toughness: -4,
								Label:     "Breya, Etherium Shaper — -4/-4",
							}.Apply(ctx)
						}
					}
					return nil
				},
			},
			{
				Label:  "{2}, Sacrifice two artifacts: You gain 5 life",
				Cost:   cost,
				Effect: Do(GainLife{Amount: 5}),
			},
		},
	})
}
