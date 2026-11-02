package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Slayer, Hatred Honed — Legendary Artifact Creature — Human Villain
// {2}, 2/1 (Marvel's Spider-Man, #4460):
//
//	"Whenever Spider-Slayer deals damage to a Spider, destroy that creature.
//	 {6}, Exile this card from your graveyard: Create two tapped 1/1
//	 colorless Robot artifact creature tokens with flying."
//
// The trigger is Stinkweed Imp's, for any damage (not only combat damage) to
// a Spider: it names the damaged creature as the object it was when the
// trigger fired, so one that died to the damage, or died and came back, is
// not destroyed (CR 400.7). "A Spider" is read through the layers as the
// damage was dealt. The graveyard ability exiles the card as a cost, then
// makes the tokens tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca68f436-2522-4007-87bf-2669c7743db0",
		Name:         "Spider-Slayer, Hatred Honed",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: spiderSlayerDealtDamageToASpider,
			Key:       spiderSlayerDestroyLabel,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, spiderSlayerDestroyLabel)
				item.Params.Object = damagedCreatureRef(g, ev.Target)
				return item
			},
			Effect: destroyDamagedCreature,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{6}, Exile this card from your graveyard: Create two tapped 1/1 colorless Robot artifact creature tokens with flying.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{6}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect: Do(CreateTokenAdvanced{
				Spec: Token(TokenCard("1/1 colorless Robot artifact with flying")).EntersTapped(),
				N:    2,
			}),
		}},
	})
}

const spiderSlayerDestroyLabel = "Spider-Slayer, Hatred Honed — destroy the Spider it dealt damage to"

// spiderSlayerDealtDamageToASpider is "Whenever Spider-Slayer deals damage to
// a Spider": damage from this card, to a Spider creature still on the
// battlefield.
func spiderSlayerDealtDamageToASpider(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source != source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	if !ok || !c.IsCreature() || !c.HasSubtypeWord("Spider") {
		return false
	}
	z := g.FindCardZoneForEffect(ev.Target)
	return z != nil && z.Kind == game.ZoneBattlefield
}
