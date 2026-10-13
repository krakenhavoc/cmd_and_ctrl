package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	borrow := HasActivatedAbilitiesOf(game.AbilityBorrow{
		Label:   clause,
		Lenders: ExiledWithLenders(ByController(Creature())),
	})
	borrow.DependsOnExile = true
	Register(Spec{
		OracleID:     "74cbe86b-cc8c-4a5f-93c3-f4eaced50393",
		Name:         "Patchwork Crawler",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{borrow},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Exile target creature card from your graveyard and put a +1/+1 counter on this creature.",
			Cost:    ManaCost("{2}{U}"),
			Targets: TargetCardInGraveyard("target creature card from your graveyard", Creature(), YouOwn()),
			Effect:  exileLinkedTargetThenCounterOnThis,
		}},
	})
}
