package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anara, Wolvid Familiar — Legendary Creature — Wolf Beast {3}{G},
// 4/4 (EDHREC rank 4066):
//
//	"During your turn, commanders you control have indestructible.
//	 Partner (You can have two commanders if both have partner.)"
//
// The green half of the Commander Legends familiar cycle, and the
// reason it is played: a commander that cannot be destroyed on your
// own turn attacks into a board of blockers and a board wipe held up
// by the player to your left, and survives both. Four mana for a 4/4
// that makes the rest of the deck's plan safe is a fine rate.
//
// THE WINDOW IS THE WHOLE POINT. "During YOUR turn" is a duration on
// a continuous effect (CR 611.2), not a trigger: the grant is live
// from the untap step to cleanup on the controller's turn and simply
// absent on everybody else's. It is modelled as a layer 6 static
// whose AppliesTo asks, on every recompute, whether the source's
// controller is the active player — so an opponent's Wrath of God on
// their own turn kills the commander, and the same Wrath cast in
// response to your attack (still your turn) does not. Recomputes run
// on every relevant state change, including the turn cursor moving,
// so nothing has to be unwound when the turn passes.
//
// "COMMANDERS YOU CONTROL", PLURAL AND WIDE. Anara herself is very
// often one of them — she has partner, so the usual configuration is
// Anara plus somebody, and she grants herself indestructible on your
// turn. A partner pair gets both members. A commander you have
// STOLEN is a commander you control and it is covered; your own
// commander somebody else has stolen is not, because the grant reads
// controller, not owner (CR 109.5).
//
// Indestructible is the engine's honoured keyword: "destroy" effects
// and lethal damage both bounce off it (CR 702.12b), while exile,
// sacrifice, -X/-X and bounce still answer the creature. That is the
// printed card, and it is why a commander wearing this is not
// actually safe.
//
// Partner is a deck-construction rule (CR 702.124h) that internal/deck
// reads off the oracle text (#2874): Anara and another partner card
// are a legal pair of commanders. Nothing about the indestructible
// grant depends on it — it reads Card.IsCommander.
//
// No simplification.
func init() {
	commanderYouControlOnYourTurn := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.IsCommander &&
			target.Controller == source.Controller &&
			isActivePlayer(g, source.Controller)
	}
	Register(Spec{
		OracleID:     "a59ff932-f758-476e-ba31-0623bd748231",
		Name:         "Anara, Wolvid Familiar",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer6Ability,
			AppliesTo: commanderYouControlOnYourTurn,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = append(c.Abilities, "indestructible")
			},
		}},
	})
}
