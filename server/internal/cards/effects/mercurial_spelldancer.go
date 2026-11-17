package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mercurial Spelldancer — Creature — Phyrexian Rogue {1}{U}, 2/1:
//
//	"This creature can't be blocked.
//	 Whenever you cast a noncreature spell, put an oil counter on this
//	 creature.
//	 Whenever this creature deals combat damage to a player, you may remove
//	 two oil counters from it. If you do, when you next cast an instant or
//	 sorcery spell this turn, copy that spell. You may choose new targets
//	 for the copy."
//
// The second ability is a "you may" resolved as the trigger resolves
// (CR 603.5). The counters come off only if this same object still has two
// (CR 400.7), and only then is the delayed trigger made (CR 603.7): it
// waits for the next instant or sorcery its controller casts this turn,
// fires once (CR 603.7b) and ends with the turn. The copy is put on the
// stack above the spell with a chance at new targets (CR 707.10c).
//
// No simplification.
func init() {
	const copyLabel = "Mercurial Spelldancer — remove two oil counters; if you do, copy the next instant or sorcery you cast this turn"
	Register(Spec{
		OracleID:     "5b90f8ed-84b2-4306-9dfa-b5654ea0e4cd",
		Name:         "Mercurial Spelldancer",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(),
				"Mercurial Spelldancer — put an oil counter on this creature",
				Do(CounterOnThis{Kind: s82OneOilCounter, N: 1})),
			Optional(WheneverThisDealsCombatDamageToAPlayer(copyLabel,
				Do(Instruction(mercurialSpelldancerRemoveAndCopy))),
				"Mercurial Spelldancer — remove two oil counters to copy your next instant or sorcery this turn?"),
		},
	})
}

// mercurialSpelldancerRemoveAndCopy removes the two oil counters from this
// Spelldancer, if it is still the object that triggered and still has two,
// and only then schedules the delayed copy.
func mercurialSpelldancerRemoveAndCopy(ctx *Context) error {
	if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID)
	if !ok || c.Counters[s82OneOilCounter] < 2 {
		return nil
	}
	if err := ctx.Game.AddCounterForEffect(ctx.Item.SourceCardID, s82OneOilCounter, -2); err != nil {
		return err
	}
	return WhenYouNextCast("Mercurial Spelldancer — copy that spell",
		game.CastFilter{Types: []string{"Instant", "Sorcery"}}, copyTheSpellBody).Apply(ctx)
}
