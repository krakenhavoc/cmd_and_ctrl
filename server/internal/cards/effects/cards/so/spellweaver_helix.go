package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spellweaver Helix — Artifact {3} (Mirrodin, #5606):
//
//	"Imprint — When this artifact enters, you may exile two target sorcery
//	 cards from a single graveyard.
//	 Whenever a player casts a card, if it has the same name as one of the
//	 cards exiled with this artifact, you may copy the other. If you do, you
//	 may cast the copy without paying its mana cost."
//
// Isochron Scepter's linked exile over two sorcery cards from one graveyard
// (Decompose's single-graveyard rule), and its copy-and-cast of an exiled card
// (CR 707.12) for "the other" one. The cast watch is an intervening "if"
// (CR 603.4), checked when a card is cast and again as the ability resolves; a
// copy that is cast is no card, so it never triggers it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9aa8eef1-fc67-4d54-9783-4d0175a76741",
		Name:         "Spellweaver Helix",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Spellweaver Helix — you may exile two target sorcery cards from a single graveyard", effects.ExileLinkedTargetsEffect),
				effects.TargetCardInGraveyard("two target sorcery cards from a single graveyard", effects.Sorcery()).WithCount(2, 2).AllShare(effects.FromASingleGraveyard())),
				"Spellweaver Helix — exile two target sorcery cards from a single graveyard?"),
			effects.Optional(effects.On(game.EventCast, effects.OldB606HelixCast,
				"Spellweaver Helix — you may copy the other exiled card and cast the copy without paying its mana cost", effects.OldB606HelixCopyTheOther),
				"Spellweaver Helix — copy the other exiled card?"),
		},
	})
}
