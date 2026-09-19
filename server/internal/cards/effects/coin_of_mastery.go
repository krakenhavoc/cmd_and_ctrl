package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coin of Mastery — Artifact {4}:
//
//	"Each creature you control enters with an additional +1/+1 counter
//	 on it for each mana from an artifact source spent to cast it.
//	 {T}: Create a Treasure token."
//
// The Treasure ability ships as an ordinary CR 602 activation. The
// headline does not, and it is blocked twice over rather than once.
//
// WHAT IS MISSING. A CR 614.1c "enters with N counters" clause reads
// the announcement through game.CastCounts, and that struct is short
// on purpose: the announced X, the number of times the spell was
// kicked, and the number of distinct colours of mana spent. It
// carries no record of what PRODUCED each mana. game.ManaToken keeps
// the source permanent's instance ID, and game.PaidCost keeps the
// tokens that paid, but the source's TYPES are not snapshotted, so a
// Treasure or a Lotus Petal that was sacrificed to pay is a source ID
// that no longer resolves to anything — and a Treasure is the most
// common artifact mana at the table. That is the open "mana-spent-to-
// cast record" seam (#761 shipped the record; the source-type
// snapshot for Kalain-style riders is the part still open).
//
// The second blocker is the shape, not the data. This clause is a
// rider Coin of Mastery hangs on OTHER creature spells its controller
// casts — the counters have to be added to somebody else's entry,
// from a permanent that is not that spell's source. The catalog can
// only declare the clause on the card that prints it
// (Spec.EntersWithCountersFromCast). That is the same granted-rider
// gap Lux Artillery waits on.
//
// Weaker than printed, never stronger: the ability that IS here
// creates one Treasure and nothing else happens.
func init() {
	Register(Spec{
		OracleID:     "d78518ee-df79-48d1-b9d5-4f968b441899",
		Name:         "Coin of Mastery",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Your creatures do not get the extra +1/+1 counters for mana spent from artifacts — only the Treasure ability works.",
		},
		Activated: []ActivatedAbility{{
			Label: "{T}: Create a Treasure token.",
			Cost:  TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
