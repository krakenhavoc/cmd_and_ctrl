package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cool but Rude — Enchantment — Class, {1}{R}:
//
//	(Gain the next level as a sorcery to add its ability.)
//	Whenever you attack, you may discard a card. If you do, draw a
//	card.
//	{1}{R}: Level 2
//	Whenever you discard a card, this Class deals 2 damage to each
//	opponent.
//	{1}{R}: Level 3
//	When this Class becomes level 3, search your library for a card,
//	put it into your hand, shuffle, then discard a card at random.
//
// A Class that builds its own engine: level 1 turns every attack into
// a rummage, level 2 turns every rummage into a Lava Spike at the
// table, and level 3 tutors — for anything — with a random discard
// that the level-2 line immediately cashes in. The three lines are
// written to be read together and the engine needs nothing new for
// any of them (ADR 0071 shipped the levels in #757).
//
// Three details carry the card, and each is a place it could have
// been written stronger than printed:
//
//   - "WHENEVER YOU ATTACK" is ONE trigger per attack declaration,
//     not one per attacker. The engine emits EventAttack per creature,
//     so OncePerBatch (#587) declines every later event of the same
//     batch. Without it a three-creature attack would rummage three
//     times and, at level 2, deal six.
//   - "YOU MAY DISCARD A CARD. IF YOU DO, DRAW A CARD" is one
//     instruction with a conditional second half, not a loot: the
//     discard happens first and from a hand that has not been
//     refilled, so the card you pitch is chosen before you see the
//     replacement. b39MayDiscardThenDraw is that shape, and the draw
//     is the RUN's count — discard nothing and you draw nothing,
//     which is what makes "if you do" mean something.
//   - The level-2 line is GATED at level 2, so the discards the
//     level-1 line was already making at level 1 pinged nobody. That
//     is the whole reason a designation is a gate rather than a
//     predicate inside AppliesTo, and the same asymmetry Wizard Class
//     documents.
//
// The level-3 line is a "becomes level N" trigger, so it fires once,
// as the level is set, and cannot fire again. Its body is Gamble's:
// the random discard is chained off the search rather than written on
// the next line, because the search RETURNS while its prompt is still
// open and a discard taken before the tutored card is in hand would
// be drawn from the wrong hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f076303-d160-4e02-aa1e-9ed6040c3735",
		Name:         "Cool but Rude",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			LevelUp(2, ManaCost("{1}{R}")),
			LevelUp(3, ManaCost("{1}{R}")),
		},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, coolButRudeAttackLabel, coolButRudeRummage)),
			AtLevel(2, On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, "Cool but Rude — 2 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 2)
			})),
			BecomesLevel(3, "Cool but Rude — level 3: search your library for a card", coolButRudeTutor),
		},
	})
}

const coolButRudeAttackLabel = "Cool but Rude — you may discard a card; if you do, draw a card"

// coolButRudeRummage is the level-1 line. n == 1 with UpTo makes the
// "you may" a real decline, and returning the discarded count as the
// draw is "if you do".
func coolButRudeRummage(g *game.Game, item *game.StackItem) error {
	return b39MayDiscardThenDraw(1, false, coolButRudeAttackLabel,
		func(discarded int) int { return discarded })(NewContext(g, item))
}

// coolButRudeTutor is the level-3 line. "A card" is every card in the
// library, so the search declares no predicate; the discard is the
// search's continuation for the reason the file comment gives.
func coolButRudeTutor(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	return SearchLibrary{
		Player:  controller,
		Dest:    game.ZoneHand,
		Limit:   1,
		Shuffle: true,
		Reason:  "Cool but Rude — search your library for a card",
		Then: func(g *game.Game, _ []uuid.UUID) error {
			return g.DiscardRandomForEffect(controller, 1)
		},
	}.Apply(NewContext(g, item))
}
