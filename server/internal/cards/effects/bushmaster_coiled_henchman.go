package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bushmaster, Coiled Henchman — Legendary Creature — Snake Human Villain
// {1}{G}, 2/1 (Marvel Super Heroes Commander, #4455):
//
//	"Deathtouch
//	 Each other creature you control with a +1/+1 counter on it has
//	 deathtouch."
//
// A layer 6 grant (Saryth's shape) that follows the counter: the keyword
// appears when the counter lands and goes when the last one leaves.
func init() {
	Register(Spec{
		OracleID:        "21067a25-f749-4b79-bc6d-820a91af2834",
		Name:            "Bushmaster, Coiled Henchman",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID &&
					target.Controller == source.Controller &&
					target.IsCreature() && target.Counters[game.CounterPlusOne] > 0
			}, "deathtouch"),
		},
	})
}
