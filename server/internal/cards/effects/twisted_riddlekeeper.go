package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twisted Riddlekeeper — Creature — Eldrazi Sphinx {8}, 5/5:
//
//	"Emerge {5}{C}{U} (You may cast this spell by sacrificing a creature
//	 and paying the emerge cost reduced by that creature's mana value.)
//	 When you cast this spell, tap up to two target permanents. Put a
//	 stun counter on each of them. (If a permanent with a stun counter
//	 would become untapped, remove one from it instead.)
//	 Flying"
//
// Emerge is the shared alternative cost (ADR 0135 §4). The reduction
// comes off the generic part only (CR 118.7a), so the {C} and the {U}
// are always owed. The cast trigger resolves above the spell: each
// target still legal is tapped and gets a stun counter (CR 122.1d).
//
// No simplification.
func init() {
	cast := WhenYouCastThisSpell("Twisted Riddlekeeper — tap up to two target permanents and stun them",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			for _, id := range legalTargetCards(item, g) {
				if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				if err := (AddCounter{Target: id, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	cast.Targets = TargetPermanent("up to two target permanents").WithCount(0, 2)
	Register(Spec{
		OracleID:         "c7ad20a3-51bf-4a22-a09c-fff8cae22765",
		Name:             "Twisted Riddlekeeper",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Emerge("{5}{C}{U}")},
		Triggered:        []game.TriggeredAbility{cast},
	})
}
