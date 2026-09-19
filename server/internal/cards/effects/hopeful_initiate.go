package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hopeful Initiate — Creature — Human Warlock {W}, 1/1 (EDHREC rank
// 6721):
//
//	"Training (Whenever this creature attacks with another creature
//	 with greater power, put a +1/+1 counter on this creature.)
//	 {2}{W}, Remove two +1/+1 counters from among creatures you
//	 control: Destroy target artifact or enchantment."
//
// The white one-drop that turns counters into removal. The activated
// ability is live since #789: "from among creatures you control" is a
// counter removal split across permanents, and the payment names a
// count per creature totalling exactly two — one each from two
// creatures, or both from one.
//
// The Initiate itself is a creature you control and a legal part of
// that payment, which is how the card is meant to be played once
// training has grown it.
//
// One declared simplification, weaker than printed: TRAINING is not
// implemented. It is a keyword ability (CR 702.138) with an attack
// trigger the engine has no shape for yet, so the Initiate never
// grows itself — the ability still works with counters from anywhere
// else (Cathars' Crusade, a Sword, another player's Ajani). Shipping
// training's counter without its trigger would be inventing a card;
// leaving it off is the #259 direction.
func init() {
	Register(Spec{
		OracleID:     "317169b8-9014-48e6-862b-ca21b706846e",
		Name:         "Hopeful Initiate",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Training isn't implemented — attacking alongside a bigger creature doesn't put a +1/+1 counter on the Initiate."},
		Activated: []ActivatedAbility{{
			Label: "{2}{W}, Remove two +1/+1 counters from among creatures you control: Destroy target artifact or enchantment.",
			Cost: Plus(
				ManaCost("{2}{W}"),
				RemoveCountersAmong(game.CounterPlusOne, 2, "creatures you control", Creature()),
			),
			Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, ref := range ctx.LegalTargets() {
					if ref.Kind == game.TargetCard {
						return DestroyTarget{Target: ref.ID}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
