package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Legion Loyalty — Enchantment {6}{W}{W} (EDHREC rank 2905):
//
//	"Creatures you control have myriad. (Whenever a creature with
//	 myriad attacks, for each opponent other than defending player,
//	 you may create a token copy that's tapped and attacking that
//	 player or a planeswalker they control. Exile the tokens at end
//	 of combat.)"
//
// Every attacker becomes one attacker per opponent. Myriad (CR
// 702.116) is a triggered ability the keyword grants, so the grant
// is written as the trigger itself, hung on the Loyalty and fired
// once per creature its controller declares as an attacker — the
// same per-creature shape a creature printed with myriad would have.
// The "you may" is the trigger's optional prompt; on yes, for each
// opponent other than the defending player (b17DefendingPlayer
// resolves the player behind a planeswalker attack) a token copy of
// the attacker (TokenCopyTemplate — name, types, P/T, oracle ID, so
// its own triggers and statics come along) enters TAPPED AND
// ATTACKING that player through CreateTokensAttackingForEffect:
// never declared, so it fires no attack trigger of its own — and no
// myriad of its own, which is the printed anti-loop (CR 508.4). A
// CR 603.7 delayed trigger exiles the copies at the beginning of
// the end of combat step. At a two-player table, or once every
// other opponent has left, the trigger is not queued at all.
//
// Two declared simplifications, both weaker than printed:
//
//   - One yes/no covers every other opponent; the printed card lets
//     you decline per opponent.
//   - The copies attack the player, never a planeswalker that player
//     controls.
//
// The exiled tokens linger in the exile zone as cards rather than
// ceasing to exist — the engine-wide token gap every "exile the
// tokens" card shares, not this card's.
func init() {
	Register(Spec{
		OracleID:     "24e39d26-6a2b-461e-b194-bda575bcf898",
		Name:         "Legion Loyalty",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Myriad asks once per attacker whether to make the copies for every other opponent, rather than letting you choose per opponent.",
			"The copies always attack the player, not a planeswalker they control.",
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b27AttackedAndAnotherOpponentRemains(ev, source, g)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Legion Loyalty — myriad: create token copies of the attacker tapped and attacking each other opponent?"},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				attacker := ev.CardID
				defender := b17DefendingPlayer(g, ev)
				return game.NewTriggeredItem(source, b27LegionLoyaltyLabel,
					func(g *game.Game, item *game.StackItem) error {
						return b27MyriadCopies(g, item, attacker, defender)
					})
			},
		}},
	})
}
