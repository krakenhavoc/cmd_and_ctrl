package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shadowspear — Legendary Artifact — Equipment for {1} (EDHREC rank
// 314, the highest-ranked Equipment in this batch):
//
//	"Equipped creature gets +1/+1 and has trample and lifelink.
//	 {1}: Permanents your opponents control lose hexproof and
//	 indestructible until end of turn.
//	 Equip {2}"
//
// One mana for a stat line good enough to justify the card, and then
// an ability that is the format's cleanest answer to the two keywords
// that otherwise blank removal. It is played in decks that own no
// other Equipment purely for the second line.
//
// THE SECOND LINE IS A TURN-SCOPED LAYER 6 EFFECT THAT TAKES THINGS
// AWAY — a `removeKeywords` data record (ADR 0041 phase 3, #1497)
// over the permanents your opponents control as it resolves. Three
// notes on it:
//
//   - The affected SET is locked when the ability resolves (CR
//     611.2c: an effect from a resolving ability that changes
//     characteristics — and abilities are characteristics, CR 109.3 —
//     affects only the objects it found then). A permanent that
//     enters afterwards keeps its hexproof and indestructible, and a
//     permanent that changes control afterwards stays as it was.
//     Before tier 3a the set was re-read on every recompute and a
//     late entrant slipped past only by timestamp, which this card
//     declared as a caveat; ADR 0041's phase 3 amendment (owner
//     decision 2) pins it, which is the printed behaviour.
//   - "Your opponents" is read at resolution, from the ability's
//     controller, so the effect keeps working if the Shadowspear
//     itself is destroyed in response, which is also correct.
//   - It costs {1} and is not a one-shot: activate it once and every
//     removal spell you cast for the rest of the turn connects.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b27326f-e7b8-4a4d-b589-df459246d19a",
		Name:         "Shadowspear",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("trample", "lifelink"),
		},
		Activated: []ActivatedAbility{
			{
				Label:  "{1}: Permanents your opponents control lose hexproof and indestructible until end of turn",
				Cost:   ManaCost("{1}"),
				Effect: shadowspearStrip,
			},
			EquipAbility("{2}"),
		},
	})
}

// shadowspearStrip is "permanents your opponents control lose hexproof
// and indestructible until end of turn".
func shadowspearStrip(g *game.Game, item *game.StackItem) error {
	return untilEndOfTurn(NewContext(g, item), uuid.Nil, OpponentControls(),
		"Shadowspear — opponents' permanents lose hexproof and indestructible",
		game.RemoveKeywordsMod("hexproof", "indestructible"))
}
