package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Doctor, EMH — Legendary Artifact Creature — Illusion Doctor
// {3}{G/W}{G/W}, 3/3 (Star Trek, #4973):
//
//	"When The Doctor enters, put a +1/+1 counter on target creature.
//	 Then put a +1/+1 counter on each creature you control with a
//	 counter on it."
//
// The target is any creature (CR 115.1). As the trigger resolves the
// target gets its counter first, then every creature its controller
// controls that has a counter of any kind on it at that moment gets one
// more: a target its controller controls gets two, an opponent's target
// one. With its only target gone the trigger does nothing at all: it
// does not resolve (CR 608.2b), so the second sentence does not happen
// either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3b04f5b9-40d9-49aa-a38d-0f946401263c",
		Name:         "The Doctor, EMH",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("The Doctor, EMH — put a +1/+1 counter on target creature, then on each creature you control with a counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
						for _, id := range creaturesWithACounter(g, item.Controller) {
							if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
								return err
							}
						}
					}
					return nil
				}),
				TargetCreature("target creature")),
		},
	})
}
