package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Patchwork Crawler — Creature — Zombie Horror {1}{U}, 1/2:
//
//	"{2}{U}: Exile target creature card from your graveyard and put a
//	 +1/+1 counter on this creature.
//	 This creature has all activated abilities of all creature cards
//	 exiled with it."
//
// Dark Impostor's shape over your own graveyard: the exile is linked
// (ExileTarget.Linked, ADR 0155 §8) and the borrow reads the cards linked
// to this Crawler object (ExiledWithLenders). A Crawler that leaves and
// comes back is a new object with no lenders (CR 400.7).
//
// No simplification.
func init() {
	const clause = "This creature has all activated abilities of all creature cards exiled with it."
	borrow := effects.HasActivatedAbilitiesOf(game.AbilityBorrow{
		Label:   clause,
		Lenders: effects.ExiledWithLenders(effects.ByController(effects.Creature())),
	})
	borrow.DependsOnExile = true
	effects.Register(effects.Spec{
		OracleID:     "74cbe86b-cc8c-4a5f-93c3-f4eaced50393",
		Name:         "Patchwork Crawler",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{borrow},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Exile target creature card from your graveyard and put a +1/+1 counter on this creature.",
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			Effect:  effects.ExileLinkedTargetThenCounterOnThis,
		}},
	})
}
