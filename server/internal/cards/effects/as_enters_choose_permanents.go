package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// as_enters_choose_permanents.go — "As this enters, <do something to>
// <other permanents you choose>" (ADR 0170, #1893): Phylactery Lich's
// "put a phylactery counter on an artifact you control", Sin, Unending
// Cataclysm's "remove all counters from any number of artifacts,
// creatures, and enchantments".
//
// CR 614.1c makes "as [this] enters" a replacement effect, and CR 614.12
// puts the choice and its consequences before the permanent is on the
// battlefield. So this is the EntryCardChoice of ADR 0098 with the
// choose action (game.EntryCardChoose): the prompt opens while the entry
// is paused, the declaration's Then acts on what was named, and only
// then does the permanent arrive. Nothing can respond in between, and a
// trigger that watches the permanent enter sees it with the counters
// Then gave it.
//
// The entering permanent and anything entering with it are never
// candidates (CR 614.13a, entryChoiceCandidatesLocked).

// AsEntersChoosePermanents is the replacement row for an as-enters
// choice of other permanents. `question` is the prompt's header and the
// row label; `choice` must have Action game.EntryCardChoose and a Then.
// The chooser is the player the permanent enters for
// (EnteringPermanentChooser), and there is no "if you don't" branch.
func AsEntersChoosePermanents(name, question string, choice *game.EntryCardChoice) game.ReplacementEffect {
	return game.ReplacementEffect{
		RowLabel:        question,
		EntryOnly:       true,
		Watches:         []game.EventKind{game.EventZoneMove},
		SelfReplacement: true,
		Label:           name,
		PromptQuestion:  question,
		EntryCardChoice: choice,
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID
		},
		Controller: EnteringPermanentChooser,
	}
}

// cardsStillHere looks the candidates up, in order, skipping any that are gone.
func cardsStillHere(g *game.Game, ids []uuid.UUID) []game.Card {
	out := make([]game.Card, 0, len(ids))
	for _, id := range ids {
		if c, ok := g.LookupCardForEffect(id); ok {
			out = append(out, c)
		}
	}
	return out
}
