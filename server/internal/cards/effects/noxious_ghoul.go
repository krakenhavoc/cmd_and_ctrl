package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Noxious Ghoul — Creature — Zombie {3}{B}{B}, 3/3 (EDHREC rank
// 4479):
//
//	"Whenever this creature or another Zombie enters, all non-Zombie
//	 creatures get -1/-1 until end of turn."
//
// The Zombie deck's board wipe on a stick. One Ghoul plus a
// Gravecrawler loop, or a Zombie token maker, is a repeatable
// one-sided shrink; two Ghouls are -2/-2 per Zombie, and a mass
// Zombie entry — Army of the Damned, Endless Ranks of the Dead —
// clears the table.
//
// Three things worth spelling out:
//
//   - "Another ZOMBIE", with no "you control": an OPPONENT'S Zombie
//     entering triggers the Ghoul too. That is printed, and in a
//     Zombie mirror it matters. The trigger asks only for the
//     subtype, so a noncreature permanent that is somehow a Zombie
//     counts, exactly as the oracle text reads.
//   - The shrink hits ALL non-Zombie creatures, yours included. The
//     Ghoul is a Zombie and is spared; your non-Zombie utility
//     creatures are not.
//   - "Until end of turn" is a CR 611.2 turn-scoped continuous
//     effect, so the affected set is SNAPSHOTTED as the trigger
//     resolves (CR 611.2c): a non-Zombie that enters afterwards is
//     not shrunk, and a creature that stops being a Zombie later in
//     the turn is not retroactively shrunk either. Both are the
//     printed behaviour.
//
// Stacking is additive: three separate triggers are three separate
// -1/-1 effects, so a 3/3 dies to the third.
//
// No simplification.
func init() {
	nonZombieCreature := func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && !c.HasSubtype("Zombie")
	}
	Register(Spec{
		OracleID:     "602e421e-afab-4488-8a52-af66b4fdce4b",
		Name:         "Noxious Ghoul",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.CardID == uuid.Nil {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.HasSubtype("Zombie")
			}, "Noxious Ghoul — all non-Zombie creatures get -1/-1 until end of turn", func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{
					Match:     nonZombieCreature,
					Power:     -1,
					Toughness: -1,
					Label:     "Noxious Ghoul — non-Zombies get -1/-1",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
