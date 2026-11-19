package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gluttonous Hellkite — Creature — Dragon {X}{X}{B}{R}{G}, 3/3:
//
//	"When you cast this spell, each player sacrifices X creatures. This
//	 creature enters with two +1/+1 counters on it for each creature
//	 sacrificed this way.
//	 Flying, trample"
//
// The cast trigger (Abundant Maw's WhenYouCastThisSpell) reads the X the
// Hellkite was cast with off its stack item, and every player in turn order
// sacrifices that many creatures together (Planetary Annihilation's
// simultaneous run: a player with fewer sacrifices what they have). The
// creatures sacrificed are counted in the run's continuation and written to
// the spell's stack item as the number it records (CastValues), where the
// Hellkite's own entry replacement (CR 614.1c, 614.12) reads it as it
// enters: two +1/+1 counters for each. A Hellkite that is countered, or
// that no longer is on the stack when the sacrifices finish, enters with
// none.
func init() {
	const name = "Gluttonous Hellkite"
	cast := WhenYouCastThisSpell(name+" — each player sacrifices X creatures; this creature enters with two +1/+1 counters on it for each creature sacrificed this way",
		gluttonousHellkiteCast)
	Register(Spec{
		OracleID:        "3093796d-c429-4c79-a33a-31226e00b12c",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered:       []game.TriggeredAbility{cast},
		Replacements: []game.ReplacementEffect{{
			Label:     name + ": enters with two +1/+1 counters for each creature sacrificed this way",
			EntryOnly: true,
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
				// The resolving spell's record: the stack has already let go
				// of it, and the entry happens inside the resolution.
				_, item, ok := g.ResolvingSpellForEffect(ev.CardID)
				if !ok {
					return nil
				}
				if n, ok := item.CastValue(gluttonousHellkiteKey); ok && n > 0 {
					ev.AddCounterAtETB(game.CounterPlusOne, 2*n)
				}
				return nil
			},
		}},
	})
}

// gluttonousHellkiteKey is the name the sacrifices are recorded under on
// the spell's stack item.
const gluttonousHellkiteKey = "sacrificed"

func gluttonousHellkiteCast(g *game.Game, item *game.StackItem) error {
	spell := g.StackItemForEffect(item.SourceCardID)
	if spell == nil || spell.XValue <= 0 {
		return nil
	}
	var asks []uuid.UUID
	for _, seat := range apnapPlayers(g) {
		for i := 0; i < spell.XValue; i++ {
			asks = append(asks, seat)
		}
	}
	return g.PlayersSacrificeThenKeyedForEffect(item.SourceCardID, asks,
		sacrificeSpec("a creature", Creature()),
		"Gluttonous Hellkite — each player sacrifices X creatures",
		gluttonousHellkiteSacrificed, game.RunData{Carry: []uuid.UUID{item.SourceCardID}})
}

// gluttonousHellkiteSacrificed records how many creatures were sacrificed
// on the Hellkite's stack item. An on-disk key.
var gluttonousHellkiteSacrificed = game.RegisterPromptRunThen("prompt-run/gluttonous-hellkite-sacrificed", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	spell := g.StackItemForEffect(r.Carry[0])
	if spell == nil {
		return nil
	}
	if spell.CastValues == nil {
		spell.CastValues = map[string]int{}
	}
	spell.CastValues[gluttonousHellkiteKey] = len(r.Sacrifices().Cards())
	return nil
})
