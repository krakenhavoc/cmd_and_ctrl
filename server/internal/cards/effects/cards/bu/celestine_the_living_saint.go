package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Celestine, the Living Saint — Legendary Creature — Human Warrior {4}{W},
// 3/4 (Warhammer 40,000 Commander, S82 GAP batch 196, #5948):
//
//	"Flying, lifelink
//	 Healing Tears — At the beginning of your end step, return target creature
//	 card with mana value X or less from your graveyard to the battlefield,
//	 where X is the amount of life you gained this turn."
//
// The target clause is built as the trigger goes on the stack, reading the
// turn's life gained (every gain counts: lifelink, a spell, an ability) and
// judged again as it resolves. With no life gained X is 0, so only a creature
// card with mana value 0 (a token never is a card) can be chosen.
//
// No simplification.
func init() {
	trigger := effects.AtYourEndStep("Celestine, the Living Saint — return target creature card with mana value X or less from your graveyard to the battlefield",
		effects.B26ReturnFirstLegalGraveyardTargetToBattlefield)
	trigger.TargetsFrom = effects.GapB196HealingTearsTargets
	effects.Register(effects.Spec{
		OracleID:        "6503ed78-6499-47f6-856b-ee1003ce4a4f",
		Name:            "Celestine, the Living Saint",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
