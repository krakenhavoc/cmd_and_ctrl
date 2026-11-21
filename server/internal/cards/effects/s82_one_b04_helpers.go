package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b04_helpers.go — shared pieces of Phyrexia: All Will Be One's
// Mite and poison-counter cards (S82 batch ONE-04).

// s82OneB04EachOpponentGetsPoison is "each opponent gets a poison counter",
// the controller being the one who puts it there (CR 122.1f).
func s82OneB04EachOpponentGetsPoison(ctx *Context) error {
	for _, opp := range ctx.Opponents() {
		if err := ctx.Game.AddPlayerCounterByForEffect(ctx.Controller(), opp, game.CounterPoison, 1); err != nil {
			return err
		}
	}
	return nil
}

// s82OneB04PoisonStep is the same clause as one step of a spell's body.
func S82OneB04PoisonStep() Instruction {
	return func(ctx *Context) error { return s82OneB04EachOpponentGetsPoison(ctx) }
}

// s82OneB04ToxicDiesMite is the whole spec of a "Toxic 1. When this creature
// dies, create a 1/1 colorless Phyrexian Mite artifact creature token with
// toxic 1 and 'This token can't block.'" creature (Crawling Chorus,
// Stinging Hivemaster).
func S82OneB04ToxicDiesMite(oracleID, name string) Spec {
	return Spec{
		OracleID:        oracleID,
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"toxic 1"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisDies(name+" — create a Phyrexian Mite", Do(CreateToken{Template: PhyrexianMiteToken(), N: 1})),
		},
	}
}

// s82OneB04OpponentPoison is the poison counters your opponents have, summed,
// read live on every layer pass (Vishgraz).
func S82OneB04OpponentPoison(g *game.Game, controller uuid.UUID) int {
	n := 0
	EachOpponent(g, controller, func(opp uuid.UUID) bool {
		if p := g.PlayerByIDForEffect(opp); p != nil {
			n += p.Counters[game.CounterPoison]
		}
		return false
	})
	return n
}

// s82OneB04Edicted is the poison counter an edict hands out once every
// opponent has sacrificed (Vraska's Fall), keyed.
var S82OneB04Edicted = RegisterEdictThen("prompt-run/s82-one-b04-vraskas-fall-poison", func(ctx *Context, _ game.PromptedSacrifices) error {
	return s82OneB04EachOpponentGetsPoison(ctx)
})

// s82OneB04ReturnAnotherPermanent is "you may return another permanent you
// control to its owner's hand" as a resolution-time pick (no target): the
// source itself is never offered, and zero is a legal answer. Ambrosia
// Whiteheart and Indoctrination Attendant share it; `keyed` is what runs on
// the pick.
func S82OneB04ReturnAnotherPermanent(g *game.Game, item *game.StackItem, question string, keyed PermanentsThen) error {
	self := item.SourceCardID
	return ChoosePermanents{
		Question: question,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, id := range PermanentsControlledBy(g, of) {
				if id != self {
					out = append(out, id)
				}
			}
			return out, 0, 1
		},
		Keyed: keyed,
	}.Apply(NewContext(g, item))
}
