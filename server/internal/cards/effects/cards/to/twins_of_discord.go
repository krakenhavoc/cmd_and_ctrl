package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Twins of Discord — Creature — Eldrazi {7}, 8/6:
//
//	"Whenever you attack, choose odd or even. Creatures with mana value of
//	 that quality can't block this turn. (Zero is even.)
//	 Each other colorless creature you control has bloodthirst 2. (If an
//	 opponent was dealt damage this turn, that creature enters with two
//	 additional +1/+1 counters on it.)"
//
// The attack trigger fires once for the whole declaration (CR 508.1m,
// OncePerBatch), asks the controller odd or even (Extinction Event's
// pick), and registers a rule over every creature with a mana value of
// that parity until end of turn. "Can't block" changes no characteristic,
// so CR 611.2c leaves the set live: a creature that arrives later in the
// turn is covered too (ScopeCreaturesOddManaValue / ScopeCreaturesEvenManaValue).
// Zero is even, so a token is.
//
// The second line gives other colorless creatures its controller controls a
// bloodthirst 2 they would not print: a CR 614 entry replacement for each
// colourless creature that enters under that player's control (the
// Twins itself does not get it, "other"), adding two +1/+1 counters when an
// opponent was dealt damage this turn, with the same counter pipeline
// Bloodthirst uses, so Hardened Scales-style effects apply.
func init() {
	const name = "Twins of Discord"
	effects.Register(effects.Spec{
		OracleID:     "4a8f2104-57ec-4b23-9857-f26afa5cbcfe",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{twinsOfDiscordBloodthirst(name + ": bloodthirst 2 for each other colorless creature you control")},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttackDeclaredByYou(ev, source.Controller)
			}, name+" — choose odd or even; creatures with mana value of that quality can't block this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PickOption{
						Question: "Twins of Discord — choose odd or even",
						Options:  []game.ChoiceOption{{Label: "Odd"}, {Label: "Even"}},
						ThenKey:  twinsOfDiscordParity,
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}

// twinsOfDiscordParity registers the can't-block rule for the parity
// chosen. An on-disk key.
var twinsOfDiscordParity = effects.OptionPickThen("option-pick/twins-of-discord-parity", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Index < 0 {
		return nil
	}
	scope := game.ScopeCreaturesOddManaValue
	if r.Index == 1 {
		scope = game.ScopeCreaturesEvenManaValue
	}
	return effects.RestrictUntilEOT{
		Scope:        scope,
		Restrictions: game.CantBlock,
		Label:        "Twins of Discord — can't block this turn",
	}.Apply(ctx)
})

// twinsOfDiscordBloodthirst is bloodthirst 2 given to each other colorless
// creature the source's controller controls as it enters.
func twinsOfDiscordBloodthirst(label string) game.ReplacementEffect {
	dealt := effects.AnOpponentWasDealtAtLeastThisTurn(1)
	return game.ReplacementEffect{
		Label:     label,
		EntryOnly: true,
		Watches:   []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil || ev.CardID == src.InstanceID {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			return ok && entering.IsCreature() && entering.Controller == src.Controller && entering.IsColorless() && dealt(g, src.Controller)
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(game.CounterPlusOne, 2)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
	}
}
