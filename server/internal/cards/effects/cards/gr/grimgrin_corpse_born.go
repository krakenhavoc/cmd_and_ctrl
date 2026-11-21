package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grimgrin, Corpse-Born — Legendary Creature — Zombie Warrior {3}{U}{B},
// 5/5:
//
//	"Grimgrin enters tapped and doesn't untap during your untap step.
//	 Sacrifice another creature: Untap Grimgrin and put a +1/+1 counter
//	 on it.
//	 Whenever Grimgrin attacks, destroy target creature defending player
//	 controls, then put a +1/+1 counter on Grimgrin."
//
// The entering-tapped replacement and the untap-step restriction are
// Traxos's. The sacrifice is the cost, so the untap and the counter happen
// even if Grimgrin is responded to; "another" excludes Grimgrin itself. The
// attack trigger targets a creature the player Grimgrin is attacking
// controls (CR 802.2a); if the target is illegal as the trigger resolves
// nothing happens at all, the counter included (CR 608.2b), and once the
// destroy is done, or prevented by regeneration or indestructible, the
// counter goes on all the same ("then").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "a2aa2c8f-7058-4020-9881-663b104ea051",
		Name:                  "Grimgrin, Corpse-Born",
		Completeness:          effects.CompletenessFull,
		Replacements:          []game.ReplacementEffect{effects.SelfEntersTapped()},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.DoesntUntapDuringYourUntapStep()},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice another creature: Untap Grimgrin and put a +1/+1 counter on it.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.SacrificeAnotherN(1, "another creature", effects.Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Do(
					effects.UntapTarget{Target: item.SourceCardID},
					effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1},
				)(g, item)
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:     []game.EventKind{game.EventAttack},
			AppliesTo:   effects.ThisAttacked,
			TargetsFrom: effects.TargetCreatureDefendingPlayerControls,
			Key:         "Grimgrin, Corpse-Born — destroy target creature defending player controls, then put a +1/+1 counter on Grimgrin",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})(g, item)
			},
		}},
	})
}
