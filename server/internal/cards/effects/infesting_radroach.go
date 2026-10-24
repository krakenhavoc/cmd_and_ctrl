package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Infesting Radroach — Creature — Insect Mutant {2}{B}, 2/2:
//
//	"Flying
//	 This creature can't block.
//	 Whenever this creature deals combat damage to a player, they get
//	 that many rad counters.
//	 Whenever an opponent mills a nonland card, if this creature is in
//	 your graveyard, you may return it to your hand."
//
// #2809. "That many" is the combat damage the Radroach dealt to that
// player, the event's amount.
//
// The last ability works only from the graveyard (InGraveyard, CR
// 113.6b), so on the battlefield it never triggers. It triggers once per
// nonland card an opponent of its owner mills (the engine's EventMill is
// per card; a surveil is not a mill, CR 701.25a), whoever made them mill
// — the rad counters' own trigger included. "If this creature is in your
// graveyard" is an intervening if (CR 603.4): the zone gate is the
// trigger-time half, and as it resolves the card must still be in the
// graveyard as the same object (CondThisInYourGraveyard, declared with
// InterveningIf), so a second trigger from the same mill is removed
// without a question once the first has returned it. The "you may" is
// asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "064f371c-e4f2-423b-9f35-443191dfbb4c",
		Name:            "Infesting Radroach",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			RestrictSelf(game.CantBlock),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Infesting Radroach — they get that many rad counters",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					ev := item.Trigger.Event
					return playerGetsRadCounters(g, item.Controller, ev.Target, ev.Amount)
				}),
			Optional(InterveningIf(CondThisInYourGraveyard{}, InGraveyard(On(game.EventMill, AllOf(ByAnOpponent, ANonlandCardWasMilled),
				"Infesting Radroach — return it from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
				}))), "Infesting Radroach — return it to your hand?"),
		},
	})
}
