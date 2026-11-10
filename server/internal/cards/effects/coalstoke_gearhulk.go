package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coalstoke Gearhulk — Artifact Creature — Construct {1}{B}{B}{R}{R}, 5/4
// (Aetherdrift, #4464):
//
//	"Menace, deathtouch
//	 When this creature enters, put target creature card with mana value 4
//	 or less from a graveyard onto the battlefield under your control with
//	 a finality counter on it. That creature gains menace, deathtouch, and
//	 haste. At the beginning of your next end step, exile that creature."
//
// Puppeteer Clique's reanimation with Uchbenbak's finality counter: the
// card comes back under the Gearhulk controller's control with the counter
// as an entry counter (CR 614.1c), and a card that left the graveyard in
// response, or was sent elsewhere by an entry replacement, gains nothing and
// is never exiled. The keywords have no stated duration, so they last as
// long as the creature stays (CR 611.2a, 400.7); the exile is a delayed
// trigger at the beginning of YOUR next end step (CR 603.7), and a creature
// that is no longer on the battlefield by then is left alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2cdbaa4e-8907-49dc-9fa9-0df70db3bf23",
		Name:            "Coalstoke Gearhulk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Coalstoke Gearhulk — put target creature card with mana value 4 or less from a graveyard onto the battlefield under your control with a finality counter on it; it gains menace, deathtouch and haste; exile it at the beginning of your next end step",
				coalstokeGearhulkReanimate),
				TargetCardInGraveyard("target creature card with mana value 4 or less from a graveyard", Creature(), ManaValueLE(4))),
		},
	})
}

func coalstokeGearhulkReanimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) == 0 {
		return nil
	}
	id := legal[0].ID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ReturnFromGraveyardWithCountersKeyedForEffect(id, item.Controller, false,
		map[string]int{game.CounterFinality: 1}, ctx.ItemTail(coalstokeGearhulkEntered))
}

// coalstokeGearhulkEntered is the rest of the trigger, once the creature is
// on the battlefield: its keywords and the delayed exile. An on-disk key.
var coalstokeGearhulkEntered = EnteredTail("coalstoke-gearhulk/entered", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	if entered == uuid.Nil || !onBattlefield(g, entered) {
		return nil
	}
	if err := (ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddKeywordsMod("menace", "deathtouch", "haste")},
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Coalstoke Gearhulk — it gains menace, deathtouch and haste",
	}).Apply(ctx); err != nil {
		return err
	}
	return ScheduleDelayedTrigger{
		Label:              "Coalstoke Gearhulk — exile it",
		Cards:              []uuid.UUID{entered},
		Body:               exileListedCardsBody,
		ControllerTurnOnly: true,
	}.Apply(ctx)
})
