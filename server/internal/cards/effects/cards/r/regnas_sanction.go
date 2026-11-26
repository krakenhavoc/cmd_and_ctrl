package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Regna's Sanction — Sorcery {3}{W} (Battlebond, #5529):
//
//	For each player, choose friend or foe. Each friend puts a +1/+1 counter on each creature they control. Each foe chooses one untapped creature they control, then taps the rest.
//
// effects.NewS82OldB205FriendOrFoe asks the caster about every player, themselves first and then in turn order. Each foe with two or more untapped creatures is asked which one stays untapped; every other creature of theirs is tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b594f7f1-1db7-46e9-bbc7-42ccc16ce9a1",
		Name:         "Regna's Sanction",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return regnasSanctionAsk(ctx)
		},
	})
}

var regnasSanctionAsk func(ctx *effects.Context) error

func init() {
	regnasSanctionAsk = effects.NewS82OldB205FriendOrFoe("option-pick/regnas-sanction-friend-or-foe", "Regna's Sanction",
		func(ctx *effects.Context, friends, foes []uuid.UUID) error {
			g := ctx.Game
			for _, friend := range friends {
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == friend && c.IsCreature() {
						if err := (effects.AddCounter{Target: c.InstanceID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.AsGroupMember()); err != nil {
							return err
						}
					}
				}
			}
			for _, foe := range foes {
				var untapped []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == foe && c.IsCreature() && !c.Tapped {
						untapped = append(untapped, c.InstanceID)
					}
				}
				switch {
				case len(untapped) == 0:
				case len(untapped) == 1:
				default:
					g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
						Chooser:  foe,
						Source:   ctx.Source(),
						Question: "Regna's Sanction — choose one untapped creature you control; the rest are tapped",
						Cards:    untapped,
						Min:      1,
						Max:      1,
						ThenKey:  regnasSanctionKeep,
						LiveZone: game.ZoneBattlefield,
						Asked:    game.NewAskedItem(ctx.Item),
					})
				}
			}
			return nil
		})
}

// regnasSanctionKeep taps every creature the foe controls except the one they chose.
var regnasSanctionKeep = effects.AskedPick("choose-cards/regnas-sanction-keep", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == r.Chooser && c.IsCreature() && c.InstanceID != r.Picked[0] {
			if err := (effects.TapTarget{Target: c.InstanceID}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
	}
	return nil
})
