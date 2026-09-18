package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barkform Harvester — Artifact Creature — Shapeshifter {3}, 2/3
// (EDHREC rank 4344):
//
//	"Changeling (This card is every creature type.)
//	 Reach
//	 {2}: Put target card from your graveyard on the bottom of your
//	 library."
//
// A colourless body that fits every tribal deck (changeling) and a
// repeatable graveyard tuck that is aimed at yourself: the Harvester is
// how a deck without recursion rebuys its best card, one card and two
// mana at a time, and how a self-mill deck stops decking itself.
//
// "YOUR graveyard" is YouOwn — a card in a graveyard has no controller,
// and the printed word is whose pile it is in. It is "target CARD", not
// "target creature card": a land, an instant, anything.
//
// The Harvester can target ITSELF once it is in the graveyard — the
// ability is on the permanent, so it can only be activated from the
// battlefield, and by then the Harvester is not in a graveyard. The
// case never arises, which is why there is no "another".
//
// Both keywords are canonical: changeling is the every-creature-type
// marker read by HasSubtype, so the Harvester really is an Elf for an
// Elvish Archdruid and a Sliver for a Sliver lord, and reach is read by
// the block-pair check.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0339bd11-ad71-4998-9b5d-a32790f0e5e3",
		Name:            "Barkform Harvester",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling, "reach"},
		Activated: []ActivatedAbility{{
			Label:   "{2}: Put target card from your graveyard on the bottom of your library",
			Cost:    ManaCost("{2}"),
			Targets: TargetCardInGraveyard("target card in your graveyard", YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return g.TuckToLibraryForEffect(t.ID, true)
				}
				return nil
			},
		}},
	})
}
