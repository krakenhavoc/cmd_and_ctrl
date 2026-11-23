package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mazzy, Truesword Paladin — Legendary Creature — Halfling Knight {1}{R}{G}{W},
// 3/4 (Commander Legends: Battle for Baldur's Gate):
//
//	"Whenever an enchanted creature attacks one of your opponents, it gets +2/+0
//	 and gains trample until end of turn.
//	 Whenever an Aura you control is put into your graveyard from the battlefield,
//	 exile it. Until the end of your next turn, you may cast that card."
//
// The first trigger fires per attacker, whoever controls it: a creature with an
// Aura attached as it is declared (an Equipment does not enchant) that attacks a
// player (CR 508.3a, not that player's planeswalker) who is an opponent of
// Mazzy's controller. The pump and trample are pinned to that attacker (CR 611.2c,
// CR 400.7). The second reads the Aura as it last existed on the battlefield
// (CR 603.10a): an Aura its controller controlled that went to its controller's
// graveyard (an Aura you control owned by an opponent goes to theirs and is not
// "your graveyard"). The exile is Practiced Scrollsmith's: cast-only permission
// until the end of its controller's next turn, taken only if the card is still in
// the graveyard as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9fa75b3c-a512-4318-9c03-e2d958acb6ef",
		Name:         "Mazzy, Truesword Paladin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventAttack || ev.CardID == uuid.Nil {
					return false
				}
				p := effects.AttackedPlayer(g, ev)
				if p == uuid.Nil || p == source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && effects.GapB218IsEnchanted(g, c)
			}, "Mazzy, Truesword Paladin — the enchanted attacker gets +2/+0 and gains trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if item.Trigger == nil {
						return nil
					}
					attacker := item.Trigger.Event.CardID
					if err := (effects.BoostUntilEOT{Target: attacker, Power: 2, Label: "Mazzy, Truesword Paladin — +2/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: attacker, Keywords: []string{"trample"}, Label: "Mazzy, Truesword Paladin — trample"}.Apply(ctx)
				}),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				if !ok || !effects.LeftAsSubtype(ev, c, "Aura") {
					return false
				}
				return effects.LeftUnderControlOf(ev, c) == source.Controller && c.Owner == source.Controller
			}, "Mazzy, Truesword Paladin — exile that Aura; until the end of your next turn, you may cast that card",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					id := item.Trigger.Event.CardID
					if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					return g.ExileCardWithPermissionForEffect(id, game.CastPermission{
						Player:     item.Controller,
						Duration:   g.UntilEndOfYourNextTurnDuration(item.Controller),
						CastOnly:   true,
						Source:     item.SourceCardID,
						SourceName: "Mazzy, Truesword Paladin",
						Label:      "Mazzy, Truesword Paladin — cast this card",
					})
				}),
		},
	})
}
