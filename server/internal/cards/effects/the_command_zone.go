package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Command Zone — Plane — MagicCon (Unknown Event, #4870):
//
//	"When you planeswalk here, each player who controls their commander
//	 draws a card. Other players may put their commander from the command
//	 zone onto the battlefield.
//	 If an ability of your commander would trigger during your turn, that
//	 ability triggers an additional time.
//	 Whenever chaos ensues, the number of times each commander has been cast
//	 from the command zone becomes 0. (This resets commander tax.)"
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6), who
// is the active player, so "during your turn" is every turn the plane is
// face up. A player "controls their commander" when a commander they own is
// on the battlefield under their control; the others are offered, in turn
// order, to put each commander they have in the command zone onto the
// battlefield (a yes/no each). The doubler is a CR 603.2d trigger doubler
// declared on a face-up plane (the trigger scan reads the plane) that doubles
// every triggered ability of a permanent that is a commander the planar
// controller owns. The chaos ability empties every commander's tax tally.
//
// No simplification.
func init() {
	doubler := DoublesAbilitiesOf(nil, DoublesAbilitiesOfOptions{
		SkipControllerCheck: true,
		SourceMatch: func(g *game.Game, q game.TriggerDoublingQuery) bool {
			src := q.Source
			you := q.DoublerLKI.Controller
			if !src.IsCommander || src.Owner != you || g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
				return false
			}
			return g.Seats[g.Turn.ActiveSeat].ID == you
		},
	})
	doubler.Label = "The Command Zone"
	doubler.RowLabel = "If an ability of your commander would trigger during your turn, that ability triggers an additional time."
	Register(Spec{
		OracleID:        "eaa5979c-86d3-4647-9a45-fd80ba4bf192",
		Name:            "The Command Zone",
		Completeness:    CompletenessFull,
		TriggerDoublers: []game.TriggerDoubler{doubler},
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThis("The Command Zone — each player who controls their commander draws a card; other players may put their commander onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, id := range b13PlayersFrom(g, item.Controller) {
						p := g.PlayerByIDForEffect(id)
						if b18ControlsYourCommander(g, id) {
							if err := (DrawCards{Player: id, N: 1}).Apply(ctx); err != nil {
								return err
							}
							continue
						}
						for _, c := range p.Command.Cards {
							if !c.IsCommander {
								continue
							}
							if err := (MayChoice{
								Player:   id,
								Question: "The Command Zone — put " + c.Name + " from the command zone onto the battlefield?",
								YesLabel: "Put it onto the battlefield",
								NoLabel:  "Leave it",
								Then:     theCommandZonePutCommander,
								Carry:    []uuid.UUID{c.InstanceID},
							}).Apply(ctx); err != nil {
								return err
							}
						}
					}
					return nil
				}),
			WheneverChaosEnsues("The Command Zone — the number of times each commander has been cast from the command zone becomes 0",
				func(g *game.Game, _ *game.StackItem) error {
					g.ResetCommanderCastsForEffect()
					return nil
				}),
		},
	})
}

// theCommandZonePutCommander is a player putting their commander onto the
// battlefield. An on-disk key: never renamed, never reused.
var theCommandZonePutCommander = MayYes("confirm/the-command-zone-put-commander", func(ctx *Context, r game.ConfirmAnswer) error {
	if len(r.Carry) != 1 {
		return nil
	}
	_, err := ctx.Game.PutFromCommandZoneOntoBattlefieldForEffect(r.Carry[0], game.ZoneEntryOptions{Controller: r.Chooser})
	return err
})
