package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Geth, Thane of Contracts — Legendary Creature — Phyrexian Zombie
// {1}{B}{B}, 3/4 (ONE):
//
//	"Other creatures you control get -1/-1.
//	 {1}{B}{B}, {T}: Return target creature card from your graveyard to the
//	 battlefield. It gains "If this creature would leave the battlefield,
//	 exile it instead of putting it anywhere else." Activate only as a
//	 sorcery."
//
// The static is a layer 7c anthem over the other creatures its controller
// controls (CR 613.4c). The returned creature comes back under its owner's
// control, which is its controller's here since the card is in their
// graveyard, and the granted replacement is Whip of Erebos's shared
// ExileInsteadOfLeavingBattlefield, pinned to that object until it leaves
// (CR 614.1a, 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b868de48-3271-4f33-94a4-3dd9b67dd72c",
		Name:         "Geth, Thane of Contracts",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Others: true, YoursOnly: true}, -1, -1),
		},
		Activated: []ActivatedAbility{{
			Label:        "{1}{B}{B}, {T}: Return target creature card from your graveyard to the battlefield. It gains \"If this creature would leave the battlefield, exile it instead of putting it anywhere else.\" Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}{B}{B}"), TapCost()),
			Targets:      targetCreatureInYourGraveyard(),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				id := item.Targets[0].ID
				if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(NewContext(g, item)); err != nil {
					return err
				}
				if onBattlefield(g, id) {
					ExileInsteadOfLeavingBattlefield(g, id, item.Controller,
						"Geth, Thane of Contracts: if it would leave the battlefield, exile it instead")
				}
				return nil
			},
		}},
	})
}
