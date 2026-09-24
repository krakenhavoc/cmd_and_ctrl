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
// attackers, and the dedup — every later event of the same batch
// (OncePerBatch, see AGENTS.md §7), then the rest of the turn —
// declines every later one.
//
// The Knight-making ability is a CR 602 activation with a discard
// cost (#1381): Plus(ManaCost("{2}"), DiscardACard()), the same
// component Cryptbreaker's tap ability pays. Declared weaker than
// printed as Aurelia: with an extra combat in the same turn the
// trigger would not fire again, though the engine has no extra
// combats, so nothing reaches it today.
func init() {
	Register(Spec{
		OracleID:     "9b8d9236-0452-4509-aca4-a53a399fff85",
		Name:         "Chivalric Alliance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b16PlayerAttackedWithAtLeast(ev, source, g, 2, b28ChivalricAllianceLabel)
			}, b28ChivalricAllianceLabel, Do(DrawCards{N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}, Discard a card: Create a 2/2 white and blue Knight creature token with vigilance.",
			Cost:  Plus(ManaCost("{2}"), DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{
					Controller: item.Controller,
					Template:   TokenCard("2/2 white and blue Knight with vigilance"),
					N:          1,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

// b28ChivalricAllianceLabel is the trigger's stack label — named
// because the "two or more" dedup matches on it.
const b28ChivalricAllianceLabel = "Chivalric Alliance — you attacked with two or more creatures: draw a card"
