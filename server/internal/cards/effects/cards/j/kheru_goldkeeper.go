package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kheru Goldkeeper — Creature — Dragon {1}{B}{G}{U}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying
//	 Whenever one or more cards leave your graveyard during your turn,
//	 create a Treasure token.
//	 Renew — {2}{B}{G}{U}, Exile this card from your graveyard: Put two
//	 +1/+1 counters and a flying counter on target creature. Activate
//	 only as a sorcery."
//
// The trigger is Attuned Hunter's (left_graveyard.go): every route a
// card leaves your graveyard by counts, cards leaving together trigger it
// once (OncePerBatch, CR 603.2c), and cards that leave on an opponent's
// turn do nothing. Renew is renew.go, Champion of Dusan's shape: the
// flying counter is a keyword counter (CR 122.1b). A target that has left
// by resolution is illegal and nothing happens (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aef4ab0e-c81f-428c-ab23-64504a9fb989",
		Name:            "Kheru Goldkeeper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.OnAny(effects.LeftGraveyardEvents, effects.WhenCardsLeaveYourGraveyardDuringYourTurn,
				"Kheru Goldkeeper — create a Treasure token",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1}))),
		},
		Activated: []effects.ActivatedAbility{
			effects.Renew("{2}{B}{G}{U}", "Put two +1/+1 counters and a flying counter on target creature.",
				effects.TargetCreature("target creature"), kheruGoldkeeperRenew),
		},
	})
}

// kheruGoldkeeperRenew puts two +1/+1 counters and a flying counter on
// the target, if it is still legal (CR 608.2b).
func kheruGoldkeeperRenew(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
		return err
	}
	return effects.AddCounter{Target: id, Kind: game.CounterFlying, N: 1}.Apply(ctx)
}
