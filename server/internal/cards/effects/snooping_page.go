package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Snooping Page — Creature — Human Cleric {1}{W}{B}, 2/3 (Secrets of
// Strixhaven, #4456):
//
//	"Repartee — Whenever you cast an instant or sorcery spell that targets a
//	 creature, this creature can't be blocked this turn.
//	 Whenever this creature deals combat damage to a player, you draw a card
//	 and lose 1 life."
//
// Repartee is Informed Inkwright's trigger (informedInkwrightCast: the cast
// spell's own announced targets). "Can't be blocked this turn" belongs to
// this permanent and is skipped if it has left or come back as a new object
// (CR 400.7), as Ant-Man's is. The second ability is an ordinary combat
// damage trigger: draw, then lose 1 life, in the printed order.
//
// No simplification.
func init() {
	const label = "Snooping Page — Repartee: can't be blocked this turn"
	Register(Spec{
		OracleID:     "f07c07d7-00f1-44a5-b51f-0a22cc51fd2b",
		Name:         "Snooping Page",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, informedInkwrightCast, label,
				func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: "Snooping Page — can't be blocked"}.Apply(NewContext(g, item))
				}),
			WheneverThisDealsCombatDamageToAPlayer("Snooping Page — you draw a card and lose 1 life",
				Do(DrawCards{N: 1}, GainLife{Amount: -1})),
		},
	})
}
