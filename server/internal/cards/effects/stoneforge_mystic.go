package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stoneforge Mystic — Creature — Kor Artificer {1}{W}, 1/2 (EDHREC
// rank 1089):
//
//	"When this creature enters, you may search your library for an
//	 Equipment card, reveal it, put it into your hand, then shuffle.
//	 {1}{W}, {T}: You may put an Equipment card from your hand onto
//	 the battlefield."
//
// The Equipment tutor. The ETB is Steelshaper's Gift on a body —
// optional, so the prompt always opens and "fail to find" is always
// an answer — and it is the half every Commander deck plays the card
// for.
//
// The activated ability is PutFromHandOntoBattlefield over an
// Equipment card, optional as printed ("You may put"). The Equipment
// is PUT, not cast: no cast trigger fires and no mana is paid for it.
// It arrives unattached, which is also as printed. A player with no
// Equipment in hand is never prompted, so the ability resolves doing
// nothing, exactly as the card does.
//
// No simplification. (This file carried a "not implemented" caveat
// until #3138; the pick-from-hand prompt (#552) and the hand-to-
// battlefield move (#654) it waited on had both landed.)
func init() {
	Register(Spec{
		OracleID:     "358789f9-7d87-411d-919e-d597da665cbd",
		Name:         "Stoneforge Mystic",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, {T}: You may put an Equipment card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{W}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutFromHandOntoBattlefield{
					Match:    OfSubtype("Equipment"),
					Optional: true,
					Label:    "Stoneforge Mystic — you may put an Equipment card from your hand onto the battlefield",
				}.Apply(NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Stoneforge Mystic — search for an Equipment card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: b09IsEquipmentCard,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Stoneforge Mystic — an Equipment card, revealed, to hand",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
