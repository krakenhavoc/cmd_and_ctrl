package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rosie Cotton of South Lane — Legendary Creature — Halfling Peasant
// {2}{W}, 1/1 (EDHREC rank 1254):
//
//	"When Rosie Cotton enters, create a Food token. (It's an artifact
//	 with "{2}, {T}, Sacrifice this token: You gain 3 life.")
//	 Whenever you create a token, put a +1/+1 counter on target
//	 creature you control other than Rosie Cotton."
//
// The token-deck's counter engine. Her own Food triggers her second
// ability, as printed — the ETB resolves, the Food enters, and the
// token trigger asks for a creature. The second ability is a targeted
// trigger: the harvester drops it with no prompt when Rosie is your
// only creature (CR 603.3d), and asks otherwise.
//
// "Other than Rosie Cotton" is by name (b03NotNamed): a trigger's
// target clause never receives its source, and in a singleton format
// the name is the creature. A token copy of Rosie could not be
// chosen either, which is weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "168d5711-4459-440f-8de4-aabffd47c44d",
		Name:         "Rosie Cotton of South Lane",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"A token copy of Rosie Cotton can't be chosen for the +1/+1 counter either."},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rosie Cotton of South Lane — create a Food", Do(CreateToken{Template: FoodToken(), N: 1})),
			{
				Watches: []game.EventKind{game.EventTokenCreated},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor == source.Controller
				},
				Targets: TargetCreature("target creature you control other than Rosie Cotton",
					YouControl(), b03NotNamed("Rosie Cotton of South Lane")),
				Key: "Rosie Cotton of South Lane — put a +1/+1 counter on target creature",
				Effect: func(g *game.Game, item *game.StackItem) error {
					if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
						return nil
					}
					return AddCounter{Target: item.Targets[0].ID, Kind: "+1/+1", N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
