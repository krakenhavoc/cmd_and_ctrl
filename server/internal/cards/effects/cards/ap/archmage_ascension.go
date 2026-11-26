package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archmage Ascension — Enchantment {2}{U} (Zendikar, S82 OLD b493):
//
//	"At the beginning of each end step, if you drew two or more cards this turn,
//	 you may put a quest counter on this enchantment.
//	 As long as this enchantment has six or more quest counters on it, if you
//	 would draw a card, you may instead search your library for a card, put that
//	 card into your hand, then shuffle."
//
// The trigger is Bennie Bracks' end-step read with the draw tally (Knowledge Is
// Power's), an intervening "if" read as it triggers and again as it resolves
// (CR 603.4). The second ability is an optional draw replacement (Dredge's
// shape, DrawInstead): it applies to each draw of its controller while six or
// more quest counters sit on it, the search replaces that draw alone (nothing
// is drawn) and the rest of a multi-card draw goes on.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5966c464-5f63-4c71-9317-e2ca79da5aba",
		Name:         "Archmage Ascension",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B15EndStepBegan(ev) && effects.Old493DrewTwoOrMoreCards(g, source.Controller)
			}, "Archmage Ascension — put a quest counter",
				effects.Do(effects.If{Cond: effects.CondBoard{Name: "you drew two or more cards this turn", Test: func(ctx *effects.Context) bool {
					return effects.Old493DrewTwoOrMoreCards(ctx.Game, ctx.Controller())
				}}, Then: []effects.Applier{effects.CounterOnThis{Kind: "quest", N: 1}}})),
				"Archmage Ascension — put a quest counter on it?"),
		},
		Replacements: []game.ReplacementEffect{{
			Watches:  []game.EventKind{game.EventDrawCard},
			Optional: true,
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventDraw && ev.DrawCount > 0 && src != nil &&
					ev.DrawPlayer == src.Controller && src.Counters["quest"] >= 6
			},
			DrawInstead: game.RegisterDrawInstead("archmage-ascension-search", effects.Old493ArchmageSearchInstead),
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			PromptQuestion: "Archmage Ascension — search your library for a card instead of drawing?",
			Label:          "Archmage Ascension — if you would draw a card, you may instead search your library for a card",
		}},
	})
}
