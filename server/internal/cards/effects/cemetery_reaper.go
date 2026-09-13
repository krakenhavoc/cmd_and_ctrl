package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cemetery Reaper — Creature — Zombie {1}{B}{B}, 2/2 (EDHREC rank
// 2179):
//
//	"Other Zombie creatures you control get +1/+1.
//	 {2}{B}, {T}: Exile target creature card from a graveyard. Create
//	 a 2/2 black Zombie creature token."
//
// The Zombie lord that makes its own Zombies. The anthem is
// TribalAnthem (others, yours only — the printed controller clause);
// the activated ability is a mana plus tap cost targeting a creature
// card in ANY graveyard, picked in the zone browser and re-checked as
// it resolves (CR 608.2b): the card is exiled and the token created
// only if the target is still there, since a single-target ability
// whose target left is countered by game rules, token and all.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70fa209f-5e79-44de-81ce-1d8d0d8c1006",
		Name:         "Cemetery Reaper",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Zombie"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, {T}: Exile target creature card from a graveyard. Create a 2/2 black Zombie",
			Cost:    Plus(ManaCost("{2}{B}"), TapCost()),
			Targets: targetCreatureInAnyGraveyard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
				return CreateToken{Controller: item.Controller, Template: BlackZombieToken(), N: 1}.Apply(ctx)
			},
		}},
	})
}
