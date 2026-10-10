package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pashalik Mons — Legendary Creature — Goblin Warrior {2}{R}, 2/2
// (EDHREC rank 2685):
//
//	"Whenever Pashalik Mons or another Goblin you control dies,
//	 Pashalik Mons deals 1 damage to any target.
//	 {3}{R}, Sacrifice a Goblin: Create two 1/1 red Goblin creature
//	 tokens."
//
// The Goblin deck's Blood Artist and its own sacrifice outlet. The
// dies trigger is the source's own death (cardDied) or another
// Goblin the controller controlled going to a graveyard from the
// battlefield (b25AnotherGoblinYouControlDied — a Goblin as it last
// existed on the battlefield, CR 603.10a and #1679, so a changeling
// counts and so does a creature that was a Goblin only through a
// grant). It targets, so the ping is picked as the
// trigger goes on the stack and re-checked at resolution; the damage
// is attributed to Mons whether or not he is still on the
// battlefield, as CR 113.7a's last-known information allows. The
// activation is a mana-plus-sacrifice cost whose sacrifice may be
// any Goblin the controller controls, Mons himself included — and
// sacrificing Mons to his own ability fires his own trigger above
// the tokens, as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bdb94ccd-1bb5-4bb7-9539-e1b1c97c19c5",
		Name:         "Pashalik Mons",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return cardDied(ev, source) || b25AnotherGoblinYouControlDied(ev, source, g)
			},
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Key:     "Pashalik Mons — deal 1 damage to any target",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 1}.Apply(ctx)
				}
				return nil
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{R}, Sacrifice a Goblin: Create two 1/1 red Goblin creature tokens",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{3}{R}"), b25SacrificeAGoblin()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: RedGoblinToken(), N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
