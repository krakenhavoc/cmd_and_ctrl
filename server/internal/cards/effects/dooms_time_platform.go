package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Doom's Time Platform — {4} Artifact (Marvel Super Heroes Commander,
// #4455):
//
//	"Whenever you attack, exile target nonland card from your graveyard
//	 with two time counters on it. If it doesn't have suspend, it gains
//	 suspend. (At the beginning of your upkeep, remove a time counter. When
//	 the last is removed, you may cast it without paying its mana cost. If
//	 it's a creature, it has haste.)"
//
// "Whenever you attack" is one trigger per attack declaration (the engine
// emits an attack event per attacker and OncePerBatch keeps the first of
// the batch). The target is a nonland card in YOUR graveyard, chosen as the
// trigger goes on the stack and judged again as it resolves (CR 608.2b).
// The exile is a keyed continuation, so a replacement on the exile is a
// restore point and a card that never reached exile is not counted; the
// continuation puts the two time counters on the card and gives suspend to
// one that has none (game.GainSuspendForEffect, CR 702.62a).
//
// No simplification.
func init() {
	const label = "Doom's Time Platform — exile target nonland card from your graveyard with two time counters on it; it gains suspend"
	Register(Spec{
		OracleID:     "78019626-1154-483f-b862-73982f9ddc7f",
		Name:         "Doom's Time Platform",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(OncePerBatch(On(game.EventAttack,
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, label, func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard || t.ID == uuid.Nil {
							continue
						}
						return g.ExileCardKeyedForEffect(t.ID, game.KeyedTail{
							Ref:    game.GainSuspendExileTail,
							Player: ctx.Controller(),
							N:      2,
						})
					}
					return nil
				})),
				TargetCardInGraveyard("target nonland card from your graveyard", Nonland(), YouOwn())),
		},
	})
}
