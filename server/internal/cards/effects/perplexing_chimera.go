package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Perplexing Chimera — Enchantment Creature — Chimera {4}{U}, 3/3:
//
//	"Whenever an opponent casts a spell, you may exchange control of
//	 this creature and that spell. If you do, you may choose new
//	 targets for the spell. (If the spell becomes a permanent, you
//	 control that permanent.)"
//
// ADR 0104's exchange of a spell and a permanent (CR 701.12). Both
// halves are indefinite and share one CR 613.7 timestamp, so the
// opponent keeps the Chimera — and, controlling it now, gets the same
// trigger against everyone else, which is how the card plays.
//
// "That spell" is read off the triggering event at resolution
// (item.Trigger), never captured. The exchange is all or nothing
// (CR 701.12a): a Chimera that left the battlefield, or came back as a
// new object (CR 400.7), exchanges nothing, and neither does a spell
// that has left the stack. "If you do" is the exchange's own report.
//
// The "you may" declares Trade: answering yes trades this creature
// for that spell, and a bot reads that off the prompt to weigh the
// trade (ADR 0104 owner decision 8) instead of accepting every trigger
// it controls.
func init() {
	const label = "Perplexing Chimera — exchange control of this creature and that spell"
	trigger := Optional(On(game.EventCast, AnOpponentCast(nil), label,
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			if item.Trigger == nil || ctx.isNewSourceObject(item.SourceCardID) {
				return nil
			}
			return ExchangeControlOfSpellAnd{
				Spell:            item.Trigger.Event.CardID,
				Permanent:        item.SourceCardID,
				ChooseNewTargets: true,
				Label:            label,
			}.Apply(ctx)
		}), "Exchange control of Perplexing Chimera and that spell?")
	trigger.OptionalPrompt.Trade = true
	Register(Spec{
		OracleID:     "7d075b8a-a606-4590-b52b-b4ef3a9e342f",
		Name:         "Perplexing Chimera",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
