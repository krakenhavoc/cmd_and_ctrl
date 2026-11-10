package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lasyd Prowler — Creature — Snake Ranger {2}{G}{G}, 5/5 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, you may mill cards equal to the number of
//	 lands you control.
//	 Renew — {1}{G}, Exile this card from your graveyard: Put X +1/+1
//	 counters on target creature, where X is the number of land cards in
//	 your graveyard. Activate only as a sorcery."
//
// The enters trigger is Deathcap Marionette's optional mill, asked as it
// resolves, with the lands counted then (CR 608.2h); declining mills
// nothing. Renew is renew.go, and X is counted as the ability resolves,
// after its cost has exiled the Prowler, so a target that has left by
// then is illegal and nothing happens (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f780d6f6-540b-4773-8a48-e56e95c2d39e",
		Name:         "Lasyd Prowler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEnters("Lasyd Prowler — mill cards equal to the number of lands you control",
				func(g *game.Game, item *game.StackItem) error {
					n := len(landsControlledByPlayer(g, item.Controller))
					return MillCards{Player: item.Controller, N: n}.Apply(NewContext(g, item))
				}),
				"Lasyd Prowler — mill cards equal to the number of lands you control?"),
		},
		Activated: []ActivatedAbility{
			Renew("{1}{G}", "Put X +1/+1 counters on target creature, where X is the number of land cards in your graveyard.",
				TargetCreature("target creature"), lasydProwlerRenew),
		},
	})
}

// lasydProwlerRenew puts one +1/+1 counter per land card in its
// controller's graveyard on the target, if it is still legal.
func lasydProwlerRenew(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	n := b17LandCardsInGraveyard(g, item.Controller)
	if n <= 0 {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
}
