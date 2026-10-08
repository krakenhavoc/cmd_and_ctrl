package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Celestus — {3} Legendary Artifact (#2190, seam #2561):
//
//	"If it's neither day nor night, it becomes day as The Celestus enters.
//	 {T}: Add one mana of any color.
//	 {3}, {T}: If it's night, it becomes day. Otherwise, it becomes
//	 night. Activate only as a sorcery.
//	 Whenever day becomes night or night becomes day, you gain 1 life.
//	 You may draw a card. If you do, discard a card."
//
// The first card on day and night (ADR 0132), and the one that uses all
// of it: the as-enters clause is AsEnters (off the stack, CR 614.12, so
// it is already day by the time anything else sees the permanent), the
// toggle is a sorcery-speed activation, and the trigger fires on a
// FLIP only — the first designation a game gains is not "day becomes
// night". It also fires on the Celestus's own toggle and on the
// untap-step check, whoever's turn it is, because the flip belongs to the
// game.
//
// The trigger's "you may draw a card. If you do, discard a card" is one
// linked clause (CR 607.2), asked when the ability resolves; declining
// the draw declines the discard. The mana ability is the ordinary
// any-colour tap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0ad2b5f-066b-424b-bddf-d3014731e599",
		Name:         "The Celestus",
		Completeness: CompletenessFull,
		AsEnters:     BecomesDayAsEnters(),
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{3}, {T}: If it's night, it becomes day. Otherwise, it becomes night. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}"), TapCost()),
			SorcerySpeed: true,
			Effect:       Do(ToggleDayNight{}),
		}},
		Triggered: []game.TriggeredAbility{
			WheneverDayBecomesNightOrNightBecomesDay("The Celestus — gain 1 life, may loot", celestusFlip),
		},
	})
}

// celestusFlip is the trigger's body: gain 1 life, then the linked
// "you may draw a card. If you do, discard a card".
func celestusFlip(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (GainLife{Amount: 1}).Apply(ctx); err != nil {
		return err
	}
	return MayChoice{
		Question: "The Celestus — draw a card, then discard a card?",
		YesLabel: "Draw",
		NoLabel:  "Decline",
		OnYes: func(ctx *Context) error {
			return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 1, 1, "")
		},
	}.Apply(ctx)
}
