package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mangara, the Diplomat — Legendary Creature — Human Cleric {3}{W},
// 2/4:
//
//	"Lifelink
//	 Whenever an opponent attacks with creatures, if two or more of
//	 those creatures are attacking you and/or planeswalkers you
//	 control, draw a card.
//	 Whenever an opponent casts their second spell each turn, draw a
//	 card."
//
// Trouble in Pairs' attack and cast clauses, through the same two
// shared conditions (opponentAttackedYouWithAtLeast and
// opponentCastTheirSecondSpellThisTurn) and the same once-per-batch
// guard on the attack half: the engine emits one attack event per
// declared attacker, so without it a five-creature alpha strike would
// draw four cards. An attack at your planeswalker counts as an attack
// at you.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cbcb6d9a-6ae5-4bcd-8013-2b657553764a",
		Name:            "Mangara, the Diplomat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, mangaraOpponentAttackedYouWithTwo, mangaraAttackLabel, Do(DrawCards{N: 1}))),
			On(game.EventCast, mangaraOpponentCastTheirSecond, mangaraCastLabel, Do(DrawCards{N: 1})),
		},
	})
}

const (
	mangaraAttackLabel = "Mangara, the Diplomat — an opponent attacked you with two or more creatures: draw a card"
	mangaraCastLabel   = "Mangara, the Diplomat — an opponent cast their second spell: draw a card"
)

func mangaraOpponentAttackedYouWithTwo(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return opponentAttackedYouWithAtLeast(ev, source, g, 2, mangaraAttackLabel)
}

func mangaraOpponentCastTheirSecond(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return opponentCastTheirSecondSpellThisTurn(ev, source, g)
}
