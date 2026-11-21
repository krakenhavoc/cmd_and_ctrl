package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ossification — Enchantment — Aura {1}{W} (EDHREC rank 4295):
//
//	"Enchant basic land you control
//	 When this Aura enters, exile target creature or planeswalker an
//	 opponent controls until this Aura leaves the battlefield."
//
// Oblivion Ring for two mana, with the drawback moved from the mana
// cost to the deckbuilding: it needs a basic land, and a Commander
// mana base full of duals and fetch targets may not have one spare.
// The reward is that it is one mana cheaper than every other card
// printed with this text.
//
// "Until this Aura leaves the battlefield" is CR 610.3: the exile is a
// one-shot, and the return is a SECOND one-shot created immediately
// after the Aura leaves — not a triggered ability (#1729). ExileUntil
// records it, and the engine performs it before anyone gets priority,
// with no stack in between: destroy the Aura and the permanent is
// already back when the destroying spell has finished. It returns even
// if the Aura's owner leaves the game, which takes the Aura off the
// battlefield with them.
//
// CR 610.3b, and the ruling of 2023-02-04: "If Ossification leaves the
// battlefield before its triggered ability resolves, the target
// permanent won't be exiled." ExileUntil reads the Aura as the object
// it was when the ability triggered, so a removal in response leaves
// the target where it is.
//
// The leave trigger the card carried before #1729 stays as a legacy
// row (UntilThisLeavesLegacyReturn): it fires only for a card exiled by
// an older binary, with no record to bring it back.
//
// The exiled permanent returns under its OWNER's control (CR 610.3c),
// not under the Aura controller's: destroying an Ossification on your
// own Ravenous Chupacabra gives it back to you, and destroying one on
// an opponent's commander gives it back to them.
//
// The enchant clause is a REAL restriction, re-run every turn by the
// CR 704.5m legality check: a basic land that stops being one — or
// stops being yours — takes the Aura to the graveyard, which returns
// the exiled permanent. That is the card's honest weakness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e29bfd62-286f-4982-813f-7086573c333b",
		Name:         "Ossification",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant basic land you control", effects.B41BasicLandYouControl()),
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.B06SelfETB,
				Targets: effects.TargetPermanent("target creature or planeswalker an opponent controls",
					effects.B41CreatureOrPlaneswalkerAnOpponentControls()),
				Key:    effects.B41OssificationExileLabel,
				Effect: effects.ExileChosenTargetUntilThisLeaves("Ossification — the exiled card returns when Ossification leaves the battlefield"),
			},
			effects.UntilThisLeavesLegacyReturn("Ossification — return the exiled card", effects.B41OssificationExileLabel),
		},
	})
}
