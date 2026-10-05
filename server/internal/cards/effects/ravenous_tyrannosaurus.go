package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ravenous Tyrannosaurus — "Devour 3 (As this creature enters, you may
// sacrifice any number of creatures. It enters with three times that
// many +1/+1 counters on it.) Whenever this creature attacks, it deals
// damage equal to its power to up to one other target creature. Excess
// damage is dealt to that creature's controller instead."
//
// Devour is devour.go's. The attack trigger splits the damage at CR
// 120.4a's lethal line: the creature takes what it still needed to die
// (toughness less damage already marked, floored at zero) and the rest
// goes to its controller. Power is read as the trigger resolves; a
// Tyrannosaurus that has left deals nothing. No simplifications.
func init() {
	Register(Spec{
		OracleID:     "9cfac390-4638-4654-8b50-65c0f0886b18",
		Name:         "Ravenous Tyrannosaurus",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{Devour("Ravenous Tyrannosaurus", 3)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			},
			Key:     "Ravenous Tyrannosaurus — damage equal to its power to up to one other target creature, excess to its controller",
			Targets: Another(TargetCreature("up to one other target creature")).WithCount(0, 1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				src, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok {
					return nil
				}
				victim, ok := g.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				power := src.CurrentPower()
				lethal := victim.CurrentToughness() - victim.DamageMarked
				if lethal < 0 {
					lethal = 0
				}
				toCreature, excess := power, 0
				if power > lethal {
					toCreature, excess = lethal, power-lethal
				}
				if err := (DealDamage{Source: item.SourceCardID, Target: id, Amount: toCreature}).Apply(ctx); err != nil {
					return err
				}
				return DealDamage{Source: item.SourceCardID, Target: victim.Controller, Amount: excess}.Apply(ctx)
			},
		}},
	})
}
