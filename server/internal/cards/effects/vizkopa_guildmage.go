package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vizkopa Guildmage — Creature — Human Wizard {W}{B}, 2/2 (EDHREC
// rank 4471):
//
//	"{1}{W}{B}: Target creature gains lifelink until end of turn.
//	 {1}{W}{B}: Whenever you gain life this turn, each opponent loses
//	 that much life."
//
// The two-mana guildmage that ends games in a lifegain deck: switch
// the second ability on, then gain a pile of life, and the whole
// table loses it. The first ability is the cheap way to make the
// second one fire — a lifelinked commander connecting for eight
// drains three opponents for eight each.
//
// The second ability makes a floating triggered ability, which is a
// repeating DelayedOnEvent (CR 603.7b, Lyra's shape): "whenever you gain
// life" is an EventChangeLife on its controller with a positive amount,
// once per gain, and the body reads that amount off the fired item's
// trigger ("that much"). It lapses at end of turn. Activating it twice
// stacks two watchers, as printed.
//
// The first ability is complete. Lifelink is a canonical keyword and
// the grant is turn-scoped, pinned to the targeted creature with its
// battlefield-entry stamp, so a creature flickered in response is a
// new object and is not granted anything (CR 400.7, CR 611.2c).
func init() {
	Register(Spec{
		OracleID:     "f19e7c5c-67fa-4ae4-89b8-afa0e08a6c48",
		Name:         "Vizkopa Guildmage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}{B}: Target creature gains lifelink until end of turn.",
			Cost:    ManaCost("{1}{W}{B}"),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return GrantKeywordUntilEOT{
							Target:   t.ID,
							Keywords: []string{"lifelink"},
							Label:    "Vizkopa Guildmage — lifelink until end of turn",
						}.Apply(ctx)
					}
				}
				return nil
			},
		}, {
			Label:   "{1}{W}{B}: Whenever you gain life this turn, each opponent loses that much life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{1}{W}{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DelayedOnEvent{
					Label:     "Vizkopa Guildmage — each opponent loses that much life",
					On:        []game.EventKind{game.EventChangeLife},
					Condition: vizkopaGuildmageGainCondition,
					Body:      vizkopaGuildmageDrainBody,
					Repeats:   true,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

var (
	vizkopaGuildmageGainCondition = game.DelayedCondition("caveat-h2/vizkopa-guildmage-you-gain-life",
		func(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
			return ev.Kind == game.EventChangeLife && ev.Target == dt.Controller && ev.Amount > 0
		})
	vizkopaGuildmageDrainBody = game.SimpleDelayedBody("caveat-h2/vizkopa-guildmage-drain",
		func(g *game.Game, item *game.StackItem) error {
			if item.Trigger == nil {
				return nil
			}
			return eachOpponentLosesLife(g, item, item.Trigger.Event.Amount)
		})
)
