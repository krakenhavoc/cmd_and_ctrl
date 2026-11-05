package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Julian Bashir, Frontier Doctor — Legendary Creature — Human Doctor
// {2}{W}, 1/1 (Star Trek Commander, #4974):
//
//	"Seek out — Whenever you draw your second card each turn, put a +1/+1
//	 counter on Julian Bashir.
//	 Whenever Julian Bashir attacks, return target creature card with mana
//	 value X or less from your graveyard to the battlefield, where X is
//	 Julian Bashir's power."
//
// Seek out is an ability word: the trigger is Atlantean Cavalry's (the
// turn's draw tally reads exactly two on the second draw, so it fires
// once a turn). The attack trigger is Winter Soldier's: the mana-value
// bound is data on the clause (RelativeToSource with
// ManaValueNoGreaterThanPower), judged against Bashir's power through the
// layers as the target is chosen and again as the ability resolves (CR
// 608.2b), and against his last-known power once he has left (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bbc8c2b4-9949-416f-a722-b1645eff64d7",
		Name:         "Julian Bashir, Frontier Doctor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Julian Bashir — seek out: put a +1/+1 counter on Julian Bashir",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			Targeting(
				On(game.EventAttack, ThisAttacked,
					"Julian Bashir — return target creature card with mana value X or less from your graveyard to the battlefield, where X is his power",
					returnFirstLegalGraveyardTargetToBattlefield),
				RelativeToSource(
					TargetCardInGraveyard("target creature card with mana value X or less from your graveyard, where X is Julian Bashir's power",
						Creature(), YouOwn()),
					ManaValueNoGreaterThanPower())),
		},
	})
}
