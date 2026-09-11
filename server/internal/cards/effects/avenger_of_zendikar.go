package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avenger of Zendikar — Creature — Elemental, {5}{G}{G}, 5/5:
//
//	"When this creature enters, create a 0/1 green Plant creature
//	 token for each land you control."
//
//	"Landfall — Whenever a land you control enters, you may put a
//	 +1/+1 counter on each Plant creature you control."
//
// The green ramp payoff: seven mana on an eight-land board makes
// eight bodies, and the next land turns them into 1/2s, and the one
// after that 2/3s. With a Scapeshift or a Splendid Reclamation it
// ends the game on the spot.
//
// # Two abilities, both ordinary
//
// The ETB counts lands at RESOLUTION, not at announce — a land that
// entered in response makes one more Plant. CreateToken with N taken
// from a live battlefield walk is exactly that.
//
// The landfall trigger fires on any land YOU CONTROL entering,
// played or put onto the battlefield by an effect, which is why
// Avenger and fetchlands and Harrow all want each other.
//
// # The Plant token
//
// Declared locally rather than in tokens.go, per the convention #231
// set: concurrent card batches collide on shared helper files, and a
// one-card token belongs with its card.
//
// Colors is set explicitly to {"G"}. That is not cosmetic — a token
// has no mana cost, so EffectiveColors() has nothing to derive a
// colour from, and a colour predicate would see a colourless Plant.
// bastion_of_remembrance.go makes the same note for the same reason.
//
// # The "may"
//
// CR 603.4: the counter clause is optional, and TriggerOptionalPrompt
// is the engine's yes/no on a trigger. It is wired rather than
// assumed-yes deliberately — putting +1/+1 counters on your board is
// nearly always right, but "nearly always" is how a card ends up
// stronger than printed, and the prompt costs nothing. The counters
// go on each Plant CREATURE you control, which is every Plant
// including ones from another source, and not on the Avenger.
//
// The +1/+1 counters route through AddCounterForEffect, so a
// Doubling Season or a Hardened Scales sees them through the CR 614
// replacement pipeline, as printed.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID: "4ba5b3f6-503b-43e6-b66e-4f8c55cffed7",
		Name:     "Avenger of Zendikar",
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Avenger of Zendikar — a Plant for each land you control",
						func(g *game.Game, item *game.StackItem) error {
							n := 0
							for _, c := range g.BattlefieldCardsForEffect() {
								if c.IsLand() && c.Controller == item.Controller {
									n++
								}
							}
							if n == 0 {
								return nil
							}
							return CreateToken{
								Controller: item.Controller,
								Template:   greenPlantToken(),
								N:          n,
							}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.IsLand() && c.Controller == source.Controller
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{
					Question: "Avenger of Zendikar — put a +1/+1 counter on each Plant you control?",
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Avenger of Zendikar — landfall, +1/+1 counter on each Plant",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							for _, c := range g.BattlefieldCardsForEffect() {
								if c.Controller != item.Controller || !c.IsCreature() {
									continue
								}
								if !hasSubtype(c, "Plant") {
									continue
								}
								if err := (AddCounter{
									Target: c.InstanceID,
									Kind:   game.CounterPlusOne,
									N:      1,
								}).Apply(ctx); err != nil {
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

// greenPlantToken is Avenger of Zendikar's 0/1 green Plant. Colors is
// explicit because a token has no mana cost to derive a colour from.
func greenPlantToken() game.Card {
	return game.Card{
		Name:      "Plant",
		TypeLine:  "Token Creature — Plant",
		Power:     0,
		Toughness: 1,
		Colors:    []string{"G"},
	}
}
