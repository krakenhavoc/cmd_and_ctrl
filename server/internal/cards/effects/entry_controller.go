package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// entry_controller.go — "This enchantment enters under the control of
// an opponent of your choice" (ADR 0102, #1759). The engine half —
// the prompt, CR 616.1b's ordering, the would-be controller and the
// landing — is game/entry_controller.go. A card file writes one line:
//
//	Replacements: []game.ReplacementEffect{
//	    EntersUnderTheControlOfAnOpponentOfYourChoice("Captive Audience", game.ControlForHarm),
//	},
//
// The purpose is the card's own reading of what the gift does to its
// recipient — harm (Captive Audience, Xantcha) or benefit (Pendant of
// Prosperity). The engine never reads it; the bot does.
//
// Never hand-roll the effect: the constructor carries the self-scoped
// AppliesTo, the chooser (EnteringPermanentChooser — whoever the
// permanent would enter under), the selector and CR 616.1b's tier flag,
// and a hand-built one missing the flag would let Kismet or Authority
// of the Consuls be ordered before the control change.

// EntersUnderTheControlOfAnOpponentOfYourChoice is the whole clause
// (CR 614.1d, CR 614.12). `name` is the card's name, used for the
// prompt's header and the CR 616 label.
func EntersUnderTheControlOfAnOpponentOfYourChoice(name string, purpose game.ControlPurpose) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:         []game.EventKind{game.EventZoneMove},
		SelfReplacement: true,
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID
		},
		Controller: EnteringPermanentChooser,
		EntryController: &game.EntryControllerChoice{
			Purpose:  purpose,
			Question: name + " — choose an opponent to control it",
		},
		ChangesEntryController: true,
		Label:                  name + ": enters under the control of an opponent of your choice",
	}
}
