package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skrelv, Defector Mite — toxic 1, can't block, and a Phyrexian-white
// tap ability granting another creature toxic 1, hexproof from the
// chosen color and evasion against that color until end of turn.
// The target is announced first; color is chosen during resolution.
func init() {
	Register(Spec{
		OracleID:        "20053847-6623-493c-8cdb-a69cda3b1577",
		Name:            "Skrelv, Defector Mite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"toxic 1"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Activated: []ActivatedAbility{{
			Label:   "{W/P}, {T}: Choose a color. Another target creature you control gains toxic 1 and hexproof from that color until end of turn. It can't be blocked by creatures of that color this turn.",
			Cost:    Plus(ManaCost("{W/P}"), TapCost()),
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				targets := ctx.LegalTargets()
				if len(targets) == 0 {
					return nil
				}
				target := targets[0].ID
				ChooseColorThen(game.ColorForProtection, g, item.Controller, item.SourceCardID,
					"Skrelv, Defector Mite — choose a color", func(g *game.Game, color string) error {
						return ScopedEffectFor{
							Target: target,
							Mods: []game.Mod{
								game.AddKeywordsMod("toxic 1", game.HexproofFromColor(color)),
								game.CantBeBlockedByColorMod(color),
							},
							Duration: DurationUntilEndOfTurn(NewContext(g, item)),
							Label:    "Skrelv, Defector Mite — toxic 1, hexproof and evasion from " + game.ColorName(color),
						}.Apply(NewContext(g, item))
					})
				return nil
			},
		}},
	})
}
