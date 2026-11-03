package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eriette of the Charmed Apple — Legendary Creature — Human Warlock
// {1}{W}{B}, 2/4 (Secrets of Strixhaven Commander, #4834):
//
//	"Each creature that's enchanted by an Aura you control can't attack
//	 you or planeswalkers you control.
//	 At the beginning of your end step, each opponent loses X life and
//	 you gain X life, where X is the number of Auras you control."
//
// The static is the "can't attack you or planeswalkers you control"
// restriction (CR 508.1c) that Promise of Loyalty grants, written by a
// layer 6 static onto every creature an Aura its controller controls is
// attached to: "you" is the controller of Eriette when the layers run,
// so it follows a change of control, and the Aura's controller is read
// the same way. A Role token is controlled by the player who created it,
// even on another player's creature (the ruling). A battle you protect is
// not covered.
//
// X is counted as the end step trigger resolves (the ruling, CR 608.2h);
// life loss, not damage. You gain X once, however many opponents lost it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f43f5855-f297-4433-984b-18263b734cfc",
		Name:         "Eriette of the Charmed Apple",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{socB5EnchantedByYourAuraCantAttackYou()},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Eriette of the Charmed Apple — each opponent loses X life and you gain X life (X = Auras you control)",
				func(g *game.Game, item *game.StackItem) error {
					n := countControlled(g, item.Controller, func(c game.Card) bool { return c.IsAura() })
					if n <= 0 {
						return nil
					}
					return b40DrainEachOpponent(n)(g, item)
				}),
		},
	})
}
