package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mavren Fein, Dusk Apostle — Legendary Creature — Vampire Cleric
// {2}{W}, 2/2 (EDHREC rank 3809):
//
//	"Whenever one or more nontoken Vampires you control attack,
//	 create a 1/1 white Vampire creature token with lifelink."
//
// The Vampire token engine: one token per combat in which any
// nontoken Vampire — Mavren himself included — attacks. The engine
// emits one attack event per creature, so the condition carries the
// OncePerBatch dedup: the second Vampire declared in
// the same combat sees the first's trigger pending or on the stack
// and declines. Without it the card would ship STRONGER than
// printed. The tokens it makes are Vampires but are tokens, so they
// never fire it on their own.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b94a11b-21b0-4465-a520-69608f022fb4",
		Name:         "Mavren Fein, Dusk Apostle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b36NontokenVampiresYouControlAttacked(ev, source, g)
			}, "Mavren Fein, Dusk Apostle — create a 1/1 white Vampire with lifelink", b34CreateTokens(b36WhiteVampireLifelinkToken, 1))),
		},
	})
}
