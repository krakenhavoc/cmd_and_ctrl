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
// "Put an Equipment card from your hand onto the battlefield" is a
// pick-from-hand prompt with a hand-to-battlefield move, and neither
// exists (the Archaeomancer's Map / Kodama of the East Tree gap; the
// only hand picker the engine has is the discard modal). Shipping
// the ability with an auto-pick would be a choice the player never
// made; shipping it as "put the first Equipment" would be the same.
// So the Mystic is a tutor and nothing more, which is weaker than
// printed, never stronger, and the caveat says so. It becomes whole
// the day a put-from-hand prompt lands.
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
