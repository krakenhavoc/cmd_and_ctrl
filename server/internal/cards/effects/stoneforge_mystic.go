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
// Sandbox simplification, declared — the Ketria Triome posture, one
// whole ability omitted: the activated ability is NOT implemented.
// So the Mystic is a tutor and nothing more, which is weaker than
// printed, never stronger, and the caveat says so.
//
// The machinery it was waiting on has since landed and the ability
// is now a small edit rather than a seam: the pick-from-hand prompt
// shipped in #552 and the hand-to-battlefield move in #654, so
// "{1}{W}, {T}: You may put an Equipment card from your hand onto
// the battlefield" is an ActivatedAbility over
// PutFromHandOntoBattlefield{Match: …Equipment…, Optional: true}.
// Writing it is card work with its own tests, which is why #654 left
// it for a roadmap batch.
func init() {
	Register(Spec{
		OracleID:     "358789f9-7d87-411d-919e-d597da665cbd",
		Name:         "Stoneforge Mystic",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The {1}{W}, {T} ability isn't implemented — Equipment has to be cast from your hand the ordinary way."},
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
