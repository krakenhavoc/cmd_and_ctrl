package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Puppet Master, String Puller — Legendary Creature — Human Artificer
// Villain, {2}{R}, 2/4:
//
//	"Whenever you attack, goad target creature an opponent controls.
//	It can't block this turn. (Until your next turn, that creature
//	attacks each combat if able and attacks a player other than you
//	if able.)
//	Whenever one or more goaded creatures deal combat damage to one
//	of your opponents, create a Treasure token."
//
// #1599: "whenever you attack" is ONE trigger per attack declaration,
// not one per attacker (the engine emits EventAttack per creature) —
// OncePerBatch declines every later event of the same batch, the same
// dedup Cool but Rude and Gimli's Reckless Might use for the identical
// printed phrase. The goad is GoadTarget (goad.go); "it can't block
// this turn" is a plain RestrictUntilEOT(CantBlock) on the same
// target, applied after the goad succeeds.
//
// The Treasure trigger is "one or more … to one of your opponents":
// creatureDealtCombatDamageToAnOpponentOf already reads that shape
// (opponent_politics.go), so the only addition is the Goaded check
// on the dealer — ANY creature's goad counts, not only one this card
// made, matching the printed line's "goaded creatures" with no
// "you've goaded". OncePerBatch keeps a multi-creature connection to
// one confused-onto-a-single occurrence.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "376c6257-4a1a-42d4-8931-72068f38727f",
		Name:         "Puppet Master, String Puller",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				},
				Targets: TargetCreature("target creature an opponent controls", OpponentControls()),
				Key:     "Puppet Master, String Puller — goad target creature; it can't block this turn",
				Effect:  puppetMasterGoadAndPreventBlock,
			}),
			OncePerBatch(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return puppetMasterGoadedCreatureHitAnOpponent(ev, source, g)
				},
				Key:    "Puppet Master, String Puller — create a Treasure",
				Effect: Do(CreateToken{Template: TreasureToken(), N: 1}),
			}),
		},
	})
}

func puppetMasterGoadAndPreventBlock(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := goadAndScheduleClear(ctx, []uuid.UUID{id}); err != nil {
		return err
	}
	return RestrictUntilEOT{
		Target:       id,
		Restrictions: game.CantBlock,
		Label:        "Puppet Master, String Puller — can't block this turn",
	}.Apply(ctx)
}

func puppetMasterGoadedCreatureHitAnOpponent(ev game.Event, source *game.Card, g *game.Game) bool {
	if !creatureDealtCombatDamageToAnOpponentOf(ev, source.Controller, g) {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && dealer.Goaded()
}
