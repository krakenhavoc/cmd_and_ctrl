package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kitsune Mystic // Autumn-Tail, Kitsune Sage — {3}{W} Creature — Fox
// Wizard 2/3 // Legendary Creature — Fox Wizard 4/5, a flip card (#3160,
// ADR 0151):
//
//	Top:     "At the beginning of the end step, if this creature is
//	          enchanted by two or more Auras, flip it."
//	Flipped: "{1}: Attach target Aura attached to a creature to another
//	          creature."
//
// The flip is every turn's end step, its intervening "if" checked as the
// step begins and again as the ability resolves (CR 603.4). Autumn-Tail
// targets only the Aura; "another creature" is chosen as the ability
// resolves, from the creatures other than the one the Aura is on, and
// the choice is offered only the ones the Aura can legally enchant (CR
// 701.3b, game.AttachLegalForEffect). AttachForEffect refuses an illegal
// host too, so an Aura that cannot move stays where it is. Autumn-Tail is
// white and {3}{W} (CR 710.1c).
//
// No simplification.
const kitsuneMysticOracleID = "b269770d-2b48-4314-a1b8-0d79beb879a3"

// enchantedByTwoAuras is "this creature is enchanted by two or more
// Auras".
func enchantedByTwoAuras(g *game.Game, host uuid.UUID) bool {
	n := 0
	for _, id := range g.AttachmentsOf(host) {
		if a, ok := g.LookupCardForEffect(id); ok && a.IsAura() {
			n++
		}
	}
	return n >= 2
}

func init() {
	Register(Spec{
		OracleID:     kitsuneMysticOracleID,
		Name:         "Kitsune Mystic",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return enchantedByTwoAuras(g, source.InstanceID)
			}, "Kitsune Mystic — flip it", func(g *game.Game, item *game.StackItem) error {
				if sourceIsNewObject(g, item) || !enchantedByTwoAuras(g, item.SourceCardID) {
					return nil
				}
				return FlipThis{}.Apply(NewContext(g, item))
			}),
		},
	})
	Register(Spec{
		OracleID:     kitsuneMysticOracleID + "#1",
		Name:         "Autumn-Tail, Kitsune Sage",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}: Attach target Aura attached to a creature to another creature.",
			Cost:  ManaCost("{1}"),
			Targets: battlefieldSpec("permanent", "target Aura attached to a creature",
				func(g *game.Game, _ uuid.UUID, c game.Card) bool {
					if !c.IsAura() || c.AttachedTo.Kind != game.TargetCard {
						return false
					}
					host, ok := g.LookupCardForEffect(c.AttachedTo.ID)
					return ok && host.IsCreature()
				}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return offerAutumnTailHosts(g, item, t.ID)
				}
				return nil
			},
		}},
	})
}

// offerAutumnTailHosts asks the controller which creature other than the
// Aura's current host to attach it to, among the ones it can enchant, and
// attaches it. Caller holds g.mu.
func offerAutumnTailHosts(g *game.Game, item *game.StackItem, auraID uuid.UUID) error {
	aura, ok := g.LookupCardForEffect(auraID)
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	current := aura.AttachedTo.ID
	var hosts []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != current && c.IsCreature() && g.AttachLegalForEffect(auraID, c.InstanceID) {
			hosts = append(hosts, c.InstanceID)
		}
	}
	if len(hosts) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Autumn-Tail, Kitsune Sage — attach " + aura.Name + " to another creature",
		Cards:    hosts,
		Min:      1,
		Max:      1,
		ThenKey:  autumnTailAttachPick,
		LiveZone: game.ZoneBattlefield,
		Carry:    []uuid.UUID{auraID},
	})
	return nil
}

// autumnTailAttachPick is the Aura (the one carried ID) attached to the
// creature picked, keyed (#4253).
var autumnTailAttachPick = game.RegisterChooseCardsThen("choose-cards/autumn-tail-kitsune-sage-attach", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	return g.AttachForEffect(r.Carry[0], game.TargetRef{Kind: game.TargetCard, ID: r.Picked[0]})
})
