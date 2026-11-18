package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:              "a2aa2c8f-7058-4020-9881-663b104ea051",
		Name:                  "Grimgrin, Corpse-Born",
		Completeness:          CompletenessFull,
		Replacements:          []game.ReplacementEffect{SelfEntersTapped()},
		UntapStepRestrictions: []game.UntapStepRestriction{doesntUntapDuringYourUntapStep()},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice another creature: Untap Grimgrin and put a +1/+1 counter on it.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    SacrificeAnotherN(1, "another creature", Creature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Do(
					UntapTarget{Target: item.SourceCardID},
					CounterOnThis{Kind: game.CounterPlusOne, N: 1},
				)(g, item)
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:     []game.EventKind{game.EventAttack},
			AppliesTo:   ThisAttacked,
			TargetsFrom: TargetCreatureDefendingPlayerControls,
			Key:         "Grimgrin, Corpse-Born — destroy target creature defending player controls, then put a +1/+1 counter on Grimgrin",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})(g, item)
			},
		}},
	})
}
