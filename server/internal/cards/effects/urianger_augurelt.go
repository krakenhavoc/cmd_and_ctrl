package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urianger Augurelt — Legendary Creature — Elf Advisor {W}{U}, 1/4
// (#3787, deck request #3657):
//
//	"Whenever you play a land from exile or cast a spell from exile,
//	 you gain 2 life.
//	 Draw Arcanum — {T}: Look at the top card of your library. You may
//	 exile it face down.
//	 Play Arcanum — {T}: Until end of turn, you may play cards exiled
//	 with Urianger Augurelt. Spells you cast this way cost {2} less to
//	 cast."
//
// Three pieces, each a shape the engine now has (ADR 0091's and ADR
// 0066's 2026-10-15 amendments):
//
//   - DRAW ARCANUM is ExileTopFaceDownLinked: CR 701.20's look (only
//     the controller becomes a knower), a "you may" prompt over the one
//     card, then the CR 406.3 exile. The card lands FaceDownHidden, so
//     Urianger's controller may look at it and nobody else may, and it
//     carries Card.ExiledWith naming THIS Urianger as the object it is
//     now (CR 607.2a, 400.7): a Urianger that left and came back has
//     exiled nothing. The source is captured from the activation
//     (HiddenRefOfActivation), so an ability still on the stack when
//     Urianger dies still exiles the card against the old object.
//   - PLAY ARCANUM is PlayCardsExiledWith: a stored standing permission
//     over exile for the rest of the turn, narrowed to the cards that
//     link names, carrying CostLess 2. "Play": lands too, which cost
//     nothing to reduce. The reduction is generic only and never takes
//     the cost below {0} (CR 601.2f); it comes off after every increase.
//   - THE TRIGGER is Prosper, Tome-Bound's and Rocco's read of the same
//     two events: a spell cast from exile (EventCast carries the origin
//     in OldZone) or a land PLAYED from exile (Event.Played, CR 305.4).
//     Any grant counts, not only Urianger's own: a card an opponent's
//     Ragavan handed you gains you 2 life too, as printed.
//
// Both abilities need {T}, so a single Urianger does one of them per
// turn cycle unless something untaps it. The permission is checked when
// the card is played, so a card exiled by a second Draw Arcanum after
// Play Arcanum resolved (an untap effect) is playable the same turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ca584e7-4621-4b4e-a381-a0051acbf66a",
		Name:         "Urianger Augurelt",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventCast, game.EventZoneMove}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b20PlayedACardFromExile(ev, source, g)
			}, "Urianger Augurelt — you gain 2 life", Do(GainLife{Amount: 2})),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "Draw Arcanum — {T}: Look at the top card of your library. You may exile it face down.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return ExileTopFaceDownLinked{
						Source: HiddenRefOfActivation(item),
						Name:   "Draw Arcanum",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "Play Arcanum — {T}: Until end of turn, you may play cards exiled with Urianger Augurelt. Spells you cast this way cost {2} less to cast.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    TapCost(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return PlayCardsExiledWith{
						Source:   HiddenRefOfActivation(item),
						CostLess: 2,
						Label:    "Play Arcanum — play cards exiled with Urianger Augurelt (spells cost {2} less)",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
