package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dovin Baan — Legendary Planeswalker — Dovin {2}{W}{U}, starting
// loyalty 3:
//
//	"+1: Until your next turn, up to one target creature gets -3/-0 and
//	 its activated abilities can't be activated.
//	 −1: You gain 2 life and draw a card.
//	 −7: You get an emblem with "Your opponents can't untap more than
//	 two permanents during their untap steps.""
//
// THE EMBLEM is ADR 0109 §5's EmblemSpec.UntapCaps (#1899): Static Orb's
// ceiling, read by the untap step's walk of every seat's emblems beside
// the battlefield, with the emblem as the source, so it binds only an
// untap step whose active player is an opponent of the emblem's owner
// (CR 114.2). The active player chooses which two of their permanents
// untap (CR 502.3, ADR 0070's untap_choice prompt). The emblem leaves
// only with its owner (CR 800.4a).
//
// THE +1 is one data record pinned to the creature, until Dovin's
// controller's next turn: -3/-0 in layer 7c, and "its activated
// abilities can't be activated", which, as on Arrest, includes its mana
// abilities (a mana ability is an activated ability, CR 605.1a; only a
// card that says "unless they're mana abilities", such as Faith's
// Fetters, spares them), so both CantActivate and CantActivateMana. A
// creature that leaves the battlefield is a
// new object the record no longer follows (CR 400.7). "Up to one" may
// choose nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b33bfbf-e4e1-43de-9c96-ccdc2916510a",
		Name:         "Dovin Baan",
		Completeness: CompletenessFull,
		// The fallback for tokens, fixtures and the dev spawner; an
		// imported deck reads printed loyalty (ADR 0032 §1).
		StartingLoyalty: 3,
		Emblem: &EmblemSpec{
			Label: "Dovin Baan emblem",
			Text:  "Your opponents can't untap more than two permanents during their untap steps.",
			UntapCaps: []game.UntapCap{
				opponentsCantUntapMoreThan("Your opponents can't untap more than two permanents during their untap steps.", 2, nil),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Until your next turn, up to one target creature gets -3/-0 and its activated abilities can't be activated.",
				Cost:    LoyaltyCost(1),
				Targets: TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  dovinBaanPlusOne,
			},
			{
				Label:  "−1: You gain 2 life and draw a card.",
				Cost:   LoyaltyCost(-1),
				Effect: Do(GainLife{Amount: 2}, DrawCards{N: 1}),
			},
			{
				Label: "−7: You get an emblem with \"Your opponents can't untap more than two permanents during their untap steps.\"",
				Cost:  LoyaltyCost(-7),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// dovinBaanPlusOne pins -3/-0 and "can't be activated" to the chosen
// creature until its controller's next turn. Nothing chosen, or a target
// that left in response, does nothing.
func dovinBaanPlusOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 || targets[0].Kind != game.TargetCard {
		return nil
	}
	return ScopedEffectFor{
		Target:   targets[0].ID,
		Mods:     []game.Mod{game.ModifyPTMod(-3, 0), game.AddRestrictionsMod(game.CantActivate | game.CantActivateMana)},
		Duration: DurationUntilYourNextTurn(ctx, ctx.Controller()),
		Label:    "Dovin Baan — -3/-0 and its activated abilities can't be activated until your next turn",
	}.Apply(ctx)
}
