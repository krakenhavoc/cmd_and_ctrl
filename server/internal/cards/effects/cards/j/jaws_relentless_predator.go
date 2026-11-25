package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jaws, Relentless Predator — Legendary Creature — Shark {3}{R}{R}, 5/5
// (Secret Lair Drop):
//
//	"Trample, haste
//	 Whenever Jaws deals combat damage to a player, create that many Blood
//	 tokens.
//	 Whenever a noncreature artifact is sacrificed or destroyed, Jaws deals 1
//	 damage to each opponent."
//
// Tana's combat-damage shape with the Blood token. The second ability is one
// trigger on the battlefield exit (EventZoneMove, whose LeaveCause says
// whether it was a destroy or a sacrifice, CR 701.8, CR 701.21a): a
// permanent that left as a noncreature artifact (last-known types, CR 603.10a,
// so a Treasure token that has ceased to exist still counts), whoever
// controlled it. A state-based death (CR 704.5g) is not a "destroy" effect.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1202725c-170b-42d7-9f82-a431a987f30e",
		Name:            "Jaws, Relentless Predator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Jaws, Relentless Predator — create that many Blood tokens",
				effects.Do(effects.CreateTokensForEach{
					Template: effects.BloodToken(),
					Name:     "the combat damage dealt",
					Count:    func(ctx *effects.Context) int { return ctx.Trigger().Event.Amount },
				})),
			{
				Watches:   []game.EventKind{game.EventZoneMove},
				AppliesTo: oldB10NoncreatureArtifactSacrificedOrDestroyed,
				Key:       "Jaws, Relentless Predator — 1 damage to each opponent",
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					// One printed instruction: one instance of damage (CR 615.8).
					return g.DamageInstanceForEffect(func() error {
						for _, p := range g.Seats {
							if p == nil || p.Eliminated || p.ID == item.Controller {
								continue
							}
							if err := (effects.DealDamage{Source: item.SourceCardID, Target: p.ID, Amount: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
				},
			},
		},
	})
}

// oldB10NoncreatureArtifactSacrificedOrDestroyed is Jaws' trigger condition.
func oldB10NoncreatureArtifactSacrificedOrDestroyed(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.OldZone != game.ZoneBattlefield || ev.CardID == uuid.Nil {
		return false
	}
	if ev.LeaveCause != game.LeaveCauseDestroy && ev.LeaveCause != game.LeaveCauseSacrifice {
		return false
	}
	var c game.Card
	if found, ok := g.LookupCardForEffect(ev.CardID); ok {
		c = found
	}
	return effects.LeftAsType(ev, c, "artifact") && !effects.LeftAsType(ev, c, "creature")
}
