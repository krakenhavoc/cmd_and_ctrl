package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winter Soldier, Reborn Avenger — Legendary Creature — Human Soldier Hero
// {4}{W}, 3/5 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever Winter Soldier attacks, return target creature card with mana
//	 value less than or equal to Winter Soldier's power from your graveyard
//	 to the battlefield. If a Hero enters this way, it enters with an
//	 additional +1/+1 counter on it."
//
// The mana-value bound is data on the clause (RelativeToSource with
// ManaValueNoGreaterThanPower), judged against Winter Soldier's effective
// power as the target is chosen and again as the ability resolves (CR 608.2b),
// and against his last-known power once he has left the battlefield
// (CR 608.2h), so a pump or shrink in response moves the bound. The counter is
// the entry's own (Heroic Return's shared body, CR 614.12): a Hero that comes
// back enters with one more +1/+1 counter; any other creature does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "df81ecf2-b49e-4be2-8757-ee2050784c9e",
		Name:         "Winter Soldier, Reborn Avenger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventAttack, effects.Msc6ThisAttacks,
					"Winter Soldier — return target creature card with mana value less than or equal to his power from your graveyard to the battlefield",
					winterSoldierReturn),
				effects.RelativeToSource(
					effects.TargetCardInGraveyard("target creature card with mana value less than or equal to Winter Soldier's power from your graveyard",
						effects.Creature(), effects.YouOwn()),
					effects.ManaValueNoGreaterThanPower())),
		},
	})
}

func winterSoldierReturn(g *game.Game, item *game.StackItem) error {
	return effects.ReturnTargetCardGivingHeroesCounters(effects.NewContext(g, item), 1)
}
