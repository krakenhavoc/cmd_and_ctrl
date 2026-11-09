package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kinscaer Sentry — Creature — Kithkin Soldier {1}{W}, 2/2 (Lorwyn
// Eclipsed, #4458):
//
//	"First strike, lifelink
//	 Whenever this creature attacks, you may put a creature card with mana
//	 value X or less from your hand onto the battlefield tapped and
//	 attacking, where X is the number of attacking creatures you control."
//
// Nick Fury's put (hand_attacker_put.go): the creature enters tapped and
// attacking a player, planeswalker or battle its controller chooses (CR
// 508.4), never declared as an attacker so it triggers nothing that
// watches attacks. X is counted once, as the ability resolves (CR 608.2h),
// over the attacking creatures its controller has then; a card with {X} in
// its cost counts as X = 0. Nothing follows the entry, so the tail is a
// no-op.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fdf0fba8-f335-419c-b96b-e99037727c43",
		Name:            "Kinscaer Sentry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Kinscaer Sentry — you may put a creature card with mana value X or less from your hand onto the battlefield tapped and attacking, X the number of attacking creatures you control",
				kinscaerSentryAttack),
		},
	})
}

// kinscaerSentryPutter is the put: keys "choose-cards/kinscaer-sentry-put"
// and "option-pick/kinscaer-sentry-attacks".
var kinscaerSentryPutter = NewHandAttackerPutter("kinscaer-sentry", "Kinscaer Sentry", kinscaerSentryEntered)

// kinscaerSentryEntered follows the entry with nothing: "if you do" is not
// printed. An on-disk key.
var kinscaerSentryEntered = EnteredTail("kinscaer-sentry/entered", func(_ *Context, _ uuid.UUID, _ game.TailOutcome) error {
	return nil
})

func kinscaerSentryAttack(g *game.Game, item *game.StackItem) error {
	x := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && c.AttackingTarget != uuid.Nil {
			x++
		}
	}
	kinscaerSentryPutter.Offer(g, item, handCardsMatching(g, item.Controller, And(Creature(), ManaValueLE(x))),
		"Kinscaer Sentry — you may put a creature card with mana value X or less from your hand onto the battlefield tapped and attacking")
	return nil
}
