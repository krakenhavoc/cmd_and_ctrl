package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chivalric Alliance — Enchantment {1}{W} (EDHREC rank 3034):
//
//	"Whenever you attack with two or more creatures, draw a card.
//	 {2}, Discard a card: Create a 2/2 white and blue Knight creature
//	 token with vigilance."
//
// The Wilds of Eldraine go-wide draw engine. The trigger is
// b16PlayerAttackedWithAtLeast (Aurelia's / Firemane Commando's
// shape): the engine emits EventAttack per creature, so the ability
// fires on the declaration that brings the controller to two
// attackers and the per-label dedup (queued, on the stack, already
// fired this turn) declines every later one.
//
// DECLARED SIMPLIFICATION, weaker than printed: the Knight-making
// ability is not offered. "Discard a card" is a cost component the
// engine cannot express (AbilityCost carries tap, sacrifice, mana,
// life, loyalty and crew — the Fauna Shaman / Tortured Existence
// seam), and an ability with its cost omitted would be stronger than
// printed (#259). The enchantment is still recognisably itself
// without it: the two-attacker draw is the card. Declared weaker
// than printed too, as Aurelia: with an extra combat in the same turn
// the trigger would not fire again, though the engine has no extra
// combats, so nothing reaches it today.
func init() {
	Register(Spec{
		OracleID:     "9b8d9236-0452-4509-aca4-a53a399fff85",
		Name:         "Chivalric Alliance",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Knight-making ability isn't available — discarding a card isn't a cost the engine can pay."},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b16PlayerAttackedWithAtLeast(ev, source, g, 2, b28ChivalricAllianceLabel)
			}, b28ChivalricAllianceLabel, Do(DrawCards{N: 1}))),
		},
	})
}

// b28ChivalricAllianceLabel is the trigger's stack label — named
// because the "two or more" dedup matches on it.
const b28ChivalricAllianceLabel = "Chivalric Alliance — you attacked with two or more creatures: draw a card"
