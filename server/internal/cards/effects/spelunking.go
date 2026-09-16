package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spelunking — Enchantment {2}{G} (EDHREC rank 740):
//
//	"When this enchantment enters, draw a card, then you may put a
//	 land card from your hand onto the battlefield. If you put a Cave
//	 onto the battlefield this way, you gain 4 life.
//	 Lands you control enter untapped."
//
// The static is the card: every tapped dual, Temple and bounce land
// you play enters untapped. It is a CR 614 replacement on the entering
// land's move that clears EntersTapped. When the land carries its own
// "enters tapped" replacement, both apply to the same event and CR
// 616.1 has the land's controller order them — choosing Spelunking's
// last is what makes the land untapped, exactly the paper interaction.
// A land with no such clause sees Spelunking's alone, which is a no-op
// with no prompt.
//
// The ETB is "draw a card, THEN you may put a land card from your
// hand onto the battlefield" — the shared clause from #654,
// sequenced after the draw so a land just drawn is a legal pick, with
// the Cave rider in its Then continuation. The rider has to be a
// continuation rather than the next step of the Do: the prompt is
// asynchronous, so anything sequenced beside it would run before the
// player had answered and would be gaining life for a Cave nobody
// had put down yet.
//
// Sandbox simplification, weaker than printed:
//
//   - A land put onto the battlefield TAPPED by an effect (Cultivate,
//     Evolving Wilds, Riveteers Overlook — the SearchLibrary
//     TappedOnEntry flag) still enters tapped: that flag is applied
//     after the replacement pipeline and OR-ed with it. Only the
//     land's own enters-tapped clause is overridden.
func init() {
	Register(Spec{
		OracleID:     "2962fe4c-bf48-454b-8a6b-0f8253352ae8",
		Name:         "Spelunking",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"A land an effect puts onto the battlefield tapped still enters tapped; only a land's own enters-tapped text is overridden.",
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Spelunking — draw a card, then you may put a land from your hand onto the battlefield",
				Do(DrawCards{N: 1}, spelunkingLandDrop())),
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil {
					return false
				}
				if ev.Actor != src.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsLand()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.EntersTapped = false
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Spelunking: lands you control enter untapped",
		}},
	})
}

// spelunkingLandDrop is the ETB's second half: the shared "you may
// put a land card from your hand onto the battlefield" clause with
// Spelunking's own rider hung off its continuation.
//
// "If you put a CAVE onto the battlefield this way, you gain 4 life"
// — read off the permanent that actually entered, so a decline pays
// nothing and a Cave a replacement kept off the battlefield pays
// nothing either. Package-level rather than a closure so it captures
// nothing: the continuation outlives the trigger's resolution frame.
func spelunkingLandDrop() PutFromHandOntoBattlefield {
	p := MayPutALandFromHand("Spelunking")
	p.Then = func(g *game.Game, res PutFromHandResult) error {
		if res.Entered == uuid.Nil {
			return nil
		}
		c, ok := g.LookupCardForEffect(res.Entered)
		if !ok || !c.HasSubtype("Cave") {
			return nil
		}
		return g.ChangePlayerLifeForEffect(res.Source, res.Player, spelunkingCaveLife)
	}
	return p
}

// spelunkingCaveLife is the printed 4.
const spelunkingCaveLife = 4
