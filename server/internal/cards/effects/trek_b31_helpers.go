package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b31_helpers.go — shared pieces for the Star Trek Commander batch 31
// (#4974).

// trekB31SeekOut is the ability word "Seek out — Whenever you draw your
// second card each turn, …": Gixian Puppeteer's trigger condition (the
// second draw of any turn, never the third; b40IsYourSecondDrawThisTurn)
// with the card's own effect.
func trekB31SeekOut(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return b40IsYourSecondDrawThisTurn(ev, source, g)
	}, label, effect)
}

// trekB31ArtifactCreaturesProtectionFromChosenColor is Borg Incursion's
// "artifact creatures you control gain protection from the color of your
// choice until end of turn" once the colour is known. The affected set is
// read as the colour is answered, still inside the spell's resolution
// (CR 611.2c), so the two Borg the spell just made are in it. The tail's
// Word is the colour, its first Words entry the label's card name. An
// on-disk key: never renamed, never reused.
var trekB31ArtifactCreaturesProtectionFromChosenColor = CardTail("trek-b31/artifact-creatures-protection-from-chosen-color", func(ctx *Context, o game.TailOutcome) error {
	token := game.ProtectionFromColor(o.Tail.Word)
	if token == "" || len(o.Tail.Words) == 0 {
		return nil
	}
	return GrantKeywordUntilEOT{
		Match:    And(Artifact(), Creature(), YouControl()),
		Keywords: []string{token},
		Label:    o.Tail.Words[0] + " — " + token,
	}.Apply(ctx)
})

// trekB31AssimilateFromExileThen is assimilate (trekB11Assimilate's
// entry) on one card in exile, with `then` run once the card has landed
// (or at once, when it could not be put).
func trekB31AssimilateFromExileThen(ctx *Context, card uuid.UUID, label string, then game.KeyedTail) error {
	var entries []game.BatchEntry
	if z := ctx.Game.FindCardZoneForEffect(card); z != nil && z.Kind == game.ZoneExile {
		if c, ok := ctx.Game.LookupCardForEffect(card); ok && !c.IsToken() && c.IsPermanent() {
			entries = append(entries, game.BatchEntry{CardID: card, From: game.ZoneExile})
		}
	}
	if len(entries) == 0 {
		return then.Run(ctx.Game, nil)
	}
	return ctx.Game.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{
		Controller:         ctx.Controller(),
		EntersWithCounters: map[string]int{game.CounterPlusOne: 1},
		EntersAs: &game.EntersAs{
			AddTypes:      []string{"Artifact"},
			CreatureTypes: ctx.TextSubtypes("Borg"),
			Label:         label,
		},
	}, then)
}
