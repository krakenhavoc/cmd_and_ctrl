package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chocobo Knights — Creature — Human Knight {3}{W}, 3/3 (EDHREC rank
// 4310):
//
//	"Whenever you attack, creatures you control with counters on them
//	 gain double strike until end of turn."
//
// The +1/+1-counters deck's finisher: a board of modest creatures with
// counters on them all deal their damage twice, and the Knights
// themselves are not among them unless something put a counter on them
// first. That last part is the design — the card wants to be in a deck
// that was already making counters, not to be one.
//
// Two details carry the card:
//
//   - "Whenever YOU ATTACK" is ONE trigger per attack declaration, not
//     one per attacker. The engine emits EventAttack per creature, so
//     OncePerBatch declines every later event of the same batch;
//     without it a three-creature attack would grant double strike
//     three times, which would be harmless here but is the #259
//     direction and would not be on the next card that shares this
//     shape.
//   - "WITH COUNTERS ON THEM" is any counter of any kind, not just
//     +1/+1 — a creature carrying a stun, shield or oil counter
//     qualifies, which is the printed text and occasionally relevant.
//
// The affected set is snapshotted when the trigger resolves
// (CR 611.2c), so a creature that gets its first counter after combat
// damage is not retroactively given double strike, and a creature
// flickered in response comes back a new object (CR 400.7) without it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6cbefcba-0c1c-48fd-9b78-905ed61aae1e",
		Name:         "Chocobo Knights",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, "Chocobo Knights — creatures with counters gain double strike",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl(), effects.B41HasACounter()),
						Keywords: []string{"double strike"},
						Label:    "Chocobo Knights — double strike until end of turn",
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
