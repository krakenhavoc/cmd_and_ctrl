package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Distinguished Conjurer — Creature — Human Wizard {1}{W}, 1/2
// (EDHREC rank 4498):
//
//	"Whenever another creature you control enters, you gain 1 life.
//	 {4}{W}, {T}: Exile another target creature you control, then
//	 return it to the battlefield under its owner's control."
//
// A two-mana blink OUTLET. Five mana an activation is a bad rate on
// its own; what a blink deck is buying is a repeatable effect on a
// cheap body that dodges sorcery-speed removal and can be untapped —
// and the lifegain half turns every ETB the deck was already making
// into a trigger for a Well of Lost Dreams or an Archangel of
// Thune.
//
// The blink is Flicker: exile and return in one resolution, so the
// creature is a NEW object — counters and damage fall off, Auras and
// Equipment fall off, and every enters trigger fires again, the
// Conjurer's own lifegain included. "Under its OWNER'S control" is
// the printed clause and it is the reason this is not a Cloudshift:
// blinking a creature you stole with a Control Magic hands it back.
//
// "Another" is excluded twice, the way Dour Port-Mage does it: by
// name at announce (a singleton format, so the same name is the same
// creature) and by instance at resolution, since a target clause
// cannot see its own source.
//
// The lifegain trigger is "another CREATURE YOU CONTROL enters" —
// tokens count, an opponent's creature does not, and the Conjurer's
// own entry does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6df57d67-2fd9-4e7a-b67b-f361fc30e496",
		Name:         "Distinguished Conjurer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureEnteredUnderYourControl,
				"Distinguished Conjurer — you gain 1 life",
				func(g *game.Game, item *game.StackItem) error {
					return GainLife{Player: item.Controller, Amount: 1}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{W}, {T}: Exile another target creature you control, then return it to the battlefield under its owner's control.",
			Cost:    Plus(ManaCost("{4}{W}"), TapCost()),
			Targets: TargetCreature("another target creature you control", YouControl(), b03NotNamed("Distinguished Conjurer")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
					return nil
				}
				if item.Targets[0].ID == item.SourceCardID {
					return nil
				}
				c, ok := g.LookupCardForEffect(item.Targets[0].ID)
				if !ok {
					return nil
				}
				return Flicker{Target: c.InstanceID, Controller: c.Owner}.Apply(NewContext(g, item))
			},
		}},
	})
}
