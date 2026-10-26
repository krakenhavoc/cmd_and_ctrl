package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	primalElderKittyPump  = "primal-elder-kitty/pump"
	primalElderKittyDies  = "primal-elder-kitty/dies-counters"
)

// Primal Elder Kitty — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Primal", "Elder", "Kitty"; three art stickers.
//	{TK}{TK} — {1}: This creature gets +1/-1 until end of turn.
//	{TK}{TK}{TK} — When this creature dies, you may put X +1/+1 counters on target creature, where X is this creature's power.
//	{TK}{TK} — 5/1
//	{TK}{TK}{TK}{TK} — 4/7
//
// The first sticker gives a pump ability. The second gives a dies trigger with a "you may" and a target creature: X is the creature's power as it last existed (CR 603.10a, counters included).
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7a00246-54e7-4213-b95b-907ab9015e53",
		Name:         "Primal Elder Kitty",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: primalElderKittyPump,
				Activated: []ActivatedAbility{{
					Label:  "{1}: This creature gets +1/-1 until end of turn.",
					Cost:   ManaCost("{1}"),
					Effect: thisGetsUntilEndOfTurn(1, -1, "Primal Elder Kitty sticker — +1/-1"),
				}},
				Text: "{1}: This creature gets +1/-1 until end of turn.",
			},
			{
				Key: primalElderKittyDies,
				Triggered: []game.TriggeredAbility{
					Optional(Targeting(
						WhenThisDies("Primal Elder Kitty sticker — you may put X +1/+1 counters on target creature, X this creature's power",
							func(g *game.Game, item *game.StackItem) error {
								ctx := NewContext(g, item)
								id, ok := stickerFirstLegalTarget(ctx)
								if !ok {
									return nil
								}
								died, ok := ctx.TriggeringPermanent()
								if !ok || died.Power <= 0 {
									return nil
								}
								return AddCounter{Target: id, Kind: game.CounterPlusOne, N: died.Power}.Apply(ctx)
							}),
						TargetCreature("target creature")),
						"Put +1/+1 counters on target creature?"),
				},
				Text: "When this creature dies, you may put X +1/+1 counters on target creature, where X is this creature's power.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: primalElderKittyPump},
			{Grant: primalElderKittyDies},
		}},
	})
}
