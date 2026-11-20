package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raving Dead — Creature — Zombie {4}{B}, 2/6 (Commander 2014, #5553):
//
//	"Deathtouch
//	 At the beginning of combat on your turn, choose an opponent at random.
//	 This creature attacks that player this combat if able.
//	 Whenever this creature deals combat damage to a player, that player
//	 loses half their life, rounded down."
//
// The random opponent is Knight Rampager's: drawn from the keyed random
// stream as the trigger resolves, announced in the log, and turned into a
// requirement on this object to attack that player this combat
// (AttackPlayerIfAbleMod; CR 508.1d lets it stay home when it can't). The
// life loss is half the damaged player's life as the trigger resolves,
// rounded down, so a player at 1 loses nothing; it is loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "30e5d70e-6547-483a-a4ab-4c2eed296543",
		Name:            "Raving Dead",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Raving Dead — choose an opponent at random; this creature attacks that player this combat if able",
				ravingDeadChooseVictim),
			WheneverThisDealsCombatDamageToAPlayer("Raving Dead — that player loses half their life, rounded down",
				func(g *game.Game, item *game.StackItem) error {
					victim := g.PlayerByIDForEffect(NewContext(g, item).Trigger().Event.Target)
					if victim == nil || victim.Life < 2 {
						return nil
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, victim.ID, -(victim.Life / 2))
				}),
		},
	})
}

// ravingDeadChooseVictim picks an opponent at random, logs the pick and makes
// this creature attack that player this combat if able.
func ravingDeadChooseVictim(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	pick := g.ChooseAtRandomForEffect(randomDraw(ctx), ctx.Opponents(), 1)
	if len(pick) != 1 {
		return nil
	}
	opp := pick[0]
	name := "an opponent"
	if p := g.PlayerByIDForEffect(opp); p != nil && p.Name != "" {
		name = p.Name
	}
	g.EmitEvent(game.Event{
		Kind:   game.EventOptionChosen,
		Actor:  item.Controller,
		CardID: item.SourceCardID,
		Label:  "Raving Dead: " + name + " (at random)",
	})
	if sourceIsNewObject(g, item) {
		return nil
	}
	d, ok := DurationUntilEndOfCombat(ctx)
	if !ok {
		return nil
	}
	return ScopedEffectFor{
		Target:   item.SourceCardID,
		Mods:     []game.Mod{game.AttackPlayerIfAbleMod(opp)},
		Duration: d,
		Label:    "Raving Dead — attacks that player this combat if able",
	}.Apply(ctx)
}
