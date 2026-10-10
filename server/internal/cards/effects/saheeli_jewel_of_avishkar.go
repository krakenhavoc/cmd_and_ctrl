package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saheeli, Jewel of Avishkar — Legendary Creature — Human Artificer
// {2}{U}{R}, 2/4:
//
//	"Thopters you control have haste.
//	 Whenever you cast a noncreature spell, create a 1/1 colorless
//	 Thopter artifact creature token with flying."
//
// The haste is a Layer 6 grant to the controller's Thopters (effective
// subtypes, so a changeling counts). The trigger reads the spell off the
// stack, where its type line is intact.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83705258-1f42-41db-a32e-de99ffd759eb",
		Name:         "Saheeli, Jewel of Avishkar",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && target.HasSubtype("Thopter")
			}, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, b10NoncreatureSpellCastByYou,
				"Saheeli, Jewel of Avishkar — create a 1/1 Thopter with flying",
				Do(CreateToken{Template: TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
	})
}
