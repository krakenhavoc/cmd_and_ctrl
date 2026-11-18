package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reaper's Talisman — Artifact — Equipment {B} (Adventures in the
// Forgotten Realms, #5499):
//
//	"Whenever equipped creature attacks, it gains deathtouch until end of
//	 turn.
//	 Whenever equipped creature attacks alone, defending player loses 2
//	 life and you gain 2 life.
//	 Equip {2}"
//
// Two printed triggers, two registered ones, both on the Equipment.
// Baseball Bat's "equipped creature attacked" gates the deathtouch; Bilbo's
// Ring's attacks-alone count (the whole declaration is stamped before any
// attack event fires, so a lone attacker is alone rather than "alone so
// far") gates Brutal Hordechief's drain of 2. The defending player is read
// off the attack as it is declared; a defender who has left loses nothing
// and you still gain 2. The deathtouch lands on the creature the event
// names, if it is still there (CR 400.7).
//
// No simplification.
func init() {
	const name = "Reaper's Talisman"
	Register(Spec{
		OracleID:     "bc2ea8c8-5360-4256-91c5-d821b5e59c0f",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack,
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attachedCreatureAttacked(ev, source)
				},
				name+" — equipped creature gains deathtouch until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					id := item.Trigger.Event.CardID
					if !onBattlefield(g, id) {
						return nil
					}
					return GrantKeywordUntilEOT{
						Target:   id,
						Keywords: []string{"deathtouch"},
						Label:    name + " — deathtouch",
					}.Apply(NewContext(g, item))
				}),
			attackDrainNTriggerWhen(name, 2, attachedCreatureAttackedAlone),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
