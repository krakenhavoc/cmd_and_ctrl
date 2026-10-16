package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Summon: Brynhildr — Enchantment Creature — Saga Knight {1}{R}, 2/1:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Chain — Exile the top card of your library. During any turn you
//	 put a lore counter on this Saga, you may play that card.
//	 II, III — Gestalt Mode — When you next cast a creature spell this
//	 turn, it gains haste until end of turn."
//
// A Saga creature. Chapter I is a conditional exile play (#3192, ADR
// 0066's second 2026-10-13 amendment): the grant lasts for as long as
// the card stays exiled and is live only during a turn a lore counter
// was put on THIS Saga (game.CondLoreCounterOnObject, naming the Saga
// object the chapter came from, so a Brynhildr that left and came back
// is a different Saga). Chapter I itself goes on the stack because of a
// lore counter, so the card is playable the turn it is exiled, and again
// on the turns chapters II and III are reached; once the Saga is
// sacrificed no lore counter can be put on it and the card stays exiled.
//
// Chapters II and III are Yuna, Grand Summoner's shape: a one-use promise
// on the next creature spell its controller casts this turn
// (game.NextSpellPromise), spent as that spell becomes cast. Its
// follow-up gives the spell haste until end of turn from this Saga
// (GrantKeywordsToSpellForEffect), which the permanent it becomes keeps
// (CR 400.7a). Not spent by the end of the turn, it lapses. The delayed
// trigger is not a stack object here, so it can't itself be countered.
//
// No simplification.
func init() {
	const name = "Summon: Brynhildr"
	gestalt := func(g *game.Game, item *game.StackItem) error {
		return GrantNextSpellPromise{
			From: name,
			Promise: game.NextSpellPromise{
				Filter:   game.PermissionFilter{CreatureOnly: true},
				FollowUp: brynhildrHaste,
				Text:     "When you next cast a creature spell this turn, it gains haste until end of turn.",
			},
		}.Apply(NewContext(g, item))
	}
	Register(Spec{
		OracleID:     "bab5e9ce-6d55-4d1e-a9ac-c2b954191be9",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "Chain — exile the top card of your library"),
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopWithPermission{
						From: item.Controller, GrantTo: item.Controller, N: 1,
						WhileExiled: true,
						Condition: &game.PermissionCondition{
							Key:         game.CondLoreCounterOnObject,
							ObjectID:    item.SourceObject.ID,
							ObjectEpoch: item.SourceObject.Epoch,
						},
					}.Apply(NewContext(g, item))
				}),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "Gestalt Mode — your next creature spell this turn gains haste"), gestalt),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "Gestalt Mode — your next creature spell this turn gains haste"), gestalt),
		},
	})
}

// brynhildrHaste is Gestalt Mode's "it gains haste until end of turn",
// run as the promised creature spell becomes cast.
var brynhildrHaste = game.RegisterCastFollowUp("summon-brynhildr-gestalt-haste",
	func(g *game.Game, f game.CastFollowUp) error {
		g.GrantKeywordsToSpellForEffect(f.Source, f.Spell, []string{"haste"}, g.UntilEndOfTurnDuration(),
			"Summon: Brynhildr — it gains haste until end of turn")
		return nil
	})
