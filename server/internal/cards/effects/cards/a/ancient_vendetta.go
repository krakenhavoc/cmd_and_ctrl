package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ancient Vendetta — Sorcery {3}{B} (Aetherdrift, #4464):
//
//	"Choose a card name. Search target opponent's graveyard, hand, and
//	 library for up to four cards with that name and exile them. Then
//	 that player shuffles."
//
// Two steps, each waiting on the one before. The name is chosen as the
// spell resolves, with Demonic Consultation's card-name prompt; its answer
// is the keyed tail's Word, and the opponent rides the tail's Carry. Then
// the extraction family's one prompt over the three zones
// (AlsoGraveyardAndHand), capped at four: "up to" lets the caster take
// fewer, none included, and the opponent's library is searched and
// shuffled whatever was taken. A target opponent who is no longer a legal
// target (they left the game) makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f053f0d-c9c7-4fcd-b218-a29d1d080104",
		Name:         "Ancient Vendetta",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim := effects.FirstLegalTargetPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
				"Ancient Vendetta — choose a card name",
				effects.WithCarry(ctx.ItemTail(ancientVendettaNamed), victim))
			return nil
		},
	})
}

// ancientVendettaNamed is the search once the name (the tail's Word) is
// known. An on-disk key.
var ancientVendettaNamed = effects.CardTail("then/ancient-vendetta-named", func(ctx *effects.Context, o game.TailOutcome) error {
	victim, name := effects.TailCarry0(o), o.Tail.Word
	if victim == uuid.Nil {
		return nil
	}
	return effects.SearchLibrary{
		Player:       ctx.Controller(),
		LibraryOwner: victim,
		Predicate: func(c game.Card) bool {
			return game.CardNameMatches(c, name)
		},
		Dest:                 game.ZoneExile,
		Limit:                4,
		AlsoGraveyardAndHand: true,
		Shuffle:              true,
		Reason:               "Ancient Vendetta — search their graveyard, hand and library for up to four cards named " + name,
	}.Apply(ctx)
})
