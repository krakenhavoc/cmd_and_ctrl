package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soulbright Flamekin — Creature — Elemental Shaman {1}{R}, 2/1 (Lorwyn,
// #5593):
//
//	"{2}: Target creature gains trample until end of turn. If this is the
//	 third time this ability has resolved this turn, you may add
//	 {R}{R}{R}{R}{R}{R}{R}{R}."
//
// Soulbright Seeker's count: the per-turn tally of this Flamekin's
// resolutions of this ability, which includes the one resolving
// (Game.ResolvedThisTurn), per object (CR 400.7). Only the third resolution
// offers the mana; the fourth and later grant trample alone. The mana is a
// "you may" asked as it resolves, and it does not depend on the target, so
// a target gone by then still lets the third resolution add it. Trample ends
// in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ea8426a-3b66-4fd0-9863-2b68d379671f",
		Name:         "Soulbright Flamekin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   soulbrightFlamekinLabel,
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCreature("target creature"),
			Purpose: soulbrightFlamekinPurpose(),
			Effect:  soulbrightFlamekinAbility,
		}},
	})
}

const soulbrightFlamekinLabel = "{2}: Target creature gains trample until end of turn. If this is the third time this ability has resolved this turn, you may add {R}{R}{R}{R}{R}{R}{R}{R}."

// soulbrightFlamekinPurpose declares the trample on the target and the
// eight red mana of the third resolution this turn.
func soulbrightFlamekinPurpose() game.Purpose {
	p := effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample"))
	p.ResolutionMana = &game.ResolutionMana{Resolution: 3, Mana: "{R}{R}{R}{R}{R}{R}{R}{R}"}
	return p
}

func soulbrightFlamekinAbility(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
		if err := (effects.GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"trample"},
			Label:    "Soulbright Flamekin — trample until end of turn",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	if g.ResolvedThisTurn(item.SourceCardID, soulbrightFlamekinLabel) != 3 {
		return nil
	}
	return effects.MayChoice{
		Question: "Soulbright Flamekin — add {R}{R}{R}{R}{R}{R}{R}{R}?",
		Then:     soulbrightFlamekinAddMana,
	}.Apply(ctx)
}

// soulbrightFlamekinAddMana is the "yes": eight red mana. An on-disk key.
var soulbrightFlamekinAddMana = effects.MayYes("may/soulbright-flamekin-add-mana", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.AddMana{Produced: "{R}{R}{R}{R}{R}{R}{R}{R}"}.Apply(ctx)
})
