package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gyrus, Waker of Corpses — Legendary Creature — Hydra {X}{B}{R}{G}, 0/0
// (Commander 2018, #5526):
//
//	"Gyrus enters with a number of +1/+1 counters on it equal to the amount
//	 of mana spent to cast it.
//	 Whenever Gyrus attacks, you may exile target creature card with lesser
//	 power from your graveyard. If you do, create a token that's a copy of
//	 that card and that's tapped and attacking. Exile the token at end of
//	 combat."
//
// The entry counters are Kurbis' CR 614.1c clause over the cast
// (CastCounts.ManaSpent, ReadsManaSpent so a waived payment asks what was
// paid). Gyrus put onto the battlefield without being cast spent nothing
// and gets no counters.
//
// The attack trigger targets a creature card in your graveyard whose power
// is less than Gyrus' (RelativeToSource, LesserPower: judged against Gyrus'
// effective power as the target is chosen and again as it resolves,
// last-known if Gyrus has left), and the "may" is asked as it resolves
// (Mirko's shape). The exile is The Scarab God's ExileTargetThenTokenCopy:
// only a card that really went to exile is copied ("if you do"). The token
// is Flamerush Rider's: a copy of the card's copiable values, entering
// tapped and attacking a player, planeswalker or battle of your choice
// (Gyrus' own defender offered first), never declared as an attacker (CR
// 508.4). "Exile the token at end of combat" is a delayed trigger at the
// beginning of that step naming the token (CR 603.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "2deae472-3cfe-4d7b-b9cf-a045f07a002e",
		Name:           "Gyrus, Waker of Corpses",
		ReadsManaSpent: game.ReadsOwnPayment,
		Completeness:   CompletenessFull,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{{
			Kind:  game.CounterPlusOne,
			Count: func(cast game.CastCounts) int { return cast.ManaSpent },
		}},
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				WheneverThisAttacks("Gyrus, Waker of Corpses — you may exile target creature card with lesser power from your graveyard and create a tapped and attacking token copy of it; exile the token at end of combat",
					func(g *game.Game, item *game.StackItem) error {
						return ExileTargetThenTokenCopy(g, item, gyrusCopyTail)
					}),
				RelativeToSource(TargetCardInGraveyard("target creature card with lesser power from your graveyard",
					Creature(), YouOwn()), LesserPower())),
				"Exile a creature card with lesser power from your graveyard and create a tapped and attacking token copy of it?"),
		},
	})
}

// gyrusCopyTail is "If you do, create a token that's a copy of that card and
// that's tapped and attacking", run once the card has landed in exile. An
// on-disk key: never renamed, never reused.
var gyrusCopyTail = CardTail("then/gyrus-waker-of-corpses-copy", func(ctx *Context, o game.TailOutcome) error {
	if !oneWentThisWay(o) {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(ctx.Game, o.IDs[0])
	if !ok {
		return nil
	}
	return CreateTokensAttackingYourChoice{
		Template: tmpl,
		Copy:     true,
		N:        1,
		Tapped:   true,
		Prefer:   triggerAttackTarget(ctx.Item),
		Label:    "Gyrus, Waker of Corpses",
		Tail:     ctx.ItemTail(gyrusExileToken),
	}.Apply(ctx)
})

// gyrusExileToken is "Exile the token at end of combat." as a registered
// tail told the token that landed. An on-disk key: never renamed, never
// reused.
var gyrusExileToken = AttackingTokensThen("attacking-tokens/gyrus-waker-of-corpses-exile-at-end-of-combat",
	func(ctx *Context, created []uuid.UUID, tail game.KeyedTail) error {
		item := tail.Asked.StackItem()
		if len(created) == 0 || item == nil {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:    game.StepEndCombat,
			Label: "Gyrus, Waker of Corpses — exile the token",
			Cards: created,
			Body:  exileListedCardsBody,
		}.Apply(NewContext(ctx.Game, item))
	})
