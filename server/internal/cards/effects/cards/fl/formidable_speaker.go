package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Formidable Speaker — Creature — Elf Druid {2}{G}, 2/4 (Lorwyn
// Eclipsed, #4458):
//
//	"When this creature enters, you may discard a card. If you do,
//	 search your library for a creature card, reveal it, put it into
//	 your hand, then shuffle.
//	 {1}, {T}: Untap another target permanent."
//
// The enters trigger is Gert and Old Lace's: the discard is an "up to
// one" prompt, and the search runs only if a card was discarded (CR
// 608.2c, "if you do"). The untap is Kelpie Guide's, without the "you
// control": any other permanent, an opponent's included.
//
// No simplification.
func init() {
	const label = "Formidable Speaker — you may discard a card; if you do, search for a creature card"
	effects.Register(effects.Spec{
		OracleID:     "5bd06f9e-9a42-455d-ad26-24bc04e51c7c",
		Name:         "Formidable Speaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player:   ctx.Controller(),
					Source:   ctx.Source(),
					N:        1,
					UpTo:     true,
					Question: label,
				}, formidableSpeakerSearch, nil)
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Untap another target permanent.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetPermanent("another target permanent", effects.Permanent())),
			Effect:  effects.B22UntapEachLegalTarget,
		}},
	})
}

// formidableSpeakerSearch is the "if you do": a creature card, revealed,
// to the hand. An on-disk key.
var formidableSpeakerSearch = game.RegisterDiscardThen("choose-cards/formidable-speaker-search", func(g *game.Game, r game.DiscardsDone) error {
	if r.Discarded.Count() == 0 {
		return nil
	}
	return effects.SearchLibrary{
		Player:    r.Player,
		Predicate: func(c game.Card) bool { return c.IsCreature() },
		Dest:      game.ZoneHand,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Formidable Speaker — a creature card, to your hand",
		Source:    r.Source,
	}.Apply(effects.KeyedContext(g, r.Player, r.Source))
})
