package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Suture Priest — Creature — Phyrexian Cleric {1}{W}, 1/1 (EDHREC rank
// 1182):
//
//	"Whenever another creature you control enters, you may gain 1 life.
//	 Whenever a creature an opponent controls enters, you may have that
//	 player lose 1 life."
//
// Two optional triggers, both the Priest's controller's choice — the
// second is "you may HAVE that player lose", so the yes/no goes to
// the controller, not the opponent. The first is Soul Warden's
// condition narrowed to your own creatures; the second is the mirror.
// Each is one prompt per creature, as printed. The losing player is
// captured in Build off the entering creature's controller, so a
// creature that changes hands before the trigger resolves still
// costs the player who brought it in.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4d36522-3ace-4bfb-bf2d-1a366f458698",
		Name:         "Suture Priest",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(WheneverAnotherCreatureEntersUnderYourControl("Suture Priest — you gain 1 life", Do(GainLife{Amount: 1})), "Suture Priest — gain 1 life?"),
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.IsCreature() && c.Controller != source.Controller
				},
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
					c, ok := g.LookupCardForEffect(ev.CardID)
					if !ok {
						return nil
					}
					victim := c.Controller
					return game.NewTriggeredItem(source, "Suture Priest — that player loses 1 life",
						func(g *game.Game, item *game.StackItem) error {
							if g.PlayerByIDForEffect(victim) == nil {
								return nil
							}
							return g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -1)
						})
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Suture Priest — that player loses 1 life?"},
			},
		},
	})
}
